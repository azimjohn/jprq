package config

import (
	"errors"
	"os"
)

type Config struct {
	DomainName          string
	MaxTunnelsPerUser   uint16
	MaxConsPerTunnel    uint16
	EventServerPort     uint16
	PublicServerPort    uint16
	PublicServerTLSPort uint16
	TLSCertFile         string
	TLSKeyFile          string
	GithubClientID      string
	GithubClientSecret  string
}

func (c *Config) Load() error {
	c.MaxTunnelsPerUser = 4
	c.MaxConsPerTunnel = 25
	c.PublicServerPort = 80
	c.EventServerPort = 4321
	c.PublicServerTLSPort = 443
	c.DomainName = os.Getenv("JPRQ_DOMAIN")
	c.TLSKeyFile = os.Getenv("JPRQ_TLS_KEY")
	c.TLSCertFile = os.Getenv("JPRQ_TLS_CERT")
	c.GithubClientID = os.Getenv("GITHUB_CLIENT_ID")
	c.GithubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	if c.DomainName == "" {
		return errors.New("jprq domain env is not set")
	}
	if c.TLSKeyFile == "" || c.TLSCertFile == "" {
		return errors.New("TLS key/cert file is missing")
	}
	if c.GithubClientID == "" || c.GithubClientSecret == "" {
		return errors.New("github client id/secret is missing")
	}
	return nil
}
