package jprq

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
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

func (j Jprq) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	query := r.URL.Query()
	usernames := query["username"]
	ports := query["port"]

	if len(usernames) != 1 || len(ports) != 1 {
		log.Error("Websocket Connection: Bad Request: ", query)
		return
	}

	username := usernames[0]
	port, _ := strconv.Atoi(ports[0])

	tunnel := j.AddTunnel(username, port, ws)
	defer j.DeleteTunnel(tunnel.host)

	message := TunnelMessage{tunnel.host, tunnel.token}
	messageContent, err := json.Marshal(message)
	time.Sleep(time.Second)

	ws.WriteMessage(websocket.TextMessage, messageContent)

	go tunnel.DispatchRequests()
	go tunnel.DispatchResponses()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}

		response := ResponseMessage{}
		err = json.Unmarshal(message, &response)
		if err != nil {
			log.Error("Failed to Unmarshal Websocket Message: ", string(message), err)
			continue
		}

		if response.Token != tunnel.token {
			log.Error("Authentication Failed: ", tunnel.host)
			continue
		}

		tunnel.responseChan <- response
	}
}
