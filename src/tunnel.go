package main

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/gorilla/websocket"
	"github.com/gosimple/slug"
	"github.com/labstack/gommon/log"
)

var tunnels map[string]Tunnel

type Tunnel struct {
	host  string
	port  int
	conn  *websocket.Conn
	token string
}

func GetTunnelByHost(host string) (Tunnel, error) {
	t, ok := tunnels[host]
	if !ok {
		return t, errors.New("Tunnel doesn't exist")
	}

	return t, nil
}

func AddTunnel(username string, port int, conn *websocket.Conn) Tunnel {
	adj := GetRandomAdj()
	username = slug.Make(username)

	host := fmt.Sprintf("%s-%s.%s", adj, username, config.BaseHostName)
	token, _ := GetJWToken(host)
	tunnel := Tunnel{host, port, conn, token}

	log.Info("New Tunnel: ", host)
	tunnels[host] = tunnel
	return tunnel
}

func DeleteTunnel(host string) {
	log.Info("Deleted Tunnel: ", host)
	delete(tunnels, host)
}
