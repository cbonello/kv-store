package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/cbonello/kv-store/go/pkg/kv"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type (
	peer struct {
		addr netAddr
	}
	kvServer struct{}
)

var (
	serverIP   netAddr
	optPeersIP []string
	store      map[string]string
	storeMutex sync.RWMutex
	peers      map[string]peer
	serverCmd  = &cobra.Command{
		Use:   "server",
		Short: "Key-value store server.",
		Run:   server,
	}
)

func server(cmd *cobra.Command, args []string) {
	store = make(map[string]string)
	storeMutex = sync.RWMutex{}

	if len(optIP) == 0 {
		serverIP = netAddr{"127.0.0.1", 4000}
	} else {
		var err error
		serverIP, err = isValidIP(optIP)
		if err != nil {
			log.Fatalln("error:", err)
		}
	}
	getPeersAddr(args)

	handleRequests()
}

func handleRequests() {
	lis, err := net.Listen("tcp", serverIP.String())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterClientServer(s, &kvServer{})
	reflection.Register(s)
	fmt.Printf("Listening on %s...\n", serverIP)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *kvServer) Get(ctx context.Context, in *pb.GetKey) (*pb.GetReply, error) {
	var key = in.Key
	if value, ok := store[key]; ok {
		explain("received GET request for key '%s': value = '%s'", key, value)
		return &pb.GetReply{Value: value, Defined: true}, nil
	}
	explain("received GET request for key '%s': value = undefined", key)
	return &pb.GetReply{Value: "", Defined: false}, nil
}

func (s *kvServer) Set(ctx context.Context, in *pb.SetKey) (*pb.SetReply, error) {
	var key = in.Key
	var value = in.Value
	var broadcast = in.Broadcast
	store[key] = value
	if broadcast {
		explain("received SET request for key '%s': new value = '%s'", key, value)
		updatePeers(key, value)
	} else {
		explain("received peer update for key '%s': new value = '%s'", key, value)
	}
	return &pb.SetReply{Value: value}, nil
}

func (s *kvServer) List(ctx context.Context, _ *pb.Void) (*pb.StoreReply, error) {
	explain("received LIST request")
	return &pb.StoreReply{Store: store}, nil
}

func (s *kvServer) RegisterWithPeer(ctx context.Context, in *pb.IP) (*pb.StoreReply, error) {
	var peerIP = in.Ip
	explain("received new peer registration: %s", peerIP)
	if ip, err := isValidIP(peerIP); err == nil {
		if _, ok := peers[peerIP]; ok == false {
			peers[peerIP] = peer{ip}
		}
	}
	return &pb.StoreReply{Store: store}, nil
}

func introduceOurself(peerIP netAddr) {
	explain("registering with peer %s...", peerIP)
	conn, err := grpc.Dial(peerIP.String(), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewClientClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.RegisterWithPeer(ctx, &pb.IP{Ip: serverIP.String()})
}

func updatePeers(key, value string) {
	for _, peerIP := range peers {
		explain("updating peer '%s': '%s' = '%s'", peerIP.addr, key, value)
		conn, err := grpc.Dial(peerIP.addr.String(), grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := pb.NewClientClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		c.Set(ctx, &pb.SetKey{Key: key, Value: value})
	}
}

func getPeersAddr(args []string) (err error) {
	peers = make(map[string]peer)
	for _, peerIP := range args {
		var ip netAddr
		ip, err = isValidIP(peerIP)
		if err != nil {
			log.Fatalln("error:", err)
		}
		peers[ip.String()] = peer{ip}
		introduceOurself(ip)
	}
	return
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&optIP, "ip", "i", "", "set server IP address (IPv4 only!)")
}
