package repository

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mssola/user_agent"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
)

//GetGap input userID, gapID
func GetGap(q Queryer, userID string, gapID string) (*entity.GetGapModel, error) {

	var g entity.GetGapModel
	var follow interface{}
	var ownerID string
	err := q.QueryRow(`
		SELECT gap.id, gap.name->>'text', gap.bio, gap.cover->>'normal',
			   gap.display->>'normal', COALESCE(user_kycs.is_idcard, 'false'), gap.count->>'follower', gap.count->>'post',
			   follow_gap.status, gap.user_id, gap.username->>'text'
		  FROM gap
	 LEFT JOIN follow_gap
			ON follow_gap.gap_id = gap.id
		   AND follow_gap.owner_id = $1
		   AND follow_gap.used = true
	 LEFT JOIN user_kycs
			ON user_kycs.user_id = gap.user_id
		 WHERE gap.id = $2 OR gap.username->>'text' = $2;
		`, userID, gapID).Scan(&g.ID, &g.Name, &g.Bio, &g.Cover,
		&g.Display, &g.IsVerify, &g.Count.Follower, &g.Count.Post,
		&follow, &ownerID, &g.Username)

	if err != nil {
		return nil, err
	}

	if follow != nil {
		g.IsFollow = follow.(bool)
	}

	if ownerID == userID {
		g.IsOwner = true
	}

	return &g, nil
}

//GetGapDetail input userID, gapID
func GetGapDetail(q Queryer, gapID string) (*entity.GetGapDetailModel, error) {

	var g entity.GetGapDetailModel
	err := q.QueryRow(`
		SELECT gap.id, gap.name->>'text', gap.username->>'text', gap.display->>'middle',
			   gap.display->>'mini', gap.user_id, gap.count->>'follower', gap.count->>'popular'
		  FROM gap
		 WHERE gap.id = $1;
		`, gapID).Scan(&g.ID, &g.Name, &g.Username, &g.Display,
		&g.DisplayMini, &g.UserID, &g.CountFollower, &g.CountPopular)

	if err != nil {
		return nil, err
	}

	return &g, nil
}

//GetGapSetting input userID, gapID
func GetGapSetting(q Queryer, rd *redis.Client, userID string, gapID string) (*entity.GetGapSettingModel, error) {

	var g entity.GetGapSettingModel
	var topicID string
	var nameTime int64
	err := q.QueryRow(`
		SELECT gap.id, gap.name->>'text', gap.bio, gap.created_at,
			   gap.topic_id, gap.username->>'text', contact.tel, contact.email,
			   contact.social, contact.city, contact.country, contact.address,
			   gap.cover->>'normal', gap.display->>'mini', COALESCE(user_kycs.is_idcard, 'false'), gap.count->>'follower',
			   gap.count->>'post', gap.username->>'swap', gap.name->>'time'
		  FROM gap
	 LEFT JOIN contact
			ON gap.id = contact.owner_id AND contact.owner_type = 1
	 LEFT JOIN user_kycs
			ON user_kycs.user_id = gap.user_id
		 WHERE (gap.id = $1 OR gap.username->>'text' = $1) AND gap.user_id = $2 ;
		`, gapID, userID).Scan(&g.ID, &g.Name, &g.Bio, &g.CreatedAt,
		&topicID, &g.Username, &g.Tel, &g.Email,
		&g.Social, &g.City, &g.Country, &g.Address,
		&g.Cover, &g.Display, &g.IsVerify, &g.Count.Follower,
		&g.Count.Post, &g.IsChangeUsername, &nameTime)

	if err != nil {
		return nil, err
	}

	duration := time.Now().Unix() - nameTime

	if duration > config.ExpiredNameGapDuration {
		g.IsChangeName = true
	}

	if g.City == "" {
		g.City = entity.GetProvinceName(1)
	}

	if g.Country == "" {
		g.Country = "ไม่ระบุ"
	}

	g.Topic = "ไม่ระบุ"

	if topicID != "" {
		g.Topic = getTopicName(rd, topicID)
	}

	return &g, nil
}

//GetGapStatistic input userID, gapID
func GetGapStatistic(q Queryer, rd *redis.Client, userID string, gapID string) (*entity.GetGapSettingModel, error) {

	var g entity.GetGapSettingModel
	err := q.QueryRow(`
		SELECT gap.id, gap.name->>'text', gap.bio, gap.username->>'text',
			   gap.cover->>'normal', gap.display->>'mini'
		  FROM gap
		 WHERE (gap.id = $1 OR gap.username->>'text' = $1) AND gap.user_id = $2 ;
		`, gapID, userID).Scan(&g.ID, &g.Name, &g.Bio, &g.Username,
		&g.Cover, &g.Display)

	if err != nil {
		return nil, err
	}

	return &g, nil
}

//GetGapRevenue input userID, gapID
func GetGapRevenue(q Queryer, userID string, gapID string) (*entity.GetGapRevenueModel, error) {

	var g entity.GetGapRevenueModel
	err := q.QueryRow(`
		SELECT gap.id, gap.name->>'text', gap.display->>'mini', gap.username->>'text',
			   COALESCE(wallets.saving, '0.00')
		  FROM gap
	 LEFT JOIN wallets
	 		ON wallets.gap_id = gap.id
		 WHERE (gap.id = $1 OR gap.username->>'text' = $1) AND gap.user_id = $2 ;
		`, gapID, userID).Scan(&g.ID, &g.Name, &g.Display, &g.Username,
		&g.Wallets)

	if err != nil {
		return nil, err
	}

	return &g, nil
}

//ListGapRecommend list Gap input number
func ListGapRecommend(q Queryer, userID string, limit int) ([]*entity.GapRecommendModel, error) {

	rows, err := q.Query(`
		SELECT gap.id, gap.name->>'text', gap.display->>'middle', gap.cover->>'mini',
			   gap.count->>'follower', follow_gap.status, gap.user_id, gap.username->>'text'
		  FROM gap_recommend
	 LEFT JOIN gap
	        ON gap_recommend.gap_id = gap.id
	 LEFT JOIN follow_gap
			ON gap_recommend.gap_id = follow_gap.gap_id
		   AND follow_gap.owner_id = $1
	  ORDER BY random()
	     LIMIT $2;
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gid string
	var list []*entity.GapRecommendModel

	for rows.Next() {

		var like interface{}
		var owner string
		var g entity.GapRecommendModel
		err := rows.Scan(
			&g.ID, &g.Name, &g.Display,
			&g.Cover, &g.FollowerCount, &like, &owner, &g.Username)
		if err != nil {
			return nil, err
		}

		if gid != g.ID {

			if like != nil {
				g.IsFollow = like.(bool)
			}

			if g.Username == "" {
				g.Username = g.ID
			}

			if owner == userID {
				g.IsOwner = true
			}

			gid = g.ID
			list = append(list, &g)
		}

	}

	return list, nil
}

//ListFollowGapID list gap id
func ListFollowGapID(q Queryer, userID string) []string {

	rows, err := q.Query(`
			SELECT gap_id
			  FROM public.follow_gap
			 WHERE owner_id = $1 AND status = true AND used = true;
		`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var gapIDs []string

	for rows.Next() {

		var gapID string
		err := rows.Scan(&gapID)
		if err != nil {
			return nil
		}

		gapIDs = append(gapIDs, gapID)
	}

	return gapIDs
}

//ListGapID list gap id
func ListGapID(q Queryer, userID string) []string {

	rows, err := q.Query(`
			SELECT id
			  FROM gap
			 WHERE user_id = $1
		`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var gapIDs []string

	for rows.Next() {

		var gapID string
		err := rows.Scan(&gapID)
		if err != nil {
			return nil
		}

		gapIDs = append(gapIDs, gapID)
	}

	return gapIDs
}

//CheckGap เอาไว้สำหรับตรวจสอบดูว่ามีแก๊ปนี้หรือไม่
func CheckGap(q Queryer, gapID string) error {

	var used bool
	err := q.QueryRow(`
		SELECT used
		  FROM public.gap
		 WHERE id = $1
		   AND used = true;
	`, gapID).Scan(&used)
	if err != nil {
		return err
	}

	return nil
}

//CheckGapRecommend เอาไว้สำหรับตรวจสอบดูว่ามีแก๊ปนี้หรือไม่
func CheckGapRecommend(q Queryer, gapID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM gap_recommend
		 WHERE gap_id = $1;
	`, gapID).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

//CheckOwnerGap is return true return Gap
func CheckOwnerGap(rd *redis.Client, gapID string, userID string) (*entity.RedisGapModel, bool) {

	arrStringGap, err := rd.Keys(config.RedisGap + userID + ":" + gapID).Result()
	if err != nil {
		return nil, false
	}

	// arrStringGap, err := redis.Strings(c.Do("KEYS", config.RedisGap+userID+":"+gapID))
	// if err != nil {
	// 	return nil, false
	// }

	if len(arrStringGap) > 0 {

		var g entity.RedisGapModel
		gBytes, _ := rd.Get(arrStringGap[0]).Bytes()
		//gBytes, _ := redis.Bytes(c.Do("GET", arrStringGap[0]))
		gob.NewDecoder(bytes.NewReader(gBytes)).Decode(&g)

		return &g, true
	}

	return nil, false
}

//CheckFollowGap ตรวจสอบว่าเคย follow แล้วหรือยังถ้ายัง ให้ create ถ้าเคยแล้วจะ return (bool) true or false
func CheckFollowGap(q Queryer, userID string, gapID string) (bool, error) {

	var isLike bool
	err := q.QueryRow(`
		SELECT status
		  FROM public.follow_gap
		 WHERE gap_id = $1
		   AND owner_id = $2
		   AND used = true;
	`, gapID, userID).Scan(&isLike)

	return isLike, err

}

//CreateFollowGap (ต้องใช้ Tx) สร้าง follow ที่ไม่เคยมีขึ้นมาใหม่ หลังจากสร้างสำเร็จจะทำการ update ค่า count following gap
func CreateFollowGap(q Queryer, userID string, gapID string) error {

	_, err := q.Exec(`
		INSERT INTO public.follow_gap
					(gap_id, owner_id, status, used)
			 VALUES ($1, $2, $3, $4) RETURNING id;
	`, gapID, userID, true, true)
	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE public.gap
		   SET count = count::jsonb || CONCAT('{"follower":', COALESCE(count->>'follower','0')::int + 1, '}')::jsonb
		 WHERE id = $1;
	`, gapID)
	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE users
		   SET count = count::jsonb || CONCAT('{"gap":', COALESCE(count->>'gap','0')::int + 1, '}')::jsonb
		 WHERE id = $1;
	`, userID)
	if err != nil {
		return err
	}

	return nil
}

//FollowGap (isFollow คือ follow ที่ user ต้องการให้เป็นเช่น user ต้องการ follow gap นี้ isFollow = true)
//เรียกใช้เมื่อ user นี้ เคย follow gap นี้แล้ว จำทำการเปลี่ยนจาก follow เป็น unfollow หรือ unfollow เป็น follow
func FollowGap(q Queryer, userID string, gapID string, isFollow bool) error {

	_, err := q.Exec(`
		UPDATE public.follow_gap
		   SET status = $1
		 WHERE gap_id = $2
		   AND owner_id = $3
		   AND used = true;
	`, isFollow, gapID, userID)

	set := "+ 1"
	if !isFollow {
		set = "- 1"
	}

	_, err = q.Exec(`
		UPDATE public.gap
		   SET count = count::jsonb || CONCAT('{"follower":', COALESCE(count->>'follower','0')::int `+set+`, '}')::jsonb
		 WHERE id = $1;
	`, gapID)

	_, err = q.Exec(`
		UPDATE users
		   SET count = count::jsonb || CONCAT('{"gap":', COALESCE(count->>'gap','0')::int `+set+`, '}')::jsonb
		 WHERE id = $1;
	`, userID)

	return err
}

//ListGap input userID, limit list popular DESC
func ListGap(q Queryer, userID string, limit int) ([]*entity.GapList, error) {

	rows, err := q.Query(`
		SELECT gap.id, gap.name->>'text', gap.display->>'normal', gap.username->>'text'
		  FROM gap
		 WHERE user_id = $1 AND used = true
	  ORDER BY COALESCE(count->>'popular','0')::int DESC
	     LIMIT $2;
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lg []*entity.GapList

	for rows.Next() {

		var g entity.GapList
		err := rows.Scan(&g.ID, &g.Name, &g.Display, &g.Username)
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

//UpdateCountGapPost update count post
func UpdateCountGapPost(q Queryer, gapID string) error {

	_, err := q.Exec(`
		UPDATE public.gap
		SET count = count::jsonb || CONCAT('{"post":', COALESCE(count->>'post','0')::int + 1, '}')::jsonb
		 WHERE id = $1;
	`, gapID)

	return err
}

//MinusCountGapPost update count post
func MinusCountGapPost(q Queryer, gapID string) error {

	_, err := q.Exec(`
		UPDATE public.gap
		SET count = count::jsonb || CONCAT('{"post":', COALESCE(count->>'post','0')::int - 1, '}')::jsonb
		 WHERE id = $1;
	`, gapID)

	return err
}

//CreateGap input db, model
func CreateGap(q Queryer, req *CreateGapModal) (string, error) {

	now := time.Now()
	var id string
	err := q.QueryRow(`
		INSERT INTO public.gap
			(bio, cat_id, count, cover,
			display, name, status, topic_id,
			used, user_id, username, verify)
			VALUES ($1, $2, $3, $4,
				$5, $6, $7, $8,
				$9, $10, $11, $12)
				RETURNING id;
				`, req.Bio, "0",
		convJSON(countGap{
			Popular:  0,
			Follower: 0,
			View:     0,
			Post:     0,
		}),
		convJSON(req.Cover),
		convJSON(req.Display),
		convJSON(gapName{
			Text: req.Name.Text,
			Time: now.AddDate(0, 0, -90).Unix(),
		}),
		convJSON(statusGap{
			CreatedAt: now.Format(time.RFC3339),
			ExpreAt:   now.Format(time.RFC3339),
			Level:     0,
		}),
		req.TopicID,
		true,
		req.UserID,
		convJSON(usernameGap{
			Text: "",
			Swap: true,
		}),
		convJSON(verifyGap{
			Status: false,
			Level:  0,
		})).Scan(&id)

	if err != nil {
		return "", err
	}

	_, err = q.Exec(`
		INSERT INTO public.contact
					(owner_id, owner_type, tel, social,
					website, email, address, city,
					country)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id;
	`, id, 1, "", "",
		"", "", "", "ไม่ระบุ",
		"ไม่ระบุ")
	if err != nil {
		return "", err
	}

	return id, nil
}

//ListUserFollowrGap list count DESC
func ListUserFollowrGap(q Queryer, gapID string, limit int) ([]*entity.UserFollowerGapModel, error) {

	rows, err := q.Query(`
		   SELECT users.id, users.firstname, users.lastname, users.display->>'mini',
				  gap.id, gap.name->>'text', gap.display->>'mini', follow_gap.created_at,
				  gap.username->>'text'
			 FROM (SELECT * FROM follow_gap
				  WHERE gap_id = $1 AND status = true AND used = true
				  ORDER BY created_at DESC
				  LIMIT $2) as follow_gap
		LEFT JOIN users
			   ON users.id = follow_gap.owner_id
		LEFT JOIN gap
			   ON follow_gap.owner_id = gap.user_id
		 ORDER BY follow_gap.created_at DESC, cast(gap.count->>'follower' as integer) DESC;
		`, gapID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lu []*entity.UserFollowerGapModel
	var uid string
	count := 0
	for rows.Next() {

		var u entity.UserFollowerGapModel
		var g entity.GapUserFollower
		var gID, gName, gDisplay, gUsername interface{}
		var firstname, lastname string
		err := rows.Scan(&u.ID, &firstname, &lastname, &u.Display, &gID, &gName, &gDisplay, &u.CreatedAt, &gUsername)
		if err != nil {
			return nil, err
		}

		if gID != nil {
			g.ID = gID.(string)
		}

		if gUsername != nil {
			g.Username = gUsername.(string)
		}

		if gName != nil {
			g.Name = gName.(string)
		}

		if gDisplay != nil {
			g.Display = gDisplay.(string)
		}

		if g.Username == "" {
			g.Username = g.ID
		}

		if u.ID != uid {
			uid = u.ID
			count = 0

			u.Name = firstname + " " + lastname

			if g.ID != "" {
				u.Gap = append(u.Gap, g)
			}

			lu = append(lu, &u)
			count++
			continue
		}

		if count < 4 {

			if g.ID != "" {
				lu[len(lu)-1].Gap = append(lu[len(lu)-1].Gap, g)
			}
			count++
			continue
		}

	}

	return lu, nil
}

//ListUserFollowrGapNextLoad list count DESC (NextLoad)
func ListUserFollowrGapNextLoad(q Queryer, gapID string, timeLoad time.Time, limit int) ([]*entity.UserFollowerGapModel, error) {

	rows, err := q.Query(`
		SELECT users.id, users.firstname, users.lastname, users.display->>'mini',
			   gap.id, gap.name->>'text', gap.display->>'mini', follow_gap.created_at,
			   gap.username->>'text'
   		  FROM (SELECT * FROM follow_gap
			   WHERE gap_id = $1 AND status = true AND used = true AND created_at < $2
			   ORDER BY created_at DESC
			   LIMIT $3) as follow_gap
	 LEFT JOIN users
	 	  	ON users.id = follow_gap.owner_id
	 LEFT JOIN gap
	 	 	ON follow_gap.owner_id = gap.user_id
	  ORDER BY follow_gap.created_at DESC, cast(gap.count->>'follower' as integer) DESC;
		`, gapID, timeLoad, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lu []*entity.UserFollowerGapModel
	var uid string
	count := 0
	for rows.Next() {

		var u entity.UserFollowerGapModel
		var g entity.GapUserFollower
		var gID, gName, gDisplay, gUsername interface{}
		var firstname, lastname string
		err := rows.Scan(&u.ID, &firstname, &lastname, &u.Display, &gID, &gName, &gDisplay, &u.CreatedAt, &gUsername)
		if err != nil {
			return nil, err
		}

		if gID != nil {
			g.ID = gID.(string)
		}

		if gName != nil {
			g.Name = gName.(string)
		}

		if gUsername != nil {
			g.Username = gUsername.(string)
		}

		if gDisplay != nil {
			g.Display = gDisplay.(string)
		}

		if g.Username == "" {
			g.Username = g.ID
		}

		if u.ID != uid {
			uid = u.ID
			count = 0

			u.Name = firstname + " " + lastname

			if g.ID != "" {
				u.Gap = append(u.Gap, g)
			}

			lu = append(lu, &u)
			count++

			continue
		}

		if count < 4 {

			if g.ID != "" {
				lu[len(lu)-1].Gap = append(lu[len(lu)-1].Gap, g)
			}
			count++
			continue
		}

	}

	return lu, nil
}

//CheckTimeEditGapName check time
func CheckTimeEditGapName(q Queryer, gapID string, userID string) (string, error) {
	var t int64
	var name string
	err := q.QueryRow(`
		SELECT name->>'time', name->>'text'
		  FROM gap
		 WHERE id = $1 AND user_id = $2;
	`, gapID, userID).Scan(&t, &name)

	if err != nil {
		return "", err
	}

	duration := time.Now().Unix() - t

	if duration < config.ExpiredNameGapDuration {
		return "", fmt.Errorf("Can't edit name")
	}

	return name, nil
}

//CheckGapUsername check username gap return error, boolean
func CheckGapUsername(q Queryer, gapID string, username string) (bool, error) {

	var id string
	var swap bool
	var text string

	err := q.QueryRow(`
		SELECT id
		  FROM gap
		 WHERE username->>'text' = $1;
	`, username).Scan(&id)

	if err != nil && err != sql.ErrNoRows {
		return false, fmt.Errorf("Username มีในระบบแล้ว")
	}

	if id != "" {
		return false, fmt.Errorf("Username มีในระบบแล้ว")
	}

	err = q.QueryRow(`
		SELECT username->>'text', username->>'swap'
		  FROM gap
		 WHERE id = $1;
	`, gapID).Scan(&text, &swap)

	if err != nil {
		return false, fmt.Errorf("ไม่สามารถเปลี่ยน Username gap ได้")
	}

	if text == "" {
		return true, nil
	}

	if swap {
		return false, nil
	}

	return true, fmt.Errorf("ไม่สามารถเปลี่ยน Username gap ได้อีก")
}

//EditGapName edit gap name
func EditGapName(q Queryer, gapID string, userID string, name string) error {

	n := gapName{
		Text: name,
		Time: time.Now().Unix(),
	}

	_, err := q.Exec(`
		UPDATE gap
		   SET name = $1
		 WHERE id = $2 AND user_id = $3;
	`, convJSON(n), gapID, userID)

	return err

}

//EditGapBio edit gap name
func EditGapBio(q Queryer, gapID string, userID string, bio string) error {

	_, err := q.Exec(`
		UPDATE gap
		   SET bio = $1
		 WHERE id = $2 AND user_id = $3;
	`, bio, gapID, userID)

	return err

}

//EditGapUsername edit gap name
func EditGapUsername(q Queryer, gapID string, userID string, username string, swap bool) error {

	u := gapUsername{
		Text: username,
		Swap: swap,
	}

	_, err := q.Exec(`
		UPDATE gap
		   SET username = $1
		 WHERE id = $2 AND user_id = $3;
	`, convJSON(u), gapID, userID)

	return err

}

//EditGapDisplay edit gap display
func EditGapDisplay(q Queryer, gapID string, userID string, display DisplayGap) error {

	_, err := q.Exec(`
		UPDATE gap
		   SET display = $1
		 WHERE id = $2 AND user_id = $3;
	`, convJSON(display), gapID, userID)

	return err

}

//EditGapCover edit gap cover
func EditGapCover(q Queryer, gapID string, userID string, cover CoverGap) error {

	_, err := q.Exec(`
		UPDATE gap
		   SET cover = $1
		 WHERE id = $2 AND user_id = $3;
	`, convJSON(cover), gapID, userID)

	return err

}

//EditGapContact edit gap contact
func EditGapContact(q Queryer, gapID string, tel string, email string, social string) error {

	_, err := q.Exec(`
		UPDATE contact
		   SET tel = $1, email = $2, social = $3
		 WHERE owner_id = $4 AND owner_type = 1;
	`, tel, email, social, gapID)

	return err
}

//EditGapAddress edit gap address
func EditGapAddress(q Queryer, gapID string, address string, city string, country string) error {

	_, err := q.Exec(`
		UPDATE contact
		   SET address = $1, city = $2, country = $3
		 WHERE owner_id = $4 AND owner_type = 1;
	`, address, city, country, gapID)

	return err

}

//GetGapView return time.Unix
func GetGapView(q Queryer, gapID string, userID string) (time.Time, error) {

	var t time.Time
	err := q.QueryRow(`
				SELECT created_at
				FROM public.gap_view
				WHERE gap_id = $1
				AND owner_id = $2
				ORDER BY created_at DESC;
				`, gapID, userID).Scan(&t)
	if err != nil {
		return t, err
	}

	return t, nil
}

//GetGapGuestView return time.Unix
func GetGapGuestView(q Queryer, gapID string, vsID string) (time.Time, error) {

	var t time.Time
	err := q.QueryRow(`
				SELECT created_at
				FROM public.gap_guest
				WHERE gap_id = $1
				AND visitor_id = $2
				ORDER BY created_at DESC;
				`, gapID, vsID).Scan(&t)
	if err != nil {
		return t, err
	}

	return t, nil
}

//CreateViewGap create new row view gap
func CreateViewGap(q Queryer, gapID string, userID string, referrer string, userAgent string) error {

	_, err := q.Exec(`
					INSERT INTO public.gap_view
					(gap_id, owner_id, referrer, user_agent)
					VALUES ($1, $2, $3, $4);
					`, gapID, userID, referrer, userAgent)
	if err != nil {
		return err
	}

	return nil
}

//CreateGuestViewGap create new row guest gap
func CreateGuestViewGap(q Queryer, gapID string, vsID string, referrer string, userAgent string) error {

	_, err := q.Exec(`
					INSERT INTO public.gap_guest
					(gap_id, visitor_id, referrer, user_agent)
					VALUES ($1, $2, $3, $4);
					`, gapID, vsID, referrer, userAgent)
	if err != nil {
		return err
	}

	return nil
}

//UpdateCountPopularGap update popular count
func UpdateCountPopularGap(q Queryer, gapID string) error {

	_, err := q.Exec(`
					UPDATE public.gap
					   SET count = count::jsonb || CONCAT('{"popular":', COALESCE(count->>'popular','0')::int + 1, '}')::jsonb
					 WHERE id = $1;
					`, gapID)
	if err != nil {
		return err
	}

	return nil
}

//UpdateCountViewGap update view count
func UpdateCountViewGap(q Queryer, gapID string) error {

	_, err := q.Exec(`
						UPDATE public.gap
						SET count = count::jsonb || CONCAT('{"view":', COALESCE(count->>'view','0')::int + 1, '}')::jsonb
						 WHERE id = $1;
						`, gapID)
	if err != nil {
		return err
	}

	return nil
}

//UpdateCountGuestViewGap update view count
func UpdateCountGuestViewGap(q Queryer, gapID string) error {

	_, err := q.Exec(`
						UPDATE public.gap
						   SET guest_view_count = guest_view_count + 1
						 WHERE id = $1;
						`, gapID)
	if err != nil {
		return err
	}

	return nil
}

//ListGapCountUserAgentView return model guestView sum count userAgent
func ListGapCountUserAgentView(q Queryer, gapID string, a, b time.Time) (*entity.CountUserAgent, error) {

	rows, err := q.Query(`
	   SELECT gap_view.user_agent, count(DISTINCT gap_view.*)
		 FROM (SELECT id FROM gap WHERE id = $1) as gap
	LEFT JOIN gap_view
	 	   ON gap_view.gap_id = gap.id
		WHERE gap_view.created_at >= $2 and gap_view.created_at <= $3
	GROUP BY gap_view.user_agent
		`, gapID, a, b)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ag entity.CountUserAgent

	for rows.Next() {
		var userAgent string
		var count int
		err := rows.Scan(&userAgent, &count)
		if err != nil {
			return nil, err
		}

		us := user_agent.New(userAgent)
		if us.Mobile() {
			ag.Mobile += count
			continue
		}

		ag.Desktop += count
	}

	return &ag, nil
}

//ListGapCountUserAgentGuestView return model guestView sum count userAgent
func ListGapCountUserAgentGuestView(q Queryer, gapID string, a, b time.Time) (*entity.CountUserAgent, error) {

	rows, err := q.Query(`
	   SELECT gap_guest.user_agent, count(DISTINCT gap_guest.*)
		 FROM (SELECT id FROM gap WHERE id = $1) as gap
	LEFT JOIN gap_guest
	 	   ON gap_guest.gap_id = gap.id
		WHERE gap_guest.created_at >= $2 and gap_guest.created_at <= $3
	GROUP BY gap_guest.user_agent
		`, gapID, a, b)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ag entity.CountUserAgent

	for rows.Next() {
		var userAgent string
		var count int
		err := rows.Scan(&userAgent, &count)
		if err != nil {
			return nil, err
		}

		us := user_agent.New(userAgent)
		if us.Mobile() {
			ag.Mobile += count
			continue
		}

		ag.Desktop += count
	}

	return &ag, nil
}

//ListCountGapViewHour return view hour
func ListCountGapViewHour(q Queryer, gapID string, a, b time.Time) ([]*entity.CountViewHour, error) {

	rows, err := q.Query(`
	  SELECT COALESCE(extract(hour from gap_view.created_at at time zone 'UTC+7' at time zone 'UTC'), 0) as date, count(DISTINCT  gap_view.*)
	    FROM (SELECT id FROM gap WHERE id = $1) as gap
   LEFT JOIN gap_view
		  ON gap_view.gap_id = gap.id AND gap_view.created_at >= $2 AND gap_view.created_at <= $3
	GROUP BY date
		`, gapID, a, b)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lc []*entity.CountViewHour

	for rows.Next() {

		var c entity.CountViewHour
		err := rows.Scan(&c.Hour, &c.Count)
		if err != nil {
			return nil, err
		}

		lc = append(lc, &c)
	}

	return lc, nil
}

//ListCountGapGuestViewHour return view hour
func ListCountGapGuestViewHour(q Queryer, gapID string, a, b time.Time) ([]*entity.CountViewHour, error) {

	rows, err := q.Query(`
	 	SELECT COALESCE(extract(hour from gap_guest.created_at at time zone 'UTC+7' at time zone 'UTC'), 0) as date, count(DISTINCT  gap_guest.*)
	      FROM (SELECT id FROM gap WHERE id = $1) as gap
     LEFT JOIN gap_guest
		    ON gap_guest.gap_id = gap.id AND gap_guest.created_at >= $2 AND gap_guest.created_at <= $3
	  GROUP BY date
		`, gapID, a, b)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lc []*entity.CountViewHour

	for rows.Next() {

		var c entity.CountViewHour
		err := rows.Scan(&c.Hour, &c.Count)
		if err != nil {
			return nil, err
		}

		lc = append(lc, &c)
	}

	return lc, nil
}

// //GetCountViewGap return model view and guestView sum
// func GetCountViewGap(q Queryer, gapID string, a, b time.Time) (*entity.GetViewModel, error) {

// 	var v entity.GetViewModel

// 	err := q.QueryRow(`
// 		SELECT count(gap_view.id)
// 		  FROM (SELECT id FROM gap WHERE user_id = $1) as gap
// 	 LEFT JOIN gap_view
// 	 		ON gap_view.gap_id = gap.id
// 		 WHERE gap_view.created_at >= $2 and gap_view.created_at <= $3
// 		`, userID, a, b).Scan(&v.View)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = q.QueryRow(`
// 		SELECT count(gap_guest.id)
// 		  FROM (SELECT id FROM gap WHERE user_id = $1) as gap
// 	 LEFT JOIN gap_guest
// 	 		ON gap_guest.gap_id = gap.id
// 			 WHERE gap_guest.created_at >= $2 and gap_guest.created_at <= $3
// 		`, userID, a, b).Scan(&v.GuestView)
// 	if err != nil {
// 		return nil, err
// 	}

// 	v.All = v.Sum()

// 	return &v, nil
// }

//CreateGapModal is struct modal
type CreateGapModal struct {
	UserID  string
	Name    NameGap
	Bio     string
	TopicID string
	Display DisplayGap
	Cover   CoverGap
}

//NameGap gap name
type NameGap struct {
	Text string `json:"text"`
	Time string `json:"time"`
}

// Contact data in struct
type contactGap struct {
	Website string `json:"website"`
	City    string `json:"city"`
	Tel     string `json:"tel"`
	Address string `json:"address"`
	Country string `json:"country"`
	Social  string `json:"social"`
	Email   string `json:"email"`
}

type countGap struct {
	Popular  int `json:"popular"`
	Follower int `json:"follower"`
	Post     int `json:"post"`
	View     int `json:"view"`
}

//CoverGap model
type CoverGap struct {
	Mini   string `json:"mini"`
	Normal string `json:"normal"`
}

//DisplayGap model
type DisplayGap struct {
	Mini   string `json:"mini"`
	Middle string `json:"middle"`
	Normal string `json:"normal"`
}

type statusGap struct {
	CreatedAt string `json:"created_at"`
	ExpreAt   string `json:"expre_at"`
	Level     int    `json:"gap_level"`
}

// username Gap
type usernameGap struct {
	Text string `json:"text"`
	Swap bool   `json:"swap"`
}

// verify Gap
type verifyGap struct {
	Status bool `json:"status"`
	Level  int  `json:"level"`
}

// gapName Gap
type gapName struct {
	Text string `json:"text"`
	Time int64  `json:"time"`
}

// gapName Gap
type gapUsername struct {
	Text string `json:"text"`
	Swap bool   `json:"swap"`
}
