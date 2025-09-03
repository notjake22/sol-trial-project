package routers

import (
	"main/internal/server/rest/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(app *gin.Engine) {
	apiAuth := app.Group("/api", middleware.Authenticate)

	app.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "solana-api",
			"version": "1.0.0",
		})
	})
	setupSolanaRoutes(app, apiAuth)
}
