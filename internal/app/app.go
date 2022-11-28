package app

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"go-donate-reacter/internal/services/donationalerts"
	"go-donate-reacter/internal/services/http"
	"go-donate-reacter/internal/storage/badgerDB"
	"log"
	"os"
	"time"
)

type App struct {
	Storage       badgerDB.Storage
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

	a.Storage.Init()

	a.httpServerStart(ctx)
	a.connectToDA(ctx)
}

func (a *App) Stop() {
	a.Storage.Close()
	a.ctxCancelFunc()
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
	/*err := daClient.NewToken(os.Getenv("GDR_DA_CODE"))
	if err != nil {
		log.Println(err)
	}
	log.Printf("%#v", daClient.Token.AccessToken)*/

	if err := a.extractToken(&daClient); err != nil {
		log.Fatalln(err)
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

func (a *App) extractToken(c *donationalerts.Client) error {
	token, err := a.Storage.GetToken()
	if err == badger.ErrKeyNotFound {
		//TODO: make login func
		code := os.Getenv("GDR_DA_CODE")
		if code == "" {
			return errors.New("empty code")
		}

		if err := c.NewToken(code); err != nil {
			return err
		}
		token = c.Token

		if err := a.Storage.SaveToken(token); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// DA sends deepish in ExpiresIn
	/*expiredDate := time.Unix(token.ExpiresIn, 0)
	if expiredDate.After(time.Now()) {
		if err := c.RefreshToken(); err != nil {
			return err
		}
		token = c.Token

		if err := a.Storage.SaveToken(token); err != nil {
			return err
		}
	}*/

	return nil
}
