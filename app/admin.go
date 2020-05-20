package app

import (
	"database/sql"
	"image"
	"image/color"
	"image/draw"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/acoshift/pgsql"
	"github.com/asaskevich/govalidator"
	"github.com/moonrhythm/hime"
	"github.com/shopspring/decimal"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/repository"
	"github.com/workdestiny/oilbets/service"
)

func adminIndexGetHandler(ctx *hime.Context) error {
	return ctx.RedirectTo("admin.verify")
}

func adminVerifyGetHandler(ctx *hime.Context) error {

	lrq, err := repository.ListUserRequest(db)
	must(err)

	p := page(ctx)
	p["ListVerify"] = lrq
	return ctx.View("admin/verify", p)
}

func adminVerifyDetailGetHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	email := ctx.FormValue("email")
	if !govalidator.IsEmail(email) {
		f.Add("ErrEmail", "อีเมลไม่ถูกรูปแบบ")
		return ctx.RedirectToGet()
	}

	user := repository.AdminGetUser(db, email)
	if user.ID == "" {
		f.Add("ErrUser", "ไม่มีผู้ใช้ในระบบ")
		return ctx.RedirectToGet()
	}

	p := page(ctx)
	p["UserData"] = user
	return ctx.View("admin/verify-detail", p)
}

func adminVerifyPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()

	email := strings.ToLower(ctx.PostFormValueTrimSpace("email"))
	if !govalidator.IsEmail(email) {
		f.Add("ErrEmail", "อีเมลไม่ถูกรูปแบบ")
		return ctx.RedirectToGet()
	}

	lrq, err := repository.SearchListUserRequest(db, email)
	must(err)

	p := page(ctx)
	p["ListVerify"] = lrq
	return ctx.View("admin/verify", p)
}

func adminVerifyIDCardHandler(ctx *hime.Context) error {

	f := getSession(ctx).Flash()

	userID := ctx.PostFormValue("id")

	err := repository.AdminCheckUser(db, userID)
	if err != nil {
		f.Add("ErrUser", "ไม่มีผู้ใช้ในระบบ")
		return ctx.RedirectTo("admin.verify")
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		gapIDs := repository.ListGapID(tx, userID)
		if err != nil {
			return err
		}

		err := repository.VerifyIDCard(tx, userID)
		if err != nil {
			return err
		}

		err = repository.UpdatePostVerifyIDCard(tx, gapIDs)
		if err != nil {
			return err
		}

		err = repository.AdminUpdateListVerify(tx, userID, entity.RequestVerifyIDCard)
		if err != nil {
			return err
		}
		return nil
	})
	must(err)

	user := repository.GetUser(db, userID)
	if user.ID == "" {
		f.Add("ErrUser", "ไม่มีผู้ใช้ในระบบ")
		return ctx.RedirectToGet()
	}

	addSuccess(f, "แก้ไขข้อมูลเรียบร้อยแล้ว!")
	return ctx.RedirectTo("admin.verify")
}

func adminVerifyBookbankHandler(ctx *hime.Context) error {

	f := getSession(ctx).Flash()

	userID := ctx.PostFormValue("id")

	err := repository.AdminCheckUser(db, userID)
	if err != nil {
		f.Add("ErrUser", "ไม่มีผู้ใช้ในระบบ")
		return ctx.RedirectTo("admin.verify")
	}

	_, err = repository.GetBookbank(db, userID)
	if err != nil {
		f.Add("ErrUser", "ผู้ใช้ไม่มีข้อมูล bookbank")
		return ctx.RedirectTo("admin.verify")
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {
		err := repository.VerifyBookbank(tx, userID)
		if err != nil {
			return err
		}

		err = repository.AdminUpdateListVerify(tx, userID, entity.RequestVerifyBookBank)
		if err != nil {
			return err
		}
		return nil
	})
	must(err)

	addSuccess(f, "แก้ไขข้อมูลเรียบร้อยแล้ว!")
	return ctx.RedirectTo("admin.verify")
}

func adminRevenueGetHandler(ctx *hime.Context) error {

	list, err := repository.ListRevenue(db)
	must(err)

	p := page(ctx)
	p["ListRevenue"] = list
	return ctx.View("admin/revenue", p)
}

func adminRevenuePostHandler(ctx *hime.Context) error {

	text := ctx.FormValueTrimSpace("text")

	list, err := repository.SearchListRevenue(db, text)
	must(err)

	p := page(ctx)
	p["TextSearch"] = text
	p["ListRevenue"] = list
	return ctx.View("admin/revenue", p)
}

func adminRevenueDetailGetHandler(ctx *hime.Context) error {

	revenueID := ctx.FormValue("id")

	revenue, err := repository.GetRevenue(db, revenueID)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	gap, err := repository.AdminGetGapRevenue(db, revenue.Gap.ID)
	must(err)

	View, err := repository.GetCountViewRevenue(db, gap.ID, revenue.CreatedAt)
	must(err)

	wallets, err := decimal.NewFromString(gap.Wallets)
	must(err)

	View.ViewAmount = View.AmountView()
	View.GuestViewAmount = View.AmountGuestView()
	View.AllAmount = View.AmountAll().Add(wallets)
	View.Percent = math.Round(View.AmountPercent()*100) / 100

	Bookbank, err := repository.GetBookbank(db, gap.UserID)
	must(err)

	post, err := repository.AdminListPostCountViewRevenue(db, gap.ID, revenue.CreatedAt, config.LimitListPostCountViewRevenue+1)
	must(err)

	var next time.Time
	isNext := false
	if len(post) > config.LimitListPostCountViewRevenue {
		isNext = true
		next = post[config.LimitListPostCountViewRevenue-1].CreatedAt
		post = post[:config.LimitListPostCountViewRevenue]
	}

	p := page(ctx)
	p["Gap"] = gap
	p["View"] = View
	p["Bookbank"] = Bookbank
	p["DateTime"] = service.FormatRevenueDateType(revenue.CreatedAt)
	p["DateTimeHour"] = service.FormatRevenueTimeType(revenue.CreatedAt)
	p["Post"] = post
	p["NextLoad"] = next
	p["IsNext"] = isNext
	p["Revenue"] = revenue
	return ctx.View("admin/revenue-detail", p)
}

func ajaxAdminListPostCountViewRevenueHandler(ctx *hime.Context) error {

	var req entity.RequestRevenue
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponsePostRevenue{}

	post, err := repository.AdminListPostCountViewRevenueNextLoad(db, req.ID, req.Next, req.Time, config.LimitListPostCountView+1)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	if len(post) == 0 {
		return ctx.NoContent()
	}

	if len(post) > config.LimitListPostCountViewRevenue {
		res.Next = post[len(post)-1].CreatedAt
		post = post[:config.LimitListPostCountViewRevenue]
	}

	res.Post = post
	res.Next = post[len(post)-1].CreatedAt

	return ctx.Status(http.StatusOK).JSON(&res)
}

func adminRevenueHistoryGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("admin/revenue-history", p)
}

func ajaxAdminRejectPostStatusRevenuePostHandler(ctx *hime.Context) error {

	var req entity.RequestReject
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	email, title, err := repository.GetUserByPostID(db, req.ID)
	if err == sql.ErrNoRows {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}
	must(err)

	err = repository.AdminCheckPostNote(db, req.ID)
	if err == sql.ErrNoRows {
		//insert
		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err := repository.AdminUpdatePostStatusRevenue(tx, req.ID)
			if err != nil {
				return err
			}

			err = repository.AdminCreatePostNote(tx, req.ID, req.Text)
			if err != nil {
				return err
			}

			return nil
		})
		must(err)

		go repository.SendEmailAdminReject(email, title, req.Text)

		return ctx.NoContent()
	}
	must(err)

	//update
	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.AdminUpdatePostStatusRevenue(tx, req.ID)
		if err != nil {
			return err
		}

		err = repository.AdminUpdatePostNote(tx, req.ID, req.Text)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	go repository.SendEmailAdminReject(email, title, req.Text)

	return ctx.NoContent()
}

func ajaxAdminDeletePostPostHandler(ctx *hime.Context) error {

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	email, title, err := repository.GetUserByPostID(db, req.ID)
	if err == sql.ErrNoRows {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.AdminDeletePost(tx, req.ID)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	go repository.SendEmailAdminDelete(email, title)

	return ctx.NoContent()
}

func ajaxAdminApproveRevenuePostHandler(ctx *hime.Context) error {

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	revenue, err := repository.GetRevenue(db, req.ID)
	if err == sql.ErrNoRows {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}
	must(err)

	gap, err := repository.AdminGetGapRevenue(db, revenue.Gap.ID)
	must(err)

	View, err := repository.GetCountViewRevenue(db, gap.ID, revenue.CreatedAt)
	must(err)

	wallets, err := decimal.NewFromString(gap.Wallets)
	must(err)

	View.AllAmount = View.AmountAll().Add(wallets)

	pay := View.AllAmount.Truncate(2)

	wallet := View.AllAmount.Sub(pay)

	postIDs, err := repository.AdminListPostID(db, gap.ID)
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.AdminUpdateGapWallet(tx, wallet.String(), gap.ID)
		if err != nil {
			return err
		}

		err = repository.AdminUpdateRevenue(tx, revenue.ID, pay.String(), View.View, View.GuestView, entity.Approve)
		if err != nil {
			return err
		}

		err = repository.AdminUpdateViewRevenue(tx, postIDs, revenue.ID, revenue.CreatedAt)
		if err != nil {
			return err
		}

		err = repository.AdminUpdateGuestViewRevenue(tx, postIDs, revenue.ID, revenue.CreatedAt)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.NoContent()
}

func ajaxAdminRejectRevenuePostHandler(ctx *hime.Context) error {

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	revenue, err := repository.GetRevenue(db, req.ID)
	if err == sql.ErrNoRows {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.AdminUpdateRevenue(tx, revenue.ID, revenue.Total.String(), 0, 0, entity.Reject)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.NoContent()
}

func adminVerifyUserHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("admin/user", p)
}

func adminVerifyUserPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	email := ctx.PostFormValueTrimSpace("email")
	if !govalidator.IsEmail(email) {
		f.Add("ErrEmail", "อีเมลไม่ถูกรูปแบบกรุณาตรวจสอบ")
		return ctx.RedirectToGet()
	}

	user := repository.AdminGetUser(db, email)
	if user.ID == "" {
		f.Add("ErrUser", "ไม่มีผู้ใช้ในระบบ")
		return ctx.RedirectToGet()
	}

	p := page(ctx)
	p["UserData"] = user
	return ctx.View("admin/user", p)
}

func adminGapRecommendGetHandler(ctx *hime.Context) error {
	gapRecommend, err := repository.AdminListGapRecommend(db)
	must(err)

	p := page(ctx)
	p["GapRecommend"] = gapRecommend
	return ctx.View("admin/gap-recommend", p)
}

func adminGapRecommendPostHandler(ctx *hime.Context) error {
	text := ctx.PostFormValueTrimSpace("text")

	gap, err := repository.AdminSearchGapRecommend(db, text, 10)
	must(err)

	gapRecommend, err := repository.AdminListGapRecommend(db)
	must(err)

	p := page(ctx)
	p["GapRecommend"] = gapRecommend
	p["Gap"] = gap
	p["TextSearch"] = text
	return ctx.View("admin/gap-recommend", p)
}

func adminAddGapRecommendPostHandler(ctx *hime.Context) error {
	id := ctx.PostFormValueTrimSpace("id")

	err := repository.CheckGap(db, id)
	if err != nil {
		return ctx.RedirectTo("admin.gap.recommend")
	}

	err = repository.CheckGapRecommend(db, id)
	if err == nil {
		return ctx.RedirectTo("admin.gap.recommend")
	}
	if err != sql.ErrNoRows {
		must(err)
	}

	err = repository.AdminAddGapRecommend(db, id)
	must(err)

	return ctx.RedirectTo("admin.gap.recommend")
}

func adminDeleteGapRecommendPostHandler(ctx *hime.Context) error {
	id := ctx.PostFormValueTrimSpace("id")

	count, err := repository.AdminCheckCountGapRecommend(db)
	must(err)
	if count <= config.LimitGapRecommend {
		return ctx.RedirectTo("admin.gap.recommend")
	}

	err = repository.AdminDeleteGapRecommend(db, id)
	must(err)

	return ctx.RedirectTo("admin.gap.recommend")
}

func adminCategoryGetHandler(ctx *hime.Context) error {

	c, err := repository.AdminListCategory(db)
	must(err)

	p := page(ctx)
	p["Category"] = c
	return ctx.View("admin/category", p)
}

func adminCategoryCreateGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("admin/category-create", p)
}

func adminCategoryCreatePostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	name := ctx.PostFormValueTrimSpace("name")

	if utf8.RuneCountInString(name) > 50 {
		f.Add("Errors", "ไม่เกิน 50 ตัว")
		return ctx.RedirectToGet()
	}

	err := repository.AdminCheckCategoryCode(db, name)
	if err != sql.ErrNoRows {
		must(err)
	}
	if err == nil {
		f.Add("Errors", "ไม่สามารถใช้ชื่อนี้ได้")
		return ctx.RedirectToGet()
	}

	err = repository.AdminCreateCategory(db, name)
	must(err)

	return ctx.RedirectTo("admin.category")
}

func adminCategoryEditGetHandler(ctx *hime.Context) error {
	id := ctx.FormValue("id")
	name := ctx.FormValue("name")

	p := page(ctx)
	p["ID"] = id
	p["Name"] = name
	return ctx.View("admin/category-edit", p)
}

func adminCategoryEditPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	id := ctx.PostFormValueTrimSpace("id")
	name := ctx.PostFormValueTrimSpace("name")

	if utf8.RuneCountInString(name) > 50 {
		f.Add("Errors", "ไม่เกิน 50 ตัว")
		return ctx.Redirect("/admin/category/edit", ctx.Param("id", id))
	}

	err := repository.AdminCheckCategoryCode(db, name)
	if err != sql.ErrNoRows {
		must(err)
	}
	if err == nil {
		f.Add("Errors", "ไม่สามารถใช้ชื่อนี้ได้")
		return ctx.Redirect("/admin/category/edit", ctx.Param("id", id))
	}

	err = repository.CheckCategory(db, id)
	if err == sql.ErrNoRows {
		f.Add("Errors", "ไอดีนี้ไม่ถูกต้อง")
		return ctx.Redirect("/admin/category/edit", ctx.Param("id", id))
	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.AdminUpdateCategory(tx, id, name)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.RedirectTo("admin.category")
}

func adminTopicGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("admin/topic", p)
}

func adminTopicPostHandler(ctx *hime.Context) error {
	text := ctx.PostFormValueTrimSpace("text")

	topic, err := repository.AdminSearchTopic(db, text)
	must(err)

	p := page(ctx)
	p["Topic"] = topic
	return ctx.View("admin/topic", p)
}

func adminTopicCreateGetHandler(ctx *hime.Context) error {

	c, err := repository.AdminListCategory(db)
	must(err)

	p := page(ctx)
	p["Category"] = c
	return ctx.View("admin/topic-create", p)
}

func adminTopicCreatePostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	name := ctx.PostFormValueTrimSpace("name")
	catID := ctx.PostFormValueTrimSpace("cat_id")
	title := ctx.PostFormValueTrimSpace("title")
	description := ctx.PostFormValueTrimSpace("description")
	tagline := ctx.PostFormValueTrimSpace("tagline")

	_, err := repository.CheckTopicName(db, name)
	if err != sql.ErrNoRows {
		must(err)
	}
	if err == nil {
		f.Add("Errors", "ไม่สามารถใช้ชื่อนี้ได้")
		return ctx.RedirectToGet()
	}

	err = repository.CheckCategory(db, catID)
	if err == sql.ErrNoRows {
		f.Add("Errors", "ไอดีนี้ไม่ถูกต้อง")
		return ctx.RedirectToGet()
	}
	must(err)

	fileHeader, err := ctx.FormFileHeaderNotEmpty("image")
	if err != nil {
		f.Add("Errors", "image")
		return ctx.RedirectToGet()
	}

	m, err := service.GetImageFromFile(fileHeader)
	if err != nil {
		f.Add("Errors", "image")
		return ctx.RedirectToGet()
	}

	dst := image.NewRGBA(m.Bounds())
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), m, m.Bounds().Min, draw.Over)

	img := resizeTopicImage(dst)

	fileName := generateTopicName(name)

	err = upload(ctx, img, fileName)
	must(err)

	fileHeader, err = ctx.FormFileHeaderNotEmpty("image_fb")
	if err != nil {
		f.Add("Errors", "image")
		return ctx.RedirectToGet()
	}

	m, err = service.GetImageFromFile(fileHeader)
	if err != nil {
		f.Add("Errors", "image")
		return ctx.RedirectToGet()
	}

	dst = image.NewRGBA(m.Bounds())
	draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), m, m.Bounds().Min, draw.Over)

	imgFb, h, w := resizeImage(dst)
	if h < config.HeightTopicFacebook || w < config.WidthTopicFacebook {
		f.Add("Errors", "ขนาดรูปไม่ต่ำกว่า 1200x630 px")
		return ctx.RedirectToGet()
	}

	fileNameFb := generateTopicName(name + "_fb")

	err = upload(ctx, imgFb, fileNameFb)
	must(err)

	topicID, err := repository.AdminCreateTopic(db, catID, name, generateDownloadURL(fileName), title, description, generateDownloadURL(fileNameFb), tagline)
	must(err)

	repository.AdminAddTopicVerifyToRedis(myRedis, entity.RedisTopicModel{
		ID:        topicID,
		CatID:     catID,
		Code:      name,
		Name:      name,
		Count:     0,
		UsedCount: 0,
		Images: entity.Image{
			Mini:   generateDownloadURL(fileName),
			Normal: generateDownloadURL(fileName),
		},
	})
	must(err)
	return ctx.RedirectTo("admin.topic.seo")
}

func adminTopicEditGetHandler(ctx *hime.Context) error {
	id := ctx.FormValue("id")

	c, err := repository.AdminListCategory(db)
	must(err)

	topic, err := repository.GetTopic(db, id)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("admin.topic")
	}
	must(err)

	p := page(ctx)
	p["Category"] = c
	p["Topic"] = topic

	return ctx.View("admin/topic-edit", p)
}

func adminTopicEditPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	action := ctx.FormValue("action")
	id := ctx.PostFormValueTrimSpace("id")
	name := ctx.PostFormValueTrimSpace("name")
	catID := ctx.PostFormValueTrimSpace("cat_id")

	topic, err := repository.GetTopic(db, id)
	if err == sql.ErrNoRows {
		f.Add("Errors", "id ไม่มีในระบบ")
		return ctx.Redirect("/admin/topic/edit", ctx.Param("id", id))
	}
	must(err)

	if topic.Name != name {
		id, err := repository.CheckTopicName(db, name)
		if err != sql.ErrNoRows {
			must(err)
		}
		if id != "" {
			f.Add("Errors", "ไม่สามารถใช้ชื่อนี้ได้")
			return ctx.Redirect("/admin/topic/edit", ctx.Param("id", id))
		}
	}

	err = repository.CheckCategory(db, catID)
	if err == sql.ErrNoRows {
		f.Add("Errors", "ไอดีนี้ไม่ถูกต้อง")
		return ctx.Redirect("/admin/topic/edit", ctx.Param("id", id))
	}
	must(err)

	urlImage := repository.ImageTopic{
		Mini:   topic.ImageMini,
		Normal: topic.Image,
	}

	fileHeader, err := ctx.FormFileHeader("image")
	if err == nil {

		m, err := service.GetImageFromFile(fileHeader)
		must(err)

		dst := image.NewRGBA(m.Bounds())
		draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(dst, dst.Bounds(), m, m.Bounds().Min, draw.Over)

		img := resizeTopicImage(dst)

		fileName := generateTopicName(name)

		err = upload(ctx, img, fileName)
		must(err)

		urlImage.Mini = generateDownloadURL(fileName)
		urlImage.Normal = generateDownloadURL(fileName)
	}
	if err != nil {
		if action == "verify" {
			f.Add("Errors", "ต้องการรูป")
			return ctx.Redirect("/admin/topic/edit", ctx.Param("id", id))
		}
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		if action == "save" {
			err = repository.AdminUpdateTopic(tx, id, catID, name, urlImage)
			if err != nil {
				return err
			}

			if topic.Verify {
				repository.AdminDeleteTopicVerifyToRedis(myRedis, topic.CatID, topic.ID, topic.Code)
				repository.AdminAddTopicVerifyToRedis(myRedis, entity.RedisTopicModel{
					ID:        topic.ID,
					CatID:     catID,
					Code:      name,
					Name:      name,
					Count:     topic.Count,
					UsedCount: topic.UsedCount,
					Images: entity.Image{
						Mini:   urlImage.Mini,
						Normal: urlImage.Normal,
					},
				})
			}

			if !topic.Verify {
				repository.AdminDeleteTopicNotVerifyToRedis(myRedis, topic.CatID, topic.ID, topic.Code)
				repository.AdminAddTopicNotVerifyToRedis(myRedis, entity.RedisTopicModel{
					ID:        topic.ID,
					CatID:     catID,
					Code:      name,
					Name:      name,
					Count:     topic.Count,
					UsedCount: topic.UsedCount,
					Images: entity.Image{
						Mini:   urlImage.Mini,
						Normal: urlImage.Normal,
					},
				})
			}

		}

		if action == "verify" {
			err = repository.AdminUpdateTopicVerify(tx, id, catID, name, urlImage)
			if err != nil {
				return err
			}

			if topic.Verify {
				repository.AdminDeleteTopicVerifyToRedis(myRedis, topic.CatID, topic.ID, topic.Code)
				repository.AdminAddTopicVerifyToRedis(myRedis, entity.RedisTopicModel{
					ID:        topic.ID,
					CatID:     catID,
					Code:      name,
					Name:      name,
					Count:     topic.Count,
					UsedCount: topic.UsedCount,
					Images: entity.Image{
						Mini:   urlImage.Mini,
						Normal: urlImage.Normal,
					},
				})
			}

			if !topic.Verify {
				repository.AdminDeleteTopicNotVerifyToRedis(myRedis, topic.CatID, topic.ID, topic.Code)
				repository.AdminAddTopicVerifyToRedis(myRedis, entity.RedisTopicModel{
					ID:        topic.ID,
					CatID:     catID,
					Code:      name,
					Name:      name,
					Count:     topic.Count,
					UsedCount: topic.UsedCount,
					Images: entity.Image{
						Mini:   urlImage.Mini,
						Normal: urlImage.Normal,
					},
				})

			}
		}

		return nil
	})
	must(err)

	return ctx.RedirectTo("admin.topic")
}

func adminTopicSEOGetHandler(ctx *hime.Context) error {
	topic, err := repository.ListTopicVerified(db, GetUserRole(ctx))
	must(err)

	p := page(ctx)
	p["Topic"] = topic
	return ctx.View("admin/topic-seo", p)
}

func adminTopicSEOEditGetHandler(ctx *hime.Context) error {
	code := ctx.FormValue("code")

	topic, err := repository.GetTopicByCode(db, code)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	p := page(ctx)
	p["Topic"] = topic

	return ctx.View("admin/topic-seo-edit", p)
}

func adminTopicSEOEditPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	code := ctx.PostFormValueTrimSpace("code")
	title := ctx.PostFormValueTrimSpace("title")
	description := ctx.PostFormValueTrimSpace("description")
	tagline := ctx.PostFormValueTrimSpace("tagline")

	topic, err := repository.GetTopicByCode(db, code)
	if err == sql.ErrNoRows {
		f.Add("Errors", "id ไม่มีในระบบ")
		return ctx.Redirect("/admin/topic/seo/edit", ctx.Param("code", code))
	}
	must(err)

	fileHeader, err := ctx.FormFileHeader("image_fb")
	if err == nil {

		m, err := service.GetImageFromFile(fileHeader)
		if err != nil {
			f.Add("Errors", "image")
			return ctx.RedirectToGet()
		}

		dst := image.NewRGBA(m.Bounds())
		draw.Draw(dst, dst.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(dst, dst.Bounds(), m, m.Bounds().Min, draw.Over)

		imgFb, h, w := resizeImage(dst)
		if h < config.HeightTopicFacebook || w < config.WidthTopicFacebook {
			f.Add("Errors", "ขนาดรูปไม่ต่ำกว่า 1200x630 px")
			return ctx.RedirectToGet()
		}

		fileNameFb := generateTopicName(topic.Name + "_fb")

		err = upload(ctx, imgFb, fileNameFb)
		must(err)

		topic.ImageFacebook = generateDownloadURL(fileNameFb)
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {
		err := repository.AdminUpdateTopicSEOVerify(tx, topic.ID, title, description, tagline, topic.ImageFacebook)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)
	f.Add("Success", "แก้ไข SEO เรียบร้อยแล้ว")
	return ctx.Redirect("/admin/topic/seo/edit", ctx.Param("code", code))
}

func adminSelectUserGetHandler(ctx *hime.Context) error {

	return ctx.View("app/selectuser", page(ctx))
}

func adminAddCoinGetHandler(ctx *hime.Context) error {
	email := ctx.FormValue("email")
	if email == "" {
		return ctx.RedirectTo("admin.selectuser")
	}

	user := repository.GetUserByEmail(db, email)
	if user.ID == "" {
		return ctx.RedirectTo("admin.selectuser")
	}

	p := page(ctx)
	p["User"] = user

	return ctx.View("app/addcoin", p)
}

func adminAddCoinPostHandler(ctx *hime.Context) error {
	userID := ctx.PostFormValue("userID")
	wallet := ctx.PostFormValueInt64("wallet")
	bonus := ctx.PostFormValueInt64("bonus")

	f := getSession(ctx).Flash()
	f.Clear()

	if wallet == 0 {
		f.Add("Errors", "กรุณากรอกจำนวนเงินให้ถูกต้อง")
		return ctx.RedirectToGet()
	}

	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.AddWalletAndBonusUser(tx, userID, wallet, bonus)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	f.Add("Success", "เติมเงินเข้าระบบเรียบร้อยแล้ว")
	return ctx.RedirectToGet()
}
