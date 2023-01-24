package main

import (
	"fmt"
	"log"
	"net"
)

type TCPServer struct {
	listener    net.Listener
	connections chan net.Conn
}

func (s *TCPServer) Init(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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

func (s *TCPServer) Connections() <-chan net.Conn {
	return s.connections
}
