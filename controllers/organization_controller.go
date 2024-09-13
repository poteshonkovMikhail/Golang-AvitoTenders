package controllers

import (
	"context"
	"log"
	"net/http"

	"avito/tender/config"
	//"avito/tender/helpers/jwt_actions"
	"avito/tender/models"

	"github.com/gin-gonic/gin"
)

func CreateOrganization(c *gin.Context) {
	var org models.CreateOrganizationStruct
	creatorID := c.MustGet("user_id").(string)

	if err := c.ShouldBindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := models.CreateOrganization(config.DB, org, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Organization created successfully"})
}

func GetOrganizations(c *gin.Context) {
	organizations, err := models.GetOrganizations(config.DB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get organizations"})
		return
	}

	c.JSON(http.StatusOK, organizations)
}

func AssignEmployeeRole(c *gin.Context, orgID string) {
	var exists bool
	responsibleID := c.MustGet("user_id").(string)
	var organisationResponsibleRequest struct {
		UserID string `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&organisationResponsibleRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err1 := config.DB.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM organization_responsible WHERE organization_id=$1 AND user_id=$2)", orgID, responsibleID).Scan(&exists)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign organisation responsible"})
		return
	}
	if exists {
		_, err2 := config.DB.Exec(context.Background(), "INSERT INTO organization_responsible (organization_id, user_id) VALUES ($1, $2)", orgID, organisationResponsibleRequest.UserID)
		if err2 != nil {
			log.Printf("%v", err2)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign organisation responsible"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Organisation responsible assigned successfully"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "You cannot assign responsibility for this organization because you are not responsible for it."})
	}
}
