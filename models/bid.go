package models

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	BID_CREATED = "CREATED"
)

type Bid struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Status  string `json:"status"`
	CreateBidStruct
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatorUsername string    `json:"creator_username"`
}

type CreateBidStruct struct {
	TenderID    string `json:"tender_id"`
	Amount      int    `json:"amount"`
	Description string `json:"description"`
}

type Review struct {
	ID             string    `json:"id"`
	TenderID       string    `json:"tender_id"`
	BidID          string    `json:"bid_id"` // Добавлено поле BidID
	AuthorUsername string    `json:"author_username"`
	Rating         int       `json:"rating"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateReviewInput struct {
	//TenderID string `json:"tender_id" binding:"required"`
	//BidID    string `json:"bid_id" binding:"required"`
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment" binding:"required"`
}

func GetBids(db *pgxpool.Pool) ([]Bid, error) {
	rows, err := db.Query(context.Background(), "SELECT id, version, tender_id, amount, description, status, created_at, updated_at, creator_username FROM bids")
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	defer rows.Close()

	var bids []Bid
	for rows.Next() {
		var bid Bid
		if err := rows.Scan(&bid.ID, &bid.Version, &bid.TenderID, &bid.Amount, &bid.Description, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt, &bid.CreatorUsername); err != nil {
			log.Printf("%v", err)
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, rows.Err()
}

func GetBidByID(db *pgxpool.Pool, id string) (*Bid, error) {
	row := db.QueryRow(context.Background(), "SELECT id, version, tender_id, amount, description, status, created_at, updated_at, creator_username FROM bids WHERE id=$1 AND status='PUBLISHED'", id)
	bid := &Bid{}
	err := row.Scan(&bid.ID, &bid.Version, &bid.TenderID, &bid.Amount, &bid.Description, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt, &bid.CreatorUsername)
	if err != nil {
		log.Printf("%d", err)
		return nil, err
	}
	return bid, nil
}

func GetUserBids(db *pgxpool.Pool, username string) ([]Bid, error) {
	rows, err := db.Query(context.Background(), "SELECT id, version, tender_id, amount, description, status, created_at, updated_at FROM bids WHERE creator_username=$1 AND status='PUBLISHED'", username)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	defer rows.Close()

	var bids []Bid
	for rows.Next() {
		var bid Bid
		if err := rows.Scan(&bid.ID, &bid.Version, &bid.TenderID, &bid.Amount, &bid.Description, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt); err != nil {
			log.Printf("%v", err)
			return nil, err
		}
		bids = append(bids, bid)
	}

	return bids, rows.Err()
}

func CreateBid(db *pgxpool.Pool, tenderID string, amount int, description, creatorUsername string) (*Bid, error) {
	var bid Bid

	err := db.QueryRow(context.Background(), "INSERT INTO bids (tender_id, amount, description, status, creator_username) VALUES ($1, $2, $3, $4, $5) RETURNING id, version, tender_id, amount, description, status, created_at, updated_at, creator_username", tenderID, amount, description, BID_CREATED, creatorUsername).Scan(&bid.ID, &bid.Version, &bid.TenderID, &bid.Amount, &bid.Description, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt, &bid.CreatorUsername)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	return &bid, nil
}

func UpdateBid(db *pgxpool.Pool, id string, amount int, description, status string) (*Bid, error) {
	bid, err := GetBidByID(db, id)
	if err != nil {
		log.Printf("%d", err)
		return nil, err
	}

	err = db.QueryRow(context.Background(), "UPDATE bids SET amount=$1, description=$2, status=$3, version=version WHERE id=$4 RETURNING id, version, tender_id, amount, description, created_at, updated_at", amount, description, status, id).Scan(&bid.ID, &bid.Version, &bid.TenderID, &bid.Amount, &bid.Description, &bid.CreatedAt, &bid.UpdatedAt)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	return bid, nil
}

func DeleteBid(db *pgxpool.Pool, id string) error {
	_, err := db.Exec(context.Background(), "DELETE FROM bids WHERE id=$1", id)
	return err
}

func RollbackBidToVersion(db *pgxpool.Pool, id string, version int, newCreatorUsername string) error {
	row := db.QueryRow(context.Background(), "SELECT bid_id, version, amount, description, created_at, updated_at, creator_username FROM bid_versions WHERE bid_id=$1 and version=$2", id, version)
	bid := &Bid{}
	err := row.Scan(&bid.ID, &bid.Version, &bid.Amount, &bid.Description, &bid.CreatedAt, &bid.UpdatedAt, &bid.CreatorUsername)
	if err != nil {
		log.Printf("%d", err)
		return err
	}
	// Начинаем транзакцию
	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Printf("Ошибка при начале транзакции по откату к версии %d бидa: %v\n", version, err)
		return err
	}

	// Операция обновления
	updateQuery := `UPDATE bids SET amount=$1, description=$2, version=$3, creator_username=$4 WHERE id=$5`
	if _, err := tx.Exec(context.Background(), updateQuery, bid.Amount, bid.Description, version, newCreatorUsername, id); err != nil {
		tx.Rollback(context.Background())
		log.Printf("Ошибка при выполнении запроса обновления: %v\n", err)
		return err
	}

	// Операция удаления
	deleteQuery := `DELETE FROM bid_versions WHERE bid_id=$1 AND version=$2`
	if _, err := tx.Exec(context.Background(), deleteQuery, id, version); err != nil {
		tx.Rollback(context.Background())
		log.Printf("Ошибка при выполнении запроса удаления: %v\n", err)
		return err
	}

	// Фиксируем/коммитим транзакцию
	if err := tx.Commit(context.Background()); err != nil {
		log.Printf("Ошибка при коммите транзакции по откату к версии %d бидa: %v\n", version, err)
		return err
	}

	return nil
}

func FetchBidsForTender(db *pgxpool.Pool, tenderId string) ([]Bid, error) {
	var bids []Bid
	query := "SELECT id, version, tender_id, amount, description, status, created_at, updated_at, creator_username FROM bids WHERE tender_id=$1 AND status='PUBLISHED'"
	rows, err := db.Query(context.Background(), query, tenderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bid Bid
		if err := rows.Scan(&bid.ID, &bid.Version, &bid.TenderID, &bid.Amount, &bid.Description, &bid.Status, &bid.CreatedAt, &bid.UpdatedAt, &bid.CreatorUsername); err != nil {
			return nil, err
		}
		bids = append(bids, bid)
	}
	return bids, nil
}

func FetchBidReviews(db *pgxpool.Pool, tenderId string) ([]Review, error) {
	var reviews []Review
	query := "SELECT id, tender_id, bid_id, author_username, comment, rating, created_at FROM reviews WHERE bid_id IN (SELECT id FROM bids WHERE tender_id=$1)" // Обновлен запрос для выборки reviews с учетом BidID
	rows, err := db.Query(context.Background(), query, tenderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ID, &review.TenderID, &review.BidID, &review.AuthorUsername, &review.Comment, &review.Rating, &review.CreatedAt); err != nil { // Добавлено поле BidID
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil
}

func CreateReview(db *pgxpool.Pool, review CreateReviewInput, authorUsername, tenderId, bidId string) (Review, error) {
	var newReview Review

	query := `
    INSERT INTO reviews (tender_id, bid_id, author_username, rating, comment)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id, tender_id, bid_id, author_username, rating, comment, created_at, updated_at
    `

	err := db.QueryRow(context.Background(), query,
		tenderId, bidId, authorUsername, review.Rating, review.Comment).Scan(
		&newReview.ID,
		&newReview.TenderID,
		&newReview.BidID,
		&newReview.AuthorUsername,
		&newReview.Rating,
		&newReview.Comment,
		&newReview.CreatedAt,
		&newReview.UpdatedAt,
	)

	if err != nil {
		return newReview, err
	}

	return newReview, nil
}
