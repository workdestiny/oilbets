package repository

import (
	"time"

	"github.com/workdestiny/convbox"
	"github.com/workdestiny/oilbets/entity"
)

//GetLiveHighlow input price
func GetLiveHighlow(q Queryer, id string) (entity.HighlowBet, error) {

	var hl entity.HighlowBet
	var high, low, n11, n1, n2, n3, n4, n5, n6, n12, n13, n14, n15, n16, n23, n24, n25, n26, n34, n35, n36, n45, n46, n56 int
	var h1, h2, h3, h4, h5, h6 int
	var l1, l2, l3, l4, l5, l6 int
	err := q.QueryRow(`
		SELECT id, dice, total, created_at,
				high, low, n11, n1,
				n2, n3, n4, n5,
				n6, n12, n13, n14,
				n15, n16, n23, n24,
				n25, n26, n34, n35,
				n36, n45, n46, n56,
				high_1, high_2, high_3, high_4,
				high_5, high_6, low_1, low_2,
				low_3, low_4, low_5, low_6,
				open, dice_2, dice_3, updated_at,
				status
		  FROM high_low
		 WHERE id = $1
	  ORDER BY created_at DESC;
		`, id).Scan(&hl.ID, &hl.Dice, &hl.Total, &hl.CreatedAt,
		&high, &low, &n11, &n1,
		&n2, &n3, &n4, &n5,
		&n6, &n12, &n13, &n14,
		&n15, &n16, &n23, &n24,
		&n25, &n26, &n34, &n35,
		&n36, &n45, &n46, &n56,
		&h1, &h2, &h3, &h4,
		&h5, &h6, &l1, &l2,
		&l3, &l4, &l5, &l6,
		&hl.Open, &hl.Dice2, &hl.Dice3, &hl.UpdatedAt,
		&hl.Status)

	if err != nil {
		return hl, err
	}

	hl.High = convbox.ShortNumber(high)
	hl.Low = convbox.ShortNumber(low)
	hl.N11 = convbox.ShortNumber(n11)
	hl.N1 = convbox.ShortNumber(n1)
	hl.N2 = convbox.ShortNumber(n2)
	hl.N3 = convbox.ShortNumber(n3)
	hl.N4 = convbox.ShortNumber(n4)
	hl.N5 = convbox.ShortNumber(n5)
	hl.N6 = convbox.ShortNumber(n6)
	hl.N12 = convbox.ShortNumber(n12)
	hl.N13 = convbox.ShortNumber(n13)
	hl.N14 = convbox.ShortNumber(n14)
	hl.N15 = convbox.ShortNumber(n15)
	hl.N16 = convbox.ShortNumber(n16)
	hl.N23 = convbox.ShortNumber(n23)
	hl.N24 = convbox.ShortNumber(n24)
	hl.N25 = convbox.ShortNumber(n25)
	hl.N26 = convbox.ShortNumber(n26)
	hl.N34 = convbox.ShortNumber(n34)
	hl.N35 = convbox.ShortNumber(n35)
	hl.N36 = convbox.ShortNumber(n36)
	hl.N45 = convbox.ShortNumber(n45)
	hl.N46 = convbox.ShortNumber(n46)
	hl.N56 = convbox.ShortNumber(n56)
	hl.High1 = convbox.ShortNumber(h1)
	hl.High2 = convbox.ShortNumber(h2)
	hl.High3 = convbox.ShortNumber(h3)
	hl.High4 = convbox.ShortNumber(h4)
	hl.High5 = convbox.ShortNumber(h5)
	hl.High6 = convbox.ShortNumber(h6)
	hl.Low1 = convbox.ShortNumber(l1)
	hl.Low2 = convbox.ShortNumber(l2)
	hl.Low3 = convbox.ShortNumber(l3)
	hl.Low4 = convbox.ShortNumber(l4)
	hl.Low5 = convbox.ShortNumber(l5)
	hl.Low6 = convbox.ShortNumber(l6)

	return hl, nil
}

//GetLiveHighlowID get id
func GetLiveHighlowID(q Queryer) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM high_low
		 WHERE status = false
	  ORDER BY created_at DESC;
		`).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//GetLiveOpenHighlowID get id
func GetLiveOpenHighlowID(q Queryer) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM high_low
		 WHERE status = false AND open = true
	  ORDER BY created_at DESC;
		`).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//CheckCountHighlowUserBet return count
func CheckCountHighlowUserBet(q Queryer, id string) (int, error) {
	var count int
	err := q.QueryRow(`
	SELECT COUNT(DISTINCT user_id)
	  FROM (SELECT user_id
			FROM high_low_bet
			WHERE highlow_id = $1) as high_low_bet;
	`, id).Scan(&count)
	return count, err
}

//ListHighlowUserBet list gap id
func ListHighlowUserBet(q Queryer, id string, limit int) ([]entity.HighlowUserBet, error) {

	rows, err := q.Query(`
			SELECT hl.id, hl.user_id, hl.price, hl.bet,
				   hl.created_at, users.firstname, users.lastname
			  FROM high_low_bet as hl
		 LEFT JOIN users
				ON users.id = hl.user_id
			 WHERE hl.highlow_id = $1
		  ORDER BY hl.created_at DESC LIMIT $2
		`, id, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hlUsers []entity.HighlowUserBet

	for rows.Next() {

		var hlUser entity.HighlowUserBet
		err := rows.Scan(&hlUser.ID, &hlUser.UserID, &hlUser.Price, &hlUser.Bet,
			&hlUser.CreatedAt, &hlUser.FirstName, &hlUser.LastName)
		if err != nil {
			return nil, err
		}

		hlUser.BetString = entity.HighlowBetString[entity.TypeHighlowBet(hlUser.Bet)]
		hlUsers = append(hlUsers, hlUser)
	}

	return hlUsers, nil
}

//ListHighlowMyBet list gap id
func ListHighlowMyBet(q Queryer, id string, userID string, limit int) ([]entity.HighlowUserBet, error) {

	rows, err := q.Query(`
			SELECT id, price, bet, created_at,
				   status, total
			  FROM high_low_bet
			 WHERE highlow_id = $1 AND user_id = $2
		  ORDER BY created_at DESC LIMIT $3
		`, id, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hlUsers []entity.HighlowUserBet

	for rows.Next() {

		var hlUser entity.HighlowUserBet
		err := rows.Scan(&hlUser.ID, &hlUser.Price, &hlUser.Bet, &hlUser.CreatedAt,
			&hlUser.Status, &hlUser.Total)
		if err != nil {
			return nil, err
		}

		hlUser.BetString = entity.HighlowBetString[entity.TypeHighlowBet(hlUser.Bet)]
		hlUsers = append(hlUsers, hlUser)
	}

	return hlUsers, nil
}

//GetHasMoney list gap id
func GetHasMoney(q Queryer, userID string) (int, error) {

	rows, err := q.Query(`
			SELECT total
			  FROM high_low_bet
			 WHERE status = true AND is_withdraw = false AND user_id = $1
		  ORDER BY created_at DESC
		`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var total int

	for rows.Next() {

		var money int
		err := rows.Scan(&money)
		if err != nil {
			return 0, err
		}

		total = total + money
	}

	return total, nil
}

//GetHighlow input price
func GetHighlow(q Queryer, id string) (bool, error) {

	var open bool
	err := q.QueryRow(`
		SELECT open
		  FROM high_low
		 WHERE id = $1 AND status = false
	  ORDER BY created_at DESC;
		`, id).Scan(&open)
	if err != nil {
		return open, err
	}

	return open, nil
}

//UpdateOpenHighlow input id
func UpdateOpenHighlow(q Queryer, id string) error {

	_, err := q.Exec(`
		UPDATE high_low
		   SET open = true, updated_at = $1
		 WHERE id = $2;
		 `, time.Now().UTC(), id)
	if err != nil {
		return err
	}

	return nil
}

//UpdateCloseHighlow input id
func UpdateCloseHighlow(q Queryer, id string, r1, r2, r3 int) error {

	_, err := q.Exec(`
		UPDATE high_low
		   SET status = true, dice = $1, dice_2 = $2, dice_3 = $3
		 WHERE id = $4;
		 `, r1, r2, r3, id)
	if err != nil {
		return err
	}

	return nil
}

//UpdateWinNumberHighlowBet input id
func UpdateWinNumberHighlowBet(q Queryer, id string, n int, sum int) error {

	_, err := q.Exec(`
		UPDATE high_low_bet
		   SET status = true, total = price * $1
		 WHERE highlow_id = $2 AND bet = $3;
		 `, sum, id, n)
	if err != nil {
		return err
	}

	return nil
}

//UpdateWinHLHighlowBet input id
func UpdateWinHLHighlowBet(q Queryer, id string, n int) error {

	_, err := q.Exec(`
		UPDATE high_low_bet
		   SET status = true, total = price * 2
		 WHERE highlow_id = $1 AND bet = $2;
		 `, id, n)
	if err != nil {
		return err
	}

	return nil
}

//UpdateWin11HighlowBet input id
func UpdateWin11HighlowBet(q Queryer, id string, n int) error {

	_, err := q.Exec(`
		UPDATE high_low_bet
		   SET status = true, total = price * 6
		 WHERE highlow_id = $1 AND bet = $2;
		 `, id, n)
	if err != nil {
		return err
	}

	return nil
}

//UpdateDuoNumberHighlowBet input id
func UpdateDuoNumberHighlowBet(q Queryer, id string, n int) error {

	_, err := q.Exec(`
		UPDATE high_low_bet
		   SET status = true, total = price * 5
		 WHERE highlow_id = $1 AND bet = $2;
		 `, id, n)
	if err != nil {
		return err
	}

	return nil
}

//UpdateHLNumberHighlowBet input id
func UpdateHLNumberHighlowBet(q Queryer, id string, n int) error {

	_, err := q.Exec(`
		UPDATE high_low_bet
		   SET status = true, total = price * 4
		 WHERE highlow_id = $1 AND bet = $2;
		 `, id, n)
	if err != nil {
		return err
	}

	return nil
}

//CreateHighlow input db, model
func CreateHighlow(q Queryer) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO high_low
					(total)
			 VALUES (0)
		  RETURNING id;
				`).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//CreateHighlowBet input db, model
func CreateHighlowBet(q Queryer, req *CreateHighlowBetModel) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO high_low_bet
					(user_id, highlow_id, price, bet)
			 VALUES ($1, $2, $3, $4)
		  RETURNING id;
				`, req.UserID, req.HighlowID, req.Price, req.Bet).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//UpdateTotalHighlow input map[bet]
func UpdateTotalHighlow(q Queryer, id string, price int64, arrBet map[entity.TypeHighlowBet]int64) error {

	_, err := q.Exec(`
		UPDATE high_low
		   SET total = total + $1, high = high + $2, low = low + $3, n11 = n11 + $4,
				   n1 = n1 + $5, n2 = n2 + $6, n3 = n3 + $7, n4 = n4 + $8, n5 = n5 + $9,
				   n6 = n6 + $10, n12 = n12 + $11, n13 = n13 + $12, n14 = n14 + $13,
				   n15 = n15 + $14, n16 = n16 + $15, n23 = n23 + $16, n24 = n24 + $17,
				   n25 = n25 + $18, n26 = n26 + $19, n34 = n34 + $20, n35 = n35 + $21,
				   n36 = n36 + $22, n45 = n45 + $23, n46 = n46 + $24, n56 = n56 + $25,
				   high_1 = high_1 + $26, high_2 = high_2 + $27, high_3 = high_3 + $28, high_4 = high_4 + $29,
				   high_5 = high_5 + $30, high_6 = high_6 + $31, low_1 = low_1 + $32, low_2 = low_2 + $33,
				   low_3 = low_3 + $34, low_4 = low_4 + $35, low_5 = low_5 + $36, low_6 = low_6 + $37
		 WHERE id = $38;
		 `, price, arrBet[0], arrBet[1], arrBet[2],
		arrBet[3], arrBet[4], arrBet[5], arrBet[6], arrBet[7],
		arrBet[8], arrBet[9], arrBet[10], arrBet[11],
		arrBet[12], arrBet[13], arrBet[14], arrBet[15],
		arrBet[16], arrBet[17], arrBet[18], arrBet[19],
		arrBet[20], arrBet[21], arrBet[22], arrBet[23],
		arrBet[24], arrBet[25], arrBet[26], arrBet[27],
		arrBet[28], arrBet[29], arrBet[30], arrBet[31],
		arrBet[32], arrBet[33], arrBet[34], arrBet[35],
		id)
	if err != nil {
		return err
	}

	return nil
}

//GetTotalWinHighlowMyBet list gap id
func GetTotalWinHighlowMyBet(q Queryer, userID string) (int64, error) {

	rows, err := q.Query(`
			SELECT total
			  FROM high_low_bet
			 WHERE user_id = $1 AND status = true AND is_withdraw = false
		  ORDER BY created_at DESC
		`, userID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var total int64

	for rows.Next() {

		var money int64
		err := rows.Scan(&money)
		if err != nil {
			return total, err
		}

		total = total + money
	}

	return total, nil
}

//UpdateHighlowIsWithdrawUser input wallet, bonus
func UpdateHighlowIsWithdrawUser(q Queryer, userID string) error {

	_, err := q.Exec(`
		UPDATE high_low_bet
		   SET is_withdraw = true
		 WHERE user_id = $1;
		 `, userID)
	if err != nil {
		return err
	}

	return nil
}

//UpdateWalletUserHighlow input wallet, bonus
func UpdateWalletUserHighlow(q Queryer, userID string, total int64) error {

	_, err := q.Exec(`
		UPDATE users
		   SET wallet = wallet + $1
		 WHERE id = $2;
		 `, total, userID)
	if err != nil {
		return err
	}

	return nil
}

//CreateHighlowBetModel is create model
type CreateHighlowBetModel struct {
	UserID    string
	HighlowID string
	Price     int64
	Bet       entity.TypeHighlowBet
}
