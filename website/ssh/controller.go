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
	"strconv"
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

func (app *App) ConnectionHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		http.Error(w, "error accepting webSocket connection", 400)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "closed")

	params := r.URL.Query()
	connBase64 := params.Get("connection")
	connInfo, err := base64.StdEncoding.DecodeString(connBase64)
	if err != nil {
		log.Println("could not decode base64 connection string: ", connBase64)
		return
	}

	var info ConnectionInfo
	if err := json.Unmarshal(connInfo, &info); err != nil {
		log.Println("error parsing connection info: ", connBase64)
		return
	}

	window := WindowSize{}
	window.Cols, _ = strconv.Atoi(params.Get("cols"))
	window.Rows, _ = strconv.Atoi(params.Get("rows"))

	if err := app.ShellOverWS(r.Context(), conn, info, window); err != nil {
		log.Println("could not establish ssh connection: ", err.Error())
	}
}

func (app *App) ShellOverWS(ctx context.Context, ws *websocket.Conn, info ConnectionInfo, window WindowSize) error {
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

	sshSession, err := session.Config(window.Cols, window.Rows)
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
