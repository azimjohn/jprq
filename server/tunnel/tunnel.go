package tunnel

type Tunnel interface {
	Open()
	Close()
	Hostname() string
	Protocol() string
	PublicServerPort() uint16
	PrivateServerPort() uint16
}
