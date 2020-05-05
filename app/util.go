package app

import (
	"github.com/acoshift/httprouter"
	"github.com/moonrhythm/hime"
)

func getParams(ctx *hime.Context, s string) string {
	return httprouter.GetParams(ctx).ByName(s)
}
