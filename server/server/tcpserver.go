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

func (s *TCPServer) Start() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("closing server on port %d", s.Port())
			return
		}
		s.connections <- conn
	}
}

func (s *TCPServer) Stop() error {
	close(s.connections)
	return s.listener.Close()
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
