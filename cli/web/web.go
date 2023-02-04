package web

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//go:embed static/index.html
var html string

//go:embed static/style.css
var css string

//go:embed static/script.js
var js string

type webServer struct {
	listeners map[int64]chan<- interface{}
}

type WebServer interface {
	Run(port uint16) error
	DispatchEvent(event interface{})
}

func NewWebServer() WebServer {
	listeners := make(map[int64]chan<- interface{})
	web := &webServer{listeners: listeners}
	http.HandleFunc("/", contentHandler(html, "text/html"))
	http.HandleFunc("/script.js", contentHandler(js, "text/javascript"))
	http.HandleFunc("/style.css", contentHandler(css, "text/css"))
	http.HandleFunc("/events", web.eventHandler)
	http.HandleFunc("/store-token", authHandler)
	return web
}

func authHandler(w http.ResponseWriter, r *http.Request) {

}

func (web *webServer) Run(port uint16) error {
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func (web *webServer) DispatchEvent(event interface{}) {
	for _, listener := range web.listeners {
		listener := listener
		go func() { listener <- event }()
	}
}

func contentHandler(content string, contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(content))
	}
}

func (web *webServer) eventHandler(w http.ResponseWriter, r *http.Request) {
	events := make(chan interface{})
	requestId := time.Now().UnixNano()
	web.listeners[requestId] = events
	defer close(events)
	defer delete(web.listeners, requestId)

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
