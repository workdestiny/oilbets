package entity

import "time"

//NotificationModel is model
type NotificationModel struct {
	ID          string           `json:"id"`
	User        NotificationUser `json:"user"`
	Gap         NotificationGap  `json:"gap"`
	Type        NotificationType `json:"type"`
	PostID      string           `json:"postID"`
	Title       string           `json:"title"`
	CreatedAt   time.Time        `json:"createdAt"`
	Time        string           `json:"time"`
	Count       int              `json:"count"`
	Read        bool             `json:"read"`
	CommentText string           `json:"commentText"`
}

//NotificationUser is user do noti
type NotificationUser struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Display string `json:"display"`
}

//NotificationGap is gap noti
type NotificationGap struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Display  string `json:"display"`
	Username string `json:"username"`
}

//NotificationType is type like, comment, follow
type NotificationType int

const (
	// Like is LikePost
	Like NotificationType = iota
	// Comment is CommentPost
	Comment
	// Follow is FollowGap
	Follow
)

var mapNotificationTypeString = map[NotificationType]string{
	Like:    "like",
	Comment: "comment",
	Follow:  "follow",
}
