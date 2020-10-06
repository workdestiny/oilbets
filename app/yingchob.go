package app

import (
	"database/sql"
	"math/rand"
	"net/http"
	"time"

	"github.com/acoshift/pgsql"
	"github.com/moonrhythm/hime"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/repository"
)

func yingChobGetHandler(ctx *hime.Context) error {

	// usersBet, err := repository.ListYingchobUserBet(db, 20)
	// must(err)

	// p := page(ctx)
	// p["UsersBet"] = AddBotYingchob(usersBet)

	return ctx.View("app/yingchob", page(ctx))
}

//AddBotYingchob a
func AddBotYingchob(usersBet []entity.YingchobUserBet) *[]entity.YingchobUserBet {
	for _, v := range usersBet {
		//ค้อน
		if v.Yingchob == 0 {
			//แพ้
			if v.Status == 0 {
				v.YingchobBot = 2
				continue
			}
			//ชนะ
			if v.Status == 1 {
				v.YingchobBot = 1
				continue
			}
			//เสมอ
			if v.Status == 2 {
				v.YingchobBot = 0
				continue
			}
		}
		//กรรไกร
		if v.Yingchob == 1 {
			//แพ้
			if v.Status == 0 {
				v.YingchobBot = 0
				continue
			}
			//ชนะ
			if v.Status == 1 {
				v.YingchobBot = 2
				continue
			}
			//เสมอ
			if v.Status == 2 {
				v.YingchobBot = 1
				continue
			}
		}
		//กระดาษ
		if v.Yingchob == 2 {
			//แพ้
			if v.Status == 0 {
				v.YingchobBot = 1
				continue
			}
			//ชนะ
			if v.Status == 1 {
				v.YingchobBot = 0
				continue
			}
			//เสมอ
			if v.Status == 2 {
				v.YingchobBot = 2
				continue
			}
		}
	}
	return &usersBet
}

//SumBotYingchob a
func SumBotYingchob(status int, yingchob int) int {
	//ค้อน
	if yingchob == 0 {
		//แพ้
		if status == 0 {
			return 2
		}
		//ชนะ
		if status == 1 {
			return 1
		}
		//เสมอ
		if status == 2 {
			return 0
		}
	}
	//กรรไกร
	if yingchob == 1 {
		//แพ้
		if status == 0 {
			return 0
		}
		//ชนะ
		if status == 1 {
			return 2
		}
		//เสมอ
		if status == 2 {
			return 1
		}
	}
	//กระดาษ
	if yingchob == 2 {
		//แพ้
		if status == 0 {
			return 1
		}
		//ชนะ
		if status == 1 {
			return 0
		}
		//เสมอ
		if status == 2 {
			return 2
		}
	}
	return 0
}

func ajaxYingchobBetPostHandler(ctx *hime.Context) error {

	user := getUser(ctx)
	if user.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestYingchobBet
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	if len(req.Yingchob) != req.Number {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var res entity.ResponseYc
	res.Price = req.Price
	res.Wallet = user.Wallet
	res.Bonus = user.Bonus
	//check wallet user
	if (user.Wallet + user.Bonus) < req.Price {
		res.NoMoney = true
		return ctx.Status(http.StatusOK).JSON(&res)
	}

	//set random
	rand.Seed(time.Now().UnixNano())
	// set win rate
	winRate := config.YingchobWinrate
	drawRate := config.YingchobDrawrate
	if req.Number == 2 {
		res.Price = req.Price * config.YingchobMuti2
		winRate = config.YingchobWinrate21
		drawRate = config.YingchobDrawrate21
	}
	if req.Number == 3 {
		res.Price = req.Price * config.YingchobMuti3
		winRate = config.YingchobWinrate31
		drawRate = config.YingchobDrawrate31
	}

	yingchob, err := repository.GetYingchob(db, req.Price, req.Number)
	if err == sql.ErrNoRows {
		//new table
		var c repository.CreateYingchobModel
		c.Pricebet = req.Price
		c.Lose = 1
		c.Number = req.Number

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			//loop
			BetStatus := 1
			for i := 0; i < req.Number; i++ {

				var cb repository.CreateYingchobBetModel
				cb.UserID = user.ID
				cb.Yingchob = req.Yingchob[i]
				cb.Price = req.Price

				//bet random win lose
				roll := rand.Intn(100) + 1
				if req.Number == 2 {
					if i == 1 {
						winRate = config.YingchobWinrate22
						drawRate = config.YingchobDrawrate22
					}
				}

				if req.Number == 3 {
					if i == 1 {
						winRate = config.YingchobWinrate32
						drawRate = config.YingchobDrawrate32
					}
					if i == 2 {
						winRate = config.YingchobWinrate33
						drawRate = config.YingchobDrawrate33
					}
				}

				//win
				if roll <= winRate {
					cb.Status = 1
					if BetStatus == 1 {
						BetStatus = 1
					}
					if i > 0 && BetStatus != 1 {
						BetStatus = 0
					}

					var yc entity.ResponseYingchob
					yc.Yingchob = req.Yingchob[i]
					yc.Status = 1
					yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
					res.Yingchob = append(res.Yingchob, yc)
				}
				//draw
				if roll <= drawRate && roll > winRate {
					cb.Status = 2
					if BetStatus == 2 || i == 0 {
						BetStatus = 2
					}
					if i > 0 && BetStatus != 2 {
						BetStatus = 0
					}

					var yc entity.ResponseYingchob
					yc.Yingchob = req.Yingchob[i]
					yc.Status = 2
					yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
					res.Yingchob = append(res.Yingchob, yc)
				}
				//lose
				if roll > drawRate {
					cb.Status = 0
					BetStatus = 0

					var yc entity.ResponseYingchob
					yc.Yingchob = req.Yingchob[i]
					yc.Status = 0
					yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
					res.Yingchob = append(res.Yingchob, yc)
				}

				_, err = repository.CreateYingchobBet(tx, &cb)
				if err != nil {
					return err
				}

			}
			wStatus := false
			res.Status = 0
			if BetStatus == 0 {
				c.Lose = 1
				c.Win = 0
			}
			if BetStatus == 1 {
				c.Lose = 0
				c.Win = 1
				wStatus = true
				res.Status = 1
			}
			if BetStatus == 2 {
				c.Lose = 0
				c.Win = 0
				res.Status = 2
			}

			//end loop
			_, err := repository.CreateYingchob(tx, &c)
			if err != nil {
				return err
			}

			if BetStatus != 2 {
				muti := 1
				if req.Number == 2 {
					muti = config.YingchobMuti2
				}
				if req.Number == 3 {
					muti = config.YingchobMuti3
				}
				wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, wStatus, int64(muti))
				res.Wallet = wallet
				res.Bonus = bonus

				err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
				if err != nil {
					return err
				}
			}

			return nil
		})
		must(err)

		return ctx.Status(http.StatusOK).JSON(&res)
	}
	must(err)

	total := yingchob.Win + yingchob.Lose
	if total >= 100 {
		//new table
		var c repository.CreateYingchobModel
		c.Pricebet = req.Price
		c.Lose = 1
		c.Number = req.Number

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			//loop
			BetStatus := 1
			for i := 0; i < req.Number; i++ {

				var cb repository.CreateYingchobBetModel
				cb.UserID = user.ID
				cb.Yingchob = req.Yingchob[i]
				cb.Price = req.Price

				//bet random win lose
				roll := rand.Intn(100) + 1
				if req.Number == 2 {
					if i == 1 {
						winRate = config.YingchobWinrate22
						drawRate = config.YingchobDrawrate22
					}
				}

				if req.Number == 3 {
					if i == 1 {
						winRate = config.YingchobWinrate32
						drawRate = config.YingchobDrawrate32
					}
					if i == 2 {
						winRate = config.YingchobWinrate33
						drawRate = config.YingchobDrawrate33
					}
				}

				//win
				if roll <= winRate {
					cb.Status = 1
					if BetStatus == 1 {
						BetStatus = 1
					}
					if i > 0 && BetStatus != 1 {
						BetStatus = 0
					}

					var yc entity.ResponseYingchob
					yc.Yingchob = req.Yingchob[i]
					yc.Status = 1
					yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
					res.Yingchob = append(res.Yingchob, yc)
				}
				//draw
				if roll <= drawRate && roll > winRate {
					cb.Status = 2
					if BetStatus == 2 || i == 0 {
						BetStatus = 2
					}
					if i > 0 && BetStatus != 2 {
						BetStatus = 0
					}

					var yc entity.ResponseYingchob
					yc.Yingchob = req.Yingchob[i]
					yc.Status = 2
					yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
					res.Yingchob = append(res.Yingchob, yc)
				}
				//lose
				if roll > drawRate {
					cb.Status = 0
					BetStatus = 0

					var yc entity.ResponseYingchob
					yc.Yingchob = req.Yingchob[i]
					yc.Status = 0
					yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
					res.Yingchob = append(res.Yingchob, yc)
				}

				_, err = repository.CreateYingchobBet(tx, &cb)
				if err != nil {
					return err
				}
			}
			wStatus := false
			res.Status = 0
			if BetStatus == 0 {
				c.Lose = 1
				c.Win = 0
			}
			if BetStatus == 1 {
				c.Lose = 0
				c.Win = 1
				wStatus = true
				res.Status = 1
			}
			if BetStatus == 2 {
				c.Lose = 0
				c.Win = 0
				res.Status = 2
			}

			//end loop
			_, err := repository.CreateYingchob(tx, &c)
			if err != nil {
				return err
			}

			if BetStatus != 2 {
				muti := 1
				if req.Number == 2 {
					muti = config.YingchobMuti2
				}
				if req.Number == 3 {
					muti = config.YingchobMuti3
				}
				wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, wStatus, int64(muti))
				res.Wallet = wallet
				res.Bonus = bonus

				err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
				if err != nil {
					return err
				}
			}

			return nil
		})
		must(err)

		return ctx.Status(http.StatusOK).JSON(&res)
	}

	//update table
	if yingchob.Win >= config.YingchobLimitWin {
		//loser 100% max limit user win

		wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, false, int64(req.Number))
		res.Wallet = wallet
		res.Bonus = bonus
		res.Status = 0

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.UpdateYingchob(tx, yingchob.ID, yingchob.Win, yingchob.Lose+1)
			if err != nil {
				return err
			}

			for i := 0; i < req.Number; i++ {
				if i == req.Number-1 {
					var cb repository.CreateYingchobBetModel
					cb.UserID = user.ID
					cb.Yingchob = req.Yingchob[i]
					cb.Price = req.Price
					cb.Status = 0

					var yc entity.ResponseYingchob
					yc.Yingchob = req.Yingchob[i]
					yc.Status = 0
					yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
					res.Yingchob = append(res.Yingchob, yc)

					_, err = repository.CreateYingchobBet(tx, &cb)
					if err != nil {
						return err
					}
					continue
				}
				roll := rand.Intn(3)
				var cb repository.CreateYingchobBetModel
				cb.UserID = user.ID
				cb.Yingchob = req.Yingchob[i]
				cb.Price = req.Price
				cb.Status = roll

				var yc entity.ResponseYingchob
				yc.Yingchob = req.Yingchob[i]
				yc.Status = roll
				yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
				res.Yingchob = append(res.Yingchob, yc)

				_, err = repository.CreateYingchobBet(tx, &cb)
				if err != nil {
					return err
				}
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
	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		BetStatus := 1
		for i := 0; i < req.Number; i++ {

			var cb repository.CreateYingchobBetModel
			cb.UserID = user.ID
			cb.Yingchob = req.Yingchob[i]
			cb.Price = req.Price

			//bet random win lose
			roll := rand.Intn(100) + 1
			if req.Number == 2 {
				if i == 1 {
					winRate = config.YingchobWinrate22
					drawRate = config.YingchobDrawrate22
				}
			}

			if req.Number == 3 {
				if i == 1 {
					winRate = config.YingchobWinrate32
					drawRate = config.YingchobDrawrate32
				}
				if i == 2 {
					winRate = config.YingchobWinrate33
					drawRate = config.YingchobDrawrate33
				}
			}

			//win
			if roll <= winRate {
				cb.Status = 1
				if BetStatus == 1 {
					BetStatus = 1
				}
				if i > 0 && BetStatus != 1 {
					BetStatus = 0
				}

				var yc entity.ResponseYingchob
				yc.Yingchob = req.Yingchob[i]
				yc.Status = 1
				yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
				res.Yingchob = append(res.Yingchob, yc)
			}
			//draw
			if roll <= drawRate && roll > winRate {
				cb.Status = 2
				if BetStatus == 2 || i == 0 {
					BetStatus = 2
				}
				if i > 0 && BetStatus != 2 {
					BetStatus = 0
				}

				var yc entity.ResponseYingchob
				yc.Yingchob = req.Yingchob[i]
				yc.Status = 2
				yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
				res.Yingchob = append(res.Yingchob, yc)
			}
			//lose
			if roll > drawRate {
				cb.Status = 0
				BetStatus = 0
				var yc entity.ResponseYingchob
				yc.Yingchob = req.Yingchob[i]
				yc.Status = 0
				yc.YingchobBot = SumBotYingchob(yc.Status, yc.Yingchob)
				res.Yingchob = append(res.Yingchob, yc)
			}

			_, err = repository.CreateYingchobBet(tx, &cb)
			if err != nil {
				return err
			}
		}

		wStatus := false
		res.Status = 0
		if BetStatus == 1 {
			wStatus = true
			yingchob.Win = yingchob.Win + 1
			res.Status = 1
		}
		if BetStatus == 0 {
			yingchob.Lose = yingchob.Lose + 1
			res.Status = 0
		}
		if BetStatus == 2 {
			res.Status = 2
		}

		err = repository.UpdateYingchob(tx, yingchob.ID, yingchob.Win, yingchob.Lose)
		if err != nil {
			return err
		}

		if BetStatus != 2 {
			muti := 1
			if req.Number == 2 {
				muti = config.YingchobMuti2
			}
			if req.Number == 3 {
				muti = config.YingchobMuti3
			}
			wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, wStatus, int64(muti))
			res.Wallet = wallet
			res.Bonus = bonus

			err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
			if err != nil {
				return err
			}
		}

		return nil
	})
	must(err)

	return ctx.Status(http.StatusOK).JSON(&res)
}
