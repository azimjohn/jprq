package jprq

import (
	"github.com/go-errors/errors"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"gopkg.in/mgo.v2/bson"
)

type Tunnel struct {
	host           string
	port           int
	conn           *websocket.Conn
	token          string
	requests       map[uuid.UUID]RequestMessage
	requestChan    chan RequestMessage
	responseChan   chan ResponseMessage
	numOfReqServed int
}

func (j Jprq) GetTunnelByHost(host string) (*Tunnel, error) {
	t, ok := j.tunnels[host]
	if !ok {
		return t, errors.New("Tunnel doesn't exist")
	}

	return t, nil
}

func (j *Jprq) AddTunnel(host string, port int, conn *websocket.Conn) *Tunnel {
	token := generateToken()
	requests := make(map[uuid.UUID]RequestMessage)
	requestChan, responseChan := make(chan RequestMessage), make(chan ResponseMessage)
	tunnel := Tunnel{
		host:         host,
		port:         port,
		conn:         conn,
		token:        token,
		requests:     requests,
		requestChan:  requestChan,
		responseChan: responseChan,
	}

	log.Info("New Tunnel: ", host)
	j.tunnels[host] = &tunnel
	return &tunnel
}

func (j *Jprq) DeleteTunnel(host string) {
	tunnel, ok := j.tunnels[host]
	if !ok {
		return
	}
	log.Infof("Deleted Tunnel: %s, Number Of Requests Served: %d", host, tunnel.numOfReqServed)
	close(tunnel.requestChan)
	close(tunnel.responseChan)
	delete(j.tunnels, host)
}

func (tunnel *Tunnel) DispatchRequests() {
	for {
		select {
		case requestMessage, more := <-tunnel.requestChan:
			if !more {
				return
			}
			messageContent, _ := bson.Marshal(requestMessage)
			tunnel.requests[requestMessage.ID] = requestMessage
			tunnel.conn.WriteMessage(websocket.BinaryMessage, messageContent)
		}
	}
}

func (tunnel *Tunnel) DispatchResponses() {
	for {
		select {
		case responseMessage, more := <-tunnel.responseChan:
			if !more {
				return
			}
			requestMessage, ok := tunnel.requests[responseMessage.RequestId]
			if !ok {
				log.Error("Request Not Found", responseMessage.RequestId)
				continue
			}

			requestMessage.ResponseChan <- responseMessage
			delete(tunnel.requests, requestMessage.ID)
			tunnel.numOfReqServed++
		}
	}
}
