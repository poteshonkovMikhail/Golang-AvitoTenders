package models

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterUser struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func CreateUser(db *pgxpool.Pool, username, firstname, lastname, hashedPassword string) (*User, error) {

	// Начинаем транзакцию
	tx, err := db.Begin(context.Background())
	if err != nil {
		log.Fatalf("Ошибка при начале транзакции при INSERT'е нового пользователя: %v\n", err)
		return nil, err
	}

	var user User
	ctx := context.Background()

	err = tx.QueryRow(ctx,
		"INSERT INTO employee (username, first_name, last_name) VALUES ($1, $2, $3) RETURNING id, username, first_name, last_name, created_at",
		username, firstname, lastname,
	).Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.CreatedAt)
	if err != nil {
		log.Printf("Не удалось вставить нового сотрудника: %v", err)
		return nil, err
	}

	if _, err := tx.Exec(context.Background(), "INSERT INTO employee_passwords (employee_id, password_hash) VALUES ($1, $2)", user.ID, hashedPassword); err != nil {
		tx.Rollback(context.Background())
		log.Fatalf("Ошибка при записи хэша пароля пользователя: %v\n", err)
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		log.Printf("Не удалось зафиксировать транзакцию по созданию пользователя: %v", err)
		return nil, err
	}

	return &user, nil
}
