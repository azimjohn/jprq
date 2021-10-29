package jprq_tcp

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net"
	"sync"
)

type Tunnel struct {
	conn                 *websocket.Conn
	publicServer         *net.Listener
	privateServer        *net.Listener
	publicPrivateMap     map[int]int
	publicConnections    map[int]*net.Conn
	privateConnections   map[int]*net.Conn
	publicConnectionChan chan *net.Conn
}

func (t *Tunnel) Close() {
	(*t.privateServer).Close()
	(*t.publicServer).Close()
	close(t.publicConnectionChan)
}

func (t *Tunnel) SendTunnelStartedEvent() {
	event := TunnelStartedEvent{
		PublicServerPort:  t.PublicServerPort(),
		PrivateServerPort: t.PrivateServerPort(),
	}
	message, _ := json.Marshal(event)
	t.conn.WriteMessage(websocket.TextMessage, message)
}

func (t *Tunnel) PublicServerPort() int {
	return (*t.publicServer).Addr().(*net.TCPAddr).Port
}

func (t *Tunnel) PrivateServerPort() int {
	return (*t.privateServer).Addr().(*net.TCPAddr).Port
}

func (t *Tunnel) AcceptPublicConnections() {
	for {
		c, err := (*t.publicServer).Accept()
		if err != nil {
			break
		}
		port := c.RemoteAddr().(*net.TCPAddr).Port
		t.publicConnections[port] = &c
		t.publicConnectionChan <- &c
	}
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
			log.Errorf("error reading from private client: %s", err)
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
		port := (*conn).RemoteAddr().(*net.TCPAddr).Port
		event := ConnectionReceivedEvent{port}
		message, _ := json.Marshal(event)
		t.conn.WriteMessage(websocket.TextMessage, message)
	}
}

func (t *Tunnel) PairConnections(publicClientPort, privateClientPort int) {
	t.publicPrivateMap[publicClientPort] = privateClientPort
	defer delete(t.publicPrivateMap, publicClientPort)
	defer delete(t.publicConnections, publicClientPort)
	defer delete(t.privateConnections, privateClientPort)

	publicClient, found1 := t.publicConnections[publicClientPort]
	privateClient, found2 := t.privateConnections[privateClientPort]

	if !found1 || !found2 {
		log.Error("connection not found from connections map")
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go PumpReadToWrite(publicClient, privateClient, &wg)
	go PumpReadToWrite(privateClient, publicClient, &wg)

	wg.Wait()
}
