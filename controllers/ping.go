package controllers

// ✅
import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping проверка для проверки состояния сервера
func Ping(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
