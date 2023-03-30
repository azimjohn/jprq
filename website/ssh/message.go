package main

import (
	"encoding/base64"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"io"
)

const (
	SSHWebSocketMessageTypeTerminal  = "terminal"
	SSHWebSocketMessageTypeHeartbeat = "heartbeat"
	SSHWebSocketMessageTypeResize    = "resize"
)

type ConnectionInfo struct {
	Host       string       `json:"host"`
	Port       int          `json:"port"`
	Username   string       `json:"username"`
	Password   string       `json:"password"`
	WindowSize WindowResize `json:"window"`
}

type SSHWebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"` // json.RawMessage
}

type TerminalMessage struct {
	DataBase64 string `json:"base64"`
}

type WindowResize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

func DispatchMessage(sshSession *ssh.Session, data []byte, wc io.WriteCloser) error {
	var socketData json.RawMessage
	socketStream := SSHWebSocketMessage{
		Data: &socketData,
	}

	if err := json.Unmarshal(data, &socketStream); err != nil {
		return nil
	}

	switch socketStream.Type {
	case SSHWebSocketMessageTypeHeartbeat:
		return nil
	case SSHWebSocketMessageTypeResize:
		var resize WindowResize
		if err := json.Unmarshal(socketData, &resize); err != nil {
			return nil // skip error
		}
		sshSession.WindowChange(resize.Rows, resize.Cols)
	case SSHWebSocketMessageTypeTerminal:
		var message TerminalMessage
		if err := json.Unmarshal(socketData, &message); err != nil {
			return err
		}
		decodeBytes, _ := base64.StdEncoding.DecodeString(message.DataBase64)
		if _, err := wc.Write(decodeBytes); err != nil {
			return err
		}
	}
	return nil
}
