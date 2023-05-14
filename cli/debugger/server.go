package debugger

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/djherbis/buffer"
	"github.com/djherbis/nio/v3"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Conn interface {
	Request() io.Writer
	Response() io.Writer
}

type Debugger interface {
	Run(port int) (int, error)
	Connection(id uint16) Conn
}

type conn struct {
	requestReader  io.Reader
	requestWriter  io.Writer
	responseReader io.Reader
	responseWriter io.Writer
}

type debugger struct {
	listeners   map[int64]chan<- interface{}
	connections map[uint16]*conn
}

func New() Debugger {
	d := &debugger{
		listeners:   make(map[int64]chan<- interface{}),
		connections: make(map[uint16]*conn),
	}
	http.HandleFunc("/", contentHandler(html, "text/html"))
	http.HandleFunc("/script.js", contentHandler(js, "text/javascript"))
	http.HandleFunc("/style.css", contentHandler(css, "text/css"))
	http.HandleFunc("/events", d.eventHandler)
	return d
}

func (d *debugger) Run(port int) (int, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return 0, err
	}
	go http.Serve(listener, nil)
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func (c *conn) Request() io.Writer {
	return c.requestWriter
}

func (c *conn) Response() io.Writer {
	return c.responseWriter
}

func (d *debugger) Connection(id uint16) Conn {
	c := &conn{}
	d.connections[id] = c
	c.requestReader, c.requestWriter = nio.Pipe(buffer.New(1 << 18))
	c.responseReader, c.responseWriter = nio.Pipe(buffer.New(1 << 18))
	go parseRequests(c.requestReader, strconv.Itoa(int(id)), d.dispatchEvent)
	go parseResponses(c.responseReader, strconv.Itoa(int(id)), d.dispatchEvent)
	return c
}

func (d *debugger) dispatchEvent(event interface{}) {
	for _, listener := range d.listeners {
		listener := listener
		go func() { listener <- event }()
	}
}

func (d *debugger) eventHandler(w http.ResponseWriter, r *http.Request) {
	events := make(chan interface{})
	listenerId := time.Now().UnixNano()
	d.listeners[listenerId] = events
	defer close(events)
	defer delete(d.listeners, listenerId)

	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(200)

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-events:
			data, _ := json.Marshal(event)
			content := fmt.Sprintf("data: %s\n\n", string(data))
			w.Write([]byte(content))
			w.(http.Flusher).Flush()
		}
	}
}

//go:embed static/index.html
var html string

//go:embed static/style.css
var css string

//go:embed static/script.js
var js string

func contentHandler(content string, contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(content))
	}
}
