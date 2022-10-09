package jprq_tcp

import (
	"net"
	"sync"
)

func BindTCPConnections(remoteConn, localConn *net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		pumpReadToWrite(remoteConn, localConn)
		wg.Done()
	}()
	go func() {
		pumpReadToWrite(localConn, remoteConn)
		wg.Done()
	}()

	wg.Wait()
}

func pumpReadToWrite(readClient, writeClient *net.Conn) {
	defer (*writeClient).Close()

	buffer := make([]byte, 1024*256)
	for {
		length, err := (*readClient).Read(buffer)
		if err != nil {
			break
		}
		_, err = (*writeClient).Write(buffer[:length])
		if err != nil {
			break
		}
	}
}
