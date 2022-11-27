package main

import (
	"context"
	"go-donate-reacter/internal/services/http"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := http.Server{
		Addr: ":13077",
	}
	server.Run(ctx)

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	sig := <-signChan

	log.Println("cleanup started with", sig, "signal")
	cleanupStart := time.Now()

	// TODO: close all we need
	time.Sleep(time.Duration(3) * time.Second)

	cleanupElapsed := time.Since(cleanupStart)
	log.Printf("cleanup completed in %v seconds\n", cleanupElapsed.Seconds())

	os.Exit(1)
}
