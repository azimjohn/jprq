package jprq_tcp

import (
	"io"
	"net"
	"sync"
)

func PumpReadToWrite(readClient *net.Conn, writeClient *net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	defer (*writeClient).Close()

	buffer := make([]byte, 1024)
	for {
		length, err := (*readClient).Read(buffer)
		if err == io.EOF {
			break
		}
		_, err = (*writeClient).Write(buffer[:length])
		if err != nil {
			break
		}
	}
}
