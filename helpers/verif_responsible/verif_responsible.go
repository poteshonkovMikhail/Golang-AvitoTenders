package verif_responsible

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func VerifResponsible(db *pgxpool.Pool, organizationId, userId string) (bool, error) {
	var exists bool
	err := db.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM organization_responsible WHERE organization_id=$1 AND user_id=$2)", organizationId, userId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func VerifBidEditor(db *pgxpool.Pool, bidId string, userId string) (bool, error) {
	var exists bool
	var creatorUsername string
	var creatorId string
	err := db.QueryRow(context.Background(), "SELECT creator_username FROM bids WHERE bid_id=$1)", bidId).Scan(&creatorUsername)
	if err != nil {
		return false, err
	}

	err = db.QueryRow(context.Background(), "SELECT id FROM employee WHERE username=$1)", creatorUsername).Scan(&creatorId)
	if err != nil {
		return false, err
	}

	rows, err := db.Query(context.Background(), "SELECT organization_id FROM organization_responsible WHERE user_id=$1", creatorId)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var organization_id string
		if err := rows.Scan(&organization_id); err != nil {
			return false, err
		}
		err = db.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM organization_responsible WHERE organization_id=$1 AND user_id=$2)", organization_id, userId).Scan(&exists)
		if err != nil {
			return false, err
		}
		if exists {
			continue
		}
	}

	return exists, nil
}
