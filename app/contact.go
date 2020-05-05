package app

import "github.com/moonrhythm/hime"

func contactGetHandler(ctx *hime.Context) error {
	return ctx.View("app/contact", page(ctx))
}
