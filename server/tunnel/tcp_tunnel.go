package tunnel

import (
	"fmt"
	"github.com/azimjohn/jprq/server/server"
	"net"
)

type TCPTunnel struct {
	hostname       string
	publicServer   server.TCPServer
	privateServer  server.TCPServer
	privateCons    map[uint16]net.Conn
	publicCons     map[uint16]net.Conn
	publicConsChan chan net.Conn
}

func NewTCP(hostname string) (*TCPTunnel, error) {
	t := &TCPTunnel{
		hostname:       hostname,
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

func (t *TCPTunnel) PublicConnections() <-chan net.Conn {
	return t.publicConsChan
}

func (t *TCPTunnel) Open() error {
	go t.privateServer.Start()
	go t.publicServer.Start()

	// handle private and public servers
	return nil
}

func (t *TCPTunnel) Close() error {
	close(t.publicConsChan)

	if err := t.privateServer.Stop(); err != nil {
		return err
	}
	if err := t.publicServer.Stop(); err != nil {
		return err
	}
	return nil
}
