package tunnel

import (
	"github.com/azimjohn/jprq/server/server"
	"net"
)

const DefaultHttpPort = 80

type HTTPTunnel struct {
	hostname       string
	privateServer  server.TCPServer
	initialBuffer  map[uint16][]byte
	privateCons    map[uint16]net.Conn
	publicCons     map[uint16]net.Conn
	publicConsChan chan net.Conn
}

func NewHTTP(hostname string) (*HTTPTunnel, error) {
	t := &HTTPTunnel{
		hostname:       hostname,
		publicCons:     make(map[uint16]net.Conn),
		privateCons:    make(map[uint16]net.Conn),
		publicConsChan: make(chan net.Conn),
	}
	t.hostname = hostname
	if err := t.privateServer.Init(0); err != nil {
		return t, err
	}
	return t, nil
}

func (t *HTTPTunnel) Protocol() string {
	return "http"
}

func (t *HTTPTunnel) Hostname() string {
	return t.hostname
}

func (t *HTTPTunnel) PrivateServerPort() uint16 {
	return t.privateServer.Port()
}

func (t *HTTPTunnel) PublicServerPort() uint16 {
	return DefaultHttpPort
}

func (t *HTTPTunnel) PublicConnections() chan<- net.Conn {
	return t.publicConsChan
}

func (t *HTTPTunnel) Open() error {
	go t.privateServer.Start()
	return nil
}

func (t *HTTPTunnel) Close() error {
	if err := t.privateServer.Stop(); err != nil {
		return err
	}
	close(t.publicConsChan)
	return nil
}

func (t *HTTPTunnel) PublicConnectionHandler(conn net.Conn, initialBuffer []byte) {
	port := uint16(conn.RemoteAddr().(*net.TCPAddr).Port)
	t.publicCons[port] = conn
	t.initialBuffer[port] = initialBuffer
}
