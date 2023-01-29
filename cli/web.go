package main

import (
	_ "embed"
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
}

func contentHandler(content string, contentType string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write([]byte(content))
	}
}

func (w *webServer) Run(port uint16) error {
	http.HandleFunc("/", contentHandler(html, "text/html"))
	http.HandleFunc("/script.js", contentHandler(js, "text/javascript"))
	http.HandleFunc("/style.css", contentHandler(css, "text/css"))
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		// send new request/responses
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
