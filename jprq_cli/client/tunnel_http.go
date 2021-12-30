package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/azimjohn/jprq/jprq_http"
	"github.com/gorilla/websocket"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	urlpkg "net/url"
	"os/user"
)

type HTTPTunnel struct {
	Host    string `bson:"host"`
	Token   string `bson:"token"`
	Warning string `bson:"warning"`
	Error   string `bson:"error"`
}

func openHTTPTunnel(port int, host string, subdomain string, ctx context.Context) {
	if subdomain == "" {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("Please specify -subdomain")
		}
		subdomain = u.Username
	}

	query := fmt.Sprintf("port=%d&username=%s&version=%s", port, subdomain, version)
	url := urlpkg.URL{Scheme: "wss", Host: host, Path: "/_ws/", RawQuery: query}

	ws, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Fatalf("Error Connecting to %s: %s\n", host, err.Error())
	}
	defer ws.Close()

	var tunnel HTTPTunnel
	_, message, err := ws.ReadMessage()
	if err != nil {
		log.Fatalf("Error Reading Message from Server: %s\n", err.Error())
	}

	err = bson.Unmarshal(message, &tunnel)
	if err != nil {
		log.Fatalf("Error Decoding Tunnel Info: %s\n", err.Error())
	}

	if tunnel.Warning != "" {
		fmt.Printf("WARNING: %s", tunnel.Warning)
	}

	if tunnel.Error != "" {
		log.Fatal(tunnel.Error)
	}

	fmt.Println("\033[32mTunnel Status: \t\tOnline\033[00m")
	fmt.Printf("Forwarded:\t\t%s → 127.0.0.1:%d\n\n", tunnel.Host, port)

	requests := make(chan jprq_http.RequestMessage)
	defer close(requests)

	go handleHTTPRequests(ws, requests)

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case request := <-requests:
			go handleHTTPRequest(ws, tunnel.Token, port, request)
		}
	}

	fmt.Println("\n\033[31mjprq tunnel closed\033[00m")
}

func handleHTTPRequests(ws *websocket.Conn, requests chan<- jprq_http.RequestMessage) {
	for {
		var requestMessage jprq_http.RequestMessage
		_, message, err := ws.ReadMessage()
		if err != nil {
			return
		}
		err = bson.Unmarshal(message, &requestMessage)
		if err != nil {
			log.Printf("Error Decoding Message %s\n", message)
		}
		requests <- requestMessage
	}
}

func handleHTTPRequest(ws *websocket.Conn, token string, port int, r jprq_http.RequestMessage) {
	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, r.URL)
	request, err := http.NewRequest(r.Method, url, bytes.NewReader(r.Body))
	if err != nil {
		fmt.Printf("Failed to Build Request: %s\n", err.Error())
		return
	}

	for key, val := range r.Header {
		request.Header.Add(key, val)
	}

	var client http.Client
	response, err := client.Do(request)

	if err != nil {
		fmt.Printf("Failed to Perform Request: %s\n", err.Error())
		return
	}

	responseMessage := jprq_http.ResponseMessage{}

	responseMessage.Header = make(map[string]string)
	for name, values := range response.Header {
		responseMessage.Header[name] = values[0]
	}

	if response.Body != nil {
		responseMessage.Body, _ = ioutil.ReadAll(response.Body)
		response.Body.Close()
	}

	responseMessage.Status = response.StatusCode
	responseMessage.RequestId = r.ID
	responseMessage.Token = token

	message, err := bson.Marshal(responseMessage)
	if err != nil {
		fmt.Printf("Error Encoding Response Message: %s\n", err.Error())
		return
	}

	err = ws.WriteMessage(websocket.BinaryMessage, message)
	if err != nil {
		fmt.Printf("Error Sending Message to Server: %s", err)
		return
	}

	fmt.Println(r.Method, r.URL, responseMessage.Status)
}
