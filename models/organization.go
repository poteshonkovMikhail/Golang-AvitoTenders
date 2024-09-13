package models

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Organization struct {
	ID string `json:"id"`
	CreateOrganizationStruct
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOrganizationStruct struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func CreateOrganization(db *pgxpool.Pool, org CreateOrganizationStruct, creatorID string) error {
	var orgID string

	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Printf("Ошибка при начале транзакции по созданию организации: %v\n", err)
		return err
	}
	err = tx.QueryRow(context.Background(), "INSERT INTO organization (name, description, type) VALUES ($1, $2, $3) RETURNING id", org.Name, org.Description, org.Type).Scan(&orgID)
	if err != nil {
		log.Printf("Ошибка при вставке в table organization для создания организации: %v\n", err)
		return err
	}

	_, err = tx.Exec(context.Background(), "INSERT INTO organization_responsible (organization_id, user_id) VALUES ($1, $2)", orgID, creatorID)
	if err != nil {
		log.Printf("Ошибка при вставке в table organization_responsible для создания организации: %v\n", err)
		return err
	}

	if err := tx.Commit(context.Background()); err != nil {
		log.Fatalf("Ошибка при коммите транзакции по созданию организации: %v\n", err)
		return err
	}
	return nil
}

func GetOrganizations(db *pgxpool.Pool) ([]Organization, error) {
	rows, err := db.Query(context.Background(), "SELECT id, name, description, type, created_at, updated_at FROM organization")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var organizations []Organization
	for rows.Next() {
		var organization Organization
		if err := rows.Scan(&organization.ID, &organization.Name, &organization.Type, &organization.Description, &organization.CreatedAt, &organization.UpdatedAt); err != nil {
			return nil, err
		}
		organizations = append(organizations, organization)
	}

	return organizations, rows.Err()
}
