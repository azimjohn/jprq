package config

type Environment string

const (
	Development Environment = "dev"
	Production  Environment = "prod"
)

type Config struct {
	Environment         Environment
	PublicServerPort    int
	WebSocketServerPort int
}

func (c *Config) Load() error {
	c.Environment = Development
	c.PublicServerPort = 80
	c.WebSocketServerPort = 4321

	return nil
}
