package config

type Environment string

const (
	Development Environment = "dev"
	Production  Environment = "prod"
)

type Config struct {
	Environment         Environment
	EventServerPort     int
	PublicServerPort    int
	EventServerTLSPort  int
	PublicServerTLSPort int
	TLSCertFile         string
	TLSKeyFile          string
}

func (c *Config) Load() error {
	c.Environment = Development
	c.PublicServerPort = 80
	c.EventServerPort = 4321
	c.PublicServerTLSPort = 443
	c.EventServerTLSPort = 4322
	c.TLSKeyFile = "~/.cert/jprqpkg.key"   // todo read from env
	c.TLSCertFile = "~/.cert/jprqpkg.cert" // todo read from env
	return nil
}
