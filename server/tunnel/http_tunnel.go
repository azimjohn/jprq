package tunnel

import (
	"errors"
	"github.com/azimjohn/jprq/server/server"
	"regexp"
	"strings"
)

const DefaultHttpPort = 80

var regex = regexp.MustCompile(`^[a-z0-9]+[a-z0-9\-]+[a-z0-9]$`)
var blockList = map[string]bool{"www": true, "jprq": true}

type HTTPTunnel struct {
	hostname      string
	privateServer server.TCPServer
}

func NewHTTPTunnel(hostname string) (HTTPTunnel, error) {
	var t HTTPTunnel
	t.hostname = hostname
	if err := validate(hostname); err != nil {
		return t, err
	}
	if err := t.privateServer.Init(0); err != nil {
		return t, err
	}
	return t, nil
}

func (t *HTTPTunnel) Protocol() string {
	return "http"
}

func (t *HTTPTunnel) Hostname() string {
	return t.hostname
}

func (t *HTTPTunnel) PrivateServerPort() uint16 {
	return t.privateServer.Port()
}

func (t *HTTPTunnel) PublicServerPort() uint16 {
	return DefaultHttpPort
}

func (t *HTTPTunnel) Start() {
	go t.privateServer.Start()
	// todo handle private connections
}

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
