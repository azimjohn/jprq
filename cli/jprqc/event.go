package jprqc

import (
	"encoding/binary"
	"github.com/azimjohn/jprq/server/events"
	"io"
	"log"
	"net"
	"time"
)

func (j *JprqClient) handleEvent(event events.ConnectionReceived) {
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

	if j.httpDebugger == nil {
		go bind(localCon, remoteCon, nil)
		bind(remoteCon, localCon, nil)
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
		if debugCon != nil {
			_, err = debugCon.Write(buf[:n])
		}
	}
	return nil
}
