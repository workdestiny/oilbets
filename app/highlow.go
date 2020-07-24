package app

import (
	"database/sql"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/acoshift/pgsql"
	"github.com/moonrhythm/hime"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/repository"
)

func highlowBetGetHandler(ctx *hime.Context) error {

	id, err := repository.GetLiveHighlowID(db)
	must(err)

	return ctx.RedirectTo("highlow.get", id)
}

func getHighlowBetGetHandler(ctx *hime.Context) error {
	highlowID := getParams(ctx, "highlowID")
	//ข้อมูลกระดาน
	highlow, err := repository.GetLiveHighlow(db, highlowID)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	//รายการผู้เล่น
	listUser, err := repository.ListHighlowUserBet(db, highlow.ID, 20)
	must(err)

	//รายการของฉัน
	listMyBet, err := repository.ListHighlowMyBet(db, highlow.ID, getUserID(ctx), 20)
	must(err)

	//เช็คจำนวนเงินในระบบ hl
	hMoney, err := repository.GetHasMoney(db, getUserID(ctx))
	must(err)

	p := page(ctx)
	p["Highlow"] = highlow
	p["ListUser"] = listUser
	p["ListMyBet"] = listMyBet
	p["HasMoney"] = hMoney
	p["User"] = getUser(ctx)

	return ctx.View("app/highlow", p)
}

func highlowDuration(t time.Time, open bool) int {
	if !open {
		return config.HighlowCountdown
	}
	now := time.Now().UTC()
	iv := config.HighlowCountdown - now.Sub(t).Seconds()
	if iv <= 0 {
		iv = 0
	}
	return int(iv)
}

func ajaxHighlowBetPostHandler(ctx *hime.Context) error {

	user := getUser(ctx)
	if user.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestHighlowBet
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var res entity.ResponseFrontback
	res.Price = req.Price

	//ตรวจสอบจำนวน wallet คงเหลือ
	if (user.Wallet + user.Bonus) < req.Price {
		res.Wallet = user.Wallet
		res.Bonus = user.Bonus
		res.NoMoney = true
		return ctx.Status(http.StatusOK).JSON(&res)
	}

	open, err := repository.GetHighlow(db, req.HighlowID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	if !open {
		//open เปิดกระดาน
		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err := repository.UpdateOpenHighlow(tx, req.HighlowID)
			if err != nil {
				return err
			}
			return nil
		})
		must(err)

		//rutine 5 min สั่งให้ bot เริ่มนับถอยหลัง และ ออกผลรางวัล
		go BotAlgorithmHighlowBet(req.HighlowID)
	}

	wallet, bonus := WalletAndBonus(req.Price, user.Wallet, user.Bonus, false)
	res.Wallet = wallet
	res.Bonus = bonus

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		//insert new bet
		_, err = repository.CreateHighlowBet(tx, &repository.CreateHighlowBetModel{
			UserID:    user.ID,
			HighlowID: req.HighlowID,
			Price:     req.Price,
			Bet:       req.Bet,
		})
		if err != nil {
			return err
		}

		var arrBet = map[entity.TypeHighlowBet]int64{}
		arrBet[req.Bet] = req.Price

		//update total highlow and sum type bet
		err = repository.UpdateTotalHighlow(tx, req.HighlowID, req.Price, arrBet)
		if err != nil {
			return err
		}

		//update wallet user
		err = repository.UpdateWalletAndBonusUser(tx, user.ID, wallet, bonus)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)
	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxHighlowBetUpdatePostHandler(ctx *hime.Context) error {

	user := getUser(ctx)
	if user.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestHighlowBet
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	highlow, err := repository.GetLiveHighlow(db, req.HighlowID)
	if err == sql.ErrNoRows {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}
	must(err)

	var res entity.ResponseHighlowUpdate
	//เวลา
	countdown := highlowDuration(highlow.UpdatedAt, highlow.Open)

	//จำนวนผู้เล่น
	count, err := repository.CheckCountHighlowUserBet(db, highlow.ID)
	must(err)

	//รายการผู้เล่น
	listUser, err := repository.ListHighlowUserBet(db, highlow.ID, 20)
	must(err)

	//รายการเล่น
	listMyBet, err := repository.ListHighlowMyBet(db, highlow.ID, user.ID, 20)
	must(err)

	//เช็คจำนวนเงินในระบบ hl
	hMoney, err := repository.GetHasMoney(db, user.ID)
	must(err)

	res.Highlow = highlow
	res.Countdown = countdown
	res.CountUser = count
	res.ListUser = listUser
	res.ListMyBet = listMyBet
	res.HasMoney = hMoney

	return ctx.Status(http.StatusOK).JSON(&res)
}

//BotAlgorithmHighlowBet bot hl
func BotAlgorithmHighlowBet(highlowID string) {
	time.Sleep(config.HighlowCountdown * time.Second)

	//ออกผล
	rand.Seed(time.Now().UnixNano())
	r1 := rand.Intn(6) + 1
	r2 := rand.Intn(6) + 1
	r3 := rand.Intn(6) + 1

	//ปิดกระดาน
	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.UpdateCloseHighlow(tx, highlowID, r1, r2, r3)
		if err != nil {
			return err
		}
		return nil
	})
	must(err)

	log.Println(r1, r2, r3)

	//ตรวจผลรางวัล
	var lucky []int
	//สูง ต่ำ 11
	sum := r1 + r2 + r3
	if sum == 11 {
		lucky = append(lucky, 2)
	}
	if sum < 11 {
		lucky = append(lucky, 1)
		//ต่ำตรงเลข
		if r1 == 1 || r2 == 1 || r3 == 1 {
			lucky = append(lucky, 30)
		}
		if r1 == 2 || r2 == 2 || r3 == 2 {
			lucky = append(lucky, 31)
		}
		if r1 == 3 || r2 == 3 || r3 == 3 {
			lucky = append(lucky, 32)
		}
		if r1 == 4 || r2 == 4 || r3 == 4 {
			lucky = append(lucky, 33)
		}
		if r1 == 5 || r2 == 5 || r3 == 5 {
			lucky = append(lucky, 34)
		}
		if r1 == 6 || r2 == 6 || r3 == 6 {
			lucky = append(lucky, 35)
		}
	}
	if sum > 11 {
		lucky = append(lucky, 0)
		//สูงตรงเลข
		if r1 == 1 || r2 == 1 || r3 == 1 {
			lucky = append(lucky, 24)
		}
		if r1 == 2 || r2 == 2 || r3 == 2 {
			lucky = append(lucky, 25)
		}
		if r1 == 3 || r2 == 3 || r3 == 3 {
			lucky = append(lucky, 26)
		}
		if r1 == 4 || r2 == 4 || r3 == 4 {
			lucky = append(lucky, 27)
		}
		if r1 == 5 || r2 == 5 || r3 == 5 {
			lucky = append(lucky, 28)
		}
		if r1 == 6 || r2 == 6 || r3 == 6 {
			lucky = append(lucky, 29)
		}
	}

	//ตรงเลข
	if r1 == 1 || r2 == 1 || r3 == 1 {
		lucky = append(lucky, 3)
	}
	if r1 == 2 || r2 == 2 || r3 == 2 {
		lucky = append(lucky, 4)
	}
	if r1 == 3 || r2 == 3 || r3 == 3 {
		lucky = append(lucky, 5)
	}
	if r1 == 4 || r2 == 4 || r3 == 4 {
		lucky = append(lucky, 6)
	}
	if r1 == 5 || r2 == 5 || r3 == 5 {
		lucky = append(lucky, 7)
	}
	if r1 == 6 || r2 == 6 || r3 == 6 {
		lucky = append(lucky, 8)
	}

	//โต๊ด
	if r1 == 1 || r2 == 1 || r3 == 1 {
		if r1 == 2 || r2 == 2 || r3 == 2 {
			lucky = append(lucky, 9)
		}
		if r1 == 3 || r2 == 3 || r3 == 3 {
			lucky = append(lucky, 10)
		}
		if r1 == 4 || r2 == 4 || r3 == 4 {
			lucky = append(lucky, 11)
		}
		if r1 == 5 || r2 == 5 || r3 == 5 {
			lucky = append(lucky, 12)
		}
		if r1 == 6 || r2 == 6 || r3 == 6 {
			lucky = append(lucky, 13)
		}
	}

	if r1 == 2 || r2 == 2 || r3 == 2 {
		if r1 == 1 || r2 == 1 || r3 == 1 {
			lucky = append(lucky, 9)
		}
		if r1 == 3 || r2 == 3 || r3 == 3 {
			lucky = append(lucky, 14)
		}
		if r1 == 4 || r2 == 4 || r3 == 4 {
			lucky = append(lucky, 15)
		}
		if r1 == 5 || r2 == 5 || r3 == 5 {
			lucky = append(lucky, 16)
		}
		if r1 == 6 || r2 == 6 || r3 == 6 {
			lucky = append(lucky, 17)
		}
	}

	if r1 == 3 || r2 == 3 || r3 == 3 {
		if r1 == 1 || r2 == 1 || r3 == 1 {
			lucky = append(lucky, 10)
		}
		if r1 == 2 || r2 == 2 || r3 == 2 {
			lucky = append(lucky, 14)
		}
		if r1 == 4 || r2 == 4 || r3 == 4 {
			lucky = append(lucky, 18)
		}
		if r1 == 5 || r2 == 5 || r3 == 5 {
			lucky = append(lucky, 19)
		}
		if r1 == 6 || r2 == 6 || r3 == 6 {
			lucky = append(lucky, 20)
		}
	}

	if r1 == 4 || r2 == 4 || r3 == 4 {
		if r1 == 1 || r2 == 1 || r3 == 1 {
			lucky = append(lucky, 11)
		}
		if r1 == 2 || r2 == 2 || r3 == 2 {
			lucky = append(lucky, 15)
		}
		if r1 == 3 || r2 == 3 || r3 == 3 {
			lucky = append(lucky, 19)
		}
		if r1 == 5 || r2 == 5 || r3 == 5 {
			lucky = append(lucky, 21)
		}
		if r1 == 6 || r2 == 6 || r3 == 6 {
			lucky = append(lucky, 22)
		}
	}

	if r1 == 5 || r2 == 5 || r3 == 5 {
		if r1 == 1 || r2 == 1 || r3 == 1 {
			lucky = append(lucky, 12)
		}
		if r1 == 2 || r2 == 2 || r3 == 2 {
			lucky = append(lucky, 16)
		}
		if r1 == 3 || r2 == 3 || r3 == 3 {
			lucky = append(lucky, 19)
		}
		if r1 == 4 || r2 == 4 || r3 == 4 {
			lucky = append(lucky, 21)
		}
		if r1 == 6 || r2 == 6 || r3 == 6 {
			lucky = append(lucky, 23)
		}
	}

	if r1 == 6 || r2 == 6 || r3 == 6 {
		if r1 == 1 || r2 == 1 || r3 == 1 {
			lucky = append(lucky, 13)
		}
		if r1 == 2 || r2 == 2 || r3 == 2 {
			lucky = append(lucky, 17)
		}
		if r1 == 3 || r2 == 3 || r3 == 3 {
			lucky = append(lucky, 20)
		}
		if r1 == 4 || r2 == 4 || r3 == 4 {
			lucky = append(lucky, 22)
		}
		if r1 == 5 || r2 == 5 || r3 == 5 {
			lucky = append(lucky, 23)
		}
	}

	log.Println("Lucky : ", lucky)

	//ตรวจคนถูกรางวัล และให้รางวัล
	for _, v := range lucky {
		//case ตัวเลข 3 - 8
		if v == 3 || v == 4 || v == 5 || v == 6 || v == 7 || v == 8 {
			//เปลี่ยนสถานะ รายการที่ชนะ
			vSum := 1
			if v == 3 {
				if r1 == 1 {
					vSum++
				}
				if r2 == 1 {
					vSum++
				}
				if r3 == 1 {
					vSum++
				}
			}
			if v == 4 {
				if r1 == 2 {
					vSum++
				}
				if r2 == 2 {
					vSum++
				}
				if r3 == 2 {
					vSum++
				}
			}
			if v == 5 {
				if r1 == 3 {
					vSum++
				}
				if r2 == 3 {
					vSum++
				}
				if r3 == 3 {
					vSum++
				}
			}
			if v == 6 {
				if r1 == 4 {
					vSum++
				}
				if r2 == 4 {
					vSum++
				}
				if r3 == 4 {
					vSum++
				}
			}
			if v == 7 {
				if r1 == 5 {
					vSum++
				}
				if r2 == 5 {
					vSum++
				}
				if r3 == 5 {
					vSum++
				}
			}
			if v == 8 {
				if r1 == 6 {
					vSum++
				}
				if r2 == 6 {
					vSum++
				}
				if r3 == 6 {
					vSum++
				}
			}
			//อัพเดทคนถูกรางวัล
			err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				err := repository.UpdateWinNumberHighlowBet(tx, highlowID, v, vSum)
				if err != nil {
					return err
				}
				return nil
			})
			log.Println(err)
			must(err)
		}

		//case สูงต่ำ 0, 1
		if v == 0 || v == 1 {
			//อัพเดทคนถูกรางวัล
			err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				err := repository.UpdateWinHLHighlowBet(tx, highlowID, v)
				if err != nil {
					return err
				}
				return nil
			})
			must(err)
			log.Println(err)
		}

		//case 11
		if v == 2 {
			//อัพเดทคนถูกรางวัล
			err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				err := repository.UpdateWin11HighlowBet(tx, highlowID, v)
				if err != nil {
					return err
				}
				return nil
			})
			must(err)
			log.Println(err)
		}

		//case โต๊ด
		if v == 9 || v == 10 || v == 11 || v == 12 || v == 13 || v == 14 || v == 15 || v == 16 || v == 17 || v == 18 || v == 19 || v == 20 || v == 21 || v == 22 || v == 23 {
			//อัพเดทคนถูกรางวัล
			err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				err := repository.UpdateDuoNumberHighlowBet(tx, highlowID, v)
				if err != nil {
					return err
				}
				return nil
			})
			must(err)
			log.Println(err)
		}

		//สูงต่ำเลข
		if v == 24 || v == 25 || v == 26 || v == 27 || v == 28 || v == 29 || v == 30 || v == 31 || v == 32 || v == 33 || v == 34 || v == 35 {
			//อัพเดทคนถูกรางวัล
			err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				err := repository.UpdateHLNumberHighlowBet(tx, highlowID, v)
				if err != nil {
					return err
				}
				return nil
			})
			must(err)
			log.Println(err)
		}
	}

	//เปิดกระดานต่อไป
	_, err = repository.CreateHighlow(db)
	must(err)
}

func ajaxHighlowBetWithdrawPostHandler(ctx *hime.Context) error {

	user := getUser(ctx)
	if user.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var res entity.ResponseFrontback
	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		//get total sum money user
		total, err := repository.GetTotalWinHighlowMyBet(tx, user.ID)
		if err != nil {
			return err
		}

		if total == 0 {
			return NewAppError("total0")
		}

		res.Wallet = total + user.Wallet

		//update is withdraw
		err = repository.UpdateHighlowIsWithdrawUser(tx, user.ID)
		if err != nil {
			return err
		}

		//update wallet user
		err = repository.UpdateWalletUserHighlow(tx, user.ID, total)
		if err != nil {
			return err
		}

		return nil
	})
	if IsAppError(err) {
		return ctx.Status(http.StatusOK).JSON(&res)
	}
	must(err)
	return ctx.Status(http.StatusOK).JSON(&res)
}
