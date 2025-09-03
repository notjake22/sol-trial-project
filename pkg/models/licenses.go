package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LicenseKey struct {
	Collection *mongo.Collection
}

type License struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key        string             `bson:"key" json:"key"`
	Name       string             `bson:"name" json:"name"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt  *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	UsageCount int64              `bson:"usage_count" json:"usage_count"`
	UsageLimit *int64             `bson:"usage_limit,omitempty" json:"usage_limit,omitempty"`
}

type CreateLicenseRequest struct {
	Name       string    `json:"name"`
	Expiry     time.Time `json:"expiry,omitempty"`
	UsageLimit int64     `json:"usage_limit,omitempty"`
}

type LicenseKeyService interface {
	CreateLicense(name string, expiry *time.Time, limit *int64) (*License, error)
	ValidateLicense(key string) (*License, error)
	IncrementUsage(key string) error
}
