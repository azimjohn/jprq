package main

import (
	"fmt"
	"github.com/azimjohn/jprq/server/events"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var version = "2.0"
var authUrl = "https://jprq.io/auth"

func main() {
	log.SetFlags(0)

	port := 3000
	subdomain := ""
	protocol := events.HTTP

	if !canReachServer(port) {
		log.Fatalf("error: can't reach server on port: %d\n", port)
	}

	var conf Config
	if err := conf.Load(); err != nil {
		log.Fatal(err)
	}
	client := jprqClient{
		config:    conf,
		protocol:  protocol,
		subdomain: subdomain,
	}

	go client.Start(port)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}

func canReachServer(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
