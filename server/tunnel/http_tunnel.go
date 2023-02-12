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

const DefaultHttpPort = 80

type HTTPTunnel struct {
	hostname      string
	maxConsLimit  int
	eventWriter   io.Writer
	privateServer server.TCPServer
	initialBuffer map[uint16][]byte
	publicCons    map[uint16]net.Conn
}

func NewHTTP(hostname string, eventWriter io.Writer, maxConsLimit int) (*HTTPTunnel, error) {
	t := &HTTPTunnel{
		hostname:      hostname,
		eventWriter:   eventWriter,
		maxConsLimit:  maxConsLimit,
		initialBuffer: make(map[uint16][]byte),
		publicCons:    make(map[uint16]net.Conn),
	}
	t.hostname = hostname
	if err := t.privateServer.Init(0, "http-tunnel-private-server"); err != nil {
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
	go t.privateServer.Serve(func(privateCon net.Conn) error {
		defer privateCon.Close()
		buffer := make([]byte, 2)
		if _, err := privateCon.Read(buffer); err != nil {
			return err
		}
		port := binary.LittleEndian.Uint16(buffer)
		publicCon, found := t.publicCons[port]
		if !found {
			return errors.New("public connection not found, cannot pair")
		}
		defer publicCon.Close()
		delete(t.publicCons, port)
		defer delete(t.initialBuffer, port)
		if _, err := privateCon.Write(t.initialBuffer[port]); err != nil {
			return err
		}

		go io.Copy(publicCon, privateCon)
		io.Copy(privateCon, publicCon)
		return nil
	})
}

func (t *HTTPTunnel) PublicConnectionHandler(publicCon net.Conn, initialBuffer []byte) error {
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
		event.Write(t.eventWriter)
		return errors.New(fmt.Sprintf("[connections-limit-reached]: %s", t.hostname))
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
	t.initialBuffer[port] = initialBuffer
	return nil
}

func (t *HTTPTunnel) Close() {
	t.privateServer.Stop()
	for port, con := range t.publicCons {
		con.Close()
		delete(t.publicCons, port)
	}
}
