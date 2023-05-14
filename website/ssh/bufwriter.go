package main

import (
	"bytes"
	"context"
	"nhooyr.io/websocket"
	"sync"
)

type WebSocketBufferWriter struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (w *WebSocketBufferWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(p)
}

func (w *WebSocketBufferWriter) Flush(ctx context.Context, messageType websocket.MessageType, ws *websocket.Conn) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.buffer.Len() != 0 {
		err := ws.Write(ctx, messageType, w.buffer.Bytes())
		if err != nil {
			return err
		}
		w.buffer.Reset()
	}
	return nil
}
