package app

import (
	"database/sql"
	"math/rand"
	"net/http"

	"github.com/acoshift/pgsql"
	"github.com/moonrhythm/hime"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/repository"
)

func frontbackBetGetHandler(ctx *hime.Context) error {

	usersBet, err := repository.ListFrontbackUserBet(db, 20)
	must(err)

	p := page(ctx)
	p["UsersBet"] = AddTextFrontback(usersBet)

	return ctx.View("app/frontback", p)
}

func ajaxFrontbackBetPostHandler(ctx *hime.Context) error {

	user := getUser(ctx)
	if user.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestFrontbackBet
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var res entity.ResponseFrontback
	res.Price = req.Price
	res.Frontback = !req.Frontback
	if (user.Wallet + user.Bonus) < req.Price {
		res.NoMoney = true
		return ctx.Status(http.StatusOK).JSON(&res)
	}

	frontback, err := repository.GetFrontback(db, req.Price)
	if err == sql.ErrNoRows {
		//new table
		var c repository.CreateFrontbackModel
		c.Pricebet = req.Price
		c.Lose = 1

		var cb repository.CreateFrontbackBetModel
		cb.UserID = user.ID
		cb.Frontback = req.Frontback
		cb.Price = req.Price

		//bet random win lose
		if rand.Intn(100) <= config.FrontbackWinrate {
			c.Lose = 0
			c.Win = 1
			cb.Status = true
			res.Frontback = req.Frontback
			res.Status = true
		}

		wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, cb.Status)
		res.Wallet = wallet
		res.Bonus = bonus

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			_, err := repository.CreateFrontback(tx, &c)
			if err != nil {
				return err
			}

			_, err = repository.CreateFrontbackBet(tx, &cb)
			if err != nil {
				return err
			}

			err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
			if err != nil {
				return err
			}

			return nil
		})
		must(err)

		return ctx.Status(http.StatusOK).JSON(&res)
	}
	must(err)

	total := frontback.Win + frontback.Lose
	if total >= 100 {
		//new table
		var c repository.CreateFrontbackModel
		c.Pricebet = req.Price
		c.Lose = 1

		var cb repository.CreateFrontbackBetModel
		cb.UserID = user.ID
		cb.Frontback = req.Frontback
		cb.Price = req.Price

		//bet random win lose
		if rand.Intn(100) <= config.FrontbackWinrate {
			c.Lose = 0
			c.Win = 1
			cb.Status = true
			res.Frontback = req.Frontback
			res.Status = true
		}

		wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, cb.Status)
		res.Wallet = wallet
		res.Bonus = bonus

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			_, err := repository.CreateFrontback(tx, &c)
			if err != nil {
				return err
			}

			_, err = repository.CreateFrontbackBet(tx, &cb)
			if err != nil {
				return err
			}

			err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
			if err != nil {
				return err
			}

			return nil
		})
		must(err)

		return ctx.Status(http.StatusOK).JSON(&res)
	}

	//update table
	if frontback.Win >= config.FrontbackWinrate {
		//loser 100% max limit user win
		var cb repository.CreateFrontbackBetModel
		cb.UserID = user.ID
		cb.Frontback = req.Frontback
		cb.Price = req.Price

		wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, false)
		res.Wallet = wallet
		res.Bonus = bonus

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.UpdateFrontback(tx, frontback.ID, frontback.Win, frontback.Lose+1)
			if err != nil {
				return err
			}

			_, err = repository.CreateFrontbackBet(tx, &cb)
			if err != nil {
				return err
			}

			err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
			if err != nil {
				return err
			}

			return nil
		})
		must(err)
		return ctx.Status(http.StatusOK).JSON(&res)
	}
	//update table
	frontback.Lose = frontback.Lose + 1

	var cb repository.CreateFrontbackBetModel
	cb.UserID = user.ID
	cb.Frontback = req.Frontback
	cb.Price = req.Price

	//bet random win lose
	if rand.Intn(100) <= config.FrontbackWinrate {
		frontback.Win = frontback.Win + 1
		frontback.Lose = frontback.Lose - 1
		cb.Status = true
		res.Frontback = req.Frontback
		res.Status = true

	}

	wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, cb.Status)
	res.Wallet = wallet
	res.Bonus = bonus

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.UpdateFrontback(tx, frontback.ID, frontback.Win, frontback.Lose)
		if err != nil {
			return err
		}

		_, err = repository.CreateFrontbackBet(tx, &cb)
		if err != nil {
			return err
		}

		err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.Status(http.StatusOK).JSON(&res)

}

//WalletAndBonus sum wallet and bonus (Return Wallet, Bonus)
func WalletAndBonus(price, wallet, bonus int64, status bool) (int64, int64) {

	if status {
		return wallet + price, bonus
	}

	sum := wallet - price
	wallet = wallet - price

	if sum > 0 {
		sum = 0
	}

	bonus = bonus + sum

	if wallet < 0 {
		wallet = 0
	}

	return wallet, bonus
}

//AddTextFrontback a
func AddTextFrontback(usersBet []entity.FrontbackUserBet) *[]entity.FrontbackUserBet {
	for i, v := range usersBet {
		if v.Status {
			usersBet[i].StatusText = "ชนะ"
		}
		if !v.Status {
			usersBet[i].StatusText = "แพ้"
		}
		if v.Frontback {
			usersBet[i].FrontbackText = "หัว"
		}
		if !v.Frontback {
			usersBet[i].FrontbackText = "ก้อย"
		}
	}
	return &usersBet
}
