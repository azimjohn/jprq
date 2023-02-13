package tunnel

import (
	"io"
	"net"
)

const DefaultHttpPort = 80

type HTTPTunnel struct {
	tunnel
}

func NewHTTP(hostname string, eventWriter io.Writer, maxConsLimit int) (*HTTPTunnel, error) {
	t := &HTTPTunnel{
		tunnel: newTunnel(hostname, eventWriter, maxConsLimit),
	}
	if err := t.privateServer.Init(0, "http-tunnel-private-server"); err != nil {
		return t, err
	}
	return t, nil
}

func (t *HTTPTunnel) Protocol() string {
	return "http"
}

func (t *HTTPTunnel) PublicServerPort() uint16 {
	return DefaultHttpPort
}

func (t *HTTPTunnel) Open() {
	go t.privateServer.Start(t.privateConnectionHandler)
}

func (t *HTTPTunnel) PublicConnectionHandler(publicCon net.Conn, initialBuffer []byte) error {
	port := uint16(publicCon.RemoteAddr().(*net.TCPAddr).Port)
	t.initialBuffer[port] = initialBuffer
	return t.publicConnectionHandler(publicCon)
}
