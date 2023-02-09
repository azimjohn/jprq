package main

import (
	"github.com/azimjohn/jprq/cli/web"
	"log"
	"time"
)

type Request struct {
	Id      uint64            `json:"id"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

type Response struct {
	RequestId uint64            `json:"request_id"`
	Status    int               `json:"status"`
	Headers   map[string]string `json:"headers"`
	Body      string            `json:"body"`
}

func main() {
	w := web.NewWebServer()

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
