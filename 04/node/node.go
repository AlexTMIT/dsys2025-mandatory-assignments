package main

import (
	"flag"
	"fmt"
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

	id := flag.Int("id", 0, "node id")
	port := flag.String("port", ":5000", "listen port")
	p := flag.String("peers", "", "comma-separated peer addresses")

	flag.Parse()
	peers := strings.Split(*p, ",")

	n := &Node{
		id:           *id,
		addr:         *port,
		peers:        peers,
		timestamp:    0,
		requesting:   false,
		reqTimestamp: 0,
		deferred:     make(map[int]bool),
		mu:           sync.Mutex{},
	}

	fmt.Printf("node init: \n%+v\n", n)
}
