package app

import (
	"context"
	"go-donate-reacter/internal/services/donationalerts"
	"go-donate-reacter/internal/services/http"
	"go-donate-reacter/internal/storage/badgerDB"
	"log"
	"os"
	"time"
)

type App struct {
	Storage badgerDB.Storage

	codeChan      chan string
	ctxCancelFunc context.CancelFunc
}

func New() App {
	storage := badgerDB.NewStorage()
	return App{
		Storage: storage,
	}
}

func (a *App) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	a.ctxCancelFunc = cancel
	a.codeChan = make(chan string, 1)

	a.Storage.Init()

	a.httpServerStart(ctx)
	a.connectToDA(ctx)
}

func (a *App) Stop() {
	a.Storage.Close()
	a.ctxCancelFunc()
}

func (a *App) httpServerStart(ctx context.Context) {
	server := http.NewServer(":13077", a.codeChan)
	server.Run(ctx)
}

func (a *App) connectToDA(ctx context.Context) {
	daToken, err := a.Storage.GetToken()
	if err != nil {
		log.Println(err)
	}

	daClient := donationalerts.NewClient(
		ctx,
		os.Getenv("GDR_DA_CLIENT_ID"),
		os.Getenv("GDR_DA_CLIENT_SECRET"),
		os.Getenv("GDR_DA_REDIRECT_URL"),
		a.codeChan,
		daToken,
	)

	if err := a.Storage.SaveToken(daClient.Token); err != nil {
		log.Println(err)
	}

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
