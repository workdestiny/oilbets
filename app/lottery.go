package app

import "github.com/moonrhythm/hime"

func lotteryGetHandler(ctx *hime.Context) error {

	//แสดงผลล่าสุด
	//แสดงรายการ การแท่งหวยของเรา
	// add table lottery
	//แสดง rate ราคา

	p := page(ctx)
	return ctx.View("app/lottery", p)
}
