package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	query := r.URL.Query()
	usernames := query["username"]
	ports := query["port"]

	if len(usernames) != 1 || len(ports) != 1 {
		return
	}

	username := usernames[0]
	port, _ := strconv.Atoi(ports[0])

	tunnel := AddTunnel(username, port, ws)
	defer DeleteTunnel(tunnel.host)
	message := Message{
		tunnelCreated, TunnelCreatedMessage{tunnel.host, tunnel.token},
	}

	time.Sleep(time.Second * 5)
	ws.WriteMessage(websocket.TextMessage, message.Bytes())

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		// todo: receive request and return http response to client
		fmt.Println(string(message))
	}
}
