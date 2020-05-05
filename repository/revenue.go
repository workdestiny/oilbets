package repository

import (
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/workdestiny/amlporn/entity"
)

// CheckRevenue is get pending
func CheckRevenue(q Queryer, gapID string) (bool, error) {

	var s int
	err := q.QueryRow(`
		SELECT status
		  FROM revenue_beta
		 WHERE gap_id = $1 AND status = $2
	`, gapID, entity.Pending).Scan(&s)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return false, nil
}

// RegisterRevenue is register new revenue
func RegisterRevenue(q Queryer, t time.Time, gapID, total, viewRate, guestViewRate string) error {

	_, err := q.Exec(`
	INSERT INTO revenue_beta
				(gap_id, total, count_view, status,
				view_rate, guest_view_rate, count_guest_view, created_at,
			 	updated_at)
	     VALUES ($1, $2, $3, $4,
				$5, $6, $7, $8,
				$8);`,
		gapID, total, 0, 0,
		viewRate, guestViewRate, 0, t)
	if err != nil {
		return err
	}

	return nil
}

//GetGapWallets get gap by admin
func GetGapWallets(q Queryer, gapID string) (*entity.GetGapRevenueModel, error) {
	var g entity.GetGapRevenueModel
	err := q.QueryRow(`
		SELECT gap.id, gap.name->>'text', gap.display->>'mini', gap.username->>'text',
			   COALESCE(wallets.saving, '0.00')
		  FROM gap
	 LEFT JOIN wallets
	 		ON wallets.gap_id = gap.id
		 WHERE (gap.id = $1 OR gap.username->>'text' = $1);
		`, gapID).Scan(&g.ID, &g.Name, &g.Display, &g.Username,
		&g.Wallets)

	if err != nil {
		return nil, err
	}

	return &g, nil
}

//GetRevenue get a revenues
func GetRevenue(q Queryer, revenueID string) (*entity.RevenueModel, error) {
	var r entity.RevenueModel
	err := q.QueryRow(`
	 SELECT revenue.id, revenue.total, revenue.created_at, gap.id,
	     	gap.name->>'text', gap.display->>'mini', gap.username->>'text'
	   FROM revenue_beta as revenue
  LEFT JOIN gap
 		 ON gap.id = revenue.gap_id
	  WHERE revenue.status = $1 AND revenue.id = $2
		`, entity.Pending, revenueID).Scan(&r.ID, &r.Total, &r.CreatedAt, &r.Gap.ID,
		&r.Gap.Name, &r.Gap.Display, &r.Gap.Username)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

//ListRevenue list revenues
func ListRevenue(q Queryer) ([]*entity.RevenueModel, error) {
	rows, err := q.Query(`
		SELECT revenue.id, revenue.total, revenue.created_at, gap.id,
			   gap.name->>'text', gap.display->>'mini', gap.username->>'text'
		  FROM revenue_beta as revenue
	 LEFT JOIN gap
			ON gap.id = revenue.gap_id
		 WHERE revenue.status = $1
	  ORDER BY revenue.created_at ASC
	`, entity.Pending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lr []*entity.RevenueModel
	for rows.Next() {

		var r entity.RevenueModel
		err := rows.Scan(&r.ID, &r.Total, &r.CreatedAt, &r.Gap.ID,
			&r.Gap.Name, &r.Gap.Display, &r.Gap.Username)
		if err != nil {
			return nil, err
		}

		lr = append(lr, &r)
	}

	return lr, nil
}

//SearchListRevenue Search list revenues
func SearchListRevenue(q Queryer, text string) ([]*entity.RevenueModel, error) {
	rows, err := q.Query(`
	SELECT revenue.id, revenue.total, revenue.created_at, gap.id,
	gap.name->>'text', gap.display->>'mini', gap.username->>'text'
	FROM revenue_beta as revenue
	LEFT JOIN gap
	ON gap.id = revenue.gap_id
	WHERE revenue.status = $1 AND (gap.name->>'text' LIKE $2 OR gap.id = $3)
	ORDER BY revenue.created_at ASC
	`, entity.Pending, text+"%", text)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lr []*entity.RevenueModel
	for rows.Next() {

		var r entity.RevenueModel
		err := rows.Scan(&r.ID, &r.Total, &r.CreatedAt, &r.Gap.ID,
			&r.Gap.Name, &r.Gap.Display, &r.Gap.Username)
		if err != nil {
			return nil, err
		}

		lr = append(lr, &r)
	}

	return lr, nil
}

// AdminUpdateRevenue is add Wallet
func AdminUpdateRevenue(q Queryer, revenueID, total string, countView, countGuestView int64, status entity.RevenueStatus) error {

	_, err := q.Exec(`
		UPDATE revenue_beta
		SET total = $1, count_view = $2, count_guest_view = $3, status = $4
		WHERE id = $5
		`, total, countView, countGuestView, status, revenueID)
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdateViewRevenue is update view revenue
func AdminUpdateViewRevenue(q Queryer, postIDs []string, revenueID string, revenueTime time.Time) error {

	_, err := q.Exec(`
		UPDATE view
		SET revenue_id = $1, status_withdraw = $2
		WHERE post_id = ANY($3) AND created_at < $4
		`, revenueID, true, pq.Array(postIDs), revenueTime)
	if err != nil {
		return err
	}

	return nil
}

// AdminUpdateGuestViewRevenue is update guest view revenue
func AdminUpdateGuestViewRevenue(q Queryer, postIDs []string, revenueID string, revenueTime time.Time) error {

	_, err := q.Exec(`
	UPDATE guest
	SET revenue_id = $1, status_withdraw = $2
	WHERE post_id = ANY($3) AND created_at < $4
	`, revenueID, true, pq.Array(postIDs), revenueTime)
	if err != nil {
		return err
	}

	return nil
}

//AdminListPostID list post ids
func AdminListPostID(q Queryer, gapID string) ([]string, error) {
	rows, err := q.Query(`
		SELECT id
		  FROM post
		 WHERE owner_id = $1
	  ORDER BY created_at DESC
	`, gapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {

		var id string
		err := rows.Scan(&id)
		if err != nil {
			return ids, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
