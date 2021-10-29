package jprq_tcp

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (j Jprq) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	defer ws.Close()
	if err != nil {
		return
	}

	tunnel, err := j.OpenTunnel(ws)
	defer tunnel.Close()

	for {
		_, _, closedErr := ws.ReadMessage()
		if closedErr != nil {
			return
		}
	}
}
