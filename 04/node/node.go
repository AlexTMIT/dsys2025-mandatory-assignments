package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"sync"
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

	fmt.Println("Node init:")
	fmt.Println(n)

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

	peers := strings.Join(n.peers, ", ")

	deferredIDs := make([]int, 0, len(n.deferred))
	for id := range n.deferred {
		deferredIDs = append(deferredIDs, id)
	}
	sort.Ints(deferredIDs)

	deferredStrs := make([]string, len(deferredIDs))
	for i, id := range deferredIDs {
		deferredStrs[i] = fmt.Sprintf("%d", id)
	}
	deferred := strings.Join(deferredStrs, ", ")

	status := "idle"
	if n.requesting {
		status = fmt.Sprintf("requesting@%d", n.reqTimestamp)
	}

	return fmt.Sprintf(
		"Node{id:%d, addr:%s, lamport:%d, status:%s, peers:[%s], deferred:[%s]}",
		n.id, n.addr, n.timestamp, status, peers, deferred,
	)
}
