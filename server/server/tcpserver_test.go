package server

import (
	"testing"
)

func TestTCPServerInit(t *testing.T) {
	s := &TCPServer{}
	err := s.Init(1234, "test_server")
	defer s.Stop()

	if err != nil {
		t.Fatalf("failed to init server: %v", err)
	}

	if s.listener == nil {
		t.Fatalf("listener not created")
	}
}

func TestTCPServerPort(t *testing.T) {
	s := &TCPServer{}
	defer s.Stop()

	err := s.Init(1234, "test_server")
	if err != nil {
		t.Fatalf("failed to init server: %v", err)
	}

	port := s.Port()
	if port != 1234 {
		t.Fatalf("expected %d, got %d", 1234, port)
	}
}
