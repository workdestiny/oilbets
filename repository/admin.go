package repository

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/service"
)

//AdminCheckUser เอาไว้สำหรับตรวจสอบดูว่ามีผู้ใช้นี้หรือไม่
func AdminCheckUser(q Queryer, userID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM users
		 WHERE id = $1;
	`, userID).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

// AdminGetUser view can get Me retrun user data
func AdminGetUser(q Queryer, email string) *entity.Me {

	me := &entity.Me{
		IsSignin: false,
	}

	var m entity.Me
	rows, err := q.Query(`
			SELECT users.id, users.birthdate, users.count->>'topic', users.count->>'gap',
			       users.role, users.display->>'mini', users.display->>'middle', users.display->>'normal', users.email->>'email',
				   users.firstname, users.lastname, users.gender, COALESCE(user_kycs.is_verify_email, 'false'),
				   COALESCE(user_kycs.is_idcard, 'false'), COALESCE(user_kycs.is_bookbank, 'false'), COALESCE(gap.id, ''), COALESCE(gap.name->>'text', ''),
				COALESCE(gap.display->>'middle', ''), COALESCE(gap.username->>'text', ''), users.notification
			  FROM users
		 LEFT JOIN user_kycs
				ON user_kycs.user_id = users.id
		 LEFT JOIN (SELECT * FROM gap ORDER BY count->>'popular' DESC) as gap
				ON gap.user_id = users.id
			 WHERE users.email->>'email' = $1;
			 `, email)
	if err != nil {
		return me
	}
	defer rows.Close()

	for rows.Next() {
		var g entity.GapList
		err := rows.Scan(&m.ID, &m.BirthDate, &m.Count.Topic, &m.Count.Gap,
			&m.Role, &m.DisplayImage.Mini, &m.DisplayImage.Middle, &m.DisplayImage.Normal, &m.Email,
			&m.FirstName, &m.LastName, &m.Gender, &m.IsVerify,
			&m.IsVerifyIDCard, &m.IsVerifyBookBank, &g.ID, &g.Name,
			&g.Display, &g.Username, &m.IsNotification)
		if err != nil {
			return me
		}

		if g.Username == "" {
			g.Username = g.ID
		}

		if g.ID != "" {
			m.Gap = append(m.Gap, g)
		}

	}

	return &m
}

// VerifyIDCard is verify idcard
func VerifyIDCard(q Queryer, userID string) error {

	_, err := q.Exec(`
		UPDATE user_kycs
		   SET is_idcard = $1, is_idcard_created_at = $2
		 WHERE user_id = $3;
	`, true, time.Now().UTC(), userID)

	if err != nil {
		return err
	}

	return nil
}

// VerifyBookbank is verify bookbank
func VerifyBookbank(q Queryer, userID string) error {

	_, err := q.Exec(`
		UPDATE user_kycs
		   SET is_bookbank = $1, is_bookbank_created_at = $2
		 WHERE user_id = $3;
	`, true, time.Now().UTC(), userID)

	if err != nil {
		return err
	}

	return nil
}

// UpdatePostVerifyIDCard is verify idcard update post
func UpdatePostVerifyIDCard(q Queryer, gapIDs []string) error {

	_, err := q.Exec(`
		UPDATE post
		   SET verify = $1
		 WHERE owner_id = ANY($2);
	`, 2, pq.Array(gapIDs))
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdateGapWallet is add Wallet
func AdminUpdateGapWallet(q Queryer, wallet, gapID string) error {

	_, err := q.Exec(`
		UPDATE wallets
		   SET saving = $1
		 WHERE gap_id = $2
	`, wallet, gapID)
	if err != nil {
		return err
	}

	return nil
}

//AdminGetGapRevenue input  gapID
func AdminGetGapRevenue(q Queryer, gapID string) (*entity.GetGapRevenueModel, error) {

	var g entity.GetGapRevenueModel
	err := q.QueryRow(`
		SELECT gap.id, gap.name->>'text', gap.display->>'mini', gap.username->>'text',
			   gap.user_id, COALESCE(wallets.saving, '0.00')
		  FROM gap
	 LEFT JOIN wallets
	 		ON wallets.gap_id = gap.id
		 WHERE (gap.id = $1 OR gap.username->>'text' = $1);
		`, gapID).Scan(&g.ID, &g.Name, &g.Display, &g.Username,
		&g.UserID, &g.Wallets)

	if err != nil {
		return nil, err
	}

	return &g, nil
}

//AdminListPostCountViewRevenue return model view and guestView sum
func AdminListPostCountViewRevenue(q Queryer, gapID string, t time.Time, limit int) ([]*entity.PostRevenueModel, error) {

	rows, err := q.Query(`
	 SELECT COALESCE(post.id, ''), COALESCE(post.slug, ''), COALESCE(post.title, ''), COALESCE(post.description, ''),
			COALESCE(post.created_at, now()), COALESCE(post.status_revenue, 0), COALESCE(post.image_url, ''), COALESCE(view.c, 0),
			COALESCE(guest.c, 0), COALESCE(post_note.note, '')
	   FROM (SELECT * FROM post WHERE owner_id = $1 AND status = true AND used = true LIMIT $2) as post
  LEFT JOIN (SELECT post_id, count(*) as c FROM view WHERE status_withdraw = $3 AND created_at < $4 GROUP BY post_id) as view
		 ON view.post_id = post.id
  LEFT JOIN (SELECT post_id, count(*) as c FROM guest WHERE status_withdraw = $3 AND created_at < $4 GROUP BY post_id) as guest
		 ON guest.post_id = post.id
  LEFT JOIN post_note
	 	 ON post_note.post_id = post.id
   ORDER BY post.created_at DESC
		`, gapID, limit, false, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.PostRevenueModel
	var id string
	for rows.Next() {

		var reject int
		var note string
		var p entity.PostRevenueModel
		err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description,
			&p.CreatedAt, &reject, &p.Image, &p.View.View,
			&p.View.GuestView, &note)
		if err != nil {
			return nil, err
		}

		if id != p.ID {
			p.Title = service.ShortTextTitleStripHTML(config.LimitTitleRevenuePostCountView, p.Title)

			if p.Title == "" {
				p.Description = service.ShortTextTitleStripHTML(config.LimitTitleRevenuePostCountView, p.Description)
				p.Title = p.Description
				if p.Description == "" {
					p.Title = p.Slug
				}
			}

			if p.Image == "" {
				p.Image = config.ImagePostRelate
			}

			if reject == 2 {
				var d decimal.Decimal
				p.View.AllAmount = d
				p.Reject = true
			}

			if reject == 0 {
				p.View.AllAmount = p.View.AmountAll()
			}

			p.Note = append(p.Note, note)
			p.Time = service.FormatRevenueDateType(p.CreatedAt) + " / " + service.FormatRevenueTimeType(p.CreatedAt)

			id = p.ID
			lp = append(lp, &p)
			continue
		}

		if len(lp) > 1 {
			lp[len(lp)-1].Note = append(lp[len(lp)-1].Note, note)
		}
	}

	return lp, nil
}

//AdminListPostCountViewRevenueNextLoad return model view and guestView sum
func AdminListPostCountViewRevenueNextLoad(q Queryer, gapID string, nextTime, t time.Time, limit int) ([]*entity.PostRevenueModel, error) {

	rows, err := q.Query(`
	 SELECT COALESCE(post.id, ''), COALESCE(post.slug, ''), COALESCE(post.title, ''), COALESCE(post.description, ''),
			COALESCE(post.created_at, now()), COALESCE(post.status_revenue, 0), COALESCE(post.image_url, ''), COALESCE(view.c, 0),
			COALESCE(guest.c, 0), COALESCE(post_note.note, '')
	   FROM (SELECT * FROM post WHERE owner_id = $1 AND status = true AND used = true AND created_at < $2 LIMIT $3) as post
  LEFT JOIN (SELECT post_id, count(*) as c FROM view WHERE status_withdraw = $4 AND created_at < $5 GROUP BY post_id) as view
		 ON view.post_id = post.id
  LEFT JOIN (SELECT post_id, count(*) as c FROM guest WHERE status_withdraw = $4 AND created_at < $5 GROUP BY post_id) as guest
		 ON guest.post_id = post.id
  LEFT JOIN post_note
	 	 ON post_note.post_id = post.id
   ORDER BY post.created_at DESC
		`, gapID, nextTime, limit, false, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.PostRevenueModel
	var id string
	for rows.Next() {

		var reject int
		var note string
		var p entity.PostRevenueModel
		err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description,
			&p.CreatedAt, &reject, &p.Image, &p.View.View,
			&p.View.GuestView, &note)
		if err != nil {
			return nil, err
		}

		if id != p.ID {
			p.Title = service.ShortTextTitleStripHTML(config.LimitTitleRevenuePostCountView, p.Title)

			if p.Title == "" {
				p.Description = service.ShortTextTitleStripHTML(config.LimitTitleRevenuePostCountView, p.Description)
				p.Title = p.Description
				if p.Description == "" {
					p.Title = p.Slug
				}
			}

			if p.Image == "" {
				p.Image = config.ImagePostRelate
			}

			if reject == 2 {
				var d decimal.Decimal
				p.View.AllAmount = d
				p.Reject = true
			}

			if reject == 0 {
				p.View.AllAmount = p.View.AmountAll()
			}

			p.View.Amount = service.Currency(p.View.AllAmount)
			p.Note = append(p.Note, note)
			p.Time = service.FormatRevenueDateType(p.CreatedAt) + " / " + service.FormatRevenueTimeType(p.CreatedAt)

			id = p.ID
			lp = append(lp, &p)
			continue
		}

		if len(lp) > 1 {
			lp[len(lp)-1].Note = append(lp[len(lp)-1].Note, note)
		}
	}

	return lp, nil
}

//AdminCheckPostNote เอาไว้สำหรับตรวจสอบดูว่า post มี หมายเหตุหรือไม่
func AdminCheckPostNote(q Queryer, postID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM post_note
		 WHERE post_id = $1;
	`, postID).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdatePostNote edit post note
func AdminUpdatePostNote(q Queryer, postID, note string) error {

	_, err := q.Exec(`
		UPDATE post_note
		   SET note = $1, updated_at = $2
		 WHERE post_id = $3;
	`, note, time.Now().UTC(), postID)
	if err != nil {
		return err
	}

	return nil
}

// AdminCreatePostNote create post note
func AdminCreatePostNote(q Queryer, postID, note string) error {

	_, err := q.Exec(`
	INSERT INTO post_note
				(post_id, note)
		 VALUES ($1, $2);
		 `, postID, note)
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdatePostStatusRevenue edit post statusRevenue
func AdminUpdatePostStatusRevenue(q Queryer, postID string) error {

	_, err := q.Exec(`
		UPDATE post
		   SET status_revenue = $1
		 WHERE id = $2;
	`, 2, postID)
	if err != nil {
		return err
	}

	return nil
}

// AdminDeletePost delete post
func AdminDeletePost(q Queryer, postID string) error {

	_, err := q.Exec(`
		UPDATE post
		   SET status = $1
		 WHERE id = $2;
	`, false, postID)
	if err != nil {
		return err
	}

	return nil
}

//GetUserByPostID Check Post ID
func GetUserByPostID(q Queryer, postID string) (string, string, error) {

	var email, title, slug, description string
	err := q.QueryRow(`
		SELECT COALESCE(users.email ->> 'email', ''), post.title, post.slug, post.description
		  FROM post
	 LEFT JOIN gap
			ON gap.ID = post.owner_id
	 LEFT JOIN users
	 		ON users.ID = gap.user_id
		 WHERE post.id = $1
		   AND post.status = true
		   AND post.used = true;
	`, postID).Scan(&email, &title, &slug, &description)
	if err != nil {
		return "", "", err
	}

	title = service.ShortTextTitleStripHTML(config.LimitTitleRevenuePostCountView, title)

	if title == "" {
		description = service.ShortTextTitleStripHTML(config.LimitTitleRevenuePostCountView, description)
		title = description
		if description == "" {
			title = slug
		}
	}

	return email, title, nil
}

//ListUserRequest list user request
func ListUserRequest(q Queryer) ([]*entity.UserRequestModel, error) {
	rows, err := q.Query(`
	SELECT users.id, users.email->>'email', users.firstname, users.lastname,
	 	   users.display->>'mini', rq.created_at, rq.type
	  FROM user_request as rq
 LEFT JOIN users
		ON users.id = rq.user_id
	 WHERE rq.status = $1
  ORDER BY rq.created_at DESC
	`, false)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lu []*entity.UserRequestModel

	for rows.Next() {
		var u entity.UserRequestModel
		err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName,
			&u.Display, &u.CreatedAt, &u.Type)
		if err != nil {
			return nil, err
		}

		lu = append(lu, &u)
	}

	return lu, nil
}

//SearchListUserRequest list user request
func SearchListUserRequest(q Queryer, text string) ([]*entity.UserRequestModel, error) {
	rows, err := q.Query(`
	SELECT users.id, users.email->>'email', users.firstname, users.lastname,
	 	   users.display->>'mini', rq.created_at, rq.type
	  FROM user_request as rq
 LEFT JOIN users
		ON users.id = rq.user_id
	 WHERE rq.status = $1 AND users.email->>'email' = $2
  ORDER BY rq.created_at DESC
	`, false, text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lu []*entity.UserRequestModel

	for rows.Next() {
		var u entity.UserRequestModel
		err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName,
			&u.Display, &u.CreatedAt, &u.Type)
		if err != nil {
			return nil, err
		}

		lu = append(lu, &u)
	}

	return lu, nil
}

//InsertUserRequest user request want to verify
func InsertUserRequest(q Queryer, userID string, t entity.UserRequestType) error {
	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM user_request
		 WHERE user_id = $1 AND type = $2 AND status = $3;
	`, userID, t, false).Scan(&id)
	if err == sql.ErrNoRows {

		_, err := q.Exec(`
	INSERT INTO user_request as rq
				(user_id, type, status)
		 VALUES ($1, $2, $3);
		 `, userID, t, false)
		if err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdateListVerify edit list verify
func AdminUpdateListVerify(q Queryer, userID string, typ entity.UserRequestType) error {

	_, err := q.Exec(`
		UPDATE user_request as rq
		   SET status = $1
		 WHERE user_id = $2 AND type = $3;
	`, true, userID, typ)
	if err != nil {
		return err
	}

	return nil
}

//AdminListGapRecommend list Gap input number
func AdminListGapRecommend(q Queryer) ([]*entity.GapRecommendModel, error) {

	rows, err := q.Query(`
		SELECT gap.id, gap.name->>'text', gap.display->>'normal', gap.username->>'text'
		  FROM gap_recommend
	 LEFT JOIN gap
	        ON gap_recommend.gap_id = gap.id;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.GapRecommendModel

	for rows.Next() {

		var g entity.GapRecommendModel
		err := rows.Scan(
			&g.ID, &g.Name, &g.Display, &g.Username)
		if err != nil {
			return nil, err
		}

		if g.Username == "" {
			g.Username = g.ID
		}

		list = append(list, &g)

	}

	return list, nil
}

//AdminSearchGapRecommend list Gap input number
func AdminSearchGapRecommend(q Queryer, text string, limit int) ([]*entity.GapRecommendModel, error) {

	rows, err := q.Query(`
		SELECT gap.id, gap.name->>'text', gap.display->>'mini', gap.username->>'text'
		  FROM gap
	 LEFT JOIN gap_recommend
			ON gap_recommend.gap_id = gap.id
		 WHERE gap.id = $1 OR gap.name->>'text' LIKE $2
		 LIMIT $3;
	`, text, text+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.GapRecommendModel

	for rows.Next() {

		var g entity.GapRecommendModel
		err := rows.Scan(
			&g.ID, &g.Name, &g.Display, &g.Username)
		if err != nil {
			return nil, err
		}

		if g.Username == "" {
			g.Username = g.ID
		}

		list = append(list, &g)

	}

	return list, nil
}

//AdminAddGapRecommend list Gap input number
func AdminAddGapRecommend(q Queryer, gapID string) error {
	_, err := q.Exec(`
	INSERT INTO gap_recommend
				(gap_id)
		 VALUES ($1) RETURNING id;
	`, gapID)
	return err
}

//AdminDeleteGapRecommend list Gap input number
func AdminDeleteGapRecommend(q Queryer, gapID string) error {
	_, err := q.Exec(`
	DELETE FROM gap_recommend
		  WHERE gap_id = $1;
	`, gapID)
	return err
}

//AdminCheckCountGapRecommend return count
func AdminCheckCountGapRecommend(q Queryer) (int, error) {
	var count int
	err := q.QueryRow(`
	SELECT COUNT(*)
	  FROM gap_recommend;
	`).Scan(&count)
	return count, err
}

//AdminListCategory list category
func AdminListCategory(q Queryer) ([]*entity.CategoryList, error) {

	rows, err := q.Query(`
			SELECT category.id, category.code, category.count, category.name->>'th'
			  FROM category;
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ct []*entity.CategoryList

	for rows.Next() {

		var c entity.CategoryList
		err := rows.Scan(&c.ID, &c.Code, &c.Count, &c.Name)
		if err != nil {
			return nil, err
		}

		ct = append(ct, &c)
	}

	return ct, nil
}

//AdminListCategoryVerified list category verified
func AdminListCategoryVerified(q Queryer) ([]*entity.CategoryList, error) {

	rows, err := q.Query(`
			SELECT category.id, category.code, category.count, category.name->>'th'
			  FROM category
			 WHERE verify = $1;
		`, true)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ct []*entity.CategoryList

	for rows.Next() {

		var c entity.CategoryList
		err := rows.Scan(&c.ID, &c.Code, &c.Count, &c.Name)
		if err != nil {
			return nil, err
		}

		ct = append(ct, &c)
	}

	return ct, nil
}

//AdminCreateCategory new category
func AdminCreateCategory(q Queryer, name string) error {

	n := NameTopic{
		Th: name,
		En: name,
	}

	var id string
	err := q.QueryRow(`
		INSERT INTO category
					(code, count, images, name,
					verify)
		     VALUES ($1, $2, $3, $4,
			        $5)
		  RETURNING id;
					`, name, 0, convJSON(ImageTopic{}), convJSON(n),
		true).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdateCategory edit category
func AdminUpdateCategory(q Queryer, catID, name string) error {

	n := NameTopic{
		Th: name,
		En: name,
	}

	_, err := q.Exec(`
		UPDATE category
		   SET code = $1, name = $2
		 WHERE id = $3;
	`, name, convJSON(n), catID)
	if err != nil {
		return err
	}

	return nil
}

//AdminCheckCategoryCode เอาไว้สำหรับตรวจสอบดูว่า category มีหรือไม่
func AdminCheckCategoryCode(q Queryer, code string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM category
		 WHERE code = $1;
	`, code).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

//AdminSearchTopic search
func AdminSearchTopic(q Queryer, text string) ([]*entity.TopicList, error) {

	rows, err := q.Query(`
			SELECT id, code, name->>'th'
			  FROM topic
			 WHERE name->>'th' LIKE $1
			 LIMIT 10;
		`, text+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lt []*entity.TopicList

	for rows.Next() {

		var t entity.TopicList

		err := rows.Scan(&t.ID, &t.Code, &t.Name)
		if err != nil {
			return nil, err
		}
		lt = append(lt, &t)
	}

	return lt, nil
}

//AdminCreateTopic new topic
func AdminCreateTopic(q Queryer, catID, name, image, title, description, imageFacebook, tagline string) (string, error) {

	n := NameTopic{
		Th: name,
		En: name,
	}

	i := ImageTopic{
		Normal: image,
		Mini:   image,
	}

	var id string
	err := q.QueryRow(`
		INSERT INTO topic
					(code, count, images, name,
					verify, cat_id, used_count, title,
					description, image_fb, tagline)
		     VALUES ($1, $2, $3, $4,
					$5, $6, $7, $8,
					$9, $10, $11)
		  RETURNING id;
					`, name, 0, convJSON(i), convJSON(n),
		true, catID, 0, title,
		description, imageFacebook, tagline).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

// AdminUpdateTopic edit topic
func AdminUpdateTopic(q Queryer, topicID, catID, name string, i ImageTopic) error {

	n := NameTopic{
		Th: name,
		En: name,
	}

	_, err := q.Exec(`
		UPDATE topic
		   SET code = $1, name = $2, cat_id = $3, images = $4
		 WHERE id = $5;
	`, name, convJSON(n), catID, convJSON(i),
		topicID)
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdateTopicVerify edit topic and verify
func AdminUpdateTopicVerify(q Queryer, topicID, catID, name string, i ImageTopic) error {

	n := NameTopic{
		Th: name,
		En: name,
	}

	_, err := q.Exec(`
		UPDATE topic
		   SET code = $1, name = $2, cat_id = $3, images = $4,
		       verify = $5
		 WHERE id = $6;
	`, name, convJSON(n), catID, convJSON(i),
		true, topicID)
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdateTopicSEOVerify edit topic and verify
func AdminUpdateTopicSEOVerify(q Queryer, topicID, title, description, tagline, imageFb string) error {

	_, err := q.Exec(`
		UPDATE topic
		   SET title = $1, description = $2, tagline = $3, image_fb = $4
		 WHERE id = $5;
	`, title, description, tagline, imageFb,
		topicID)
	if err != nil {
		return err
	}

	return nil
}

//AddWalletAndBonusUser input wallet, bonus
func AddWalletAndBonusUser(q Queryer, userID string, wallet int64, bonus int64) error {

	_, err := q.Exec(`
		UPDATE users
		   SET wallet = wallet + $1, bonus = bonus + $2
		 WHERE id = $3;
		 `, wallet, bonus, userID)
	if err != nil {
		return err
	}

	return nil
}
