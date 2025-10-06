package main

import (
	"fmt"
	"net"
	"tcp-sim/computer"
	"tcp-sim/header"
)

func main() {
	var c computer.Computer
	c.HandleArgs()
	c.Dial()
	defer c.Conn.Close()

	// send syn
	var syn header.Header
	header.SetSyn(&syn)
	header.SetSeq(&syn, c.Seq)
	header.FillPorts(&syn, c.Conn, c.Conn.RemoteAddr().(*net.UDPAddr).Port)

	c.Conn.Write(syn[:])
	syn.Print("CLIENT: SENT SYN")

	// receive synack
	synAck, _ := c.ReadHeader()
	if header.IsSynAck(&synAck) {
		synAck.Print("CLIENT: RECEIVED SYNACK")
	}

	// send ack
	var ack header.Header
	header.SetAck(&ack)
	header.SetSeq(&ack, header.GetAckNum(&synAck))
	header.SetAckNum(&ack, header.GetSeq(&synAck)+1)
	header.FillPorts(&ack, c.Conn, c.Conn.RemoteAddr().(*net.UDPAddr).Port)

	c.Conn.Write(ack[:])
	ack.Print("CLIENT: SENT ACK")

	fmt.Printf("\nHandshake complete.")
}
