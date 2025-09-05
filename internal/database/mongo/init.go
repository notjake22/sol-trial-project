package mongo

import (
	"context"
	"log"
	"main/pkg/config"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var Database *mongo.Database

func Init() {
	err := connect(config.Config.MongoDbName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
}

func connect(dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.Config.MongoUri))
	if err != nil {
		return err
	}

	if err = client.Ping(ctx, nil); err != nil {
		return err
	}

	Client = client
	Database = client.Database(dbName)
	log.Println("Connected to MongoDB")
	return nil
}

func Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return Client.Disconnect(ctx)
}
