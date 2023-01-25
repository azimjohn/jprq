package events

import (
	"bytes"
	"encoding/gob"
)

type Protocol string

const (
	TCP  Protocol = "tcp"
	HTTP Protocol = "http"
)

type EventType interface {
	TunnelRequested | TunnelStarted | ConnectionReceived
}

type Event[Type EventType] struct {
	Data *Type
}

func (e *Event[EventType]) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e.Data); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	return data, nil
}

func (e *Event[EventType]) Decode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&e.Data); err != nil {
		return err
	}
	return nil
}

type TunnelRequested struct {
	Subdomain  string
	Protocol   Protocol
	AuthToken  string
	CliVersion string
}

type TunnelStarted struct {
	Host           string   `json:"host"`
	Protocol       Protocol `json:"protocol"`
	PublicServer   uint16   `json:"public_server"`
	PrivateServer  uint16   `json:"private_server"`
	UserMessage    string   `json:"user_message"`
	MaxConnections uint16   `json:"max_connections"`
}

type ConnectionReceived struct {
	ClientIP    string `json:"client_ip"`
	ClientPort  uint16 `json:"client_port"`
	RateLimited bool   `json:"rate_limited"`
}
