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
	"sync"
)

type TCPTunnel struct {
	PublicServerPort  int `json:"public_server_port"`
	PrivateServerPort int `json:"private_server_port"`
}

type ConnectionRequest struct {
	Flag int `json:"public_client_port"`
}

func openTCPTunnel(port int, ctx context.Context) {
	url := urlpkg.URL{Scheme: "wss", Host: tcpBaseHost, Path: "/_ws/"}
	ws, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatalf("Error Connecting to %s: %s\n", tcpBaseHost, err.Error())
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
		tcpBaseHost, tunnel.PublicServerPort, port)

	connRequests := make(chan int)
	defer close(connRequests)

	go receiveTCPConnectionRequests(ws, connRequests)

	out:
	for {
		select {
		case <-ctx.Done():
			break out
		case flag := <-connRequests:
			go handleTCPConnection(flag, port, tunnel.PrivateServerPort, ctx)
		}
	}

	fmt.Println("\n\033[31mjprq tunnel closed\033[00m")
}

func receiveTCPConnectionRequests(ws *websocket.Conn, connRequests chan<- int) {
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
		connRequests <- connReq.Flag
	}
}

func handleTCPConnection(flag, localServerPort, remoteServerPort int, ctx context.Context) {
	fmt.Println("New Connection +1")
	defer fmt.Println("Connection Closed")

	localConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", localServerPort))
	if err != nil {
		fmt.Printf("Error Connecting to Local Server: %s\n", err.Error())
		return
	}
	defer localConn.Close()

	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", tcpBaseHost, remoteServerPort))
	if err != nil {
		fmt.Printf("Error Connecting to Remote Server: %s\n", err.Error())
		return
	}
	defer remoteConn.Close()

	_, err = remoteConn.Write([]byte{byte(flag >> 8 & 0xFF), byte(flag & 0xFF)})
	if err != nil {
		fmt.Printf("Error Sending Data to Remote Server: %s\n", err.Error())
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go jprq_tcp.PumpReadToWrite(&remoteConn, &localConn, &wg)
	go jprq_tcp.PumpReadToWrite(&localConn, &remoteConn, &wg)

	wg.Wait()
}
