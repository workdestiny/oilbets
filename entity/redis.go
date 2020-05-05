package entity

//RedisUserModel is Model in redis
type RedisUserModel struct {
	ID               string
	Username         string
	FirstName        string
	LastName         string
	DisplayImage     string
	DisplayImageMini string
	Level            UserLevelType
}

//RedisGapModel is Model in redis
type RedisGapModel struct {
	ID               string `json:"id"`
	UserID           string `json:"userID"`
	Username         string `json:"username"`
	Name             string `json:"name"`
	DisplayImage     string `json:"display"`
	DisplayImageMini string `json:"displayMini"`
	CountFollower    int    `json:"countFollower"`
	CountPopular     int    `json:"countPopular"`
}

//RedisCategoryModel is Model in redis
type RedisCategoryModel struct {
	ID     int64
	Code   string
	Name   Name
	Images Image
	Count  int
}

//RedisTopicModel is Model in redis
type RedisTopicModel struct {
	ID        string `json:"id"`
	CatID     string `json:"catID"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Images    Image  `json:"image"`
	Count     int    `json:"count"`
	UsedCount int    `json:"usedCount"`
}

//Name Topic and category
type Name struct {
	Th string
	En string
}

//Image Topic and category
type Image struct {
	Normal string `json:"normal"`
	Mini   string `json:"mini"`
}
