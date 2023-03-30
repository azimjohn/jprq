package main

import (
	"context"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
	"time"
)

func main() {
	app := &App{}
	http.HandleFunc("/ws/ssh", app.HandleWebsocket)
	log.Fatal(http.ListenAndServe(":2222", nil))
}

type App struct {
}

func (a *App) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		http.Error(w, "error accepting webSocket connection", 400)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "closed")

	info := ConnectionInfo{} // todo read from query params
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

	var wg sync.WaitGroup
	wg.Add(1)

	go func(wc io.WriteCloser) {
		defer wg.Done()
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
		defer wg.Done()
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
		defer wg.Done()
		if err := sshSession.Wait(); err != nil {
			log.Println("ssh exist from server", err)
		}
	}()

	wg.Wait()
	stopper <- true
	return nil
}
