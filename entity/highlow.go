package entity

import "time"

//HighlowBet entity model
type HighlowBet struct {
	ID        string
	Dice      int
	Dice2     int
	Dice3     int
	Total     int64
	Status    bool
	Open      bool
	CreatedAt time.Time
	High      int64
	Low       int64
	N11       int64
	N1        int64
	N2        int64
	N3        int64
	N4        int64
	N5        int64
	N6        int64
	N12       int64
	N13       int64
	N14       int64
	N15       int64
	N16       int64
	N23       int64
	N24       int64
	N25       int64
	N26       int64
	N34       int64
	N35       int64
	N36       int64
	N45       int64
	N46       int64
	N56       int64
	High1     int64
	High2     int64
	High3     int64
	High4     int64
	High5     int64
	High6     int64
	Low1      int64
	Low2      int64
	Low3      int64
	Low4      int64
	Low5      int64
	Low6      int64
}

//HighlowUserBet is
type HighlowUserBet struct {
	ID        string
	UserID    string
	FirstName string
	LastName  string
	Price     int64
	Bet       int64
	Status    bool
	Total     int64
	CreatedAt time.Time
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
