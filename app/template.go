package app

import (
	"html/template"

	"github.com/workdestiny/convbox"
	"github.com/workdestiny/oilbets/service"
)

// TemplateFuncs returns template funcs
func (app *App) TemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"static": func(s string) string {
			p := app.Static[s]
			return "/-" + p
		},
		"shortText":          service.ShortText,          // short for wording
		"shortTextPost":      service.ShortTextPost,      // short text for post only
		"shortTextStripHTML": service.ShortTextStripHTML, // short text return html
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"sanitize": func(s string) string {
			return service.StripHTML(s)
		},
		"stripHtml": func(s string) template.HTML {
			text := service.StripHTML(s)
			return template.HTML(text)
		},
		"shortNumber": func(i int) string {
			return convbox.ShortNumber(i)

		},
		"formatTime":   service.FormatTime,
		"formatDate":   service.FormatDate,
		"reFormatTime": service.ReFormatTime,
		"facebookAppID": func() string {
			return facebookAppID
		},
		"postNextURL": func(templateName string) string {
			var path string
			switch templateName {
			case "app/discover":
				path = app.Hime.Route("ajax.post.discover")
			case "app/follow":
				path = app.Hime.Route("ajax.post.follow")
			case "app/public":
				path = app.Hime.Route("ajax.post.public")
			case "app/topic":
				path = app.Hime.Route("ajax.post.topic")
			case "app/gap":
				path = app.Hime.Route("ajax.post.gap")
			case "app/categoryget":
				path = app.Hime.Route("ajax.post.category")
			case "app/post-read":
				path = app.Hime.Route("ajax.post.read")
			case "app/liked":
				path = app.Hime.Route("ajax.post.liked")
			default:
				path = ""
			}
			return path
		},
		"currency": service.Currency,
		"add": func(x, y int) int {
			return x + y
		},
		"isNewEditor": service.IsNewEditor,
	}
}
