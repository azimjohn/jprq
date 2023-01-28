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
