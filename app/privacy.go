package app

import (
	"github.com/moonrhythm/hime"
)

func privacyGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("app/privacy", p)
}

func conditionGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("app/condition", p)
}

func testGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("app/test", p)
}
