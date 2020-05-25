package app

import (
	"github.com/moonrhythm/hime"
)

func discoverGetHandler(ctx *hime.Context) error {

	return ctx.View("app/discover", page(ctx))
}
