package app

import (
	"database/sql"
	"image"
	"image/color"
	"image/draw"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shopspring/decimal"

	"github.com/asaskevich/govalidator"
	"github.com/mssola/user_agent"

	"github.com/acoshift/pgsql"

	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
	"github.com/workdestiny/amlporn/service"
)

func createGapGetHandler(ctx *hime.Context) error {

	role := GetUserRole(ctx)
	if role == entity.RoleUser {
		return ctx.RedirectTo("notfound")
	}

	topic, err := repository.ListTopicVerified(db, role)
	must(err)

	p := page(ctx)
	p["Topic"] = topic

	return ctx.View("app/create-gap", p)
}

func createGapPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	f := getSession(ctx).Flash()

	name := ctx.PostFormValueTrimSpace("name")
	bio := ctx.PostFormValueTrimSpace("bio")
	topicID := ctx.PostFormValue("topic-id")

	if utf8.RuneCountInString(name) == 0 || utf8.RuneCountInString(name) > 50 {
		f.Add("ErrName", "ชื่อเพจต้องไม่เกิน 50 ตัวอักษร")
	}

	if !checkNameGap(name) {
		f.Add("ErrName", "ไม่สามารถใช้อักษรพิเศษได้")
	}

	if !checkReservedWordsUsername(name) {
		f.Add("ErrName", "ไม่สามารถใช้ชื่อเฉพาะได้")
	}

	if utf8.RuneCountInString(bio) > 200 {
		f.Add("ErrBio", "คำอธิบายเพจต้องไม่เกิน 200 ตัวอักษร")
	}

	if f.Has("ErrName") || f.Has("ErrBio") {
		f.Set("NameGap", name)
		f.Set("Bio", bio)
		f.Set("TopicID", topicID)
		return ctx.RedirectToGet()
	}

	if topicID != "" {

		err := repository.CheckTopicID(db, topicID)
		if err == sql.ErrNoRows {
			f.Add("ErrTopic", "Topic ไม่ถูกต้อง")
			return ctx.RedirectToGet()
		}
		must(err)
	}

	var newID string
	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		returnID, err := repository.CreateGap(tx, &repository.CreateGapModal{
			UserID: id,
			Name: repository.NameGap{
				Text: name,
			},
			Bio:     bio,
			TopicID: topicID,
			Display: repository.DisplayGap{
				Mini:   config.ImageGapMini,
				Normal: config.ImageGap,
				Middle: config.ImageGap,
			},
			Cover: repository.CoverGap{
				Normal: config.ImageCoverGap,
				Mini:   config.ImageCoverGapMini,
			},
		})
		if err != nil {
			return err
		}

		newID = returnID

		return nil
	})
	must(err)

	//add to Redis
	go repository.AddNewGapToRedis(myRedis, entity.RedisGapModel{
		ID:               newID,
		UserID:           id,
		Username:         newID,
		Name:             name,
		DisplayImage:     config.ImageGap,
		DisplayImageMini: config.ImageGapMini,
		CountFollower:    0,
		CountPopular:     0,
	})

	return ctx.RedirectTo("gap", newID)
}

func gapGetHandler(ctx *hime.Context) error {
	userID := GetMyID(ctx)
	gapID := getParams(ctx, "gapID")

	if gapID == "" {
		return ctx.RedirectTo("notfound")
	}

	gap, err := repository.GetGap(db, userID, gapID)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	if gap.Username != gapID && gap.Username != "" {
		return ctx.RedirectTo("gap", gap.Username)
	}

	p := page(ctx)
	p["Gap"] = gap
	p["ParamID"] = gap.ID
	p["IsGapOfficial"] = false
	if gap.ID == config.GapOfficialID {
		p["IsGapOfficial"] = true
	}
	p["M"] = setMeta(
		service.ShortTextMeta(70, gap.Name),
		service.ShortTextMeta(170, gap.Bio),
		gap.Display,
		300,
		300,
	)

	if gap.IsOwner {
		// sum popular
		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.UpdateCountPopularGap(tx, gap.ID)
			if err != nil {
				return err
			}

			return nil
		})
		must(err)
	}

	if userID != "" {
		// add view
		go countGapView(gap.ID, userID, ctx.Request.Referer(), ctx.Request.UserAgent())
	}

	if userID == "" {
		userAgent := ctx.Request.UserAgent()
		ua := user_agent.New(userAgent)
		if !ua.Bot() {
			go countGapGuestView(gap.ID, GetVisitorID(ctx), ctx.Request.Referer(), userAgent)
		}
	}

	return ctx.View("app/gap", p)
}

func ajaxGapPostHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	gapID := getParams(ctx, "gapID")

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponsePost{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		_, isOwner := repository.CheckOwnerGap(myRedis, gapID, me.ID)

		if me.IsAdmin() {
			isOwner = true
		}

		if isOwner {

			post, err := repository.ListPostOwnerGap(db, myRedis, me.ID, gapID, config.LimitListPostGap)
			if err != nil {
				return ctx.Status(http.StatusInternalServerError).JSON(nil)
			}

			if len(post) == 0 {
				return ctx.NoContent()
			}

			res.Post = post
			res.Next = post[len(post)-1].CreatedAt
			return ctx.Status(http.StatusOK).JSON(&res)

		}

		post, err := repository.ListPostGap(db, myRedis, me.ID, gapID, config.LimitListPostGap)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(nil)
		}

		if len(post) == 0 {
			return ctx.NoContent()
		}

		res.Post = post
		res.Next = post[len(post)-1].CreatedAt
		return ctx.Status(http.StatusOK).JSON(&res)
	}

	post, err := repository.ListPostGapNextLoad(db, myRedis, me.ID, gapID, req.Next, config.LimitListPostGapNext)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	if len(post) == 0 {
		return ctx.NoContent()
	}

	res.Post = post
	res.Next = post[len(post)-1].CreatedAt

	return ctx.Status(http.StatusOK).JSON(&res)
}

func gapSettingGetHandler(ctx *hime.Context) error {
	userID := getUserID(ctx)
	gapID := getParams(ctx, "gapID")

	gap, err := repository.GetGapSetting(db, myRedis, userID, gapID)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	if gap.Username != gapID && gap.Username != "" {
		return ctx.Redirect("/gap/" + gap.Username + "/setting")
	}

	p := page(ctx)
	p["Gap"] = gap
	p["Province"] = entity.ProvinceData
	return ctx.View("app/gap.setting", p)
}

func gapSettingInfoPostHandler(ctx *hime.Context) error {
	id := getUserID(ctx)
	f := getSession(ctx).Flash()

	name := ctx.PostFormValueTrimSpace("name")
	bio := service.StripHTML(ctx.PostFormValueTrimSpace("bio"))
	isEditName := ctx.PostFormValueTrimSpace("edit-name")
	gapID := ctx.PostFormValue("id")

	_, isErr := repository.CheckOwnerGap(myRedis, gapID, id)
	if !isErr {
		f.Add("ErrOwner", "คุณไม่ใช่เจ้าของเพจ")
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	if utf8.RuneCountInString(bio) > 200 {
		f.Add("ErrBio", "คำอธิบายเพจต้องไม่เกิน 200 ตัวอักษร")
	}

	if isEditName == "true" {

		if utf8.RuneCountInString(name) == 0 || utf8.RuneCountInString(name) > 50 {
			f.Add("ErrName", "ไม่สามารถเปลี่ยนชื่อเพจได้")
		}

		if !checkReservedWordsUsername(name) {
			f.Add("ErrName", "ไม่สามารถใช้ชื่อเฉพาะได้")
		}

		if !checkNameGap(name) {
			f.Add("ErrName", "ไม่สามารถใช้อักษรพิเศษได้")
		}

	}

	if f.Has("ErrBio") || f.Has("ErrName") {
		f.Set("Name", name)
		f.Set("Bio", bio)
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	err := repository.EditGapBio(db, gapID, id, bio)
	if err != nil {
		f.Add("ErrBio", "ไม่สามารถเปลี่ยนคำอธิบายเพจได้")
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	if isEditName == "true" {

		nameGap, err := repository.CheckTimeEditGapName(db, gapID, id)
		if err != nil {
			f.Add("ErrName", "ไม่สามารถเปลี่ยนชื่อเพจได้ ต้องรอหลังจาก 60 วัน ที่เปลี่ยนชื่อครั้งล่าสุด")
			return ctx.Redirect("/gap/" + gapID + "/setting")
		}

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.EditGapName(tx, gapID, id, name)
			if err != nil {
				return NewAppError("ไม่สามารถแก้ไขชื่อเพจได้")
			}

			return nil
		})
		if IsAppError(err) {
			f.Add("ErrName", err.Error())
			return ctx.Redirect("/gap/" + gapID + "/setting")
		}

		gap, err := repository.GetGapDetail(db, gapID)
		must(err)

		//set to Redis
		go repository.SetGapToRedis(myRedis, entity.RedisGapModel{
			ID:               gapID,
			UserID:           gap.UserID,
			Username:         gap.Username,
			Name:             name,
			DisplayImage:     gap.Display,
			DisplayImageMini: gap.DisplayMini,
			CountFollower:    gap.CountFollower,
			CountPopular:     gap.CountPopular,
		}, nameGap)

	}

	f.Add("SuccessInfo", "บันทึกเรียบร้อย")
	return ctx.Redirect("/gap/" + gapID + "/setting")
}

func gapSettingUserNamePostHandler(ctx *hime.Context) error {

	f := getSession(ctx).Flash()

	id := getUserID(ctx)
	isEditUsername := ctx.PostFormValueTrimSpace("edit-username")
	gapID := ctx.PostFormValue("id")
	username := ctx.PostFormValueTrimSpace("username")
	username = strings.ToLower(username)

	_, isErr := repository.CheckOwnerGap(myRedis, gapID, id)
	if !isErr {
		f.Add("ErrOwner", "คุณไม่ใช่เจ้าของเพจ")
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	if isEditUsername == "true" {

		if username == "" {
			f.Add("ErrUsername", "กรุณากรอก Username gap ของท่าน")
			f.Set("Username", "")
			return ctx.Redirect("/gap/" + gapID + "/setting")
		}

		if utf8.RuneCountInString(username) > 30 {
			f.Add("ErrUsername", "Username gap ต้องไม่เกิน 30 ตัวอักษร")
			return ctx.Redirect("/gap/" + gapID + "/setting")
		}

		if !checkUsername(username) {
			f.Add("ErrUsername", "ต้องขึ้นต้นด้วยตัวอักษรและไม่สามารถใช้อักษรพิเศษได้")
		}

		if !checkReservedWordsUsername(username) {
			f.Add("ErrUsername", "ไม่สามารถใช้ Username gap นี้ได้")
		}

	}

	if f.Has("ErrUsername") {
		f.Set("Username", username)
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	if isEditUsername == "true" {

		swap, err := repository.CheckGapUsername(db, gapID, username)
		if err != nil {
			f.Add("ErrUsername", err.Error())
			return ctx.Redirect("/gap/" + gapID + "/setting")
		}

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.EditGapUsername(tx, gapID, id, username, swap)
			if err != nil {
				return NewAppError("ไม่สามารถเปลี่ยน Username gap ได้")
			}

			return nil
		})
		if IsAppError(err) {
			f.Add("ErrUsername", err.Error())
			return ctx.Redirect("/gap/" + gapID + "/setting")
		}

		gap, err := repository.GetGapDetail(db, gapID)
		must(err)

		//set to Redis
		go repository.SetGapToRedis(myRedis, entity.RedisGapModel{
			ID:               gapID,
			UserID:           gap.UserID,
			Username:         username,
			Name:             gap.Name,
			DisplayImage:     gap.Display,
			DisplayImageMini: gap.DisplayMini,
			CountFollower:    gap.CountFollower,
			CountPopular:     gap.CountPopular,
		}, "")
	}

	f.Add("SuccessUsername", "บันทึกเรียบร้อย")
	return ctx.Redirect("/gap/" + gapID + "/setting")
}

func gapSettingContactPostHandler(ctx *hime.Context) error {

	id := getUserID(ctx)
	f := getSession(ctx).Flash()

	tel := ctx.PostFormValueTrimSpace("tel")
	social := ctx.PostFormValueTrimSpace("social")
	email := ctx.PostFormValueTrimSpace("email")
	gapID := ctx.PostFormValue("id")

	_, isErr := repository.CheckOwnerGap(myRedis, gapID, id)
	if !isErr {
		f.Add("ErrOwner", "คุณไม่ใช่เจ้าของเพจ")
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	if utf8.RuneCountInString(tel) > 20 {
		f.Add("ErrTel", "เบอร์โทรศัพท์ต้องไม่เกิน 20 ตัวอักษร")
	}

	if email != "" {
		if !govalidator.IsEmail(email) {
			f.Add("ErrEmail", "อีเมลไม่ถูกรูปแบบ")
		}
	}

	if f.Has("ErrTel") || f.Has("ErrEmail") {
		f.Set("Tel", tel)
		f.Set("Social", social)
		f.Set("Email", email)
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.EditGapContact(db, gapID, tel, email, social)
		if err != nil {
			return NewAppError("ไม่สามารถเปลี่ยนข้อมูลการติดต่อได้")
		}

		return nil
	})
	if IsAppError(err) {
		f.Add("ErrContact", err.Error())
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	f.Add("SuccessContact", "บันทึกเรียบร้อย")
	return ctx.Redirect("/gap/" + gapID + "/setting")
}

func gapSettingAddressPostHandler(ctx *hime.Context) error {

	id := getUserID(ctx)
	f := getSession(ctx).Flash()

	city := ctx.PostFormValueTrimSpace("city")
	country := ctx.PostFormValueTrimSpace("country")
	address := ctx.PostFormValueTrimSpace("address")
	gapID := ctx.PostFormValue("id")

	_, isErr := repository.CheckOwnerGap(myRedis, gapID, id)
	if !isErr {
		f.Add("ErrOwner", "คุณไม่ใช่เจ้าของเพจ")
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.EditGapAddress(tx, gapID, address, city, country)
		if err != nil {
			return NewAppError("ไม่สามารถเปลี่ยนข้อมูลสถานที่ได้")
		}

		return nil
	})
	if IsAppError(err) {
		f.Add("ErrAddress", err.Error())
		return ctx.Redirect("/gap/" + gapID + "/setting")
	}

	f.Add("SuccessAddress", "บันทึกเรียบร้อย")
	return ctx.Redirect("/gap/" + gapID + "/setting")
}

func ajaxUploadGapDisplayPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	gapID := getParams(ctx, "gapID")

	_, isErr := repository.CheckOwnerGap(myRedis, gapID, id)
	if !isErr {
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

	nMini, nMiddle, nNormal := generateGapDisplayName(id)

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

		err = repository.EditGapDisplay(tx, gapID, id, repository.DisplayGap{
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

	gap, err := repository.GetGapDetail(db, gapID)
	must(err)

	//set to Redis
	go repository.SetGapToRedis(myRedis, entity.RedisGapModel{
		ID:               gapID,
		UserID:           gap.UserID,
		Username:         gap.Username,
		Name:             gap.Name,
		DisplayImage:     generateDownloadURL(nMiddle),
		DisplayImageMini: generateDownloadURL(nMini),
		CountFollower:    gap.CountFollower,
		CountPopular:     gap.CountPopular,
	}, "")

	res := entity.ResDraftImage{
		URL: generateDownloadURL(nNormal),
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxUploadGapCoverPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	gapID := getParams(ctx, "gapID")

	_, isErr := repository.CheckOwnerGap(myRedis, gapID, id)
	if !isErr {
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
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), m, m.Bounds().Min, draw.Over)

	mini, normal := resizeUploadCoverImage(dst)

	nMini, nNormal := generateGapCoverName(id)

	err = upload(ctx, mini, nMini)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	err = upload(ctx, normal, nNormal)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.EditGapCover(tx, gapID, id, repository.CoverGap{
			Mini:   generateDownloadURL(nMini),
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

	res := entity.ResDraftImage{
		URL: generateDownloadURL(nNormal),
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxFollowGapPostHandler(ctx *hime.Context) error {
	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	if req.ID == config.GapOfficialID {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.CheckGap(db, req.ID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResFollow{
		IsFollow: false,
	}

	follow, err := repository.CheckFollowGap(db, id, req.ID)
	if err == sql.ErrNoRows {

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.CreateFollowGap(tx, id, req.ID)
			if err != nil {
				return NewAppError("ไม่สามารถติดตามเพจได้")
			}

			return nil
		})
		if IsAppError(err) {
			return ctx.JSON(map[string]interface{}{
				"Error": err.Error(),
			})
		}
		must(err)

		go repository.AddNotificationFollow(db, myRedis, id, req.ID)

		res.IsFollow = true
		return ctx.Status(http.StatusOK).JSON(&res)
	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {
		err = repository.FollowGap(tx, id, req.ID, !follow)

		if err != nil {
			return NewAppError("ไม่สามารถติดตามเพจได้")
		}

		return nil
	})
	if IsAppError(err) {
		return ctx.JSON(map[string]interface{}{
			"Error": err.Error(),
		})
	}
	must(err)

	res.IsFollow = !follow

	if res.IsFollow {
		go repository.AddNotificationFollow(db, myRedis, id, req.ID)
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxListUserFollowerGapPostHandler(ctx *hime.Context) error {

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.CheckGap(db, req.ID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponseUserFollowerGapModel{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		list, err := repository.ListUserFollowrGap(db, req.ID, config.LimitListUserFollowrGap+1)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(nil)
		}

		if len(list) == 0 {
			return ctx.NoContent()
		}

		if len(list) > config.LimitListUserFollowrGap {
			res.IsNext = true
			res.Next = list[config.LimitListUserFollowrGap-1].CreatedAt
			list = list[:config.LimitListUserFollowrGap]
		}

		res.User = list
		return ctx.Status(http.StatusOK).JSON(&res)
	}

	list, err := repository.ListUserFollowrGapNextLoad(db, req.ID, req.Next, config.LimitListUserFollowrGap+1)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	if len(list) == 0 {
		return ctx.NoContent()
	}

	if len(list) > config.LimitListUserFollowrGap {
		res.IsNext = true
		res.Next = list[config.LimitListUserFollowrGap-1].CreatedAt
		list = list[:config.LimitListUserFollowrGap]
	}

	res.User = list
	return ctx.Status(http.StatusOK).JSON(&res)
}

//countGapView update view count
func countGapView(gapID string, userID string, referrer string, userAgent string) {

	t, err := repository.GetGapView(db, gapID, userID)
	if err != nil && err != sql.ErrNoRows {
		return
	}

	if err == nil {
		now := time.Now().UTC().Unix()
		if (now - t.Unix()) <= 86400 {
			return
		}
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.CreateViewGap(tx, gapID, userID, referrer, userAgent)
		if err != nil {
			return err
		}

		err = repository.UpdateCountViewGap(tx, gapID)
		if err != nil {
			return err
		}

		return nil
	})
}

//countGapGuestView update guestView count gap
func countGapGuestView(gapID string, vsID string, referrer string, userAgent string) {

	t, err := repository.GetGapGuestView(db, gapID, vsID)
	if err != nil && err != sql.ErrNoRows {
		return
	}

	if err == nil {
		now := time.Now().UTC().Unix()
		if (now - t.Unix()) <= config.LimitDurationGuestView {
			return
		}
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.CreateGuestViewGap(tx, gapID, vsID, referrer, userAgent)
		if err != nil {
			return err
		}

		err = repository.UpdateCountGuestViewGap(tx, gapID)
		if err != nil {
			return err
		}

		return nil
	})
}

func checkNameGap(text string) bool {
	Re := regexp.MustCompile(`^[\sa-zA-Zก-๙0-9._-]+$`)
	return Re.MatchString(text)
}

func gapInsightsGetHandler(ctx *hime.Context) error {
	userID := GetMyID(ctx)
	gapID := getParams(ctx, "gapID")
	section := ctx.FormValue("section")
	typ := ctx.FormValue("type")
	from := ctx.FormValue("from")
	to := ctx.FormValue("to")
	var startTime, endTime time.Time

	if gapID == "" {
		return ctx.RedirectTo("notfound")
	}

	gap, err := repository.GetGapStatistic(db, myRedis, userID, gapID)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	if gap.Username != gapID && gap.Username != "" {
		return ctx.Redirect("/gap/" + gap.Username + "/insights")
	}

	switch section {
	case "overview":
		startTime, endTime = service.GetParserTime("all", "", "")

	case "advance":
		if typ == "" {
			typ = "thismonth"
			startTime, endTime = service.GetParserTime(typ, "", "")
			break
		}
		startTime, endTime = service.GetParserTime(typ, from, to)

	default:
		section = "overview"
		startTime, endTime = service.GetParserTime("all", "", "")
	}

	ViewPost, ViewGap, err := repository.GetCountView(db, gap.ID, startTime, endTime)
	must(err)

	post, err := repository.ListPostCountView(db, gap.ID, 0, startTime, endTime)
	must(err)

	p := page(ctx)
	if section == "advance" {
		percentUserAgent := percentUserAgent(gap.ID, startTime, endTime)
		p["PercentUserAgent"] = percentUserAgent

		listCountViewHour := listCountViewHour(gap.ID, startTime, endTime)
		p["ListCountViewHour"] = listCountViewHour

		if typ == "custom" {
			p["FormatToTime"] = service.FormatCustomType(endTime)
			p["FormatFromTime"] = service.FormatCustomType(startTime)
			p["ToTime"] = to
			p["FromTime"] = from
		}
	}

	p["Gap"] = gap
	p["ViewPost"] = ViewPost
	p["ViewGap"] = ViewGap
	p["ViewAll"] = ViewPost.All + ViewGap.All
	p["Post"] = post
	p["Section"] = section
	p["Type"] = typ
	return ctx.View("app/gap.insights", p)
}

func gapRevenueGetHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	gapID := getParams(ctx, "gapID")

	gap, err := repository.GetGapRevenue(db, me.ID, gapID)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	if gap.Username != gapID && gap.Username != "" {
		return ctx.Redirect("/gap/" + gap.Username + "/revenue")
	}

	t := time.Now().UTC()
	View, err := repository.GetCountViewRevenue(db, gap.ID, t)
	must(err)

	wallets, err := decimal.NewFromString(gap.Wallets)
	must(err)

	View.ViewAmount = View.AmountView()
	View.GuestViewAmount = View.AmountGuestView()
	View.AllAmount = View.AmountAll().Add(wallets)
	fa, _ := View.AllAmount.Float64()
	View.Percent = math.Round(View.AmountPercent()*100) / 100

	Bookbank, err := repository.GetBookbank(db, me.ID)
	must(err)

	post, err := repository.ListPostCountViewRevenue(db, gap.ID, config.LimitListPostCountViewRevenue+1)
	must(err)

	var next time.Time
	isNext := false
	if len(post) > config.LimitListPostCountViewRevenue {
		isNext = true
		next = post[config.LimitListPostCountViewRevenue-1].CreatedAt
		post = post[:config.LimitListPostCountViewRevenue]
	}

	isRevenue, err := repository.CheckRevenue(db, gap.ID)
	must(err)

	if isRevenue {
		if fa < config.MinimumPay {
			isRevenue = false
		}
	}

	now := time.Now()
	p := page(ctx)
	p["Gap"] = gap
	p["View"] = View
	p["Bookbank"] = Bookbank
	p["DateTime"] = service.FormatRevenueDateType(now)
	p["DateTimeHour"] = service.FormatRevenueTimeType(now)
	p["Post"] = post
	p["IsRevenue"] = isRevenue
	p["NextLoad"] = next
	p["IsNext"] = isNext
	return ctx.View("app/gap.revenue", p)
}

func getMonthNow() time.Time {
	t := time.Now()
	day := t.Day()
	return t.AddDate(0, 0, -(day - 1))
}

func ajaxListPostCountViewHandler(ctx *hime.Context) error {

	userID := GetMyID(ctx)
	if userID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestGapStatistic
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	_, isOwner := repository.CheckOwnerGap(myRedis, req.ID, userID)
	if !isOwner {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	startTime, endTime := service.GetParserTime(req.Type, req.Start, req.End)

	post, err := repository.ListPostCountView(db, req.ID, (req.Page*config.LimitListPostCountView)-config.LimitListPostCountView, startTime, endTime)
	must(err)

	if len(post) == 0 {
		return ctx.NoContent()
	}

	res := entity.ResponseGapStatistic{
		Post: post,
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxListPostCountViewRevenueHandler(ctx *hime.Context) error {
	userID := GetMyID(ctx)

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		log.Println("json")
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	_, isOwner := repository.CheckOwnerGap(myRedis, req.ID, userID)
	if !isOwner {
		log.Println("gap")
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponsePostRevenue{}

	post, err := repository.ListPostCountViewRevenueNextLoad(db, req.ID, req.Next, config.LimitListPostCountViewRevenue)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	if len(post) == 0 {
		return ctx.NoContent()
	}

	res.Next = post[len(post)-1].CreatedAt
	res.Post = post

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxRevenuePostHandler(ctx *hime.Context) error {
	me := getUser(ctx)

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	gap, err := repository.GetGapWallets(db, req.ID)
	must(err)
	if gap.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	t := time.Now().UTC()
	View, err := repository.GetCountViewRevenue(db, gap.ID, t)
	must(err)

	wallets, err := decimal.NewFromString(gap.Wallets)
	must(err)

	amount := View.AmountAll().Add(wallets)
	fa, _ := amount.Float64()
	if fa < config.MinimumPay {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.RegisterRevenue(db, t, gap.ID, amount.String(), config.RevenueRateView.String(), config.RevenueRateGuestView.String())
	must(err)

	go repository.SendEmailRevenue(me.ID, me.FirstName+" "+me.LastName, me.Email, gap.ID, gap.Name)

	return ctx.NoContent()
}

func percentUserAgent(gapID string, a time.Time, b time.Time) entity.GetUserAgentViewModel {

	agentView := entity.GetUserAgentViewModel{}
	agentPostView, err := repository.ListPostCountUserAgentView(db, gapID, a, b)
	must(err)

	agentPostGuestView, err := repository.ListPostCountUserAgentGuestView(db, gapID, a, b)
	must(err)

	agentGapView, err := repository.ListGapCountUserAgentView(db, gapID, a, b)
	must(err)

	agentGapGuestView, err := repository.ListGapCountUserAgentGuestView(db, gapID, a, b)
	must(err)

	agentView.Desktop = agentPostView.Desktop + agentPostGuestView.Desktop + agentGapView.Desktop + agentGapGuestView.Desktop
	agentView.Mobile = agentPostView.Mobile + agentPostGuestView.Mobile + agentGapView.Mobile + agentGapGuestView.Mobile
	agentView.DesktopPercent = agentView.CalculatePercentDesktop()
	agentView.MobilePercent = agentView.CalculatePercentMobile()

	return agentView
}

func listCountViewHour(gapID string, a time.Time, b time.Time) []entity.ViewHour {

	cv, err := repository.ListCountViewHour(db, gapID, a, b)
	must(err)
	cg, err := repository.ListCountGuestViewHour(db, gapID, a, b)
	must(err)
	cgv, err := repository.ListCountGapViewHour(db, gapID, a, b)
	must(err)
	cgg, err := repository.ListCountGapGuestViewHour(db, gapID, a, b)
	must(err)

	lc := []entity.ViewHour{
		{Hour: "00:00", Count: 0},
		{Hour: "01:00", Count: 0},
		{Hour: "02:00", Count: 0},
		{Hour: "03:00", Count: 0},
		{Hour: "04:00", Count: 0},
		{Hour: "05:00", Count: 0},
		{Hour: "06:00", Count: 0},
		{Hour: "07:00", Count: 0},
		{Hour: "08:00", Count: 0},
		{Hour: "09:00", Count: 0},
		{Hour: "10:00", Count: 0},
		{Hour: "11:00", Count: 0},
		{Hour: "12:00", Count: 0},
		{Hour: "13:00", Count: 0},
		{Hour: "14:00", Count: 0},
		{Hour: "15:00", Count: 0},
		{Hour: "16:00", Count: 0},
		{Hour: "17:00", Count: 0},
		{Hour: "18:00", Count: 0},
		{Hour: "19:00", Count: 0},
		{Hour: "20:00", Count: 0},
		{Hour: "21:00", Count: 0},
		{Hour: "22:00", Count: 0},
		{Hour: "23:00", Count: 0},
	}

	for _, u := range cv {
		lc[u.Hour].Count += u.Count
	}

	for _, u := range cg {
		lc[u.Hour].Count += u.Count
	}

	for _, u := range cgv {
		lc[u.Hour].Count += u.Count
	}

	for _, u := range cgg {
		lc[u.Hour].Count += u.Count
	}

	return lc
}
