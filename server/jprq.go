package main

import (
	"github.com/azimjohn/jprq/server/config"
	"github.com/azimjohn/jprq/server/server"
	"net"
)

type Jprq struct {
	config          config.Config
	eventServer     server.TCPServer
	publicServer    server.TCPServer
	eventServerTLS  server.TCPServer
	publicServerTLS server.TCPServer
}

func (j *Jprq) Init(conf config.Config) error {
	j.config = conf
	err := j.eventServer.Init(conf.EventServerPort)
	if err != nil {
		return err
	}
	err = j.publicServer.Init(conf.PublicServerPort)
	if err != nil {
		return err
	}
	err = j.eventServerTLS.InitTLS(conf.EventServerPort, conf.TLSCertFile, conf.TLSKeyFile)
	if err != nil {
		return err
	}
	err = j.publicServerTLS.InitTLS(conf.PublicServerPort, conf.TLSCertFile, conf.TLSKeyFile)
	if err != nil {
		return err
	}
	return nil
}

func (j *Jprq) Start() {
	go j.eventServer.Start()
	go j.publicServer.Start()
	go j.eventServerTLS.Start()
	go j.publicServerTLS.Start()

	go j.eventServer.Serve(j.serveEventConn)
	go j.publicServer.Serve(j.servePublicConn)
	go j.eventServerTLS.Serve(j.serveEventConn)
	go j.publicServerTLS.Serve(j.servePublicConn)
}

func (j *Jprq) serveEventConn(conn net.Conn) {
	// todo
}

func (j *Jprq) servePublicConn(conn net.Conn) {
	// todo
}
