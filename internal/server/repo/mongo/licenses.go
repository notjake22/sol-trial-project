package mongo

import (
	"context"
	"main/pkg/models"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LicenseKey models.LicenseKey
type LicenseKeyImpl models.LicenseKeyService

func (l *LicenseKey) CreateLicense(name string, expiry *time.Time, limit *int64) (*models.License, error) {
	key, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	apiKey := &models.License{
		ID:         primitive.NewObjectID(),
		Key:        key.String(),
		Name:       name,
		CreatedAt:  time.Now(),
		ExpiresAt:  expiry,
		UsageCount: 0,
		UsageLimit: limit,
	}

	_, err = l.Collection.InsertOne(context.Background(), apiKey)
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func (l *LicenseKey) ValidateLicense(key string) (*models.License, error) {
	var apiKey models.License
	filter := bson.M{
		"key":       key,
		"is_active": true,
	}

	err := l.Collection.FindOne(context.Background(), filter).Decode(&apiKey)
	if err != nil {
		return nil, err
	}

	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, mongo.ErrNoDocuments
	}

	if apiKey.UsageLimit != nil && apiKey.UsageCount >= *apiKey.UsageLimit {
		return nil, mongo.ErrNoDocuments
	}

	return &apiKey, nil
}

func (l *LicenseKey) IncrementUsage(key string) error {
	filter := bson.M{"key": key}
	update := bson.M{"$inc": bson.M{"usage_count": 1}}

	_, err := l.Collection.UpdateOne(context.Background(), filter, update)
	return err
}
