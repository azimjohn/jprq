package tunnel

import (
	"fmt"
	"github.com/azimjohn/jprq/server/server"
	"io"
)

type TCPTunnel struct {
	tunnel
	publicServer server.TCPServer
}

func NewTCP(hostname string, eventWriter io.Writer, maxConsLimit int) (*TCPTunnel, error) {
	t := &TCPTunnel{
		tunnel: newTunnel(hostname, eventWriter, maxConsLimit),
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

func (t *TCPTunnel) PublicServerPort() uint16 {
	return t.publicServer.Port()
}

func (t *TCPTunnel) Open() {
	go t.publicServer.Start(t.publicConnectionHandler)
	go t.privateServer.Start(t.privateConnectionHandler)
}

func (t *TCPTunnel) Close() {
	t.publicServer.Stop()
	t.tunnel.Close()
}
