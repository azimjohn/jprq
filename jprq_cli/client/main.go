package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var version = "1.0.0"
var tcpBaseHost = "tcp.jprq.io"
var httpBaseHost = "open.jprq.io"

func main() {
	subdomain := flag.String("subdomain", "", "Subdomain for HTTP Tunnel")
	flag.Parse()
	log.SetFlags(0)

	if len(os.Args) < 3 {
		log.Fatalf("Usage: jprq <PROTOCOL> <PORT>\n" +
			"  Supported Protocols: [tcp, http]\n" +
			"  Optional Argument: -subdomain\n" +
			"  Client Version: %s\n", version)
	}

	protocol := os.Args[1]
	if protocol != "tcp" && protocol != "http" {
		log.Fatalf("Invalid Protocol: %s\n", protocol)
	}

	port, err := strconv.Atoi(os.Args[2])
	if err != nil || port < 0 || port > 65535 {
		log.Fatalf("Invalid Port Number: %d\n", port)
	}

	if !canReachServer(port) {
		log.Fatalf("No server is running on port: %d\n", port)
	}

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	if os.Args[1] == "tcp" {
		go openTCPTunnel(port, ctx)
	} else if os.Args[1] == "http" {
		go openHTTPTunnel(port, *subdomain, ctx)
	}

	<-signalChan
	cancelFunc()
}

func canReachServer(port int) bool {
	timeout := 500 * time.Millisecond
	address := fmt.Sprintf("127.0.0.1:%d", port)

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}

	conn.Close()
	return true
}
