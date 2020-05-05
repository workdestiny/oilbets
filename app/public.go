package app

import (
	"net/http"

	"github.com/mileusna/useragent"

	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
)

func publicGetHandler(ctx *hime.Context) error {

	referrer := getReferrer(ctx)
	if referrer != "" {
		removeSessionReferrer(ctx)
		return ctx.Redirect(referrer)
	}

	id := GetMyID(ctx)

	p := page(ctx)
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

	return ctx.View("app/public", p)
}

func publicPostHandler(ctx *hime.Context) error {

	mo := ua.Parse(ctx.UserAgent())
	userID := GetMyID(ctx)
	if userID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	//topicIDs := repository.ListFollowTopicID(db, userID)

	res := entity.ResponsePost{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		post, err := repository.ListPostPublic(db, myRedis, userID, req.Type, mo.Mobile, config.LimitPublic)
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
	post, err := repository.ListPostPublicNextLoad(db, myRedis, userID, req.Type, req.Next, mo.Mobile, config.LimitPublicNext)
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

func readPostHandler(ctx *hime.Context) error {

	userID := GetMyID(ctx)
	postID := getParams(ctx, "postID")

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	topicIDs := repository.ListTagTopicIDOnPost(db, postID)

	if len(topicIDs) == 0 {
		return ctx.NoContent()
	}

	res := entity.ResponsePost{}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {

		t, err := repository.GetPostCreateTime(db, postID)
		if err != nil {
			return ctx.Status(http.StatusInternalServerError).JSON(nil)
		}

		post, err := repository.ListPostRead(db, myRedis, userID, t, topicIDs, postID, config.LimitPublic)
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

	post, err := repository.ListPostReadNextLoad(db, myRedis, userID, req.Type, topicIDs, postID, req.Next, config.LimitPublicNext)
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
