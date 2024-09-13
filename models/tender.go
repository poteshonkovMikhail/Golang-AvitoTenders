package models

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	TENDER_CREATED = "CREATED"
)

type Tender struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Status  string `json:"status"`
	CreateTenderStruct
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	CreatorUsername string    `json:"creator_username"`
}

type CreateTenderStruct struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	ServiceType    string `json:"service_type"`
	OrganizationID string `json:"organization_id"`
}

func GetTenders(db *pgxpool.Pool) ([]Tender, error) {
	rows, err := db.Query(context.Background(), "SELECT id, version, name, description, service_type, status, created_at, updated_at, organization_id, creator_username FROM tenders WHERE status='PUBLISHED'")
	if err != nil {
		log.Printf("%v", err)

		return nil, err
	}
	defer rows.Close()

	var tenders []Tender
	for rows.Next() {
		var tender Tender
		if err := rows.Scan(&tender.ID, &tender.Version, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.CreatedAt, &tender.UpdatedAt, &tender.OrganizationID, &tender.CreatorUsername); err != nil {
			log.Printf("%v", err)

			return nil, err
		}
		tenders = append(tenders, tender)
	}

	return tenders, rows.Err()
}

func GetTenderByID(db *pgxpool.Pool, id string) (*Tender, error) {
	row := db.QueryRow(context.Background(), "SELECT id, version, name, description, service_type, status, created_at, updated_at, organization_id, creator_username FROM tenders WHERE id=$1 AND status='PUBLISHED'", id)
	tender := &Tender{}
	err := row.Scan(&tender.ID, &tender.Version, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.CreatedAt, &tender.UpdatedAt, &tender.OrganizationID, &tender.CreatorUsername)
	if err != nil {
		return nil, err
	}
	return tender, nil
}

func GetUserTenders(db *pgxpool.Pool, username string) ([]Tender, error) {
	rows, err := db.Query(context.Background(), "SELECT id, version, name, description, service_type, status, created_at, updated_at, organization_id, creator_username FROM tenders WHERE creator_username=$1 AND status='PUBLISHED'", username)
	if err != nil {
		log.Printf("%v", err)

		return nil, err
	}
	defer rows.Close()

	var tenders []Tender
	for rows.Next() {
		var tender Tender
		if err := rows.Scan(&tender.ID, &tender.Version, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.CreatedAt, &tender.UpdatedAt, &tender.OrganizationID, &tender.CreatorUsername); err != nil {
			log.Printf("%v", err)

			return nil, err
		}
		tenders = append(tenders, tender)
	}

	return tenders, rows.Err()
}

func CreateTender(db *pgxpool.Pool, name, description, service_type, organizationID, creator_username string) (*Tender, error) {
	var tender Tender

	err := db.QueryRow(context.Background(), "INSERT INTO tenders (name, description, service_type, status, organization_id, creator_username) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, version, name, description, service_type, status, created_at, updated_at, organization_id, creator_username", name, description, service_type, TENDER_CREATED, organizationID, creator_username).Scan(&tender.ID, &tender.Version, &tender.Name, &tender.Description, &tender.ServiceType, &tender.Status, &tender.CreatedAt, &tender.UpdatedAt, &tender.OrganizationID, &tender.CreatorUsername)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	return &tender, nil
}

func UpdateTender(db *pgxpool.Pool, id, name, description, service_type, status string) (*Tender, error) {
	tender, err := GetTenderByID(db, id)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}

	err = db.QueryRow(context.Background(), "UPDATE tenders SET name=$1, description=$2, service_type=$3, status=$4, version=version WHERE id=$5 RETURNING id, version, name, description, created_at, updated_at", name, description, service_type, status, id).Scan(&tender.ID, &tender.Version, &tender.Name, &tender.Description, &tender.CreatedAt, &tender.UpdatedAt)
	if err != nil {
		log.Printf("%v", err)
		return nil, err
	}
	return tender, nil
}

func DeleteTender(db *pgxpool.Pool, id int) error {
	_, err := db.Exec(context.Background(), "DELETE FROM tenders WHERE id=$1", id)
	return err
}

func RollbackTenderToVersion(db *pgxpool.Pool, id string, version int) error {
	row := db.QueryRow(context.Background(), "SELECT tender_id, version, name, description, created_at, updated_at, creator_username, service_type FROM tender_versions WHERE tender_id=$1 and version=$2", id, version)
	tender := &Tender{}
	err := row.Scan(&tender.ID, &tender.Version, &tender.Name, &tender.Description, &tender.CreatedAt, &tender.UpdatedAt, &tender.CreatorUsername, &tender.ServiceType)
	if err != nil {
		return err
	}
	// Начинаем транзакцию
	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Fatalf("Ошибка при начале транзакции по откату к версии %d тендера: %v\n", version, err)
		return err
	}

	// Операция обновления
	updateQuery := `UPDATE tenders SET name=$1, description=$2, service_type=$3 WHERE id=$4`
	if _, err := tx.Exec(context.Background(), updateQuery, tender.Name, tender.Description, tender.ServiceType, id); err != nil {
		tx.Rollback(context.Background())
		log.Fatalf("Ошибка при выполнении запроса обновления: %v\n", err)
		return err
	}

	// Операция удаления
	deleteQuery := `DELETE FROM tender_versions WHERE tender_id=$1 AND version=$2`
	if _, err := tx.Exec(context.Background(), deleteQuery, id, version); err != nil {
		tx.Rollback(context.Background())
		log.Fatalf("Ошибка при выполнении запроса удаления: %v\n", err)
		return err
	}

	// Фиксируем/коммитим транзакцию
	if err := tx.Commit(context.Background()); err != nil {
		log.Fatalf("Ошибка при коммите транзакции по откату к версии %d тендера: %v\n", version, err)
		return err
	}

	return nil
}
