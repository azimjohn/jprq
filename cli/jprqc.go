package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/azimjohn/jprq/cli/debugger"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/tunnel"
)

type jprqClient struct {
	config       Config
	protocol     string
	subdomain    string
	cname        string
	localServer  string
	remoteServer string
	publicServer string
	httpDebugger debugger.Debugger
}

func (j *jprqClient) Start(port int, debug bool) {
	eventCon, err := net.Dial("tcp", j.config.Remote.Events)
	if err != nil {
		log.Fatalf("failed to connect to event server: %s\n", err)
	}
	defer eventCon.Close()

	request := events.Event[events.TunnelRequested]{
		Data: &events.TunnelRequested{
			Protocol:   j.protocol,
			Subdomain:  j.subdomain,
			CanonName:  j.cname,
			AuthToken:  j.config.Local.AuthToken,
			CliVersion: version,
		},
	}
	if err := request.Write(eventCon); err != nil {
		log.Fatalf("failed to send request: %s\n", err)
	}

	var t events.Event[events.TunnelOpened]
	if err := t.Read(eventCon); err != nil {
		log.Fatalf("failed to receive tunnel info: %s\n", err)
	}
	if t.Data.ErrorMessage != "" {
		log.Fatalf(t.Data.ErrorMessage)
	}

	j.localServer = fmt.Sprintf("127.0.0.1:%d", port)
	j.remoteServer = fmt.Sprintf("jprq.%s:%d", j.config.Remote.Domain, t.Data.PrivateServer)
	j.publicServer = fmt.Sprintf("%s:%d", t.Data.Hostname, t.Data.PublicServer)

	fmt.Printf("Status: \t Online \n")
	fmt.Printf("Protocol: \t %s \n", strings.ToUpper(j.protocol))
	fmt.Printf("Forwarded: \t %s -> %s \n", strings.TrimSuffix(j.publicServer, ":80"), j.localServer)

	if j.protocol == "http" {
		j.publicServer = fmt.Sprintf("https://%s", t.Data.Hostname)
	}
	if j.protocol == "http" && debug {
		j.httpDebugger = debugger.New()
		if port, err := j.httpDebugger.Run(0); err == nil {
			fmt.Printf("Http Debugger: \t http://127.0.0.1:%d \n", port)
		}
	}

	var event events.Event[events.ConnectionReceived]
	for {
		if err := event.Read(eventCon); err != nil {
			log.Fatalf("failed to receive connection-received event: %s\n", err)
		}
		go j.handleEvent(*event.Data)
	}
}

func (j *jprqClient) handleEvent(event events.ConnectionReceived) {
	localCon, err := net.Dial("tcp", j.localServer)
	if err != nil {
		log.Printf("failed to connect to local server: %s\n", err)
		return
	}
	defer localCon.Close()

	remoteCon, err := net.Dial("tcp", j.remoteServer)
	if err != nil {
		log.Printf("failed to connect to remote server: %s\n", err)
		return
	}
	defer remoteCon.Close()

	buffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(buffer, event.ClientPort)
	remoteCon.Write(buffer)

	if j.httpDebugger == nil {
		go tunnel.Bind(localCon, remoteCon, nil)
		tunnel.Bind(remoteCon, localCon, nil)
		return
	}

	debugCon := j.httpDebugger.Connection(event.ClientPort)
	go tunnel.Bind(localCon, remoteCon, debugCon.Response())
	tunnel.Bind(remoteCon, localCon, debugCon.Request())
}
