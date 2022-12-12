package donationalerts

import (
	"github.com/centrifugal/centrifuge-go"
	"github.com/imroc/req/v3"
	"go-donate-reacter/internal/services/donationalerts/request"
	"go-donate-reacter/internal/services/donationalerts/response"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"strconv"
)

const (
	DA_CENT_WEBSOCKET_URL string = "wss://centrifugo.donationalerts.com/connection/websocket"
)

type Centrifuge struct {
	AccessToken string
	SocketToken string
	UserID      int64
	ClientID    string
	Channel     Channel
	Sub         *centrifuge.Subscription
	Client      *centrifuge.Client
}

type Channel struct {
	ID    string
	Token string
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
	handler := &testEventHandler{
		onConnect: func(c *centrifuge.Client, e centrifuge.ConnectEvent) {
			log.Println("[cent] connected to wss.")
			//log.Printf("[cent] %#v\n", e)
			cent.ClientID = e.ClientID
			if err := cent.GetChannelToken(); err != nil {
				log.Fatalln("[cent] fatal ", err)
			}
			cent.Sub = cent.Subscribe()
		},
		onDisconnect: func(c *centrifuge.Client, e centrifuge.DisconnectEvent) {
			log.Println("[cent] disconnected")
			//log.Printf("[cent] %#v\n", e)
		},
		onMessage: func(c *centrifuge.Client, e centrifuge.MessageEvent) {
			log.Println("[cent] onMessage")
			log.Printf("[cent] %#v\n", e)
		},
		onPrivateSub: func(c *centrifuge.Client, e centrifuge.PrivateSubEvent) (string, error) {
			log.Println("[cent] subscribing to private channel...")
			//log.Printf("[cent] %#v\n", e)
			return cent.Channel.Token, nil
		},
	}
	client.OnConnect(handler)
	client.OnDisconnect(handler)
	client.OnMessage(handler)
	client.OnPrivateSub(handler)

	if err := client.Connect(); err != nil {
		log.Fatalln("[cent] fatal ", err)
	}

	cent.Client = client

	go func() {
		select {
		case <-ctx.Done():
			log.Println("[cent] closing all connections")
			if cent.Sub != nil {
				if err := cent.Sub.Close(); err != nil {
					log.Println("[cent] ", err)
				}
			}
			if err := client.Close(); err != nil {
				log.Println("[cent] ", err)
			}
		}
	}()
}

func (cent *Centrifuge) Subscribe() *centrifuge.Subscription {
	sub, err := cent.Client.NewSubscription(cent.Channel.ID)
	if err != nil {
		log.Println("[cent] ", err)
	}

	handler := &testPublishHandler{
		onPublish: func(c *centrifuge.Subscription, e centrifuge.PublishEvent) {
			// TODO: main logic entry point
			log.Println("[cent] Donation received.")
			log.Printf("[cent] %#v\n", string(e.Data))
		},
	}
	sub.OnPublish(handler)

	if err = sub.Subscribe(); err != nil {
		log.Println("[cent] ", err)
	}

	log.Println("[cent] subscribed. Waiting for donations... :)")
	return sub
}

func (cent *Centrifuge) GetChannelToken() error {
	channel := "$alerts:donation_" + strconv.FormatInt(cent.UserID, 10)
	r := request.SubscribeChannel{
		Channels: []string{channel},
		Client:   cent.ClientID,
	}
	channelToken := response.SubscribeChannel{}
	resp, err := req.R().
		SetBody(r).
		SetResult(&channelToken).
		SetBearerAuthToken(cent.AccessToken).
		SetContentType("application/json").
		Post(DA_HOST + "/api/v1/centrifuge/subscribe")

	if err != nil {
		return err
	}

	if !resp.IsSuccess() {
		log.Println("[cent] ", resp.String())
		return resp.Err
	}

	cent.Channel.ID = channel
	cent.Channel.Token = channelToken.Channels[0]["token"]

	return nil
}
