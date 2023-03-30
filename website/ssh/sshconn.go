package main

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"strconv"
)

type Node struct {
	Host   string
	Port   int
	client *ssh.Client
}

func NewSSHNode(host string, port int) Node {
	return Node{Host: host, Port: port, client: nil}
}

func (node *Node) GetClient() (*ssh.Client, error) {
	if node.client == nil {
		return nil, errors.New("no connection, connect first")
	}
	return node.client, nil
}

func (node *Node) Connect(username string, auth ssh.AuthMethod) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(node.Host, strconv.Itoa(node.Port)), config)
	if err != nil {
		return err
	}
	node.client = client
	return nil
}

type SSHShellSession struct {
	Node
	// calling Write() to write data to ssh server
	StdinPipe io.WriteCloser
	// Write() be called to receive data from ssh server
	WriterPipe io.Writer
	session    *ssh.Session
}

func (s *SSHShellSession) Config(cols, rows int) (*ssh.Session, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, err
	}
	s.session = session

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal("failed to set IO stdin: ", err)
		return nil, err
	}
	s.StdinPipe = stdin

	if s.WriterPipe == nil {
		return nil, errors.New("WriterPipe is nil")
	}
	session.Stdout = s.WriterPipe
	session.Stderr = s.WriterPipe

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", rows, cols, modes); err != nil {
		log.Fatal("request for pseudo terminal failed: ", err)
		return nil, err
	}
	if err := session.Shell(); err != nil {
		log.Fatal("failed to start remote shell: ", err)
		return nil, err
	}
	return session, nil
}

func (s *SSHShellSession) Close() {
	if s.session != nil {
		s.session.Close()
	}
	if s.client != nil {
		s.client.Close()
	}
}
