package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/azimjohn/jprq/server/config"
	"github.com/azimjohn/jprq/server/events"
	"github.com/azimjohn/jprq/server/github"
	"github.com/azimjohn/jprq/server/server"
	"github.com/azimjohn/jprq/server/tunnel"
)

const dateFormat = "2006/01/02 15:04:05"

type Jprq struct {
	config          config.Config
	eventServer     server.TCPServer
	publicServer    server.TCPServer
	publicServerTLS server.TCPServer
	allowedUsers    map[string]string
	allowedLastMod  time.Time
	authenticator   github.Authenticator
	cnameMap        map[string]string
	tcpTunnels      map[uint16]*tunnel.TCPTunnel
	httpTunnels     map[string]*tunnel.HTTPTunnel
	userTunnels     map[string]map[string]tunnel.Tunnel
	mu              sync.RWMutex
}

func (j *Jprq) Init(conf config.Config, oauth github.Authenticator) error {
	j.config = conf
	j.authenticator = oauth
	j.allowedUsers = make(map[string]string)
	j.cnameMap = make(map[string]string)
	j.tcpTunnels = make(map[uint16]*tunnel.TCPTunnel)
	j.httpTunnels = make(map[string]*tunnel.HTTPTunnel)
	j.userTunnels = make(map[string]map[string]tunnel.Tunnel)

	if err := j.eventServer.Init(conf.EventServerPort, "jprq_event_server"); err != nil {
		return err
	}
	if err := j.publicServer.Init(conf.PublicServerPort, "jprq_public_server"); err != nil {
		return err
	}
	err := j.publicServerTLS.InitTLS(conf.PublicServerTLSPort, "jprq_public_server_tls", conf.TLSCertFile, conf.TLSKeyFile)
	return err
}

func (j *Jprq) Start() {
	go j.eventServer.Start(j.serveEventConn)
	go j.publicServer.Start(j.servePublicConn)
	go j.publicServerTLS.Start(j.servePublicConn)

	go func() { // periodically load allowed users
		j.loadAllowedUsers()
		for range time.Tick(5 * time.Second) {
			j.loadAllowedUsers()
		}
	}()
}

func (j *Jprq) Stop() error {
	if err := j.eventServer.Stop(); err != nil {
		return err
	}
	if err := j.publicServer.Stop(); err != nil {
		return err
	}
	err := j.publicServerTLS.Stop()
	return err
}

func (j *Jprq) servePublicConn(conn net.Conn) error {
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	host, buffer, err := parseHost(conn)
	if err != nil || host == "" {
		writeResponse(conn, 400, "Bad Request", "Bad Request")
		return nil
	}
	if tunnelHost := j.resolveCname(host); tunnelHost != "" {
		host = tunnelHost
	}
	host = strings.ToLower(host)
	t, found := j.getHTTPTunnel(host)
	if !found {
		writeResponse(conn, 404, "Not Found", "tunnel not found. create one at jprq.io")
		return fmt.Errorf("unknown host requested %s", host)
	}
	return t.PublicConnectionHandler(conn, buffer)
}

func (j *Jprq) serveEventConn(conn net.Conn) error {
	defer conn.Close()

	var event events.Event[events.TunnelRequested]
	if err := event.Read(conn); err != nil {
		return err
	}

	request := event.Data
	if request.Protocol != events.HTTP && request.Protocol != events.TCP {
		return events.WriteError(conn, "invalid protocol %s", request.Protocol)
	}
	user, err := j.authenticator.Authenticate(request.AuthToken)
	if err != nil {
		return events.WriteError(conn, "authentication failed %s", "\n\tobtain auth token from https://jprq.io/auth\n")
	}

	if !j.isUserAllowed(user.Login) && !user.Allowed {
		return events.WriteError(conn, "jprq is now invite-only service %s\n", "\n\tbuy membership - https://buymeacoffee.com/azimjon \n")
	}
	if j.countUserTunnels(user.Login) >= j.config.MaxTunnelsPerUser {
		return events.WriteError(conn, "tunnels limit reached for %s", user.Login)
	}
	if request.Subdomain == "" {
		request.Subdomain = user.Login
	}
	if err := validate(&request.Subdomain); err != nil {
		return events.WriteError(conn, "invalid subdomain %s: %s", request.Subdomain, err.Error())
	}
	hostname := fmt.Sprintf("%s.%s", request.Subdomain, j.config.DomainName)
	if _, ok := j.getHTTPTunnel(hostname); ok {
		return events.WriteError(conn, "subdomain is busy: %s, try another one", request.Subdomain)
	}
	cname := request.CanonName
	if cname != "" && j.cnameTaken(cname) {
		return events.WriteError(conn, "cname is busy: %s, try another one", request.CanonName)
	}

	var t tunnel.Tunnel
	var maxConsLimit = j.config.MaxConsPerTunnel

	switch request.Protocol {
	case events.HTTP:
		tn, err := tunnel.NewHTTP(hostname, conn, maxConsLimit)
		if err != nil {
			return events.WriteError(conn, "failed to create http tunnel", err.Error())
		}
		subdomainBusy, cnameBusy := j.registerHTTPTunnel(hostname, cname, tn)
		if subdomainBusy {
			tn.Close()
			return events.WriteError(conn, "subdomain is busy: %s, try another one", request.Subdomain)
		}
		if cnameBusy {
			tn.Close()
			return events.WriteError(conn, "cname is busy: %s, try another one", request.CanonName)
		}
		defer j.unregisterHTTPTunnel(hostname, cname)
		t = tn
	case events.TCP:
		tn, err := tunnel.NewTCP(hostname, conn, maxConsLimit)
		if err != nil {
			return events.WriteError(conn, "failed to create tcp tunnel", err.Error())
		}
		j.registerTCPTunnel(tn)
		defer j.unregisterTCPTunnel(tn.PublicServerPort())
		t = tn
	}

	tunnelId := fmt.Sprintf("%s:%d", t.Hostname(), t.PublicServerPort())
	j.addUserTunnel(user.Login, tunnelId, t)
	defer j.removeUserTunnel(user.Login, tunnelId)

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
		_ = conn.SetReadDeadline(time.Now().Add(time.Minute))
		if _, err := conn.Read(buffer); err == io.EOF {
			break
		}
		if !j.isUserAllowed(user.Login) && !user.Allowed {
			break
		}
	}
	fmt.Printf("%s [tunnel-closed] %s: %s\n", time.Now().Format(dateFormat), user.Login, tunnelId)
	return nil
}

func (j *Jprq) loadAllowedUsers() {
	stat, err := os.Stat(j.config.AllowedUsersFile)
	if err != nil {
		log.Printf("failed to stat blocked users file: %s", err)
		return
	}
	if !stat.ModTime().After(j.allowedLastMod) {
		return
	}
	file, err := os.Open(j.config.AllowedUsersFile)
	if err != nil {
		log.Printf("failed to read blocked users file: %s", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	allowedUsers := make(map[string]string)

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ",")
		if len(fields) >= 2 {
			login := strings.TrimSpace(fields[0])
			allowedUsers[login] = strings.ToLower(fields[1])
		}
	}

	j.mu.Lock()
	j.allowedUsers = allowedUsers
	j.mu.Unlock()

	j.allowedLastMod = stat.ModTime()
	log.Println("allow-list updated")
}

func (j *Jprq) isUserAllowed(login string) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	_, found := j.allowedUsers[login]
	return found
}

func (j *Jprq) countUserTunnels(login string) int {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return len(j.userTunnels[login])
}

func (j *Jprq) resolveCname(host string) string {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.cnameMap[host]
}

func (j *Jprq) cnameTaken(cname string) bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	_, ok := j.cnameMap[cname]
	return ok
}

func (j *Jprq) getHTTPTunnel(host string) (*tunnel.HTTPTunnel, bool) {
	j.mu.RLock()
	defer j.mu.RUnlock()
	t, found := j.httpTunnels[host]
	return t, found
}

func (j *Jprq) registerHTTPTunnel(hostname, cname string, tn *tunnel.HTTPTunnel) (subdomainBusy, cnameBusy bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if _, ok := j.httpTunnels[hostname]; ok {
		return true, false
	}
	if _, ok := j.cnameMap[cname]; ok && cname != "" {
		return false, true
	}
	j.cnameMap[cname] = hostname
	j.httpTunnels[hostname] = tn
	return false, false
}

func (j *Jprq) unregisterHTTPTunnel(hostname, cname string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.cnameMap, cname)
	delete(j.httpTunnels, hostname)
}

func (j *Jprq) registerTCPTunnel(tn *tunnel.TCPTunnel) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.tcpTunnels[tn.PublicServerPort()] = tn
}

func (j *Jprq) unregisterTCPTunnel(port uint16) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.tcpTunnels, port)
}

func (j *Jprq) addUserTunnel(login, tunnelId string, t tunnel.Tunnel) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.userTunnels[login] == nil {
		j.userTunnels[login] = make(map[string]tunnel.Tunnel)
	}
	j.userTunnels[login][tunnelId] = t
}

func (j *Jprq) removeUserTunnel(login, tunnelId string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	if tunnels := j.userTunnels[login]; tunnels != nil {
		delete(tunnels, tunnelId)
	}
}
