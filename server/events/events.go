package events

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"net"
)

const (
	TCP  string = "tcp"
	HTTP string = "http"
)

type EventType interface {
	TunnelRequested | TunnelOpened | ConnectionReceived
}

type Event[Type EventType] struct {
	Data *Type
}

type TunnelRequested struct {
	Protocol   string
	Subdomain  string
	AuthToken  string
	CliVersion string
}

type TunnelOpened struct {
	Hostname      string
	Protocol      string
	PublicServer  uint16
	PrivateServer uint16
	ErrorMessage  string
}

type ConnectionReceived struct {
	ClientIP    net.IP
	ClientPort  uint16
	RateLimited bool
}

func WriteError(eventWriter io.Writer, message string, args ...string) error {
	event := Event[TunnelOpened]{
		Data: &TunnelOpened{
			ErrorMessage: fmt.Sprintf(message, args),
		},
	}
	event.Write(eventWriter)
	return errors.New(event.Data.ErrorMessage)
}

func (e *Event[EventType]) encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e.Data); err != nil {
		return nil, err
	}
	data := buf.Bytes()
	return data, nil
}

func (e *Event[EventType]) decode(data []byte) error {
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
	if err := e.decode(buffer); err != nil {
		return err
	}
	return nil
}

func (e *Event[EventType]) Write(conn io.Writer) error {
	data, err := e.encode()
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
