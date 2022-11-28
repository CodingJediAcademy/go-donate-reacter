package donationalerts

import (
	"github.com/imroc/req/v3"
	"go-donate-reacter/internal/services/donationalerts/response"
	"strconv"

	"log"
	"net/url"
)

const (
	DA_HOST string = "https://www.donationalerts.com"
)

type Client struct {
	id          string
	secret      string
	redirectUrl string

	Scope string
	Token response.Token
}

func NewClient(
	id string,
	secret string,
	redirectUrl string,
) Client {
	return Client{
		id:          id,
		secret:      secret,
		redirectUrl: redirectUrl,
		Scope:       "oauth-user-show oauth-donation-subscribe oauth-donation-index",
	}
}

func (c *Client) AuthLink() string {
	url, _ := url.Parse(DA_HOST + "/oauth/authorize")

	q := url.Query()
	q.Add("client_id", c.id)
	q.Add("response_type", "code")
	q.Add("scope", c.Scope)
	q.Add("redirect_uri", c.redirectUrl)
	url.RawQuery = q.Encode()

	str := url.String()

	return str
}

func (c *Client) NewToken(code string) error {
	data := response.Token{}
	resp, err := req.R().
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     c.id,
			"client_secret": c.secret,
			"redirect_uri":  c.redirectUrl,
			"code":          code,
		}).
		SetResult(&data).
		SetContentType("application/x-www-form-urlencoded").
		Post(DA_HOST + "/oauth/token")

	if err != nil || !resp.IsSuccess() {
		log.Println(err, resp.Error())
		return resp.Err
	}

	c.Token = data
	return nil
}

func (c *Client) RefreshToken() error {
	data := response.Token{}
	resp, err := req.R().
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": c.Token.RefreshToken,
			"client_id":     c.id,
			"client_secret": c.secret,
			"scope":         c.Scope,
		}).
		SetResult(&data).
		SetContentType("application/x-www-form-urlencoded").
		Post(DA_HOST + "/oauth/Token")

	if err != nil || !resp.IsSuccess() {
		log.Println(err, resp.Error())
		return resp.Err
	}

	c.Token = data
	return nil
}

func (c *Client) Profile() (response.Profile, error) {
	profile := response.Profile{}
	resp, err := req.R().
		SetResult(&profile).
		SetBearerAuthToken(c.Token.AccessToken).
		Get(DA_HOST + "/api/v1/user/oauth")

	if err != nil || !resp.IsSuccess() {
		log.Println(err, resp.Error())
		return response.Profile{}, resp.Err
	}

	return profile, nil
}

func (c *Client) Donations(page int) (response.Donations, error) {
	var strPage string
	var data response.Donations
	strPage = "?page=" + strconv.Itoa(page)
	resp, err := req.R().
		SetResult(&data).
		SetBearerAuthToken(c.Token.AccessToken).
		Get(DA_HOST + "/api/v1/alerts/donations" + strPage)

	if err != nil || !resp.IsSuccess() {
		log.Println(err, resp.Error(), resp.Status)
		return response.Donations{}, resp.Err
	}

	return data, nil
}
