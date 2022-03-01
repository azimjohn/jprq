package jprq_http

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/gommon/log"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func buildErrMessage(msg string) []byte {
	return []byte(fmt.Sprintf(`{"error": "%s"`, msg))
}

func (j Jprq) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	defer ws.Close()
	if err != nil {
		return
	}

	query := r.URL.Query()
	versions := query["version"]
	hostnames := query["hostname"]

	if len(hostnames) != 1 || len(versions) != 1 {
		ws.WriteMessage(websocket.TextMessage, buildErrMessage("missing params"))
		log.Error("bad request: ", query)
		return
	}

	hostname := strings.ToLower(hostnames[0])
	tunnel, err := j.OpenTunnel(hostname, ws)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, buildErrMessage(err.Error()))
		log.Error("could not open tunnel: ", err.Error())
		return
	}
	defer j.CloseTunnel(tunnel)

	for {
		_, _, closedErr := ws.ReadMessage()
		if closedErr != nil {
			return
		}
	}
}
