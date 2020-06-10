package app

import (
	"log"

	"github.com/moonrhythm/hime"
)

func discoverGetHandler(ctx *hime.Context) error {
	log.Println("111")

	return ctx.View("app/discover", page(ctx))
}
