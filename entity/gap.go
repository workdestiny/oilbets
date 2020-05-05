package entity

import "time"

//GapRecommendModel is model discover page
type GapRecommendModel struct {
	ID            string
	Name          string
	Username      string
	Display       string
	Cover         string
	FollowerCount int
	IsFollow      bool
	IsOwner       bool
}

//GapList use page create post
type GapList struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Display  string `json:"display"`
}

//GetGapModel use page Gap
type GetGapModel struct {
	ID       string
	Name     string
	Username string
	Bio      string
	Cover    string
	Display  string
	IsFollow bool
	IsVerify bool
	IsOwner  bool
	Count    CountGetGap
}

//GetUsername use page Gap
func (g *GetGapModel) GetUsername() string {
	if g.Username == "" {
		return g.ID
	}
	return g.Username
}

//GetGapDetailModel use page Gap
type GetGapDetailModel struct {
	ID            string
	Name          string
	Username      string
	Display       string
	DisplayMini   string
	UserID        string
	CountFollower int
	CountPopular  int
}

//GetGapSettingModel use Gap Setting
type GetGapSettingModel struct {
	ID               string
	Name             string
	IsChangeName     bool
	Bio              string
	Cover            string
	Display          string
	IsVerify         bool
	Count            CountGetGap
	CreatedAt        time.Time
	Topic            string
	Username         string
	IsChangeUsername bool
	Tel              string
	Email            string
	Social           string
	City             string
	Country          string
	Address          string
}

//GetUsername getUsername
func (g *GetGapSettingModel) GetUsername() string {
	if g.Username == "" {
		return g.ID
	}
	return g.Username
}

//GetGapRevenueModel use Gap revenue
type GetGapRevenueModel struct {
	ID       string
	Name     string
	Display  string
	Username string
	UserID   string
	Wallets  string
}

//GetUsername getUsername
func (g *GetGapRevenueModel) GetUsername() string {
	if g.Username == "" {
		return g.ID
	}
	return g.Username
}

//CountGetGap use page Gap
type CountGetGap struct {
	Post     int
	Follower int
}

//CountUserAgent count agent
type CountUserAgent struct {
	Mobile  int
	Desktop int
}

//CountViewHour count hour view
type CountViewHour struct {
	Hour  int
	Count int
}

//ViewHour model
type ViewHour struct {
	Hour  string
	Count int
}
