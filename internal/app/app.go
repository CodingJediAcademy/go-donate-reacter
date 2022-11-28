package app

import (
	"context"
	"go-donate-reacter/internal/services/donationalerts"
	"go-donate-reacter/internal/services/http"
	"log"
	"os"
	"time"
)

type App struct {
	cancelFunc context.CancelFunc
}

func New() App {
	return App{}
}

func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel

	a.httpServerStart(ctx)
	a.connectToDA(ctx)
}

func (a *App) Stop() {
	a.cancelFunc()
}

func (a *App) httpServerStart(ctx context.Context) {
	server := http.Server{
		Addr: ":13077",
	}
	server.Run(ctx)
}

func (a *App) connectToDA(ctx context.Context) {
	daClient := donationalerts.NewClient(
		os.Getenv("GDR_DA_CLIENT_ID"),
		os.Getenv("GDR_DA_CLIENT_SECRET"),
		os.Getenv("GDR_DA_REDIRECT_URL"),
	)
	log.Println(daClient.AuthLink())
	err := daClient.NewToken(os.Getenv("GDR_DA_CODE"))
	if err != nil {
		log.Println(err)
	}
	log.Printf("%#v", daClient.Token.AccessToken)

	profile, err := daClient.Profile()
	if err != nil {
		log.Println(err)
	}
	log.Printf("%#v", profile)

	cent := donationalerts.Centrifuge{
		AccessToken: daClient.Token.AccessToken,
		SocketToken: profile.Data.SocketToken,
		UserID:      profile.Data.ID,
	}

	cent.Connect(ctx)
	time.Sleep(time.Duration(2) * time.Second)
	cent.Subscribe(ctx)
}
