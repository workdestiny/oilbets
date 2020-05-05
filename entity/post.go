package entity

import (
	"math"
	"time"
)

// PostModel is post
type PostModel struct {
	ID              string
	Slug            string
	OwnerID         string
	OwnerType       int
	Title           string
	TagTopics       []TagTopic
	Description     string
	LinkDescription string
	ImageURL        string
	ImageURLMobile  string
	Height          int
	Width           int
	Link            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Type            int
	catID           int64
	TopicID         int64
	Province        int
	LikeCount       int
	CommentCount    int
	ShareCount      int
	ViewCount       int
	Status          bool
	StatusVerify    bool
	StatusUsed      bool
}

//DiscoverModel is post
type DiscoverModel struct {
	ID              string     `json:"id"`
	Slug            string     `json:"slug"`
	Owner           PostOwner  `json:"owner"`
	Title           string     `json:"title"`
	TagTopics       []TagTopic `json:"tagTopics"`
	Description     string     `json:"description"`
	LinkDescription string     `json:"linkDescription"`
	ImageURL        string     `json:"imageURL"`
	Height          int        `json:"height"`
	Width           int        `json:"width"`
	IsLike          bool       `json:"isLike"`
	Link            string     `json:"link"`
	CreatedAt       time.Time  `json:"createdAt"`
	Time            string     `json:"time"`
	Type            int        `json:"type"`
	Count           PostCount  `json:"count"`
}

//GetPostModel is post
type GetPostModel struct {
	ID              string     `json:"id"`
	Slug            string     `json:"slug"`
	Owner           PostOwner  `json:"owner"`
	Title           string     `json:"title"`
	TagTopics       []TagTopic `json:"tagTopics"`
	Description     string     `json:"description"`
	LinkDescription string     `json:"linkDescription"`
	ImageURL        string     `json:"imageURL"`
	Height          int        `json:"height"`
	Width           int        `json:"width"`
	ImageShareURL   string     `json:"imageShareURL"`
	HeightShare     int        `json:"heightShare"`
	WidthShare      int        `json:"widthShare"`
	IsLike          bool       `json:"isLike"`
	Link            string     `json:"link"`
	VdoURL          string     `json:"vdoURL"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Province        string     `json:"province"`
	Type            int        `json:"type"`
	Count           PostCount  `json:"count"`
}

//GetViewModel view and guest model
type GetViewModel struct {
	View      int
	GuestView int
	All       int
}

//Sum GetViewModel
func (v GetViewModel) Sum() int {
	return v.View + v.GuestView
}

//GetUserAgentViewModel view and guest model
type GetUserAgentViewModel struct {
	Mobile         int
	MobilePercent  float64
	Desktop        int
	DesktopPercent float64
}

//CalculatePercentMobile GetViewModel
func (v GetUserAgentViewModel) CalculatePercentMobile() float64 {
	if v.Mobile == 0 {
		return 0.0
	}

	max := v.Mobile + v.Desktop
	percent := float64(v.Mobile) / float64(max)
	return math.Round(percent*100) / 100
}

//CalculatePercentDesktop GetViewModel
func (v GetUserAgentViewModel) CalculatePercentDesktop() float64 {
	if v.Desktop == 0 {
		return 0.0
	}

	max := v.Mobile + v.Desktop
	percent := float64(v.Desktop) / float64(max)
	return math.Round(percent*100) / 100
}

//PostStatisticModel view and guest model
type PostStatisticModel struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	View        int    `json:"view"`
	GuestView   int    `json:"guestView"`
	All         int    `json:"all"`
	Like        int    `json:"like"`
	Comment     int    `json:"comment"`
	CreatedAt   string `json:"createdAt"`
}

//Sum on PostStatisticModel
func (p PostStatisticModel) Sum() int {
	return p.View + p.GuestView
}

//PostRevenueModel view and guest model
type PostRevenueModel struct {
	ID          string              `json:"id"`
	Slug        string              `json:"slug"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Image       string              `json:"image"`
	View        GetViewRevenueModel `json:"view"`
	Note        []string            `json:"note"`
	Reject      bool                `json:"reject"`
	CreatedAt   time.Time           `json:"createdAt"`
	Time        string              `json:"time"`
}

//GetDraftPostModel is post
type GetDraftPostModel struct {
	ID          string
	Owner       DraftOwner
	Title       string
	Description string
	Type        int
	Time        string
}

//TopPostRelateModel is model
type TopPostRelateModel struct {
	ID       string
	Slug     string
	ImageURL string
	Title    string
	View     int
}

//PostOwner Gap
type PostOwner struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Display  string `json:"display"`
	IsFollow bool   `json:"isFollow"`
	IsVerify bool   `json:"isVerify"`
	IsOwner  bool   `json:"isOwner"`
}

//DraftOwner Gap
type DraftOwner struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Display string `json:"display"`
}

//PostCount Count Like Comment View Post
type PostCount struct {
	Like    string `json:"like"`
	Comment string `json:"comment"`
	View    string `json:"view"`
}

// TypePost Type
type TypePost int

const (
	// TypePostArticle is text
	TypePostArticle TypePost = iota
	// TypePostImage is image
	TypePostImage
	// TypePostLink is link
	TypePostLink
)

var mapTypePost = map[TypePost]string{
	TypePostArticle: "article",
	TypePostImage:   "image",
	TypePostLink:    "link",
}

func (x TypePost) String() string {
	return mapTypePost[x]
}

// TagTopic post have TagTopic =< 3 Tag
type TagTopic struct {
	TopicID string `json:"id"`
	Name    string `json:"tag"`
}

//Request model ajax
type Request struct {
	ID   string    `json:"id"`
	Next time.Time `json:"next"`
}

//RequestRevenue model ajax
type RequestRevenue struct {
	ID   string    `json:"id"`
	Next time.Time `json:"next"`
	Time time.Time `json:"time"`
}

//RequestNextLoad model ajax
type RequestNextLoad struct {
	Next time.Time `json:"next"`
	Type string    `json:"type"`
}

//RequestNotificationNextLoad model ajax
type RequestNotificationNextLoad struct {
	Next time.Time        `json:"next"`
	Type NotificationType `json:"type"`
}

//RequestEmailVerify model ajax
type RequestEmailVerify struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	RepeatPassword string `json:"repeatPassword"`
}

//RequestFollowTopicList model
type RequestFollowTopicList struct {
	ID []RequestFollowTopicID
}

//RequestFollowTopicID model
type RequestFollowTopicID struct {
	ID string `json:"id"`
}

// RequestComment modal ajax
type RequestComment struct {
	GapID  string           `json:"gapID"`
	PostID string           `json:"postID"`
	Type   TypeOwnerComment `json:"type"`
	Text   string           `json:"text"`
}

//RequestDeleteComment modal ajax
type RequestDeleteComment struct {
	CommentID string `json:"commentID"`
	PostID    string `json:"postID"`
}

//RequestEditComment modal ajax
type RequestEditComment struct {
	CommentID string `json:"commentID"`
	Text      string `json:"text"`
}

//RequestReject modal ajax
type RequestReject struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

//RequestPost modal ajax
type RequestPost struct {
	OwnerID         string   `json:"ownerID"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	LinkDescription string   `json:"linkDescription"`
	Link            string   `json:"link"`
	Type            TypePost `json:"type"`
	Province        int      `json:"province"`
}

//ResponsePost modal ajax
type ResponsePost struct {
	Next time.Time        `json:"next"`
	Post []*DiscoverModel `json:"post"`
}

//ResponseUserFollowerGapModel modal ajax
type ResponseUserFollowerGapModel struct {
	Next   time.Time               `json:"next"`
	IsNext bool                    `json:"isNext"`
	User   []*UserFollowerGapModel `json:"user"`
}

//ResponseDeleteComment modal ajax
type ResponseDeleteComment struct {
	CommentID string `json:"commentID"`
}

//TagTopicRequest model ajax
type TagTopicRequest struct {
	Text string `json:"text"`
}

//RequestSearch model ajax
type RequestSearch struct {
	Text string `json:"text"`
}

//ResponseSearch model ajax
type ResponseSearch struct {
	Topic []*RedisTopicModel `json:"topic"`
	// Gap   []*RedisGapModel   `json:"gap"`
	Post []*DiscoverModel `json:"post"`
}

//ResponseTopic model ajax
type ResponseTopic struct {
	Topic []*TopicList `json:"topicList"`
}

//RequestGapStatistic model ajax
type RequestGapStatistic struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Start string `json:"start"`
	End   string `json:"end"`
	Page  int    `json:"page"`
}

//ResponseGapStatistic model ajax
type ResponseGapStatistic struct {
	Post []*PostStatisticModel `json:"post"`
}

//ResponseNotification model ajax
type ResponseNotification struct {
	Notification []*NotificationModel `json:"notification"`
	Count        int                  `json:"count"`
	Next         time.Time            `json:"next"`
	IsNext       bool                 `json:"isNext"`
}

//ResLike model ajax
type ResLike struct {
	IsLike bool `json:"isLike"`
}

//ResFollow model ajax
type ResFollow struct {
	IsFollow bool `json:"isfollow"`
}

//ResDraftImage model ajax
type ResDraftImage struct {
	URL      string              `json:"url,omitempty"`
	Uploaded bool                `json:"uploaded"`
	Error    *ResErrorDraftImage `json:"error,omitempty"`
}

//ResErrorDraftImage model ajax
type ResErrorDraftImage struct {
	Message string `json:"message,omitempty"`
}

//ResponseError model ajax
type ResponseError struct {
	Message string `json:"message"`
	Errors  string `json:"errors"`
}

//ResponseShortenerURL model ajax
type ResponseShortenerURL struct {
	URL string `json:"url"`
}

//CommentModel is model
type CommentModel struct {
	ID        string       `json:"id"`
	Owner     CommentOwner `json:"owner"`
	Text      string       `json:"text"`
	Time      string       `json:"time"`
	CreatedAt time.Time    `json:"createdAt"`
}

//CommentOwner Owner
type CommentOwner struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Display  string `json:"display"`
	Type     string `json:"type"`
}

//ResponseComment modal ajax
type ResponseComment struct {
	ID    string               `json:"id"`
	Owner ResponseCommentOwner `json:"owner"`
	Text  string               `json:"text"`
	Time  string               `json:"time"`
}

//ResponseCommentOwner modal ajax
type ResponseCommentOwner struct {
	ID      string           `json:"id"`
	Name    string           `json:"name"`
	Display string           `json:"display"`
	Type    TypeOwnerComment `json:"type"`
}

//ResponseCommentList modal ajax
type ResponseCommentList struct {
	Comment []*CommentModel `json:"comment"`
	Next    time.Time       `json:"next"`
	IsNext  bool            `json:"isNext"`
}

//ResponseUserFollowGapList modal ajax
type ResponseUserFollowGapList struct {
	Gap    []*UserFollowGapListModel `json:"gap"`
	Next   time.Time                 `json:"next"`
	IsNext bool                      `json:"isNext"`
}

// TypeOwnerComment Type
type TypeOwnerComment int

const (
	// TypeOwnerCommentUser is user
	TypeOwnerCommentUser TypeOwnerComment = iota
	// TypeOwnerCommentGap is gap
	TypeOwnerCommentGap
)

var mapTypeOwnerComment = map[TypeOwnerComment]string{
	TypeOwnerCommentUser: "user",
	TypeOwnerCommentGap:  "gap",
}

func (x TypeOwnerComment) String() string {
	return mapTypeOwnerComment[x]
}

//ResponsePostRevenue modal ajax
type ResponsePostRevenue struct {
	Next time.Time           `json:"next"`
	Post []*PostRevenueModel `json:"post"`
}
