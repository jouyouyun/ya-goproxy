package main

import (
	"flag"
	"fmt"
	"github.com/jouyouyun/ya-goproxy/protocol"
	"net"
)

var (
	addr = flag.String("a", "127.0.0.1", "the listen address")
)

func main() {
	flag.Parse()

	laddr, err := net.ResolveIPAddr("ip4", *addr)
	if err != nil {
		fmt.Println("Failed to resolve address:", *addr, err)
		return
	}

	conn, err := net.ListenIP("ip4:tcp", laddr)
	if err != nil {
		fmt.Println("Failed to listen ip:", *addr, err)
		return
	}
	defer conn.Close()

	for {
		var buf = make([]byte, 1480)
		n, raddr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("Failed to read from:", err)
			continue
		}
		fmt.Println("Read data:", n, raddr)
		h := protocol.UnmarshalTCPHeader(buf)
		fmt.Println("TCP:", h.String())
	}
}
