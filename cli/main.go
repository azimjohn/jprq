package main

import "log"

func main() {
	w := webServer{}
	if err := w.Run(4444); err != nil {
		log.Fatalf("fail to run web server: %v", err)
	}
}
