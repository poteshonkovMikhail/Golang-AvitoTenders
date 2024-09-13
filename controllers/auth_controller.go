package controllers

// ready
import (
	"context"
	"database/sql"
	"net/http"

	"avito/tender/config"
	"avito/tender/helpers/jwt_actions"
	"avito/tender/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var user models.RegisterUser
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	if _, err := models.CreateUser(config.DB, user.Username, user.FirstName, user.LastName, user.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful"})
}

func Login(c *gin.Context) {
	var user models.LoginUser
	var dbUser models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	////////////////
	row1 := config.DB.QueryRow(context.Background(), "SELECT id, username, first_name, last_name FROM employee WHERE username = $1", user.Username)
	if err := row1.Scan(&dbUser.ID, &dbUser.Username, &dbUser.FirstName, &dbUser.LastName); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	row2 := config.DB.QueryRow(context.Background(), "SELECT password_hash FROM employee_passwords WHERE employee_id = $1", dbUser.ID)
	if err := row2.Scan(&dbUser.Password); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Cannot retrive password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}
	/////////////////////////

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(user.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	tokenString, _ := jwt_actions.GenerateJWT(config.DB, dbUser.Username, dbUser.ID)
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "username": dbUser.Username})
}
