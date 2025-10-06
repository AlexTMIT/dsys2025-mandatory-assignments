// tcp header 20 bytes, w/o options
package header

import (
	"fmt"
	"net"
	"strings"
)

type Header [20]byte

const ( // flags
	SYN = 1 << 1
	ACK = 1 << 4
)

func SetSyn(h *Header) {
	h[13] |= SYN // bitwise or since flags are bits in a byte
}

func IsSyn(h *Header) bool {
	return h[13]&SYN != 0
}

func SetAck(h *Header) {
	h[13] |= ACK
}

func IsAck(h *Header) bool {
	return h[13]&ACK != 0
}

func SetSynAck(h *Header) {
	SetAck(h)
	SetSyn(h)
}

func IsSynAck(h *Header) bool {
	return IsSyn(h) && IsAck(h)
}

func SetSeq(h *Header, seq uint32) {
	h[4] = byte(seq >> 24)
	h[5] = byte(seq >> 16)
	h[6] = byte(seq >> 8)
	h[7] = byte(seq)
}

func GetSeq(h *Header) uint32 {
	return (uint32(h[4]) << 24) |
		(uint32(h[5]) << 16) |
		(uint32(h[6]) << 8) |
		uint32(h[7])
}

func SetAckNum(h *Header, ack uint32) {
	h[8] = byte(ack >> 24)
	h[9] = byte(ack >> 16)
	h[10] = byte(ack >> 8)
	h[11] = byte(ack)
}

func GetAckNum(h *Header) uint32 {
	return (uint32(h[8]) << 24) |
		(uint32(h[9]) << 16) |
		(uint32(h[10]) << 8) |
		uint32(h[11])
}

func SetSrcPort(h *Header, port uint16) {
	h[0] = byte(port >> 8)
	h[1] = byte(port)
}

func GetSrcPort(h *Header) uint16 {
	hi := uint16(h[0]) << 8
	lo := uint16(h[1])
	return hi | lo
}

func GetDstPort(h *Header) uint16 {
	hi := uint16(h[2]) << 8
	lo := uint16(h[3])
	return hi | lo
}

func SetDstPort(h *Header, port uint16) {
	h[2] = byte(port >> 8)
	h[3] = byte(port)
}

func FillPorts(h *Header, conn *net.UDPConn, dst int) {
	localPort := conn.LocalAddr().(*net.UDPAddr).Port
	remotePort := dst
	SetSrcPort(h, uint16(localPort))
	SetDstPort(h, uint16(remotePort))
}

// Print method by ChatGPT (because not relevant to course, just fun to see)
func (h *Header) Print(label string) {
	fmt.Printf("\n[%s]\n", label)

	// Ports
	fmt.Printf(" SrcPort: %-5d  DstPort: %-5d\n", GetSrcPort(h), GetDstPort(h))

	// Sequence & Ack
	fmt.Printf(" Seq: %-10d  Ack: %-10d\n", GetSeq(h), GetAckNum(h))

	// Flags
	flags := []string{}
	if IsSyn(h) {
		flags = append(flags, "SYN")
	}
	if IsAck(h) {
		flags = append(flags, "ACK")
	}
	if len(flags) == 0 {
		flags = append(flags, "NONE")
	}
	fmt.Printf(" Flags: %s\n", strings.Join(flags, " | "))

	// Raw bytes in rows of 8
	fmt.Print(" Raw:  ")
	for i, b := range h {
		fmt.Printf("%02X ", b)
		if (i+1)%8 == 0 {
			fmt.Print("\n        ")
		}
	}
	fmt.Println()
}
