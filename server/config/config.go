package config

import (
	"errors"
	"os"
)

type Config struct {
	EventServerPort     uint16
	PublicServerPort    uint16
	EventServerTLSPort  uint16
	PublicServerTLSPort uint16
	TLSCertFile         string
	TLSKeyFile          string
}

func (c *Config) Load() error {
	c.PublicServerPort = 80
	c.EventServerPort = 4321
	c.EventServerTLSPort = 4322
	c.PublicServerTLSPort = 443
	c.TLSKeyFile = os.Getenv("TLS_KEY_FILE")
	c.TLSCertFile = os.Getenv("TLS_CERT_FILE")

	if c.TLSKeyFile == "" || c.TLSCertFile == "" {
		return errors.New("TLS key/cert file is missing")
	}
	return nil
}
