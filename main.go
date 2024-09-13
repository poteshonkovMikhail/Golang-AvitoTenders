package main

import (
	"avito/tender/config"
	"avito/tender/routes"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализация базы данных
	config.InitDB()

	// Создаем роутер Gin
	router := gin.Default()

	// Добавляем маршруты
	routes.InitRoutes(router)

	// Адрес сервера из переменной окружения
	serverAddress := os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		serverAddress = "0.0.0.0:8080"
	}

	fmt.Printf("Server running at %s\n", serverAddress)
	router.Run(serverAddress)
}
