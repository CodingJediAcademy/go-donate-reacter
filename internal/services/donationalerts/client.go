package donationalerts

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/browser"
	"go-donate-reacter/internal/services/donationalerts/response"
	"golang.org/x/oauth2"
	"io"
	"log"
	"net/http"
	"strconv"
)

const (
	DA_HOST string = "https://www.donationalerts.com"
)

type Client struct {
	ctx        context.Context
	conf       *oauth2.Config
	httpClient *http.Client
	Token      *oauth2.Token
}

func NewClient(
	ctx context.Context,
	id string,
	secret string,
	redirectUrl string,
	codeChan <-chan string,
	token *oauth2.Token,
) Client {
	conf := &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   DA_HOST + "/oauth/authorize",
			TokenURL:  DA_HOST + "/oauth/token",
			AuthStyle: oauth2.AuthStyleInParams,
		},
		RedirectURL: redirectUrl,
		Scopes:      []string{"oauth-user-show", "oauth-donation-subscribe", "oauth-donation-index"},
	}

	client := Client{
		ctx:  ctx,
		conf: conf,
	}

	if token.AccessToken == "" {
		token = client.Auth(codeChan)
	}

	httpClient := conf.Client(ctx, token)
	client.httpClient = httpClient
	client.Token = token

	return client
}

func (c *Client) Auth(codeChan <-chan string) *oauth2.Token {

	url := c.conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	if err := browser.OpenURL(url); err != nil {
		fmt.Printf("Can't open your browser. Open it yourself: %v\n", url)
	}

	fmt.Println("Waiting for code...")
	code := <-codeChan
	fmt.Println("Got code! Now retrieving token...")
	tok, err := c.conf.Exchange(c.ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Got token! Ready to run!")
	return tok
}

func (c *Client) Profile() (*response.Profile, error) {
	profile := response.Profile{}
	resp, err := c.httpClient.Get(DA_HOST + "/api/v1/user/oauth")
	if err != nil {
		log.Fatalln(err)
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err = json.Unmarshal(body, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (c *Client) Donations(page int) (*response.Donations, error) {
	var strPage string
	var data response.Donations
	strPage = "?page=" + strconv.Itoa(page)
	resp, err := c.httpClient.Get(DA_HOST + "/api/v1/alerts/donations" + strPage)
	if err != nil {
		log.Fatalln(err)
	}
	defer func(Body io.ReadCloser) { _ = Body.Close() }(resp.Body)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
