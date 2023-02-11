package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/azimjohn/jprq/server/config"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/github"
	"github.com/azimjohn/jprq/server/server"
	"github.com/azimjohn/jprq/server/tunnel"
	"io"
	"net"
	"regexp"
	"strings"
)

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

func (j *Jprq) servePublicConn(conn net.Conn) error {
	reader := bufio.NewReader(conn)

	first, err := reader.ReadString('\n')
	if err != nil {
		return conn.Close() // todo write http response: bad request
	}
	second, err := reader.ReadString('\n')
	if err != nil {
		return conn.Close() // todo write http response: bad request
	}
	i := strings.Index(second, ":")
	if i < 0 {
		return conn.Close() // todo write http response: bad request
	}
	host := strings.Trim(second[i+1:], "\r\n")
	host = strings.ToLower(host)
	t, found := j.httpTunnels[host]
	if !found {
		return conn.Close() // todo write http response: not found
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
		return events.WriteError("authentication failed", conn)
	}

	if request.Protocol != events.HTTP && request.Protocol != events.TCP {
		return events.WriteError("invalid protocol", conn)
	}

	if len(j.userTunnels[user.Login]) >= j.config.MaxTunnelsPerUser {
		return events.WriteError("tunnels limit reached", conn)
	}

	if _, ok := j.httpTunnels[request.Hostname]; ok {
		return events.WriteError("host is currently busy", conn)
	}

	if err := validate(request.Hostname); err != nil {
		return events.WriteError(err.Error(), conn)
	}

	var t tunnel.Tunnel
	var maxConsLimit = j.config.MaxConsPerTunnel

	switch request.Protocol {
	case events.HTTP:
		tn, err := tunnel.NewHTTP(request.Hostname, conn, maxConsLimit)
		if err != nil {
			return events.WriteError("failed to create tunnel", conn)
		}
		j.httpTunnels[request.Hostname] = tn
		defer delete(j.httpTunnels, request.Hostname)
		t = tn
	case events.TCP:
		tn, err := tunnel.NewTCP(request.Hostname, conn, maxConsLimit)
		if err != nil {
			return events.WriteError("failed to create tunnel", conn)
		}
		j.tcpTunnels[tn.PublicServerPort()] = tn
		defer delete(j.tcpTunnels, tn.PublicServerPort())
		t = tn
	}

	tunnelId := fmt.Sprintf("%s:%d", t.Hostname(), t.PublicServerPort())
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
	buffer := make([]byte, 8) // wait until connection is closed
	for {
		if _, err := conn.Read(buffer); err == io.EOF {
			return err
		}
	}
}

var regex = regexp.MustCompile(`^[a-z0-9]+[a-z0-9\-]+[a-z0-9]$`)
var blockList = map[string]bool{"www": true, "jprq": true}

func validate(hostname string) error {
	domains := strings.Split(hostname, ".")
	if len(domains) != 3 {
		return errors.New("invalid hostname")
	}
	subdomain := domains[0]
	if len(subdomain) > 42 || len(subdomain) < 3 {
		return errors.New("subdomain length must be between 3 and 42")
	}
	if blockList[subdomain] {
		return errors.New("subdomain is in deny list")
	}
	if !regex.MatchString(subdomain) {
		return errors.New("subdomain must be lowercase & alphanumeric")
	}
	return nil
}
