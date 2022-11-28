package response

type Profile struct {
	Data struct {
		ID          int64  `json:"id"`
		Code        string `json:"code"`
		Name        string `json:"name"`
		Avatar      string `json:"avatar"`
		Email       string `json:"email"`
		SocketToken string `json:"socket_connection_token"`
	} `json:"data"`
}
