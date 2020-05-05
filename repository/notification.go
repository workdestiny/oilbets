package repository

import (
	"database/sql"
	"log"
	"time"

	"github.com/acoshift/pgsql"
	"github.com/go-redis/redis"
	"github.com/workdestiny/amlporn/config"
	"github.com/workdestiny/amlporn/entity"
	"github.com/workdestiny/amlporn/service"
)

//ListNotification list all myNotification
func ListNotification(q Queryer, userID string, limit int) ([]*entity.NotificationModel, error) {
	rows, err := q.Query(`
			SELECT notification.id, notification.type, notification.created_at, notification.count,
				   notification.read, users.id, users.firstname,
				   users.display->>'mini', gap.id, gap.name->>'text', gap.display->>'mini',
				   gap.username->>'text', COALESCE(post.title, ''), COALESCE(post.id, ''), notification.comment_text
			  FROM (SELECT *
					FROM notification
					WHERE notification.owner_id = $1 AND used = $2 AND main = $3
					ORDER BY created_at DESC
					LIMIT $4) as notification
		 LEFT JOIN users
				ON users.id = notification.user_id
		 LEFT JOIN gap
				ON gap.id = notification.gap_id
		 LEFT JOIN post
				ON post.id = notification.post_id
		  ORDER BY notification.created_at DESC;
		`, userID, true, true, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ln []*entity.NotificationModel
	for rows.Next() {

		var firstname string
		var t time.Time
		var n entity.NotificationModel
		err := rows.Scan(&n.ID, &n.Type, &t, &n.Count,
			&n.Read, &n.User.ID, &firstname,
			&n.User.Display, &n.Gap.ID, &n.Gap.Name, &n.Gap.Display,
			&n.Gap.Username, &n.Title, &n.PostID, &n.CommentText)
		if err != nil {
			return nil, err
		}

		if n.Gap.Username == "" {
			n.Gap.Username = n.Gap.ID
		}

		n.Title = service.ShortTextNoti(config.LimitTitleNotification, n.Title)
		n.CommentText = service.UnescapeString(n.CommentText)
		n.User.Name = service.ShortTextNoti(config.LimitNameNotification, firstname)
		n.Gap.Name = service.ShortTextNoti(config.LimitNameGapNotification, n.Gap.Name)
		n.CreatedAt = t
		n.Time = service.ReFormatTime(t)

		ln = append(ln, &n)
	}

	return ln, nil
}

//ListNotificationNextLoad list all myNotification
func ListNotificationNextLoad(q Queryer, userID string, timeLoad time.Time, limit int) ([]*entity.NotificationModel, error) {
	rows, err := q.Query(`
			SELECT notification.id, notification.type, notification.created_at, notification.count,
				   notification.read, users.id, users.firstname,
				   users.display->>'mini', gap.id, gap.name->>'text', gap.display->>'mini',
				   gap.username->>'text', COALESCE(post.title, ''), COALESCE(post.id, ''), notification.comment_text
			  FROM (SELECT *
					FROM notification
					WHERE notification.owner_id = $1 AND used = $2 AND main = $3 AND created_at < $4
					ORDER BY created_at DESC
					LIMIT $5) as notification
		 LEFT JOIN users
				ON users.id = notification.user_id
		 LEFT JOIN gap
				ON gap.id = notification.gap_id
		 LEFT JOIN post
				ON post.id = notification.post_id
		  ORDER BY notification.created_at DESC;
		`, userID, true, true, timeLoad, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ln []*entity.NotificationModel
	for rows.Next() {

		var t time.Time
		var firstname string
		var n entity.NotificationModel
		err := rows.Scan(&n.ID, &n.Type, &t, &n.Count,
			&n.Read, &n.User.ID, &firstname,
			&n.User.Display, &n.Gap.ID, &n.Gap.Name, &n.Gap.Display,
			&n.Gap.Username, &n.Title, &n.PostID, &n.CommentText)
		if err != nil {
			return nil, err
		}

		if n.Gap.Username == "" {
			n.Gap.Username = n.Gap.ID
		}

		n.Title = service.ShortTextNoti(config.LimitTitleNotification, n.Title)
		n.CommentText = service.UnescapeString(n.CommentText)
		n.User.Name = service.ShortTextNoti(config.LimitNameNotification, firstname)
		n.Gap.Name = service.ShortTextNoti(config.LimitNameGapNotification, n.Gap.Name)
		n.CreatedAt = t
		n.Time = service.ReFormatTime(t)

		ln = append(ln, &n)
	}

	return ln, nil
}

//ListNotificationType list all myNotification type (input)
func ListNotificationType(q Queryer, userID string, t entity.NotificationType, limit int) ([]*entity.NotificationModel, error) {
	rows, err := q.Query(`
			SELECT notification.id, notification.type, notification.created_at, notification.count,
				   notification.read, users.id, users.firstname, users.lastname,
				   users.display->>'mini', gap.id, gap.name->>'text', gap.display->>'mini',
				   gap.username->>'text', COALESCE(post.title, ''), COALESCE(post.id, ''), notification.comment_text
			  FROM (SELECT *
					FROM notification
					WHERE notification.owner_id = $1 AND used = $2 AND main = $3 AND type = $4
					ORDER BY created_at DESC
					LIMIT $5) as notification
		 LEFT JOIN users
				ON users.id = notification.user_id
		 LEFT JOIN gap
				ON gap.id = notification.gap_id
		 LEFT JOIN post
				ON post.id = notification.post_id
		  ORDER BY notification.created_at DESC;
		`, userID, true, true, t, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ln []*entity.NotificationModel
	for rows.Next() {

		var t time.Time
		var firstname, lastname string
		var n entity.NotificationModel
		err := rows.Scan(&n.ID, &n.Type, &t, &n.Count,
			&n.Read, &n.User.ID, &firstname, &lastname,
			&n.User.Display, &n.Gap.ID, &n.Gap.Name, &n.Gap.Display,
			&n.Gap.Username, &n.Title, &n.PostID, &n.CommentText)
		if err != nil {
			return nil, err
		}

		if n.Gap.Username == "" {
			n.Gap.Username = n.Gap.ID
		}

		n.Title = service.ShortTextNoti(config.LimitTitleNotificationType, n.Title)
		n.CommentText = service.UnescapeString(n.CommentText)
		n.User.Name = service.ShortTextNoti(config.LimitNameNotificationType, firstname+" "+lastname)
		n.Gap.Name = service.ShortTextNoti(config.LimitNameGapNotificationType, n.Gap.Name)
		n.CreatedAt = t
		n.Time = service.ReFormatTime(t)

		ln = append(ln, &n)
	}
	return ln, nil
}

//ListNotificationTypeNextLoad list all myNotification type (input)
func ListNotificationTypeNextLoad(q Queryer, userID string, t entity.NotificationType, timeLoad time.Time, limit int) ([]*entity.NotificationModel, error) {
	rows, err := q.Query(`
			SELECT notification.id, notification.type, notification.created_at, notification.count,
				   notification.read, users.id, users.firstname, users.lastname,
				   users.display->>'mini', gap.id, gap.name->>'text', gap.display->>'mini',
				   gap.username->>'text', COALESCE(post.title, ''), COALESCE(post.id, ''), notification.comment_text
			  FROM (SELECT *
					FROM notification
					WHERE notification.owner_id = $1 AND used = $2 AND main = $3 AND type = $4 AND created_at < $5
					ORDER BY created_at DESC
					LIMIT $6) as notification
		 LEFT JOIN users
				ON users.id = notification.user_id
		 LEFT JOIN gap
				ON gap.id = notification.gap_id
		 LEFT JOIN post
				ON post.id = notification.post_id
		  ORDER BY notification.created_at DESC;
		`, userID, true, true, t, timeLoad, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ln []*entity.NotificationModel
	for rows.Next() {

		var t time.Time
		var firstname, lastname string
		var n entity.NotificationModel
		err := rows.Scan(&n.ID, &n.Type, &t, &n.Count,
			&n.Read, &n.User.ID, &firstname, &lastname,
			&n.User.Display, &n.Gap.ID, &n.Gap.Name, &n.Gap.Display,
			&n.Gap.Username, &n.Title, &n.PostID, &n.CommentText)
		if err != nil {
			return nil, err
		}

		if n.Gap.Username == "" {
			n.Gap.Username = n.Gap.ID
		}

		n.Title = service.ShortTextNoti(config.LimitTitleNotificationType, n.Title)
		n.CommentText = service.UnescapeString(n.CommentText)
		n.User.Name = service.ShortTextNoti(config.LimitNameNotificationType, firstname+" "+lastname)
		n.Gap.Name = service.ShortTextNoti(config.LimitNameGapNotificationType, n.Gap.Name)
		n.CreatedAt = t
		n.Time = service.ReFormatTime(t)

		ln = append(ln, &n)
	}

	return ln, nil
}

//CountNotification is check count
func CountNotification(q Queryer, userID string) (int, error) {

	var count int
	err := q.QueryRow(`
		SELECT COUNT(*)
		  FROM (SELECT *
				FROM notification
				WHERE owner_id = $1 AND main = $2 AND used = $3 AND read = $4) as notification;
	`, userID, true, true, false).Scan(&count)

	return count, err
}

//CheckNotification is check like
func CheckNotification(q Queryer, notificationID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM notification
		 WHERE id = $1;
	`, notificationID).Scan(&id)

	return err
}

//CheckNotificationLike is check like
func CheckNotificationLike(q Queryer, userID string, postID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM notification
		 WHERE user_id = $1
		   AND post_id = $2
		   AND type = $3;
	`, userID, postID, entity.Like).Scan(&id)

	return err
}

//CheckNotificationComment is check comment
func CheckNotificationComment(q Queryer, userID string, postID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM notification
		 WHERE user_id = $1
		   AND post_id = $2
		   AND type = $3;
	`, userID, postID, entity.Comment).Scan(&id)

	return err
}

//CheckNotificationFollow is check follow
func CheckNotificationFollow(q Queryer, userID string, gapID string) error {

	var id string
	err := q.QueryRow(`
		SELECT id
		  FROM notification
		 WHERE user_id = $1
		   AND gap_id = $2
		   AND type = $3;
	`, userID, gapID, entity.Follow).Scan(&id)

	return err
}

//AddNotificationLike is insert
func AddNotificationLike(db *sql.DB, myRedis *redis.Client, userID string, postID string, gapID string) error {

	gap := getRedisGap(myRedis, gapID)

	if userID == gap.UserID {
		return nil
	}

	err := CheckNotificationLike(db, userID, postID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		return nil
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		var id string
		var count int
		var t time.Time
		err := tx.QueryRow(`
			SELECT id, count, created_at
			  FROM notification
			 WHERE post_id = $1
			   AND main = $2
			   AND read = $3
			   AND type = $4
		  ORDER BY created_at DESC;
		`, postID, true, false, entity.Like).Scan(&id, &count, &t)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if id != "" {

			duration := time.Now().Unix() - t.Unix()

			if duration < 86400 {
				_, err = tx.Exec(`
					UPDATE notification
					   SET main = $1, read = $2
					 WHERE id = $3;
				`, false, true, id)
				if err != nil {
					return err
				}

				count = count + 1
			}

			if duration > 86400 {
				count = 0
			}

		}

		_, err = tx.Exec(`
			INSERT INTO notification
						(owner_id, user_id, post_id, type,
						read, main, count, used,
						gap_id)
				 VALUES ($1, $2, $3, $4,
						$5, $6, $7, $8,
						$9);`,
			gap.UserID, userID, postID, entity.Like,
			false, true, count, true,
			gapID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			UPDATE users
			   SET notification = $1
			 WHERE id = $2;
		`, true, gap.UserID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//AddNotificationFollow is insert
func AddNotificationFollow(db *sql.DB, myRedis *redis.Client, userID string, gapID string) error {

	gap := getRedisGap(myRedis, gapID)

	if userID == gap.UserID {
		return nil
	}

	err := CheckNotificationFollow(db, userID, gapID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		return nil
	}

	err = pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

		var id string
		var count int
		var t time.Time
		err := tx.QueryRow(`
			SELECT id, count, created_at
			  FROM notification
			 WHERE gap_id = $1
			   AND main = $2
			   AND read = $3
			   AND type = $4
		  ORDER BY created_at DESC;
		`, gapID, true, false, entity.Follow).Scan(&id, &count, &t)
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		if id != "" {

			duration := time.Now().Unix() - t.Unix()

			if duration < 86400 {
				_, err = tx.Exec(`
					UPDATE notification
					   SET main = $1, read = $2
					 WHERE id = $3;
				`, false, true, id)
				if err != nil {
					return err
				}

				count = count + 1
			}

			if duration > 86400 {
				count = 0
			}

		}

		_, err = tx.Exec(`
			INSERT INTO notification
						(owner_id, user_id, post_id, type,
						read, main, count, used,
						gap_id)
				 VALUES ($1, $2, $3, $4,
						$5, $6, $7, $8,
						$9);`,
			gap.UserID, userID, "", entity.Follow,
			false, true, count, true,
			gapID)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			UPDATE users
			   SET notification = $1
			 WHERE id = $2;
		`, true, gap.UserID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

//AddNotificationComment is insert
func AddNotificationComment(db *sql.DB, myRedis *redis.Client, userID string, postID string, gapID string, text string) error {

	gap := getRedisGap(myRedis, gapID)

	// if userID == gap.UserID {
	// 	return nil
	// }
	IDs := ListUserIDOnCommentPost(db, postID, userID)
	IDs = append(IDs, gap.UserID)
	if userID != config.AdminID {
		IDs = append(IDs, config.AdminID)
	}
	if userID != config.AdminID2 {
		IDs = append(IDs, config.AdminID2)
	}

	for _, u := range IDs {

		err := pgsql.RunInTx(db, nil, func(tx *sql.Tx) error {

			var id string
			var count int
			var t time.Time
			err := tx.QueryRow(`
			SELECT id, count, created_at
			  FROM notification
			 WHERE post_id = $1
			   AND main = $2
			   AND user_id = $3
			   AND type = $4
			   AND owner_id = $5
		  ORDER BY created_at DESC;
		`, postID, true, userID, entity.Comment, u).Scan(&id, &count, &t)
			if err != nil && err != sql.ErrNoRows {
				return err
			}

			if id != "" {

				duration := time.Now().Unix() - t.Unix()

				if duration < 86400 {
					_, err = tx.Exec(`
					UPDATE notification
					   SET main = $1, read = $2
					 WHERE id = $3;
				`, false, true, id)
					if err != nil {
						return err
					}

					count = count + 1
				}

				if duration > 86400 {
					count = 0
				}

			}

			_, err = tx.Exec(`
			INSERT INTO notification
						(owner_id, user_id, post_id, type,
						read, main, count, used,
						gap_id, comment_text)
				 VALUES ($1, $2, $3, $4,
						$5, $6, $7, $8,
						$9, $10);`,
				u, userID, postID, entity.Comment,
				false, true, count, true,
				gapID, service.ShortText(config.LimitTextCommentNotification, text))
			if err != nil {
				return err
			}

			_, err = tx.Exec(`
			UPDATE users
			   SET notification = $1
			 WHERE id = $2;
		`, true, u)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

//ListUserIDOnCommentPost list count DESC
func ListUserIDOnCommentPost(q Queryer, postID, userID string) []string {

	rows, err := q.Query(`
			SELECT owner_id
			  FROM comment
			 WHERE post_id = $1 AND owner_id != $2 AND owner_type = 0 AND status = true AND used = true;
		`, postID, userID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var userIDs []string

	for rows.Next() {

		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil
		}

		userIDs = append(userIDs, userID)
	}

	return userIDs
}

//ReadNotification read a notification
func ReadNotification(q Queryer, userID string, notificationID string) error {
	_, err := q.Exec(`
		UPDATE notification
		   SET read = $1
		 WHERE owner_id = $2 AND id = $3;
	`, true, userID, notificationID)

	return err
}

//ReadAllNotification read all notification
func ReadAllNotification(q Queryer, userID string) error {
	_, err := q.Exec(`
		UPDATE notification
		   SET read = $1
		 WHERE owner_id = $2;
	`, true, userID)

	return err
}

//ResetNotification reset red bell notification
func ResetNotification(q Queryer, userID string) error {
	_, err := q.Exec(`
		UPDATE users
		   SET notification = $1
		 WHERE id = $2;
	`, false, userID)

	return err
}
