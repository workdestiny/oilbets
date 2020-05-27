package entity

import (
	"time"
)

// Me is data user
type Me struct {
	ID               string
	Email            string
	FirstName        string
	LastName         string
	DisplayImage     DisplayImage
	BirthDate        time.Time
	Gender           string
	Count            CountMe
	Role             Role
	Gap              []GapList
	IsVerify         bool
	IsSignin         bool
	IsNotification   bool
	IsVerifyIDCard   bool
	IsVerifyBookBank bool
	Wallet           int64
	Bonus            int64
	WithdrawRate     int64
}

// IsPost is check if user can post
func (m *Me) IsPost() bool {
	return m.Role > RoleUser
}

// HasPage is check if user has page
func (m *Me) HasPage() bool {
	return len(m.Gap) > 0
}

// IsAdmin get data RoleAdmin
func (m *Me) IsAdmin() bool {
	if m.Role == RoleAdmin {
		return true
	}
	return false
}

// GetLevel get UserLevelType
func (m *Me) GetLevel() UserLevelType {
	if m.Email == "" {
		return NoEmail
	}
	if !m.IsVerify {
		return NotVerify
	}
	if !m.IsVerifyIDCard {
		return VerifyEmail
	}
	if !m.IsVerifyBookBank {
		return VerifyIDCard
	}
	return VerifyBookBank
}

// GetLevelPost get UserLevelType
func (m *Me) GetLevelPost() int {
	if !m.IsVerify {
		if m.IsVerifyIDCard {
			return 2
		}
		return 0
	}
	if !m.IsVerifyIDCard {
		return 1
	}
	return 2
}

// UserLevelType type user level
type UserLevelType int

const (
	// NoEmail is UserLevelType
	NoEmail UserLevelType = iota
	// NotVerify is UserLevelType
	NotVerify
	// VerifyEmail is UserLevelType
	VerifyEmail
	// VerifyIDCard is UserLevelType
	VerifyIDCard
	// VerifyFaceIDCard is UserLevelType
	VerifyFaceIDCard
	// VerifyBookBank is UserLevelType
	VerifyBookBank
)

// IsEmail get data Email
func (m *Me) IsEmail() bool {
	if m.Email == "" {
		return false
	}
	return true
}

// DisplayImage data in struct Me
type DisplayImage struct {
	Normal string
	Mini   string
	Middle string
}

// AboutMe data in struct Me
type AboutMe struct {
	Type             string
	Value            string
	StatusUpdateTime time.Time
}

// BirthDate data in struct Me
type BirthDate struct {
	Day   string
	Month string
	Year  string
}

// Contact data in struct Me
type Contact struct {
	Address string `json:"address"`
	City    string `json:"city"`
	Country string `json:"country"`
}

// CountMe data in struct Me
type CountMe struct {
	Topic int
	Gap   int
}

// ToolTip .data in struct Me
type ToolTip struct {
	Post     bool
	Category bool
	Follow   bool
}

// EmailVerifyMe .data in struct Me
type EmailVerifyMe struct {
	Status bool
	At     time.Time
}
