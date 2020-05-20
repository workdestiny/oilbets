package repository

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/acoshift/pgsql"
	"github.com/lib/pq"
	"github.com/mssola/user_agent"
	"github.com/workdestiny/convbox"

	"github.com/go-redis/redis"
	"github.com/workdestiny/oilbets/config"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/service"
)

const (
	querySelectPost = `SELECT
	post.id,
	post.slug,
	post.owner_id,
	post.owner_type,
	post.title,
	post.description,
	post.link,
	post.link_description,
	post.image_url,
	post.type,
	post.height,
	post.width,
	post.province,
	post.created_at,
	post.like_count,
	post.comment_count,
	post.share_count,
	post.view_count`

	querySelectDiscovery = `SELECT
	post.id,
	post.slug,
	post.owner_id,
	post.owner_type,
	post.title,
	post.description,
	post.link,
	post.link_description,
	post.image_url,
	post.type,
	post.height,
	post.width,
	post.created_at,
	post.like_count,
	post.comment_count,
	post.share_count,
	post.view_count`

	selectTopic                  = ` , tagtopic.topic_id`
	selectGap                    = ` , gap.id, gap.name->>'text', gap.display->>'mini'`
	fromPost                     = " FROM post"
	queryWherePostOnlineDiscover = ` post.status = true AND post.verify = 0`
	queryWherePostOnlinePublic   = ` post.status = true`
	queryWherePostOnlineFollow   = ` post.status = true`
	queryWhereTopic              = " AND tagtopic.topic_id = ANY($2) AND tagtopic.main = true"
	queryWhereGap                = " AND post.owner_id = ANY($2)"
	lateTime                     = ` ORDER BY created_at DESC`
)

//ListPostDiscover get list post all
func ListPostDiscover(q Queryer, rd *redis.Client, userID string, t string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text',
			   gap.display->>'mini', gap.username->>'text', likepost.status, post.guest_view_count
		  FROM (SELECT * FROM post WHERE status = true AND verify = 2 AND created_at < $1 ORDER BY created_at DESC LIMIT $2) as post
	 LEFT JOIN gap
		    ON post.owner_id = gap.id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $3
	  ORDER BY post.created_at DESC
		`, time.Now().UTC(), limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		//var topicID string
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			owner.IsVerify = true

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostDiscoverNextLoad get list post all (NextLoad)
func ListPostDiscoverNextLoad(q Queryer, rd *redis.Client, userID string, t string, timeLoad time.Time, isMobile bool, limit int) ([]*entity.DiscoverModel, error) {

	queryWhereType := ""
	if t != "" {
		switch t {
		case "article":
			queryWhereType = " AND type = 0"
		case "gallery":
			queryWhereType = " AND type = 1"
		}
	}
	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text',
			   gap.display->>'mini', gap.username->>'text', likepost.status, post.guest_view_count
		  FROM (SELECT * FROM post WHERE status = true AND verify = 2`+queryWhereType+` AND created_at < $1 ORDER BY created_at DESC LIMIT $2) as post
	 LEFT JOIN gap
		    ON post.owner_id = gap.id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $3
	  ORDER BY post.created_at DESC
		`, timeLoad, limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			if isMobile {
				p.ImageURL = p.ImageURLMobile
			}

			owner.IsVerify = true

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostPublic get list post all follow topic
func ListPostPublic(q Queryer, rd *redis.Client, userID string, t string, isMobile bool, limit int) ([]*entity.DiscoverModel, error) {
	//WHERE status = true AND verify > 0 AND tg.topic_id = ANY($1)
	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count,
			   post.image_url_mobile
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count, post.image_url_mobile FROM post
				 WHERE status = true AND verify > 0 AND post.created_at < $1
				 ORDER BY post.created_at DESC LIMIT $2) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $3
			ORDER BY post.created_at DESC
		`, time.Now().UTC(), limit, userID)
	//	  ORDER BY post.created_at DESC
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		//var topicID string
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display,
			&owner.Username, &like, &owner.IsVerify, &countGuestView,
			&p.ImageURLMobile)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			if isMobile {
				if len(p.ImageURLMobile) != 0 {
					p.ImageURL = p.ImageURLMobile
				}
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
				// TagTopics: []entity.TagTopic{
				// 	entity.TagTopic{
				// 		TopicID: topicID,
				// 		Name:    getTopicName(rd, topicID),
				// 	},
				// },
			})
			continue
		}

		// lp[len(lp)-1].TagTopics = append(lp[len(lp)-1].TagTopics, entity.TagTopic{
		// 	TopicID: topicID,
		// 	Name:    getTopicName(rd, topicID),
		// })

	}

	return lp, err
}

//ListPostPublicNextLoad get list post all follow topic (NextLoad)
func ListPostPublicNextLoad(q Queryer, rd *redis.Client, userID string, t string, timeLoad time.Time, isMobile bool, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count,
			   post.image_url_mobile
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count, post.image_url_mobile FROM post
				 WHERE status = true AND verify > 0 AND post.created_at < $1
				 ORDER BY post.created_at DESC LIMIT $2) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $3
	  ORDER BY post.created_at DESC
		`, timeLoad, limit, userID)
	// ORDER BY post.created_at DESC
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		//var topicID string
		var countGustView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display,
			&owner.Username, &like, &owner.IsVerify, &countGustView,
			&p.ImageURLMobile)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			if isMobile {
				if len(p.ImageURLMobile) != 0 {
					p.ImageURL = p.ImageURLMobile
				}
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGustView + len(p.Description)),
				},
				// TagTopics: []entity.TagTopic{
				// 	entity.TagTopic{
				// 		TopicID: topicID,
				// 		Name:    getTopicName(rd, topicID),
				// 	},
				// },
			})
			continue
		}

		// lp[len(lp)-1].TagTopics = append(lp[len(lp)-1].TagTopics, entity.TagTopic{
		// 	TopicID: topicID,
		// 	Name:    getTopicName(rd, topicID),
		// })

	}

	return lp, err
}

//ListPostFollow get list post all follow gap
func ListPostFollow(q Queryer, rd *redis.Client, userID string, t string, gapIDs []string, limit int) ([]*entity.DiscoverModel, error) {

	queryWhereType := ""
	if t != "" {
		switch t {
		case "article":
			queryWhereType = " AND type = 0"
		case "gallery":
			queryWhereType = " AND type = 1"
		}
	}

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT *
				  FROM post
			     WHERE status = true`+queryWhereType+` AND owner_id = ANY($1)
			  ORDER BY created_at DESC
			     LIMIT $2) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $3
	  ORDER BY post.created_at DESC
		`, pq.Array(gapIDs), limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostFollowNextLoad get list post all follow gap (NextLoad)
func ListPostFollowNextLoad(q Queryer, rd *redis.Client, userID string, t string, gapIDs []string, timeLoad time.Time, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT *
				  FROM post
			     WHERE status = true AND owner_id = ANY($1) AND created_at < $2
			  ORDER BY created_at DESC
			     LIMIT $3) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	    	ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $4
	  ORDER BY post.created_at DESC, tagtopic ASC
		`, pq.Array(gapIDs), timeLoad, limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostGap get list post all gap
func ListPostGap(q Queryer, rd *redis.Client, userID string, gapID string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT *
				  FROM post
			     WHERE status = true AND owner_id = $1 AND created_at < $2
			  ORDER BY created_at DESC
			     LIMIT $3) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $4
	  ORDER BY post.created_at DESC
		`, gapID, time.Now().UTC(), limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostGapNextLoad get list post all gap (NextLoad)
func ListPostGapNextLoad(q Queryer, rd *redis.Client, userID string, gapID string, timeLoad time.Time, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT *
				  FROM post
			     WHERE status = true AND owner_id = $1 AND created_at < $2
			  ORDER BY created_at DESC
			     LIMIT $3) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $4
	  ORDER BY post.created_at DESC;
		`, gapID, timeLoad, limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostOwnerGap get list post all gap
func ListPostOwnerGap(q Queryer, rd *redis.Client, userID string, gapID string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT *
				  FROM post
			     WHERE status = true AND owner_id = $1
			  ORDER BY created_at DESC
			     LIMIT $2) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $3
	  ORDER BY post.created_at DESC
		`, gapID, limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel

		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostTopic get list post all topic
func ListPostTopic(q Queryer, rd *redis.Client, userID string, topicCode string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count FROM post
			 LEFT JOIN tagtopic as tg
					ON post.id = tg.post_id
			 LEFT JOIN topic as tc
					ON tc.code = $1
				 WHERE post.status = true AND tg.topic_id = tc.id AND post.created_at < $2 ORDER BY post.created_at DESC LIMIT $3) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $4
	  ORDER BY post.created_at DESC
		`, topicCode, time.Now().UTC(), limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostTopicNextLoad get list post all topic (NextLoad)
func ListPostTopicNextLoad(q Queryer, rd *redis.Client, userID string, topicCode string, timeLoad time.Time, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count FROM post
			 LEFT JOIN tagtopic as tg
					ON post.id = tg.post_id
			 LEFT JOIN topic as tc
					ON tc.code = $1
				 WHERE post.status = true AND tg.topic_id = tc.id AND post.created_at < $2 ORDER BY post.created_at DESC LIMIT $3) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $4
	  ORDER BY post.created_at DESC
		`, topicCode, timeLoad, limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostCategory get list post all on category
func ListPostCategory(q Queryer, rd *redis.Client, userID string, categoryCode string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count FROM post
			 LEFT JOIN tagtopic as tg
					ON post.id = tg.post_id
			 LEFT JOIN topic as tc
					ON tc.cat_id = $1 AND tc.verify = true
				 WHERE post.status = true AND tg.topic_id = tc.id and post.created_at < $2 ORDER BY post.created_at DESC LIMIT $3) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $4
	  ORDER BY post.created_at DESC
		`, categoryCode, time.Now().UTC(), limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostCategoryNextLoad get list post all topic on category (NextLoad)
func ListPostCategoryNextLoad(q Queryer, rd *redis.Client, userID string, topicCode string, timeLoad time.Time, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count FROM post
			 LEFT JOIN tagtopic as tg
					ON post.id = tg.post_id
			 LEFT JOIN topic as tc
					ON tc.cat_id = $1 AND tc.verify = true
				 WHERE post.status = true AND tg.topic_id = tc.id AND post.created_at < $2 ORDER BY post.created_at DESC LIMIT $3) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $4
	  ORDER BY post.created_at DESC
		`, topicCode, timeLoad, limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostLiked get list post all
func ListPostLiked(q Queryer, rd *redis.Client, userID string, t string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
	SELECT post.id, post.slug, post.owner_id, post.owner_type,
		   post.title, post.description, post.link, post.link_description,
		   post.image_url, post.type, post.height, post.width,
		   post.created_at, post.like_count, post.comment_count, post.share_count,
		   post.view_count, gap.id, gap.name->>'text',
		   gap.display->>'mini', gap.username->>'text', likepost.status, post.guest_view_count
	  FROM (SELECT p.id, p.slug, p.owner_id, p.owner_type,
				   p.title, p.description, p.link, p.link_description,
				   p.image_url, p.type, p.height, p.width,
				   p.created_at, p.like_count, p.comment_count, p.share_count,
				   p.view_count, p.guest_view_count
			  FROM post as p
		 LEFT JOIN likepost
				ON likepost.status = true AND likepost.owner_id = $2
			 WHERE p.status = true AND p.verify = 2 AND p.id = likepost.post_id
		  ORDER BY p.created_at DESC LIMIT $1) as post
 LEFT JOIN gap
		ON post.owner_id = gap.id
 LEFT JOIN likepost
		ON post.id = likepost.post_id AND likepost.owner_id = $2
  ORDER BY post.created_at DESC;
		`, limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			owner.IsVerify = true

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostLikedNextLoad get list post all (NextLoad)
func ListPostLikedNextLoad(q Queryer, rd *redis.Client, userID string, t string, timeLoad time.Time, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
	SELECT post.id, post.slug, post.owner_id, post.owner_type,
		   post.title, post.description, post.link, post.link_description,
		   post.image_url, post.type, post.height, post.width,
		   post.created_at, post.like_count, post.comment_count, post.share_count,
		   post.view_count, gap.id, gap.name->>'text',
		   gap.display->>'mini', gap.username->>'text', likepost.status, post.guest_view_count
	  FROM (SELECT p.id, p.slug, p.owner_id, p.owner_type,
				   p.title, p.description, p.link, p.link_description,
				   p.image_url, p.type, p.height, p.width,
				   p.created_at, p.like_count, p.comment_count, p.share_count,
				   p.view_count, p.guest_view_count
			  FROM post as p
		 LEFT JOIN likepost
				ON likepost.status = true AND likepost.owner_id = $3
			 WHERE p.status = true AND p.verify = 2 AND p.id = likepost.post_id AND p.created_at < $1
		  ORDER BY p.created_at DESC LIMIT $2) as post
 LEFT JOIN gap
		ON post.owner_id = gap.id
 LEFT JOIN likepost
		ON post.id = likepost.post_id AND likepost.owner_id = $3
  ORDER BY post.created_at DESC;
		`, timeLoad, limit, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			owner.IsVerify = true

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostRead get list post all follow topic
func ListPostRead(q Queryer, rd *redis.Client, userID string, t time.Time, topicIDs []string, postID string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count FROM post
			 LEFT JOIN tagtopic as tg
				    ON post.id = tg.post_id
				 WHERE status = true AND verify > 0 AND post.id != $1 AND tg.topic_id = ANY($2) AND post.created_at < $3 ORDER BY post.created_at DESC LIMIT $4) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $5
	  ORDER BY post.created_at DESC;
		`, postID, pq.Array(topicIDs), t.UTC(), limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//ListPostReadNextLoad get list post all follow topic (NextLoad)
func ListPostReadNextLoad(q Queryer, rd *redis.Client, userID string, t string, topicIDs []string, postID string, timeLoad time.Time, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count FROM post
			 LEFT JOIN tagtopic as tg
				    ON post.id = tg.post_id
				 WHERE status = true AND verify > 0 AND post.id != $1 AND tg.topic_id = ANY($2) AND post.created_at < $3 ORDER BY post.created_at DESC LIMIT $4) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $5
	  ORDER BY post.created_at DESC;
		`, postID, pq.Array(topicIDs), timeLoad, limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGustView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGustView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGustView + len(p.Description)),
				},
			})
			continue
		}
	}

	return lp, err
}

//CheckPostID Check Post ID
func CheckPostID(q Queryer, postID string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT owner_id
		  FROM public.post
		 WHERE id = $1
		   AND status = true
		   AND used = true;
	`, postID).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil

}

//CheckPostSlug Check Post Slug
func CheckPostSlug(q Queryer, slug string) error {

	var used bool
	err := q.QueryRow(`
		SELECT used
		  FROM public.post
		 WHERE slug = $1;
	`, slug).Scan(&used)
	if err != nil {
		return err
	}

	return nil

}

//GetPostSlug get Post Slug
func GetPostSlug(q Queryer, id string) (string, error) {

	var slug string
	err := q.QueryRow(`
		SELECT slug
		  FROM public.post
		 WHERE id = $1;
	`, id).Scan(&slug)
	if err != nil {
		return "", err
	}

	return slug, nil
}

//GetPostCreateTime get Post Time
func GetPostCreateTime(q Queryer, id string) (time.Time, error) {

	var createdAt time.Time
	err := q.QueryRow(`
		SELECT created_at
		  FROM public.post
		 WHERE id = $1;
	`, id).Scan(&createdAt)
	if err != nil {
		return createdAt, err
	}

	return createdAt, nil
}

//CheckOwnerPost check owner post
func CheckOwnerPost(q Queryer, userID string, postID string) error {

	var used bool
	err := q.QueryRow(`
		SELECT used
		  FROM public.post
		 WHERE id = $1
		   AND user_id = $2
		   AND status = true
		   AND used = true;
	`, postID, userID).Scan(&used)
	if err != nil {
		return err
	}

	return nil
}

//FindGapOwnerPost check owner post
func FindGapOwnerPost(q Queryer, postID string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT owner_id
		  FROM public.post
		 WHERE id = $1
	`, postID).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func getTopicName(rd *redis.Client, topicID string) string {

	stringTopic, _ := rd.Keys(config.RedisTopicAny + "*:" + topicID + ":*").Result()

	//stringTopic, _ := redis.Strings(c.Do("KEYS", config.RedisTopicAny+"*:"+topicID+":*"))
	if len(stringTopic) > 0 {
		var topic entity.RedisTopicModel
		bytesTopic, _ := rd.Get(stringTopic[0]).Bytes()
		//bytesTopic, _ := redis.Bytes(c.Do("GET", stringTopic[0]))
		gob.NewDecoder(bytes.NewReader(bytesTopic)).Decode(&topic)
		return topic.Name
	}

	return ""
}

//ListTopicRecommend list input number topic recommend
func ListTopicRecommend(q Queryer, limit int) ([]*entity.TopicRecommendModel, error) {

	rows, err := q.Query(`
		SELECT id, name->>'th'
		  FROM topic
	 	 WHERE verify = true AND is_customer = false
	  ORDER BY count DESC
		 LIMIT $1;`, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var list []*entity.TopicRecommendModel
	for rows.Next() {

		var t entity.TopicRecommendModel
		err := rows.Scan(&t.ID, &t.Name)
		if err != nil {
			return nil, err
		}
		list = append(list, &t)

	}

	return list, nil

}

//ListTopicRecommendForCustomer list input number topic recommend
func ListTopicRecommendForCustomer(q Queryer, limit int) ([]*entity.TopicRecommendModel, error) {

	rows, err := q.Query(`
		SELECT id, name->>'th'
		  FROM topic
	 	 WHERE verify = true
	  ORDER BY count DESC
		 LIMIT $1;`, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var list []*entity.TopicRecommendModel
	for rows.Next() {

		var t entity.TopicRecommendModel
		err := rows.Scan(&t.ID, &t.Name)
		if err != nil {
			return nil, err
		}
		list = append(list, &t)

	}

	return list, nil

}

//CheckLike put userID and postID can like and unLike
func CheckLike(q Queryer, userID string, postID string) (bool, error) {

	var isLike bool
	err := q.QueryRow(`
		SELECT status
		  FROM public.likepost
		 WHERE post_id = $1
		   AND owner_id = $2
		   AND used = true;
	`, postID, userID).Scan(&isLike)

	return isLike, err

}

//CreateLike put userID and postID created like post
func CreateLike(q Queryer, userID string, postID string) error {

	_, err := q.Exec(`
		INSERT INTO public.likepost
					(post_id, owner_id, status, used)
			 VALUES ($1, $2, $3, $4) RETURNING id;
	`, postID, userID, true, true)

	_, err = q.Exec(`
		UPDATE public.post
		   SET like_count = like_count + 1
		 WHERE id = $1;
	`, postID)

	return err
}

//LikePost put userID and postID set like post
func LikePost(q Queryer, userID string, postID string, isLike bool) error {

	_, err := q.Exec(`
		UPDATE public.likepost
		   SET status = $1
		 WHERE post_id = $2
		   AND owner_id = $3
		   AND used = true;
	`, isLike, postID, userID)

	set := "+ 1"
	if !isLike {
		set = "- 1"
	}

	_, err = q.Exec(`
		UPDATE public.post
		   SET like_count = like_count `+set+`
		 WHERE id = $1;
	`, postID)

	return err
}

// GetPost select input postID
func GetPost(q Queryer, rd *redis.Client, userID string, postID string) (*entity.GetPostModel, error) {

	var p entity.GetPostModel
	rows, err := q.Query(`
		SELECT post.id, post.slug, post.title, post.description,
		       post.link, post.link_description, post.image_url, post.type,
		       post.height, post.width, post.created_at, post.like_count,
			   post.comment_count, post.view_count, tagtopic.topic_id, gap.id,
			   gap.name->>'text', gap.display->>'middle', COALESCE(user_kycs.is_idcard, 'false'), likepost.status,
			   follow_gap.status, post.province, COALESCE(post.image_share_url, ''), COALESCE(post.width_share, 0),
			   COALESCE(post.height_share, 0), post.guest_view_count, post.updated_at, post.vdo_url
		  FROM post
	 LEFT JOIN tagtopic
	        ON post.id = tagtopic.post_id
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	        ON gap.user_id = user_kycs.user_id
	 LEFT JOIN likepost
			ON post.id = likepost.post_id AND likepost.used = true AND likepost.owner_id = $1
	 LEFT JOIN follow_gap
			ON post.owner_id = follow_gap.gap_id AND follow_gap.used = true AND follow_gap.owner_id = $2
		 WHERE post.id = $3 OR post.slug = $3
		   AND post.status = true
	  ORDER BY tagtopic ASC;
 		`, userID, userID, postID)
	if err != nil {
		return &p, err
	}
	defer rows.Close()

	for rows.Next() {

		var topicID string
		var province int
		var countLike, countComment, countView, countGuestView int
		var like, followGap interface{}
		err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description,
			&p.Link, &p.LinkDescription, &p.ImageURL, &p.Type,
			&p.Height, &p.Width, &p.CreatedAt, &countLike,
			&countComment, &countView, &topicID, &p.Owner.ID,
			&p.Owner.Name, &p.Owner.Display, &p.Owner.IsVerify, &like,
			&followGap, &province, &p.ImageShareURL, &p.WidthShare,
			&p.HeightShare, &countGuestView, &p.UpdatedAt, &p.VdoURL)
		if err != nil {
			return &p, err
		}

		if len(p.TagTopics) == 0 {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			isFollowGap := false
			if followGap != nil {
				isFollowGap = followGap.(bool)
			}

			p.Count.Like = convbox.ShortNumber(countLike + (len(p.Description) / config.AMLLuckieNumber))
			p.Count.Comment = convbox.ShortNumber(countComment)
			p.Count.View = convbox.ShortNumber(countView + countGuestView + len(p.Description))

			p.Province = entity.GetProvinceName(province)
			p.IsLike = isLike
			p.Owner.IsFollow = isFollowGap
			p.TagTopics = append(p.TagTopics, entity.TagTopic{
				TopicID: topicID,
				Name:    getTopicName(rd, topicID),
			})
			continue

		}

		p.TagTopics = append(p.TagTopics, entity.TagTopic{
			TopicID: topicID,
			Name:    getTopicName(rd, topicID),
		})

	}

	return &p, err
}

//SearchPost get list post input keyword
func SearchPost(q Queryer, rd *redis.Client, userID string, keyword string, limit int) ([]*entity.DiscoverModel, error) {

	rows, err := q.Query(`
		SELECT post.id, post.slug, post.owner_id, post.owner_type,
			   post.title, post.description, post.link, post.link_description,
			   post.image_url, post.type, post.height, post.width,
			   post.created_at, post.like_count, post.comment_count, post.share_count,
			   post.view_count, gap.id, gap.name->>'text', gap.display->>'mini',
			   gap.username->>'text', likepost.status, COALESCE(user_kycs.is_idcard, 'false'), post.guest_view_count
		  FROM (SELECT post.id, post.slug, post.owner_id, post.owner_type,
			           post.title, post.description, post.link, post.link_description,
			           post.image_url, post.type, post.height, post.width,
			           post.created_at, post.like_count, post.comment_count, post.share_count,
			           post.view_count, post.guest_view_count FROM post
				 WHERE status = true AND verify > 0 AND (lower(post.title) LIKE $1 OR lower(post.description) LIKE $1) ORDER BY post.created_at DESC LIMIT $2) as post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
	 LEFT JOIN user_kycs
	 		ON user_kycs.user_id = gap.user_id
	 LEFT JOIN likepost
		    ON post.id = likepost.post_id AND likepost.owner_id = $3
	  ORDER BY post.created_at DESC
		`, "%"+keyword+"%", limit, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.DiscoverModel
	var pid string
	for rows.Next() {

		var p entity.PostModel
		var countGuestView int
		var like interface{}
		var owner entity.PostOwner
		err := rows.Scan(&p.ID, &p.Slug, &p.OwnerID, &p.OwnerType,
			&p.Title, &p.Description, &p.Link, &p.LinkDescription,
			&p.ImageURL, &p.Type, &p.Height, &p.Width, &p.CreatedAt,
			&p.LikeCount, &p.CommentCount, &p.ShareCount, &p.ViewCount,
			&owner.ID, &owner.Name, &owner.Display, &owner.Username, &like, &owner.IsVerify, &countGuestView)
		if err != nil {
			return nil, err
		}

		if pid != p.ID {

			isLike := false
			if like != nil {
				isLike = like.(bool)
			}

			if owner.Username == "" {
				owner.Username = owner.ID
			}

			pid = p.ID
			lp = append(lp, &entity.DiscoverModel{
				ID:              p.ID,
				Slug:            p.Slug,
				Owner:           owner,
				Title:           service.ShortTextStripHTML(80, p.Title),
				Description:     service.ShortTextStripHTML(60, p.Description),
				Link:            p.Link,
				IsLike:          isLike,
				LinkDescription: p.LinkDescription,
				ImageURL:        p.ImageURL,
				Type:            p.Type,
				Height:          p.Height,
				Width:           p.Width,
				Time:            service.FormatTime(p.CreatedAt),
				CreatedAt:       p.CreatedAt,
				Count: entity.PostCount{
					Like:    convbox.ShortNumber(p.LikeCount + (len(p.Description) / config.AMLLuckieNumber)),
					Comment: convbox.ShortNumber(p.CommentCount),
					View:    convbox.ShortNumber(p.ViewCount + countGuestView + len(p.Description)),
				},
			})
			continue
		}

	}

	return lp, err
}

// ListTopPostGap input gapID and limit get top post
func ListTopPostGap(q Queryer, gapID string, postID string, limit int) ([]*entity.TopPostRelateModel, error) {

	var lp []*entity.TopPostRelateModel
	rows, err := q.Query(`
		SELECT post.id, post.slug, post.title, post.description,
		       post.image_url, post.view_count, post.guest_view_count
		  FROM post
		 WHERE post.owner_id = $1
		   AND post.id != $2
		   AND post.status = true
	  ORDER BY post.created_at DESC
	     LIMIT $3;
		`, gapID, postID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var p entity.TopPostRelateModel
		var countGuestView int
		var description string
		err := rows.Scan(&p.ID, &p.Slug, &p.Title, &description, &p.ImageURL, &p.View, &countGuestView)
		if err != nil {
			return nil, err
		}

		if p.ImageURL == "" {
			p.ImageURL = config.ImagePostRelate
		}

		if p.Title == "" {
			p.Title = description
		}

		p.View = p.View + countGuestView + len(description)

		lp = append(lp, &p)

	}

	return lp, err

}

// ListPostRelateTagTopic input topicID and limit get top post
func ListPostRelateTagTopic(q Queryer, postID string, provinceID int, limit int) ([]*entity.TopPostRelateModel, error) {

	var lp []*entity.TopPostRelateModel
	rows, err := q.Query(`
		SELECT post.id, post.slug, post.title, post.description,
		       post.image_url, post.view_count, post.guest_view_count
		  FROM post
		 WHERE post.status = true AND post.province = $1 AND post.id != $2
	  ORDER BY random()
		 LIMIT $3;
		`, provinceID, postID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var p entity.TopPostRelateModel
		var countGuestView int
		var description string
		err := rows.Scan(&p.ID, &p.Slug, &p.Title, &description, &p.ImageURL, &p.View, &countGuestView)
		if err != nil {
			return nil, err
		}

		if p.ImageURL == "" {
			p.ImageURL = config.ImagePostRelate
		}

		if p.Title == "" {
			p.Title = description
		}

		p.View = p.View + countGuestView + len(description)

		lp = append(lp, &p)

	}

	return lp, err

}

//ListComment input postID and limit list comment
func ListComment(q Queryer, rd *redis.Client, postID string, limit int) ([]*entity.CommentModel, error) {

	var lc []*entity.CommentModel
	rows, err := q.Query(`
		SELECT comment.id, comment.owner_id, comment.owner_type, comment.text,
			   comment.created_at, gap.username->>'text'
		  FROM comment
	 LEFT JOIN gap
		    ON comment.owner_id = gap.id
		 WHERE comment.post_id = $1 AND comment.status = true AND comment.used = true
	  ORDER BY created_at DESC
		 LIMIT $2;
		`, postID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var ownerType entity.TypeOwnerComment
		var c entity.CommentModel
		var username interface{}
		err := rows.Scan(&c.ID, &c.Owner.ID, &ownerType, &c.Text, &c.CreatedAt, &username)
		if err != nil {
			return nil, err
		}

		c.Time = service.FormatTime(c.CreatedAt)

		if ownerType == entity.TypeOwnerCommentUser {

			rdUser := getRedisUser(rd, c.Owner.ID)

			c.Owner.Name = rdUser.FirstName + " " + rdUser.LastName
			c.Owner.Display = rdUser.DisplayImage
			c.Owner.Type = ownerType.String()
		}

		if ownerType == entity.TypeOwnerCommentGap {

			if username != nil {
				c.Owner.Username = username.(string)
			}

			if c.Owner.Username == "" {
				c.Owner.Username = c.Owner.ID
			}

			rdGap := getRedisGap(rd, c.Owner.ID)

			c.Owner.Name = rdGap.Name
			c.Owner.Display = rdGap.DisplayImage
			c.Owner.Type = ownerType.String()
		}

		c.Text = service.UnescapeString(c.Text)

		lc = append(lc, &c)

	}

	return lc, err

}

//ListCommentNextLoad input postID and limit list comment (NextLoad)
func ListCommentNextLoad(q Queryer, rd *redis.Client, postID string, timeLoad time.Time, limit int) ([]*entity.CommentModel, error) {

	var lc []*entity.CommentModel
	rows, err := q.Query(`
		SELECT id, owner_id, owner_type, text,
		       created_at
		  FROM comment
		 WHERE post_id = $1 AND status = true AND used = true AND created_at < $2
	  ORDER BY created_at DESC
		 LIMIT $3;
		`, postID, timeLoad, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var ownerType entity.TypeOwnerComment
		var c entity.CommentModel
		err := rows.Scan(&c.ID, &c.Owner.ID, &ownerType, &c.Text, &c.CreatedAt)
		if err != nil {
			return nil, err
		}

		c.Time = service.FormatTime(c.CreatedAt)

		if ownerType == entity.TypeOwnerCommentUser {

			rdUser := getRedisUser(rd, c.Owner.ID)

			c.Owner.Name = rdUser.FirstName + " " + rdUser.LastName
			c.Owner.Display = rdUser.DisplayImage
			c.Owner.Type = ownerType.String()
		}

		if ownerType == entity.TypeOwnerCommentGap {

			rdGap := getRedisGap(rd, c.Owner.ID)

			c.Owner.Name = rdGap.Name
			c.Owner.Display = rdGap.DisplayImage
			c.Owner.Type = ownerType.String()
		}

		c.Text = service.UnescapeString(c.Text)

		lc = append(lc, &c)

	}

	return lc, err

}

//CheckOwnerComment Check Owner comment
func CheckOwnerComment(q Queryer, userID string, commentID string) error {

	var ownerID string
	var ty entity.TypeOwnerComment
	err := q.QueryRow(`
		SELECT owner_id, owner_type
		  FROM public.comment
		 WHERE id = $1
		   AND status = true
		   AND used = true;
	`, commentID).Scan(&ownerID, &ty)
	if err != nil {
		return err
	}

	if ty == entity.TypeOwnerCommentUser {

		if userID == ownerID {
			return nil
		}
		return fmt.Errorf("not owner")
	}

	// type Gap Comment
	var used bool
	err = q.QueryRow(`
		SELECT used
		  FROM public.gap
		 WHERE id = $1
		   AND user_id = $2
		   AND used = true;
	`, ownerID, userID).Scan(&used)
	if err != nil {
		return err
	}

	return nil

}

//CommentPost (ใช้ Tx) input id user or id gap input type user or gap , text is user comment
func CommentPost(q Queryer, postID string, id string, ty entity.TypeOwnerComment, text string) (string, error) {

	var returnID string
	err := q.QueryRow(`
		INSERT INTO public.comment
					(post_id, owner_id, owner_type, text,
					status, used)
			 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;
	`, postID, id, ty, text, true, true).Scan(&returnID)
	if err != nil {
		return "", err
	}

	_, err = q.Exec(`
		UPDATE public.post
		   SET comment_count = comment_count + 1
		 WHERE id = $1;
	`, postID)
	if err != nil {
		return "", err
	}

	return returnID, nil
}

//DeleteComment input comment id ที่จะลบ
func DeleteComment(q Queryer, postID string, commentID string) error {

	_, err := q.Exec(`
		UPDATE public.comment
		   SET status = false
		 WHERE id = $1
		   AND post_id = $2;
	`, commentID, postID)
	if err != nil {
		return err
	}

	_, err = q.Exec(`
		UPDATE public.post
		   SET comment_count = comment_count - 1
		 WHERE id = $1;
	`, postID)
	if err != nil {
		return err
	}

	return nil
}

//EditComment input comment id and text ที่จะ edit
func EditComment(q Queryer, commentID string, text string) error {

	_, err := q.Exec(`
		UPDATE public.comment
		   SET text = $1
		 WHERE id = $2;
	`, text, commentID)
	if err != nil {
		return err
	}

	return nil
}

//CheckDraftPost check is draft
func CheckDraftPost(q Queryer, userID string) (string, error) {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM public.post
		 WHERE user_id = $1
		   AND draft = true
		   AND status = false
		   AND used = true;
	`, userID).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

// GetDraftPost select input userID
func GetDraftPost(q Queryer, userID string) (*entity.GetDraftPostModel, error) {

	var d entity.GetDraftPostModel
	var updatedAt time.Time
	err := q.QueryRow(`
		SELECT post.id, post.title, post.description, gap.id,
			   gap.name->>'text', gap.display->>'mini', post.updated_at
		  FROM post
	 LEFT JOIN gap
			ON post.owner_id = gap.id
		 WHERE post.user_id = $1
		   AND post.status = false
		   AND post.draft = true
		   AND post.used = true;
		 `, userID).Scan(&d.ID, &d.Title, &d.Description, &d.Owner.ID,
		&d.Owner.Name, &d.Owner.Display, &updatedAt)
	if err != nil {
		return nil, err
	}

	d.Time = updatedAt.Format(time.RFC3339)

	return &d, nil
}

//UpdateDraftPost update
func UpdateDraftPost(q Queryer, postID string, p CreatePostModel) error {

	_, err := q.Exec(`
			  UPDATE post
				 SET owner_id = $1, owner_type = $2, title = $3, description = $4,
					 link = $5, image_url = $6, type = $7, province = $8,
					 height = $9, width = $10, verify = $11, link_description = $12,
					 updated_at = $13
			   WHERE id = $14
				 AND draft = true
				 AND status = false
				 AND used = true;
					`, p.OwnerID, 1, p.Title, p.Description,
		p.Link, p.ImageURL, p.Type, p.Province,
		p.Height, p.Width, p.StatusVerify, p.LinkDescription,
		time.Now().UTC(),
		postID)
	if err != nil {
		return err
	}

	return nil
}

//CreateDraftPost create
func CreateDraftPost(q Queryer, userID string, p CreatePostModel) error {

	var id string
	err := q.QueryRow(`
		INSERT INTO post
					(user_id, owner_id, owner_type, title,
					description, link, image_url, type,
					province, height, width, verify,
					link_description, like_count, comment_count, share_count,
					view_count, draft, status, used,
					slug, guest_view_count)
		     VALUES ($1, $2, $3, $4,
			 		$5, $6, $7, $8,
			 		$9, $10, $11, $12,
			 		$13, $14, $15, $16,
			 		$17, $18, $19, $20,
					$21, $22) RETURNING id;
					`, userID, p.OwnerID, 1, p.Title,
		p.Description, p.Link, p.ImageURL, p.Type,
		p.Province, p.Height, p.Width, p.StatusVerify,
		p.LinkDescription, 0, 0, 0,
		0, true, false, true,
		"", 0).Scan(&id)
	if err != nil {
		return err
	}

	return nil
}

//DeleteDraftPost delete
func DeleteDraftPost(q Queryer, userID string, postID string) error {

	_, err := q.Exec(`
		UPDATE public.post
		   SET draft = false
		 WHERE id = $1
		   AND user_id = $2;
	`, postID, userID)
	if err != nil {
		return err
	}

	return nil
}

//CreateTagTopic new TagTopic
func CreateTagTopic(q Queryer, postID string, topicIDs []string) error {

	for i, v := range topicIDs {

		isMain := false
		if i == 0 {
			isMain = true
		}

		_, err := q.Exec(`
			INSERT INTO tagtopic
						(post_id, topic_id, main)
				 VALUES ($1, $2, $3);
			`, postID, v, isMain)
		if err != nil {
			return err
		}
	}

	return nil
}

//UpdateTagTopic update TagTopic
func UpdateTagTopic(db *sql.DB, q Queryer, rd *redis.Client, postID string, topics []string) error {

	rows, err := q.Query(`
		SELECT topic_id
		  FROM tagtopic
		 WHERE post_id = $1;
		`, postID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {

		var id string
		err := rows.Scan(&id)
		if err != nil {
			return err
		}

		ids = append(ids, id)
	}

	go updateTagTopicCountToRedis(db, rd, ids)

	_, err = q.Exec(`
		DELETE FROM tagtopic
		      WHERE post_id = $1;
		`, postID)
	if err != nil {
		return err
	}

	for i, v := range topics {

		isMain := false
		if i == 0 {
			isMain = true
		}

		_, err := q.Exec(`
				INSERT INTO tagtopic
				(post_id, topic_id, main)
				VALUES ($1, $2, $3);
				`, postID, v, isMain)
		if err != nil {
			return err
		}
	}

	return nil

}

//DeleteTagTopic delete TagTopic
func DeleteTagTopic(db *sql.DB, q Queryer, rd *redis.Client, postID string) error {

	rows, err := q.Query(`
		SELECT topic_id
		  FROM tagtopic
		 WHERE post_id = $1;
		`, postID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {

		var id string
		err := rows.Scan(&id)
		if err != nil {
			return err
		}

		ids = append(ids, id)
	}

	go updateTagTopicCountToRedis(db, rd, ids)

	return nil

}

func updateTagTopicCountToRedis(db *sql.DB, rd *redis.Client, topicIDs []string) {

	pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {
		for _, v := range topicIDs {

			_, err := tx.Exec(`
			UPDATE public.topic
			   SET used_count = used_count - 1
			 WHERE id = $1;
		`, v)
			if err != nil {
				continue
			}

			t, err := GetTopic(tx, v)
			if err == nil {

				var topicRedis entity.RedisTopicModel
				topicRedis.ID = t.ID
				topicRedis.CatID = t.CatID
				topicRedis.Code = t.Code
				topicRedis.Name = t.Name
				topicRedis.Count = t.Count
				topicRedis.UsedCount = t.UsedCount
				topicRedis.Images = entity.Image{
					Mini:   t.ImageMini,
					Normal: t.Image,
				}

				buf := bytes.Buffer{}
				gob.NewEncoder(&buf).Encode(&topicRedis)

				if t.Verify {

					SetTopicVerifyToRedis(rd, t.CatID, t.ID, t.Code, buf.Bytes())
					continue
				}

				if !t.Verify {

					SetTopicNotVerifyToRedis(rd, t.CatID, t.ID, t.Code, buf.Bytes())
					continue
				}

			}
		}

		return nil
	})

}

//GetViewPost return time.Unix
func GetViewPost(q Queryer, postID string, userID string) (time.Time, error) {

	var t time.Time
	err := q.QueryRow(`
				SELECT created_at
				FROM public.view
				WHERE post_id = $1
				AND owner_id = $2
				ORDER BY created_at DESC;
				`, postID, userID).Scan(&t)
	if err != nil {
		return t, err
	}

	return t, nil
}

//GetGuestViewPost return time.Unix
func GetGuestViewPost(q Queryer, postID string, vsID string) (time.Time, error) {

	var t time.Time
	err := q.QueryRow(`
				SELECT created_at
				FROM public.guest
				WHERE post_id = $1
				AND visitor_id = $2
				ORDER BY created_at DESC;
				`, postID, vsID).Scan(&t)
	if err != nil {
		return t, err
	}

	return t, nil
}

//CreateViewPost create new row view post
func CreateViewPost(q Queryer, postID string, userID string, userAgent string, referrer string) error {

	_, err := q.Exec(`
					INSERT INTO public.view
					(post_id, owner_id, user_agent, referrer)
					VALUES ($1, $2, $3, $4);
					`, postID, userID, userAgent, referrer)
	if err != nil {
		return err
	}

	return nil
}

//CreateGuestViewPost create new row guest post
func CreateGuestViewPost(q Queryer, postID string, vsID string, userAgent string, referrer string) error {

	_, err := q.Exec(`
					INSERT INTO public.guest
					(post_id, visitor_id, user_agent, referrer)
					VALUES ($1, $2, $3, $4);
					`, postID, vsID, userAgent, referrer)
	if err != nil {
		return err
	}

	return nil
}

//UpdateCountViewPost update view count
func UpdateCountViewPost(q Queryer, postID string) error {

	_, err := q.Exec(`
						UPDATE public.post
						   SET view_count = view_count + 1
						 WHERE id = $1;
						`, postID)
	if err != nil {
		return err
	}

	return nil
}

//UpdateCountGuestViewPost update view count
func UpdateCountGuestViewPost(q Queryer, postID string) error {

	_, err := q.Exec(`
						UPDATE public.post
						   SET guest_view_count = guest_view_count + 1
						 WHERE id = $1;
						`, postID)
	if err != nil {
		return err
	}

	return nil
}

//CreateImage create new image upload
func CreateImage(q Queryer, userID string, imageURL string, h int, w int) error {

	_, err := q.Exec(`
							INSERT INTO public.imageinbucket
							            (owner_id, image, height, width,
								        status)
								 VALUES ($1, $2, $3, $4, $5);
								`, userID, imageURL, h, w, false)
	if err != nil {
		return err
	}

	return nil
}

//ConfirmImage Confirm new image upload
func ConfirmImage(q Queryer, userID string, imageURL string) error {

	_, err := q.Exec(`
			  UPDATE public.imageinbucket
				 SET status = true
			   WHERE owner_id = $1
			     AND image = $2;
					`, userID, imageURL)
	if err != nil {
		return err
	}

	return nil
}

//CreatePost create new post
func CreatePost(q Queryer, userID string, p CreatePostModel) (string, error) {

	var id string
	err := q.QueryRow(`
		INSERT INTO post
					(user_id, owner_id, owner_type, title,
					description, link, image_url, type,
					province, height, width, verify,
					link_description, like_count, comment_count, share_count,
					view_count, draft, status, used,
					slug, image_share_url, height_share, width_share,
					guest_view_count, created_at, image_url_mobile, vdo_url)
		     VALUES ($1, $2, $3, $4,
			 		$5, $6, $7, $8,
			 		$9, $10, $11, $12,
			 		$13, $14, $15, $16,
			 		$17, $18, $19, $20,
					$21, $22, $23, $24,
					$25, $26, $27, $28) RETURNING id;
					`, userID, p.OwnerID, 1, p.Title,
		p.Description, p.Link, p.ImageURL, p.Type,
		p.Province, p.Height, p.Width, p.StatusVerify,
		p.LinkDescription, 0, 0, 0,
		0, false, true, true,
		p.Slug, p.ImageShareURL, p.HeightShare, p.WidthShare,
		0, p.Onair, p.ImageURLMobile, p.VdoURL).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

//UpdateCreatePost update create post
func UpdateCreatePost(q Queryer, postID string, p CreatePostModel) error {

	t := time.Now().UTC()
	_, err := q.Exec(`
		UPDATE post
		   SET owner_id = $1, owner_type = $2, title = $3, description = $4,
			   link = $5, image_url = $6, type = $7, province = $8,
			   height = $9, width = $10, verify = $11, link_description = $12,
			   draft = $13, status = $14, slug = $15, created_at = $16,
			   updated_at = $17, image_share_url = $18, height_share = $19, width_share = $20,
			   image_url_mobile = $21, vdo_url = $22
		 WHERE id = $23
		   AND draft = true
		   AND status = false
		   AND used = true;
			  `, p.OwnerID, 1, p.Title, p.Description,
		p.Link, p.ImageURL, p.Type, p.Province,
		p.Height, p.Width, p.StatusVerify, p.LinkDescription,
		false, true, p.Slug, p.Onair, t,
		p.ImageShareURL, p.HeightShare, p.WidthShare, p.ImageURLMobile,
		p.VdoURL, postID)
	if err != nil {
		return err
	}

	return nil
}

//DeletePost delete post
func DeletePost(q Queryer, postID string) error {

	_, err := q.Exec(`
		UPDATE public.post
		   SET status = false
		 WHERE id = $1
	`, postID)
	if err != nil {
		return err
	}

	return nil
}

//EditPost edit post
func EditPost(q Queryer, p EditPostModel) error {

	t := time.Now().UTC()
	_, err := q.Exec(`
		UPDATE post
		   SET title = $1, description = $2, link = $3, image_url = $4,
			   type = $5, province = $6, height = $7, width = $8,
			   verify = $9, link_description = $10, updated_at = $11, image_share_url = $12,
			   height_share = $13, width_share = $14, image_url_mobile = $15, vdo_url =$16
		 WHERE id = $17
		   AND draft = false
		   AND status = true
		   AND used = true;
			  `, p.Title, p.Description, p.Link, p.ImageURL,
		p.Type, p.Province, p.Height, p.Width,
		p.StatusVerify, p.LinkDescription, t, p.ImageShareURL,
		p.HeightShare, p.WidthShare, p.ImageURLMobile, p.VdoURL,
		p.ID)
	if err != nil {
		return err
	}

	return nil
}

//UpdatePostVerify update post verify level
func UpdatePostVerify(q Queryer, userID string, level int) error {

	gapIDs := ListGapID(q, userID)

	_, err := q.Exec(`
			  UPDATE public.post
				 SET verify = $1
			   WHERE owner_id = ANY($2);
					`, level, pq.Array(gapIDs))
	if err != nil {
		return err
	}

	return nil
}

//GetCountView return model view and guestView sum
func GetCountView(q Queryer, gapID string, a, b time.Time) (*entity.GetViewModel, *entity.GetViewModel, error) {

	var p entity.GetViewModel
	var g entity.GetViewModel

	err := q.QueryRow(`
		SELECT COALESCE(sum(view.c), 0)::int, COALESCE(sum(guest.c), 0)::int, COALESCE(avg(gap_view.c), 0)::int, COALESCE(avg(gap_guest.c), 0)::int
		  FROM (SELECT id FROM gap WHERE id = $1) as gap
	 LEFT JOIN post
	 		ON post.owner_id = gap.id AND post.status = true AND post.used = true
	 LEFT JOIN (SELECT post_id, count(*) as c FROM view WHERE created_at >= $2 AND created_at <= $3 GROUP BY post_id) as view
	   		ON view.post_id = post.id
	 LEFT JOIN (SELECT post_id, count(*) as c FROM guest WHERE created_at >= $2 AND created_at <= $3 GROUP BY post_id) as guest
	   		ON guest.post_id = post.id
	 LEFT JOIN (SELECT gap_id, count(*) as c FROM gap_view WHERE created_at >= $2 AND created_at <= $3 GROUP BY gap_id) as gap_view
	   		ON gap_view.gap_id = gap.id
	 LEFT JOIN (SELECT gap_id, count(*) as c FROM gap_guest WHERE created_at >= $2 AND created_at <= $3 GROUP BY gap_id) as gap_guest
	   		ON gap_guest.gap_id = gap.id
		`, gapID, a, b).Scan(&p.View, &p.GuestView, &g.View, &g.GuestView)
	if err != nil {
		return nil, nil, err
	}

	p.All = p.Sum()
	g.All = g.Sum()

	return &p, &g, nil
}

//ListPostCountView return model view and guestView sum
func ListPostCountView(q Queryer, gapID string, offset int, a, b time.Time) ([]*entity.PostStatisticModel, error) {

	rows, err := q.Query(`
	 SELECT post.id, post.slug, post.title, post.description,
			post.created_at, COALESCE(view.c, 0), COALESCE(guest.c, 0), COALESCE(likepost.c, 0),
			COALESCE(comment.c, 0)
	   FROM (SELECT id FROM gap WHERE id = $1) as gap
  LEFT JOIN post
 		 ON post.owner_id = gap.id AND post.status = true AND post.used = true
  LEFT JOIN (SELECT post_id, count(*) as c FROM view WHERE created_at >= $2 AND created_at <= $3 GROUP BY post_id) as view
		 ON view.post_id = post.id
  LEFT JOIN (SELECT post_id, count(*) as c FROM guest WHERE created_at >= $2 AND created_at <= $3 GROUP BY post_id) as guest
	 	 ON guest.post_id = post.id
  LEFT JOIN (SELECT post_id, count(*) as c FROM likepost WHERE created_at >= $2 AND created_at <= $3 AND status = true AND used = true GROUP BY post_id) as likepost
  		 ON likepost.post_id = post.id
  LEFT JOIN (SELECT post_id, count(*) as c FROM comment WHERE created_at >= $2 AND created_at <= $3 AND status = true AND used = true GROUP BY post_id) as comment
		 ON comment.post_id = post.id
	  WHERE view.c > 0 OR guest.c > 0
   ORDER BY post.created_at DESC
	 OFFSET $4
	  LIMIT $5
		`, gapID, a, b, offset, config.LimitListPostCountView)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lp []*entity.PostStatisticModel
	for rows.Next() {

		var t time.Time
		var p entity.PostStatisticModel
		err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description,
			&t, &p.View, &p.GuestView, &p.Like,
			&p.Comment)
		if err != nil {
			return nil, err
		}

		p.Title = service.ShortTextTitleStripHTML(config.LimitTitlePostCountView, p.Title)

		if p.Title == "" {
			p.Description = service.ShortTextTitleStripHTML(config.LimitTitlePostCountView, p.Description)
			p.Title = p.Description
			if p.Description == "" {
				p.Title = p.Slug
			}
		}

		p.All = p.Sum()
		p.CreatedAt = service.FormatPostViewCount(t)

		lp = append(lp, &p)
	}

	return lp, nil
}

//ListPostCountUserAgentView return model view and guestView sum
func ListPostCountUserAgentView(q Queryer, gapID string, a, b time.Time) (*entity.CountUserAgent, error) {

	rows, err := q.Query(`
	  SELECT view.user_agent, count(DISTINCT view.*)
		FROM (SELECT id FROM gap WHERE id = $1) as gap
   LEFT JOIN post
		  ON post.owner_id = gap.id AND post.status = true AND post.used = true
   LEFT JOIN view
		  ON view.post_id = post.id AND view.created_at >= $2 AND view.created_at <= $3
	   WHERE view.created_at >= $2 AND view.created_at <= $3
	GROUP BY view.user_agent
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

//ListPostCountUserAgentGuestView return model guestView sum count userAgent
func ListPostCountUserAgentGuestView(q Queryer, gapID string, a, b time.Time) (*entity.CountUserAgent, error) {

	rows, err := q.Query(`
	  SELECT guest.user_agent, count(DISTINCT guest.*)
		FROM (SELECT id FROM gap WHERE id = $1) as gap
   LEFT JOIN post
		  ON post.owner_id = gap.id AND post.status = true AND post.used = true
   LEFT JOIN guest
		  ON guest.post_id = post.id AND guest.created_at >= $2 AND guest.created_at <= $3
	   WHERE guest.created_at >= $2 AND guest.created_at <= $3
	GROUP BY guest.user_agent
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

//ListCountViewHour return view hour
func ListCountViewHour(q Queryer, gapID string, a, b time.Time) ([]*entity.CountViewHour, error) {

	rows, err := q.Query(`
	  SELECT COALESCE(extract(hour from view.created_at at time zone 'UTC+7' at time zone 'UTC'), 0) as date, count(DISTINCT  view.*)
	    FROM (SELECT id FROM gap WHERE id = $1) as gap
   LEFT JOIN post
	  	  ON post.owner_id = gap.id AND post.status = true AND post.used = true
   LEFT JOIN view
		  ON view.post_id = post.id AND view.created_at >= $2 AND view.created_at <= $3
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

//ListCountGuestViewHour return view hour
func ListCountGuestViewHour(q Queryer, gapID string, a, b time.Time) ([]*entity.CountViewHour, error) {

	rows, err := q.Query(`
	  SELECT COALESCE(extract(hour from guest.created_at at time zone 'UTC+7' at time zone 'UTC'), 0) as date, count(DISTINCT  guest.*)
	    FROM (SELECT id FROM gap WHERE id = $1) as gap
   LEFT JOIN post
	  	  ON post.owner_id = gap.id AND post.status = true AND post.used = true
   LEFT JOIN guest
		  ON guest.post_id = post.id AND guest.created_at >= $2 AND guest.created_at <= $3
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

//GetCountViewRevenue return model view and guestView sum
func GetCountViewRevenue(q Queryer, gapID string, t time.Time) (*entity.GetViewRevenueModel, error) {

	var p entity.GetViewRevenueModel

	err := q.QueryRow(`
		SELECT COALESCE(sum(view.c), 0)::int, COALESCE(sum(guest.c), 0)::int
		  FROM (SELECT id FROM gap WHERE id = $1) as gap
	 LEFT JOIN post
	 		ON post.owner_id = gap.id AND post.status = true AND post.used = true AND post.status_revenue != 2
	 LEFT JOIN (SELECT post_id, count(*) as c FROM view WHERE status_withdraw = $2 AND created_at < $3 GROUP BY post_id) as view
	   		ON view.post_id = post.id
	 LEFT JOIN (SELECT post_id, count(*) as c FROM guest WHERE status_withdraw = $2 AND created_at < $3 GROUP BY post_id) as guest
			ON guest.post_id = post.id

		`, gapID, false, t).Scan(&p.View, &p.GuestView)
	if err != nil {
		return nil, err
	}

	p.All = p.Sum()

	return &p, nil
}

//ListPostCountViewRevenue return model view and guestView sum
func ListPostCountViewRevenue(q Queryer, gapID string, limit int) ([]*entity.PostRevenueModel, error) {

	rows, err := q.Query(`
	 SELECT COALESCE(post.id, ''), COALESCE(post.slug, ''), COALESCE(post.title, ''), COALESCE(post.description, ''),
			COALESCE(post.created_at, now()), COALESCE(post.status_revenue, 0), COALESCE(post.image_url, ''), COALESCE(view.c, 0),
			COALESCE(guest.c, 0), COALESCE(post_note.note, '')
	   FROM (SELECT * FROM post WHERE owner_id = $1 AND status = true AND used = true ORDER BY created_at DESC LIMIT $2) as post
  LEFT JOIN (SELECT post_id, count(*) as c FROM view WHERE status_withdraw = $3 GROUP BY post_id) as view
		 ON view.post_id = post.id
  LEFT JOIN (SELECT post_id, count(*) as c FROM guest WHERE status_withdraw = $3 GROUP BY post_id) as guest
		 ON guest.post_id = post.id
  LEFT JOIN post_note
	 	 ON post_note.post_id = post.id
   ORDER BY post.created_at DESC
		`, gapID, limit, false)
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

//ListPostCountViewRevenueNextLoad return model view and guestView sum
func ListPostCountViewRevenueNextLoad(q Queryer, gapID string, timeLoad time.Time, limit int) ([]*entity.PostRevenueModel, error) {

	rows, err := q.Query(`
	 SELECT COALESCE(post.id, ''), COALESCE(post.slug, ''), COALESCE(post.title, ''), COALESCE(post.description, ''),
			COALESCE(post.created_at, now()), COALESCE(post.status_revenue, 2), COALESCE(post.image_url, ''), COALESCE(view.c, 0),
			COALESCE(guest.c, 0), COALESCE(post_note.note, '')
	   FROM (SELECT * FROM post WHERE owner_id = $1 AND created_at < $2 AND status = true AND used = true ORDER BY created_at DESC LIMIT $3) as post
  LEFT JOIN (SELECT post_id, count(*) as c FROM view WHERE status_withdraw = $4 GROUP BY post_id) as view
		 ON view.post_id = post.id
  LEFT JOIN (SELECT post_id, count(*) as c FROM guest WHERE status_withdraw = $4 GROUP BY post_id) as guest
		 ON guest.post_id = post.id
  LEFT JOIN post_note
	 	 ON post_note.post_id = post.id
   ORDER BY post.created_at DESC
		`, gapID, timeLoad, config.LimitListPostCountView, false)
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

//GetShortenerURL input postID return (code, error)
func GetShortenerURL(q Queryer, postID string) (string, error) {
	var code string
	err := q.QueryRow(`
			SELECT code
			  FROM shortener_url
			 WHERE post_id = $1
	`, postID).Scan(&code)

	return code, err
}

//CheckCodeShortenerURL input code return err
func CheckCodeShortenerURL(q Queryer, code string) error {
	var id string
	err := q.QueryRow(`
			SELECT id
			  FROM shortener_url
			 WHERE code = $1
	`, code).Scan(&id)

	return err
}

//InsertShortenerURL insert data return err
func InsertShortenerURL(q Queryer, code, postID, link string) error {
	var returnID string
	err := q.QueryRow(`
	INSERT INTO shortener_url
				(post_id, code, link)
		 VALUES ($1, $2, $3) RETURNING id;
`, postID, code, link).Scan(&returnID)

	return err
}

// CreatePostModel is struct model create
type CreatePostModel struct {
	Slug            string
	OwnerID         string
	Onair           time.Time
	Title           string
	TagTopics       []TagTopic
	Description     string
	LinkDescription string
	ImageURL        string
	ImageURLMobile  string
	Height          int
	Width           int
	ImageShareURL   string
	HeightShare     int
	WidthShare      int
	Link            string
	VdoURL          string
	Type            entity.TypePost
	Province        int
	StatusVerify    int
}

// EditPostModel is struct model edit
type EditPostModel struct {
	ID              string
	Title           string
	TagTopics       []TagTopic
	Description     string
	LinkDescription string
	ImageURL        string
	ImageURLMobile  string
	Height          int
	Width           int
	ImageShareURL   string
	HeightShare     int
	WidthShare      int
	Link            string
	VdoURL          string
	Type            entity.TypePost
	Province        int
	StatusVerify    int
}

//TagTopic create model
type TagTopic struct {
	TopicID string `json:"id"`
	Name    string `json:"tag"`
}

type tagTopic struct {
	ID      string
	TopicID string
}
