package entity

import "time"

//FrontbackUserBet entity model
type FrontbackUserBet struct {
	ID            string
	UserID        string
	FirstName     string
	LastName      string
	Price         int64
	Status        bool
	StatusText    string
	Frontback     bool
	FrontbackText string
	CreatedAt     time.Time
}

//Frontback entity model
type Frontback struct {
	ID        string
	PriceBet  int64
	Win       int
	Lose      int
	CreatedAt time.Time
}

//RequestFrontbackBet model
type RequestFrontbackBet struct {
	Frontback bool  `json:"frontback"`
	Price     int64 `json:"price"`
}

//ResponseFrontback model
type ResponseFrontback struct {
	Status    bool  `json:"status"`
	Price     int64 `json:"price"`
	Wallet    int64 `json:"wallet"`
	Bonus     int64 `json:"bonus"`
	Frontback bool  `json:"frontback"`
	NoMoney   bool  `json:"noMoney"`
}
