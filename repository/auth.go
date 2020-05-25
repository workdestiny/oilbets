package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/workdestiny/oilbets/entity"

	"github.com/go-redis/redis"
	"github.com/workdestiny/oilbets/config"

	// postgre
	_ "github.com/lib/pq"
)

// Signin repository connect ds check signin
func Signin(q Queryer, email string) (string, string, error) {

	var id string
	var pwd string
	err := q.QueryRow(`SELECT id, provider->>'password'
		                 FROM user_provider
		                WHERE provider->>'email' = $1;`, email).Scan(&id, &pwd)
	if err != nil {
		return "", "", err
	}
	return id, pwd, nil

}

// CheckEmail input email check in ds
func CheckEmail(q Queryer, email string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM user_provider
		 WHERE provider->>'email' = $1;
		 `, email).Scan(&id)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}

	return fmt.Errorf("can't use email")

}

// CreateUser is struct model
type CreateUser struct {
	ID        string
	Email     string
	Password  string
	FirstName string
	LastName  string
	Display   Display
}

//CreateUserProvider model
type CreateUserProvider struct {
	Provider CreateProvider
}

//CreateProvider model
type CreateProvider struct {
	GoogleID     string `json:"google_id"`
	GoogleCode   string `json:"google_code"`
	FacebookID   string `json:"facebook_id"`
	FacebookCode string `json:"facebook_code"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	XCode        string `json:"x_code"`
}

// Contact data in struct
type contact struct {
	Address string `json:"address"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type count struct {
	Topic int `json:"topic"`
	Gap   int `json:"gap"`
}

// Display User
type Display struct {
	Mini   string `json:"mini"`
	Middle string `json:"middle"`
	Normal string `json:"normal"`
}

// statusUser User
type statusUser struct {
	Level     int   `json:"user_level"`
	CreatedAt int64 `json:"created_at"`
	ExpreAt   int64 `json:"expre_at"`
}

type emailUser struct {
	Verify bool   `json:"verify"`
	TimeAt int64  `json:"time_at"`
	Code   string `json:"code"`
	Email  string `json:"email"`
}

// bookbankUser User
type bookbankUser struct {
	Name     string `json:"name"`
	Number   string `json:"number"`
	BankName string `json:"bankName"`
	Image    string `json:"image"`
}

// username User
type username struct {
	Text string `json:"text"`
	Swap bool   `json:"swap"`
}

// verify User
type verify struct {
	At    int64 `json:"at"`
	Level int   `json:"level"`
}

//GetForgetPassword check email id in db
func GetForgetPassword(q Queryer, email string) (string, time.Time, int, error) {

	var t time.Time
	var count int
	var id string
	err := q.QueryRow(`
		SELECT id, time, count
		  FROM user_provider
		 WHERE provider->>'email' = $1;`,
		email).Scan(&id, &t, &count)
	if err != nil {
		return "", t, 0, err
	}
	return id, t, count, nil
}

//GetEmail get email user
func GetEmail(q Queryer, id string) (string, error) {

	var email string
	err := q.QueryRow(`
		SELECT email->>'email'
		  FROM users
		 WHERE id = $1;
		 `, id).Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

//CheckCodeEmailVerify check emailand code in db
func CheckCodeEmailVerify(q Queryer, email, code string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM users
		 WHERE email->>'email' = $1 AND email->>'code' = $2;`,
		email, code).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

//VerifyEmail verify email
func VerifyEmail(q Queryer, id string) error {

	_, err := q.Exec(`
		UPDATE user_kycs
		   SET is_verify_email = $1, is_verify_email_created_at = $2
		 WHERE user_id = $3;
	`, true, time.Now().UTC(), id)

	return err
}

//SetCodeResetPassword check email id in db
func SetCodeResetPassword(q Queryer, id string, t time.Time, i int, code string) error {

	_, err := q.Exec(`
		UPDATE user_provider
		   SET time = $1, count = $2, provider = provider::jsonb - 'x_code' || '{"x_code": "`+code+`"}'::jsonb
		 WHERE id = $3;
	`, t, i, id)

	return err
}

//SetResetPassword reset password in db (user_provider)
func SetResetPassword(q Queryer, email, code, hashPassword string) error {

	_, err := q.Exec(`
		UPDATE user_provider
		   SET provider = provider::jsonb - 'password' || '{"password": "`+hashPassword+`"}'::jsonb - 'x_code' || '{"x_code": ""}'::jsonb
		 WHERE provider->>'email' = $1
		   AND provider->>'x_code' = $2;
	`, email, code)

	return err
}

//CheckFacebookID check facebook id in db
func CheckFacebookID(q Queryer, facebookID string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM user_provider
		 WHERE provider->>'facebook_id' = $1;`,
		facebookID).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

//CheckGoogleID check google id in db
func CheckGoogleID(q Queryer, googleID string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM user_provider
		 WHERE provider->>'google_id' = $1;`,
		googleID).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

// CreateSocialUserProvicer new user signin facebook
func CreateSocialUserProvicer(q Queryer, data *CreateProvider) string {

	var id string
	q.QueryRow(`INSERT INTO user_provider
			                (code_social_signin, count, provider, update_email)
			        VALUES ($1, $2, $3, $4) RETURNING id;`,
		"", 0, convJSON(data), "{}").Scan(&id)

	return id
}

// CreateSocialUser new user signin facebook
func CreateSocialUser(q Queryer, redis *redis.Client, user *CreateUser) error {

	now := time.Now()

	_, err := q.Exec(`
	INSERT INTO users
				(id, about_me, contact, count,
				display, email, firstname, gender,
				gap, lastname, role, status,
				used, username)
	     VALUES ($1, $2, $3, $4,
		        $5, $6, $7, $8,
		        $9, $10, $11, $12,
		        $13, $14);`,
		user.ID, "{}",
		convJSON(contact{
			Address: "",
			City:    "",
			Country: "",
		}),
		convJSON(count{
			Topic: 0,
			Gap:   0,
		}),
		convJSON(user.Display),
		convJSON(emailUser{
			Email:  user.Email,
			Code:   "",
			TimeAt: now.Unix(),
			Verify: false,
		}),
		user.FirstName,
		"other",
		false,
		user.LastName,
		0,
		convJSON(statusUser{
			Level:     0,
			CreatedAt: now.Unix(),
			ExpreAt:   now.Unix(),
		}),
		true,
		convJSON(username{
			Text: "",
			Swap: true,
		}))
	if err != nil {
		log.Println("22222")
		return err
	}

	_, err = q.Exec(`
	INSERT INTO user_kycs
				(user_id, is_email, is_phone, is_verify_email,
				is_idcard, is_bookbank, is_ban)
	     VALUES ($1, $2, $3, $4,
		        $5, $6, $7);`,
		user.ID, false, false, false,
		false, false, false)
	if err != nil {
		log.Println("33333")
		return err
	}

	//เพิ่มข้อมูล New User To Redis
	go AddUserToRedis(redis, entity.RedisUserModel{
		ID:               user.ID,
		Username:         "",
		FirstName:        user.FirstName,
		LastName:         user.LastName,
		DisplayImage:     user.Display.Middle,
		DisplayImageMini: user.Display.Mini,
		Level:            0,
	})

	return nil
}

// Register is create new user to ds
func Register(q Queryer, redis *redis.Client, req *CreateUser, agent string) (string, string, error) {

	randomCode := RandStr(40)
	now := time.Now()

	// create new userProvider
	var id string
	err := q.QueryRow(`
		INSERT INTO user_provider
			        (code_social_signin, count, provider, update_email)
			VALUES ($1, $2, $3, $4)
		 RETURNING id;
					`,
		"", 0, convJSON(CreateProvider{
			Email:    req.Email,
			Password: req.Password,
		}), "{}").Scan(&id)
	if err != nil {
		return "", "", err
	}

	verifyCode := randomCode + id

	_, err = q.Exec(`
		INSERT INTO users
					(id, about_me, contact, count,
					display, email, firstname, gender,
					gap, lastname, role, status,
					used, username)
			 VALUES ($1, $2, $3, $4,
					$5, $6, $7, $8,
					$9, $10, $11, $12,
					$13, $14);
					`,
		id,
		"{}",
		convJSON(contact{
			Address: "",
			City:    "",
			Country: "",
		}),
		convJSON(count{
			Topic: 0,
			Gap:   0,
		}),
		convJSON(req.Display),
		convJSON(emailUser{
			Email:  req.Email,
			Code:   "",
			TimeAt: now.Unix(),
			Verify: false,
		}),
		req.FirstName,
		"other",
		false,
		req.LastName,
		0,
		convJSON(statusUser{
			Level:     0,
			CreatedAt: now.Unix(),
			ExpreAt:   now.Unix(),
		}),
		true,
		convJSON(username{
			Text: "",
			Swap: true,
		}))
	if err != nil {
		return "", "", err
	}

	_, err = q.Exec(`
		INSERT INTO user_kycs
					(user_id, is_email, is_phone, is_verify_email,
					is_idcard, is_bookbank, is_ban)
	   	     VALUES ($1, $2, $3, $4,
			        $5, $6, $7);`,
		id, false, false, false,
		false, false, false)
	if err != nil {
		return "", "", err
	}

	_, err = q.Exec(`
	INSERT INTO user_bookbank
				(id, number, bank, owner)
			VALUES ($1, $2, $3, $4);`,
		id, "", "", "")
	if err != nil {
		return "", "", err
	}

	//เพิ่มข้อมูล New User To Redis
	go AddUserToRedis(redis, entity.RedisUserModel{
		ID:               id,
		Username:         "",
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		DisplayImage:     req.Display.Middle,
		DisplayImageMini: req.Display.Mini,
		Level:            0,
	})

	return id, verifyCode, nil
}

// CheckBirthDateFormat check format BirthDate
func CheckBirthDateFormat(day, month, year string) (time.Time, bool) {

	// check BirthDate
	layout := "2006-01-02T15:04:05.000Z"
	str := year + "-" + month + "-" + day + "T00:00:00.000Z"
	t, err := time.Parse(layout, str)
	if err != nil {
		return t, false
	}

	// check use A.D. (Anno Domini)
	now := time.Now()
	yearNow := now.Format("2006")
	yearInput := t.Format("2006")

	intYearNow, _ := strconv.ParseInt(yearNow, 0, 64)
	intYear, _ := strconv.ParseInt(yearInput, 0, 64)

	if intYear > (intYearNow - 13) {
		return t, false
	}

	return t, true
}

// GetImageProfileURL input gender output images rerate genter
func GetImageProfileURL(gender string) (string, string) {

	if gender == "male" {
		return config.ImageProfileM, config.ImageProfileMMini
	}

	if gender == "female" {

		return config.ImageProfileF, config.ImageProfileFMini
	}

	return config.ImageProfileO, config.ImageProfileOMini
}
