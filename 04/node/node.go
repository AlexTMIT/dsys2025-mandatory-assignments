package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"
)

type Node struct {
	id           int
	addr         string
	peers        []string
	timestamp    int64
	requesting   bool
	reqTimestamp int64
	deferred     map[int]bool
	mu           sync.Mutex
}

func main() {
	n := readArgs()

	fmt.Println("\n--- node init ---")
	fmt.Printf("%s\n\n", n)

	for i := range n.peers {
		go listen(n.peers[i])
	}

	for {
		time.Sleep(1000)
	}
}

func readArgs() *Node {
	id := flag.Int("id", 0, "node id")
	port := flag.String("port", ":5000", "listen port")
	p := flag.String("peers", "", "comma-separated peer addresses")
	flag.Parse()
	peers := strings.Split(*p, ",")

	return &Node{
		id:           *id,
		addr:         *port,
		peers:        peers,
		timestamp:    0,
		requesting:   false,
		reqTimestamp: 0,
		deferred:     make(map[int]bool),
		mu:           sync.Mutex{},
	}
}

func (n *Node) String() string {
	n.mu.Lock()
	defer n.mu.Unlock()

	status := "idle"
	if n.requesting {
		status = fmt.Sprintf("requesting@%d", n.reqTimestamp)
	}

	peers := "[" + strings.Join(n.peers, ", ") + "]"

	ids := make([]string, 0, len(n.deferred))
	for id := range n.deferred {
		ids = append(ids, fmt.Sprint(id))
	}
	deferred := "[" + strings.Join(ids, ", ") + "]"

	return fmt.Sprintf(
		"Node{id:%d, addr:%s, lamport:%d, status:%s, peers:%s, deferred:%s}",
		n.id, n.addr, n.timestamp, status, peers, deferred,
	)
}

func listen(peer string) {
	fmt.Println("going with peer " + peer)
}
