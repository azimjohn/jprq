package main

import "os"

var config Config

type Config struct {
	BaseHostName  string
	JwtSigningKey string
}

func (c *Config) Load() {
	c.BaseHostName = os.Getenv("BASE_HOST_NAME")
	c.JwtSigningKey = os.Getenv("JWT_SIGNING_KEY")
}
