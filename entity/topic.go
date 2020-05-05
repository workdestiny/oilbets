package entity

//TopicRecommendModel is model
type TopicRecommendModel struct {
	ID   string
	Name string
}

//TopicList is model
type TopicList struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	UsedCount int    `json:"usedCount"`
	Title     string `json:"title"`
}

//CategoryList is model
type CategoryList struct {
	ID    string           `json:"id"`
	Code  string           `json:"code"`
	Name  string           `json:"name"`
	Count int              `json:"count"`
	Topic []TopicListModel `json:"topic"`
}

//CategoryDiscoverList is model
type CategoryDiscoverList struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

//TopicModel is model
type TopicModel struct {
	ID            string
	CatID         string
	Code          string
	Name          string
	Image         string
	ImageMini     string
	Count         int
	UsedCount     int
	Verify        bool
	Title         string
	Description   string
	ImageFacebook string
	Tagline       string
}

//TopicListModel is model
type TopicListModel struct {
	ID       string
	CatID    string
	Code     string
	Name     string
	Image    string
	Count    int
	IsFollow bool
}
