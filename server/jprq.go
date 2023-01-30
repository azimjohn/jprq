package main

import (
	"github.com/azimjohn/jprq/server/config"
	"github.com/azimjohn/jprq/server/server"
	"github.com/azimjohn/jprq/server/tunnel"
	"net"
)

type Jprq struct {
	config          config.Config
	eventServer     server.TCPServer
	publicServer    server.TCPServer
	publicServerTLS server.TCPServer
	userTunnels     map[string]tunnel.Tunnel
	tcpTunnels      map[uint16]tunnel.TCPTunnel
	httpTunnels     map[string]tunnel.HTTPTunnel
}

func (j *Jprq) Init(conf config.Config) error {
	j.config = conf
	j.userTunnels = make(map[string]tunnel.Tunnel)
	j.tcpTunnels = make(map[uint16]tunnel.TCPTunnel)
	j.httpTunnels = make(map[string]tunnel.HTTPTunnel)

	err := j.eventServer.Init(conf.EventServerPort)
	if err != nil {
		return err
	}
	err = j.publicServer.Init(conf.PublicServerPort)
	if err != nil {
		return err
	}
	err = j.publicServerTLS.InitTLS(conf.PublicServerTLSPort, conf.TLSCertFile, conf.TLSKeyFile)
	if err != nil {
		return err
	}
	return nil
}

func (j *Jprq) Start() {
	go j.eventServer.Start()
	go j.publicServer.Start()
	go j.publicServerTLS.Start()

	go j.eventServer.Serve(j.serveEventConn)
	go j.publicServer.Serve(j.servePublicConn)
	go j.publicServerTLS.Serve(j.servePublicConn)
}

func (j *Jprq) Stop() error {
	if err := j.eventServer.Stop(); err != nil {
		return err
	}
	if err := j.publicServer.Stop(); err != nil {
		return err
	}
	if err := j.publicServerTLS.Stop(); err != nil {
		return err
	}
	return nil
}

func (j *Jprq) serveEventConn(conn net.Conn) {
	// todo
}

func (j *Jprq) servePublicConn(conn net.Conn) {
	// todo
}
