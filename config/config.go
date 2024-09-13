package config

import (
	"context"
	"fmt"
	"log"

	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

//func InitDB() {
//	ctx := context.Background()
//	connStr := os.Getenv("POSTGRES_CONN")
//	config, err := pgxpool.ParseConfig(connStr)
//	if err != nil {
//		log.Fatalf("Unable to parse DATABASE_URL %v", err)
//	}
//
//	pool, err := pgxpool.NewWithConfig(ctx, config)
//	if err != nil {
//		log.Fatalf("Unable to create connection pool %v", err)
//	}
//
//	DB = pool
//	fmt.Println("Database connection established")
//}

func InitDB() {
	var err error

	// Используем pgxpool для создания пула соединений
	DB, err = pgxpool.New(context.Background(), os.Getenv("POSTGRES_CONN"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Проверка состояния подключения
	if err = DB.Ping(context.Background()); err != nil {
		log.Fatalf("Database is unreachable: %v\n", err)
	} else {
		fmt.Println("Database connected successfully")
	}
}
