package main

import (
	"fmt"
	"net"
)

func tcpDump(conn net.Conn) {
	fmt.Println("Local Address:")
	loc := conn.LocalAddr()
	fmt.Println("\tNetwork:", loc.Network())
	fmt.Println("\tString:", loc.String())

	fmt.Println("Remote Address:")
	remote := conn.RemoteAddr()
	fmt.Println("\tNetwork:", remote.Network())
	fmt.Println("\tString:", remote.String())

	fmt.Println("Start to read connection data:")
	for {
		var buf = make([]byte, 256)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("[Error] Failed to read:", err)
			return
		}
		fmt.Printf("\tRead %d Data: %v\n", n, buf)
		fmt.Printf("\tRead %d Data: %v\n", n, string(buf))
	}
}
