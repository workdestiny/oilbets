package entity

import (
	"time"

	"github.com/acoshift/ds"
)

// UserProviderModel is model in ds
type UserProviderModel struct {
	ds.Model //ID = userID , kind = UserProvider
	ds.StampModel
	Providers        []ProviderData
	Time             time.Time
	Count            int
	CodeSocialSignin string
	UpdateEmail      UpdateEmail
}

// NewKey auto set ds Key UserProviderModel
func (userProviderModel *UserProviderModel) NewKey() {
	userProviderModel.NewIncomplateKey(KindUserProvider, nil)
}

// UserModel is model in ds
type UserModel struct {
	ds.Model
	ds.StampModel
	Email     string
	Username  Username
	Tokens    []Token
	FirstName string
	LastName  string
	//Bio              string
	DisplayImage     string
	DisplayImageMini string
	//CoverImage       string
	//CoverImageMini   string
	BirthDate   time.Time
	Gender      string
	Contact     Contact
	AboutMe     []AboutMe
	Status      Status
	EmailVerify EmailVerify
	Verify      Verify
	ToolTip     ToolTip
	Count       Count
	Role        Role
	HasPage     bool
	StatusUsed  bool
}

// NewKey auto set ds Key UserModel
func (userModel *UserModel) NewKey() {
	userModel.NewIncomplateKey(KindUser, nil)
}

// UserAudit is model in ds
type UserAudit struct {
	ds.Model
	ds.StampModel
	ID       int64
	Type     string
	Snapshot *UserModel
}

// NewKey auto set ds Key UserAudit
func (UserAudit *UserAudit) NewKey() {
	UserAudit.NewIncomplateKey(KindUserAudit, nil)
}

// FollowPageModel is model in ds
type FollowPageModel struct {
	ds.Model
	ds.StampModel
	OwnerID    int64
	PageID     int64
	Status     bool
	StatusUsed bool
}

// NewKey auto set ds Key FollowPageModel
func (FollowPageModel *FollowPageModel) NewKey() {
	FollowPageModel.NewIncomplateKey(KindFollowPage, nil)
}

// ProviderData is data user auth
type ProviderData struct {
	Type       UserProviderType
	Password   string `datastore:",omitempty,noindex"`
	ProviderID string
	Code       string
}

// UpdateEmail is model
type UpdateEmail struct {
	NewEmail     string
	NewEmailCode string
	OldEmailCode string
	ExpreAt      time.Time
}

// UserProviderType is user signin with, Google, Facebook
type UserProviderType int

const (
	// UserProviderTypeX is UserProviderType
	UserProviderTypeX UserProviderType = iota
	// UserProviderTypeGoogle is UserProviderType
	UserProviderTypeGoogle
	// UserProviderTypeFacebook is UserProviderType
	UserProviderTypeFacebook
)

// StatisticsToken is Model Token
type StatisticsToken struct {
	ds.Model
	ds.StampModel
	// UserID ของผู้ใช้ที่ Signin เข้ามาในระบบ
	UserID       int64
	RefreshToken string
	UserAgent    string
	Location     Location
	SignOutAt    time.Time
	StatusSignIn bool
	SigninType   SigninType
}

// NewKey is auto create key to ds Model
func (StatisticsToken *StatisticsToken) NewKey() {
	StatisticsToken.NewIncomplateKey(KindStatisticsToken, nil)
}

// Location is model Location Signin
type Location struct {
	Lat  string
	Long string
}

// SigninType is Check User Signin with , Facebook, Google
type SigninType int

const (
	// TtX Signin with
	TtX SigninType = iota
	// TtFacebook Signin with Facebook
	TtFacebook
	// TtGoogle Signin with Google
	TtGoogle
)

// Count data in struct Me
type Count struct {
	Topic int
	Page  int
}

// Username is model Username
type Username struct {
	Text string
	Swap bool
}

// Token is Model Token
type Token struct {
	Type   TokenType
	Token  string
	TimeAt time.Time
}

// Status is Model
type Status struct {
	UserStatusLevel int
	CreatedAt       time.Time
	ExpreAt         time.Time
}

// EmailVerify is Model
type EmailVerify struct {
	Status bool
	At     time.Time
	Code   string
}

// Verify is Model
type Verify struct {
	At              time.Time
	UserVerifyLevel UserVerifyLevel
}

// TokenType ตรวจสอบว่า Signin ผ่านทางไหน
type TokenType int

const (
	// WebToken ผ่านหน้าเว็ป
	WebToken TokenType = iota
	// AndroidToken ผ่านระบบ Android
	AndroidToken
	// IosToken ผ่านระบบ Apple IOS
	IosToken
)

// UserVerifyLevel is Model ตรวจสอบว่าเป็น user ระดับไหน Normal Creator Official
type UserVerifyLevel int

const (
	// Normal เป็นผู้ใช้ธรรมดา
	Normal UserVerifyLevel = iota
	// Creator ผู้ใช้ระดับ Creator
	Creator
	// Official ผู้ใช้ระดับ Official
	Official
)

// UserStatusLevel is Status User
type UserStatusLevel int

const (
	// UserStatusLevelActive is Normal Status
	UserStatusLevelActive UserStatusLevel = iota
	// UserStatusLevelNonPost is Ban Post Status
	UserStatusLevelNonPost
	// UserStatusLevelNonLogin is Ban Signin Status
	UserStatusLevelNonLogin
	// UserStatusLevelBan is Ban Account Status
	UserStatusLevelBan
)

// Role User
type Role int

const (
	// RoleUser is normal
	RoleUser Role = iota
	// RoleSuperUser is super
	RoleSuperUser
	// RoleAdmin is admin
	RoleAdmin
	// RoleAgent is agent
	RoleAgent
)

var mapRoleString = map[Role]string{
	RoleUser:      "user",
	RoleSuperUser: "superuser",
	RoleAdmin:     "admin",
	RoleAgent:     "agent",
}

func (x Role) String() string {
	return mapRoleString[x]
}

// UserFollowerGapModel is model
type UserFollowerGapModel struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Display   string            `json:"display"`
	Gap       []GapUserFollower `json:"gap"`
	CreatedAt time.Time         `json:"createdAt"`
}

// GapUserFollower is model
type GapUserFollower struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Display  string `json:"display"`
}

// UserFollowGapListModel is model
type UserFollowGapListModel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Display   string    `json:"display"`
	IsVerify  bool      `json:"isVerify"`
	CreatedAt time.Time `json:"createdAt"`
}
