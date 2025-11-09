package main

import (
	"flag"
	"fmt"
	"strings"
)

func main() {
	id := flag.Int("id", 0, "node id")
	port := flag.String("port", ":5000", "listen port")
	peers := flag.String("peers", "", "comma-separated peer addresses")
	flag.Parse()

	peerList := []string{}
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}

	fmt.Println("ID:", *id)
	fmt.Println("Port:", *port)
	fmt.Println("Peers:", peerList)
}
