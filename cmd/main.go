package main

import (
	"go-donate-reacter/internal/app"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	a := app.New()
	a.Start()
	defer a.Stop()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	sig := <-signChan

	log.Println("cleanup started with", sig, "signal")
	cleanupStart := time.Now()

	// TODO: close all we need
	a.Stop()

	cleanupElapsed := time.Since(cleanupStart)
	log.Printf("cleanup completed in %v seconds\n", cleanupElapsed.Seconds())

	os.Exit(1)
}
