package main

import (
	"log"
	"time"
)

func main() {
	w := webServer{}

	go func() {
		r := Request{}

		ticker := time.NewTicker(1 * time.Second)
		for _ = range ticker.C {
			w.Dispatch(r)
		}
	}()
	
	if err := w.Run(4444); err != nil {
		log.Fatalf("fail to run web server: %v", err)
	}
}
