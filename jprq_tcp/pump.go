package jprq_tcp

import (
	"net"
	"sync"
)

func PumpReadToWrite(readClient, writeClient *net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
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
