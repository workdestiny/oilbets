package entity

// HelpPostModel is type for help post
type HelpPostModel struct {
	Title string
	URL   []string
}

// HelpPost is data for help post
var HelpPost = []HelpPostModel{
	{
		Title: "เพิ่มสถานที่ให้กับคอนเทนต์",
		URL: []string{
			"",
		},
	},
	{
		Title: "การบันทึกฉบับร่าง",
		URL: []string{
			"",
		},
	},
	{
		Title: "เพิ่มแท็กให้กับคอนเทนต์",
		URL: []string{
			"",
		},
	},
	{
		Title: "เครื่องมือจัดรูปแบบเนื้อหา",
		URL: []string{
			"",
			"",
		},
	},
	{
		Title: "เครื่องมือเสริมการเขียนคอนเทนต์",
		URL: []string{
			"",
			"",
		},
	},
}
