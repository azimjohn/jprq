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
	"github.com/azimjohn/jprq/cli/i18n"
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
		log.Fatalf("%s %s\n", i18n.T("failed_connect_event_server"), err)
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
		log.Fatalf("%s %s\n", i18n.T("failed_sent_request"), err)
	}

	var t events.Event[events.TunnelOpened]
	if err := t.Read(eventCon); err != nil {
		log.Fatalf("%s %s\n", i18n.T("failed_receive_tunnel"), err)
	}
	if t.Data.ErrorMessage != "" {
		log.Fatalf(t.Data.ErrorMessage)
	}

	j.localServer = fmt.Sprintf("localhost:%d", port)
	j.remoteServer = fmt.Sprintf("jprq.%s:%d", j.config.Remote.Domain, t.Data.PrivateServer)
	j.publicServer = fmt.Sprintf("%s:%d", t.Data.Hostname, t.Data.PublicServer)

	if j.protocol == "http" {
		j.publicServer = fmt.Sprintf("https://%s", t.Data.Hostname)
	}

	fmt.Printf("%s \n", i18n.T("status_online"))
	fmt.Printf("%s \t %s \n", i18n.T("protocol"), strings.ToUpper(j.protocol))
	fmt.Printf("%s \t %s -> %s \n", i18n.T("forwarded"), strings.TrimSuffix(j.publicServer, ":80"), j.localServer)

	if j.protocol == "http" && debug {
		j.httpDebugger = debugger.New()
		if port, err := j.httpDebugger.Run(0); err == nil {
			fmt.Printf("Http %s: \t http://127.0.0.1:%d \n", i18n.T("debugger"), port)
		}
	}

	var event events.Event[events.ConnectionReceived]
	for {
		if err := event.Read(eventCon); err != nil {
			log.Fatalf("%s %s\n", i18n.T("failed_receive_event"), err)
		}
		go j.handleEvent(*event.Data)
	}
}

func (j *jprqClient) handleEvent(event events.ConnectionReceived) {
	localCon, err := net.Dial("tcp", j.localServer)
	if err != nil {
		log.Printf("%s %s\n", i18n.T("failed_connect_local_server"), err)
		return
	}
	defer localCon.Close()

	remoteCon, err := net.Dial("tcp", j.remoteServer)
	if err != nil {
		log.Printf("%s %s\n", i18n.T("failed_connect_remote_server"), err)
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
