package main

import (
	"log"
	"os"
	"os/signal"
)

func main() {
	var (
		conf Config
		jprq Jprq
	)

	err := conf.Load()
	if err != nil {
		log.Fatalf("failed to load conf %v", err)
	}

	err = jprq.Init(conf)
	if err != nil {
		log.Fatalf("failed to init jprq %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	jprq.Start()

	<-signalChan
	log.Printf("jprq server stopped")
}
