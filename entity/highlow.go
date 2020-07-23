package entity

import "time"

//HighlowBet entity model
type HighlowBet struct {
	ID        string    `json:"id"`
	Dice      int       `json:"dice"`
	Dice2     int       `json:"dice2"`
	Dice3     int       `json:"dice3"`
	Total     int64     `json:"total"`
	Status    bool      `json:"status"`
	Open      bool      `json:"open"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	High      string    `json:"high"`
	Low       string    `json:"low"`
	N11       string    `json:"n11"`
	N1        string    `json:"n1"`
	N2        string    `json:"n2"`
	N3        string    `json:"n3"`
	N4        string    `json:"n4"`
	N5        string    `json:"n5"`
	N6        string    `json:"n6"`
	N12       string    `json:"n12"`
	N13       string    `json:"n13"`
	N14       string    `json:"n14"`
	N15       string    `json:"n15"`
	N16       string    `json:"n16"`
	N23       string    `json:"n23"`
	N24       string    `json:"n24"`
	N25       string    `json:"n25"`
	N26       string    `json:"n26"`
	N34       string    `json:"n34"`
	N35       string    `json:"n35"`
	N36       string    `json:"n36"`
	N45       string    `json:"n45"`
	N46       string    `json:"n46"`
	N56       string    `json:"n56"`
	High1     string    `json:"high1"`
	High2     string    `json:"high2"`
	High3     string    `json:"high3"`
	High4     string    `json:"high4"`
	High5     string    `json:"high5"`
	High6     string    `json:"high6"`
	Low1      string    `json:"low1"`
	Low2      string    `json:"low2"`
	Low3      string    `json:"low3"`
	Low4      string    `json:"low4"`
	Low5      string    `json:"low5"`
	Low6      string    `json:"low6"`
}

//HighlowUserBet is
type HighlowUserBet struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Price     int64     `json:"price"`
	Bet       int64     `json:"bet"`
	BetString string    `json:"bet_string"`
	Status    bool      `json:"status"`
	Total     int64     `json:"total"`
	CreatedAt time.Time `json:"created_at"`
}

//RequestHighlowBet model
type RequestHighlowBet struct {
	Bet       TypeHighlowBet `json:"bet"`
	Price     int64          `json:"price"`
	HighlowID string         `json:"highlow_id"`
}

//ResponseHighlow model
type ResponseHighlow struct {
	Status    bool  `json:"status"`
	Price     int64 `json:"price"`
	Wallet    int64 `json:"wallet"`
	Bonus     int64 `json:"bonus"`
	Frontback bool  `json:"frontback"`
	NoMoney   bool  `json:"noMoney"`
}

type ResponseHighlowUpdate struct {
	Highlow   HighlowBet       `json:"highlow"`
	Countdown int              `json:"countdown"`
	CountUser int              `json:"count_user"`
	ListUser  []HighlowUserBet `json:"list_user"`
	ListMyBet []HighlowUserBet `json:"list_mybet"`
	HasMoney  int              `json:"has_money"`
}
type TypeHighlowBet int

const (
	//0
	High TypeHighlowBet = iota
	//1
	Low
	//2
	N11
	//3
	N1
	//4
	N2
	//5
	N3
	//6
	N4
	//7
	N5
	//8
	N6
	//9
	N12
	//10
	N13
	//11
	N14
	//12
	N15
	//13
	N16
	//14
	N23
	//15
	N24
	//16
	N25
	//17
	N26
	//18
	N34
	//19
	N35
	//20
	N36
	//21
	N45
	//22
	N46
	//23
	N56
	//24
	High1
	//25
	High2
	//26
	High3
	//27
	High4
	//28
	High5
	//29
	High6
	//30
	Low1
	//31
	Low2
	//32
	Low3
	//33
	Low4
	//34
	Low5
	//35
	Low6
)

var HighlowBetString = map[TypeHighlowBet]string{
	High:  "สูง",
	Low:   "ต่ำ",
	N11:   "11 ไฮโล",
	N1:    "1",
	N2:    "2",
	N3:    "3",
	N4:    "4",
	N5:    "5",
	N6:    "6",
	N12:   "โต้ด 1 2",
	N13:   "โต้ด 1 3",
	N14:   "โต้ด 1 4",
	N15:   "โต้ด 1 5",
	N16:   "โต้ด 1 6",
	N23:   "โต้ด 2 3",
	N24:   "โต้ด 2 4",
	N25:   "โต้ด 2 5",
	N26:   "โต้ด 2 6",
	N34:   "โต้ด 3 4",
	N35:   "โต้ด 3 5",
	N36:   "โต้ด 3 6",
	N45:   "โต้ด 4 5",
	N46:   "โต้ด 4 6",
	N56:   "โต้ด 5 6",
	High1: "สูง 1",
	High2: "สูง 2",
	High3: "สูง 3",
	High4: "สูง 4",
	High5: "สูง 5",
	High6: "สูง 6",
	Low1:  "ต่ำ 1",
	Low2:  "ต่ำ 2",
	Low3:  "ต่ำ 3",
	Low4:  "ต่ำ 4",
	Low5:  "ต่ำ 5",
	Low6:  "ต่ำ 6",
}

var HighlowBetRate = map[TypeHighlowBet]int64{
	High:  2,
	Low:   2,
	N11:   6,
	N1:    2,
	N2:    2,
	N3:    2,
	N4:    2,
	N5:    2,
	N6:    2,
	N12:   6,
	N13:   6,
	N14:   6,
	N15:   6,
	N16:   6,
	N23:   6,
	N24:   6,
	N25:   6,
	N26:   6,
	N34:   6,
	N35:   6,
	N36:   6,
	N45:   6,
	N46:   6,
	N56:   6,
	High1: 4,
	High2: 4,
	High3: 4,
	High4: 4,
	High5: 4,
	High6: 4,
	Low1:  4,
	Low2:  4,
	Low3:  4,
	Low4:  4,
	Low5:  4,
	Low6:  4,
}
