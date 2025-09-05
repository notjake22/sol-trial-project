package config

import "os"

func Load() {
	Config = &Structure{
		RpcUri:      os.Getenv("RPC_URI"),
		Port:        os.Getenv("PORT"),
		MongoDbName: os.Getenv("MONGO_DB_NAME"),
		MongoUri:    os.Getenv("MONGO_URI"),
		RedisUri:    os.Getenv("REDIS_URL"),
	}
}
