package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

type TCPServer struct {
	listener    net.Listener
	connections chan net.Conn
}

func (s *TCPServer) Init(port uint16) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.listener = ln
	s.connections = make(chan net.Conn)
	return nil
}

func (s *TCPServer) InitTLS(port uint16, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), &config)
	if err != nil {
		return err
	}
	s.listener = ln
	s.connections = make(chan net.Conn)
	return nil
}

func (s *TCPServer) Start() <-chan net.Conn {
	defer s.listener.Close()
	defer close(s.connections)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("could not accept client %v", err)
			continue
		}
		s.connections <- conn
	}
}

func (s *TCPServer) Serve(handler func(conn net.Conn)) {
	for conn := range s.connections {
		go handler(conn)
	}
}

func (s *TCPServer) Connections() <-chan net.Conn {
	return s.connections
}

func (s *TCPServer) Port() uint16 {
	return uint16(s.listener.Addr().(*net.TCPAddr).Port)
}
