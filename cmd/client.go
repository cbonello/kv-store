package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func getServerAddr(s string) (server netAddr, err error) {
	var host, port string
	if host, port, err = net.SplitHostPort(s); err != nil {
		return server, fmt.Errorf("invalid peer: %v", err)
	}
	p, _ := strconv.Atoi(port)
	server = netAddr{host, p}
	return
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Println("error: missing server info")
			log.Println("Try '--help' for more information.")
			os.Exit(1)
		}
		ip, _ := getServerAddr(args[0])
		server := fmt.Sprintf("http://%s", ip)

		var netClient = &http.Client{
			Timeout: time.Second * 10,
		}
		get, err := cmd.LocalFlags().GetString("get")
		if err != nil {
			printError("could not retrieve --get value: %s", err)
		}
		if len(get) != 0 {
			re := regexp.MustCompile(`[a-zA-Z0-9]+`)
			if re.MatchString(get) == false {
				log.Fatalf("error: -get %v: malformed input", get)
			}
			u := fmt.Sprintf("%s/get?key=%s", server, get)
			resp, err := netClient.Get(u)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			buf, _ := ioutil.ReadAll(resp.Body)
			log.Printf("GET: %s = %s\n", get, buf)
		}

		set, err := cmd.LocalFlags().GetString("set")
		if err != nil {
			printError("could not retrieve --set value: %s", err)
		}
		if len(set) != 0 {
			re := regexp.MustCompile(`[a-zA-Z0-9]+=[a-zA-Z0-9]+`)
			if re.MatchString(set) == false {
				log.Fatalf("error: -set %v: malformed input", set)
			}
			eq := strings.Index(set, "=")
			key := set[:eq]
			value := set[eq+1:]

			u := fmt.Sprintf("%s/set", server)
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
			log.Println("SET: Success")
		}
	},
}

func init() {
	var get, set string

	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVarP(&get, "get", "g", "", "return value associated with KEY")
	clientCmd.Flags().StringVarP(&set, "set", "s", "", "set KEY to VALUE")
}
