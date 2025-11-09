package main

// go run . --id=1 --port=:50051 --peers=:50052
// go run . --id=2 --port=:50052 --peers=:50051

import (
	"context"
	"flag"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	pb "ra/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Peer struct {
	addr   string
	conn   *grpc.ClientConn
	client pb.EchoClient
}

type Node struct {
	pb.UnimplementedEchoServer

	id         int
	addr       string
	peers      map[string]*Peer
	timestamp  int64
	replies    int
	requesting bool
	deferred   map[int]bool

	mu     sync.Mutex
	server *grpc.Server
}

func main() {
	n := readArgs()

	log.Println("\n--- node init ---")

	err := n.startServer()
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}

	for addr := range n.peers {
		go n.dialPeer(addr)
	}

	select {}
}

func readArgs() *Node {
	id := flag.Int("id", 0, "node id")
	port := flag.String("port", ":5000", "listen addr")
	ps := flag.String("peers", "", "comma-separated peer addrs")
	flag.Parse()

	peers := map[string]*Peer{}
	for _, s := range strings.Split(*ps, ",") {
		s = "localhost" + strings.TrimSpace(s)
		peers[s] = &Peer{addr: s}
	}

	return &Node{
		id:     *id,
		addr:   "localhost" + *port,
		peers:  peers,
		server: grpc.NewServer(),
	}
}

func (n *Node) startServer() error {
	lis, err := net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}

	pb.RegisterEchoServer(n.server, n)

	go func() {
		log.Printf("listening on %s", n.addr)
		if err := n.server.Serve(lis); err != nil {
			log.Fatalf("grpc Serve error: %v", err)
		}
	}()
	return nil
}

func (n *Node) dialPeer(addr string) {
	for {
		_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		cancel()

		if err != nil {
			log.Printf("dial %s: %v (retrying)", addr, err)
			time.Sleep(800 * time.Millisecond)
			continue
		}

		n.mu.Lock()
		if p, ok := n.peers[addr]; ok {
			p.conn = conn
			p.client = pb.NewEchoClient(conn)
		}
		n.mu.Unlock()

		log.Printf("connected -> %s", addr)
		return
	}
}

func (n *Node) Msg(ctx context.Context, req *pb.Request) (*pb.Reply, error) {
	n.mu.Lock()

	if int64(req.Timestamp) >= n.timestamp {
		n.timestamp = int64(req.Timestamp) + 1
	} else {
		n.timestamp++
	}

	ts := n.timestamp
	n.mu.Unlock()

	return &pb.Reply{Id: uint64(n.id), Timestamp: uint64(ts)}, nil
}

func (n *Node) requestAccess() {
	n.mu.Lock()
	n.timestamp++
	n.requesting = true
	n.replies = 0
	n.mu.Unlock()

	for _, p := range n.peers {
		go func(pc pb.EchoClient, a string) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			_, err := pc.Msg(ctx, &pb.Request{Id: uint64(n.id), Timestamp: uint64(n.timestamp)})
			if err != nil {
				n.mu.Lock()
				n.replies++
				n.mu.Unlock()
			}
		}(p.client, p.addr)
	}

	for n.replies != len(n.peers)-1 {
		time.Sleep(time.Second)
	}
}
