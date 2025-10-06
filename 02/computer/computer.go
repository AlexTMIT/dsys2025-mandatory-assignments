package computer

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"tcp-sim/header"
)

var BUF = make([]byte, 20)

type Computer struct {
	Port string
	Seq  uint32
	Conn *net.UDPConn
}

func (c *Computer) HandleArgs() {
	expectedArgs := 3

	if len(os.Args) != expectedArgs {
		fmt.Println("Usage:", "go run ./server|client <server-port> <isn>")
		os.Exit(1)
	}

	c.Port = os.Args[1]
	seqInt, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	c.Seq = uint32(seqInt)
}

func (c *Computer) GetAddr(listen bool) *net.UDPAddr {
	addr, err := net.ResolveUDPAddr("udp", ":"+c.Port)
	if err != nil {
		panic(err)
	}
	if listen {
		fmt.Printf("Listening on port %s\n", c.Port)
	}
	return addr
}

func (c *Computer) Listen() {
	addr := c.GetAddr(true)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	c.Conn = conn
}

func (c *Computer) Dial() {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+c.Port)
	if err != nil {
		panic(err)
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	c.Conn = conn
}

func (c *Computer) ReadHeader() (header.Header, *net.UDPAddr) {
	_, clientAddr, err := c.Conn.ReadFromUDP(BUF)
	if err != nil {
		panic(err)
	}
	var h header.Header
	copy(h[:], BUF)
	return h, clientAddr
}

func (c *Computer) SendHeader(h header.Header, addr *net.UDPAddr) {
	_, err := c.Conn.WriteToUDP(h[:], addr)
	if err != nil {
		panic(err)
	}
}
