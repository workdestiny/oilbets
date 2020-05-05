package app

import (
	"net/http"

	"github.com/mileusna/useragent"
	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
)

func discoverGetHandler(ctx *hime.Context) error {

	referrer := getReferrer(ctx)
	if referrer != "" {
		removeSessionReferrer(ctx)
		return ctx.Redirect(referrer)
	}

	id := GetMyID(ctx)
	//mo := ua.Parse(ctx.UserAgent())

	if id != "" {
		if !GetUserFollowTopic(ctx) {

			topic, err := repository.ListTopicVerified(db, GetUserRole(ctx))
			must(err)

			p := page(ctx)
			p["ListTopic"] = topic
			p["NonFollow"] = true
			//p["IsMobile"] = mo.Mobile

			return ctx.View("app/discover", p)
		}

		t, err := repository.ListTopicRecommendForCustomer(db, config.LimitListTopicRecommend)
		must(err)
		p := page(ctx)
		p["Topics"] = t
		p["NonFollow"] = false
		//p["IsMobile"] = mo.Mobile

		return ctx.View("app/discover", p)
	}

	t, err := repository.ListTopicRecommend(db, config.LimitListTopicRecommend)
	must(err)

	p := page(ctx)
	p["Topics"] = t
	p["NonFollow"] = false
	//p["IsMobile"] = mo.Mobile

	return ctx.View("app/discover", p)
}

func ajaxDiscoverPostHandler(ctx *hime.Context) error {
	// if GetMyID(ctx) != "" {
	// 	if !GetUserFollowTopic(ctx) {
	// 		return discoverPostHandler(ctx)
	// 	}
	// 	return publicPostHandler(ctx)
	// }
	// return discoverPostHandler(ctx)
	return discoverPostHandler(ctx)
}

func discoverPostHandler(ctx *hime.Context) error {
	userID := GetMyID(ctx)
	mo := ua.Parse(ctx.UserAgent())

	var req entity.RequestNextLoad
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	res := entity.ResponsePost{}

	//topicIDs := repository.ListTopicIDVerified(db)
	//topicIDs := []string{"21f00a69-0617-11e9-8c97-363671da6b2f", "d724074a-042d-11e9-8c97-363671da6b2f"}

	if req.Next.String() == "1970-01-01 00:00:00 +0000 UTC" {
		//post, err := repository.ListPostDiscover(db, myRedis, userID, req.Type, config.LimitDiscover)
		post, err := repository.ListPostPublic(db, myRedis, userID, req.Type, mo.Mobile, config.LimitDiscover)
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
	//post, err := repository.ListPostDiscoverNextLoad(db, myRedis, userID, req.Type, req.Next, config.LimitDiscover)
	post, err := repository.ListPostPublicNextLoad(db, myRedis, userID, req.Type, req.Next, mo.Mobile, config.LimitDiscoverNext)
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
