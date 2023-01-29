package main

import (
	"log"
	"time"
)

type Request struct {
	Id      uint64            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"header"`
}

type Response struct {
	RequestId uint64            `json:"request_id"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"header"`
	Body      string            `json:"body"`
}

func main() {
	w := NewWebServer()

	go func() {
		r := Request{}

		ticker := time.NewTicker(1 * time.Second)
		for _ = range ticker.C {
			w.DispatchEvent(r)
		}
	}()

	if err := w.Run(4444); err != nil {
		log.Fatalf("fail to run web server: %v", err)
	}
}
