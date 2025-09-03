package handlers

import (
	"main/pkg/models"
	"main/pkg/queue"
	"sync"

	"github.com/gin-gonic/gin"
)

func GetSolanaBalance(c *gin.Context) {
	var request models.WalletsRequest
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(400, models.GenericResponse[any]{
			Object:  nil,
			Error:   "Invalid request body",
			Success: false,
		})
		return
	}

	var result []models.WalletBalance
	wg := sync.WaitGroup{}
	wg.Add(len(request.Wallets))
	for _, wallet := range request.Wallets {
		go func(wallet string) {
			defer wg.Done()
			waitChan := queue.AddWalletToQueue(wallet)
			res := <-waitChan
			bal := models.WalletBalance{
				Wallet: wallet,
			}
			if res.Error != nil {
				bal.Balance = res.Error.Error()
			} else {
				bal.Balance = res.Result
			}
			if res.Cache {
				bal.Cache = "hit"
			} else {
				bal.Cache = "miss"
			}
			result = append(result, bal)
		}(wallet)
	}
	wg.Wait()

	c.JSON(200, models.GenericResponse[[]models.WalletBalance]{
		Object:  result,
		Error:   "",
		Success: true,
	})
}
