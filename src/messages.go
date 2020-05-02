package main

import (
	"encoding/json"
)

const tunnelCreated = "tunnel_created"

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type TunnelCreatedMessage struct {
	Host  string `json:"host"`
	Token string `json:"token"`
}

func (m Message) Bytes() []byte {
	bytes, _ := json.Marshal(m)
	return bytes
}
