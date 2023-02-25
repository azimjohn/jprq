package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"time"
)

var version = "2.0"

func printHelp() {
	fmt.Println("Usage: jprq <command> [arguments]\n")
	fmt.Println("Commands:")
	fmt.Println("  auth <token>               Set authentication token from jprq.io/auth")
	fmt.Println("  tcp <port>                 Start a TCP tunnel on the specified port")
	fmt.Println("  http <port>                Start an HTTP tunnel on the specified port")
	fmt.Println("  http <port> -s <subdomain> Start an HTTP tunnel with a custom subdomain")
	fmt.Println("  --help                     Show this help message")
	fmt.Println("  --version                  Show the version number")
	os.Exit(0)
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("no command specified")
	}

	command := os.Args[1]
	args := os.Args[2:]
	protocol := ""

	switch command {
	case "auth":
		handleAuth(args)
	case "tcp", "http":
		protocol = command
	case "help", "--help":
		printHelp()
	case "version", "--version":
		printVersion()
	default:
		log.Fatalf("unknown command: %s, jprq --help", command)
	}

	if len(args) < 1 {
		log.Fatal("please specify port number, jprq --help")
	}
	port, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("port number must be an integer")
	}
	subdomain := ""
	if len(args) == 3 && args[1] == "-s" {
		subdomain = validate(args[2])
	}

	var conf Config
	if err := conf.Load(); err != nil {
		log.Fatal(err)
	}
	if !canReachServer(port) {
		log.Fatalf("error: cannot reach server on port: %d\n", port)
	}

	fmt.Printf("jprq: \t%s\n\n", version)
	defer log.Println("jprq tunnel closed")

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

func validate(subdomain string) string {
	subdomainRegex := `^[a-z\d](?:[a-z\d]|-[a-z\d]){0,38}$`
	if !regexp.MustCompile(subdomainRegex).MatchString(subdomain) {
		log.Fatalf("error: subdomain must be lowercase & alphanumeric")
	}
	return subdomain
}

func handleAuth(args []string) {
	if len(args) != 1 {
		log.Fatalf("invalid command, jprq --help")
	}
	config := Config{
		Local: struct {
			AuthToken string `json:"auth_token"`
		}{args[0]},
	}
	if err := config.Write(); err != nil {
		log.Fatalf("error writing config: %s", err)
	}
	log.Println("auth token has been set")
	os.Exit(0)
}

func canReachServer(port int) bool {
	address := fmt.Sprintf("127.0.0.1:%d", port)
	conn, err := net.DialTimeout("tcp", address, 512*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func printVersion() {
	log.Printf("v%s", version)
	os.Exit(0)
}
