package tunnel

import (
	"errors"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/server"
	"io"
	"net"
)

const DefaultHttpPort = 80

type HTTPTunnel struct {
	hostname       string
	maxConsLimit   int
	eventWriter    io.Writer
	publicConsChan chan net.Conn
	privateServer  server.TCPServer
	initialBuffer  map[uint16][]byte
	privateCons    map[uint16]net.Conn
	publicCons     map[uint16]net.Conn
}

func NewHTTP(hostname string, eventWriter io.Writer, maxConsLimit int) (*HTTPTunnel, error) {
	t := &HTTPTunnel{
		hostname:       hostname,
		eventWriter:    eventWriter,
		maxConsLimit:   maxConsLimit,
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

func (t *HTTPTunnel) Open() {
	go t.privateServer.Start()

	// todo: handle private connections
}

func (t *HTTPTunnel) Close() {
	t.privateServer.Stop()
	close(t.publicConsChan)
}

func (t *HTTPTunnel) PublicConnectionHandler(conn net.Conn, initialBuffer []byte) error {
	ip := conn.RemoteAddr().(*net.TCPAddr).IP
	port := uint16(conn.RemoteAddr().(*net.TCPAddr).Port)

	if len(t.publicCons) >= t.maxConsLimit {
		event := events.Event[events.ConnectionReceived]{
			Data: &events.ConnectionReceived{
				ClientIP:    ip,
				RateLimited: true,
			},
		}
		conn.Close()
		event.Write(t.eventWriter)
		return errors.New("connection rate limited")
	}

	event := events.Event[events.ConnectionReceived]{
		Data: &events.ConnectionReceived{
			ClientIP:    ip,
			ClientPort:  port,
			RateLimited: false,
		},
	}
	if err := event.Write(t.eventWriter); err != nil {
		return conn.Close()
	}
	t.publicCons[port] = conn
	t.initialBuffer[port] = initialBuffer
	return nil
}
