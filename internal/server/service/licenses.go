package service

import (
	"fmt"
	"main/internal/database/mongo"
	mongo2 "main/internal/server/repo/mongo"
	"main/pkg/models"
)

var (
	licenseKeys       *mongo2.LicenseKey
	licenseKeyService mongo2.LicenseKeyImpl
)

func init() {
	// Only initialize if MongoDB is available
	if mongo.Database != nil {
		licenseKeys = &mongo2.LicenseKey{
			Collection: mongo.Database.Collection("license_keys"),
		}
		licenseKeyService = mongo2.LicenseKeyImpl(licenseKeys)
	}
}

func CreateLicense(request models.CreateLicenseRequest) (*models.License, error) {
	if licenseKeyService == nil {
		return nil, fmt.Errorf("license service not initialized")
	}
	result, err := licenseKeyService.CreateLicense(request.Name, &request.Expiry, &request.UsageLimit)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func ValidateLicense(key string) (*models.License, error) {
	if licenseKeyService == nil {
		return nil, fmt.Errorf("license service not initialized")
	}
	result, err := licenseKeyService.ValidateLicense(key)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func IncrementLicenseUsage(key string) error {
	if licenseKeyService == nil {
		return fmt.Errorf("license service not initialized")
	}
	err := licenseKeyService.IncrementUsage(key)
	if err != nil {
		return err
	}

	return nil
}
