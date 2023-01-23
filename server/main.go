package main

import (
	"context"
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

	ctx := context.Background()
	ctx, cancelFunc := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	go jprq.Start(ctx)
	defer jprq.Stop()

	<-signalChan
	cancelFunc()
}
