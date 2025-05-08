package core

type Client struct {
	ClientID string `db:"client_id"`
	Capacity int    `db:"capacity"`
	Tokens   int    `db:"tokens"`
}

type ClientRequest struct {
	ClientID string `json:"client_id"`
	Capacity int    `json:"capacity"`
	Tokens   int    `json:"tokens"`
}
