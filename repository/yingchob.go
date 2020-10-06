package repository

import (
	"log"

	"github.com/workdestiny/oilbets/entity"
)

//ListYingchobUserBet list user id
func ListYingchobUserBet(q Queryer, limit int) ([]entity.YingchobUserBet, error) {

	rows, err := q.Query(`
			SELECT yc.id, yc.user_id, yc.price, yc.status,
					yc.yingchob, yc.created_at, users.firstname, users.lastname
			  FROM ying_chob_bet as yc
		 LEFT JOIN users
				ON users.id = yc.user_id
		  ORDER BY yc.created_at DESC LIMIT $1
		`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ycUsers []entity.YingchobUserBet

	for rows.Next() {

		var ycUser entity.YingchobUserBet
		err := rows.Scan(&ycUser.ID, &ycUser.UserID, &ycUser.Price, &ycUser.Status,
			&ycUser.Yingchob, &ycUser.CreatedAt, &ycUser.FirstName, &ycUser.LastName)
		if err != nil {
			return nil, err
		}

		ycUsers = append(ycUsers, ycUser)
	}

	return ycUsers, nil
}

//GetYingchob input price, number_bet
func GetYingchob(q Queryer, price int64, numberBet int) (entity.Yingchob, error) {

	var fb entity.Yingchob
	err := q.QueryRow(`
		SELECT id, win, lose
		  FROM ying_chob
		 WHERE price_bet = $1 AND number_bet = $2
	  ORDER BY created_at DESC;
		`, price, numberBet).Scan(&fb.ID, &fb.Win, &fb.Lose)

	if err != nil {
		return fb, err
	}

	return fb, nil
}

//CreateYingchob input db, model
func CreateYingchob(q Queryer, req *CreateYingchobModel) (string, error) {
	log.Println(req.Number)
	var id string
	err := q.QueryRow(`
		INSERT INTO ying_chob
					(price_bet, win, lose, number_bet)
			 VALUES ($1, $2, $3, $4)
		  RETURNING id;
				`, req.Pricebet, req.Win, req.Lose, req.Number).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//CreateYingchobModel a
type CreateYingchobModel struct {
	Pricebet int64
	Win      int
	Lose     int
	Number   int
}

//CreateYingchobBetModel model create front_back_bet table
type CreateYingchobBetModel struct {
	UserID   string
	Price    int64
	Status   int
	Yingchob int
}

//CreateYingchobBet input db, model
func CreateYingchobBet(q Queryer, req *CreateYingchobBetModel) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO ying_chob_bet
					(user_id, price, status, ying_chob)
			 VALUES ($1, $2, $3, $4)
		  RETURNING id;
				`, req.UserID, req.Price, req.Status, req.Yingchob).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//UpdateYingchob input wallet, bonus
func UpdateYingchob(q Queryer, id string, win int, lose int) error {

	_, err := q.Exec(`
		UPDATE ying_chob
		   SET win = $1, lose = $2
		 WHERE id = $3;
		 `, win, lose, id)
	if err != nil {
		return err
	}

	return nil
}
