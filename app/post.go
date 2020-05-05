package app

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"image"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/acoshift/pgsql"
	"github.com/anthoz69/go-slugify"
	"github.com/google/uuid"
	"github.com/moonrhythm/hime"
	"github.com/mssola/user_agent"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
	"github.com/workdestiny/amlporn/service"
)

func createPostGetHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	role := GetUserRole(ctx)
	if role == entity.RoleUser {
		return ctx.RedirectTo("notfound")
	}

	listGap, err := repository.ListGap(db, id, config.LimitListGap)
	if len(listGap) == 0 {
		return ctx.RedirectTo("create.gap")
	}

	draftPost, draftErr := repository.GetDraftPost(db, id)

	listCategory, err := repository.ListCategoryAll(db)
	must(err)

	p := page(ctx)
	p["ListGap"] = listGap
	p["MainGap"] = listGap[0]

	if draftErr == nil {
		p["Draft"] = draftPost
		p["MainGap"] = draftPost.Owner
	}

	p["ListCategory"] = listCategory
	p["ProvinceData"] = entity.ProvinceData

	return ctx.View("app/create-post", p)
}

func createPostPostHandler(ctx *hime.Context) error {

	f := getSession(ctx).Flash()
	me := getUser(ctx)

	var req repository.CreatePostModel

	if ctx.PostFormValueInt("type") > 2 {
		f.Add("ErrType", "Type Post ไม่ถูกต้อง")
	}

	req.OwnerID = ctx.PostFormValue("owner-id")
	onair := ctx.PostFormValue("onair")
	req.Title = ctx.PostFormValueTrimSpace("title")
	req.Description = ctx.PostFormValueTrimSpace("description")
	req.LinkDescription = service.StripHTML(ctx.PostFormValueTrimSpace("link-description"))
	req.Link = service.StripHTML(ctx.PostFormValueTrimSpace("link"))
	req.VdoURL = service.StripHTML(ctx.PostFormValueTrimSpace("vdourl"))
	req.Type = entity.TypePost(ctx.PostFormValueInt("type"))
	req.Province = ctx.PostFormValueInt("province")

	json.Unmarshal([]byte(ctx.PostFormValue("tag-topic")), &req.TagTopics)

	createdAt := time.Now().UTC()
	if me.IsAdmin() {
		if onair != "" {
			split := strings.Split(onair, "/")
			if len(split) != 3 {
				f.Add("ErrTime", "เวลาตั้งโพสไม่ถูกต้อง")
				return ctx.RedirectToGet()
			}

			day, err := strconv.Atoi(split[0])
			if err != nil || day > 31 || day == 0 {
				f.Add("ErrTime", "เวลาตั้งโพสไม่ถูกต้อง")
				return ctx.RedirectToGet()
			}

			month, err := strconv.Atoi(split[1])
			if err != nil || month > 12 || month == 0 {
				f.Add("ErrTime", "เวลาตั้งโพสไม่ถูกต้อง")
				return ctx.RedirectToGet()
			}

			year, err := strconv.Atoi(split[2])
			if err != nil || year < 2019 {
				f.Add("ErrTime", "เวลาตั้งโพสไม่ถูกต้อง")
				return ctx.RedirectToGet()
			}

			t := time.Date(year, time.Month(month), day, 0, 1, rand.Intn(59), rand.Intn(99999), time.UTC)
			diff := t.Sub(createdAt)
			if diff > 0 {
				createdAt = t
			}
		}
	}

	_, isOwner := repository.CheckOwnerGap(myRedis, req.OwnerID, me.ID)
	if !isOwner {
		f.Add("ErrOwner", "ไม่ใช่เจ้าของ Gap")
		return ctx.RedirectToGet()
	}

	if req.Province == 0 {
		req.Province = 1
	}

	if req.Province != 01 {

		if !checkProvinceID(req.Province) {
			f.Add("ErrProvince", "จังหวัดไม่ถูกต้อง")
		}
	}

	if service.StripHTML(req.Title) == "" && service.StripHTML(req.Description) == "" {
		f.Add("Errors", "Error401")
	}

	req.Title = service.StripHTML(req.Title)
	req.Description = service.SanitizeUGCDescription(req.Description)

	if f.Has("ErrType") || f.Has("ErrProvince") || f.Has("Errors") {
		return ctx.RedirectToGet()
	}

	//check tagtopic and set
	var tagID []string

	if len(req.TagTopics) > 0 {
		for i := 0; i < len(req.TagTopics); i++ {

			if i == 5 {
				break
			}

			if req.TagTopics[i].TopicID != "0" {

				if req.TagTopics[i].TopicID == config.TopicOfficialID {

					if me.Role == entity.RoleAdmin {
						if checkTagtopic(tagID, req.TagTopics[i].TopicID) {
							tagID = append(tagID, req.TagTopics[i].TopicID)
						}
						continue
					}

					continue
				}

				err := repository.CheckTopicID(db, req.TagTopics[i].TopicID)
				if err != nil {
					continue
				}

				if checkTagtopic(tagID, req.TagTopics[i].TopicID) {
					tagID = append(tagID, req.TagTopics[i].TopicID)
				}
				continue
			}

			if req.TagTopics[i].Name == "" {
				continue
			}

			if utf8.RuneCountInString(req.TagTopics[i].Name) > 25 {
				req.TagTopics[i].Name = req.TagTopics[i].Name[:25]
			}

			err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				topicID, err := repository.CheckTopicName(tx, req.TagTopics[i].Name)
				if err == sql.ErrNoRows {
					//create
					topicID, err := repository.CreateTopic(tx, req.TagTopics[i].Name)
					if err != nil {
						return err
					}

					if checkTagtopic(tagID, topicID) {
						tagID = append(tagID, topicID)
					}
					return nil
				}

				if checkTagtopic(tagID, topicID) {
					tagID = append(tagID, topicID)
				}
				return nil
			})
			must(err)
		}
	}

	if len(tagID) == 0 {
		tagID = append(tagID, config.TopicOtherID)
	}

	//create slug
	code := strings.ToLower(repository.RandStr(6))
	slug := slugify.Marshal(req.Title, true)

	if utf8.RuneCountInString(req.Title) > 50 {
		slug = slugify.Marshal(string([]rune(req.Title)[:50]), true)
	}

	if slug != "" {
		slug = slug + "-" + code
	}

	if slug == "" {
		slug = uuid.New().String()
	}

	// imgs := findImages(req.Description)
	// for _, img := range imgs {
	// 	repository.ConfirmImage(db, me.ID, img)
	// }

	// if len(imgs) > 0 {
	// 	req.ImageShareURL = imgs[0]
	// 	req.HeightShare = config.HigthImageFacebook
	// 	req.WidthShare = config.WidthImageFacebook
	// 	req.ImageURL = imgs[0]
	// 	req.ImageURLMobile = imgs[0]
	// 	req.Width = 1000
	// 	req.Height = 1000
	// }

	//resize and upload
	imgs := findImages(req.Description)
	if len(imgs) > 0 {

		for _, img := range imgs {
			repository.ConfirmImage(db, me.ID, img)
		}

		resp, err := http.Get(imgs[0])
		must(err)
		defer resp.Body.Close()

		image, _, err := image.Decode(resp.Body)
		if err != nil {
			return err
		}

		req.ImageShareURL = imgs[0]
		req.HeightShare = config.HigthImageFacebook
		req.WidthShare = config.WidthImageFacebook

		m, w, h := resizeMainImage(image)
		name := generateMainImagePostName(me.ID)
		upload(ctx, m, name)

		mb, w, h := resizeMainImageMobile(image)
		nameMb := generateMainImagePostNameMobile(me.ID)
		uploadThumbnailMobile(ctx, mb, nameMb)

		req.ImageURL = generateDownloadURL(name)
		req.ImageURLMobile = generateDownloadURL(nameMb)
		req.Width = w
		req.Height = h
	}
	if len(imgs) == 0 {
		if strings.Contains(req.Description, `<div class="video-wrap"><iframe src="`) {
			v := strings.Split(req.Description, `<div class="video-wrap"><iframe src="`)
			if len(v) > 1 {
				link := strings.Split(v[1], `"`)
				if len(link) > 0 {

					if service.CheckRejectLinkPost(link[0]) {

						url := service.GetImageIDFromIFrame(link[0])
						if url != "" {
							resp, err := http.Get(url)
							must(err)
							defer resp.Body.Close()

							image, _, err := image.Decode(resp.Body)

							req.HeightShare = image.Bounds().Dy()
							req.WidthShare = image.Bounds().Dx()

							m, w, h := resizeMainImage(image)
							name := generateMainImagePostName(me.ID)
							upload(ctx, m, name)

							req.ImageURL = generateDownloadURL(name)
							req.Width = w
							req.Height = h

							req.ImageShareURL = req.ImageURL
						}
					}
				}
			}
		}
	}

	req.StatusVerify = me.GetLevelPost()

	postID, err := repository.CheckDraftPost(db, me.ID)
	if err == sql.ErrNoRows {
		//create
		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			loop := true
			for loop {
				err := repository.CheckPostSlug(tx, slug)
				if err == sql.ErrNoRows {
					loop = false
					break
				}
				if err != nil && err != sql.ErrNoRows {
					return err
				}

				slug = string([]rune(slug)[:utf8.RuneCountInString(slug)-7]) + "-" + strings.ToLower(repository.RandStr(6))
			}

			returnPostID, err := repository.CreatePost(tx, me.ID, repository.CreatePostModel{
				OwnerID:         req.OwnerID,
				Onair:           createdAt,
				Title:           req.Title,
				Description:     req.Description,
				Link:            req.Link,
				VdoURL:          req.VdoURL,
				ImageURL:        req.ImageURL,
				ImageURLMobile:  req.ImageURLMobile,
				Type:            req.Type,
				Province:        req.Province,
				Height:          req.Height,
				Width:           req.Width,
				StatusVerify:    req.StatusVerify,
				LinkDescription: req.LinkDescription,
				Slug:            slug,
				ImageShareURL:   req.ImageShareURL,
				HeightShare:     req.HeightShare,
				WidthShare:      req.WidthShare,
			})
			if err != nil {
				return NewAppError("ไม่สามารถสร้าง Post ได้")
			}

			err = repository.CreateTagTopic(tx, returnPostID, tagID)
			if err != nil {
				return NewAppError("ไม่สามารถสร้าง Post ได้")
			}

			return nil
		})
		if IsAppError(err) {
			f.Add("Errors", err.Error())
			return ctx.RedirectToGet()
		}
		must(err)

		go updateCountCreatePost(tagID, req.OwnerID)

		return ctx.Redirect("/gap/" + req.OwnerID)
	}
	must(err)

	//update
	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		loop := true
		for loop {
			err := repository.CheckPostSlug(tx, slug)
			if err == sql.ErrNoRows {
				loop = false
				break
			}
			if err != nil && err != sql.ErrNoRows {
				return err
			}

			slug = string([]rune(slug)[:utf8.RuneCountInString(slug)-7]) + "-" + strings.ToLower(repository.RandStr(6))
		}

		err = repository.UpdateCreatePost(tx, postID, repository.CreatePostModel{
			OwnerID:         req.OwnerID,
			Onair:           createdAt,
			Title:           req.Title,
			Description:     req.Description,
			Link:            req.Link,
			VdoURL:          req.VdoURL,
			ImageURL:        req.ImageURL,
			ImageURLMobile:  req.ImageURLMobile,
			Type:            req.Type,
			Province:        req.Province,
			Height:          req.Height,
			Width:           req.Width,
			StatusVerify:    req.StatusVerify,
			LinkDescription: req.LinkDescription,
			Slug:            slug,
			ImageShareURL:   req.ImageShareURL,
			HeightShare:     req.HeightShare,
			WidthShare:      req.WidthShare,
		})
		if err != nil {
			log.Println(err)
			return NewAppError("ไม่สามารถสร้าง Post ได้")
		}

		err = repository.CreateTagTopic(tx, postID, tagID)
		if err != nil {
			return NewAppError("ไม่สามารถสร้าง Post ได้")
		}

		return nil
	})
	if IsAppError(err) {
		f.Add("Errors", err.Error())
		return ctx.RedirectToGet()
	}
	must(err)

	//update count
	go updateCountCreatePost(tagID, req.OwnerID)

	return ctx.Redirect("/gap/" + req.OwnerID)
}

func deletePostPostHandler(ctx *hime.Context) error {

	f := getSession(ctx).Flash()
	me := getUser(ctx)

	postID := ctx.PostFormValue("id")

	gapID, err := repository.FindGapOwnerPost(db, postID)
	if err != nil {
		return ctx.RedirectTo("notfound")
	}

	if !me.IsAdmin() {
		err = repository.CheckOwnerPost(db, me.ID, postID)
		if err != nil {
			f.Add("ErrOwner", "ไม่ใช่เจ้าของ Post")
			return ctx.RedirectToGet()
		}
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.DeletePost(tx, postID)
		if err != nil {
			return err
		}

		err = repository.DeleteTagTopic(db, tx, myRedis, postID)
		if err != nil {
			return err
		}

		err = repository.MinusCountGapPost(tx, gapID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		f.Add("Errors", "ไม่สามารถลบ Post นี้ได้")
		return ctx.RedirectToGet()
	}

	return ctx.RedirectTo("gap", gapID)
}

func updateCountCreatePost(ids []string, gapID string) {

	pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.UpdateCountGapPost(tx, gapID)
		if err != nil {
			log.Println("err update count gap!")
		}

		return nil
	})

	for _, v := range ids {

		pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err := repository.UpdateUsedCountTopic(tx, v)
			if err != nil {
				log.Println("err update count topic!")
			}

			return nil
		})

		t, err := repository.GetTopic(db, v)
		if err == nil {

			var topicRedis entity.RedisTopicModel
			topicRedis.ID = t.ID
			topicRedis.CatID = t.CatID
			topicRedis.Code = t.Code
			topicRedis.Name = t.Name
			topicRedis.Count = t.Count
			topicRedis.UsedCount = t.UsedCount
			topicRedis.Images = entity.Image{
				Mini:   t.ImageMini,
				Normal: t.Image,
			}

			buf := bytes.Buffer{}
			gob.NewEncoder(&buf).Encode(&topicRedis)

			if t.UsedCount == 1 && !t.Verify {
				repository.AddTopicNotVerifyToRedis(myRedis, t.CatID, t.ID, t.Code, t.Name, buf.Bytes())
				continue
			}

			if t.Verify {

				repository.SetTopicVerifyToRedis(myRedis, t.CatID, t.ID, t.Code, buf.Bytes())
				continue
			}

			if !t.Verify {

				repository.SetTopicNotVerifyToRedis(myRedis, t.CatID, t.ID, t.Code, buf.Bytes())
				continue
			}

		}
	}
}

func updateCountEditPost(ids []string) {

	for _, v := range ids {

		pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err := repository.UpdateUsedCountTopic(tx, v)
			if err != nil {
				log.Println("err update count topic!")
			}

			return nil
		})

		t, err := repository.GetTopic(db, v)
		if err == nil {

			var topicRedis entity.RedisTopicModel
			topicRedis.ID = t.ID
			topicRedis.CatID = t.CatID
			topicRedis.Code = t.Code
			topicRedis.Name = t.Name
			topicRedis.Count = t.Count
			topicRedis.UsedCount = t.UsedCount
			topicRedis.Images = entity.Image{
				Mini:   t.ImageMini,
				Normal: t.Image,
			}

			buf := bytes.Buffer{}
			gob.NewEncoder(&buf).Encode(&topicRedis)

			if t.UsedCount == 1 && !t.Verify {

				repository.AddTopicNotVerifyToRedis(myRedis, t.CatID, t.ID, t.Code, t.Name, buf.Bytes())
				continue
			}

			if t.Verify {

				repository.SetTopicVerifyToRedis(myRedis, t.CatID, t.ID, t.Code, buf.Bytes())
				continue
			}

			if !t.Verify {

				repository.SetTopicNotVerifyToRedis(myRedis, t.CatID, t.ID, t.Code, buf.Bytes())
				continue
			}

		}
	}
}

func ajaxLikePostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	postOwnerID, err := repository.CheckPostID(db, req.ID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResLike{
		IsLike: false,
	}

	like, err := repository.CheckLike(db, id, req.ID)
	if err == sql.ErrNoRows {

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err := repository.CreateLike(tx, id, req.ID)
			if err != nil {
				return NewAppError("ไม่สามารถถูกใจได้")
			}

			return nil
		})
		if IsAppError(err) {
			return ctx.JSON(map[string]interface{}{
				"Error": err.Error(),
			})
		}
		must(err)

		go repository.AddNotificationLike(db, myRedis, id, req.ID, postOwnerID)

		res.IsLike = true
		return ctx.Status(http.StatusOK).JSON(&res)

	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.LikePost(tx, id, req.ID, !like)
		if err != nil {
			return NewAppError("ไม่สามารถติดตามได้")
		}

		return nil
	})
	if IsAppError(err) {
		return ctx.JSON(map[string]interface{}{
			"Error": err.Error(),
		})
	}
	must(err)

	res.IsLike = !like

	if res.IsLike {
		go repository.AddNotificationLike(db, myRedis, id, req.ID, postOwnerID)
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func postReadGetHandler(ctx *hime.Context) error {
	postID := getParams(ctx, "postID")
	id := GetMyID(ctx)
	//mo := ua.Parse(ctx.UserAgent())

	if postID == "" {
		return ctx.RedirectTo("notfound")
	}

	// Get Post
	post, err := repository.GetPost(db, myRedis, id, postID)
	if post.ID == "" {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	p := page(ctx)
	if post.Slug != postID {
		p["Canonical"] = ""
		//p["Canonical"] = "https://www.xs.com/post/" + post.Slug
	}

	//Check User is Owner Gap
	_, isOwner := repository.CheckOwnerGap(myRedis, post.Owner.ID, GetMyID(ctx))
	post.Owner.IsOwner = isOwner

	// Get TopPost
	listTopPost, err := repository.ListTopPostGap(db, post.Owner.ID, post.ID, config.LimitListTopPostGap)
	must(err)

	// Get RelateTagTopic
	listPostRelateTagTopic, err := repository.ListPostRelateTagTopic(db, post.ID, entity.GetProvinceID(post.Province), config.LimitListPostRelateTagTopic)
	must(err)

	var next time.Time
	isNext := false
	if post.Count.Comment != "0" {
		//Get ListComment
		listComment, err := repository.ListComment(db, myRedis, post.ID, 2+1)
		must(err)

		if len(listComment) > 2 {
			isNext = true
			next = listComment[2-1].CreatedAt
			listComment = listComment[:2]
		}

		p["ListComment"] = listComment
	}

	p["IsGapOfficial"] = false
	if post.Owner.ID == config.GapOfficialID {
		p["IsGapOfficial"] = true
	}
	p["NextLoad"] = next
	p["IsNext"] = isNext
	p["Post"] = post
	p["TopPost"] = listTopPost
	p["RelateTagTopic"] = listPostRelateTagTopic
	p["Tagline"] = ""
	p["GapAuthor"] = post.Owner.Name
	p["ParamID"] = post.ID
	//p["IsMobile"] = mo.Mobile

	var (
		title       string
		description string
	)
	if post.Title != "" {
		title = post.Title
	} else {
		title = post.Owner.Name
	}

	if post.Description != "" {
		description = post.Description
	} else {
		description = post.Title
	}

	p["M"] = setMeta(
		service.ShortTextMeta(70, title),
		service.ShortTextMeta(170, description),
		post.ImageShareURL,
		post.WidthShare,
		post.HeightShare,
	)

	if id != "" {
		// add view
		go countView(post.ID, id, ctx.Request.UserAgent(), ctx.Request.Referer())
	}

	if id == "" {
		userAgent := ctx.Request.UserAgent()
		ua := user_agent.New(userAgent)
		if !ua.Bot() {
			go countGuestView(post.ID, GetVisitorID(ctx), userAgent, ctx.Request.Referer())
		}
	}

	return ctx.View("app/post-read", p)
}

func ajaxCommentNextLoadPostHandler(ctx *hime.Context) error {

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	list, err := repository.ListCommentNextLoad(db, myRedis, req.ID, req.Next, config.LimitListComment+1)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	if len(list) == 0 {
		return ctx.NoContent()
	}

	var next time.Time
	isNext := false

	if len(list) > config.LimitListComment {
		isNext = true
		next = list[config.LimitListComment-1].CreatedAt
		list = list[:config.LimitListComment]
	}

	res := entity.ResponseCommentList{
		Comment: list,
		Next:    next,
		IsNext:  isNext,
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxDraftImagePostHandler(ctx *hime.Context) error {

	res := entity.ResDraftImage{
		Uploaded: false,
	}

	message := &entity.ResErrorDraftImage{}

	id := GetMyID(ctx)
	if id == "" {
		message.Message = "Error StatusBadRequest"
		res.Error = message
		return ctx.Status(http.StatusBadRequest).JSON(&res)
	}

	fileHeader, err := ctx.FormFileHeaderNotEmpty("upload")
	if err != nil {
		message.Message = "Error StatusBadRequest"
		res.Error = message
		return ctx.Status(http.StatusBadRequest).JSON(&res)
	}

	typ := fileHeader.Header.Get("Content-Type")
	if !strings.Contains(typ, "image") {
		message.Message = "Error StatusBadRequest"
		res.Error = message
		return ctx.Status(http.StatusBadRequest).JSON(&res)
	}

	filename := service.GeneratePostImageName(id)
	height, width, err := service.Upload(ctx, bucket.Storage, fileHeader, filename)
	if err != nil {
		message.Message = "Error StatusInternalServerError"
		res.Error = message
		return ctx.Status(http.StatusInternalServerError).JSON(&res)
	}

	imageURL := generateDownloadURL(filename)
	err = repository.CreateImage(db, id, imageURL, height, width)
	must(err)

	res.Uploaded = true
	res.URL = imageURL
	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxCommentPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestComment
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	req.Text = strings.TrimSpace(req.Text)

	req.Text = service.SanitizeUGC(req.Text)
	if req.Text == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	gapID, err := repository.CheckPostID(db, req.PostID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	u := getUser(ctx)

	res := entity.ResponseComment{
		Owner: entity.ResponseCommentOwner{
			ID:      id,
			Name:    u.FirstName + " " + u.LastName,
			Display: u.DisplayImage.Normal,
			Type:    req.Type,
		},
		Text: service.UnescapeString(req.Text),
		Time: service.FormatTime(time.Now()),
	}

	ownerID := id
	if req.Type == entity.TypeOwnerCommentGap {

		g, isOwner := repository.CheckOwnerGap(myRedis, req.GapID, id)
		if !isOwner {
			return ctx.Status(http.StatusBadRequest).JSON(nil)
		}

		res.Owner.ID = req.GapID
		res.Owner.Name = g.Name
		res.Owner.Display = g.DisplayImage
		ownerID = req.GapID
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		returnID, err := repository.CommentPost(tx, req.PostID, ownerID, req.Type, req.Text)
		if err != nil {
			return NewAppError("ไม่สามารถคอมเม้นต์ได้")
		}

		res.ID = returnID

		return nil
	})
	if IsAppError(err) {
		return ctx.JSON(map[string]interface{}{
			"Error": err.Error(),
		})
	}
	must(err)

	go repository.AddNotificationComment(db, myRedis, id, req.PostID, gapID, req.Text)

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxCommentEditPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestEditComment
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	req.Text = strings.TrimSpace(req.Text)

	req.Text = service.SanitizeUGC(req.Text)

	if req.Text == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.CheckOwnerComment(db, id, req.CommentID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.EditComment(tx, req.CommentID, req.Text)

		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	return ctx.Status(http.StatusOK).JSON(nil)
}

func ajaxCommentDeletePost(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestDeleteComment
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.CheckOwnerPost(db, id, req.PostID)
	if err == sql.ErrNoRows {

		err = repository.CheckOwnerComment(db, id, req.CommentID)
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(nil)
		}
	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.DeleteComment(tx, req.PostID, req.CommentID)

		if err != nil {
			log.Println(err)
			return err
		}

		return nil
	})
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	res := entity.ResponseDeleteComment{
		CommentID: req.CommentID,
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}

func ajaxSearchTagTopicPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	role := GetUserRole(ctx)

	var req entity.TagTopicRequest
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	req.Text = strings.ToLower(req.Text)

	t, err := repository.SearchTagTopic(myRedis, req.Text, role, config.LimitSearchTagTopic)
	must(err)

	if len(t) == 0 {
		return ctx.NoContent()
	}

	return ctx.Status(http.StatusOK).JSON(&t)
}

//CountView update view count
func countView(postID string, userID string, userAgent string, referrer string) {

	t, err := repository.GetViewPost(db, postID, userID)
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

		err = repository.CreateViewPost(tx, postID, userID, userAgent, referrer)
		if err != nil {
			return err
		}

		err = repository.UpdateCountViewPost(tx, postID)
		if err != nil {
			return err
		}

		return nil
	})
}

//countGuestView update guestView count
func countGuestView(postID string, vsID string, userAgent string, referrer string) {

	t, err := repository.GetGuestViewPost(db, postID, vsID)
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

		err = repository.CreateGuestViewPost(tx, postID, vsID, userAgent, referrer)
		if err != nil {
			return err
		}

		err = repository.UpdateCountGuestViewPost(tx, postID)
		if err != nil {
			return err
		}

		return nil
	})
}

func ajaxDraftPostArticleHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	if me.ID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestPost
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	isVerifyGap := 0
	if me.GetLevel() == entity.VerifyEmail {
		isVerifyGap = 1
	}
	if me.GetLevel() > entity.VerifyEmail {
		isVerifyGap = 2
	}

	_, isOwner := repository.CheckOwnerGap(myRedis, req.OwnerID, me.ID)
	if !isOwner {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	if service.StripHTML(req.Title) == "" && service.SanitizeUGCDescription(req.Description) == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}
	req.Title = service.StripHTML(strings.TrimSpace(req.Title))
	req.Description = service.SanitizeUGCDescription(req.Description)

	postID, err := repository.CheckDraftPost(db, me.ID)
	if err == sql.ErrNoRows {

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.CreateDraftPost(tx, me.ID, repository.CreatePostModel{
				OwnerID:      req.OwnerID,
				Title:        req.Title,
				Description:  req.Description,
				Type:         0,
				Province:     01,
				StatusVerify: isVerifyGap,
			})

			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(nil)
		}

		return ctx.NoContent()
	}
	must(err)

	err = repository.UpdateDraftPost(db, postID, repository.CreatePostModel{
		OwnerID:      req.OwnerID,
		Title:        req.Title,
		Description:  req.Description,
		Type:         0,
		Province:     01,
		StatusVerify: isVerifyGap,
	})
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	return ctx.NoContent()
}

func undraftPostHandler(ctx *hime.Context) error {

	id := GetMyID(ctx)
	if id == "" {
		return ctx.RedirectTo("signin")
	}

	postID, err := repository.CheckDraftPost(db, id)
	if err != nil {
		return ctx.RedirectTo("create.post")
	}

	err = repository.DeleteDraftPost(db, id, postID)
	must(err)

	return ctx.RedirectTo("create.post")
}

func editPostGetHandler(ctx *hime.Context) error {
	postID := getParams(ctx, "postID")
	me := getUser(ctx)

	if postID == "" {
		return ctx.RedirectTo("notfound")
	}

	// Get Post
	post, err := repository.GetPost(db, myRedis, me.ID, postID)
	if post.ID == "" {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	//Check User is Owner Gap
	if !me.IsAdmin() {
		_, isOwner := repository.CheckOwnerGap(myRedis, post.Owner.ID, GetMyID(ctx))
		if !isOwner {
			return ctx.Redirect("/post/" + postID)
		}
	}

	listGap, err := repository.ListGap(db, me.ID, config.LimitListGap)
	if len(listGap) == 0 {
		return ctx.RedirectTo("create.gap")
	}

	listCategory, err := repository.ListCategoryAll(db)
	must(err)

	p := page(ctx)
	p["Post"] = post
	p["ListGap"] = listGap
	p["MainGap"] = entity.DraftOwner{
		ID:      post.Owner.ID,
		Name:    post.Owner.Name,
		Display: post.Owner.Display,
	}
	p["ListCategory"] = listCategory
	p["ProvincePost"] = entity.GetProvince(post.Province)
	p["ProvinceData"] = entity.ProvinceData

	return ctx.View("app/post-edit", p)
}

func editPostPostHandler(ctx *hime.Context) error {

	f := getSession(ctx).Flash()
	me := getUser(ctx)

	var req repository.EditPostModel

	if ctx.PostFormValueInt("type") > 2 {
		f.Add("ErrType", "Type Post ไม่ถูกต้อง")
	}

	req.ID = getParams(ctx, "postID")
	req.Title = ctx.PostFormValueTrimSpace("title")
	req.Description = ctx.PostFormValueTrimSpace("description")
	req.LinkDescription = service.SanitizeUGC(ctx.PostFormValue("link-description"))
	req.Link = service.SanitizeUGC(ctx.PostFormValue("link"))
	req.Type = entity.TypePost(ctx.PostFormValueInt("type"))
	req.Province = ctx.PostFormValueInt("province")
	req.VdoURL = service.SanitizeUGC(ctx.PostFormValue("vdourl"))

	json.Unmarshal([]byte(ctx.PostFormValue("tag-topic")), &req.TagTopics)

	if !me.IsAdmin() {
		err := repository.CheckOwnerPost(db, me.ID, req.ID)
		if err != nil {
			f.Add("ErrOwner", "ไม่ใช่เจ้าของ Post")
			return ctx.Redirect("/edit/post/" + req.ID)
		}
	}

	if req.Province == 0 {
		req.Province = 1
	}

	if req.Province != 01 {

		if !checkProvinceID(req.Province) {
			f.Add("ErrProvince", "จังหวัดไม่ถูกต้อง")
		}
	}

	if service.StripHTML(req.Title) == "" && service.StripHTML(req.Description) == "" {
		f.Add("Errors", "Error401")
	}

	req.Title = service.StripHTML(req.Title)
	req.Description = service.SanitizeUGCDescription(req.Description)

	if f.Has("ErrType") || f.Has("ErrProvince") || f.Has("Errors") {
		return ctx.Redirect("/edit/post/" + req.ID)
	}

	//check tagtopic and set
	var tagID []string

	if len(req.TagTopics) > 0 {
		for i := 0; i < len(req.TagTopics); i++ {

			if i == 5 {
				break
			}

			if req.TagTopics[i].TopicID != "0" {

				if req.TagTopics[i].TopicID == config.TopicOfficialID {

					if me.Role == entity.RoleAdmin {
						if checkTagtopic(tagID, req.TagTopics[i].TopicID) {
							tagID = append(tagID, req.TagTopics[i].TopicID)
						}
						continue
					}

					continue
				}

				err := repository.CheckTopicID(db, req.TagTopics[i].TopicID)
				if err != nil {
					continue
				}

				if checkTagtopic(tagID, req.TagTopics[i].TopicID) {
					tagID = append(tagID, req.TagTopics[i].TopicID)
				}
				continue
			}

			if req.TagTopics[i].Name == "" {
				continue
			}

			if utf8.RuneCountInString(req.TagTopics[i].Name) > 25 {
				req.TagTopics[i].Name = req.TagTopics[i].Name[:25]
			}

			err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				topicID, err := repository.CheckTopicName(tx, req.TagTopics[i].Name)
				if err == sql.ErrNoRows {
					//create
					topicID, err := repository.CreateTopic(tx, req.TagTopics[i].Name)
					if err != nil {
						return err
					}

					if checkTagtopic(tagID, topicID) {
						tagID = append(tagID, topicID)
					}
					return nil
				}

				if checkTagtopic(tagID, topicID) {
					tagID = append(tagID, topicID)
				}
				return nil
			})
			must(err)
		}
	}

	if len(tagID) == 0 {
		tagID = append(tagID, config.TopicOtherID)
	}

	// imgs := findImages(req.Description)
	// for _, img := range imgs {
	// 	repository.ConfirmImage(db, me.ID, img)
	// }
	// if len(imgs) > 0 {
	// 	req.ImageShareURL = imgs[0]
	// 	req.HeightShare = config.HigthImageFacebook
	// 	req.WidthShare = config.WidthImageFacebook
	// 	req.ImageURL = imgs[0]
	// 	req.ImageURLMobile = imgs[0]
	// 	req.Width = 1000
	// 	req.Height = 1000
	// }

	// resize and upload
	imgs := findImages(req.Description)
	if len(imgs) > 0 {

		for _, img := range imgs {
			repository.ConfirmImage(db, me.ID, img)
		}

		resp, err := http.Get(imgs[0])
		must(err)
		defer resp.Body.Close()

		image, _, err := image.Decode(resp.Body)
		if err != nil {
			return err
		}

		req.ImageShareURL = imgs[0]
		req.HeightShare = config.HigthImageFacebook
		req.WidthShare = config.WidthImageFacebook

		m, w, h := resizeMainImage(image)
		name := generateMainImagePostName(me.ID)
		upload(ctx, m, name)

		mb, w, h := resizeMainImageMobile(image)
		nameMb := generateMainImagePostNameMobile(me.ID)
		uploadThumbnailMobile(ctx, mb, nameMb)

		req.ImageURL = generateDownloadURL(name)
		req.ImageURLMobile = generateDownloadURL(nameMb)
		req.Width = w
		req.Height = h
	}
	if len(imgs) == 0 {
		if strings.Contains(req.Description, `<div class="video-wrap"><iframe src="`) {
			v := strings.Split(req.Description, `<div class="video-wrap"><iframe src="`)
			if len(v) > 1 {
				link := strings.Split(v[1], `"`)
				if len(link) > 0 {

					if service.CheckRejectLinkPost(link[0]) {

						url := service.GetImageIDFromIFrame(link[0])
						if url != "" {
							resp, err := http.Get(url)
							must(err)
							defer resp.Body.Close()

							image, _, err := image.Decode(resp.Body)

							req.HeightShare = image.Bounds().Dy()
							req.WidthShare = image.Bounds().Dx()

							m, w, h := resizeMainImage(image)
							name := generateMainImagePostName(me.ID)
							upload(ctx, m, name)

							req.ImageURL = generateDownloadURL(name)
							req.Width = w
							req.Height = h

							req.ImageShareURL = req.ImageURL
						}
					}
				}
			}
		}
	}

	req.StatusVerify = me.GetLevelPost()

	//edit
	err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.EditPost(tx, repository.EditPostModel{
			ID:              req.ID,
			Title:           req.Title,
			Description:     req.Description,
			Link:            req.Link,
			VdoURL:          req.VdoURL,
			ImageURL:        req.ImageURL,
			ImageURLMobile:  req.ImageURLMobile,
			Type:            req.Type,
			Province:        req.Province,
			Height:          req.Height,
			Width:           req.Width,
			StatusVerify:    req.StatusVerify,
			LinkDescription: req.LinkDescription,
			ImageShareURL:   req.ImageShareURL,
			HeightShare:     req.HeightShare,
			WidthShare:      req.WidthShare,
		})
		if err != nil {
			return NewAppError("ไม่สามารถแก้ไข Post ได้")
		}

		err = repository.UpdateTagTopic(db, tx, myRedis, req.ID, tagID)
		if err != nil {
			return NewAppError("ไม่สามารถแก้ไข Post ได้")
		}

		return nil
	})
	if IsAppError(err) {
		f.Add("Errors", err.Error())
		return ctx.Redirect("/edit/post/" + req.ID)
	}
	must(err)

	//update count
	updateCountEditPost(tagID)

	slug, err := repository.GetPostSlug(db, req.ID)
	must(err)

	if slug == "" {
		slug = req.ID
	}

	return ctx.RedirectTo("post.read", slug)
}

func ajaxShortenerURLPostHandler(ctx *hime.Context) error {

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	slug, err := repository.GetPostSlug(db, req.ID)
	if err == sql.ErrNoRows {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}
	must(err)

	res := entity.ResponseShortenerURL{}

	code, err := repository.GetShortenerURL(db, req.ID)
	if err == sql.ErrNoRows {
		loop := true
		for loop {
			newCode := repository.RandStr(7)

			err := repository.CheckCodeShortenerURL(db, newCode)
			if err == sql.ErrNoRows {

				err := repository.InsertShortenerURL(db, newCode, req.ID, baseURL+"/post/"+slug)
				must(err)

				res.URL = shortenerURL + "/" + newCode
				loop = false
				break
			}
		}

		return ctx.Status(http.StatusOK).JSON(&res)
	}
	must(err)

	res.URL = shortenerURL + "/" + code
	return ctx.Status(http.StatusOK).JSON(&res)
}
