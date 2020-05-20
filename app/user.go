package app

import (
	"database/sql"
	"image"
	"image/color"
	"image/draw"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/acoshift/pgsql"

	"github.com/asaskevich/govalidator"
	"github.com/moonrhythm/hime"
	"github.com/moonrhythm/session"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/repository"
	"github.com/workdestiny/oilbets/service"
)

func userGetHandler(ctx *hime.Context) error {
	id := GetMyID(ctx)

	listGap, err := repository.ListGap(db, id, config.LimitListGap)
	must(err)

	listfollowgap, err := repository.ListFollowUser(db, id, config.LimitListFollowUser+1)
	must(err)

	var next time.Time
	isNext := false

	if len(listfollowgap) > config.LimitListFollowUser {
		isNext = true
		next = listfollowgap[config.LimitListFollowUser-1].CreatedAt
		listfollowgap = listfollowgap[:config.LimitListFollowUser]
	}

	p := page(ctx)
	p["NextLoad"] = next
	p["IsNext"] = isNext
	p["ListGap"] = listGap
	p["ListFollowGap"] = listfollowgap
	p["BankList"] = entity.BankListData
	return ctx.View("app/account", p)
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

	me := getUser(ctx)
	firstName := ctx.PostFormValueTrimSpace("firstname")
	lastName := ctx.PostFormValueTrimSpace("lastname")
	email := strings.ToLower(ctx.PostFormValueTrimSpace("email"))
	oldPassword := ctx.PostFormValue("old-password")
	newPassword := ctx.PostFormValue("new-password")
	repeatPassword := ctx.PostFormValue("repeat-password")
	changePassword := ctx.PostFormValue("change-password")
	changeEmail := ctx.PostFormValue("change-email")

	if utf8.RuneCountInString(firstName) > 50 {
		f.Add("ErrName", "ชื่อต้องไม่เกิน 50 ตัวอักษร")
	}

	if !checkOnlyText(firstName) {
		f.Add("ErrName", "ไม่อนุญาตอักษรพิเศษใน: ชื่อ")
	}

	if !checkReservedWordsName(firstName) {
		f.Add("ErrName", "ไม่อนุญาตคำเฉพาะใน: ชื่อ")
	}

	if utf8.RuneCountInString(lastName) > 50 {
		f.Add("ErrLName", "นามสกุลต้องไม่เกิน 50 ตัวอักษร")
	}

	if !checkOnlyText(lastName) {
		f.Add("ErrLName", "ไม่อนุญาตอักษรพิเศษใน: นามสกุล")
	}

	if !checkReservedWordsName(lastName) {
		f.Add("ErrLName", "ไม่อนุญาตคำเฉพาะใน: นามสกุล")
	}

	if email == "" && changePassword != "true" {
		return ctx.RedirectTo("account")
	}

	if !govalidator.IsEmail(email) {
		f.Add("ErrEmail", "อีเมลไม่ถูกรูปแบบ")
	}

	if oldPassword == "" && newPassword == "" && repeatPassword == "" {
		changePassword = "false"
	}

	if changePassword == "true" {

		if oldPassword == "" {
			f.Add("ErrOldPassword", "กรุณากรอกรหัสผ่าน")
			f.Set("changePassword", changePassword)
			f.Set("NewPassword", newPassword)
			f.Set("RepeatPassword", repeatPassword)
			return ctx.RedirectToGet()
		}

		if newPassword == "" {
			f.Add("ErrInputEmptyNewPassword", "กรุณากรอกรหัสผ่านใหม่")
			f.Set("changePassword", changePassword)
			f.Set("OldPassword", oldPassword)
			f.Set("NewPassword", newPassword)
			f.Set("RepeatPassword", repeatPassword)
			return ctx.RedirectToGet()
		}

		if repeatPassword == "" {
			f.Add("ErrInputEmptyRepeatPassword", "กรุณายืนยันรหัสผ่านใหม่")
			f.Set("changePassword", changePassword)
			f.Set("OldPassword", oldPassword)
			f.Set("NewPassword", newPassword)
			f.Set("RepeatPassword", repeatPassword)
			return ctx.RedirectToGet()
		}

		if newPassword != repeatPassword {
			f.Add("ErrPasswordMissMatch", "รหัสผ่านไม่ตรงกัน")
			f.Set("changePassword", changePassword)
			f.Set("OldPassword", oldPassword)
			return ctx.RedirectToGet()
		}

		if utf8.RuneCountInString(newPassword) < 8 || utf8.RuneCountInString(newPassword) > 20 {
			f.Add("ErrNewPassword", "รหัสผ่านต้องไม่น้อยกว่า 8 และไม่เกิน 20 ตัวอักษร")
			f.Set("changePassword", changePassword)
			f.Set("OldPassword", oldPassword)
			return ctx.RedirectToGet()
		}

	}

	if f.Has("ErrName") || f.Has("ErrLName") || f.Has("ErrEmail") || f.Has("ErrPasswordMissMatch") {
		setFalsh(f, email, firstName, lastName, changePassword)
		return ctx.RedirectToGet()
	}

	if changeEmail == "true" {
		err := repository.CheckEmail(db, email)
		if err != nil {
			f.Add("ErrEmail", "มีอีเมลนี้ในระบบแล้ว")
			setFalsh(f, email, firstName, lastName, changePassword)
			return ctx.RedirectToGet()
		}
	}

	if changeEmail == "true" {

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err := repository.UpdateNewEmail(tx, me.ID, email)
			if err != nil {
				return err
			}

			err = repository.UpdateCodeSendEmailVerify(tx, baseURL, me.ID, email, firstName+" "+lastName)
			if err != nil {
				return err
			}

			return nil
		})
		must(err)
		f.Add("SuccessEmail", true)
	}

	err := repository.EditNameUser(db, me.ID, firstName, lastName)
	if err != nil {
		f.Add("ErrUser", "ไม่สามารถเปลี่ยนข้อมูลได้")
		return ctx.RedirectToGet()
	}

	if changePassword == "true" {

		opw, err := repository.CheckOldPassword(db, me.ID)
		if err == sql.ErrNoRows {
			f.Add("ErrOldPassword", "รหัสผ่านไม่ถูกต้อง")
			f.Set("NewPassword", newPassword)
			f.Set("RepeatPassword", repeatPassword)
			f.Set("changePassword", changePassword)
		}
		must(err)

		if !repository.ComparePassword(opw, oldPassword) {
			f.Add("ErrOldPassword", "รหัสผ่านไม่ถูกต้อง")
			f.Set("NewPassword", newPassword)
			f.Set("RepeatPassword", repeatPassword)
			f.Set("changePassword", changePassword)
			return ctx.RedirectToGet()
		}

		hasPassword, err := repository.HashPassword(newPassword)
		if err != nil {
			f.Add("Password", "ไม่สามารถใช้รหัสผ่านนี้ได้")
			f.Set("changePassword", changePassword)
			return ctx.RedirectToGet()
		}

		err = repository.UpdateNewPassword(db, me.ID, hasPassword)
		if err != nil {
			f.Add("Password", "ไม่สามารถเปลี่ยนรหัสผ่านได้")
			f.Set("changePassword", changePassword)
			return ctx.RedirectToGet()
		}

	}

	go repository.SetUserToRedis(myRedis, entity.RedisUserModel{
		ID:               me.ID,
		Username:         "",
		FirstName:        firstName,
		LastName:         lastName,
		DisplayImage:     me.DisplayImage.Middle,
		DisplayImageMini: me.DisplayImage.Mini,
		Level:            entity.UserLevelType(me.GetLevel()),
	})

	addSuccess(f, "แก้ไขข้อมูลเรียบร้อยแล้ว!")
	return ctx.RedirectTo("account")
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
