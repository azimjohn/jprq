package jprq_http

import (
	"encoding/json"
	"github.com/azimjohn/jprq/jprq_tcp"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net"
)

type Tunnel struct {
	hostname             string
	conn                 *websocket.Conn
	privateServer        *net.Listener
	publicPrivateMap     map[int]int
	publicConnections    map[int]*net.Conn
	privateConnections   map[int]*net.Conn
	initialBufferByPort  map[int][]byte
	publicConnectionChan chan *net.Conn
}

func (t *Tunnel) Close() {
	(*t.privateServer).Close()
	close(t.publicConnectionChan)
	log.Infof("HTTP Tunnel Closed from IP %s", t.conn.RemoteAddr())
}

func (t *Tunnel) SendTunnelStartedEvent() {
	event := jprq_tcp.TunnelStartedEvent{
		PublicServerPort:  8081,
		PrivateServerPort: t.PrivateServerPort(),
	}
	message, _ := json.Marshal(event)
	t.conn.WriteMessage(websocket.TextMessage, message)
}

func (t *Tunnel) PrivateServerPort() int {
	return (*t.privateServer).Addr().(*net.TCPAddr).Port
}

func (t *Tunnel) AcceptPrivateConnections() {
	for {
		c, err := (*t.privateServer).Accept()
		if err != nil {
			break
		}
		privateClientPort := c.RemoteAddr().(*net.TCPAddr).Port
		t.privateConnections[privateClientPort] = &c

		buffer := make([]byte, 2) // 16 bits
		_, err = c.Read(buffer)
		if err != nil {
			log.Errorf("reading from private client failed: %s\n", err)
			return
		}

		publicClientPort := (int(buffer[0]) << 8) + int(buffer[1])
		go t.PairConnections(publicClientPort, privateClientPort)
	}
}

func (t *Tunnel) NotifyPublicConnections() {
	for {
		conn, ok := <-t.publicConnectionChan
		if !ok {
			break
		}
		ip := (*conn).RemoteAddr().(*net.TCPAddr).IP.String()
		port := (*conn).RemoteAddr().(*net.TCPAddr).Port
		event := jprq_tcp.ConnectionReceivedEvent{
			PublicClientPort: port, PublicClientIP: ip,
		}
		message, _ := json.Marshal(event)
		t.conn.WriteMessage(websocket.TextMessage, message)
	}
}

func (t *Tunnel) PairConnections(publicClientPort, privateClientPort int) {
	defer delete(t.publicPrivateMap, publicClientPort)
	defer delete(t.publicConnections, publicClientPort)
	defer delete(t.privateConnections, privateClientPort)

	t.publicPrivateMap[publicClientPort] = privateClientPort
	publicClient, found1 := t.publicConnections[publicClientPort]
	privateClient, found2 := t.privateConnections[privateClientPort]

	if !found1 || !found2 {
		log.Error("connection not found from connections map")
		return
	}

	buffer, ok := t.initialBufferByPort[publicClientPort]
	if !ok {
		log.Error("initial buffer not found")
		return
	}

	_, err := (*privateClient).Write(buffer)
	if err != nil {
		log.Error("writing initial buffer to private client failed")
		return
	}

	delete(t.initialBufferByPort, publicClientPort)
	jprq_tcp.BindTCPConnections(publicClient, privateClient)
}
