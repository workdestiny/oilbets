package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/moonrhythm/hime"
)

const desc = "เว็บไซต์ คลิปโป้ คลิปหลุด หนังx หนังav หนังr porn เรามีคลิปโป๊ครบทุกแนวไม่ว่าจะเป็นคลิปหลุดข่มขืน xxxนักศึกษา xxxนักเรียน xporn ดูหนังxหนังโป๊คลิปหลุดทางบ้านตั้งกล้องถ่ายและอีกมากมาย"

func page(ctx *hime.Context) map[string]interface{} {
	sess := getSession(ctx)
	r := ctx.Request

	x := make(map[string]interface{})
	x["QueryURI"] = r.URL.Query().Get("type")
	if r.URL.Path != "" {
		x["Path"] = r.URL.Path
	}

	x["Tagline"] = "| ดูหนังx หนังav คลิปโป๊ คลิปหลุด หนังโป๊ xxxนักเรียน xxxporn"
	x["URL"] = "https://" + getHost(r) + r.RequestURI
	x["Meta"] = map[string]string{
		"Image":       "",
		"Description": desc,
	}
	x["User"] = getUser(ctx)
	x["Flash"] = sess.Flash()
	x["Now"] = time.Now()

	return x
}

// getHost gets real host from request
func getHost(r *http.Request) string {
	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return host
}

func setMeta(title, description, image string, width, height int) map[string]string {
	if title == "" {
		title = "Doujinxxx"
	}
	if description == "" {
		description = desc
	}
	if image == "" {
		image = ""
	}
	if width == 0 {
		width = 1200
	}
	if height == 0 {
		height = 630
	}

	return map[string]string{
		"Title":       title,
		"Description": description,
		"Image":       image,
		"Width":       strconv.Itoa(width),
		"Height":      strconv.Itoa(height),
	}
}
