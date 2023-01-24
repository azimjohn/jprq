package main

import (
	"github.com/azimjohn/jprq/server/config"
	"log"
	"os"
	"os/signal"
)

func main() {
	var (
		conf config.Config
		jprq Jprq
	)

	err := conf.Load()
	if err != nil {
		log.Fatalf("failed to load conf: %v", err)
	}

	err = jprq.Init(conf)
	if err != nil {
		log.Fatalf("failed to init jprqpkg %v", err)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	jprq.Start()
	log.Println("jprq server started")

	<-signalChan
	log.Printf("jprq server stopped")
}
