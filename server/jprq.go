package main

import (
	"fmt"
	"github.com/azimjohn/jprq/server/config"
	"golang.org/x/net/context"
	"log"
	"net"
)

type Jprq struct {
	config       config.Config
	publicServer net.Listener
}

func (j *Jprq) Init(conf config.Config) error {
	j.config = conf
	err2 := j.initPublicServer()
	if err2 != nil {
		return err2
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
