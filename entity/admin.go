package entity

import (
	"time"
)

//UserRequestModel is model
type UserRequestModel struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Display   string
	Type      UserRequestType
	CreatedAt time.Time
}

// UserRequestType is Type User Request
type UserRequestType int

const (
	// RequestVerifyIDCard request
	RequestVerifyIDCard UserRequestType = iota
	// RequestVerifyBookBank request
	RequestVerifyBookBank
)
