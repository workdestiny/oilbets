package entity

import "time"

//YingchobUserBet entity model
type YingchobUserBet struct {
	ID           string
	UserID       string
	FirstName    string
	LastName     string
	Price        int64
	Status       int
	StatusText   string
	Yingchob     int
	YingchobText string
	YingchobBot  int
	CreatedAt    time.Time
}

//Yingchob entity model
type Yingchob struct {
	ID        string
	PriceBet  int64
	Win       int
	Lose      int
	CreatedAt time.Time
}

//RequestYingchobBet model
type RequestYingchobBet struct {
	Yingchob []int `json:"yingchob"`
	Price    int64 `json:"price"`
	Number   int   `json:"number"`
}

//ResponseYc model
type ResponseYc struct {
	Yingchob []ResponseYingchob `json:"yingchob"`
	NoMoney  bool               `json:"noMoney"`
	Price    int64              `json:"price"`
	Wallet   int64              `json:"wallet"`
	Bonus    int64              `json:"bonus"`
	Status   int                `json:"status"`
}

//ResponseYingchob model
type ResponseYingchob struct {
	Status      int `json:"status"`
	Yingchob    int `json:"yingchob"`
	YingchobBot int `json:"yingchob_bot"`
}
