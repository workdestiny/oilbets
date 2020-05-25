package app

import (
	"database/sql"
	"image"
	"image/color"
	"image/draw"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/acoshift/pgsql"

	"github.com/moonrhythm/hime"
	"github.com/moonrhythm/session"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/repository"
	"github.com/workdestiny/oilbets/service"
)

func userGetHandler(ctx *hime.Context) error {
	user := getUser(ctx)

	hasBookbank := true
	bookbank, _ := repository.GetUserBookbank(db, user.ID)
	if bookbank.Number == "" {
		hasBookbank = false
	}

	p := page(ctx)
	p["User"] = user
	p["HasBookbank"] = hasBookbank
	p["Bookbank"] = bookbank
	return ctx.View("app/account", p)
}

//UserWithdrawMoneyGetHandler is money
func UserWithdrawMoneyGetHandler(ctx *hime.Context) error {
	user := getUser(ctx)
	hasBookbank := true
	bookbank, _ := repository.GetUserBookbank(db, user.ID)
	if bookbank.Number == "" {
		hasBookbank = false
	}

	p := page(ctx)
	p["HasBookbank"] = hasBookbank
	p["Bookbank"] = bookbank
	p["Wallet"] = user.Wallet

	return ctx.View("app/userbookbank", p)
}

//UserWithdrawMoneyPostHandler is money
func UserWithdrawMoneyPostHandler(ctx *hime.Context) error {
	amount := ctx.PostFormValueInt64("amount")
	user := getUser(ctx)

	f := getSession(ctx).Flash()
	f.Clear()

	if amount > user.Wallet {
		f.Add("Errors", "จำนวนเงินใน wallet ไม่พอ")
		return ctx.RedirectToGet()
	}

	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		_, err := repository.CreateWithdrawMoney(tx, &repository.CreateWithdrawMoneyModel{
			UserID: user.ID,
			Amount: amount,
		})
		if err != nil {
			return err
		}

		err = repository.UpdateWalletUser(tx, user.ID, amount)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	f.Add("Success", "ดำเนินการเรียบร้อยแล้ว")
	return ctx.RedirectToGet()
}

func ajaxUserFollowGapPostHandler(ctx *hime.Context) error {

	userID := GetMyID(ctx)
	if userID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	list, err := repository.ListFollowUserNextLoad(db, userID, req.Next, config.LimitListFollowUser+1)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	if len(list) == 0 {
		return ctx.NoContent()
	}

	var next time.Time
	isNext := false

	if len(list) > config.LimitListFollowUser {
		isNext = true
		next = list[config.LimitListFollowUser-1].CreatedAt
		list = list[:config.LimitListFollowUser]
	}

	res := entity.ResponseUserFollowGapList{
		Gap:    list,
		Next:   next,
		IsNext: isNext,
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func userPostHandler(ctx *hime.Context) error {

	f := getSession(ctx).Flash()
	f.Clear()

	id := GetMyID(ctx)
	owner := ctx.PostFormValueTrimSpace("owner")
	number := ctx.PostFormValueTrimSpace("number")
	bank := ctx.PostFormValueTrimSpace("bank")

	if utf8.RuneCountInString(owner) > 50 {
		f.Add("ErrName", "ชื่อต้องไม่เกิน 50 ตัวอักษร")
		return ctx.RedirectToGet()
	}
	if utf8.RuneCountInString(number) > 20 {
		f.Add("ErrName", "เลขบัญชี 20 ตัวอักษร")
		return ctx.RedirectToGet()
	}
	if utf8.RuneCountInString(bank) > 20 {
		f.Add("ErrName", "ชื่อต้องไม่เกิน 20 ตัวอักษร")
		return ctx.RedirectToGet()
	}

	err := repository.UpdateBookbankUser(db, id, owner, number, bank)
	must(err)

	addSuccess(f, "แก้ไขข้อมูลเรียบร้อยแล้ว!")
	return ctx.RedirectToGet()
}

func setFalsh(f *session.Flash, email, firstName, lastName, changePassword string) {

	f.Set("Email", email)
	f.Set("FirstName", firstName)
	f.Set("LastName", lastName)
	f.Set("ChangePass", changePassword)
}

func ajaxUploadProfileDisplayPostHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	if me.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	fileHeader, err := ctx.FormFileHeaderNotEmpty("image")
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	m, err := service.GetImageFromFile(fileHeader)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	dst := image.NewRGBA(m.Bounds())
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White),
		image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), m, m.Bounds().Min, draw.Over)

	mini, middle, normal := resizeUploadDisplayImage(dst)

	nMini, nMiddle, nNormal := generatProfileDisplayName(me.ID)

	err = upload(ctx, mini, nMini)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	err = upload(ctx, middle, nMiddle)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	err = upload(ctx, normal, nNormal)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.EditProfileDisplay(tx, me.ID, repository.DisplayGap{
			Mini:   generateDownloadURL(nMini),
			Middle: generateDownloadURL(nMiddle),
			Normal: generateDownloadURL(nNormal),
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	go repository.SetUserToRedis(myRedis, entity.RedisUserModel{
		ID:               me.ID,
		Username:         "",
		FirstName:        me.FirstName,
		LastName:         me.LastName,
		DisplayImage:     generateDownloadURL(nMiddle),
		DisplayImageMini: generateDownloadURL(nMini),
		Level:            entity.UserLevelType(me.GetLevel()),
	})

	res := entity.ResDraftImage{
		URL: generateDownloadURL(nNormal),
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxVerifyToCreatorPostHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	if me.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	imageIDCard, err := ctx.FormFileHeaderNotEmpty("idcard")
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	imageFaceIDCard, err := ctx.FormFileHeaderNotEmpty("face_idcard")
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	imgIDCard, typIDCard, err := service.ImageToBase64(imageIDCard)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	imgFace, typFace, err := service.ImageToBase64(imageFaceIDCard)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.InsertUserRequest(db, me.ID, entity.RequestVerifyIDCard)
	must(err)

	go repository.SendEmailVerifyCreator(me.Email, me.ID, me.FirstName+" "+me.LastName, imgIDCard, imgFace, typIDCard, typFace)

	return ctx.NoContent()
}

func ajaxVerifyBookbankPostHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	if me.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	if !me.IsEmail() || !me.IsVerify || !me.IsVerifyIDCard || me.IsVerifyBookBank {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	bookbankName := ctx.PostFormValueTrimSpace("bookbank_name")
	bookbankNumber := ctx.PostFormValueTrimSpace("bookbank_number")
	bankName := ctx.PostFormValueTrimSpace("bank_name")

	if bookbankName == "" || bookbankNumber == "" || bankName == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	bankID := entity.GetBankIDByName(bankName)
	if bankID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	image, err := ctx.FormFileHeaderNotEmpty("image_bookbank")
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	img, typ, err := service.ImageToBase64(image)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	m, err := service.GetImageFromFile(image)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	URL := generateProfileBookbankName(me.ID)

	err = upload(ctx, m, URL)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	downloadURL := generateDownloadURL(URL)

	err = repository.CheckBookbank(db, me.ID)
	if err == sql.ErrNoRows {

		err := repository.InsertBookbank(db, me.ID, bookbankName, bookbankNumber, bankID, downloadURL)
		must(err)

		err = repository.InsertUserRequest(db, me.ID, entity.RequestVerifyBookBank)
		must(err)

		go repository.SendEmailVerifyBookbank(me.Email, me.ID, me.FirstName+" "+me.LastName, img, typ, bookbankName, bookbankNumber, bankName)

		return ctx.NoContent()
	}
	must(err)

	//repo save UpdateBookbank
	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.UpdateBookbank(tx, me.ID, bookbankName, bookbankNumber, bankID, downloadURL)
		if err != nil {
			return err
		}
		return nil
	})
	must(err)

	err = repository.InsertUserRequest(db, me.ID, entity.RequestVerifyBookBank)
	must(err)

	go repository.SendEmailVerifyBookbank(me.Email, me.ID, me.FirstName+" "+me.LastName, img, typ, bookbankName, bookbankNumber, bankName)

	return ctx.NoContent()
}
