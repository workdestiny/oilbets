package app

import (
	"database/sql"
	"net/http"

	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
)

func categoryGetHandler(ctx *hime.Context) error {

	category, err := repository.ListCategory(db, GetMyID(ctx), GetUserRole(ctx))
	must(err)

	p := page(ctx)
	p["Category"] = category
	return ctx.View("app/category", p)
}

func getCategoryGetHandler(ctx *hime.Context) error {
	code := getParams(ctx, "category")

	id, err := repository.GetCategoryIDByCode(db, code)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	topics, err := repository.ListTopicIsDataByCategory(db, id)
	must(err)

	p := page(ctx)
	p["Topic"] = topics
	p["ParamID"] = id
	p["Code"] = code
	return ctx.View("app/categoryget", p)
}

func ajaxCategoryPostHandler(ctx *hime.Context) error {

	id := getParams(ctx, "category")
	userID := GetMyID(ctx)

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponsePost{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		post, err := repository.ListPostCategory(db, myRedis, userID, id, config.LimitListPostTopic)
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

	post, err := repository.ListPostCategoryNextLoad(db, myRedis, userID, id, req.Next, config.LimitListPostTopicNext)
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
