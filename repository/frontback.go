package repository

import (
	"github.com/workdestiny/oilbets/entity"
)

//ListFrontbackUserBet list gap id
func ListFrontbackUserBet(q Queryer, limit int) ([]entity.FrontbackUserBet, error) {

	rows, err := q.Query(`
			SELECT fb.id, fb.user_id, fb.price, fb.status,
				   fb.frontback, fb.created_at, users.firstname, users.lastname
			  FROM front_back_bet as fb
		 LEFT JOIN users
			    ON users.id = fb.user_id
		  ORDER BY fb.created_at DESC LIMIT $1
		`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fbUsers []entity.FrontbackUserBet

	for rows.Next() {

		var fbUser entity.FrontbackUserBet
		err := rows.Scan(&fbUser.ID, &fbUser.UserID, &fbUser.Price, &fbUser.Status,
			&fbUser.Frontback, &fbUser.CreatedAt, &fbUser.FirstName, &fbUser.LastName)
		if err != nil {
			return nil, err
		}

		fbUsers = append(fbUsers, fbUser)
	}

	return fbUsers, nil
}

//GetFrontback input price
func GetFrontback(q Queryer, price int64) (entity.Frontback, error) {

	var fb entity.Frontback
	err := q.QueryRow(`
		SELECT id, win, lose
		  FROM front_back
		 WHERE price_bet = $1
	  ORDER BY created_at DESC;
		`, price).Scan(&fb.ID, &fb.Win, &fb.Lose)

	if err != nil {
		return fb, err
	}

	return fb, nil
}

//CreateFrontback input db, model
func CreateFrontback(q Queryer, req *CreateFrontbackModel) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO front_back
					(price_bet, win, lose)
			 VALUES ($1, $2, $3)
		  RETURNING id;
				`, req.Pricebet, req.Win, req.Lose).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//CreateFrontbackBet input db, model
func CreateFrontbackBet(q Queryer, req *CreateFrontbackBetModel) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO front_back_bet
					(user_id, price, status, frontback)
			 VALUES ($1, $2, $3, $4)
		  RETURNING id;
				`, req.UserID, req.Price, req.Status, req.Frontback).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//UpdateFrontback input wallet, bonus
func UpdateFrontback(q Queryer, id string, win int, lose int) error {

	_, err := q.Exec(`
		UPDATE front_back
		   SET win = $1, lose = $2
		 WHERE id = $3;
		 `, win, lose, id)
	if err != nil {
		return err
	}

	return nil
}

//CreateFrontbackModel is model create front_back table
type CreateFrontbackModel struct {
	Pricebet int64
	Win      int
	Lose     int
}

//CreateFrontbackBetModel is model create front_back_bet table
type CreateFrontbackBetModel struct {
	UserID    string
	Price     int64
	Status    bool
	Frontback bool
}
