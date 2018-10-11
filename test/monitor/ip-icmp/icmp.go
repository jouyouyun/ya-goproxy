package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/icmp"
	"net"
)

var (
	addr = flag.String("a", "127.0.0.1", "the listen address")
)

func main() {
	flag.Parse()

	laddr, err := net.ResolveIPAddr("ip4", *addr)
	if err != nil {
		fmt.Println("Failed to resolve ip addr:", *addr, err)
		return
	}

	conn, err := net.ListenIP("ip4:icmp", laddr)
	if err != nil {
		fmt.Println("Failed to listen ip:", *addr, err)
		return
	}
	defer conn.Close()

	for {
		var buf = make([]byte, 1024)
		n, saddr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("Failed to read from:", err)
			continue
		}
		fmt.Println("Read data:", n, saddr.String())
		// handle data, proto: IP4ICMP, IP6ICMP
		msg, err := icmp.ParseMessage(1, buf)
		if err != nil {
			fmt.Println("Failed to parse icmp message:", err)
			continue
		}
		fmt.Println("ICMP:", msg.Type, msg.Code, msg.Checksum)
	}
}
