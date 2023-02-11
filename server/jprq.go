package main

import (
	"errors"
	"fmt"
	"github.com/azimjohn/jprq/server/config"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/github"
	"github.com/azimjohn/jprq/server/server"
	"github.com/azimjohn/jprq/server/tunnel"
	"io"
	"net"
	"strings"
	"time"
)

const dateFormat = "2006/01/02 15:04:05"

type Jprq struct {
	config          config.Config
	eventServer     server.TCPServer
	publicServer    server.TCPServer
	publicServerTLS server.TCPServer
	authenticator   github.Authenticator
	tcpTunnels      map[uint16]*tunnel.TCPTunnel
	httpTunnels     map[string]*tunnel.HTTPTunnel
	userTunnels     map[string]map[string]tunnel.Tunnel
}

func (j *Jprq) Init(conf config.Config, oauth github.Authenticator) error {
	j.config = conf
	j.authenticator = oauth
	j.tcpTunnels = make(map[uint16]*tunnel.TCPTunnel)
	j.httpTunnels = make(map[string]*tunnel.HTTPTunnel)
	j.userTunnels = make(map[string]map[string]tunnel.Tunnel)

	if err := j.eventServer.Init(conf.EventServerPort, "jprq_event_server"); err != nil {
		return err
	}
	if err := j.publicServer.Init(conf.PublicServerPort, "jprq_public_server"); err != nil {
		return err
	}
	if err := j.publicServerTLS.InitTLS(conf.PublicServerTLSPort, "jprq_public_server_tls", conf.TLSCertFile,
		conf.TLSKeyFile); err != nil {
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

func (j *Jprq) servePublicConn(conn net.Conn) error {
	first, err := readLine(conn)
	if err != nil {
		writeResponse(conn, 400, "Bad Request", "Bad Request")
		return nil
	}
	second, err := readLine(conn)
	if err != nil {
		writeResponse(conn, 400, "Bad Request", "Bad Request")
		return nil
	}
	i := strings.Index(second, ":")
	if i < 0 {
		writeResponse(conn, 400, "Bad Request", "Bad Request")
		return errors.New("error reading host header from request")
	}
	host := strings.Trim(second[i+1:], "\r\n ")
	host = strings.ToLower(host)
	t, found := j.httpTunnels[host]
	if !found {
		writeResponse(conn, 404, "Not Found", "tunnel not found")
		return errors.New(fmt.Sprintf("unknown host requested %s", host))
	}
	return t.PublicConnectionHandler(conn, []byte(first+second))
}

func (j *Jprq) serveEventConn(conn net.Conn) error {
	defer conn.Close()

	var event events.Event[events.TunnelRequested]
	if err := event.Read(conn); err != nil {
		return err
	}

	request := event.Data
	user, err := j.authenticator.Authenticate(request.AuthToken)
	if err != nil {
		return events.WriteError(conn, "authentication failed")
	}

	if request.Protocol != events.HTTP && request.Protocol != events.TCP {
		return events.WriteError(conn, "invalid protocol %s", string(request.Protocol))
	}

	if len(j.userTunnels[user.Login]) >= j.config.MaxTunnelsPerUser {
		return events.WriteError(conn, "tunnels limit reached for %s", user.Login)
	}

	if _, ok := j.httpTunnels[request.Hostname]; ok {
		return events.WriteError(conn, "%host currently busy: %s", request.Hostname)
	}

	if err := validate(request.Hostname); err != nil {
		return events.WriteError(conn, "invalid hostname %s: %s", request.Hostname, err.Error())
	}

	var t tunnel.Tunnel
	var maxConsLimit = j.config.MaxConsPerTunnel

	switch request.Protocol {
	case events.HTTP:
		tn, err := tunnel.NewHTTP(request.Hostname, conn, maxConsLimit)
		if err != nil {
			return events.WriteError(conn, "failed to create http tunnel", err.Error())
		}
		j.httpTunnels[request.Hostname] = tn
		defer delete(j.httpTunnels, request.Hostname)
		t = tn
	case events.TCP:
		tn, err := tunnel.NewTCP(request.Hostname, conn, maxConsLimit)
		if err != nil {
			return events.WriteError(conn, "failed to create tcp tunnel", err.Error())
		}
		j.tcpTunnels[tn.PublicServerPort()] = tn
		defer delete(j.tcpTunnels, tn.PublicServerPort())
		t = tn
	}

	tunnelId := fmt.Sprintf("%s:%d", t.Hostname(), t.PublicServerPort())
	if len(j.userTunnels[user.Login]) == 0 {
		j.userTunnels[user.Login] = make(map[string]tunnel.Tunnel)
	}
	j.userTunnels[user.Login][tunnelId] = t
	defer delete(j.userTunnels[user.Login], tunnelId)

	t.Open()
	defer t.Close()
	opened := events.Event[events.TunnelOpened]{
		Data: &events.TunnelOpened{
			Hostname:      t.Hostname(),
			Protocol:      t.Protocol(),
			PublicServer:  t.PublicServerPort(),
			PrivateServer: t.PrivateServerPort(),
		},
	}
	if err := opened.Write(conn); err != nil {
		return err
	}

	fmt.Printf("%s [tunnel-opened] %s: %s\n", time.Now().Format(dateFormat), user.Login, tunnelId)
	buffer := make([]byte, 8) // wait until connection is closed
	for {
		if _, err := conn.Read(buffer); err == io.EOF {
			break
		}
	}
	fmt.Printf("%s [tunnel-closed] %s: %s\n", time.Now().Format(dateFormat), user.Login, tunnelId)
	return nil
}
