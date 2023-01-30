package tunnel

type Tunnel interface {
	Open() error
	Close() error
	Hostname() string
	Protocol() string
	PublicServerPort() uint16
	PrivateServerPort() uint16
}
