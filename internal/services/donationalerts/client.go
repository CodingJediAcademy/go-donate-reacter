package donationalerts

import (
	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"

	"go-donate-reacter/internal/services/donationalerts/response"

	"log"
	"net/http"
	"net/url"
)

const (
	DA_HOST string = "https://www.donationalerts.com"
)

type Client struct {
	ID          string
	Secret      string
	RedirectUrl string
	Scope       string
}

func NewClient(
	id string,
	secret string,
	redirectUrl string,
) Client {
	return Client{
		ID:          id,
		Secret:      secret,
		RedirectUrl: redirectUrl,
		Scope:       "oauth-user-show oauth-donation-subscribe",
	}
}

func (c *Client) AuthLink() string {
	url, _ := url.Parse(DA_HOST + "/oauth/authorize")

	q := url.Query()
	q.Add("client_id", c.ID)
	q.Add("response_type", "code")
	q.Add("scope", c.Scope)
	q.Add("redirect_uri", c.RedirectUrl)
	url.RawQuery = q.Encode()

	str := url.String()

	return str
}

func (c *Client) NewToken(code string) (response.Token, error) {
	data := response.Token{}
	resp, err := req.R().
		SetFormData(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     c.ID,
			"client_secret": c.Secret,
			"redirect_uri":  c.RedirectUrl,
			"code":          code,
		}).
		SetResult(&data).
		SetContentType("application/x-www-form-urlencoded").
		Post(DA_HOST + "/oauth/token")

	if err != nil {
		log.Println(err)
		return response.Token{}, fiber.NewError(http.StatusInternalServerError, "cannot get tokens")
	}
	if !resp.IsSuccess() {
		return response.Token{}, fiber.NewError(401, "seems like code is invalid")
	}

	return data, nil
}

func (c *Client) RefreshToken(rToken string) (response.Token, error) {
	data := response.Token{}
	resp, err := req.R().
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": rToken,
			"client_id":     c.ID,
			"client_secret": c.Secret,
			"scope":         c.Scope,
		}).
		SetResult(&data).
		SetContentType("application/x-www-form-urlencoded").
		Post(DA_HOST + "/oauth/token")

	if err != nil {
		log.Println(err)
		return response.Token{}, fiber.NewError(http.StatusInternalServerError, "cannot get tokens")
	}
	if !resp.IsSuccess() {
		return response.Token{}, fiber.NewError(401, "seems like refresh_token is invalid")
	}

	return data, nil
}

func (c *Client) Profile(token string) (response.Profile, error) {
	profile := response.Profile{}
	profileResp, err := req.R().
		SetResult(&profile).
		SetBearerAuthToken(token).
		Get(DA_HOST + "/api/v1/user/oauth")

	if err != nil || !profileResp.IsSuccess() {
		log.Println(err)
		return response.Profile{}, fiber.NewError(http.StatusInternalServerError, "cannot get profile")
	}

	return profile, nil
}
