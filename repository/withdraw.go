package repository

//CreateWithdrawMoney input db, model
func CreateWithdrawMoney(q Queryer, req *CreateWithdrawMoneyModel) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO withdraw_money
					(user_id, amount)
			 VALUES ($1, $2)
		  RETURNING id;
				`, req.UserID, req.Amount).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}

//CreateWithdrawMoneyModel is model table
type CreateWithdrawMoneyModel struct {
	UserID string
	Amount int64
}
