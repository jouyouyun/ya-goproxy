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
	tcpDump(conn)
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
