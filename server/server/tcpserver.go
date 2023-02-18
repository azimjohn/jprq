package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

type TCPServer struct {
	title    string
	listener net.Listener
}

func (s *TCPServer) Init(port uint16, title string) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.title = title
	s.listener = ln
	return nil
}

func (s *TCPServer) InitTLS(port uint16, title, certFile, keyFile string) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", port), &config)
	if err != nil {
		return err
	}
	s.title = title
	s.listener = ln
	return nil
}

func (s *TCPServer) Start(handler func(conn net.Conn) error) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go func() {
			err := handler(conn)
			if err != nil {
				log.Printf("[%s]: %s\n", s.title, err.Error())
			}
		}()
	}
}

func (s *TCPServer) Stop() error {
	return s.listener.Close()
}

func (s *TCPServer) Port() uint16 {
	return uint16(s.listener.Addr().(*net.TCPAddr).Port)
}
