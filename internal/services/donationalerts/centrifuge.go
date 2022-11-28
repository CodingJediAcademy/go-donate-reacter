package donationalerts

import (
	"fmt"
	"github.com/centrifugal/centrifuge-go"
	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"
	"go-donate-reacter/internal/services/donationalerts/request"
	"go-donate-reacter/internal/services/donationalerts/response"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	DA_CENT_WEBSOCKET_URL string = "wss://centrifugo.donationalerts.com/connection/websocket"
)

type Centrifuge struct {
	AccessToken  string
	SocketToken  string
	UserID       int64
	ClientID     string
	Channel      string
	ChannelToken string
	Client       *centrifuge.Client
}

type testPublishHandler struct {
	onPublish func(*centrifuge.Subscription, centrifuge.PublishEvent)
}

func (h *testPublishHandler) OnPublish(c *centrifuge.Subscription, e centrifuge.PublishEvent) {
	if h.onPublish != nil {
		h.onPublish(c, e)
	}
}

type testEventHandler struct {
	onConnect    func(*centrifuge.Client, centrifuge.ConnectEvent)
	onDisconnect func(*centrifuge.Client, centrifuge.DisconnectEvent)
	onMessage    func(*centrifuge.Client, centrifuge.MessageEvent)
	onPrivateSub func(*centrifuge.Client, centrifuge.PrivateSubEvent) (string, error)
}

func (h *testEventHandler) OnConnect(c *centrifuge.Client, e centrifuge.ConnectEvent) {
	if h.onConnect != nil {
		h.onConnect(c, e)
	}
}

func (h *testEventHandler) OnDisconnect(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
	if h.onDisconnect != nil {
		h.onDisconnect(c, e)
	}
}

func (h *testEventHandler) OnMessage(c *centrifuge.Client, e centrifuge.MessageEvent) {
	if h.onMessage != nil {
		h.onMessage(c, e)
	}
}

func (h *testEventHandler) OnPrivateSub(c *centrifuge.Client, e centrifuge.PrivateSubEvent) (string, error) {
	if h.onPrivateSub != nil {
		return h.onPrivateSub(c, e)
	}

	return "", nil
}

func (cent *Centrifuge) Connect(ctx context.Context) {
	header := http.Header{}
	header.Add("Authorization", "Bearer "+cent.AccessToken)
	client := centrifuge.NewJsonClient(DA_CENT_WEBSOCKET_URL, centrifuge.Config{
		PingInterval:         centrifuge.DefaultPingInterval,
		ReadTimeout:          centrifuge.DefaultReadTimeout,
		WriteTimeout:         centrifuge.DefaultWriteTimeout,
		HandshakeTimeout:     centrifuge.DefaultHandshakeTimeout,
		PrivateChannelPrefix: centrifuge.DefaultPrivateChannelPrefix,
		Header:               header,
		Name:                 centrifuge.DefaultName,
	})
	client.SetToken(cent.SocketToken)
	//defer func() { _ = client.Close() }()
	doneCh := make(chan error, 1)
	handler := &testEventHandler{
		onConnect: func(c *centrifuge.Client, e centrifuge.ConnectEvent) {
			log.Printf("%#v\n", e)
			if e.ClientID == "" {
				doneCh <- fmt.Errorf("wrong client ID value")
				return
			}
			close(doneCh)
			cent.ClientID = e.ClientID
			cent.Channel = "$alerts:donation_" + strconv.FormatInt(cent.UserID, 10)
			err := cent.GetChannelToken()
			if err != nil {
				log.Fatalln(err)
			}
		},
		onDisconnect: func(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
			log.Printf("%#v\n", e)
		},
		onMessage: func(c *centrifuge.Client, e centrifuge.MessageEvent) {
			log.Printf("%#v\n", e)
		},
		onPrivateSub: func(c *centrifuge.Client, e centrifuge.PrivateSubEvent) (string, error) {
			log.Printf("%#v\n", e)
			return cent.ChannelToken, nil
		},
	}
	client.OnConnect(handler)
	client.OnDisconnect(handler)
	client.OnMessage(handler)
	client.OnPrivateSub(handler)
	_ = client.Connect()
	select {
	case err := <-doneCh:
		if err != nil {
			log.Printf("finish with error: %v", err)
		}
	case <-time.After(5 * time.Second):
		log.Println("expecting successful connect")
	}

	cent.Client = client

	go func() {
		select {
		case <-ctx.Done():
			log.Fatal(client.Close())
		}
	}()
}

func (cent *Centrifuge) Subscribe(ctx context.Context) {
	sub, err := cent.Client.NewSubscription(cent.Channel)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(sub.Channel())
	handler := &testPublishHandler{
		onPublish: func(c *centrifuge.Subscription, e centrifuge.PublishEvent) {
			log.Printf("%#v\n", e)
		},
	}
	sub.OnPublish(handler)

	err = sub.Subscribe()
	if err != nil {
		return
	}

	go func() {
		select {
		case <-ctx.Done():
			log.Fatal(sub.Close())
		}
	}()
}

func (cent *Centrifuge) GetChannelToken() error {
	r := request.SubscribeChannel{
		Channels: []string{cent.Channel},
		Client:   cent.ClientID,
	}
	channelToken := response.SubscribeChannel{}
	resp, err := req.R().
		SetBody(r).
		SetResult(&channelToken).
		SetBearerAuthToken(cent.AccessToken).
		SetContentType("application/json").
		Post(DA_HOST + "/api/v1/centrifuge/subscribe")

	if err != nil || !resp.IsSuccess() {
		log.Println(err)
		return fiber.NewError(http.StatusInternalServerError, "cannot get channel token")
	}

	log.Printf("%#v\n", channelToken)

	cent.ChannelToken = channelToken.Channels[0]["token"]

	return nil
}
