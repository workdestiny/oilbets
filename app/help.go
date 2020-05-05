package app

import (
	"github.com/workdestiny/amlporn/entity"
	"github.com/moonrhythm/hime"
)

func helpPostGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	p["List"] = entity.HelpPost

	return ctx.View("app/help-post", p)
}
