package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

const hostname = "127.0.0.1"

var store map[string]string
var storeMutex sync.RWMutex

type peer struct {
	addr  netAddr
	alive bool
}

var peers map[string]peer
var peersMutex sync.RWMutex

var netClient = &http.Client{
	Timeout: time.Second * 10,
}

// Collect peers from command line.
func getPeersAddr(args []string) (err error) {
	peers = make(map[string]peer)
	for _, ipPort := range args {
		var host, port string
		if host, port, err = net.SplitHostPort(ipPort); err != nil {
			return fmt.Errorf("invalid peer: %v", err)
		}
		var p int
		if p, err = strconv.Atoi(port); err != nil {
			return fmt.Errorf("peer %v: port must be an integer", ipPort)
		}
		n := netAddr{host, p}
		peers[n.String()] = peer{n, true}
	}
	return
}

func reportHTTPError(w http.ResponseWriter, httpCode int, m string) {
	w.WriteHeader(httpCode)
	fmt.Fprint(w, "error: ", m)
}

// GET /get?key=<KEYNAME>
// getHandler returns the value associated with given key.
func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			reportHTTPError(w, http.StatusBadRequest, fmt.Sprintf("%s", err))
			return
		}
		if len(values.Get("key")) == 0 {
			reportHTTPError(w, http.StatusBadRequest, "missing key")
			return
		}
		key := values.Get("key")
		storeMutex.RLock()
		value, ok := store[string(key)]
		storeMutex.RUnlock()
		if ok {
			fmt.Fprint(w, value)
		} else {
			fmt.Fprintln(w, "undefined key")
		}
	} else {
		reportHTTPError(w, http.StatusMethodNotAllowed, "only GET accepted")
	}
}

func set(w http.ResponseWriter, r *http.Request, broadcast bool) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	var key, value string
	if key = r.FormValue("key"); len(key) == 0 {
		reportHTTPError(w, http.StatusBadRequest, "missing key")
		return
	}
	if value = r.FormValue("value"); len(value) == 0 {
		reportHTTPError(w, http.StatusBadRequest, "missing value")
		return
	}
	storeMutex.Lock()
	store[key] = value
	storeMutex.Unlock()
	log.Printf("set %s = %s\n", key, value)
	if broadcast {
		broadcastUpdate(key, value)
	}
}

// POST /set key=<KEYNAME>, value=<VALUE>
// setHandler adds a new key-value pair an broadcast the update to our peers.
func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		set(w, r, true)
	} else {
		reportHTTPError(w, http.StatusMethodNotAllowed, "only POST accepted")
	}
}

// GET /peerinit
// peerinitHandler returns the contents of our store (in JSON). It is called by new peers.
func peerinitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		storeMutex.RLock()
		jsonValue, err := json.Marshal(store)
		storeMutex.RUnlock()
		if err != nil {
			reportHTTPError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonValue)
	} else {
		reportHTTPError(w, http.StatusMethodNotAllowed, "only GET accepted")
	}
}

// POST /peerregister host=<HOSTNAME>, port=<PORT>
// Sent by a peer to register itself with us.
func registerpeerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		var host, port string
		if host = r.FormValue("host"); len(host) == 0 {
			reportHTTPError(w, http.StatusBadRequest, "peer handler: missing host")
			return
		}
		if port = r.FormValue("port"); len(port) == 0 {
			reportHTTPError(w, http.StatusBadRequest, "peer handler: missing port")
			return
		}
		if p, err := strconv.Atoi(port); err == nil {
			addr := netAddr{host, p}
			peers[addr.String()] = peer{addr, true}
			log.Printf("registering new peer %s...\n", addr)
		}
	} else {
		reportHTTPError(w, http.StatusMethodNotAllowed, "only POST accepted")
	}
}

// POST /peerupdate key=<KEYNAME>, value=<VALUE>
// peerupdateHandler receives key-value updates from our peers.
func peerupdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// No broadcast here since it will cause each peer to broadcast the update, which will...
		// end up badly!
		set(w, r, false)
	} else {
		reportHTTPError(w, http.StatusMethodNotAllowed, "only POST accepted")
	}
}

// introduceOurself registers ourself with our peers.
func introduceOurself(port int) {
	for _, peer := range peers {
		if peer.alive {
			u := fmt.Sprintf("http://%s/peerregister", peer.addr)
			resp, err := netClient.PostForm(u, url.Values{
				"host": {hostname},
				"port": {strconv.Itoa(port)},
			})
			if err != nil {
				// Discard peer since it is not available.
				peer.alive = false
				log.Printf("warning: cannot contact peer %s: %v", peer.addr, err)
				continue
			}
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			resp.Body.Close()
		}
	}
}

// getPeersStore retrieve our peers' store and merge all the key-value defined in our store.
func getPeersStore() {
	for _, peer := range peers {
		if peer.alive {
			log.Printf("synchronizing with peer %s...", peer.addr)
			u := fmt.Sprintf("http://%s/peerinit", peer.addr)
			resp, err := netClient.Get(u)
			if err != nil {
				peer.alive = false
				log.Printf("warning: cannot synchronize with peer %s: %v", peer.addr, err)
				continue
			}
			defer resp.Body.Close()

			decoder := json.NewDecoder(resp.Body)
			var v map[string]string
			if err := decoder.Decode(&v); err != nil {
				log.Println("error: ", err.Error())
				return
			}
			storeMutex.RLock()
			for k, val := range v {
				store[k] = val
			}
			storeMutex.RUnlock()
		}
	}

	log.Println("store contents after initial synchronization:")
	for k, v := range store {
		log.Printf("\t%s = %s\n", k, v)
	}
	log.Println("end of store dump")
}

// broadcastUpdate broadcast a key-value update to our peers.
func broadcastUpdate(key, value string) {
	for _, peer := range peers {
		if peer.alive {
			u := fmt.Sprintf("http://%s/peerupdate", peer.addr)
			resp, err := netClient.PostForm(u, url.Values{
				"key":   {key},
				"value": {value},
			})
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("broadcasting key %s update to peer %s...\n", key, peer.addr)
		}
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		port, err := cmd.LocalFlags().GetInt("port")
		if err != nil {
			printError("could not retrieve --port value: %s", err)
		}
		if port == 0 {
			fmt.Println("error: missing server port")
			fmt.Println("Try '--help' for more information.")
			os.Exit(1)
		}
		server := netAddr{hostname, port}

		store = make(map[string]string)
		storeMutex = sync.RWMutex{}
		peersMutex = sync.RWMutex{}

		if err = getPeersAddr(args); err != nil {
			log.Fatalln("error:", err)
		}
		introduceOurself(port)
		getPeersStore()

		// Clients endpoints.
		http.HandleFunc("/get", getHandler)
		http.HandleFunc("/set", setHandler)

		// Reserved for synchronisation between peers.
		http.HandleFunc("/peerinit", peerinitHandler)
		http.HandleFunc("/peerregister", registerpeerHandler)
		http.HandleFunc("/peerupdate", peerupdateHandler)

		log.Printf("listening on %s...\n", server.String())
		if err = http.ListenAndServe(server.String(), nil); err != nil {
			log.Fatalf("error: cannot start server; %v", err)
		}
	},
}

func init() {
	var port int

	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().IntVar(&port, "port", 0x0000, "set server port")
}
