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
	if err == sql.ErrNoRows {
		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {
			id, _ := repository.GetLiveHighlowID(tx)
			if id != "" {
				return nil
			}
			//เปิดกระดานต่อไป
			_, err = repository.CreateHighlow(db)
			if err != nil {
				return err
			}
			return nil
		})
		must(err)
		return ctx.RedirectToGet()
	}
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
	r1, r2, r3 := RandomRoll(highlowID, db)

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

//RandomRoll roll random
func RandomRoll(highlowID string, postgre *sql.DB) (int, int, int) {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(100)+1 >= config.HighlowWinrate {
		return rand.Intn(6) + 1, rand.Intn(6) + 1, rand.Intn(6) + 1
	}

	var r1, r2, r3 int
	highlow, err := repository.GetHighlowLink(postgre, highlowID)
	must(err)

	cost := 100000
	for i := 1; i <= 216; i++ {
		total, s1, s2, s3 := TotalRollCase(i, &highlow)
		log.Println("Cost: %s  Total: %s", cost, total)
		log.Println("Roll: %s %s %s", r1, r2, r3)
		if cost > total {
			cost = total
			r1 = s1
			r2 = s2
			r3 = s3
		}
		if cost == total {
			if rand.Intn(100)+1 >= config.HighlowRandomRoll {
				r1 = s1
				r2 = s2
				r3 = s3
			}
		}

	}
	return r1, r2, r3
}

//TotalRollCase showcase
func TotalRollCase(i int, h *entity.HighlowLinkBet) (int, int, int, int) {
	total := 0
	switch i {
	case 1:
		// 1 1 1
		total += h.N1 * 3
		total += h.Low
		total += h.Low1
		return total, 1, 1, 1
	case 2:
		// 1 1 2
		total += h.N1 * 2
		total += h.N2
		total += h.N12 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		return total, 1, 1, 2
	case 3:
		// 1 1 3
		total += h.N1 * 2
		total += h.N3
		total += h.N13 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		return total, 1, 1, 3
	case 4:
		// 1 1 4
		total += h.N1 * 2
		total += h.N4
		total += h.N14 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		return total, 1, 1, 4
	case 5:
		// 1 1 5
		total += h.N1 * 2
		total += h.N5
		total += h.N15 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low5 * 3
		return total, 1, 1, 5
	case 6:
		// 1 1 6
		total += h.N1 * 2
		total += h.N6
		total += h.N16 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low6 * 3
		return total, 1, 1, 6
	case 7:
		// 1 2 1
		total += h.N1 * 2
		total += h.N2
		total += h.N12 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		return total, 1, 2, 1
	case 8:
		// 1 2 2
		total += h.N1
		total += h.N2 * 2
		total += h.N12 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		return total, 1, 2, 2
	case 9:
		// 1 2 3
		total += h.N1
		total += h.N2
		total += h.N3
		total += h.N12 * 4
		total += h.N13 * 4
		total += h.N23 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 1, 2, 3
	case 10:
		// 1 2 4
		total += h.N1
		total += h.N2
		total += h.N4
		total += h.N12 * 4
		total += h.N14 * 4
		total += h.N24 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 1, 2, 4
	case 11:
		// 1 2 5
		total += h.N1
		total += h.N2
		total += h.N5
		total += h.N12 * 4
		total += h.N15 * 4
		total += h.N25 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 1, 2, 5
	case 12:
		// 1 2 6
		total += h.N1
		total += h.N2
		total += h.N6
		total += h.N12 * 4
		total += h.N16 * 4
		total += h.N26 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 1, 2, 6
	case 13:
		// 1 3 1
		total += h.N1 * 2
		total += h.N3
		total += h.N13 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		return total, 1, 3, 1
	case 14:
		// 1 3 2
		total += h.N1
		total += h.N2
		total += h.N3
		total += h.N12 * 4
		total += h.N13 * 4
		total += h.N23 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 1, 3, 2
	case 15:
		// 1 3 3
		total += h.N3 * 2
		total += h.N1
		total += h.N13 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		return total, 1, 3, 3
	case 16:
		// 1 3 4
		total += h.N1
		total += h.N3
		total += h.N4
		total += h.N13 * 4
		total += h.N14 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 1, 3, 4
	case 17:
		// 1 3 5
		total += h.N1
		total += h.N3
		total += h.N5
		total += h.N13 * 4
		total += h.N15 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 1, 3, 5
	case 18:
		// 1 3 6
		total += h.N1
		total += h.N3
		total += h.N6
		total += h.N13 * 4
		total += h.N16 * 4
		total += h.N36 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low6 * 3
		return total, 1, 3, 6
	case 19:
		// 1 4 1
		total += h.N1 * 2
		total += h.N4
		total += h.N14 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		return total, 1, 4, 1
	case 20:
		// 1 4 2
		total += h.N1
		total += h.N2
		total += h.N4
		total += h.N12 * 4
		total += h.N14 * 4
		total += h.N24 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 1, 4, 2
	case 21:
		// 1 4 3
		total += h.N1
		total += h.N3
		total += h.N4
		total += h.N13 * 4
		total += h.N14 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 1, 4, 3
	case 22:
		// 1 4 4
		total += h.N1
		total += h.N4 * 2
		total += h.N14 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		return total, 1, 4, 4
	case 23:
		// 1 4 5
		total += h.N1
		total += h.N4
		total += h.N5
		total += h.N14 * 4
		total += h.N15 * 4
		total += h.N45 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		total += h.Low5 * 3
		return total, 1, 4, 5
	case 24:
		// 1 4 6
		total += h.N1
		total += h.N4
		total += h.N6
		total += h.N14 * 4
		total += h.N16 * 4
		total += h.N46 * 4
		total += h.N11 * 5
		return total, 1, 4, 6
	case 25:
		// 1 5 1
		total += h.N1 * 2
		total += h.N5
		total += h.N15 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low5 * 3
		return total, 1, 5, 1
	case 26:
		//1 5 2
		total += h.N1
		total += h.N2
		total += h.N5
		total += h.N12 * 4
		total += h.N15 * 4
		total += h.N25 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 1, 5, 2
	case 27:
		// 1 5 3
		total += h.N1
		total += h.N3
		total += h.N5
		total += h.N13 * 4
		total += h.N15 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 1, 5, 3
	case 28:
		// 1 5 4
		total += h.N1
		total += h.N4
		total += h.N5
		total += h.N14 * 4
		total += h.N15 * 4
		total += h.N45 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		total += h.Low5 * 3
		return total, 1, 5, 4
	case 29:
		// 1 5 5
		total += h.N1
		total += h.N5 * 2
		total += h.N15 * 4
		total += h.N11 * 5
		return total, 1, 5, 5
	case 30:
		// 1 5 6
		total += h.N1
		total += h.N5
		total += h.N6
		total += h.N15 * 4
		total += h.N16 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 1, 5, 6
	case 31:
		// 1 6 1
		total += h.N1 * 2
		total += h.N6
		total += h.N16 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low6 * 3
		return total, 1, 6, 1
	case 32:
		// 1 6 2
		total += h.N1
		total += h.N2
		total += h.N6
		total += h.N12 * 4
		total += h.N16 * 4
		total += h.N26 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 1, 6, 2
	case 33:
		// 1 6 3
		total += h.N1
		total += h.N3
		total += h.N6
		total += h.N13 * 4
		total += h.N16 * 4
		total += h.N36 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low6 * 3
		return total, 1, 6, 3
	case 34:
		// 1 6 4
		total += h.N1
		total += h.N4
		total += h.N6
		total += h.N14 * 4
		total += h.N16 * 4
		total += h.N46 * 4
		total += h.N11 * 5
		return total, 1, 6, 4
	case 35:
		// 1 6 5
		total += h.N1
		total += h.N5
		total += h.N6
		total += h.N15 * 4
		total += h.N16 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 1, 6, 5
	case 36:
		// 1 6 6
		total += h.N1
		total += h.N6 * 2
		total += h.N16 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High6 * 3
		return total, 1, 6, 6
	case 37:
		// 2 1 1
		total += h.N1 * 2
		total += h.N2
		total += h.N12 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		return total, 2, 1, 1
	case 38:
		// 2 1 2
		total += h.N1
		total += h.N2 * 2
		total += h.N12 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		return total, 2, 1, 2
	case 39:
		// 2 1 3
		total += h.N1
		total += h.N2
		total += h.N3
		total += h.N12 * 4
		total += h.N13 * 4
		total += h.N23 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 2, 1, 3
	case 40:
		// 2 1 4
		total += h.N1
		total += h.N2
		total += h.N4
		total += h.N12 * 4
		total += h.N14 * 4
		total += h.N24 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 2, 1, 4
	case 41:
		// 2 1 5
		total += h.N1
		total += h.N2
		total += h.N5
		total += h.N12 * 4
		total += h.N15 * 4
		total += h.N25 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 2, 1, 5
	case 42:
		// 2 1 6
		total += h.N1
		total += h.N2
		total += h.N6
		total += h.N12 * 4
		total += h.N16 * 4
		total += h.N26 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 2, 1, 6
	case 43:
		// 2 2 1
		total += h.N1
		total += h.N2 * 2
		total += h.N12 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		return total, 2, 2, 1
	case 44:
		// 2 2 2
		total += h.N2 * 3
		total += h.Low
		total += h.Low2 * 3
		return total, 2, 2, 2
	case 45:
		// 2 2 3
		total += h.N2 * 2
		total += h.N3
		total += h.N23 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 2, 2, 3
	case 46:
		// 2 2 4
		total += h.N2 * 2
		total += h.N4
		total += h.N24 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 2, 2, 4
	case 47:
		// 2 2 5
		total += h.N2 * 2
		total += h.N5
		total += h.N25 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 2, 2, 5
	case 48:
		// 2 2 6
		total += h.N2 * 2
		total += h.N6
		total += h.N26 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 2, 2, 6
	case 49:
		// 2 3 1
		total += h.N1
		total += h.N2
		total += h.N3
		total += h.N12 * 4
		total += h.N13 * 4
		total += h.N23 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 2, 3, 1
	case 50:
		//  2 3 2
		total += h.N2 * 2
		total += h.N3
		total += h.N23 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 2, 3, 2
	case 51:
		// 2 3 3
		total += h.N2
		total += h.N3 * 2
		total += h.N23 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 2, 3, 3
	case 52:
		// 2 3 4
		total += h.N2
		total += h.N3
		total += h.N4
		total += h.N23 * 4
		total += h.N24 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 2, 3, 4
	case 53:
		// 2 3 5
		total += h.N2
		total += h.N3
		total += h.N5
		total += h.N23 * 4
		total += h.N25 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 2, 3, 5
	case 54:
		// 2 3 6
		total += h.N2
		total += h.N3
		total += h.N6
		total += h.N23 * 4
		total += h.N26 * 4
		total += h.N36 * 4
		total += h.N11 * 5
		return total, 2, 3, 6
	case 55:
		// 2 4 1
		total += h.N1
		total += h.N2
		total += h.N4
		total += h.N12 * 4
		total += h.N14 * 4
		total += h.N24 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 2, 4, 1
	case 56:
		// 2 4 2
		total += h.N2 * 2
		total += h.N4
		total += h.N24 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 2, 4, 2
	case 57:
		// 2 4 3
		total += h.N2
		total += h.N3
		total += h.N4
		total += h.N23 * 4
		total += h.N24 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 2, 4, 3
	case 58:
		// 2 4 4
		total += h.N2
		total += h.N4 * 2
		total += h.N24 * 4
		total += h.Low
		total += h.Low4 * 3
		return total, 2, 4, 4
	case 59:
		// 2 4 5
		total += h.N2
		total += h.N4
		total += h.N5
		total += h.N24 * 4
		total += h.N25 * 4
		total += h.N45 * 4
		total += h.N11 * 5
		return total, 2, 4, 5
	case 60:
		// 2 4 6
		total += h.N2
		total += h.N4
		total += h.N6
		total += h.N24 * 4
		total += h.N26 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 2, 4, 6
	case 61:
		// 2 5 1
		total += h.N1
		total += h.N2
		total += h.N5
		total += h.N12 * 4
		total += h.N15 * 4
		total += h.N25 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 2, 5, 1
	case 62:
		// 2 5 2
		total += h.N2 * 2
		total += h.N5
		total += h.N25 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 2, 5, 2
	case 63:
		// 2 5 3
		total += h.N2
		total += h.N3
		total += h.N5
		total += h.N23 * 4
		total += h.N25 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 2, 5, 3
	case 64:
		// 2 5 4
		total += h.N2
		total += h.N4
		total += h.N5
		total += h.N24 * 4
		total += h.N25 * 4
		total += h.N45 * 4
		total += h.N11 * 5
		return total, 2, 5, 4
	case 65:
		// 2 5 5
		total += h.N2
		total += h.N5 * 2
		total += h.N25 * 4
		total += h.High
		total += h.High2 * 3
		return total, 2, 5, 5
	case 66:
		// 2 5 6
		total += h.N2
		total += h.N5
		total += h.N6
		total += h.N25 * 4
		total += h.N26 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 2, 5, 6
	case 67:
		// 2 6 1
		total += h.N1
		total += h.N2
		total += h.N6
		total += h.N12 * 4
		total += h.N16 * 4
		total += h.N26 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 2, 6, 1
	case 68:
		// 2 6 2
		total += h.N2 * 2
		total += h.N6
		total += h.N26 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 2, 6, 2
	case 69:
		// 2 6 3
		total += h.N2
		total += h.N3
		total += h.N6
		total += h.N23 * 4
		total += h.N26 * 4
		total += h.N36 * 4
		total += h.N11 * 5
		return total, 2, 6, 3
	case 70:
		// 2 6 4
		total += h.N2
		total += h.N4
		total += h.N6
		total += h.N24 * 4
		total += h.N26 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 2, 6, 4
	case 71:
		// 2 6 5
		total += h.N2
		total += h.N5
		total += h.N6
		total += h.N25 * 4
		total += h.N26 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 2, 6, 5
	case 72:
		// 2 6 6
		total += h.N2
		total += h.N6 * 2
		total += h.N26 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High6 * 3
		return total, 2, 6, 6
	case 73:
		// 3 1 1
		total += h.N1 * 2
		total += h.N3
		total += h.N13 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		return total, 3, 1, 1
	case 74:
		// 3 1 2
		total += h.N1
		total += h.N2
		total += h.N3
		total += h.N12 * 4
		total += h.N13 * 4
		total += h.N23 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 3, 1, 2
	case 75:
		// 3 1 3
		total += h.N3 * 2
		total += h.N1
		total += h.N13 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		return total, 3, 1, 3
	case 76:
		// 3 1 4
		total += h.N1
		total += h.N3
		total += h.N4
		total += h.N13 * 4
		total += h.N14 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 3, 1, 4
	case 77:
		// 3 1 5
		total += h.N1
		total += h.N3
		total += h.N5
		total += h.N13 * 4
		total += h.N15 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 3, 1, 5
	case 78:
		// 3 1 6
		total += h.N1
		total += h.N3
		total += h.N6
		total += h.N13 * 4
		total += h.N16 * 4
		total += h.N36 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low6 * 3
		return total, 3, 1, 6
	case 79:
		// 3 2 1
		total += h.N1
		total += h.N2
		total += h.N3
		total += h.N12 * 4
		total += h.N13 * 4
		total += h.N23 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 3, 2, 1
	case 80:
		// 3 2 2
		total += h.N2 * 2
		total += h.N3
		total += h.N23 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 3, 2, 2
	case 81:
		// 3 2 3
		total += h.N2
		total += h.N3 * 2
		total += h.N23 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 3, 2, 3
	case 82:
		// 3 2 4
		total += h.N2
		total += h.N3
		total += h.N4
		total += h.N23 * 4
		total += h.N24 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 3, 2, 4
	case 83:
		// 3 2 5
		total += h.N2
		total += h.N3
		total += h.N5
		total += h.N23 * 4
		total += h.N25 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 3, 2, 5
	case 84:
		// 3 2 6
		total += h.N2
		total += h.N3
		total += h.N6
		total += h.N23 * 4
		total += h.N26 * 4
		total += h.N36 * 4
		total += h.N11 * 5
		return total, 3, 2, 6
	case 85:
		// 3 3 2
		total += h.N2
		total += h.N3 * 2
		total += h.N23 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		return total, 3, 3, 2
	case 86:
		// 3 3 3
		total += h.N3 * 3
		total += h.Low
		total += h.Low3 * 3
		return total, 3, 3, 3
	case 87:
		// 3 3 4
		total += h.N3 * 2
		total += h.N4
		total += h.N34 * 4
		total += h.Low
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 3, 3, 4
	case 88:
		// 3 3 5
		total += h.N3 * 2
		total += h.N5
		total += h.N35 * 4
		total += h.N11 * 5
		return total, 3, 3, 5
	case 89:
		// 3 3 6
		total += h.N3 * 2
		total += h.N6
		total += h.N36 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High6 * 3
		return total, 3, 3, 6
	case 90:
		// 3 4 1
		total += h.N1
		total += h.N3
		total += h.N4
		total += h.N13 * 4
		total += h.N14 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 3, 4, 1
	case 91:
		// 3 4 2
		total += h.N2
		total += h.N3
		total += h.N4
		total += h.N23 * 4
		total += h.N24 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 3, 4, 2
	case 92:
		// 3 4 3
		total += h.N3 * 2
		total += h.N4
		total += h.N34 * 4
		total += h.Low
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 3, 4, 3
	case 93:
		// 3 4 4
		total += h.N3
		total += h.N4 * 2
		total += h.N34 * 4
		total += h.N11 * 5
		return total, 3, 4, 4
	case 94:
		// 3 4 5
		total += h.N3
		total += h.N4
		total += h.N5
		total += h.N34 * 4
		total += h.N35 * 4
		total += h.N45 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 3, 4, 5
	case 95:
		// 3 4 6
		total += h.N3
		total += h.N4
		total += h.N6
		total += h.N34 * 4
		total += h.N36 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 3, 4, 6
	case 96:
		// 3 5 1
		total += h.N1
		total += h.N3
		total += h.N5
		total += h.N13 * 4
		total += h.N15 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 3, 5, 1
	case 97:
		// 3 5 2
		total += h.N2
		total += h.N3
		total += h.N5
		total += h.N23 * 4
		total += h.N25 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 3, 5, 2
	case 98:
		// 3 5 4
		total += h.N3
		total += h.N4
		total += h.N5
		total += h.N34 * 4
		total += h.N35 * 4
		total += h.N45 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 3, 5, 4
	case 99:
		// 3 5 5
		total += h.N3
		total += h.N5 * 2
		total += h.N35 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		return total, 3, 5, 5
	case 100:
		// 3 5 6
		total += h.N3
		total += h.N5
		total += h.N6
		total += h.N35 * 4
		total += h.N36 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 3, 5, 6
	case 101:
		// 3 6 1
		total += h.N1
		total += h.N3
		total += h.N6
		total += h.N13 * 4
		total += h.N16 * 4
		total += h.N36 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low6 * 3
		return total, 3, 6, 1
	case 102:
		// 3 6 2
		total += h.N2
		total += h.N3
		total += h.N6
		total += h.N23 * 4
		total += h.N26 * 4
		total += h.N36 * 4
		total += h.N11 * 5
		return total, 3, 6, 2
	case 103:
		// 3 6 4
		total += h.N3
		total += h.N4
		total += h.N6
		total += h.N34 * 4
		total += h.N36 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 3, 6, 4
	case 104:
		// 3 6 5
		total += h.N3
		total += h.N5
		total += h.N6
		total += h.N35 * 4
		total += h.N36 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 3, 6, 5
	case 105:
		// 3 6 6
		total += h.N3
		total += h.N6 * 2
		total += h.N36 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High6 * 3
		return total, 3, 6, 6
	case 106:
		// 4 1 1
		total += h.N1 * 2
		total += h.N4
		total += h.N14 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		return total, 4, 1, 1
	case 107:
		// 4 1 2
		total += h.N1
		total += h.N2
		total += h.N4
		total += h.N12 * 4
		total += h.N14 * 4
		total += h.N24 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 4, 1, 2
	case 108:
		// 4 1 3
		total += h.N1
		total += h.N3
		total += h.N4
		total += h.N13 * 4
		total += h.N14 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 4, 1, 3
	case 109:
		// 4 1 4
		total += h.N1
		total += h.N4 * 2
		total += h.N14 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		return total, 4, 1, 4
	case 110:
		// 4 1 5
		total += h.N1
		total += h.N4
		total += h.N5
		total += h.N14 * 4
		total += h.N15 * 4
		total += h.N45 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		total += h.Low5 * 3
		return total, 4, 1, 5
	case 111:
		// 4 1 6
		total += h.N1
		total += h.N4
		total += h.N6
		total += h.N14 * 4
		total += h.N16 * 4
		total += h.N46 * 4
		total += h.N11 * 5
		return total, 4, 1, 6
	case 112:
		// 4 2 1
		total += h.N1
		total += h.N2
		total += h.N4
		total += h.N12 * 4
		total += h.N14 * 4
		total += h.N24 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 4, 2, 1
	case 113:
		// 4 2 2
		total += h.N2 * 2
		total += h.N4
		total += h.N24 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low4 * 3
		return total, 4, 2, 2
	case 114:
		// 4 2 3
		total += h.N2
		total += h.N3
		total += h.N4
		total += h.N23 * 4
		total += h.N24 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 4, 2, 3
	case 115:
		// 4 2 4
		total += h.N2
		total += h.N4 * 2
		total += h.N24 * 4
		total += h.Low
		total += h.Low4 * 3
		return total, 2, 4, 4
	case 116:
		// 4 2 5
		total += h.N2
		total += h.N4
		total += h.N5
		total += h.N24 * 4
		total += h.N25 * 4
		total += h.N45 * 4
		total += h.N11 * 5
		return total, 2, 4, 5
	case 117:
		// 4 2 6
		total += h.N2
		total += h.N4
		total += h.N6
		total += h.N24 * 4
		total += h.N26 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 4, 2, 6
	case 118:
		// 4 3 1
		total += h.N1
		total += h.N3
		total += h.N4
		total += h.N13 * 4
		total += h.N14 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 4, 3, 1
	case 119:
		// 4 3 2
		total += h.N2
		total += h.N3
		total += h.N4
		total += h.N23 * 4
		total += h.N24 * 4
		total += h.N34 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 4, 3, 2
	case 120:
		// 4 3 3
		total += h.N3 * 2
		total += h.N4
		total += h.N34 * 4
		total += h.Low
		total += h.Low3 * 3
		total += h.Low4 * 3
		return total, 4, 3, 3
	case 121:
		// 4 3 4
		total += h.N3
		total += h.N4 * 2
		total += h.N34 * 4
		total += h.N11 * 5
		return total, 4, 3, 4
	case 122:
		// 4 3 5
		total += h.N3
		total += h.N4
		total += h.N5
		total += h.N34 * 4
		total += h.N35 * 4
		total += h.N45 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 4, 3, 5
	case 123:
		// 4 3 6
		total += h.N3
		total += h.N4
		total += h.N6
		total += h.N34 * 4
		total += h.N36 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 4, 3, 6
	case 124:
		// 4 4 1
		total += h.N1
		total += h.N4 * 2
		total += h.N14 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		return total, 4, 4, 1
	case 125:
		// 4 4 2
		total += h.N2
		total += h.N4 * 2
		total += h.N24 * 4
		total += h.Low
		total += h.Low4 * 3
		return total, 4, 4, 2
	case 126:
		// 4 4 3
		total += h.N3
		total += h.N4 * 2
		total += h.N34 * 4
		total += h.N11 * 5
		return total, 4, 4, 3
	case 127:
		// 4 4 4
		total += h.N4 * 3
		total += h.High
		total += h.High4 * 3
		return total, 4, 4, 4
	case 128:
		// 4 4 5
		total += h.N4 * 2
		total += h.N5
		total += h.N45 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 4, 4, 5
	case 129:
		// 4 4 6
		total += h.N4 * 2
		total += h.N6
		total += h.N46 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 4, 4, 6
	case 130:
		// 4 5 1
		total += h.N1
		total += h.N4
		total += h.N5
		total += h.N14 * 4
		total += h.N15 * 4
		total += h.N45 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		total += h.Low5 * 3
		return total, 4, 5, 1
	case 131:
		// 4 5 2
		total += h.N2
		total += h.N4
		total += h.N5
		total += h.N24 * 4
		total += h.N25 * 4
		total += h.N45 * 4
		total += h.N11 * 5
		return total, 4, 5, 2
	case 132:
		// 4 5 3
		total += h.N3
		total += h.N4
		total += h.N5
		total += h.N34 * 4
		total += h.N35 * 4
		total += h.N45 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 4, 5, 3
	case 133:
		// 4 5 4
		total += h.N4 * 2
		total += h.N5
		total += h.N45 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 4, 5, 4
	case 134:
		// 4 5 5
		total += h.N4
		total += h.N5 * 2
		total += h.N45 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 4, 5, 5
	case 135:
		// 4 5 6
		total += h.N4
		total += h.N5
		total += h.N6
		total += h.N45 * 4
		total += h.N46 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 4, 5, 6
	case 136:
		// 4 6 1
		total += h.N1
		total += h.N4
		total += h.N6
		total += h.N14 * 4
		total += h.N16 * 4
		total += h.N46 * 4
		total += h.N11 * 5
		return total, 4, 6, 1
	case 137:
		// 4 6 2
		total += h.N2
		total += h.N4
		total += h.N6
		total += h.N24 * 4
		total += h.N26 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 4, 6, 2
	case 138:
		// 4 6 3
		total += h.N3
		total += h.N4
		total += h.N6
		total += h.N34 * 4
		total += h.N36 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 4, 6, 3
	case 139:
		// 4 6 4
		total += h.N4 * 2
		total += h.N6
		total += h.N46 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 4, 6, 4
	case 140:
		// 4 6 5
		total += h.N4
		total += h.N5
		total += h.N6
		total += h.N45 * 4
		total += h.N46 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 4, 6, 5
	case 141:
		// 4 6 6
		total += h.N4
		total += h.N6 * 2
		total += h.N46 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 4, 6, 6
	case 142:
		// 5 1 1
		total += h.N1 * 2
		total += h.N5
		total += h.N15 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low5 * 3
		return total, 5, 1, 1
	case 143:
		// 5 1 2
		total += h.N1
		total += h.N2
		total += h.N5
		total += h.N12 * 4
		total += h.N15 * 4
		total += h.N25 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 5, 1, 2
	case 144:
		// 5 1 3
		total += h.N1
		total += h.N3
		total += h.N5
		total += h.N13 * 4
		total += h.N15 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 5, 1, 3
	case 145:
		// 5 1 4
		total += h.N1
		total += h.N4
		total += h.N5
		total += h.N14 * 4
		total += h.N15 * 4
		total += h.N45 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		total += h.Low5 * 3
		return total, 5, 1, 4
	case 146:
		// 5 1 5
		total += h.N1
		total += h.N5 * 2
		total += h.N15 * 4
		total += h.N11 * 5
		return total, 5, 1, 5
	case 147:
		// 5 1 6
		total += h.N1
		total += h.N5
		total += h.N6
		total += h.N15 * 4
		total += h.N16 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 1, 6
	case 148:
		// 5 2 1
		total += h.N1
		total += h.N2
		total += h.N5
		total += h.N12 * 4
		total += h.N15 * 4
		total += h.N25 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 5, 2, 1
	case 149:
		// 5 2 2
		total += h.N2 * 2
		total += h.N5
		total += h.N25 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low5 * 3
		return total, 5, 2, 2
	case 150:
		// 5 2 3
		total += h.N2
		total += h.N3
		total += h.N5
		total += h.N23 * 4
		total += h.N25 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 5, 2, 3
	case 151:
		// 5 2 4
		total += h.N2
		total += h.N4
		total += h.N5
		total += h.N24 * 4
		total += h.N25 * 4
		total += h.N45 * 4
		total += h.N11 * 5
		return total, 5, 2, 4
	case 152:
		// 5 2 5
		total += h.N2
		total += h.N5 * 2
		total += h.N25 * 4
		total += h.High
		total += h.High2 * 3
		return total, 5, 2, 5
	case 153:
		// 5 2 6
		total += h.N2
		total += h.N5
		total += h.N6
		total += h.N25 * 4
		total += h.N26 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 2, 6
	case 154:
		// 5 3 1
		total += h.N1
		total += h.N3
		total += h.N5
		total += h.N13 * 4
		total += h.N15 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 5, 3, 1
	case 155:
		// 5 3 2
		total += h.N2
		total += h.N3
		total += h.N5
		total += h.N23 * 4
		total += h.N25 * 4
		total += h.N35 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low3 * 3
		total += h.Low5 * 3
		return total, 5, 3, 2
	case 156:
		// 5 3 3
		total += h.N3 * 2
		total += h.N5
		total += h.N35 * 4
		total += h.N11 * 5
		return total, 5, 3, 3
	case 157:
		// 5 3 4
		total += h.N3
		total += h.N4
		total += h.N5
		total += h.N34 * 4
		total += h.N35 * 4
		total += h.N45 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 5, 3, 4
	case 158:
		// 5 3 5
		total += h.N3
		total += h.N5 * 2
		total += h.N35 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		return total, 5, 3, 5
	case 159:
		// 5 3 6
		total += h.N3
		total += h.N5
		total += h.N6
		total += h.N35 * 4
		total += h.N36 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 3, 6
	case 160:
		// 5 4 1
		total += h.N1
		total += h.N4
		total += h.N5
		total += h.N14 * 4
		total += h.N15 * 4
		total += h.N45 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low4 * 3
		total += h.Low5 * 3
		return total, 5, 4, 1
	case 161:
		// 5 4 2
		total += h.N2
		total += h.N4
		total += h.N5
		total += h.N24 * 4
		total += h.N25 * 4
		total += h.N45 * 4
		total += h.N11 * 5
		return total, 5, 4, 2
	case 162:
		// 5 4 3
		total += h.N3
		total += h.N4
		total += h.N5
		total += h.N34 * 4
		total += h.N35 * 4
		total += h.N45 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 5, 4, 3
	case 163:
		// 5 4 4
		total += h.N4 * 2
		total += h.N5
		total += h.N45 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 5, 4, 4
	case 164:
		// 5 4 5
		total += h.N4
		total += h.N5 * 2
		total += h.N45 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 5, 4, 5
	case 165:
		// 5 4 6
		total += h.N4
		total += h.N5
		total += h.N6
		total += h.N45 * 4
		total += h.N46 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 4, 6
	case 166:
		// 5 5 1
		total += h.N1
		total += h.N5 * 2
		total += h.N15 * 4
		total += h.N11 * 5
		return total, 5, 5, 1
	case 167:
		// 5 5 2
		total += h.N2
		total += h.N5 * 2
		total += h.N25 * 4
		total += h.High
		total += h.High2 * 3
		return total, 5, 5, 2
	case 168:
		// 5 5 3
		total += h.N3
		total += h.N5 * 2
		total += h.N35 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		return total, 5, 5, 3
	case 169:
		// 5 5 4
		total += h.N4
		total += h.N5 * 2
		total += h.N45 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		return total, 5, 5, 4
	case 170:
		// 5 5 5
		total += h.N5 * 3
		total += h.High
		total += h.High5 * 3
		return total, 5, 5, 5
	case 171:
		// 5 5 6
		total += h.N5 * 2
		total += h.N6
		total += h.N56 * 4
		total += h.High
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 5, 6
	case 172:
		// 5 6 1
		total += h.N1
		total += h.N5
		total += h.N6
		total += h.N15 * 4
		total += h.N16 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 6, 1
	case 173:
		// 5 6 2
		total += h.N2
		total += h.N5
		total += h.N6
		total += h.N25 * 4
		total += h.N26 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 6, 2
	case 174:
		// 5 6 3
		total += h.N3
		total += h.N5
		total += h.N6
		total += h.N35 * 4
		total += h.N36 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 6, 3
	case 175:
		// 5 6 4
		total += h.N4
		total += h.N5
		total += h.N6
		total += h.N45 * 4
		total += h.N46 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 6, 4
	case 176:
		// 5 6 5
		total += h.N5 * 2
		total += h.N6
		total += h.N56 * 4
		total += h.High
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 6, 5
	case 177:
		// 5 6 6
		total += h.N5
		total += h.N6 * 2
		total += h.N56 * 4
		total += h.High
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 5, 6, 6
	case 178:
		// 6 1 1
		total += h.N1 * 2
		total += h.N6
		total += h.N16 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low6 * 3
		return total, 6, 1, 1
	case 179:
		// 6 1 2
		total += h.N1
		total += h.N2
		total += h.N6
		total += h.N12 * 4
		total += h.N16 * 4
		total += h.N26 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 6, 1, 2
	case 180:
		// 6 1 3
		total += h.N1
		total += h.N3
		total += h.N6
		total += h.N13 * 4
		total += h.N16 * 4
		total += h.N36 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low6 * 3
		return total, 6, 1, 3
	case 181:
		// 6 1 4
		total += h.N1
		total += h.N4
		total += h.N6
		total += h.N14 * 4
		total += h.N16 * 4
		total += h.N46 * 4
		total += h.N11 * 5
		return total, 6, 1, 4
	case 182:
		// 6 1 5
		total += h.N1
		total += h.N5
		total += h.N6
		total += h.N15 * 4
		total += h.N16 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 1, 5
	case 183:
		// 6 1 6
		total += h.N1
		total += h.N6 * 2
		total += h.N16 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High6 * 3
		return total, 6, 1, 6
	case 184:
		// 6 2 1
		total += h.N1
		total += h.N2
		total += h.N6
		total += h.N12 * 4
		total += h.N16 * 4
		total += h.N26 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 6, 2, 1
	case 185:
		// 6 2 2
		total += h.N2 * 2
		total += h.N6
		total += h.N26 * 4
		total += h.Low
		total += h.Low2 * 3
		total += h.Low6 * 3
		return total, 6, 2, 2
	case 186:
		// 6 2 3
		total += h.N2
		total += h.N3
		total += h.N6
		total += h.N23 * 4
		total += h.N26 * 4
		total += h.N36 * 4
		total += h.N11 * 5
		return total, 6, 2, 3
	case 187:
		// 6 2 4
		total += h.N2
		total += h.N4
		total += h.N6
		total += h.N24 * 4
		total += h.N26 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 6, 2, 4
	case 188:
		// 6 2 5
		total += h.N2
		total += h.N5
		total += h.N6
		total += h.N25 * 4
		total += h.N26 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 2, 5
	case 189:
		// 6 2 6
		total += h.N2
		total += h.N6 * 2
		total += h.N26 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High6 * 3
		return total, 6, 2, 6
	case 190:
		// 6 3 1
		total += h.N1
		total += h.N3
		total += h.N6
		total += h.N13 * 4
		total += h.N16 * 4
		total += h.N36 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		total += h.Low6 * 3
		return total, 6, 3, 1
	case 191:
		// 6 3 2
		total += h.N2
		total += h.N3
		total += h.N6
		total += h.N23 * 4
		total += h.N26 * 4
		total += h.N36 * 4
		total += h.N11 * 5
		return total, 6, 3, 2
	case 192:
		// 6 3 3
		total += h.N3 * 2
		total += h.N6
		total += h.N36 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High6 * 3
		return total, 6, 3, 3
	case 193:
		// 6 3 4
		total += h.N3
		total += h.N4
		total += h.N6
		total += h.N34 * 4
		total += h.N36 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 6, 3, 4
	case 194:
		// 6 3 5
		total += h.N3
		total += h.N5
		total += h.N6
		total += h.N35 * 4
		total += h.N36 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 3, 5
	case 195:
		// 6 3 6
		total += h.N3
		total += h.N6 * 2
		total += h.N36 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High6 * 3
		return total, 6, 3, 6
	case 196:
		// 6 4 1
		total += h.N1
		total += h.N4
		total += h.N6
		total += h.N14 * 4
		total += h.N16 * 4
		total += h.N46 * 4
		total += h.N11 * 5
		return total, 6, 4, 1
	case 197:
		// 6 4 2
		total += h.N2
		total += h.N4
		total += h.N6
		total += h.N24 * 4
		total += h.N26 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 6, 4, 2
	case 198:
		// 6 4 3
		total += h.N3
		total += h.N4
		total += h.N6
		total += h.N34 * 4
		total += h.N36 * 4
		total += h.N46 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 6, 4, 3
	case 199:
		// 6 4 4
		total += h.N4 * 2
		total += h.N6
		total += h.N46 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 6, 4, 4
	case 200:
		// 6 4 5
		total += h.N4
		total += h.N5
		total += h.N6
		total += h.N45 * 4
		total += h.N46 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 4, 5
	case 201:
		// 6 4 6
		total += h.N4
		total += h.N6 * 2
		total += h.N46 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 6, 4, 6
	case 202:
		// 6 5 1
		total += h.N1
		total += h.N5
		total += h.N6
		total += h.N15 * 4
		total += h.N16 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 5, 1
	case 203:
		// 6 5 2
		total += h.N2
		total += h.N5
		total += h.N6
		total += h.N25 * 4
		total += h.N26 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 5, 2
	case 204:
		// 6 5 3
		total += h.N3
		total += h.N5
		total += h.N6
		total += h.N35 * 4
		total += h.N36 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 5, 3
	case 205:
		// 6 5 4
		total += h.N4
		total += h.N5
		total += h.N6
		total += h.N45 * 4
		total += h.N46 * 4
		total += h.N56 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 5, 4
	case 206:
		// 6 5 5
		total += h.N5 * 2
		total += h.N6
		total += h.N56 * 4
		total += h.High
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 5, 5
	case 207:
		// 6 5 6
		total += h.N5
		total += h.N6 * 2
		total += h.N56 * 4
		total += h.High
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 5, 6
	case 208:
		// 6 6 1
		total += h.N1
		total += h.N6 * 2
		total += h.N16 * 4
		total += h.High
		total += h.High1 * 3
		total += h.High6 * 3
		return total, 6, 6, 1
	case 209:
		// 6 6 2
		total += h.N2
		total += h.N6 * 2
		total += h.N26 * 4
		total += h.High
		total += h.High2 * 3
		total += h.High6 * 3
		return total, 6, 6, 2
	case 210:
		// 6 6 3
		total += h.N3
		total += h.N6 * 2
		total += h.N36 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High6 * 3
		return total, 6, 6, 3
	case 211:
		// 6 6 4
		total += h.N4
		total += h.N6 * 2
		total += h.N46 * 4
		total += h.High
		total += h.High4 * 3
		total += h.High6 * 3
		return total, 6, 6, 4
	case 212:
		// 6 6 5
		total += h.N5
		total += h.N6 * 2
		total += h.N56 * 4
		total += h.High
		total += h.High5 * 3
		total += h.High6 * 3
		return total, 6, 6, 5
	case 213:
		// 6 6 6
		total += h.N6 * 3
		total += h.High
		total += h.High6 * 3
		return total, 6, 6, 6
	case 214:
		// 3 3 1
		total += h.N1
		total += h.N3 * 2
		total += h.N13 * 4
		total += h.Low
		total += h.Low1 * 3
		total += h.Low3 * 3
		return total, 3, 3, 1
	case 215:
		// 3 5 3
		total += h.N3 * 2
		total += h.N5
		total += h.N35 * 4
		total += h.N11 * 5
		return total, 3, 5, 3
	case 216:
		// 3 6 3
		total += h.N3 * 2
		total += h.N6
		total += h.N36 * 4
		total += h.High
		total += h.High3 * 3
		total += h.High6 * 3
		return total, 3, 6, 3
	}

	log.Println("ERROR")
	return 0, 0, 0, 0
}
