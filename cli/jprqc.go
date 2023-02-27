package main

import (
	"encoding/binary"
	"fmt"
	"github.com/azimjohn/jprq/cli/web"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/tunnel"
	"log"
	"net"
	"strings"
)

type jprqClient struct {
	config       Config
	protocol     string
	subdomain    string
	localServer  string
	remoteServer string
	publicServer string
	webInterface web.WebServer
}

func (j *jprqClient) Start(port int) {
	eventCon, err := net.Dial("tcp", j.config.Remote.Events)
	if err != nil {
		log.Fatalf("error connecting to event server: %s\n", err)
	}
	defer eventCon.Close()

	request := events.Event[events.TunnelRequested]{
		Data: &events.TunnelRequested{
			Protocol:   j.protocol,
			Subdomain:  j.subdomain,
			AuthToken:  j.config.Local.AuthToken,
			CliVersion: version,
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
	j.remoteServer = fmt.Sprintf("jprq.%s:%d", j.config.Remote.Domain, tunnel.Data.PrivateServer)
	j.publicServer = fmt.Sprintf("%s:%d", tunnel.Data.Hostname, tunnel.Data.PublicServer)

	if j.protocol == "http" {
		go j.runWebInterface(4444)
		j.publicServer = fmt.Sprintf("https://%s", tunnel.Data.Hostname)
	}

	fmt.Printf("Status: \t Online \n")
	fmt.Printf("Protocol: \t %s \n", strings.ToUpper(j.protocol))
	fmt.Printf("Forwarded: \t %s -> %s \n", j.publicServer, j.localServer)

	var event events.Event[events.ConnectionReceived]
	for {
		if err := event.Read(eventCon); err != nil {
			log.Fatalf("error receiving connection received event: %s\n", err)
		}
		go j.handleEvent(*event.Data)
	}
}

func (j *jprqClient) handleEvent(event events.ConnectionReceived) {
	localCon, err := net.Dial("tcp", j.localServer)
	if err != nil {
		log.Printf("error connecting to local server: %s\n", err)
		return
	}
	defer localCon.Close()

	remoteCon, err := net.Dial("tcp", j.remoteServer)
	if err != nil {
		log.Printf("error connecting to remote server: %s\n", err)
		return
	}
	defer remoteCon.Close()

	buffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(buffer, event.ClientPort)
	remoteCon.Write(buffer)

	go tunnel.Bind(localCon, remoteCon)
	tunnel.Bind(remoteCon, localCon)
}

func (j *jprqClient) runWebInterface(port uint16) {
	j.webInterface = web.NewWebServer()
	fmt.Printf("Web Interface: \t http://127.0.0.1:%d (for debugging)\n", port)
	if err := j.webInterface.Run(port); err != nil {
		fmt.Printf("Web Interface: \t failed to run on port %d\n", port)
	}
}
