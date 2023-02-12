package tunnel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/server"
	"io"
	"net"
)

type TCPTunnel struct {
	hostname      string
	maxConsLimit  int
	eventWriter   io.Writer
	publicServer  server.TCPServer
	privateServer server.TCPServer
	publicCons    map[uint16]net.Conn
}

func NewTCP(hostname string, eventWriter io.Writer, maxConsLimit int) (*TCPTunnel, error) {
	t := &TCPTunnel{
		hostname:     hostname,
		eventWriter:  eventWriter,
		maxConsLimit: maxConsLimit,
		publicCons:   make(map[uint16]net.Conn),
	}
	if err := t.privateServer.Init(0, "tcp_tunnel_private_server"); err != nil {
		return t, fmt.Errorf("error init private server: %w", err)
	}
	if err := t.publicServer.Init(0, "tcp_tunnel_public_server"); err != nil {
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
	go t.publicServer.Start(func(publicCon net.Conn) error {
		ip := publicCon.RemoteAddr().(*net.TCPAddr).IP
		port := uint16(publicCon.RemoteAddr().(*net.TCPAddr).Port)

		if len(t.publicCons) >= t.maxConsLimit {
			event := events.Event[events.ConnectionReceived]{
				Data: &events.ConnectionReceived{
					ClientIP:    ip,
					RateLimited: true,
				},
			}
			publicCon.Close()
			return event.Write(t.eventWriter)
		}

		event := events.Event[events.ConnectionReceived]{
			Data: &events.ConnectionReceived{
				ClientIP:    ip,
				ClientPort:  port,
				RateLimited: false,
			},
		}
		if err := event.Write(t.eventWriter); err != nil {
			return publicCon.Close()
		}
		t.publicCons[port] = publicCon
		return nil
	})

	go t.privateServer.Start(func(privateCon net.Conn) error {
		defer privateCon.Close()
		buffer := make([]byte, 2)
		if _, err := privateCon.Read(buffer); err != nil {
			return err
		}
		port := binary.LittleEndian.Uint16(buffer)
		publicCon, found := t.publicCons[port]
		if !found {
			return errors.New("public connection not found")
		}

		defer publicCon.Close()
		delete(t.publicCons, port)

		go Bind(publicCon, privateCon)
		Bind(privateCon, publicCon)
		return nil
	})
}

func (t *TCPTunnel) Close() {
	t.privateServer.Stop()
	t.publicServer.Stop()
	for port, con := range t.publicCons {
		con.Close()
		delete(t.publicCons, port)
	}
}
