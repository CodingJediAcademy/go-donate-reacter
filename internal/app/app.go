package app

import (
	"context"
	"go-donate-reacter/internal/services/donationalerts"
	"go-donate-reacter/internal/services/http"
	"go-donate-reacter/internal/storage/badgerDB"
	"log"
	"os"
)

type App struct {
	Storage badgerDB.Storage

	oauthCodeChan chan string
	ctxCancelFunc []context.CancelFunc
}

func New() App {
	storage := badgerDB.NewStorage()
	return App{
		Storage: storage,
	}
}

func (a *App) ctxGenerate() *context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	a.ctxCancelFunc = append(a.ctxCancelFunc, cancel)

	return &ctx
}

func (a *App) Start() {
	a.oauthCodeChan = make(chan string, 1)

	a.Storage.Init()

	a.httpServerStart(*a.ctxGenerate())
	a.connectToDA(*a.ctxGenerate())
}

func (a *App) Stop() {
	a.Storage.Close()
	for _, cancelFunc := range a.ctxCancelFunc {
		cancelFunc()
	}
}

func (a *App) httpServerStart(ctx context.Context) {
	server := http.NewServer(":13077", a.oauthCodeChan)
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
		a.oauthCodeChan,
		daToken,
	)

	if err := a.Storage.SaveToken(daClient.Token); err != nil {
		log.Println(err)
	}

	profile, err := daClient.Profile()
	if err != nil {
		log.Println(err)
	}

	cent := donationalerts.Centrifuge{
		AccessToken: daClient.Token.AccessToken,
		SocketToken: profile.Data.SocketToken,
		UserID:      profile.Data.ID,
	}

	cent.Connect(ctx)
}
