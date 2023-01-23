package server

import (
	"context"
	"fmt"
	"log"
	"net"
)

type TCPServer struct {
	listener net.Listener
}

func (s *TCPServer) Init(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	s.listener = ln
	return nil
}

func (s *TCPServer) Start(ctx context.Context) <-chan net.Conn {
	c := make(chan net.Conn)
	go func() {
		defer s.listener.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := s.listener.Accept()
				if err != nil {
					log.Printf("could not accept client %v", err)
					continue
				}
				c <- conn
			}
		}
	}()
	return c
}
