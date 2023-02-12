package tunnel

import (
	"io"
)

type Tunnel interface {
	Open()
	Close()
	Hostname() string
	Protocol() string
	PublicServerPort() uint16
	PrivateServerPort() uint16
}

func Bind(readClient, writeClient io.ReadWriteCloser) {
	defer writeClient.Close()

	buffer := make([]byte, 1024*256)
	for {
		length, err := readClient.Read(buffer)
		if err != nil {
			break
		}
		_, err = writeClient.Write(buffer[:length])
		if err != nil {
			break
		}
	}
}
