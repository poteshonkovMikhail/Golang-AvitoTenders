package controllers

import (
	"context"
	"net/http"
	"strconv"

	"avito/tender/config"
	"avito/tender/helpers/verif_responsible"
	"avito/tender/models"

	"github.com/gin-gonic/gin"
)

const NoOrganizationUUID = "e1d9b4e4-3e10-4d4c-8f90-2aa816d349f5"

func GetUserTenders(c *gin.Context) {
	username := c.MustGet("username").(string)

	tender, err := models.GetUserTenders(config.DB, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tender not found"})
		return
	}

	c.JSON(http.StatusOK, tender)
}

func GetTenders(c *gin.Context) {
	tenders, err := models.GetTenders(config.DB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get tenders"})
		return
	}

	c.JSON(http.StatusOK, tenders)
}

func CreateTender(c *gin.Context) {
	var tender models.CreateTenderStruct

	username := c.MustGet("username").(string)
	userId := c.MustGet("user_id").(string)

	if err := c.BindJSON(&tender); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if tender.OrganizationID == "" {
		tender.OrganizationID = NoOrganizationUUID
		err := config.DB.QueryRow(context.Background(), "INSERT INTO organization_responsible VALUES organization_id=$1, user_id=$2", tender.OrganizationID, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tender"})
			return
		}
	}

	newTender, err := models.CreateTender(config.DB, tender.Name, tender.Description, tender.ServiceType, tender.OrganizationID, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tender"})
		return
	}

	c.JSON(http.StatusOK, newTender)
}

func EditTender(c *gin.Context) {
	id := c.Param("tenderId")
	userId := c.MustGet("user_id").(string)

	var organizationId string
	var tenderInput struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		ServiceType string `json:"service_type"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&tenderInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := config.DB.QueryRow(context.Background(), "SELECT organization_id FROM tenders WHERE tender_id=$1", id).Scan(&organizationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify organisation responsible"})
		return
	}

	exists, err := verif_responsible.VerifResponsible(config.DB, userId, organizationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify organisation responsible"})
		return
	}

	if exists {
		tender, err := models.UpdateTender(config.DB, id, tenderInput.Name, tenderInput.Description, tenderInput.ServiceType, tenderInput.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tender"})
			return
		}

		c.JSON(http.StatusOK, tender)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "You haven't roots to edit this tender"})
	}
}

func RollbackTenderVersion(c *gin.Context, tenderIDParam, versionParam string) {
	var organizationId string
	userId := c.MustGet("user_id").(string)
	err := config.DB.QueryRow(context.Background(), "SELECT organization_id FROM tenders WHERE tender_id=$1", tenderIDParam).Scan(&organizationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify organisation responsible"})
		return
	}

	exists, err := verif_responsible.VerifResponsible(config.DB, userId, organizationId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify organisation responsible"})
		return
	}

	if exists {

		version, err := strconv.Atoi(versionParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}

		err = models.RollbackTenderToVersion(config.DB, tenderIDParam, version)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tender"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       tenderIDParam,
			"version":  version,
			"rollback": true,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "You haven't roots to edit this tender"})
	}
}
