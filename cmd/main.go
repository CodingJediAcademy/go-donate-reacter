package main

import (
	"context"
	"go-donate-reacter/internal/services/donationalerts"
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

	daClient := donationalerts.Client{
		ID:          os.Getenv("GDR_DA_CLIENT_ID"),
		Secret:      os.Getenv("GDR_DA_CLIENT_SECRET"),
		RedirectUrl: os.Getenv("GDR_DA_REDIRECT_URL"),
	}

	log.Println(daClient.AuthLink())
	token, err := daClient.NewToken("")
	if err != nil {
		log.Println(err)
	}
	log.Printf("%#v", token)

	profile, err := daClient.Profile(token.AccessToken)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%#v", profile)

	cent := donationalerts.Centrifuge{
		AccessToken: token.AccessToken,
		SocketToken: profile.Data.SocketToken,
		UserID:      profile.Data.ID,
	}

	cent.Connect()
	time.Sleep(time.Duration(2) * time.Second)
	cent.Subscribe()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	sig := <-signChan

	log.Println("cleanup started with", sig, "signal")
	cleanupStart := time.Now()

	// TODO: close all we need
	//time.Sleep(time.Duration(3) * time.Second)

	cleanupElapsed := time.Since(cleanupStart)
	log.Printf("cleanup completed in %v seconds\n", cleanupElapsed.Seconds())

	os.Exit(1)
}
