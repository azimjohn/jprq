package tunnel

import (
	"fmt"
	"github.com/azimjohn/jprq/server/server"
	"io"
	"net"
)

type TCPTunnel struct {
	hostname       string
	eventWriter    io.Writer
	publicServer   server.TCPServer
	privateServer  server.TCPServer
	privateCons    map[uint16]net.Conn
	publicCons     map[uint16]net.Conn
	publicConsChan chan net.Conn
}

func NewTCP(hostname string, eventWriter io.Writer) (*TCPTunnel, error) {
	t := &TCPTunnel{
		hostname:       hostname,
		eventWriter:    eventWriter,
		publicCons:     make(map[uint16]net.Conn),
		privateCons:    make(map[uint16]net.Conn),
		publicConsChan: make(chan net.Conn),
	}
	if err := t.privateServer.Init(0); err != nil {
		return t, fmt.Errorf("error init private server: %w", err)
	}
	if err := t.publicServer.Init(0); err != nil {
		return t, fmt.Errorf("error init public server: %w", err)
	}
	return t, nil
}

func (t *TCPTunnel) Protocol() string {
	return "tcp"
}

func (t *TCPTunnel) Hostname() string {
	return t.hostname
}

func (t *TCPTunnel) PublicServerPort() uint16 {
	return t.publicServer.Port()
}

func (t *TCPTunnel) PrivateServerPort() uint16 {
	return t.privateServer.Port()
}

func (t *TCPTunnel) Open() {
	go t.privateServer.Start()
	go t.publicServer.Start()

	// handle private and public connections
}

func (t *TCPTunnel) Close() {
	close(t.publicConsChan)
	t.privateServer.Stop()
	t.publicServer.Stop()
}
