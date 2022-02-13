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

var version = "1.2.0"
var tcpBaseHost = "tcp.jprq.io"
var httpBaseHost = "open.jprq.io"

func main() {
	subdomain := flag.String("subdomain", "", "Subdomain for HTTP Tunnel")
	host := flag.String("host", "", "Host for Tunnel")
	flag.Parse()
	log.SetFlags(0)
	args := flag.Args()

	if len(os.Args) < 3 {
		log.Fatalf("Usage: jprq [--subdomain] [--host] <PROTOCOL> <PORT>\n"+
			"  Supported Protocols: [tcp, http]\n"+
			"  Optional Argument: -subdomain, -host\n"+
			"  Client Version: %s\n", version)
	}

	protocol := args[0]
	if protocol != "tcp" && protocol != "http" {
		log.Fatalf("Invalid Protocol: %s\n", protocol)
	}

	port, err := strconv.Atoi(args[1])
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

	if protocol == "tcp" {
		if *host == "" {
			// if no host has been supplied, fall back to default for tcp
			*host = tcpBaseHost
		}
		go openTCPTunnel(port, *host, ctx)
	} else if protocol == "http" {
		if *host == "" {
			// if no host has been supplied, fall back to default for http
			*host = httpBaseHost
		}
		fmt.Println(subdomain)
		//go openHTTPTunnel(port, *host, *subdomain, ctx)
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
