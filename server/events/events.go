package events

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
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

type TunnelRequested struct {
	Hostname   string
	Protocol   Protocol
	AuthToken  string
	CliVersion string
}

type TunnelStarted struct {
	Hostname       string   `json:"host_name"`
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

func (e *Event[EventType]) Read(conn io.Reader) error {
	buffer := make([]byte, 2)
	if _, err := conn.Read(buffer); err != nil {
		return err
	}
	length := binary.LittleEndian.Uint16(buffer)
	buffer = make([]byte, length)
	if _, err := conn.Read(buffer); err != nil {
		return err
	}
	if err := e.Decode(buffer); err != nil {
		return err
	}
	return nil
}

func (e *Event[EventType]) Write(conn io.Writer) error {
	data, err := e.Encode()
	if err != nil {
		return err
	}
	length := make([]byte, 2)
	binary.LittleEndian.PutUint16(length, uint16(len(data)))
	if _, err := conn.Write(length); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}
