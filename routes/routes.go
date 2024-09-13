package routes

import (
	"avito/tender/controllers"
	"avito/tender/middleware"

	"github.com/gin-gonic/gin"
)

func InitRoutes(router *gin.Engine) {

	//Public routes
	auth := router.Group("/auth")
	{
		//Authentification routes
		auth.POST("/register", controllers.Register) //✅
		auth.POST("/login", controllers.Login)       //✅
	}
	//Protected routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/ping", controllers.Ping) //✅

		// Organization Routes
		api.GET("/organizations", controllers.GetOrganizations)    //✅
		api.POST("/organizations", controllers.CreateOrganization) //✅
		api.POST("/organizations/assignEmployeeRole", func(c *gin.Context) {
			organisationId := c.Query("organizationId")
			controllers.AssignEmployeeRole(c, organisationId)
		}) //✅

		// Tender Routes
		api.GET("/tenders", controllers.GetTenders)                  //✅
		api.POST("/tenders/new", controllers.CreateTender)           //✅  //Ready
		api.GET("/tenders/my", controllers.GetUserTenders)           //✅
		api.PATCH("/tenders/:tenderId/edit", controllers.EditTender) //✅	//Ready
		api.PUT("/tenders/:tenderId/rollback", func(c *gin.Context) {
			bidId := c.Param("tenderId")
			version := c.Query("version")
			controllers.RollbackTenderVersion(c, bidId, version)
		}) //✅			//Ready

		// Bid/Review Routes
		api.GET("/bids/my", controllers.GetUserBids)                  //✅
		api.POST("/bids/new", controllers.CreateBid)                  //✅
		api.GET("/bids/:tenderId/list", controllers.GetBidsForTender) //✅
		api.PATCH("/bids/:bidId/edit", controllers.EditBid)           //✅      //Ready
		api.PATCH("/bids/:bidId/approve", controllers.ApproveBid)     //✅      //Ready
		api.PUT("/bids/:bidId/rollback", func(c *gin.Context) {
			bidId := c.Param("bidId")
			version := c.Query("version")
			controllers.RollbackBid(c, bidId, version)
		}) //✅ //Ready
		api.POST("/bids/:bidId/reviews/new", func(c *gin.Context) {
			tenderId := c.Query("tenderId")
			bidId := c.Param("bidId")
			controllers.CreateReviewHandler(c, tenderId, bidId)
		}) //✅      //Ready
		api.GET("/bids/:tenderId/get/reviews", func(c *gin.Context) {
			tenderId := c.Param("tenderId")
			organizationId := c.Query("organizationId")
			controllers.GetBidReviews(c, tenderId, organizationId)
		}) //✅
		api.GET("/bids/:tenderId/reviews", func(c *gin.Context) {
			tenderId := c.Param("tenderId")
			authorUsername := c.Query("authorUsername")
			organizationId := c.Query("organizationId")
			controllers.GetBidsReviewOfAuthor(c, tenderId, authorUsername, organizationId) ////////////////////////////////////////
		}) //✅ Ready
	}
}
