package cmd

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	pb "github.com/cbonello/kv-store/go/pkg/kv"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	optGet, optSet string
	optList        bool
	clientCmd      = &cobra.Command{
		Use:   "client",
		Short: "Send request(s) to a key-value store server.",
		Run:   client,
	}
)

func client(cmd *cobra.Command, args []string) {
	var ip netAddr
	if len(optIP) == 0 {
		ip = netAddr{"127.0.0.1", 4000}
	} else {
		var err error
		ip, err = isValidIP(optIP)
		if err != nil {
			log.Fatalln("error:", err)
		}
	}

	if len(optGet) != 0 {
		re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
		if re.MatchString(optGet) == false {
			log.Fatalf("error: invalid --get: expected '--get KEY'; got '--get %s'", optGet)
		}
		doGet(ip, optGet)
	}

	if len(optSet) != 0 {
		re := regexp.MustCompile(`^[a-zA-Z0-9]+=[a-zA-Z0-9]+$`)
		if re.MatchString(optSet) == false {
			log.Fatalf("error: invalid --set: expected '--set KEY=VALUE'; got '--set %s'", optSet)
		}
		eq := strings.Index(optSet, "=")
		key := optSet[:eq]
		value := optSet[eq+1:]
		doSet(ip, key, value)
	}

	if optList {
		doList(ip)
	}
}

func doGet(ip netAddr, key string) {
	explain("sending GET request to %s for key '%s'...", ip, key)
	conn, err := grpc.Dial(ip.String(), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewClientClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Get(ctx, &pb.GetKey{Key: key})
	if err != nil {
		statusCode, _ := status.FromError(err)
		log.Fatalf("%v: could not get key: %v", statusCode, err)
	}
	if r.Defined {
		fmt.Printf("'%s'='%s'\n", key, r.Value)
	} else {
		fmt.Printf("'%s': undefined\n", key)
	}
}

func doSet(ip netAddr, key, value string) {
	explain("sending SET request to %s for key '%s'...", ip, key)
	conn, err := grpc.Dial(ip.String(), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewClientClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = c.Set(ctx, &pb.SetKey{Key: key, Value: value, Broadcast: true})
	if err != nil {
		log.Fatalf("could not set key-value pair: %v", err)
	}
}

func doList(ip netAddr) {
	explain("sending LIST request to %s...", ip)
	conn, err := grpc.Dial(ip.String(), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewClientClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.List(ctx, &pb.Void{})
	if err != nil {
		log.Fatalf("could not get key-value pairs: %v", err)
	}
	fmt.Printf("Key-value pairs defined on %s:\n", ip)
	for key, value := range r.Store {
		fmt.Printf("  - '%s'='%s'\n", key, value)
	}
	fmt.Print("-- end of key-value dump --\n")
}

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.Flags().StringVarP(&optIP, "ip", "i", "", "set server IP address (IPv4 only!)")
	clientCmd.Flags().StringVarP(&optGet, "get", "g", "", "get value associated with key")
	clientCmd.Flags().StringVarP(&optSet, "set", "s", "", "set a key-value pair")
	clientCmd.Flags().BoolVarP(&optList, "list", "l", false, "get key-value pairs defined on server")
}
