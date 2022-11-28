package request

type SubscribeChannel struct {
	Channels []string `json:"channels"`
	Client   string   `json:"client"`
}
