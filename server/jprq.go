package main

import (
	"context"
	"github.com/azimjohn/jprq/server/server"
	"log"
)

type Jprq struct {
	config       Config
	publicServer server.TCPServer
	eventServer  server.TCPServer
}

func (j *Jprq) Init(conf Config) error {
	j.config = conf
	err := j.publicServer.Init(conf.PublicServerPort)
	if err != nil {
		return err
	}
	err = j.eventServer.Init(conf.EventServerPort)
	if err != nil {
		return err
	}
	return nil
}

func (j *Jprq) Start(ctx context.Context) {
	j.eventServer.Start(ctx)
	j.publicServer.Start(ctx)
	// todo handle connections
	log.Println("jprq server started")
}
