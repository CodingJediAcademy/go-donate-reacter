package response

type Publication struct {
	Id                   int         `json:"id"`
	Name                 string      `json:"name"`
	Username             string      `json:"username"`
	Message              string      `json:"message"`
	MessageType          string      `json:"message_type"`
	PayinSystem          PayinSystem `json:"payin_system"`
	Amount               float32     `json:"amount"`
	Currency             string      `json:"currency"`
	IsShown              int         `json:"is_shown"`
	AmountInUserCurrency float32     `json:"amount_in_user_currency"`
	RecipientName        string      `json:"recipient_name"`
	Recipient            Recipient   `json:"recipient"`
	CreatedAt            string      `json:"created_at"`
	CreatedAtTs          int         `json:"created_at_ts"`
	ShownAt              string      `json:"shown_at"`
	ShownAtTs            int         `json:"shown_at_ts"`
	Reason               string      `json:"reason"`
}

type PayinSystem struct {
	Title string `json:"title"`
}

type Recipient struct {
	UserId int    `json:"user_id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
