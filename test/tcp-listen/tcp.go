package main

import (
	"flag"
	"fmt"
	"net"
	"time"
)

var (
	addr = flag.String("a", "127.0.0.1:8082", "the tcp listen address")
)

func main() {
	flag.Parse()

	laddr, err := net.ResolveTCPAddr("tcp", *addr)
	if err != nil {
		fmt.Println("Failed to resolve address:", *addr, err)
		return
	}

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		fmt.Println("Failed to listen tcp:", *addr, err)
		return
	}
	defer listener.Close()

	for {
		tconn, err := listener.AcceptTCP()
		if err != nil {
			ne, ok := err.(net.Error)
			if !ok || ne == nil {
				fmt.Println("Failed to accept:", err)
				break
			}
			if !ne.Temporary() {
				fmt.Println("Failed to accept:", err)
				break
			}
			time.Sleep(time.Millisecond * 5)
		}
		fmt.Println("Conn info:", tconn.LocalAddr(), tconn.RemoteAddr())
		go handleTCPConn(tconn)
	}
}

func handleTCPConn(conn *net.TCPConn) {
	if conn == nil {
		return
	}
	defer conn.Close()

	var buf = make([]byte, 1480)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Failed to read from:", err)
		return
	}
	// fmt.Println("Read bytes:", n, buf)
	fmt.Println("Read string:", n, string(buf))

	conn.Write([]byte("Hello, TCP Monitor\n"))
}
