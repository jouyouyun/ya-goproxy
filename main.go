package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Manager struct {
	tcpListener net.Listener
	udpListener net.Listener
}

const (
	pidFile = "/tmp/ya-goproxy.pid"
)

var (
	tcpAddress = flag.String("p", "127.0.0.1:8081", "the tcp listen address")

	_m *Manager
)

func main() {
	flag.Parse()

	if checkExists() {
		fmt.Println("There has a 'ya-goproxy' running...")
		return
	}
	_m = new(Manager)
	err := _m.init()
	if err != nil {
		fmt.Println("Failed to init:", err)
		return
	}
	defer _m.finalize()

	err = genPIDFile()
	if err != nil {
		fmt.Println("Failed to generate pid file:", err)
		return
	}
	defer os.Remove(pidFile)

	go _m.handleSignal()
	_m.TCPLoop()
}

func (m *Manager) init() error {
	tcp, err := net.Listen("tcp", *tcpAddress)
	if err != nil {
		return err
	}
	m.tcpListener = tcp
	return nil
}

func (m *Manager) finalize() {
	if m.tcpListener != nil {
		m.tcpListener.Close()
		m.tcpListener = nil
	}
	if m.udpListener != nil {
		m.udpListener.Close()
		m.udpListener = nil
	}
}

func (m *Manager) handleSignal() {
	var c = make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	fmt.Println("Recieve signal:", sig.String())
	_m.finalize()
	os.Exit(-1)
}

func (m *Manager) TCPLoop() {
	for {
		conn, err := m.tcpAccept()
		if err != nil {
			fmt.Println("Failed to accept:", err)
			return
		}
		go m.tcpHandle(conn)
	}
}

func (m *Manager) tcpAccept() (net.Conn, error) {
	// if temporary error, continue
	var tempDelay time.Duration
	for {
		conn, err := m.tcpListener.Accept()
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

func (m *Manager) tcpHandle(conn net.Conn) {
	defer conn.Close()
	dumpTCPConn(conn)
}

// conn only contains data
func dumpTCPConn(conn net.Conn) {
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

func transport(rw1, rw2 io.ReadWriter) error {
	errc := make(chan error, 1)
	go func() {
		_, err := io.Copy(rw1, rw2)
		errc <- err
	}()

	go func() {
		_, err := io.Copy(rw2, rw1)
		errc <- err
	}()

	err := <-errc
	if err != nil && err == io.EOF {
		err = nil
	}
	return err
}

func genPIDFile() error {
	pid := os.Getpid()
	return ioutil.WriteFile(pidFile, []byte(fmt.Sprint(pid)), 0644)
}

func checkExists() bool {
	contents, err := ioutil.ReadFile(pidFile)
	if err != nil {
		fmt.Println("Failed to read pid file:", err)
		return false
	}
	pid, err := strconv.Atoi(string(contents))
	if err != nil {
		fmt.Println("Failed to convert string to int:", err)
		return false
	}

	contents, err = ioutil.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		fmt.Println("Failed to read pid cmdline:", err)
		return false
	}

	return strings.Contains(string(contents), "ya-goproxy")
}
