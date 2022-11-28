package response

type SubscribeChannel struct {
	Channels []map[string]string `json:"channels"`
}
