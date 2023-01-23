package events

type Protocol string

const (
	TCP  Protocol = "tcp"
	HTTP Protocol = "http"
)

type TunnelRequested struct {
	Subdomain  string
	Protocol   Protocol
	AuthToken  string
	CliVersion string
}

type TunnelStarted struct {
	Host           string   `json:"host"`
	Protocol       Protocol `json:"protocol"`
	PublicServer   int      `json:"public_server"`
	PrivateServer  int      `json:"private_server"`
	UserMessage    string   `json:"user_message"`
	MaxConnections int      `json:"max_connections"`
}

type ConnectionReceived struct {
	ClientIP    string `json:"client_ip"`
	ClientPort  int    `json:"client_port"`
	RateLimited bool   `json:"rate_limited"`
}
