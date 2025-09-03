package middleware

import (
	"errors"
	"log"
	"main/internal/server/service"

	"github.com/gin-gonic/gin"
)

func Authenticate(c *gin.Context) {
	// add in api key authentication here later

	// get x-api-key from header and validate from mongo license service
	apiKey := c.GetHeader("x-api-key")
	if apiKey == "" {
		if err := c.AbortWithError(401, gin.Error{
			Err:  errors.New("missing api key"),
			Type: gin.ErrorTypePublic,
			Meta: "Missing API key",
		}); err != nil {
			log.Println("Error aborting request: " + err.Error())
		}
		return
	}
	
	_, err := service.ValidateLicense(apiKey)
	if err != nil {
		if err = c.AbortWithError(401, gin.Error{
			Err:  err,
			Type: gin.ErrorTypePublic,
			Meta: "Invalid API key",
		}); err != nil {
			log.Println("Error aborting request: " + err.Error())
		}
	}

	go func() {
		if err = service.IncrementLicenseUsage(apiKey); err != nil {
			log.Println("Error incrementing license usage: " + err.Error())
		}
	}()

	count, err := service.GetIpRequestCount(c.ClientIP())
	if err != nil {
		if err = c.AbortWithError(429, err); err != nil {
			log.Println("Error aborting request: " + err.Error())
		}
		return
	}

	if count >= 10 {
		if err = c.AbortWithError(429, gin.Error{
			Err:  err,
			Type: gin.ErrorTypePublic,
			Meta: "Too many requests from this IP, please try again later.",
		}); err != nil {
			log.Println("Error aborting request: " + err.Error())
		}
		return
	}

	if err = service.IncrementIpRequestCount(c.ClientIP()); err != nil {
		if err = c.AbortWithError(500, err); err != nil {
			log.Println("Error aborting request: " + err.Error())
		}
		return
	}

	c.Next()
}
