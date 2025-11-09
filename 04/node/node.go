package main

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
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Node struct {
	pb.UnimplementedRAServer

	addr  string
	peers map[string]pb.RAClient

	ts         uint64
	reqTs      uint64
	requesting bool
	replies    int
	deferred   map[string]bool
	mu         sync.Mutex
	srv        *grpc.Server
}

func main() {
	n := readArgs()
	if err := n.startServer(); err != nil {
		log.Fatal(err)
	}

	for a := range n.peers {
		go n.dialPeer(a)
	}

	n.waitPeersReady()

	go func() {
		for {
			time.Sleep(3 * time.Second)
			log.Printf("[%s] REQUEST CS", n.addr)
			n.requestAccess()
			log.Printf("[%s] ENTER  CS", n.addr)
			time.Sleep(time.Second)
			log.Printf("[%s] EXIT   CS", n.addr)
			n.exitCS()
		}
	}()

	select {}
}

func readArgs() *Node {
	port := flag.String("port", ":50051", "listen addr")
	plist := flag.String("peers", "", "comma-separated peer ports, e.g. :50052,:50053")
	flag.Parse()

	peers := map[string]pb.RAClient{}
	for _, p := range strings.Split(*plist, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		addr := "localhost" + p
		peers[addr] = nil
	}

	return &Node{
		addr:     "localhost" + *port,
		peers:    peers,
		deferred: map[string]bool{},
		srv:      grpc.NewServer(),
	}
}

func (n *Node) startServer() error {
	lis, err := net.Listen("tcp", n.addr)
	if err != nil {
		return err
	}
	pb.RegisterRAServer(n.srv, n)
	go func() {
		log.Printf("[node %s] listening", n.addr)
		if err := n.srv.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()
	return nil
}

func (n *Node) dialPeer(addr string) {
	for {
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			time.Sleep(600 * time.Millisecond)
			continue
		}
		client := pb.NewRAClient(conn)

		pingCtx, pingCancel := context.WithTimeout(context.Background(), 1*time.Second)
		_, perr := client.Ping(pingCtx, &emptypb.Empty{})
		pingCancel()
		if perr != nil || conn.GetState() != connectivity.Ready {
			_ = conn.Close()
			time.Sleep(400 * time.Millisecond)
			continue
		}

		n.mu.Lock()
		n.peers[addr] = client
		n.mu.Unlock()
		log.Printf("[%s] connected -> %s", n.addr, addr)
		return
	}
}

func (n *Node) waitPeersReady() {
	for {
		n.mu.Lock()
		total := len(n.peers)
		ok := 0
		// heartbeat sweep
		for addr, c := range n.peers {
			if c == nil {
				continue
			}
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			_, err := c.Ping(ctx, &emptypb.Empty{})
			cancel()
			if err == nil {
				ok++
			} else {
				n.peers[addr] = nil
				go n.dialPeer(addr)
			}
		}
		n.mu.Unlock()
		if ok == total {
			log.Printf("[%s] all peers reachable (%d/%d); starting", n.addr, ok, total)
			return
		}
		log.Printf("[%s] waiting for peers: %d/%d reachable", n.addr, ok, total)
		time.Sleep(time.Second * 5)
	}
}

func (n *Node) Request(ctx context.Context, r *pb.Req) (*emptypb.Empty, error) {
	n.mu.Lock()
	oldTs := n.ts
	if r.Ts >= n.ts {
		n.ts = r.Ts + 1
	} else {
		n.ts++
	}

	meHasPriority := n.requesting && (n.reqTs < r.Ts || (n.reqTs == r.Ts && n.addr < r.Addr))
	if meHasPriority {
		n.deferred[r.Addr] = true
		log.Printf("[%s] recv Request from %s -> DEFER (my reqTs=%d < their ts=%d) | ts %d->%d", n.addr, r.Addr, n.reqTs, r.Ts, oldTs, n.ts)
		n.mu.Unlock()
		return &emptypb.Empty{}, nil
	}

	to := r.Addr
	log.Printf("[%s] recv Request from %s -> REPLY immediately (my reqTs=%d, their ts=%d) | ts %d->%d", n.addr, to, n.reqTs, r.Ts, oldTs, n.ts)
	n.mu.Unlock()
	go n.replyTo(to)
	return &emptypb.Empty{}, nil
}

func (n *Node) Reply(ctx context.Context, r *pb.Rep) (*emptypb.Empty, error) {
	n.mu.Lock()
	oldTs := n.ts
	if r.Ts >= n.ts {
		n.ts = r.Ts + 1
	} else {
		n.ts++
	}
	if n.requesting {
		n.replies++
		log.Printf("[%s] recv Reply #%d/%d | ts %d->%d", n.addr, n.replies, len(n.peers), oldTs, n.ts)
	}
	n.mu.Unlock()
	return &emptypb.Empty{}, nil
}

func (n *Node) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

func (n *Node) requestAccess() {
	n.mu.Lock()
	log.Printf("[%s] requesting CS (ts=%d)", n.addr, n.ts)
	n.ts++
	n.reqTs = n.ts
	n.requesting = true
	n.replies = 0

	localTs := n.reqTs
	fromAddr := n.addr
	clients := make([]pb.RAClient, 0, len(n.peers))
	for _, c := range n.peers {
		if c != nil {
			clients = append(clients, c)
		}
	}

	expected := len(clients)
	n.mu.Unlock()

	for _, c := range clients {
		go func(c pb.RAClient) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, _ = c.Request(ctx, &pb.Req{Ts: localTs, Addr: fromAddr})
			cancel()
		}(c)
	}

	for {
		time.Sleep(50 * time.Millisecond)
		n.mu.Lock()
		done := n.replies >= expected
		n.mu.Unlock()
		if done {
			break
		}
	}
}

func (n *Node) exitCS() {
	n.mu.Lock()
	n.requesting = false
	deferList := make([]string, 0, len(n.deferred))
	for addr := range n.deferred {
		deferList = append(deferList, addr)
	}
	n.deferred = map[string]bool{}
	n.mu.Unlock()

	for _, addr := range deferList {
		log.Printf("[%s] releasing deferred reply to %s", n.addr, addr)
		n.replyTo(addr)
	}
}

func (n *Node) replyTo(addr string) {
	n.mu.Lock()
	oldTs := n.ts
	c := n.peers[addr]
	if c == nil {
		n.mu.Unlock()
		return
	}
	n.ts++
	curTs := n.ts
	n.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := c.Reply(ctx, &pb.Rep{Ts: curTs})
	cancel()
	if err != nil {
		log.Printf("[%s] send Reply -> %s FAILED: %v", n.addr, addr, err)
	} else {
		log.Printf("[%s] send Reply -> %s (ts %d->%d)", n.addr, addr, oldTs, curTs)
	}
}
