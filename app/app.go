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
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
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

	m.Get(app.Hime.Route("wallet"), hime.Handler(walletGetHandler))
	m.Get(app.Hime.Route("register"), hime.Handler(registerGetHandler))
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
	m.Get(app.Hime.Route("front.back"), isUser(hime.Handler(frontbackBetGetHandler)))
	m.Get(app.Hime.Route("notfound"), hime.Handler(notFoundHandler))
	m.Get(app.Hime.Route("account"), isUser(hime.Handler(userGetHandler)))
	m.Post(app.Hime.Route("account"), isUser(hime.Handler(userPostHandler)))
	m.Get(app.Hime.Route("withdraw.money"), isUser(hime.Handler(UserWithdrawMoneyGetHandler)))
	m.Post(app.Hime.Route("withdraw.money"), isUser(hime.Handler(UserWithdrawMoneyPostHandler)))
	m.Post(app.Hime.Route("ajax.frontback.bet"), isUser(hime.Handler(ajaxFrontbackBetPostHandler)))

	admin := httprouter.New()
	admin.HandleMethodNotAllowed = false
	admin.NotFound = hime.Handler(notFoundHandler)

	admin.Get("/", hime.Handler(adminIndexGetHandler))

	admin.Get("/selectuser", hime.Handler(adminSelectUserGetHandler))
	admin.Get("/addcoin", hime.Handler(adminAddCoinGetHandler))
	admin.Post("/addcoin", hime.Handler(adminAddCoinPostHandler))
	admin.Get("/withdraw/money", hime.Handler(adminWithdrawTest))
	admin.Post("/withdraw/money", hime.Handler(adminWithdrawMoneyPostHandler))
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

	rds.Del(config.RedisIndexUserFirstName)
	rds.Del(config.RedisIndexUserLastName)

	rows, err := postgre.Query(`
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

	log.Println("Init Redis Success!!")
}
