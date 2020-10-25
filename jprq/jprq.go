package jprq

type Jprq struct {
	baseHost string
	tunnels map[string]Tunnel
}

func New(baseHost string) Jprq {
	return Jprq{
		baseHost: baseHost,
		tunnels: make(map[string]Tunnel),
	}
}
