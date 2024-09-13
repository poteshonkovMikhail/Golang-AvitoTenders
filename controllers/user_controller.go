package controllers

import (
	"context"
	"net/http"
	"time"

	"avito/tender/config"
	"avito/tender/models"

	"github.com/gin-gonic/gin"
)

func CreateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		newUser, err := models.CreateUser(config.DB, user.Username, user.FirstName, user.LastName, user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusOK, newUser)
	}(ctx)
}
