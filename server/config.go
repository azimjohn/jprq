package main

type Environment string

const (
	Development Environment = "dev"
	Production  Environment = "prod"
)

type Config struct {
	Environment      Environment
	PublicServerPort int
	EventServerPort  int
}

func (c *Config) Load() error {
	c.Environment = Development
	c.PublicServerPort = 80
	c.EventServerPort = 4321

	return nil
}
