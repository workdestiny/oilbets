package app

import (
	"net/http"

	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
)

func followGetHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	p := page(ctx)

	if me.Count.Gap > 1 {
		p["ListFollowGap"] = true
	}

	return ctx.View("app/follow", p)
}

func ajaxFollowPostHandler(ctx *hime.Context) error {

	userID := GetMyID(ctx)
	if userID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	gapIDs := repository.ListFollowGapID(db, userID)

	if len(gapIDs) == 0 {
		return ctx.NoContent()
	}

	res := entity.ResponsePost{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		post, err := repository.ListPostFollow(db, myRedis, userID, req.Type, gapIDs, config.LimitFollow)
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

	post, err := repository.ListPostFollowNextLoad(db, myRedis, userID, req.Type, gapIDs, req.Next, config.LimitFollowNext)
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
