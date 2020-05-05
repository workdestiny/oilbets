package repository

import (
	"bytes"
	"encoding/gob"
	"strings"
	"unicode/utf8"

	"github.com/go-redis/redis"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
)

// RedisUserModel is Model User in Redis
type RedisUserModel struct {
	ID           string
	Username     string
	FirstName    string
	LastName     string
	DisplayImage string
	Level        entity.UserLevelType
}

// RedisGapModel is Model Gap in Redis
type RedisGapModel struct {
	ID            int64
	UserID        int64
	Username      string
	Name          string
	DisplayImage  string
	CountFollower int
	CountPopular  int
}

// RedisCategoryModel is Model Category in Redis
type RedisCategoryModel struct {
	ID     int64
	Code   string
	Name   Name
	Images []Image
	Count  int
}

// RedisTopicModel is Model Topic in Redis
type RedisTopicModel struct {
	ID        int64
	CatID     int64
	Code      string
	Name      Name
	Images    []Image
	Count     int
	UsedCount int
}

// Name Category and Topic
type Name struct {
	Th string
	En string
}

// Image Category and Topic
type Image struct {
	Normal string
	Color  string
}

func getRedisUser(rd *redis.Client, id string) entity.RedisUserModel {

	var u entity.RedisUserModel
	uBytes, _ := rd.Get(config.RedisUser + id).Bytes()
	//uBytes, _ := redis.Bytes(c.Do("GET", config.RedisUser+id))
	gob.NewDecoder(bytes.NewReader(uBytes)).Decode(&u)

	return u
}

func getRedisGap(rd *redis.Client, id string) entity.RedisGapModel {

	arrStringGap, _ := rd.Keys(config.RedisGap + "*" + ":" + id).Result()
	//arrStringGap, _ := redis.Strings(c.Do("KEYS", config.RedisGap+"*"+":"+id))

	var g entity.RedisGapModel

	if len(arrStringGap) != 0 {

		bytesGap, _ := rd.Get(arrStringGap[0]).Bytes()
		//bytesGap, _ := redis.Bytes(c.Do("GET", arrStringGap[0]))
		gob.NewDecoder(bytes.NewReader(bytesGap)).Decode(&g)

	}

	return g
}

// AddUserToRedis add user data to redis
func AddUserToRedis(rd *redis.Client, user entity.RedisUserModel) {

	var incr func() error

	incr = func() error {

		keyUser := config.RedisUser + user.ID
		buf := bytes.Buffer{}
		gob.NewEncoder(&buf).Encode(&user)

		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.SAdd(config.RedisIndexUserFirstName, user.ID+":"+strings.ToLower(user.FirstName), 0)
				pipe.SAdd(config.RedisIndexUserLastName, user.ID+":"+strings.ToLower(user.LastName), 0)
				pipe.Set(config.RedisUser+user.ID, buf.Bytes(), 0)
				return nil
			})
			return err
		}, config.RedisIndexUserFirstName, config.RedisIndexUserLastName, keyUser)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

// SetUserToRedis set user data to redis
func SetUserToRedis(rd *redis.Client, user entity.RedisUserModel) {

	var incr func() error

	incr = func() error {

		keyUser := config.RedisUser + user.ID
		buf := bytes.Buffer{}
		gob.NewEncoder(&buf).Encode(&user)

		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.Set(config.RedisUser+user.ID, buf.Bytes(), 0)
				return nil
			})
			return err
		}, keyUser)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

// // UpdateUserToRedis update user data to redis
// func UpdateUserToRedis(myRedis *redis.Client, id int64, username string, firstname string, lastname string, displayImage string, auditFirstname string, auditLastname string, userVerifyLevel entity.UserVerifyLevel) {

// 	c := myRedis.Get()
// 	defer c.Close()

// 	var userRedis RedisUserModel
// 	userRedis.ID = id
// 	userRedis.Username = username
// 	userRedis.FirstName = firstname
// 	userRedis.LastName = lastname
// 	userRedis.DisplayImage = displayImage
// 	userRedis.VerifyLevel = userVerifyLevel
// 	//userRedis.CountFollower = countFollower

// 	sID := strconv.FormatInt(userRedis.ID, 10)

// 	for i := 0; i < 5; i++ {

// 		ok, err := redis.String(c.Do("MULTI"))
// 		if err != nil {
// 			break
// 		}
// 		if ok == "OK" {
// 			c.Do("SREM", config.RedisIndexUserFirstName, sID+":"+strings.ToLower(auditFirstname))
// 			c.Do("SREM", config.RedisIndexUserLastName, sID+":"+strings.ToLower(auditLastname))
// 			c.Do("SADD", config.RedisIndexUserFirstName, sID+":"+strings.ToLower(userRedis.FirstName))
// 			c.Do("SADD", config.RedisIndexUserLastName, sID+":"+strings.ToLower(userRedis.LastName))

// 			buf := bytes.Buffer{}
// 			gob.NewEncoder(&buf).Encode(&userRedis)
// 			c.Do("SET", config.RedisUser+sID, buf.Bytes())
// 		}
// 		status, err := redis.Values(c.Do("EXEC"))
// 		if err != nil {
// 			break
// 		}
// 		if len(status) > 0 {
// 			if status[len(status)-1] == "OK" {
// 				break
// 			}
// 		}

// 	}

// }

// // UpdateGapToRedis update Gap data to Redis
// func UpdateGapToRedis(myRedis *redis.Client, id int64, username string, name string, displayImage string, UserID int64, countFollower int, countPopular int, auditName string, userID int64) {

// 	c := myRedis.Get()
// 	defer c.Close()

// 	var gapRedis RedisGapModel
// 	gapRedis.ID = id
// 	gapRedis.Username = username
// 	gapRedis.Name = name
// 	gapRedis.UserID = userID
// 	gapRedis.DisplayImage = displayImage
// 	gapRedis.CountFollower = countFollower
// 	gapRedis.CountPopular = countPopular

// 	sID := strconv.FormatInt(gapRedis.ID, 10)
// 	sUserID := strconv.FormatInt(UserID, 10)

// 	for i := 0; i < 5; i++ {

// 		ok, err := redis.String(c.Do("MULTI"))
// 		if err != nil {
// 			return
// 		}
// 		if ok == "OK" {
// 			if len(auditName) > 0 {
// 				c.Do("SREM", config.RedisIndexPageName, sID+":"+strings.ToLower(auditName))
// 			}
// 			c.Do("SADD", config.RedisIndexPageName, sID+":"+strings.ToLower(gapRedis.Name))

// 			buf := bytes.Buffer{}
// 			gob.NewEncoder(&buf).Encode(&gapRedis)
// 			c.Do("SET", config.RedisPage+sUserID+":"+sID, buf.Bytes())
// 		}
// 		status, err := redis.Values(c.Do("EXEC"))
// 		if err != nil {
// 			return
// 		}
// 		if len(status) > 0 {
// 			if status[len(status)-1] == "OK" {
// 				break
// 			}
// 		}
// 	}

// }

//AddNewGapToRedis add new gap
func AddNewGapToRedis(rd *redis.Client, rg entity.RedisGapModel) {

	buf := bytes.Buffer{}
	gob.NewEncoder(&buf).Encode(&rg)

	rd.Set(config.RedisGap+rg.UserID+":"+rg.ID, buf.Bytes(), 0)
	//c.Do("SET", config.RedisGap+rg.UserID+":"+rg.ID, buf.Bytes())
	rd.SAdd(config.RedisIndexGapName, rg.ID+":"+strings.ToLower(rg.Name))
	//c.Do("SADD", config.RedisIndexGapName, rg.ID+":"+strings.ToLower(rg.Name))
}

// SetGapToRedis set gap data to redis
func SetGapToRedis(rd *redis.Client, gap entity.RedisGapModel, nameAudit string) {

	var incr func() error

	incr = func() error {

		keyGap := config.RedisGap + gap.UserID + ":" + gap.ID
		buf := bytes.Buffer{}
		gob.NewEncoder(&buf).Encode(&gap)

		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.SRem(config.RedisIndexGapName, gap.ID+":"+strings.ToLower(nameAudit))
				pipe.SAdd(config.RedisIndexGapName, gap.ID+":"+strings.ToLower(gap.Name), 0)
				pipe.Set(keyGap, buf.Bytes(), 0)
				return nil
			})
			return err
		}, config.RedisIndexGapName, keyGap)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

//SetTopicNotVerifyToRedis set TopicNotVerify to redis
func SetTopicNotVerifyToRedis(rd *redis.Client, CatID string, TopicID string, TopicCode string, buf []byte) {

	var incr func() error

	incr = func() error {

		keyTopic := config.RedisTopicNotVerify + CatID + ":" + TopicID + ":" + TopicCode

		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.Set(keyTopic, buf, 0)
				return nil
			})
			return err
		}, keyTopic)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

// SetTopicVerifyToRedis set TopicVerify to redis
func SetTopicVerifyToRedis(rd *redis.Client, CatID string, TopicID string, TopicCode string, buf []byte) {
	var incr func() error
	incr = func() error {
		keyTopic := config.RedisTopic + CatID + ":" + TopicID + ":" + TopicCode

		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.Set(keyTopic, buf, 0)
				return nil
			})
			return err
		}, keyTopic)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

// DeleteTopicNotVerifyToRedis dalete topic notverify to redis
func DeleteTopicNotVerifyToRedis(rd *redis.Client, CatID string, TopicID string, TopicCode string) {

	rd.Del(config.RedisTopicNotVerify + CatID + ":" + TopicID + ":" + TopicCode)
	//c.Do("DEL", config.RedisTopicNotVerify+CatID+":"+TopicID+":"+TopicCode)

}

// AdminDeleteTopicVerifyToRedis dalete (set and string) topic verify to redis
func AdminDeleteTopicVerifyToRedis(rd *redis.Client, CatID string, TopicID string, TopicCode string) {

	rd.Del(config.RedisTopic + CatID + ":" + TopicID + ":" + TopicCode)
	rd.SRem(config.RedisIndexTopicName, TopicID+":"+TopicCode)
}

// AdminDeleteTopicNotVerifyToRedis dalete (set and string) topic not verify to redis
func AdminDeleteTopicNotVerifyToRedis(rd *redis.Client, CatID string, TopicID string, TopicCode string) {

	rd.Del(config.RedisTopicNotVerify + CatID + ":" + TopicID + ":" + TopicCode)
	rd.SRem(config.RedisIndexTopicName, TopicID+":"+TopicCode)
}

// AdminAddTopicVerifyToRedis add Topic Verify to redis
func AdminAddTopicVerifyToRedis(rd *redis.Client, t entity.RedisTopicModel) {

	buf := bytes.Buffer{}
	gob.NewEncoder(&buf).Encode(&t)

	var incr func() error
	incr = func() error {

		keyTopic := config.RedisTopic + t.CatID + ":" + t.ID + ":" + t.Code
		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.SAdd(config.RedisIndexTopicName, t.ID+":"+strings.ToLower(t.Code), 0)
				pipe.Set(keyTopic, buf.Bytes(), 0)
				return nil
			})
			return err
		}, config.RedisIndexTopicName, keyTopic)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

// AdminAddTopicNotVerifyToRedis add Topic not Verify to redis
func AdminAddTopicNotVerifyToRedis(rd *redis.Client, t entity.RedisTopicModel) {

	buf := bytes.Buffer{}
	gob.NewEncoder(&buf).Encode(&t)

	var incr func() error
	incr = func() error {

		keyTopic := config.RedisTopicNotVerify + t.CatID + ":" + t.ID + ":" + t.Code
		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.SAdd(config.RedisIndexTopicName, t.ID+":"+strings.ToLower(t.Code), 0)
				pipe.Set(keyTopic, buf.Bytes(), 0)
				return nil
			})
			return err
		}, config.RedisIndexTopicName, keyTopic)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

// AddTopicNotVerifyToRedis add TopicNotVerify to redis
func AddTopicNotVerifyToRedis(rd *redis.Client, CatID string, TopicID string, TopicCode string, name string, buf []byte) {

	var incr func() error
	incr = func() error {

		keyTopic := config.RedisTopicNotVerify + CatID + ":" + TopicID + ":" + TopicCode
		err := rd.Watch(func(tx *redis.Tx) error {
			_, err := tx.Pipelined(func(pipe redis.Pipeliner) error {
				pipe.SAdd(config.RedisIndexTopicName, TopicID+":"+strings.ToLower(name), 0)
				pipe.Set(keyTopic, buf, 0)
				return nil
			})
			return err
		}, config.RedisIndexTopicName, keyTopic)
		if err == redis.TxFailedErr {
			return incr()
		}
		return err
	}
	incr()
}

var (
	cursor int64
	//GapNameitems select
	GapNameitems []string
	//TopicNameitems select
	TopicNameitems []string
	scanText       string
)

//SearchTopic Search the tagTopic
func SearchTopic(rd *redis.Client, text string) ([]*entity.RedisTopicModel, error) {

	scanText := "*:" + text + "*"
	if utf8.RuneCountInString(text) >= 3 {
		scanText = "*:*" + text + "*"
	}

	var cursor uint64

	// ค้นหา Topic ใน Redis
	values, _, err := rd.SScan(config.RedisIndexTopicName, cursor, scanText, config.RedisCount).Result()
	//values, err := redis.Values(c.Do("SSCAN", config.RedisIndexTopicName, cursor, "match", scanText, "count", config.RedisCount))
	if err != nil {
		return nil, err
	}

	var topics []*entity.RedisTopicModel
	for i := 0; i < len(values); i++ {

		va := strings.Split(values[i], ":")

		arrStringTopic, err := rd.Keys(config.RedisTopicAny + "*:" + va[0] + ":*").Result()
		//arrStringTopic, err := redis.Strings(c.Do("KEYS", config.RedisTopicAny+"*"+":"+values[0]+":*"))
		if err != nil {
			return nil, err
		}

		if len(arrStringTopic) > 0 {

			var topic entity.RedisTopicModel
			bytesTopic, err := rd.Get(arrStringTopic[0]).Bytes()
			//bytesTopic, err := redis.Bytes(c.Do("GET", arrStringTopic[0]))
			if err != nil {
				return nil, err
			}

			gob.NewDecoder(bytes.NewReader(bytesTopic)).Decode(&topic)
			topics = append(topics, &topic)
		}
	}

	for i := 0; i < len(topics); i++ {
		for j := 0; j < len(topics)-1; j++ {
			if topics[j].UsedCount < topics[j+1].UsedCount {
				// swap
				topics[j], topics[j+1] = topics[j+1], topics[j]
			}
		}
	}

	if len(topics) > config.LimitSearchTopic {
		topics = topics[:config.LimitSearchTopic]
	}

	return topics, nil

}

//SearchGap search The Gap
func SearchGap(rd *redis.Client, text string) ([]*entity.RedisGapModel, error) {

	scanText := "*:" + text + "*"
	if utf8.RuneCountInString(text) >= 3 {
		scanText = "*:*" + text + "*"
	}

	var cursor uint64

	// ค้นหา Gap ใน Redis
	values, _, err := rd.SScan(config.RedisIndexGapName, cursor, scanText, config.RedisCount).Result()
	//values, err := redis.Values(c.Do("SSCAN", config.RedisIndexGapName, cursor, "match", scanText, "count", config.RedisCount))
	if err != nil {
		return nil, err
	}

	var gaps []*entity.RedisGapModel
	for i := 0; i < len(values); i++ {

		va := strings.Split(values[i], ":")
		arrStringGap, err := rd.Keys(config.RedisGap + "*" + ":" + va[0]).Result()
		//arrStringGap, err := redis.Strings(c.Do("KEYS", config.RedisGap+"*"+":"+values[0]))
		if err != nil {
			return nil, err
		}

		if len(arrStringGap) > 0 {

			var gap entity.RedisGapModel
			bytesGap, err := rd.Get(arrStringGap[0]).Bytes()
			// bytesGap, err := redis.Bytes(c.Do("GET", arrStringGap[0]))
			if err != nil {
				return nil, err
			}

			gob.NewDecoder(bytes.NewReader(bytesGap)).Decode(&gap)

			if gap.Username == "" {
				gap.Username = gap.ID
			}

			gaps = append(gaps, &gap)
		}

	}

	for i := 0; i < len(gaps); i++ {
		for j := 0; j < len(gaps)-1; j++ {
			if gaps[j].CountFollower < gaps[j+1].CountFollower {
				// swap
				gaps[j], gaps[j+1] = gaps[j+1], gaps[j]
			}
		}
	}

	if len(gaps) > config.LimitSearchGap {
		gaps = gaps[:config.LimitSearchGap]
	}

	return gaps, nil
}
