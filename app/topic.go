package app

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/acoshift/pgsql"
	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
)

func topicGetHandler(ctx *hime.Context) error {
	code := getParams(ctx, "topic")
	id := GetMyID(ctx)

	topic, err := repository.GetTopicByCode(db, code)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	title := topic.Name
	description := "หัวข้อคอนเทนต์ที่เกี่ยวกับ: " + topic.Name
	image := ""

	if topic.Title != "" {
		title = topic.Title
	}
	if topic.Description != "" {
		description = topic.Description
	}
	if topic.ImageFacebook != "" {
		image = topic.ImageFacebook
	}

	p := page(ctx)
	p["Topic"] = topic

	if id != "" {
		t, err := repository.ListTopicRecommendForCustomer(db, config.LimitListTopicRecommend)
		must(err)
		p["Topics"] = t
	}

	if id == "" {
		t, err := repository.ListTopicRecommend(db, config.LimitListTopicRecommend)
		must(err)
		p["Topics"] = t
	}

	p["ParamID"] = code
	if topic.Tagline != "" {
		p["Tagline"] = "| " + topic.Tagline
	}

	p["M"] = setMeta(title, description, image, 0, 0)

	return ctx.View("app/topic", p)
}

func ajaxTopicPostHandler(ctx *hime.Context) error {

	topic := getParams(ctx, "topic")
	userID := GetMyID(ctx)

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponsePost{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		post, err := repository.ListPostTopic(db, myRedis, userID, topic, config.LimitListPostTopic)
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

	post, err := repository.ListPostTopicNextLoad(db, myRedis, userID, topic, req.Next, config.LimitListPostTopicNext)
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

func ajaxFollowTopicPostHandler(ctx *hime.Context) error {
	id := GetMyID(ctx)
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	if req.ID == config.TopicOfficialID {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.CheckTopicID(db, req.ID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResFollow{
		IsFollow: false,
	}

	follow, err := repository.CheckFollowTopic(db, id, req.ID)
	if err == sql.ErrNoRows {

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.CreateFollowTopic(tx, id, req.ID)
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

		res.IsFollow = true
		return ctx.Status(http.StatusOK).JSON(&res)
	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {
		err = repository.FollowTopic(tx, id, req.ID, !follow)

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

	res.IsFollow = !follow
	return ctx.Status(http.StatusOK).JSON(&res)
}

func followTopicListPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()

	id := GetMyID(ctx)

	var req entity.RequestFollowTopicList
	json.Unmarshal([]byte(ctx.PostFormValue("id")), &req.ID)

	if len(req.ID) == 0 {
		f.Add("ErrLenID", "กรุณาส่ง topic id")
	}

	if f.Has("ErrLenID") {
		return ctx.RedirectToGet()
	}

	for _, v := range req.ID {

		err := repository.CheckTopicID(db, v.ID)
		if err != nil {
			continue
		}

		follow, err := repository.CheckFollowTopic(db, id, v.ID)
		if err == sql.ErrNoRows {

			err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

				err = repository.CreateFollowTopic(tx, id, v.ID)
				if err != nil {
					return NewAppError("ไม่สามารถติดตามได้")
				}

				return nil
			})

			if IsAppError(err) {
				f.Add("ErrFollow", err.Error)
				return ctx.RedirectToGet()
			}
			must(err)

			continue
		}
		must(err)

		if !follow {

			err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {
				err = repository.FollowTopic(tx, id, v.ID, true)

				if err != nil {
					return NewAppError("ไม่สามารถติดตามได้")
				}

				return nil
			})
			if IsAppError(err) {
				f.Add("ErrFollow", err.Error)
				return ctx.RedirectToGet()
			}
			must(err)
		}
	}

	return ctx.RedirectTo("discover")
}

func ajaxListTopicPostHandler(ctx *hime.Context) error {

	var req entity.Request
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	err = repository.CheckCategory(db, req.ID)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	topics, err := repository.ListTopicVerifiedByCategory(db, req.ID)
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(nil)
	}

	if len(topics) == 0 {
		return ctx.NoContent()
	}

	res := entity.ResponseTopic{
		Topic: topics,
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}
