package main

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

//go:embed xterm.html
var html string

func main() {
	app := &App{}
	http.HandleFunc("/", contentHandler([]byte(html), "text/html"))
	http.HandleFunc("/ws/ssh", app.ConnectionHandler)
	log.Fatal(http.ListenAndServe(":2222", nil))
}

type App struct {
}

func contentHandler(content []byte, contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(content)
	}
}

func (a *App) ConnectionHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, "error accepting webSocket connection", 400)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "closed")

	var info ConnectionInfo
	connection := r.URL.Query().Get("connection")
	connInfo, err := base64.StdEncoding.DecodeString(connection)
	if err != nil {
		http.Error(w, "error parsing connection info", 400)
		return
	}

	if err := json.Unmarshal(connInfo, &info); err != nil {
		http.Error(w, "error parsing connection info", 400)
		return
	}

	if err := a.ShellOverWS(r.Context(), conn, info); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (a *App) ShellOverWS(ctx context.Context, ws *websocket.Conn, info ConnectionInfo) error {
	auth := ssh.Password(info.Password)
	session := SSHShellSession{
		Node: NewSSHNode(info.Host, info.Port),
	}
	var wsBuff WebSocketBufferWriter
	session.WriterPipe = &wsBuff

	err := session.Connect(info.Username, auth)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer session.Close()

	sshSession, err := session.Config(info.WindowSize.Cols, info.WindowSize.Rows)
	if err != nil {
		return fmt.Errorf("configure ssh error: %w", err)
	}

	defer wsBuff.Flush(ctx, websocket.MessageBinary, ws)

	done := make(chan bool, 3)
	setDone := func() { done <- true }

	go func(wc io.WriteCloser) {
		defer setDone()
		for {
			_, data, err := ws.Read(ctx)
			if err != nil {
				log.Println("error reading webSocket message:", err)
				return
			}
			if err = DispatchMessage(sshSession, data, wc); err != nil {
				log.Println("error writing data to ssh server:", err)
				return
			}
		}
	}(session.StdinPipe)

	stopper := make(chan bool)
	go func() {
		defer setDone()
		tick := time.NewTicker(time.Millisecond * 120)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				if err := wsBuff.Flush(ctx, websocket.MessageBinary, ws); err != nil {
					log.Println("error sending data to websocket:", err)
					return
				}
			case <-stopper:
				return
			}
		}
	}()

	go func() {
		defer setDone()
		if err := sshSession.Wait(); err != nil {
			log.Println("ssh session closed: ", err)
		}
	}()

	<-done
	stopper <- true
	return nil
}
