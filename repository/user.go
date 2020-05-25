package repository

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/workdestiny/oilbets/entity"
)

// // GetUser view can get Me retrun user data
// func GetUser(ctx context.Context, ds *ds.Client, userID int64) *entity.Me {
// 	me := &entity.Me{
// 		IsSignin: false,
// 	}
// 	if userID > 0 {
// 		var u entity.UserModel
// 		err := ds.GetByID(ctx, entity.KindUser, userID, &u)
// 		if err != nil {
// 			return me
// 		}
// 		return convMe(&u)
// 	}
// 	return me

// }

// GetUser view can get Me retrun user data
func GetUser(q Queryer, userID string) *entity.Me {

	me := &entity.Me{
		IsSignin: false,
	}

	if userID != "" {

		var m entity.Me
		rows, err := q.Query(`
			SELECT users.id, users.birthdate, users.count->>'topic', users.count->>'gap',
			       users.role, users.display->>'mini', users.display->>'middle', users.display->>'normal', users.email->>'email',
				   users.firstname, users.lastname, users.gender, COALESCE(user_kycs.is_verify_email, 'false'),
				   COALESCE(user_kycs.is_idcard, 'false'), COALESCE(user_kycs.is_bookbank, 'false'), users.notification, users.wallet, users.bonus
			  FROM users
		 LEFT JOIN user_kycs
				ON user_kycs.user_id = users.id
			 WHERE users.id = $1;
			 `, userID)
		if err != nil {
			log.Println(err)
			return me
		}
		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&m.ID, &m.BirthDate, &m.Count.Topic, &m.Count.Gap,
				&m.Role, &m.DisplayImage.Mini, &m.DisplayImage.Middle, &m.DisplayImage.Normal, &m.Email,
				&m.FirstName, &m.LastName, &m.Gender, &m.IsVerify,
				&m.IsVerifyIDCard, &m.IsVerifyBookBank, &m.IsNotification, &m.Wallet, &m.Bonus)
			if err != nil {
				return me
			}
		}

		m.IsSignin = true

		return &m
	}
	return me
}

// GetUserByEmail view can get Me retrun user data
func GetUserByEmail(q Queryer, email string) *entity.Me {

	var m entity.Me
	rows, err := q.Query(`
			SELECT users.id, users.birthdate, users.count->>'topic', users.count->>'gap',
			       users.role, users.display->>'mini', users.display->>'middle', users.display->>'normal', users.email->>'email',
				   users.firstname, users.lastname, users.gender, COALESCE(user_kycs.is_verify_email, 'false'),
				   COALESCE(user_kycs.is_idcard, 'false'), COALESCE(user_kycs.is_bookbank, 'false'), users.notification, users.wallet, users.bonus
			  FROM users
		 LEFT JOIN user_kycs
				ON user_kycs.user_id = users.id
			 WHERE users.email->>'email' = $1;
			 `, email)
	if err != nil {
		return &m
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&m.ID, &m.BirthDate, &m.Count.Topic, &m.Count.Gap,
			&m.Role, &m.DisplayImage.Mini, &m.DisplayImage.Middle, &m.DisplayImage.Normal, &m.Email,
			&m.FirstName, &m.LastName, &m.Gender, &m.IsVerify,
			&m.IsVerifyIDCard, &m.IsVerifyBookBank, &m.IsNotification, &m.Wallet, &m.Bonus)
		if err != nil {
			return &m
		}
	}

	return &m
}

//ListFollowUser list User FollowerGap
func ListFollowUser(q Queryer, userID string, limit int) ([]*entity.UserFollowGapListModel, error) {

	rows, err := q.Query(`
			SELECT gap.id, gap.name->>'text', gap.display->>'mini', COALESCE(user_kycs.is_idcard, 'false'),
				   follow_gap.created_at, gap.username->>'text'
			  FROM (SELECT * FROM follow_gap
				   WHERE owner_id = $1 AND status = true AND used = true
				   ORDER BY created_at DESC
				   LIMIT $2) as follow_gap
		 LEFT JOIN gap
				ON follow_gap.gap_id = gap.id
		 LEFT JOIN user_kycs
		 		ON gap.user_id = user_kycs.user_id
		  ORDER BY follow_gap.created_at DESC;
		`, userID, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var lg []*entity.UserFollowGapListModel
	for rows.Next() {
		var g entity.UserFollowGapListModel
		err := rows.Scan(&g.ID, &g.Name, &g.Display, &g.IsVerify, &g.CreatedAt, &g.Username)
		if err != nil {
			return nil, err
		}

		if g.Username == "" {
			g.Username = g.ID
		}

		lg = append(lg, &g)
	}

	return lg, nil
}

//ListFollowUserNextLoad list User FollowerGap (NextLOad)
func ListFollowUserNextLoad(q Queryer, userID string, timeLoad time.Time, limit int) ([]*entity.UserFollowGapListModel, error) {

	rows, err := q.Query(`
			SELECT gap.id, gap.name->>'text', gap.display->>'mini', COALESCE(user_kycs.is_idcard, 'false'),
			follow_gap.created_at, gap.username->>'text'
			  FROM (SELECT * FROM follow_gap
				   WHERE owner_id = $1 AND status = true AND used = true AND created_at < $2
				   ORDER BY created_at DESC
				   LIMIT $3) as follow_gap
		 LEFT JOIN gap
				ON follow_gap.gap_id = gap.id
		 LEFT JOIN user_kycs
		 		ON gap.user_id = user_kycs.user_id
		  ORDER BY follow_gap.created_at DESC;
		`, userID, timeLoad, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var lg []*entity.UserFollowGapListModel
	for rows.Next() {
		var g entity.UserFollowGapListModel
		err := rows.Scan(&g.ID, &g.Name, &g.Display, &g.IsVerify, &g.CreatedAt, &g.Username)
		if err != nil {
			return nil, err
		}

		if g.Username == "" {
			g.Username = g.ID
		}

		lg = append(lg, &g)
	}

	return lg, nil
}

//EditNameUser input userID, name, lastname
func EditNameUser(q Queryer, userID string, firstName string, lastName string) error {

	_, err := q.Exec(`
		UPDATE users
		   SET firstname = $1, lastname = $2
		 WHERE id = $3;
		 `, firstName, lastName, userID)
	if err != nil {
		return err
	}

	return nil
}

// CheckOldPassword repository connect ds check checkOldPassword
func CheckOldPassword(q Queryer, userID string) (string, error) {

	var id string
	var pwd string
	err := q.QueryRow(`
		SELECT id, provider->>'password'
		  FROM user_provider
		 WHERE id = $1;`, userID).Scan(&id, &pwd)
	if err != nil {
		return "", err
	}
	return pwd, nil

}

// UpdateNewPassword is change password
func UpdateNewPassword(q Queryer, userID string, password string) error {

	_, err := q.Exec(`
		UPDATE user_provider
		   SET provider = provider::jsonb - 'password' || '{"password": "`+password+`"}'::jsonb
		 WHERE id = $1;
		 `, userID)
	if err != nil {
		return err
	}
	return nil
}

//EditProfileDisplay edit profile display
func EditProfileDisplay(q Queryer, userID string, display DisplayGap) error {

	_, err := q.Exec(`
		UPDATE users
		   SET display = $1
		 WHERE id = $2;
	`, convJSON(display), userID)

	return err

}

//UpdateCodeSendEmailVerify update code in db and send email verify
func UpdateCodeSendEmailVerify(q Queryer, baseURL, userID, email, name string) error {

	newCode := uuid.New().String()

	code := codeEmail{
		Code: newCode,
	}

	_, err := q.Exec(`
		UPDATE users
		   SET email = email::jsonb - 'code' || $1::jsonb
		 WHERE id = $2;
	`, convJSON(code), userID)
	if err != nil {
		return err
	}

	SendEmailVerify(baseURL, email, newCode, name)
	return nil
}

// UpdateNewEmail is change new Email
func UpdateNewEmail(q Queryer, userID string, newEmail string) error {

	emailVerify := emailUser{
		Code:   "",
		Email:  newEmail,
		TimeAt: time.Now().Unix(),
		Verify: false,
	}

	_, err := q.Exec(`
		UPDATE users
		   SET email = $1
		 WHERE id = $2;
	`, convJSON(emailVerify), userID)

	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE user_provider
		   SET provider = provider::jsonb - 'email' || '{"email": "`+newEmail+`"}'::jsonb
		 WHERE id = $1;
		 `, userID)

	if err != nil {
		return err
	}

	_, err = q.Exec(`
	UPDATE user_kycs
	   SET is_verify_email = $1
	 WHERE user_id = $2;
`, false, userID)

	if err != nil {
		return err
	}

	return nil
}

// CheckBookbank check bookbank
func CheckBookbank(q Queryer, userID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM public.bookbank
		 WHERE user_id = $1;
		 `, userID).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

// InsertBookbank is change new bookbank
func InsertBookbank(q Queryer, userID, bookbankName, bookbankNumber, bankID, image string) error {

	_, err := q.Exec(`
			INSERT INTO public.bookbank
			            (user_id, number, name, bank_id, image)
				 VALUES ($1, $2, $3, $4, $5);
	`, userID, bookbankNumber, bookbankName, bankID, image)
	if err != nil {
		return err
	}

	return nil
}

// UpdateBookbank is change new bookbank
func UpdateBookbank(q Queryer, userID, bookbankName, bookbankNumber, bankID, image string) error {

	_, err := q.Exec(`
				UPDATE public.bookbank
				   SET number = $1, name = $2, bank_id = $3, image = $4
  				 WHERE user_id = $5;
	`, bookbankNumber, bookbankName, bankID, image, userID)
	if err != nil {
		return err
	}

	return nil
}

// GetBookbank is get bookbank
func GetBookbank(q Queryer, userID string) (*entity.Bookbank, error) {

	b := entity.Bookbank{}

	var bankID string
	err := q.QueryRow(`
		SELECT number, name, bank_id
		  FROM bookbank
		 WHERE user_id = $1
	  ORDER BY created_at DESC;
	`, userID).Scan(&b.Number, &b.Name, &bankID)
	if err != nil {
		return nil, err
	}

	b.BankName = entity.GetBankNameByID(bankID)

	return &b, nil
}

// GetUserBookbank is get bookbank
func GetUserBookbank(q Queryer, userID string) (*entity.UserBookbank, error) {

	b := entity.UserBookbank{}

	err := q.QueryRow(`
		SELECT number, owner, bank
		  FROM user_bookbank
		 WHERE id = $1;
	`, userID).Scan(&b.Number, &b.Owner, &b.BankName)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

//UpdateWalletUser input id, amount
func UpdateWalletUser(q Queryer, userID string, amount int64) error {

	_, err := q.Exec(`
		UPDATE users
		   SET wallet = wallet - $1
		 WHERE id = $2;
		 `, amount, userID)
	if err != nil {
		return err
	}

	return nil
}

//UpdateBookbankUser input id, amount
func UpdateBookbankUser(q Queryer, userID, owner, number, bank string) error {

	_, err := q.Exec(`
		UPDATE user_bookbank
		   SET number = $1, owner = $2, bank = $3
		 WHERE id = $4;
		 `, number, owner, bank, userID)
	if err != nil {
		return err
	}

	return nil
}

//UpdateWalletAndBonusUser input wallet, bonus
func UpdateWalletAndBonusUser(q Queryer, userID string, wallet int64, bonus int64) error {

	_, err := q.Exec(`
		UPDATE users
		   SET wallet = $1, bonus = $2
		 WHERE id = $3;
		 `, wallet, bonus, userID)
	if err != nil {
		return err
	}

	return nil
}

type codeEmail struct {
	Code string `json:"code"`
}
