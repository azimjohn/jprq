package main

import (
	"log"
	"net"
)

type Jprq struct {
	config       Config
	eventServer  TCPServer
	publicServer TCPServer
}

func (j *Jprq) Init(conf Config) error {
	j.config = conf
	err := j.eventServer.Init(conf.EventServerPort)
	if err != nil {
		return err
	}
	err = j.publicServer.Init(conf.PublicServerPort)
	if err != nil {
		return err
	}
	return nil
}

func (j *Jprq) Start() {
	go j.eventServer.Start()
	go j.publicServer.Start()

	go j.handleEventConnections()
	go j.handlePublicConnections()

	log.Println("jprq server started")
}

func (j *Jprq) handleEventConnections() {
	for conn := range j.eventServer.Connections() {
		go j.handleEventConnection(conn)
	}
}

func (j *Jprq) handlePublicConnections() {
	for conn := range j.eventServer.Connections() {
		go j.handlePublicConnection(conn)
	}
}

func (j *Jprq) handleEventConnection(conn net.Conn) {
	// todo
}

func (j *Jprq) handlePublicConnection(conn net.Conn) {
	// todo
}
