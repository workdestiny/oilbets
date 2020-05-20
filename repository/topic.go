package repository

import (
	"bytes"
	"encoding/gob"
	"strings"

	"github.com/go-redis/redis"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
)

//ListCategoryDiscover list category and topic
func ListCategoryDiscover(q Queryer) ([]*entity.CategoryDiscoverList, error) {
	rows, err := q.Query(`
			SELECT category.id, category.code, category.name->>'th'
			  FROM category
			 WHERE verify = true
		  ORDER BY random();
		`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ct []*entity.CategoryDiscoverList

	for rows.Next() {
		var c entity.CategoryDiscoverList
		err := rows.Scan(&c.ID, &c.Code, &c.Name)
		if err != nil {
			return nil, err
		}

		ct = append(ct, &c)
	}

	return ct, nil
}

//ListTopicVerified list count DESC
func ListTopicVerified(q Queryer, role entity.Role) ([]*entity.TopicList, error) {

	r := config.TopicOfficialID

	if role == entity.RoleAdmin {
		r = ""
	}

	rows, err := q.Query(`
			SELECT id, code, name->>'th', used_count, title
			  FROM topic
			 WHERE id != $1 AND verify = true
		  ORDER BY used_count, created_at DESC;
		`, r)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lt []*entity.TopicList

	for rows.Next() {

		var t entity.TopicList
		err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.UsedCount, &t.Title)
		if err != nil {
			return nil, err
		}

		lt = append(lt, &t)
	}

	return lt, nil
}

//ListTopicVerifiedByCategory list count DESC
func ListTopicVerifiedByCategory(q Queryer, catID string) ([]*entity.TopicList, error) {

	rows, err := q.Query(`
			SELECT id, code, name->>'th'
			  FROM public.topic
			 WHERE cat_id = $1 AND verify = true
		  ORDER BY used_count DESC;
		`, catID)
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

//ListTopicIsDataByCategory list count DESC
func ListTopicIsDataByCategory(q Queryer, catID string) ([]*entity.TopicList, error) {

	rows, err := q.Query(`
			SELECT id, code, name->>'th'
			  FROM public.topic
			 WHERE cat_id = $1 AND verify = true AND used_count != 0
		  ORDER BY used_count DESC;
		`, catID)
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

//ListFollowTopicID list count DESC
func ListFollowTopicID(q Queryer, userID string) []string {

	rows, err := q.Query(`
			SELECT topic_id
			  FROM public.follow_topic
			 WHERE owner_id = $1 AND status = true AND used = true;
		`, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var topicIDs []string

	for rows.Next() {

		var topicID string
		err := rows.Scan(&topicID)
		if err != nil {
			return nil
		}

		topicIDs = append(topicIDs, topicID)
	}

	return topicIDs
}

//ListTopicIDVerified list count DESC
func ListTopicIDVerified(q Queryer) []string {

	rows, err := q.Query(`
			SELECT id
			  FROM topic
			 WHERE verify = true
		  ORDER BY count DESC;
		`)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var ids []string

	for rows.Next() {

		var id string
		err := rows.Scan(&id)
		if err != nil {
			return []string{}
		}

		ids = append(ids, id)
	}

	return ids
}

//SearchTagTopic input text Search in redis name Topic (Ajax)
//ต้องตัดช่องว่าง และ คำที่จะค้นหาต้องเป็นตัวเล็กทั้งหมด
func SearchTagTopic(rd *redis.Client, text string, role entity.Role, limit int) ([]*entity.TopicList, error) {

	var cursor uint64

	scanText := "*:" + text + "*"

	// ค้นหา TagTopic ใน Redis
	values, _, err := rd.SScan(config.RedisIndexTopicName, cursor, scanText, config.RedisCount).Result()
	//values, err := redis.Values(c.Do("SSCAN", config.RedisIndexTopicName, cursor, "match", scanText, "count", config.RedisCount))
	if err != nil {
		return nil, err
	}

	var lt []*entity.TopicList
	for i := 0; i < len(values); i++ {

		if len(lt) == limit {
			break
		}

		va := strings.Split(values[i], ":")

		arrStringTopic, err := rd.Keys(config.RedisTopicAny + "*" + ":" + va[0] + ":*").Result()
		//arrStringTopic, err := redis.Strings(c.Do("KEYS", config.RedisTopicAny+"*"+":"+values[0]+":*"))
		if err != nil {
			return nil, err
		}

		if len(arrStringTopic) > 0 {

			var t entity.RedisTopicModel
			bytesPage, err := rd.Get(arrStringTopic[0]).Bytes()
			//bytesPage, err := redis.Bytes(c.Do("GET", arrStringTopic[0]))
			if err != nil {
				return nil, err
			}

			gob.NewDecoder(bytes.NewReader(bytesPage)).Decode(&t)

			if t.ID == config.TopicOfficialID {
				if role < entity.RoleAdmin {
					continue
				}
			}

			if t.ID == "" {
				continue
			}

			lt = append(lt, &entity.TopicList{
				ID:        t.ID,
				Code:      t.Code,
				Name:      t.Name,
				UsedCount: t.UsedCount,
			})
		}
	}

	for i := 0; i < len(lt); i++ {
		for j := 0; j < len(lt)-1; j++ {
			if lt[j].UsedCount < lt[j+1].UsedCount {
				// swap
				lt[j], lt[j+1] = lt[j+1], lt[j]
			}
		}
	}

	return lt, nil
}

//GetTopic get topic
func GetTopic(q Queryer, topicID string) (*entity.TopicModel, error) {

	var t entity.TopicModel
	err := q.QueryRow(`
		SELECT id, cat_id, code, name->>'th',
			   images->>'mini', images->>'normal', count, used_count,
			   verify
		  FROM public.topic
		 WHERE id = $1;
		 `, topicID).Scan(&t.ID, &t.CatID, &t.Code, &t.Name,
		&t.ImageMini, &t.Image, &t.Count, &t.UsedCount,
		&t.Verify)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

//GetTopicByCode get topic
func GetTopicByCode(q Queryer, code string) (*entity.TopicModel, error) {

	var t entity.TopicModel
	err := q.QueryRow(`
		SELECT id, cat_id, code, name->>'th',
			   images->>'mini', images->>'normal', count, used_count,
			   verify, title, description, image_fb,
			   tagline
		  FROM public.topic
		 WHERE code = $1;
		 `, code).Scan(&t.ID, &t.CatID, &t.Code, &t.Name,
		&t.ImageMini, &t.Image, &t.Count, &t.UsedCount,
		&t.Verify, &t.Title, &t.Description, &t.ImageFacebook,
		&t.Tagline)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

//GetCategoryIDByCode get cat id
func GetCategoryIDByCode(q Queryer, code string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM category
		 WHERE code = $1;
		 `, code).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

//CheckTopicID check is id
func CheckTopicID(q Queryer, topicID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM public.topic
		 WHERE id = $1;
	`, topicID).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

//CheckCategory check is category
func CheckCategory(q Queryer, catID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM category
		 WHERE id = $1;
	`, catID).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

//CheckTopicName check is name
func CheckTopicName(q Queryer, name string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM public.topic
		 WHERE name->>'th' = $1;
	`, name).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

//CheckTopicCode check is code
func CheckTopicCode(q Queryer, code string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM public.topic
		 WHERE code = $1;
	`, code).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

//CreateTopic new topic
func CreateTopic(q Queryer, name string) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO public.topic
					(cat_id, code, count, images,
					name, used_count, verify)
		     VALUES ($1, $2, $3, $4,
			        $5, $6, $7)
		  RETURNING id;
					`, config.CategoryOtherID, name, 0,
		convJSON(ImageTopic{}), convJSON(NameTopic{
			Th: name,
			En: name,
		}), 0,
		false).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

//UpdateUsedCountTopic update count
func UpdateUsedCountTopic(q Queryer, topicID string) error {

	_, err := q.Exec(`
		UPDATE public.topic
		   SET used_count = used_count + 1
		 WHERE id = $1;
	`, topicID)

	return err
}

//CheckFollowTopic ตรวจสอบว่าเคย follow แล้วหรือยังถ้ายัง
func CheckFollowTopic(q Queryer, userID string, topicID string) (bool, error) {

	var isFollow bool
	err := q.QueryRow(`
		SELECT status
		  FROM public.follow_topic
		 WHERE topic_id = $1
		   AND owner_id = $2
		   AND used = true;
	`, topicID, userID).Scan(&isFollow)

	return isFollow, err
}

//CreateFollowTopic (ต้องใช้ Tx) สร้าง follow ที่ไม่เคยมีขึ้นมาใหม่ หลังจากสร้างสำเร็จจะทำการ update ค่า count following topic
func CreateFollowTopic(q Queryer, userID string, topicID string) error {

	_, err := q.Exec(`
		INSERT INTO public.follow_topic
					(topic_id, owner_id, status, used)
			 VALUES ($1, $2, $3, $4);
	`, topicID, userID, true, true)
	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE public.topic
		   SET count = count + 1
		 WHERE id = $1;
	`, topicID)
	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE users
		   SET count = count::jsonb || CONCAT('{"topic":', COALESCE(count->>'topic','0')::int + 1, '}')::jsonb
		 WHERE id = $1;
	`, userID)
	if err != nil {
		return err
	}

	return nil
}

//CreateFollowTopicXOfficial follow ที่ไม่เคยมีขึ้นมาใหม่
func CreateFollowTopicXOfficial(q Queryer, userID string) error {

	_, err := q.Exec(`
		INSERT INTO public.follow_topic
					(topic_id, owner_id, status, used)
			 VALUES ($1, $2, $3, $4) RETURNING id;
	`, config.TopicOfficialID, userID, true, true)
	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE public.topic
		   SET count = count + 1
		 WHERE id = $1;
	`, config.TopicOfficialID)
	if err != nil {
		return err
	}

	return nil
}

//CreateFollowGapXOfficial (ต้องใช้ Tx) สร้าง follow ที่ไม่เคยมีขึ้นมาใหม่ หลังจากสร้างสำเร็จจะทำการ update ค่า count following topic
func CreateFollowGapXOfficial(q Queryer, userID string) error {

	_, err := q.Exec(`
		INSERT INTO public.follow_gap
					(gap_id, owner_id, status, used)
			 VALUES ($1, $2, $3, $4) RETURNING id;
	`, config.GapOfficialID, userID, true, true)
	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE public.gap
		   SET count = count::jsonb || CONCAT('{"follower":', COALESCE(count->>'follower','0')::int + 1, '}')::jsonb
		 WHERE id = $1;
	`, config.GapOfficialID)
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

//ListTopicFollow list count DESC
func ListTopicFollow(q Queryer, userID string, role entity.Role) ([]*entity.TopicList, error) {

	r := config.TopicOfficialID

	if role == entity.RoleAdmin {
		r = ""
	}

	rows, err := q.Query(`
			SELECT public.topic.id, public.topic.code, public.topic.name->>'th', public.topic.used_count
			  FROM public.topic
		 LEFT JOIN follow_topic
			    ON follow_topic.owner_id = $1 AND status = true AND used = true
			 WHERE public.topic.id != $2 AND public.topic.verify = true AND follow_topic.topic_id = public.topic.id
		  ORDER BY count DESC;
		`, userID, r)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lt []*entity.TopicList

	for rows.Next() {

		var t entity.TopicList
		err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.UsedCount)
		if err != nil {
			return nil, err
		}

		lt = append(lt, &t)
	}

	return lt, nil
}

//FollowTopic (isFollow คือ follow ที่ user ต้องการให้เป็นเช่น user ต้องการ follow topic นี้ isFollow = true)
//เรียกใช้เมื่อ user นี้ เคย follow topic นี้แล้ว จำทำการเปลี่ยนจาก follow เป็น unfollow หรือ unfollow เป็น follow
func FollowTopic(q Queryer, userID string, topicID string, isFollow bool) error {

	_, err := q.Exec(`
		UPDATE public.follow_topic
		   SET status = $1
		 WHERE topic_id = $2
		   AND owner_id = $3
		   AND used = true;
	`, isFollow, topicID, userID)

	set := "+ 1"
	if !isFollow {
		set = "- 1"
	}

	_, err = q.Exec(`
		UPDATE public.topic
		   SET count = count `+set+`
		 WHERE id = $1;
	`, topicID)

	_, err = q.Exec(`
		UPDATE users
		SET count = count::jsonb || CONCAT('{"topic":', COALESCE(count->>'topic','0')::int `+set+`, '}')::jsonb
		 WHERE id = $1;
	`, userID)
	if err != nil {
		return err
	}

	return err
}

//ListCategory list category and topic
func ListCategory(q Queryer, userID string, role entity.Role) ([]*entity.CategoryList, error) {

	r := config.TopicOfficialID

	if role == entity.RoleAdmin {
		r = ""
	}

	rows, err := q.Query(`
			SELECT public.topic.id, public.topic.code, public.topic.name->>'th', public.topic.count,
				   category.id, category.code, category.count, category.name->>'th',
				   follow_topic.status, public.topic.images->>'mini'
			  FROM public.topic
		 LEFT JOIN follow_topic
				ON follow_topic.owner_id = $1 AND follow_topic.topic_id = public.topic.id AND used = true
		 LEFT JOIN category
		 		ON public.topic.cat_id = category.id
			 WHERE public.topic.id != $2 AND public.topic.verify = true
		  ORDER BY public.topic.count DESC;
		`, userID, r)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ct []*entity.CategoryList
	var isPull bool

	for rows.Next() {

		var t entity.TopicListModel
		var c entity.CategoryList
		var follow interface{}
		err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Count,
			&c.ID, &c.Code, &c.Count, &c.Name,
			&follow, &t.Image)
		if err != nil {
			return nil, err
		}

		if follow != nil {
			t.IsFollow = follow.(bool)
		}

		for _, v := range ct {
			if v.ID == c.ID {
				v.Topic = append(v.Topic, t)
				isPull = true
				break
			}
		}

		if !isPull {
			c.Topic = append(c.Topic, t)
			ct = append(ct, &c)
		}
		isPull = false

	}

	return ct, nil
}

//ListCategoryAll list category verify
func ListCategoryAll(q Queryer) ([]*entity.CategoryList, error) {

	rows, err := q.Query(`
			SELECT category.id, category.code, category.count, category.name->>'th'
			  FROM category
		  ORDER BY category.count DESC;
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

//ListTagTopicIDOnPost list count DESC
func ListTagTopicIDOnPost(q Queryer, postID string) []string {

	rows, err := q.Query(`
			SELECT topic_id
			  FROM tagtopic
			 WHERE post_id = $1;
		`, postID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var topicIDs []string

	for rows.Next() {

		var topicID string
		err := rows.Scan(&topicID)
		if err != nil {
			return nil
		}

		topicIDs = append(topicIDs, topicID)
	}

	return topicIDs
}

//TopicModel model to DB
type TopicModel struct {
	ID        string
	CatID     int64
	Code      string
	Name      NameTopic
	Images    []ImageTopic
	Count     int
	UsedCount int
	Verify    bool
}

//NameTopic model name topic
type NameTopic struct {
	Th string `json:"th"`
	En string `json:"en"`
}

//ImageTopic model Image topic
type ImageTopic struct {
	Normal string `json:"normal"`
	Mini   string `json:"mini"`
}
