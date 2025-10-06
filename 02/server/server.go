package main

import (
	"fmt"
	"tcp-sim/computer"
	"tcp-sim/header"
)

func main() {
	var s computer.Computer
	s.HandleArgs()
	s.Listen()
	defer s.Conn.Close()

	// wait for syn
	syn, clientAddr := s.ReadHeader()
	if header.IsSyn(&syn) {
		syn.Print("SERVER: RECEIVED SYN")
	}

	// send synack
	var synAck header.Header
	header.SetSynAck(&synAck)
	header.SetSeq(&synAck, s.Seq)
	header.SetAckNum(&synAck, header.GetSeq(&syn)+1)
	header.FillPorts(&synAck, s.Conn, int(header.GetSrcPort(&syn)))

	s.SendHeader(synAck, clientAddr)
	synAck.Print("SERVER: SENT SYNACK")

	// receive ack
	ack, _ := s.ReadHeader()
	if header.IsAck(&ack) {
		ack.Print("SERVER: RECEIVED ACK")
		fmt.Printf("\nHandshake complete.")
	}
}
