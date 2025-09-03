package server

import (
	"log"
	"main/internal/server/rest/routers"

	"github.com/gin-gonic/gin"
)

func Start(port string) error {
	app := gin.Default()
	gin.SetMode(gin.DebugMode)

	routers.SetupRoutes(app)

	log.Println("Starting server on port " + port)
	if err := app.Run(":" + port); err != nil {
		log.Println("Error starting server: " + err.Error())
		return err
	}

	return nil
}
