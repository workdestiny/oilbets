package app

import (
	"github.com/moonrhythm/hime"
)

func registerGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("app/register", p)
}

func walletGetHandler(ctx *hime.Context) error {
	p := page(ctx)
	return ctx.View("app/wallet", p)
}
