package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/azimjohn/jprq/jprq_tcp"
	"github.com/gorilla/websocket"
	"log"
	"net"
	urlpkg "net/url"
)

type TCPTunnel struct {
	PublicServerPort  int `json:"public_server_port"`
	PrivateServerPort int `json:"private_server_port"`
}

type ConnectionRequest struct {
	IP   string `json:"public_client_ip"`
	Flag int    `json:"public_client_port"`
}

func openTCPTunnel(port int, host string, ctx context.Context) {
	url := urlpkg.URL{Scheme: "wss", Host: host, Path: "/_ws/"}
	ws, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatalf("Error Connecting to %s: %s\n", host, err.Error())
	}
	defer ws.Close()

	var tunnel TCPTunnel
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Fatalf("Error Reading Message from Server: %s\n", err.Error())
	}

	err = json.Unmarshal(message, &tunnel)
	if err != nil {
		log.Fatalf("Error Decoding Tunnel Info: %s\n", err.Error())
	}

	fmt.Println("\033[32mTunnel Status: \t\tOnline\033[00m")
	fmt.Printf("Forwarded:\t\t%s:%d → 127.0.0.1:%d\n\n",
		host, tunnel.PublicServerPort, port)

	connRequests := make(chan ConnectionRequest)
	defer close(connRequests)

	go handleTCPConnections(ws, connRequests)

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case connRequest := <-connRequests:
			go handleTCPConnection(connRequest, host, port, tunnel.PrivateServerPort)
		}
	}

	fmt.Println("\n\033[31mjprq tunnel closed\033[00m")
}

func handleTCPConnections(ws *websocket.Conn, connRequests chan<- ConnectionRequest) {
	for {
		var connReq ConnectionRequest
		_, message, err := ws.ReadMessage()
		if err != nil {
			return
		}
		err = json.Unmarshal(message, &connReq)
		if err != nil {
			log.Printf("Error Decoding Message %s\n", message)
		}
		connRequests <- connReq
	}
}

func handleTCPConnection(connRequest ConnectionRequest, host string, localServerPort int, remoteServerPort int) {
	fmt.Printf("> Opened Connection with %s\n", connRequest.IP)
	defer fmt.Printf("> Closed Connection with %s\n", connRequest.IP)

	localConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", localServerPort))
	if err != nil {
		fmt.Printf("Error Connecting to Local Server: %s\n", err.Error())
		return
	}
	defer localConn.Close()

	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, remoteServerPort))
	if err != nil {
		fmt.Printf("Error Connecting to Remote Server: %s\n", err.Error())
		return
	}
	defer remoteConn.Close()

	flag := connRequest.Flag
	_, err = remoteConn.Write([]byte{byte(flag >> 8 & 0xFF), byte(flag & 0xFF)})
	if err != nil {
		fmt.Printf("Error Sending Data to Remote Server: %s\n", err.Error())
		return
	}

	jprq_tcp.BindTCPConnections(&localConn, &remoteConn)
}
