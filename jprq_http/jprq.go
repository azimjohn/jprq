package jprq_http

import (
	"errors"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net"
	"strings"
)

type Jprq struct {
	publicServer  *net.Listener
	tunnelsByHost map[string]*Tunnel
}

func New() (*Jprq, error) {
	publicServer, err := net.Listen("tcp", ":8081")
	if err != nil {
		return nil, err
	}
	jprq := &Jprq{
		publicServer:  &publicServer,
		tunnelsByHost: make(map[string]*Tunnel),
	}
	go jprq.AcceptPublicConnections()

	return jprq, nil
}

func (j *Jprq) OpenTunnel(hostname string, conn *websocket.Conn) (*Tunnel, error) {
	if _, ok := j.tunnelsByHost[hostname]; ok {
		return nil, errors.New("hostname busy")
	}

	log.Infof("New HTTP Tunnel %s", hostname)
	privateServer, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}

	tunnel := &Tunnel{
		conn:                 conn,
		hostname:             hostname,
		privateServer:        &privateServer,
		publicPrivateMap:     make(map[int]int),
		publicConnections:    make(map[int]*net.Conn),
		privateConnections:   make(map[int]*net.Conn),
		publicConnectionChan: make(chan *net.Conn),
		initialBufferByPort:  make(map[int][]byte),
	}
	go tunnel.SendTunnelStartedEvent()
	go tunnel.AcceptPrivateConnections()
	go tunnel.NotifyPublicConnections()
	j.tunnelsByHost[hostname] = tunnel

	return tunnel, nil
}

func (j Jprq) AcceptPublicConnections() {
	for {
		conn, err := (*j.publicServer).Accept()
		if err != nil {
			log.Error("error accepting public connection: ", err.Error())
			continue
		}
		err = j.HandlePublicConnection(conn)
		if err != nil {
			log.Error("error handling public connection: ", err.Error())
			continue
		}
	}
}

func readLine(conn net.Conn) (string, error) {
	// todo: optimize & refactor
	var line []byte
	buffer := make([]byte, 1)
	for {
		if _, err := conn.Read(buffer); err != nil {
			return "", err
		}
		line = append(line, buffer...)
		if string(buffer) == "\n" {
			break
		}
	}
	return string(line), nil
}

func (j Jprq) HandlePublicConnection(conn net.Conn) error {
	firstLine, err := readLine(conn)
	if err != nil {
		return err
	}
	secondLine, err := readLine(conn)
	if err != nil {
		return err
	}

	i := strings.Index(secondLine, ":")
	if i < 0 {
		return errors.New("malformed http request")
	}
	i++
	host := strings.Trim(secondLine[i:], "\r\n ")
	host = strings.ToLower(host)

	tunnel := j.tunnelsByHost[host]
	if tunnel == nil {
		return errors.New("tunnel not found " + host)
	}

	port := conn.RemoteAddr().(*net.TCPAddr).Port
	tunnel.initialBufferByPort[port] = []byte(firstLine + secondLine)
	tunnel.publicConnections[port] = &conn
	tunnel.publicConnectionChan <- &conn

	return nil
}

func (j Jprq) CloseTunnel(tunnel *Tunnel) {
	tunnel.Close()
	delete(j.tunnelsByHost, tunnel.hostname)
}
