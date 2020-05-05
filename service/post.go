package service

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/anthoz69/bluemonday"
)

// SanitizeUGCDescription is strip all danger html and script from Description
func SanitizeUGCDescription(s string) string {
	p := bluemonday.UGCPolicy()
	p.AllowAttrs("src", "allowfullscreen").OnElements("oembed")
	p.AllowAttrs("url").OnElements("oembed")
	p.AllowAttrs("target").OnElements("a")
	p.AllowAttrs("class", "style").OnElements("li", "span", "p", "strong", "div", "pre", "figcaption", "figure", "h2", "h3", "h4")
	p.AllowElements("figcaption")
	return p.Sanitize(s)
}

// SanitizeUGC is strip all danger html and script from Title
func SanitizeUGC(s string) string {
	return bluemonday.UGCPolicy().Sanitize(s)
}

// StripHTML is remove all html from string
func StripHTML(s string) string {
	return bluemonday.StrictPolicy().Sanitize(s)
}

// ShortText return length of expect string
func ShortText(lenght int, s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	var elips string
	if p := utf8.RuneCountInString(s); lenght > p {
		lenght = p
	} else {
		elips = "..."
	}
	return string([]rune(s)[:lenght]) + elips
}

// ShortTextNotification return length of expect string
func ShortTextNotification(lenght int, s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	var elips string
	if p := utf8.RuneCountInString(s); lenght > p {
		lenght = p
	} else {
		elips = ".."
	}
	return string([]rune(s)[:lenght]) + elips
}

// ShortTextPost use for post description return string
func ShortTextPost(lenght int, s string) string {
	var elips string
	if p := utf8.RuneCountInString(s); lenght > p {
		lenght = p
	} else {
		elips = "...อ่านต่อ"
	}
	return string([]rune(s)[:lenght]) + elips
}

// ShortTextStripHTML is stript all html and return length of expect string
func ShortTextStripHTML(lenght int, s string) string {
	text := UnescapeString(StripHTML(s))
	return ShortTextPost(lenght, text)
}

// ShortTextTitleStripHTML is stript all html and return length of expect string
func ShortTextTitleStripHTML(lenght int, s string) string {
	return ShortText(lenght, UnescapeString(StripHTML(s)))
}

// ShortTextMeta is cut text for meta seo
func ShortTextMeta(lenght int, s string) string {
	return ShortTextPost(lenght, UnescapeString(StripHTML(s)))
}

// ShortTextNoti is cut text for notification
func ShortTextNoti(lenght int, s string) string {
	return ShortTextNotification(lenght, UnescapeString(StripHTML(s)))
}

// UnescapeString is cut text for notification
func UnescapeString(s string) string {
	return html.UnescapeString(s)
}

// CheckRejectLinkPost is check
func CheckRejectLinkPost(s string) bool {
	Re := regexp.MustCompile(`/(player.|www.)?(vimeo\.com|youtu(be\.com|\.be|be\.googleapis\.com))\/(video\/|embed\/|watch\?v=|v\/)?([A-Za-z0-9._%-]*)(\&\S+)?/`)
	return Re.MatchString(s)
}

// GetImageIDFromIFrame is get id
func GetImageIDFromIFrame(s string) string {
	if strings.Contains(s, "youtube") {
		v := strings.Split(s, "embed/")
		if len(v) > 1 {
			return "http://img.youtube.com/vi/" + v[1] + "/maxresdefault.jpg"
		}
	}

	// if strings.Contains(s, "vimeo") {
	// 	v := strings.Split(s, "video/")
	// 	if len(v) > 1 {
	// 		return "http://img.youtube.com/vi/" + v[1] + "/maxresdefault.jpg"
	// 	}
	// }

	return ""
}
