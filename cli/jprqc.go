package main

import (
	"encoding/binary"
	"fmt"
	"github.com/azimjohn/jprq/cli/debugger"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/tunnel"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

type jprqClient struct {
	config       Config
	protocol     string
	subdomain    string
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
			AuthToken:  j.config.Local.AuthToken,
			CliVersion: version,
		},
	}
	if err := request.Write(eventCon); err != nil {
		log.Fatalf("failed to send request: %s\n", err)
	}

	var tunnel events.Event[events.TunnelOpened]
	if err := tunnel.Read(eventCon); err != nil {
		log.Fatalf("failed to receive tunnel info: %s\n", err)
	}
	if tunnel.Data.ErrorMessage != "" {
		log.Fatalf(tunnel.Data.ErrorMessage)
	}

	j.localServer = fmt.Sprintf("127.0.0.1:%d", port)
	j.remoteServer = fmt.Sprintf("jprq.%s:%d", j.config.Remote.Domain, tunnel.Data.PrivateServer)
	j.publicServer = fmt.Sprintf("%s:%d", tunnel.Data.Hostname, tunnel.Data.PublicServer)

	fmt.Printf("Status: \t Online \n")
	fmt.Printf("Protocol: \t %s \n", strings.ToUpper(j.protocol))
	fmt.Printf("Forwarded: \t %s -> %s \n", strings.TrimSuffix(j.publicServer, ":80"), j.localServer)

	if j.protocol == "http" {
		j.publicServer = fmt.Sprintf("https://%s", tunnel.Data.Hostname)
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
		go tunnel.Bind(localCon, remoteCon)
		tunnel.Bind(remoteCon, localCon)
		return
	}

	debugCon := j.httpDebugger.Connection(event.ClientPort)

	go bind(localCon, remoteCon, debugCon.Response())
	bind(remoteCon, localCon, debugCon.Request())
}

func bind(src net.Conn, dst net.Conn, debugCon io.Writer) error {
	defer src.Close()
	defer dst.Close()
	buf := make([]byte, 4096)
	for {
		_ = src.SetReadDeadline(time.Now().Add(time.Second))
		n, err := src.Read(buf)
		if err == io.EOF {
			break
		}
		_ = dst.SetWriteDeadline(time.Now().Add(time.Second))
		_, err = dst.Write(buf[:n])
		if err != nil {
			return err
		}
		_, err = debugCon.Write(buf[:n])
	}
	return nil
}
