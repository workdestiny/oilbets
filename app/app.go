package app

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"log"
	"net/http"
	"strings"

	"github.com/acoshift/csrf"
	"github.com/acoshift/httprouter"
	"github.com/acoshift/middleware"
	"github.com/acoshift/webstatic"

	"github.com/go-redis/redis"
	"github.com/moonrhythm/hime"
	"github.com/moonrhythm/session"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
)

// Handler is return Handler
func (app *App) Handler() http.Handler {
	initConfig(app)
	// initStatic()
	// initTemplate(app)
	// initRoutes(app)

	mux := http.NewServeMux()
	mux.Handle("/healthz", http.HandlerFunc(healthzHandler))
	mux.Handle("/-/", http.StripPrefix("/-", webstatic.New(webstatic.Config{
		Dir:          "public",
		CacheControl: "public, max-age=31536000",
	})))

	mux.Handle("/favicon.ico", fileHandler("public/images/favicon.ico"))
	mux.Handle("/robots.txt", fileHandler("public/robots.txt"))
	mux.Handle("/sitemap.xml", fileHandler("public/gsc/sitemap.xml"))
	mux.Handle("/f6004f04614a71fe71a7653a689c48f4.html", fileHandler("public/f6004f04614a71fe71a7653a689c48f4.html"))

	m := httprouter.New()
	m.HandleMethodNotAllowed = false
	m.NotFound = hime.Handler(notFoundHandler)

	m.Get(app.Hime.Route("privacy"), hime.Handler(privacyGetHandler))
	m.Get(app.Hime.Route("condition"), hime.Handler(conditionGetHandler))
	m.Get(app.Hime.Route("contact"), hime.Handler(contactGetHandler))
	m.Get(app.Hime.Route("signin"), isAnonymous(hime.Handler(signInGetHandler)))
	m.Get(app.Hime.Route("signin.email"), isAnonymous(hime.Handler(signInEmailGetHandler)))
	m.Post(app.Hime.Route("signin.email"), isAnonymous(hime.Handler(signInPostHandler)))
	m.Get(app.Hime.Route("signin.facebook"), isAnonymous(hime.Handler(signInFacebookGetHandler)))
	m.Get(app.Hime.Route("signin.facebook.callback"), isAnonymous(hime.Handler(signInFacebookCallbackGetHandler)))
	m.Get(app.Hime.Route("signin.google"), isAnonymous(hime.Handler(signInGoogleGetHandler)))
	m.Get(app.Hime.Route("signin.google.callback"), isAnonymous(hime.Handler(signInGoogleCallbackGetHandler)))
	m.Get(app.Hime.Route("verify.email.callback"), hime.Handler(verifyEmailGetHandler))
	m.Get(app.Hime.Route("signup"), isAnonymous(hime.Handler(signUpGetHandler)))
	m.Post(app.Hime.Route("signup"), isAnonymous(hime.Handler(signupPostHandler)))
	m.Post(app.Hime.Route("signout"), isUser(hime.Handler(signoutPostHandler)))
	m.Get(app.Hime.Route("forgot"), isAnonymous(hime.Handler(forgotGetHandler)))
	m.Post(app.Hime.Route("forgot"), isAnonymous(hime.Handler(forgetPostHandler)))
	m.Get(app.Hime.Route("resetpassword"), isAnonymous(hime.Handler(resetpasswordGetHandler)))
	m.Post(app.Hime.Route("resetpassword"), isAnonymous(hime.Handler(resetPasswordPostHandler)))
	m.Get(app.Hime.Route("index"), isPublic(hime.Handler(discoverGetHandler)))
	m.Get(app.Hime.Route("discover"), hime.Handler(discoverGetHandler))
	m.Get(app.Hime.Route("follow"), isUser(hime.Handler(followGetHandler)))
	m.Get(app.Hime.Route("liked"), isUser(hime.Handler(likedGetHandler)))
	m.Get(app.Hime.Route("search"), hime.Handler(searchGetHandler))
	// m.Get(app.Hime.Route("public"), isUser(hime.Handler(publicGetHandler)))
	m.Get(app.Hime.Route("create.post"), isUser(hime.Handler(createPostGetHandler)))
	m.Post(app.Hime.Route("create.post"), isUser(hime.Handler(createPostPostHandler)))
	m.Post(app.Hime.Route("create.post.undraft"), isUser(hime.Handler(undraftPostHandler)))
	m.Get(app.Hime.Route("create.gap"), isUser(hime.Handler(createGapGetHandler)))
	m.Post(app.Hime.Route("create.gap"), isUser(hime.Handler(createGapPostHandler)))
	m.Get(app.Hime.Route("edit.post", ":postID"), isUser(hime.Handler(editPostGetHandler)))
	m.Post(app.Hime.Route("edit.post", ":postID"), isUser(hime.Handler(editPostPostHandler)))
	m.Post(app.Hime.Route("delete.post"), isUser(hime.Handler(deletePostPostHandler)))
	m.Get(app.Hime.Route("notfound"), hime.Handler(notFoundHandler))
	m.Get(app.Hime.Route("category"), isUser(hime.Handler(categoryGetHandler)))
	m.Get(app.Hime.Route("topic", ":topic"), hime.Handler(topicGetHandler))
	m.Get(app.Hime.Route("category.get", ":category"), hime.Handler(getCategoryGetHandler))
	m.Post(app.Hime.Route("topic.select"), isUser(hime.Handler(followTopicListPostHandler)))
	m.Get(app.Hime.Route("post.read"), hime.Handler(func(ctx *hime.Context) error {
		return ctx.Redirect("/")
	}))
	m.Get(app.Hime.Route("post.read", ":postID"), hime.Handler(postReadGetHandler))
	m.Get("/gap/:gapID", hime.Handler(gapGetHandler))
	m.Get("/gap/:gapID/setting", isUser(hime.Handler(gapSettingGetHandler)))
	m.Get("/gap/:gapID/insights", isUser(hime.Handler(gapInsightsGetHandler)))
	m.Get("/gap/:gapID/revenue", isBookbank(isUser(hime.Handler(gapRevenueGetHandler))))
	m.Post(app.Hime.Route("gap.setting.info"), isUser(hime.Handler(gapSettingInfoPostHandler)))
	m.Post(app.Hime.Route("gap.setting.username"), isUser(hime.Handler(gapSettingUserNamePostHandler)))
	m.Post(app.Hime.Route("gap.setting.contact"), isUser(hime.Handler(gapSettingContactPostHandler)))
	m.Post(app.Hime.Route("gap.setting.address"), isUser(hime.Handler(gapSettingAddressPostHandler)))
	m.Get(app.Hime.Route("account"), isUser(hime.Handler(userGetHandler)))
	m.Post(app.Hime.Route("account"), isUser(hime.Handler(userPostHandler)))
	m.Get(app.Hime.Route("notification"), isUser(hime.Handler(notificationGetHandler)))
	m.Get(app.Hime.Route("help.post"), hime.Handler(helpPostGetHandler))

	m.Post(app.Hime.Route("ajax.gap.follow"), isUser(hime.Handler(ajaxFollowGapPostHandler)))
	m.Post(app.Hime.Route("ajax.gap.list.user.follower"), hime.Handler(ajaxListUserFollowerGapPostHandler))
	m.Post(app.Hime.Route("ajax.profile.upload.display"), isUser(hime.Handler(ajaxUploadProfileDisplayPostHandler)))
	m.Post(app.Hime.Route("ajax.profile.verify.creator"), isUser(hime.Handler(ajaxVerifyToCreatorPostHandler)))
	m.Post(app.Hime.Route("ajax.profile.verify.email"), isUser(hime.Handler(ajaxEmailVerify)))
	m.Post(app.Hime.Route("ajax.profile.verify.bookbank"), isUser(hime.Handler(ajaxVerifyBookbankPostHandler)))
	m.Post(app.Hime.Route("ajax.gap.upload.display", ":gapID"), isUser(hime.Handler(ajaxUploadGapDisplayPostHandler)))
	m.Post(app.Hime.Route("ajax.gap.upload.cover", ":gapID"), isUser(hime.Handler(ajaxUploadGapCoverPostHandler)))
	m.Post(app.Hime.Route("ajax.topic.follow"), isUser(hime.Handler(ajaxFollowTopicPostHandler)))
	m.Post(app.Hime.Route("ajax.search.tagtopic"), isUser(hime.Handler(ajaxSearchTagTopicPostHandler)))
	m.Post(app.Hime.Route("ajax.search.navigationbar"), hime.Handler(ajaxNavigationBarSearchHandler))
	m.Post(app.Hime.Route("ajax.post.like"), isUser(hime.Handler(ajaxLikePostHandler)))
	m.Post(app.Hime.Route("ajax.post.image"), isUser(hime.Handler(ajaxDraftImagePostHandler)))
	m.Post(app.Hime.Route("ajax.post.draft.article"), isUser(hime.Handler(ajaxDraftPostArticleHandler)))
	m.Post(app.Hime.Route("ajax.post.discover"), hime.Handler(discoverPostHandler))
	m.Post(app.Hime.Route("ajax.post.liked"), hime.Handler(ajaxLikedPostHandler))
	m.Post(app.Hime.Route("ajax.post.follow"), isUser(hime.Handler(ajaxFollowPostHandler)))
	//m.Post(app.Hime.Route("ajax.post.public"), isUser(hime.Handler(ajaxPublicPostHandler)))
	m.Post(app.Hime.Route("ajax.post.read", ":postID"), hime.Handler(readPostHandler))
	m.Post(app.Hime.Route("ajax.post.topic", ":topic"), hime.Handler(ajaxTopicPostHandler))
	m.Post(app.Hime.Route("ajax.post.category", ":category"), hime.Handler(ajaxCategoryPostHandler))
	m.Post(app.Hime.Route("ajax.post.gap", ":gapID"), hime.Handler(ajaxGapPostHandler))
	m.Post(app.Hime.Route("ajax.post.shortener.url"), hime.Handler(ajaxShortenerURLPostHandler))
	m.Post(app.Hime.Route("ajax.user.list.gap.follow"), isUser(hime.Handler(ajaxUserFollowGapPostHandler)))
	m.Post(app.Hime.Route("ajax.comment.list"), hime.Handler(ajaxCommentNextLoadPostHandler))
	m.Post(app.Hime.Route("ajax.comment.post"), isUser(hime.Handler(ajaxCommentPostHandler)))
	m.Post(app.Hime.Route("ajax.comment.delete"), isUser(hime.Handler(ajaxCommentDeletePost)))
	m.Post(app.Hime.Route("ajax.comment.edit"), isUser(hime.Handler(ajaxCommentEditPostHandler)))
	m.Post(app.Hime.Route("ajax.notification.list"), isUser(hime.Handler(ajaxListNotificationPostHandler)))
	m.Post(app.Hime.Route("ajax.notification.list.type"), isUser(hime.Handler(ajaxListNotificationTypePostHandler)))
	m.Post(app.Hime.Route("ajax.notification.read"), isUser(hime.Handler(ajaxReadNotificationPostHandler)))
	m.Post(app.Hime.Route("ajax.notification.read.all"), isUser(hime.Handler(ajaxReadAllNotificationPostHandler)))
	m.Post(app.Hime.Route("ajax.notification.reset"), isUser(hime.Handler(ajaxResetNotificationPostHandler)))
	m.Post(app.Hime.Route("ajax.gap.statistic.post"), isUser(hime.Handler(ajaxListPostCountViewHandler)))
	m.Post(app.Hime.Route("ajax.gap.revenue.post"), isBookbank(isUser(hime.Handler(ajaxListPostCountViewRevenueHandler))))
	m.Post(app.Hime.Route("ajax.gap.revenue.register"), isBookbank(isUser(hime.Handler(ajaxRevenuePostHandler))))
	m.Post(app.Hime.Route("ajax.topic.list"), isUser(hime.Handler(ajaxListTopicPostHandler)))

	admin := httprouter.New()
	admin.HandleMethodNotAllowed = false
	admin.NotFound = hime.Handler(notFoundHandler)

	admin.Get("/", hime.Handler(adminIndexGetHandler))
	admin.Get("/verify", hime.Handler(adminVerifyGetHandler))
	admin.Post("/verify", hime.Handler(adminVerifyPostHandler))
	admin.Get("/user", hime.Handler(adminVerifyUserHandler))
	admin.Post("/user", hime.Handler(adminVerifyUserPostHandler))
	admin.Get("/verify/detail", hime.Handler(adminVerifyDetailGetHandler))
	admin.Post("/idcard/verify", hime.Handler(adminVerifyIDCardHandler))
	admin.Post("/bookbank/verify", hime.Handler(adminVerifyBookbankHandler))
	admin.Get("/revenue", hime.Handler(adminRevenueGetHandler))
	admin.Post("/revenue", hime.Handler(adminRevenuePostHandler))
	admin.Get("/revenue/detail", hime.Handler(adminRevenueDetailGetHandler))
	admin.Get("/revenue/history", hime.Handler(adminRevenueHistoryGetHandler))
	admin.Get("/gap/recommend", hime.Handler(adminGapRecommendGetHandler))
	admin.Post("/gap/recommend", hime.Handler(adminGapRecommendPostHandler))
	admin.Post("/gap/recommend/register", hime.Handler(adminAddGapRecommendPostHandler))
	admin.Post("/gap/recommend/delete", hime.Handler(adminDeleteGapRecommendPostHandler))

	admin.Get("/category", hime.Handler(adminCategoryGetHandler))
	admin.Get("/category/create", hime.Handler(adminCategoryCreateGetHandler))
	admin.Post("/category/create", hime.Handler(adminCategoryCreatePostHandler))
	admin.Get("/category/edit", hime.Handler(adminCategoryEditGetHandler))
	admin.Post("/category/edit", hime.Handler(adminCategoryEditPostHandler))

	admin.Get("/topic", hime.Handler(adminTopicGetHandler))
	admin.Post("/topic", hime.Handler(adminTopicPostHandler))
	admin.Get("/topic/create", hime.Handler(adminTopicCreateGetHandler))
	admin.Post("/topic/create", hime.Handler(adminTopicCreatePostHandler))
	admin.Get("/topic/edit", hime.Handler(adminTopicEditGetHandler))
	admin.Post("/topic/edit", hime.Handler(adminTopicEditPostHandler))
	admin.Get("/topic/seo", hime.Handler(adminTopicSEOGetHandler))
	admin.Get("/topic/seo/edit", hime.Handler(adminTopicSEOEditGetHandler))
	admin.Post("/topic/seo/edit", hime.Handler(adminTopicSEOEditPostHandler))

	admin.Post("/ajax/revenue/post", hime.Handler(ajaxAdminListPostCountViewRevenueHandler))
	admin.Post("/ajax/revenue/approve", hime.Handler(ajaxAdminApproveRevenuePostHandler))
	admin.Post("/ajax/revenue/reject", hime.Handler(ajaxAdminRejectRevenuePostHandler))
	admin.Post("/ajax/post/reject", hime.Handler(ajaxAdminRejectPostStatusRevenuePostHandler))
	admin.Post("/ajax/post/delete", hime.Handler(ajaxAdminDeletePostPostHandler))

	mux.Handle("/", m)
	mux.Handle("/admin/", http.StripPrefix("/admin", isAdmin(admin)))

	return middleware.Chain(
		DefaultCacheControl,
		logHTTP,
		noCORS,
		securityHeaders,
		methodFilter,
		csrf.New(app.CSRFConfig),
		session.Middleware(app.SessionConfig),
		getCookie,
		fetchUser(db),
		userAgent,
		panicRecovery,
	)(mux)

}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	if err := db.Ping(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := myRedis.Do("PING").Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func notFoundHandler(ctx *hime.Context) error {
	return ctx.RedirectTo("discover")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// RunRedisX run
func RunRedisX(rds *redis.Client, postgre *sql.DB) {

	rds.Del(config.RedisIndexTopicName)
	rds.Del(config.RedisIndexUserFirstName)
	rds.Del(config.RedisIndexUserLastName)
	rds.Del(config.RedisIndexGapName)

	rows, err := postgre.Query(`
		SELECT id, cat_id, code, name->>'th', images->>'mini',
		       images->>'normal', count, used_count, verify
		  FROM public.topic
	`)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()

	for rows.Next() {

		var verify bool
		var rt entity.RedisTopicModel
		err := rows.Scan(&rt.ID, &rt.CatID, &rt.Code, &rt.Name, &rt.Images.Mini, &rt.Images.Normal, &rt.Count, &rt.UsedCount, &verify)
		must(err)

		buf := bytes.Buffer{}
		gob.NewEncoder(&buf).Encode(&rt)

		if verify {
			rds.Set(config.RedisTopic+rt.CatID+":"+rt.ID+":"+rt.Code, buf.Bytes(), 0)
			//c.Do("SET", config.RedisTopic+rt.CatID+":"+rt.ID+":"+rt.Code, buf.Bytes())
		}

		if !verify {
			rds.Set(config.RedisTopicNotVerify+rt.CatID+":"+rt.ID+":"+rt.Code, buf.Bytes(), 0)
			//c.Do("SET", config.RedisTopicNotVerify+rt.CatID+":"+rt.ID+":"+rt.Code, buf.Bytes())
		}

		rds.SAdd(config.RedisIndexTopicName, rt.ID+":"+strings.ToLower(rt.Name))
		//c.Do("SADD", config.RedisIndexTopicName, rt.ID+":"+strings.ToLower(rt.Name))

	}

	rows, err = postgre.Query(`
		SELECT users.id, users.username->>'text', users.firstname, users.lastname,
			   users.display->>'mini', users.display->>'middle', COALESCE(user_kycs.is_email, 'false'), COALESCE(user_kycs.is_verify_email, 'false'),
			   COALESCE(user_kycs.is_idcard, 'false'), COALESCE(user_kycs.is_bookbank, 'false')
		  FROM users
	 LEFT JOIN user_kycs
	 		ON users.id = user_kycs.user_id
	`)
	if err != nil {

		log.Println(err)
	}

	for rows.Next() {
		var isEmail, isVerifyEmail, isIDCard, isBookBank bool
		var ru entity.RedisUserModel
		err := rows.Scan(&ru.ID, &ru.Username, &ru.FirstName, &ru.LastName,
			&ru.DisplayImageMini, &ru.DisplayImage, &isEmail, &isVerifyEmail,
			&isIDCard, &isBookBank)
		must(err)

		var level entity.UserLevelType
		if isEmail {
			level = entity.NotVerify
		}
		if isVerifyEmail {
			level = entity.VerifyEmail
		}
		if isIDCard {
			level = entity.VerifyIDCard
		}
		if isBookBank {
			level = entity.VerifyBookBank
		}

		ru.Level = level

		buf := bytes.Buffer{}
		gob.NewEncoder(&buf).Encode(&ru)

		rds.Set(config.RedisUser+ru.ID, buf.Bytes(), 0)
		//c.Do("SET", config.RedisUser+ru.ID, buf.Bytes())
		rds.SAdd(config.RedisIndexUserFirstName, ru.ID+":"+strings.ToLower(ru.FirstName))
		//c.Do("SADD", config.RedisIndexUserFirstName, ru.ID+":"+strings.ToLower(ru.FirstName))
		rds.SAdd(config.RedisIndexUserLastName, ru.ID+":"+strings.ToLower(ru.LastName))
		//c.Do("SADD", config.RedisIndexUserLastName, ru.ID+":"+strings.ToLower(ru.LastName))

	}

	rows, err = postgre.Query(`
		SELECT id, username->>'text', name->>'text', user_id,
		       display->>'mini', display->>'middle', count->>'follower', count->>'popular'
		  FROM public.gap
	`)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {

		var rg entity.RedisGapModel
		err := rows.Scan(&rg.ID, &rg.Username, &rg.Name, &rg.UserID, &rg.DisplayImageMini, &rg.DisplayImage, &rg.CountFollower, &rg.CountPopular)
		must(err)

		buf := bytes.Buffer{}
		gob.NewEncoder(&buf).Encode(&rg)

		rds.Set(config.RedisGap+rg.UserID+":"+rg.ID, buf.Bytes(), 0)
		//c.Do("SET", config.RedisGap+rg.UserID+":"+rg.ID, buf.Bytes())
		rds.SAdd(config.RedisIndexGapName, rg.ID+":"+strings.ToLower(rg.Name))
		//c.Do("SADD", config.RedisIndexGapName, rg.ID+":"+strings.ToLower(rg.Name))

	}

	log.Println("Init Redis Success!!")
}
