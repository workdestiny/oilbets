package app

import (
	"net/http"
	"strings"

	"github.com/moonrhythm/hime"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/repository"
)

func searchGetHandler(ctx *hime.Context) error {

	p := page(ctx)
	return ctx.View("app/search", p)
}

func ajaxNavigationBarSearchHandler(ctx *hime.Context) error {

	var req entity.RequestSearch
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	keyword := strings.Replace(req.Text, " ", "%", -1)
	keyword = strings.ToLower(keyword)
	req.Text = strings.Replace(req.Text, " ", "", -1)
	req.Text = strings.ToLower(req.Text)

	topics, err := repository.SearchTopic(myRedis, req.Text)
	must(err)

	// gaps, err := repository.SearchGap(myRedis, req.Text)
	// must(err)

	posts, err := repository.SearchPost(db, myRedis, GetMyID(ctx), keyword, 10)
	must(err)

	if len(topics) == 0 && len(posts) == 0 {
		return ctx.NoContent()
	}

	res := entity.ResponseSearch{
		Topic: topics,
		Post:  posts,
	}

	return ctx.Status(http.StatusOK).JSON(&res)
}
