package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
)

//go:embed web/index.html
var html string

//go:embed web/style.css
var css string

//go:embed web/script.js
var js string

type Request struct {
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	Body     string            `json:"body"`
	Headers  map[string]string `json:"header"`
	Response Response          `json:"response"`
}

type Response struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"header"`
	Body    string            `json:"body"`
}

type webServer struct {
	listeners []chan<- Request
}

func contentHandler(content string, contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(content))
	}
}

func (web *webServer) Dispatch(request Request) {
	for _, listener := range web.listeners {
		listener := listener
		go func() { listener <- request }()
	}
}

func (web *webServer) Run(port uint16) error {
	http.HandleFunc("/", contentHandler(html, "text/html"))
	http.HandleFunc("/script.js", contentHandler(js, "text/javascript"))
	http.HandleFunc("/style.css", contentHandler(css, "text/css"))
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		events := make(chan Request)
		web.listeners = append(web.listeners, events)

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)

		for event := range events {
			data, _ := json.Marshal(event)
			content := fmt.Sprintf("data: %s\n\n", string(data))
			w.Write([]byte(content))
			w.(http.Flusher).Flush()
		}
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
