package routers

import (
	"main/internal/server/rest/handlers"

	"github.com/gin-gonic/gin"
)

func setupSolanaRoutes(app *gin.Engine, apiAuth *gin.RouterGroup) {
	solana := apiAuth.Group("")
	{
		solana.POST("/get-balance", handlers.GetSolanaBalance)
	}
}
