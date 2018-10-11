package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

type Manager struct {
	listener net.Listener
	quit     bool
	// quitLocker sync.Mutex
}

var (
	port = flag.String("p", "8081", "the listen port")

	_m *Manager
)

func main() {
	flag.Parse()

	if len(os.Args) == 2 && (os.Args[1] == "-h" ||
		os.Args[1] == "--help") {
		flag.Usage()
		return
	}

	listener, err := net.Listen("tcp", "127.0.0.1:"+*port)
	if err != nil {
		fmt.Println("Failed to listen port:", err, *port)
		return
	}

	_m = &Manager{listener: listener, quit: false}
	_m.Loop()
}

func (m *Manager) Loop() {
	for {
		if m.canQuit() {
			fmt.Println("Quit...")
			return
		}
		conn, err := m.accept()
		if err != nil {
			fmt.Println("Failed to accept:", err)
			return
		}
		go m.handle(conn)
	}
}

func (m *Manager) accept() (net.Conn, error) {
	// if temporary error, continue
	var tempDelay time.Duration
	for {
		conn, err := m.listener.Accept()
		if err == nil {
			tempDelay = 0
			return conn, nil
		}
		ee, ok := err.(net.Error)
		if !ok || !ee.Temporary() {
			return nil, err
		}
		if tempDelay == 0 {
			tempDelay = 5 * time.Millisecond
		} else {
			tempDelay *= 2
		}

		if max := 1 * time.Second; tempDelay > max {
			tempDelay = max
		}
		time.Sleep(tempDelay)
	}
}

func (m *Manager) handle(conn net.Conn) {
	defer conn.Close()
	connDump(conn)
}

func (m *Manager) canQuit() bool {
	// m.quitLocker.Lock()
	r := m.quit
	// m.quitLocker.Unlock()
	return r
}

func (m *Manager) doQuit() {
	// m.quitLocker.Lock()
	m.quit = true
	// m.quitLocker.Unlock()
}

// conn only contains data
func connDump(conn net.Conn) {
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
		fmt.Printf("\tRead %d Data []byte: %v\n", n, buf)
		fmt.Printf("\tRead %d Data string: %v\n", n, string(buf))
	}
}
