package main

import (
	"flag"
	"fmt"
	"net"
	"syscall"
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
	fmt.Println("Listen successfully:", *addr)

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
		oAddr, oConn, err := getOriginDstAddr(tconn)
		tconn.Close()
		if err != nil {
			fmt.Println("Failed to get origin dst addr:", err)
			continue
		}
		fmt.Println("Origin dst addr:", oAddr.String())
		go handleTCPConn(oConn)
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

func getOriginDstAddr(conn *net.TCPConn) (net.Addr, *net.TCPConn, error) {
	fr, err := conn.File()
	if err != nil {
		return nil, nil, err
	}
	defer fr.Close()

	mreq, err := syscall.GetsockoptIPv6Mreq(int(fr.Fd()), syscall.IPPROTO_IP, 80)
	if err != nil {
		return nil, nil, err
	}
	//fmt.Println("Ipv6 interface:", mreq.Interface, ", mreq:", mreq.Multiaddr)

	// only support ip4
	ip := net.IPv4(mreq.Multiaddr[4], mreq.Multiaddr[5], mreq.Multiaddr[6],
		mreq.Multiaddr[7])
	port := uint16(mreq.Multiaddr[2])<<8 + uint16(mreq.Multiaddr[3])
	addr, err := net.ResolveTCPAddr("tcp4",
		fmt.Sprintf("%s:%d", ip.String(), port))
	if err != nil {
		return nil, nil, err
	}

	fcc, err := net.FileConn(fr)
	if err != nil {
		return nil, nil, err
	}

	c, ok := fcc.(*net.TCPConn)
	if !ok {
		return nil, nil, fmt.Errorf("not a TCP connection")
	}
	return addr, c, nil
}
