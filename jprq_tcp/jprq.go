package jprq_tcp

import (
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net"
)

type Jprq struct {
	baseHost string
}

func New(baseHost string) Jprq {
	return Jprq{
		baseHost: baseHost,
	}
}

func (j *Jprq) OpenTunnel(conn *websocket.Conn) (*Tunnel, error) {
	log.Infof("New TCP Tunnel from IP %s", conn.RemoteAddr())
	publicServer, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	privateServer, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	tunnel := &Tunnel{
		conn:                 conn,
		publicServer:         &publicServer,
		privateServer:        &privateServer,
		publicPrivateMap:     make(map[int]int),
		publicConnections:    make(map[int]*net.Conn),
		privateConnections:   make(map[int]*net.Conn),
		publicConnectionChan: make(chan *net.Conn),
	}

	go tunnel.SendTunnelStartedEvent()
	go tunnel.AcceptPublicConnections()
	go tunnel.AcceptPrivateConnections()
	go tunnel.NotifyPublicConnections()

	return tunnel, nil
}
