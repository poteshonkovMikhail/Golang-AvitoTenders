package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"avito/tender/config"
	"avito/tender/helpers/verif_responsible"
	"avito/tender/models"

	"github.com/gin-gonic/gin"
)

func GetUserBids(c *gin.Context) {
	username := c.MustGet("username").(string)

	bids, err := models.GetUserBids(config.DB, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bids not found"})
		return
	}

	c.JSON(http.StatusOK, bids)
}

func GetBids(c *gin.Context) {
	bids, err := models.GetBids(config.DB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get bids"})
		return
	}

	c.JSON(http.StatusOK, bids)
}

func GetBidsForTender(c *gin.Context) {
	tenderId := c.Param("tenderId")
	bids, err := models.FetchBidsForTender(config.DB, tenderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch bids"})
		return
	}

	c.JSON(http.StatusOK, bids)
}

func GetBidReviews(c *gin.Context, tenderId, organizationId string) {
	responsibleID := c.MustGet("user_id").(string)
	var exists bool
	err1 := config.DB.QueryRow(context.Background(), "SELECT EXISTS (SELECT 1 FROM organization_responsible WHERE organization_id=$1 AND user_id=$2)", organizationId, responsibleID).Scan(&exists)
	if err1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign organisation responsible"})
		return
	}
	if exists {
		reviews, err := models.FetchBidReviews(config.DB, tenderId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to fetch reviews"})
			return
		}

		c.JSON(http.StatusOK, reviews)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "You haven't access to this bid reviews cause you're not responsible for this organization"})
	}
}

func CreateBid(c *gin.Context) {
	var bid models.CreateBidStruct

	username := c.MustGet("username").(string)
	fmt.Println(username)

	if err := c.BindJSON(&bid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	newBid, err := models.CreateBid(config.DB, bid.TenderID, bid.Amount, bid.Description, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bid"})
		return
	}

	c.JSON(http.StatusOK, newBid)
}

func EditBid(c *gin.Context) {
	id := c.Param("bidId")
	userId := c.MustGet("user_id").(string)

	var bidInput struct {
		Amount      int    `json:"amount"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := c.ShouldBindJSON(&bidInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	exists, err := verif_responsible.VerifBidEditor(config.DB, id, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to verify bid editor"})
		return
	}

	if exists {
		bid, err := models.UpdateBid(config.DB, id, bidInput.Amount, bidInput.Description, bidInput.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bid"})
			return
		}

		c.JSON(http.StatusOK, bid)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"messages": "You haven't an access to edit this bid"})
	}
}

func RollbackBid(c *gin.Context, bidIDParam, versionParam string) {
	newCreatorUsername := c.MustGet("username").(string)
	userId := c.MustGet("user_id").(string)
	version, err := strconv.Atoi(versionParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
		return
	}

	exists, err := verif_responsible.VerifBidEditor(config.DB, bidIDParam, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to verify bid editor"})
		return
	}

	if exists {

		err = models.RollbackBidToVersion(config.DB, bidIDParam, version, newCreatorUsername)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rollback bid"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":       bidIDParam,
			"version":  version,
			"rollback": true,
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"messages": "You haven't an access to rollback this bid"})
	}
}

func GetBidsReviewOfAuthor(c *gin.Context, tenderId string, authorUsername string, organizationId string) {
	userId := c.MustGet("user_id").(string)
	exists, err := verif_responsible.VerifResponsible(config.DB, organizationId, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to verify bid editor"})
		return
	}

	if exists {

		var reviews []models.Review

		query := "SELECT id, tender_id, bid_id, author_username, comment, created_at FROM reviews WHERE bid_id IN (SELECT id FROM bids WHERE tender_id = $1)"

		rows, err := config.DB.Query(context.Background(), query, tenderId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var review models.Review
			if err := rows.Scan(&review.ID, &review.TenderID, &review.BidID, &review.AuthorUsername, &review.Comment, &review.CreatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan review"})
				return
			}
			reviews = append(reviews, review)
		}

		c.JSON(http.StatusOK, reviews)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "You cannot make this action because you haven't relevant roots for that"})
	}
}

func CreateReviewHandler(c *gin.Context, tenderId, bidId string) {
	var input models.CreateReviewInput
	authorUsername := c.MustGet("username").(string)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	newReview, err := models.CreateReview(config.DB, input, authorUsername, tenderId, bidId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create review"})
		return
	}

	c.JSON(http.StatusOK, newReview)
}

func ApproveBid(c *gin.Context) {
	bidID := c.Param("bidId")
	userId := c.MustGet("user_id").(string)
	var tenderId string

	exists, err := verif_responsible.VerifBidEditor(config.DB, bidID, userId)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Failed to verify"})
		return
	}

	if exists {
		err := config.DB.QueryRow(context.Background(), "UPDATE bids SET status=$1, version=version WHERE bid_id=$2 RETURNING tender_id", "APPROVED", bidID).Scan(&tenderId)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Failed to approve bid"})
			return
		}

		_, err = config.DB.Exec(context.Background(), "UPDATE tenders SET status=$1, version=version WHERE id=$2", "CLOSED", tenderId)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Failed to close tender"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "The bid approved and the tenders closed succsessfully"})
	} else {
		c.JSON(http.StatusForbidden, gin.H{"message": "You haven't roots to approve this bid"})
	}
}
