package main

import (
	"encoding/binary"
	"fmt"
	"github.com/azimjohn/jprq/server/events"
	"io"
	"log"
	"net"
)

type jprqClient struct {
	port      int
	config    Config
	protocol  string
	subdomain string
}

func (j *jprqClient) Start() {
	request := events.Event[events.TunnelRequested]{
		Data: &events.TunnelRequested{
			Protocol:   j.protocol,
			Subdomain:  j.subdomain,
			AuthToken:  j.config.Local.AuthToken,
			CliVersion: version,
		},
	}

	eventCon, err := net.Dial("tcp", j.config.Remote.Events)
	if err != nil {
		log.Fatalf("error connecting to event server: %s\n", err)
	}
	defer eventCon.Close()

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

	// todo: display tunnel info

	var event events.Event[events.ConnectionReceived]
	for {
		if err := event.Read(eventCon); err != nil {
			log.Fatalf("error receiving connection received event: %s\n", err)
		}
		go j.handleEvent(*event.Data, tunnel.Data.PrivateServer)
	}
}

func (j *jprqClient) handleEvent(event events.ConnectionReceived, privateServerPort uint16) {
	localCon, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", j.port))
	if err != nil {
		log.Printf("error connecting to local server: %s\n", err)
		return
	}
	defer localCon.Close()

	remoteCon, err := net.Dial("tcp", fmt.Sprintf("jprq.live:%d", privateServerPort))
	if err != nil {
		log.Printf("error connecting to remote server: %s\n", err)
		return
	}
	defer remoteCon.Close()

	buffer := make([]byte, 2)
	binary.LittleEndian.PutUint16(buffer, event.ClientPort)
	remoteCon.Write(buffer)

	go io.Copy(localCon, remoteCon)
	io.Copy(remoteCon, localCon)
}
