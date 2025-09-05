package main

import (
	"main/internal/database/mongo"
	"main/internal/database/redis"
	"main/internal/server"
	"main/pkg/config"
	"os"
)

func main() {
	config.Load()
	redis.InitRedis()
	mongo.Init()

	err := server.Start(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}
}
