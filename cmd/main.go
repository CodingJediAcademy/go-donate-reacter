package main

import (
	"go-donate-reacter/internal/services/app"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	app := app.New()
	app.Start()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	sig := <-signChan

	log.Println("cleanup started with", sig, "signal")
	cleanupStart := time.Now()

	// TODO: close all we need
	app.Stop()

	cleanupElapsed := time.Since(cleanupStart)
	log.Printf("cleanup completed in %v seconds\n", cleanupElapsed.Seconds())

	os.Exit(1)
}
