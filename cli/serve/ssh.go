package serve

import (
	"github.com/gliderlabs/ssh"
	"log"
	"net"
)

func handleSSH() int {
	listener, err := net.Listen("tcp", ":0")
	port := listener.Addr().(*net.TCPAddr).Port

	if err != nil {
		log.Fatalf("error ")
	}
	ssh.Handle(func(s ssh.Session) {
		s.Write([]byte("hello world!"))
	})
	go func() {
		log.Fatalf("failed to serve: %s", ssh.Serve(listener, nil))
	}()
	return port
}
