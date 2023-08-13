package jprqc

import (
	"fmt"
	"github.com/azimjohn/jprq/cli/debugger"
	"github.com/azimjohn/jprq/server/events"
	"log"
	"net"
	"os/user"
	"strings"
)

var Version = "2.2"

type JprqClient struct {
	Config       Config
	Server       string
	Protocol     string
	Subdomain    string
	localServer  string
	remoteServer string
	publicServer string
	httpDebugger debugger.Debugger
}

func (j *JprqClient) Start(port int, debug bool) {
	eventCon, err := net.Dial("tcp", j.Config.Remote.Events)
	if err != nil {
		log.Fatalf("error connecting to event server: %s\n", err)
	}
	defer eventCon.Close()

	request := events.Event[events.TunnelRequested]{
		Data: &events.TunnelRequested{
			Protocol:   j.Protocol,
			Subdomain:  j.Subdomain,
			AuthToken:  j.Config.Local.AuthToken,
			CliVersion: Version,
		},
	}
	if err := request.Write(eventCon); err != nil {
		log.Fatalf("error sendind request: %s\n", err)
	}

	var tunnel events.Event[events.TunnelOpened]
	if err := tunnel.Read(eventCon); err != nil {
		log.Fatalf("error receiving tunnel info: %s\n", err)
	}
	if tunnel.Data.ErrorMessage != "" {
		log.Fatalf(tunnel.Data.ErrorMessage)
	}

	j.localServer = fmt.Sprintf("127.0.0.1:%d", port)
	j.remoteServer = fmt.Sprintf("jprq.%s:%d", j.Config.Remote.Domain, tunnel.Data.PrivateServer)
	j.publicServer = fmt.Sprintf("%s:%d", tunnel.Data.Hostname, tunnel.Data.PublicServer)

	fmt.Printf("Status: \t Online \n")
	fmt.Printf("Protocol: \t %s \n", strings.ToUpper(j.Protocol))
	fmt.Printf("Forwarded: \t %s -> %s \n", j.publicServer, j.localServer)

	if j.Protocol == "http" {
		j.publicServer = fmt.Sprintf("https://%s", tunnel.Data.Hostname)
	}

	if j.Protocol == "http" && debug {
		j.httpDebugger = debugger.New()
		if port, err := j.httpDebugger.Run(0); err == nil {
			fmt.Printf("Http Debugger: \t http://127.0.0.1:%d \n", port)
		}
	}

	if j.Server == "ssh" {
		u, _ := user.Current()
		conn := fmt.Sprintf("%s@%d", u.Username, tunnel.Data.PublicServer)
		log.Printf("Web Console: \t https://ssh.jprq.io?c=%s", conn)
	}

	var event events.Event[events.ConnectionReceived]
	for {
		if err := event.Read(eventCon); err != nil {
			log.Fatalf("error receiving connection received event: %s\n", err)
		}
		go j.handleEvent(*event.Data)
	}
}
