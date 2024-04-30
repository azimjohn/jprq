package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"
)

var version = "2.3"

type Flags struct {
	debug     bool
	cname     string
	subdomain string
}

func printVersion() {
	log.Printf("v%s", version)
	os.Exit(0)
}

func printHelp() {
	fmt.Printf(`
Usage: jprq <command> [arguments]

Commands:
  auth  <token>               Set authentication token from jprq.io/auth
  tcp   <port>                Start a TCP tunnel on the specified port
  http  <port>                Start an HTTP tunnel on the specified port
  http  <port> -s <subdomain> Start an HTTP tunnel with a custom subdomain
  http  <port> --debug        Debug an HTTP tunnel with Jprq Debugger
  serve <dir>                 Serve files with built-in Http Server
  --help                      Show this help message
  --version                   Show the version number
`)
	os.Exit(0)
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Println("no command specified")
		printHelp()
	}

	switch os.Args[1] {
	case "help", "--help":
		printHelp()
	case "version", "--version":
		printVersion()
	}

	if len(os.Args) < 3 {
		log.Println("no arg supplied")
		printHelp()
	}

	protocol, port := "", 0
	command, arg := os.Args[1], os.Args[2]
	flags := parseFlags(os.Args[3:])

	switch command {
	case "auth":
		handleAuth(arg)
	case "serve":
		protocol, port = handleServe(arg)
	case "tcp", "http":
		protocol = command
		port, _ = strconv.Atoi(arg)
	default:
		log.Fatalf("unknown command: %s, jprq --help", command)
	}

	if port <= 0 {
		log.Fatalf("port number must be a positive integer")
	}

	var conf Config
	if err := conf.Load(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("jprq %s \t press Ctrl+C to quit\n\n", version)
	defer log.Println("jprq tunnel closed")

	client := jprqClient{
		config:    conf,
		protocol:  protocol,
		subdomain: flags.subdomain,
		cname:     flags.cname,
	}

	go client.Start(port, flags.debug)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}

func parseFlags(args []string) Flags {
	var flags Flags
	for i, arg := range args {
		switch arg {
		case "-d", "-debug", "--debug":
			flags.debug = true
		case "-s", "-subdomain", "--subdomain":
			flags.subdomain = args[i+1]
		case "-c", "-cname", "--cname":
			flags.cname = args[i+1]
		}
	}
	return flags
}

func handleAuth(token string) {
	config := Config{
		Local: struct {
			AuthToken string `json:"auth_token"`
		}{token},
	}
	if err := config.Write(); err != nil {
		log.Fatalf("error writing config: %s", err)
	}
	log.Println("auth token has been set")
	os.Exit(0)
}

func handleServe(dir string) (string, int) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("no such dir %s", dir)
	}

	handler := http.FileServer(http.Dir(dir))
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("failed to start server: %s", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	go func() {
		if err := http.Serve(listener, handler); err != nil {
			log.Fatalf("cannot serve files on %s: %s", dir, err)
		}
	}()

	time.AfterFunc(600*time.Millisecond, func() {
		log.Println("Serving: \t", dir)
	})
	return "http", port
}
