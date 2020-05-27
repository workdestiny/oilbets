package app

import (
	"database/sql"
	"encoding/json"
	"image"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/acoshift/pgsql"
	"github.com/asaskevich/govalidator"
	"github.com/moonrhythm/hime"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
)

func signInGetHandler(ctx *hime.Context) error {
	referrer := ctx.Request.Referer()
	if strings.Contains(referrer, domain) {
		SaveSessionReferrer(ctx, referrer)
	}

	return ctx.View("app/signin", page(ctx))
}

func signInPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()
	email := strings.ToLower(ctx.PostFormValue("email"))
	password := ctx.PostFormValue("password")

	if !govalidator.IsEmail(email) {
		f.Add("Errors", "อีเมลไม่ถูกรูปแบบ")
		return ctx.RedirectToGet()
	}

	id, pw, err := repository.Signin(db, email)
	if err == sql.ErrNoRows {
		f.Add("Errors", "อีเมล หรือรหัสผ่านไม่ถูกต้อง")
		return ctx.RedirectToGet()
	}
	must(err)

	if !repository.ComparePassword(pw, password) {
		f.Add("Errors", "อีเมล หรือรหัสผ่านไม่ถูกต้อง")
	}

	if f.Has("Errors") {
		return ctx.RedirectToGet()
	}

	//save session
	SaveSession(ctx, id)

	return ctx.RedirectTo("discover")
}

func signInEmailGetHandler(ctx *hime.Context) error {
	return ctx.View("app/signin-email", page(ctx))
}

func signUpGetHandler(ctx *hime.Context) error {
	return ctx.View("app/signup", page(ctx))
}

func forgotGetHandler(ctx *hime.Context) error {
	return ctx.View("app/forgotpassword", page(ctx))
}

func resetpasswordGetHandler(ctx *hime.Context) error {

	email := strings.ToLower(ctx.FormValue("email"))
	code := ctx.FormValue("code")

	if !govalidator.IsEmail(email) {
		return ctx.RedirectTo("notfound")
	}

	if code == "" {
		return ctx.RedirectTo("notfound")
	}

	return ctx.View("app/resetpassword", page(ctx))
}

func resetPasswordPostHandler(ctx *hime.Context) error {

	email := strings.ToLower(ctx.FormValue("email"))
	code := ctx.FormValue("code")
	password := ctx.PostFormValue("password")
	repeatPassword := ctx.PostFormValue("repeat-password")

	f := getSession(ctx).Flash()

	if !govalidator.IsEmail(email) {
		f.Add("Errors", "อีเมลไม่ถูกรูปแบบ")
	}

	if utf8.RuneCountInString(password) < 8 || utf8.RuneCountInString(password) > 20 {
		f.Add("Errors", "รหัสผ่านต้องไม่น้อย 8 และไม่เกิน 20 ตัวอักษร")
	}

	if password != repeatPassword {
		f.Add("Errors", "รหัสผ่านไม่ตรงกัน")
	}

	hashPassword, err := repository.HashPassword(password)
	if err != nil {
		f.Add("Errors", "รหัสผ่านไม่สามารถเข้ารหัสได้ กรุณาลองใหม่อีกครั้ง")
	}

	if f.Has("Errors") {
		return ctx.RedirectToGet()
	}

	err = repository.SetResetPassword(db, email, code, hashPassword)
	must(err)

	removeSession(ctx)
	addSuccess(f, "เปลี่ยนรหัสผ่านเรียบร้อยแล้ว")
	return ctx.RedirectToGet()
}

func verifyEmailGetHandler(ctx *hime.Context) error {

	me := getUser(ctx)
	email := strings.ToLower(ctx.FormValue("email"))
	code := ctx.FormValue("code")

	if !govalidator.IsEmail(email) {
		return ctx.RedirectTo("notfound")
	}

	id, err := repository.CheckCodeEmailVerify(db, email, code)
	if err == sql.ErrNoRows {
		return ctx.RedirectTo("notfound")
	}
	must(err)

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.VerifyEmail(tx, id)
		if err != nil {
			return err
		}

		if me.GetLevel() == entity.NoEmail && !me.IsVerifyIDCard {

			err = repository.UpdatePostVerify(tx, id, 1)
			if err != nil {
				return err
			}
		}

		return nil
	})
	must(err)

	return ctx.RedirectTo("discover")
}

func ajaxEmailVerify(ctx *hime.Context) error {

	me := getUser(ctx)

	if me.IsVerify {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	var req entity.RequestEmailVerify
	err := bindJSON(ctx.Request.Body, &req)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(nil)
	}

	if me.Email == "" {

		res := entity.ResponseError{}
		if !govalidator.IsEmail(req.Email) {
			res.Message = "อีเมลไม่ถูกรูปแบบ"
			res.Errors = "email"
			return ctx.Status(http.StatusOK).JSON(&res)
		}

		if utf8.RuneCountInString(req.Password) < 8 || utf8.RuneCountInString(req.Password) > 20 {
			res.Message = "รหัสผ่านต้องไม่น้อยกว่า 8 ตัวอักษรและไม่เกิน 20 ตัวอักษร"
			res.Errors = "password"
			return ctx.Status(http.StatusOK).JSON(&res)
		}

		if req.Password != req.RepeatPassword {
			res.Message = "รหัสผ่านไม่ตรงกัน"
			res.Errors = "password"
			return ctx.Status(http.StatusOK).JSON(&res)
		}

		hash, err := repository.HashPassword(req.Password)
		if err != nil {
			res.Message = "รหัสผ่านไม่สามารถเข้ารหัสได้กรุณาใช้รหัสผ่านอื่น"
			res.Errors = "password"
			return ctx.Status(http.StatusOK).JSON(&res)
		}

		err = repository.CheckEmail(db, req.Email)
		if err != nil {
			res.Message = "อีเมลนี้มีในระบบแล้ว"
			res.Errors = "email"
			return ctx.Status(http.StatusOK).JSON(&res)
		}

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.UpdateNewEmail(tx, me.ID, req.Email)
			if err != nil {
				return err
			}

			err = repository.UpdateNewPassword(tx, me.ID, hash)
			if err != nil {
				return err
			}

			return nil
		})
		must(err)

		me.Email = req.Email
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err := repository.UpdateCodeSendEmailVerify(tx, baseURL, me.ID, me.Email, me.FirstName+" "+me.LastName)
		if err != nil {
			return err
		}

		return nil
	})
	must(err)

	return ctx.NoContent()
}

func forgetPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()

	email := strings.ToLower(ctx.PostFormValueTrimSpace("email"))

	if !govalidator.IsEmail(email) {
		f.Add("Errors", "อีเมลไม่ถูกรูปแบบ")
	}

	id, t, count, err := repository.GetForgetPassword(db, email)
	if err != nil {
		f.Add("Errors", "ไม่มีอีเมลนี้ในระบบ")
	}

	ct := t
	now := time.Now()

	if now.Unix() < ct.Add(15*time.Minute).Unix() {
		f.Add("Errors", "กรุณาลองใหม่หลัง 15 นาที")
	}

	if f.Has("Errors") {
		return ctx.RedirectToGet()
	}

	ranCode := repository.RandStr(40)

	if now.Year() > ct.Year() || now.Month() > ct.Month() || now.Day() > ct.Day() {

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			err = repository.SetCodeResetPassword(tx, id, now.UTC(), 1, ranCode)
			if err != nil {
				f.Add("Errors", "ไม่สามารถส่งคำขอลืมรหัสผ่านได้")
				return err
			}

			return nil
		})

		if err == nil {
			go repository.SendEmailForgetPassword(baseURL, email, ranCode)
			addSuccess(f, "ระบบได้ทำการส่งคำขอไปยังอีเมลของท่านแล้ว กรุณาตรวจสอบอีเมล")
		}

		return ctx.RedirectToGet()
	}

	if count >= 3 {
		f.Add("Errors", "กรุณาลองใหม่หลัง 24 ชั่วโมง")
		return ctx.RedirectToGet()
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		err = repository.SetCodeResetPassword(tx, id, now.UTC(), count+1, ranCode)
		if err != nil {
			f.Add("Errors", "ไม่สามารถส่งคำขอลืมรหัสผ่านได้")
			return err
		}

		return nil
	})

	go repository.SendEmailForgetPassword(baseURL, email, ranCode)
	addSuccess(f, "ระบบได้ทำการส่งคำขอไปยังอีเมลของท่านแล้ว กรุณาตรวจสอบอีเมล")
	return ctx.RedirectToGet()

}

func signInGoogleGetHandler(ctx *hime.Context) error {

	gothic.Store = sessions.NewCookieStore(googleCookieSecret)
	goth.UseProviders(
		gplus.New(
			googleClient,
			googleSecret,
			baseURL+config.GoogleCallbackURL,
		),
	)

	q := ctx.Request.URL.Query()
	q.Set("provider", "gplus")
	ctx.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(ctx.ResponseWriter(), ctx.Request)

	return ctx.Redirect("notfound")
}

func signInGoogleCallbackGetHandler(ctx *hime.Context) error {

	q := ctx.Request.URL.Query()
	q.Set("provider", "gplus")
	ctx.Request.URL.RawQuery = q.Encode()
	u, err := gothic.CompleteUserAuth(ctx.ResponseWriter(), ctx.Request)
	if err != nil {
		return ctx.RedirectTo("signin")
	}

	id, err := repository.CheckGoogleID(db, u.UserID)
	if err == sql.ErrNoRows {
		// create user

		resp, err := http.Get(u.AvatarURL)
		if err != nil {
			return ctx.RedirectTo("signin")
		}
		defer resp.Body.Close()

		image, _, err := image.Decode(resp.Body)

		m := resizeDisplayImage(image)
		displayname := generateDisplayImageName(u.UserID)
		upload(ctx, m, displayname)

		m = resizeDisplayMiniImage(image)
		displaymininame := generateDisplayMiniImageName(u.UserID)
		upload(ctx, m, displaymininame)

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			id := repository.CreateSocialUserProvicer(tx, &repository.CreateProvider{
				Email:    "",
				GoogleID: u.UserID,
			})

			err = repository.CreateSocialUser(tx, myRedis, &repository.CreateUser{
				ID:        id,
				Email:     "",
				FirstName: u.FirstName,
				LastName:  u.LastName,
				Display: repository.Display{
					Mini:   generateDownloadURL(displaymininame),
					Middle: generateDownloadURL(displayname),
					Normal: generateDownloadURL(displayname),
				},
			})
			if err != nil {
				return err
			}

			err = repository.CreateFollowGapXOfficial(tx, id)
			if err != nil {
				return err
			}

			err = repository.CreateFollowTopicXOfficial(tx, id)
			if err != nil {
				return err
			}

			SaveSession(ctx, id)
			return nil
		})
		must(err)
		//save session
		return ctx.RedirectTo("discover")

	}

	must(err)

	//save session
	SaveSession(ctx, id)

	return ctx.RedirectTo("discover")
}

func signInFacebookGetHandler(ctx *hime.Context) error {
	return ctx.Redirect("https://www.facebook.com/v2.9/dialog/oauth",
		ctx.Param("client_id", config.FacebookAppID), ctx.Param("redirect_uri", baseURL+config.FacebookCallbackURL))
}

func signInFacebookCallbackGetHandler(ctx *hime.Context) error {

	facebookOauth2 := oauth2.Config{
		ClientID:     config.FacebookAppID,
		ClientSecret: facebookSecret,
		RedirectURL:  baseURL + config.FacebookCallbackURL,
		Scopes:       []string{"public_profile", "email"},
		Endpoint:     facebook.Endpoint,
	}

	code := ctx.Request.URL.Query().Get("code")

	//https: //graph.facebook.com/v2.9/oauth/access_token?client_id=appid&redirect_uri=link&client_secret&code=code
	tokenFacebook, err := facebookOauth2.Exchange(ctx, code)
	if err != nil {
		return ctx.RedirectTo("signin")
	}

	//https: //graph.facebook.com/v2.9/me?fields=id,name,email,picture.type(large)&access_token=
	resp, err := http.Get("https://graph.facebook.com/v2.9/me?fields=id,name,email,picture.type(large)&access_token=" + tokenFacebook.AccessToken)
	if err != nil {
		return ctx.RedirectTo("signin")
	}
	defer resp.Body.Close()

	var facebookData entity.FacebookOauth2
	err = json.NewDecoder(resp.Body).Decode(&facebookData)
	if err != nil {
		return ctx.RedirectTo("signin")
	}

	id, err := repository.CheckFacebookID(db, facebookData.ID)
	if err == sql.ErrNoRows {
		// create user
		resp, err = http.Get(facebookData.Picture.Data.URL)
		if err != nil {
			return ctx.RedirectTo("signin")
		}
		defer resp.Body.Close()

		var FirstNameFacebook string
		var LastNameFacebook string

		image, _, err := image.Decode(resp.Body)
		values := strings.Split(facebookData.Name, " ")
		if len(values) > 1 {
			LastNameFacebook = values[1]
		}
		FirstNameFacebook = values[0]

		m := resizeDisplayImage(image)
		displayname := generateDisplayImageName(facebookData.ID)
		upload(ctx, m, displayname)

		m = resizeDisplayMiniImage(image)
		displaymininame := generateDisplayMiniImageName(facebookData.ID)
		upload(ctx, m, displaymininame)

		err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			id := repository.CreateSocialUserProvicer(tx, &repository.CreateProvider{
				Email:      "",
				FacebookID: facebookData.ID,
			})

			err = repository.CreateSocialUser(tx, myRedis, &repository.CreateUser{
				ID:        id,
				Email:     "",
				FirstName: FirstNameFacebook,
				LastName:  LastNameFacebook,
				Display: repository.Display{
					Mini:   generateDownloadURL(displaymininame),
					Middle: generateDownloadURL(displayname),
					Normal: generateDownloadURL(displayname),
				},
			})
			if err != nil {
				log.Println("1")
				return err
			}

			err = repository.CreateFollowGapXOfficial(tx, id)
			if err != nil {
				log.Println("11")
				return err
			}

			err = repository.CreateFollowTopicXOfficial(tx, id)
			if err != nil {
				log.Println("122")
				return err
			}

			SaveSession(ctx, id)
			return nil
		})
		must(err)
		//save session
		return ctx.RedirectTo("discover")

	}

	must(err)

	//save session
	SaveSession(ctx, id)

	return ctx.RedirectTo("discover")

}

func signupPostHandler(ctx *hime.Context) error {
	f := getSession(ctx).Flash()

	firstName := ctx.PostFormValueTrimSpace("firstname")
	lastName := ctx.PostFormValueTrimSpace("lastname")
	email := strings.ToLower(ctx.PostFormValueTrimSpace("email"))
	password := ctx.PostFormValue("password")
	repeatPassword := ctx.PostFormValue("repeat-password")

	if utf8.RuneCountInString(firstName) > 50 {
		f.Add("ErrName", "ชื่อต้องไม่เกิน 50 ตัวอักษร")
	}

	if !checkOnlyText(firstName) {
		f.Add("ErrName", "ไม่อนุญาตอักษรพิเศษใน: ชื่อ")
	}

	if !checkReservedWordsName(firstName) {
		f.Add("ErrName", "ไม่อนุญาตคำเฉพาะใน: ชื่อ")
	}

	if utf8.RuneCountInString(lastName) > 50 {
		f.Add("ErrLName", "นามสกุลต้องไม่เกิน 50 ตัวอักษร")
	}

	if !checkOnlyText(lastName) {
		f.Add("ErrLName", "ไม่อนุญาตอักษรพิเศษใน: นามสกุล")
	}

	if !checkReservedWordsName(lastName) {
		f.Add("ErrLName", "ไม่อนุญาตคำเฉพาะใน: นามสกุล")
	}

	if utf8.RuneCountInString(email) == 10 {
		f.Add("ErrEmail", "เบอร์โทรศัพท์ไม่ถูกรูปแบบ")
	}

	if utf8.RuneCountInString(password) < 8 || utf8.RuneCountInString(password) > 20 {
		f.Add("ErrPassword", "รหัสผ่านต้องไม่น้อย 8 และไม่เกิน 20 ตัวอักษร")
	}

	if password != repeatPassword {
		f.Add("ErrPassword", "รหัสผ่านไม่ตรงกัน")
	}

	hasPassword, err := repository.HashPassword(password)
	if err != nil {
		f.Add("ErrPassword", "ไม่สามารถใช้รหัสนี้ได้")
	}

	err = repository.CheckEmail(db, email)
	if err != nil {
		f.Add("ErrEmail", "มีอีเมลนี้ในระบบแล้ว")
	}

	if f.Has("ErrName") || f.Has("ErrLName") || f.Has("ErrEmail") || f.Has("ErrPassword") {
		f.Set("Email", email)
		f.Set("FirstName", firstName)
		f.Set("LastName", lastName)
		return ctx.RedirectToGet()
	}

	normal, mini := repository.GetImageProfileURL("other")

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		id, _, err := repository.Register(tx, myRedis, &repository.CreateUser{
			Email:     email,
			Password:  hasPassword,
			FirstName: firstName,
			LastName:  lastName,
			Display: repository.Display{
				Normal: normal,
				Middle: normal,
				Mini:   mini,
			},
		}, GetUserAgent(ctx))
		if err != nil {
			return err
		}

		// //save session
		SaveSession(ctx, id)

		return nil
	})
	must(err)

	return ctx.RedirectTo("discover")
}

func signoutPostHandler(ctx *hime.Context) error {

	removeSession(ctx)
	return ctx.RedirectTo("discover")
}

func checkOnlyText(text string) bool {
	Re := regexp.MustCompile(`^[a-zA-Zก-๙0-9]+$`)
	return Re.MatchString(text)
}

func checkUsername(username string) bool {
	Re := regexp.MustCompile(`^[a-z][0-9a-z]*$`)
	return Re.MatchString(username)
}

func checkReservedWordsUsername(s string) bool {

	s = strings.ToLower(s)
	var ReservedWords = []string{
		"admin",
		"supports",
		"official",
		"officials"}

	for i := 0; i < len(ReservedWords); i++ {
		if s == ReservedWords[i] {
			return false
		}
	}

	return true
}

func checkReservedWordsName(s string) bool {

	s = strings.Replace(s, " ", "", -1)
	s = strings.ToLower(s)
	var ReservedWords = []string{
		"admin",
	}

	for i := 0; i < len(ReservedWords); i++ {
		if s == ReservedWords[i] {
			return false
		}
	}

	return true
}

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

func checkGender(g string) bool {

	if g == "male" {
		return true
	}
	if g == "female" {
		return true
	}
	if g == "other" {
		return true
	}
	return false

}
