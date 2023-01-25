package tunnel

import (
	"fmt"
	"github.com/azimjohn/jprq/server/server"
)

type TCPTunnel struct {
	hostname      string
	publicServer  server.TCPServer
	privateServer server.TCPServer
}

func NewTCPTunnel(hostname string) (TCPTunnel, error) {
	var t TCPTunnel
	t.hostname = hostname
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

func (t *TCPTunnel) Start() {
	go t.privateServer.Start()
	go t.publicServer.Start()

	// handle private and public servers
}
