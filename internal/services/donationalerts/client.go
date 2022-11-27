package donationalerts

import (
	"github.com/gofiber/fiber/v2"
	"github.com/imroc/req/v3"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	ID          string
	Secret      string
	RedirectUrl string
}

const (
	DA_HOST         string = "https://www.donationalerts.com"
	DA_ACCESS_SCOPE string = "oauth-user-show oauth-donation-subscribe"
)

func (c *Client) AuthLink() string {
	url, _ := url.Parse(DA_HOST + "/oauth/authorize")

	q := url.Query()
	q.Add("client_id", c.ID)
	q.Add("response_type", "code")
	q.Add("scope", DA_ACCESS_SCOPE)
	q.Add("redirect_uri", c.RedirectUrl)
	url.RawQuery = q.Encode()

	str := url.String()

	return str
}

type TokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func (c *Client) NewToken(code string) (TokensResponse, error) {
	data := TokensResponse{}
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
		return TokensResponse{}, fiber.NewError(http.StatusInternalServerError, "cannot get tokens")
	}
	if !resp.IsSuccess() {
		return TokensResponse{}, fiber.NewError(401, "seems like code is invalid")
	}

	return data, nil
}

func (c *Client) RefreshToken(rToken string) (TokensResponse, error) {
	data := TokensResponse{}
	resp, err := req.R().
		SetFormData(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": rToken,
			"client_id":     c.ID,
			"client_secret": c.Secret,
			"scope":         DA_ACCESS_SCOPE,
		}).
		SetResult(&data).
		SetContentType("application/x-www-form-urlencoded").
		Post(DA_HOST + "/oauth/token")

	if err != nil {
		log.Println(err)
		return TokensResponse{}, fiber.NewError(http.StatusInternalServerError, "cannot get tokens")
	}
	if !resp.IsSuccess() {
		return TokensResponse{}, fiber.NewError(401, "seems like refresh_token is invalid")
	}

	return data, nil
}

type ProfileResponse struct {
	Data struct {
		ID          int64  `json:"id"`
		Code        string `json:"code"`
		Name        string `json:"name"`
		Avatar      string `json:"avatar"`
		Email       string `json:"email"`
		SocketToken string `json:"socket_connection_token"`
	} `json:"data"`
}

func (c *Client) Profile(token string) (ProfileResponse, error) {
	profile := ProfileResponse{}
	profileResp, err := req.R().
		SetResult(&profile).
		SetBearerAuthToken(token).
		Get(DA_HOST + "/api/v1/user/oauth")

	if err != nil || !profileResp.IsSuccess() {
		log.Println(err)
		return ProfileResponse{}, fiber.NewError(http.StatusInternalServerError, "cannot get profile")
	}

	return profile, nil
}
