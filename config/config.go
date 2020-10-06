package config

import (
	"time"

	"github.com/shopspring/decimal"
)

var (

	//RevenueRateView ตัวคูณเงิน
	RevenueRateView = decimal.New(5, -2)
	//RevenueRateGuestView ตัวคูณเงิน
	RevenueRateGuestView = decimal.New(1, -3)
	//PivotDay Change Every thing
	PivotDay = time.Date(2018, 11, 11, 0, 0, 0, 0, time.UTC)
)

const (
	//FrontbackWinrate is Win Rate 30% (game Frontback)
	FrontbackWinrate = 40
	//YingchobWinrate is Win Rate 30% (game Yingchob)
	YingchobWinrate = 30
	//YingchobDrawrate is Win Rate 30% (game Yingchob)
	YingchobDrawrate = 60

	//YingchobWinrate21 is Win Rate 35% (game Yingchob)
	YingchobWinrate21 = 40
	//YingchobDrawrate21 is Win Rate 30% (game Yingchob)
	YingchobDrawrate21 = 70
	//YingchobWinrate22 is Win Rate 30% (game Yingchob)
	YingchobWinrate22 = 35
	//YingchobDrawrate22 is Win Rate 30% (game Yingchob)
	YingchobDrawrate22 = 65

	//YingchobWinrate31 is Win Rate 50% (game Yingchob)
	YingchobWinrate31 = 50
	//YingchobDrawrate31 is Win Rate 30% (game Yingchob)
	YingchobDrawrate31 = 80
	//YingchobWinrate32 is Win Rate 40% (game Yingchob)
	YingchobWinrate32 = 40
	//YingchobDrawrate32 is Win Rate 30% (game Yingchob)
	YingchobDrawrate32 = 70
	//YingchobWinrate33 is Win Rate 30% (game Yingchob)
	YingchobWinrate33 = 25
	//YingchobDrawrate33 is Win Rate 25% (game Yingchob)
	YingchobDrawrate33 = 55
	//YingchobLimitWin is Win Rate 30% (game Yingchob)

	YingchobLimitWin = 40
	//YingchobMuti2 muti roll
	YingchobMuti2 = 3
	//YingchobMuti3 muti roll
	YingchobMuti3 = 10

	//HighlowWinrate is Win Rate 50% (game Hilo)
	HighlowWinrate = 30
	//HighlowRandomRoll is Rate 50%
	HighlowRandomRoll = 95
	//HighlowCountdown is duration time countdown
	HighlowCountdown = 180
	//LimitDiscover limit post
	LimitDiscover = 12
	//LimitPublic limit post
	LimitPublic = 12
	//LimitFollow limit post
	LimitFollow = 12
	//LimitListPostGap limit
	LimitListPostGap = 12
	//LimitListPostTopic limit
	LimitListPostTopic = 12
	//LimitDiscoverNext limit post
	LimitDiscoverNext = 12
	//LimitPublicNext limit post
	LimitPublicNext = 12
	//LimitFollowNext limit post
	LimitFollowNext = 12
	//LimitListPostGapNext limit
	LimitListPostGapNext = 12
	//LimitListPostTopicNext limit
	LimitListPostTopicNext = 12

	//LimitListUserFollowrGap limit
	LimitListUserFollowrGap = 10
	// LimitAllNotificationCount notification
	LimitAllNotificationCount = 48
	//LimitListGap limit
	LimitListGap = 1000
	//LimitListTopic limit
	LimitListTopic = 3
	//LimitListTopPostGap limit
	LimitListTopPostGap = 3
	//LimitListPostRelateTagTopic limit
	LimitListPostRelateTagTopic = 5
	//LimitListComment limit
	LimitListComment = 5
	//LimitSearchTagTopic limit
	LimitSearchTagTopic = 5
	//LimitListFollowUser limit
	LimitListFollowUser = 6
	//LimitListGapRecommend limit
	LimitListGapRecommend = 15
	//LimitListTopicRecommend limit
	LimitListTopicRecommend = 25

	//LimitSearchTopic limit
	LimitSearchTopic = 10
	//LimitSearchGap limit
	LimitSearchGap = 10

	//LimitTitleNotification limit
	LimitTitleNotification = 10
	//LimitNameNotification limit
	LimitNameNotification = 10
	//LimitNameGapNotification limit
	LimitNameGapNotification = 10
	//LimitTextCommentNotification limit
	LimitTextCommentNotification = 10
	//LimitListNotification limit
	LimitListNotification = 10

	//LimitTitleNotificationType limit
	LimitTitleNotificationType = 20
	//LimitNameNotificationType limit
	LimitNameNotificationType = 20
	//LimitNameGapNotificationType limit
	LimitNameGapNotificationType = 20
	//LimitTextCommentNotificationType limit
	LimitTextCommentNotificationType = 20

	//LimitDurationGuestView limit
	LimitDurationGuestView = 7200

	//LimitListPostCountView limit
	LimitListPostCountView = 5
	//LimitListPostCountViewRevenue limit
	LimitListPostCountViewRevenue = 5
	//LimitTitlePostCountView limit
	LimitTitlePostCountView = 20

	//LimitTitleRevenuePostCountView limit
	LimitTitleRevenuePostCountView = 60

	//DeletePageDuration = 86400 * 15
	DeletePageDuration = 15
	// UpdateEmailDuration ระยะเวลาหมดอายุ email ระยะเวลาการใช้งานของ code ที่จะแก้ไข email
	UpdateEmailDuration = 7

	// HeightDisplay And ขนาด รูป display
	HeightDisplay = 300
	// WidthDisplay  And ขนาด รูป display
	WidthDisplay = 300
	// HeightDisplayMiddle And ขนาด รูป display
	HeightDisplayMiddle = 150
	// WidthDisplayMiddle  And ขนาด รูป display
	WidthDisplayMiddle = 150
	// HeightDisplayMini And ขนาด รูป display
	HeightDisplayMini = 50
	// WidthDisplayMini  And ขนาด รูป display
	WidthDisplayMini = 50
	// HeightTopic And ขนาด รูป display
	HeightTopic = 300
	// WidthTopic  And ขนาด รูป display
	WidthTopic = 300
	// HeightTopicFacebook And ขนาด รูป display
	HeightTopicFacebook = 630
	// WidthTopicFacebook  And ขนาด รูป display
	WidthTopicFacebook = 1200

	//HeightCropCoverMini ขนาด รูป display Mini
	HeightCropCoverMini = 350
	//WidthCropCoverMini ขนาด รูป display Mini
	WidthCropCoverMini = 700

	//HeightResizeCoverMini is resize
	HeightResizeCoverMini = 125
	//WidthResizeCoverMini is resize
	WidthResizeCoverMini = 255

	//HeightCover ขนาด รูป Cover
	HeightCover = 400
	//WidthCover ขนาด รูป Cover
	WidthCover = 1920

	//WidthMainPost ขนาดรูป MainPost
	WidthMainPost = 300
	//HigthMainPost ขนาดรูป MainPost
	HigthMainPost = 517
	//WidthMainPostMobile ขนาดรูป MainPost
	WidthMainPostMobile = 200
	//HigthMainPostMobile ขนาดรูป MainPost
	HigthMainPostMobile = 344

	//WidthPostImage ขนาด รูป Post
	WidthPostImage = 1200

	//WidthImageFacebook ขนาดรูป fb
	WidthImageFacebook = 1200
	//HigthImageFacebook ขนาดรูป fb
	HigthImageFacebook = 628

	//ExpiredNameGapDuration นับเวลาเปลี่ยนชื่อ GAP
	ExpiredNameGapDuration = 5184000

	//MinimumPay ขั้นต่ำ 500 บาท
	MinimumPay = 500

	//TopicOtherID Topic other
	TopicOtherID = "5654313976201216"
	//CategoryOtherID category other
	CategoryOtherID = "5639445604728832"

	//RedisCount ค่าค้นหาสูงสุดของ Redis
	RedisCount              = 1000
	RedisIndexSession       = "index:session"
	RedisIndexUserFirstName = "index:user:firstname"
	RedisIndexUserLastName  = "index:user:lastname"
	RedisIndexGapName       = "index:gap:name"
	RedisIndexTopicName     = "index:topic:name"
	RedisUser               = "user:"
	RedisGap                = "gap:"
	RedisCategory           = "category:verify:"
	RedisCategoryNotVerify  = "category:notverify:"
	RedisTopic              = "topic:verify:"
	RedisTopicNotVerify     = "topic:notverify:"
	RedisTopicAny           = "topic:*:"
	RedisProvince           = "province:"

	ImageProfileM     = "https://storage.googleapis.com/powerwork-bucket/bet/frontback/headuser1.png"
	ImageProfileF     = "https://storage.googleapis.com/powerwork-bucket/bet/frontback/headuser1.png"
	ImageProfileO     = "https://storage.googleapis.com/powerwork-bucket/bet/frontback/headuser1.png"
	ImageProfileMMini = "https://storage.googleapis.com/powerwork-bucket/bet/frontback/headuser1.png"
	ImageProfileFMini = "https://storage.googleapis.com/powerwork-bucket/bet/frontback/headuser1.png"
	ImageProfileOMini = "https://storage.googleapis.com/powerwork-bucket/bet/frontback/headuser1.png"
	ImageCoverM       = "https://storage.googleapis.com/swapgap-stage-bucket/system/gap-image/cover/normal/cover.jpg"
	ImageCoverMMini   = "https://storage.googleapis.com/swapgap-stage-bucket/system/gap-image/cover/mini/cover-mini.png"
	ImageCoverF       = "https://storage.googleapis.com/swapgap-stage-bucket/system/gap-image/cover/normal/cover.jpg"
	ImageCoverFMini   = "https://storage.googleapis.com/swapgap-stage-bucket/system/gap-image/cover/mini/cover-mini.png"
	ImageCoverO       = "https://storage.googleapis.com/swapgap-stage-bucket/system/gap-image/cover/normal/cover.jpg"
	ImageCoverOMini   = "https://storage.googleapis.com/swapgap-stage-bucket/system/gap-image/cover/mini/cover-mini.png"
	ImageGap          = "https://storage.googleapis.com/swapgap-bucket/system/gap-image/avatar_gap.svg"
	ImageGapMini      = "https://storage.googleapis.com/swapgap-bucket/system/gap-image/avatar_gap.svg"
	ImageCoverGap     = "https://storage.googleapis.com/swapgap-bucket/system/gap-image/swg_imagecovergap.jpg"
	ImageCoverGapMini = "https://storage.googleapis.com/swapgap-bucket/system/gap-image/swg_imagecovergap.jpg"
	ImageCategory     = "https://storage.googleapis.com/swapgap-stage-bucket/system/category/normal/cat-300.jpg"
	ImagePostRelate   = "https://storage.googleapis.com/powerwork-bucket/static/post/avatar-relate.svg"

	//TODO c
	GapOfficialID   = "6245188464803840"
	TopicOfficialID = "4698262711828480"
	AdminID         = "aa5a05ac-0139-11ea-a6b0-82c626904a38"
	AdminID2        = "e8e44244-05ec-11ea-a6b1-82c626904a38"

	URLVerifyEmailCallBack   = "/verify/email/callback"
	URLResetPasswordCallBack = "/resetpassword"

	GoogleCallbackURL = "/signin/google/callback"

	//TODO c
	FacebookAppID       = "414711119471060"
	FacebookCallbackURL = "/signin/facebook/callback"

	AMLLuckieNumber = 3
)
