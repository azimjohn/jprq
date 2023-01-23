package main

import (
	"context"
	"fmt"
	"log"
	"net"
)

type Jprq struct {
	config       Config
	publicServer net.Listener
}

func (j *Jprq) Init(conf Config) error {
	j.config = conf
	err := j.initPublicServer()
	if err != nil {
		return err
	}
	return nil
}

func (j *Jprq) initPublicServer() error {
	port := fmt.Sprintf(":%d", j.config.PublicServerPort)
	ln, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	j.publicServer = ln
	return nil
}

func (j *Jprq) Start(ctx context.Context) {
	// todo start all servers
	log.Println("jprq server started")
}

func (j *Jprq) Stop() {
	j.publicServer.Close()
	log.Printf("jprq server stopped")
}
