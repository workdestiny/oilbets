package app

import (
	"log"
	"net/http"

	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
)

func likedGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("app/liked", p)
}

func ajaxLikedPostHandler(ctx *hime.Context) error {
	userID := GetMyID(ctx)

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponsePost{}

	//topicIDs := repository.ListTopicIDVerified(db)

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		post, err := repository.ListPostLiked(db, myRedis, userID, req.Type, config.LimitDiscover)
		if err != nil {
			log.Println(err)
			return ctx.Status(http.StatusInternalServerError).JSON(nil)
		}

		if len(post) == 0 {
			return ctx.NoContent()
		}

		res.Post = post
		res.Next = post[len(post)-1].CreatedAt

		return ctx.Status(http.StatusOK).JSON(&res)
	}

	post, err := repository.ListPostLikedNextLoad(db, myRedis, userID, req.Type, req.Next, config.LimitDiscoverNext)
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
