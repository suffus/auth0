package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceService struct {
	db *gorm.DB
}

func NewDeviceService(db *gorm.DB) *DeviceService {
	return &DeviceService{db: db}
}

// CreateDevice creates a new device
func (s *DeviceService) CreateDevice(userID uuid.UUID, deviceType, identifier, secret string, active bool) (*database.Device, error) {
	validTypes := []string{"yubikey", "totp", "sms", "email"}
	validType := false
	for _, t := range validTypes {
		if deviceType == t {
			validType = true
			break
		}
	}
	if !validType {
		return nil, fmt.Errorf("device type must be one of: %v", validTypes)
	}

	// Check if user exists
	var user database.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Generate secret for TOTP if not provided
	if secret == "" && deviceType == "totp" {
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			return nil, fmt.Errorf("failed to generate secret: %w", err)
		}
		secret = hex.EncodeToString(secretBytes)
	}

	device := database.Device{
		ID:         uuid.New(),
		UserID:     userID,
		Type:       deviceType,
		Identifier: identifier,
		Secret:     secret,
		Active:     active,
		VerifiedAt: time.Now(),
	}

	if err := s.db.Create(&device).Error; err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return &device, nil
}

// GetDeviceByID retrieves a device by ID
func (s *DeviceService) GetDeviceByID(deviceID uuid.UUID) (*database.Device, error) {
	var device database.Device
	if err := s.db.Preload("User").Where("id = ?", deviceID).First(&device).Error; err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	return &device, nil
}

// GetDeviceByIdentifier retrieves a device by identifier and type
func (s *DeviceService) GetDeviceByIdentifier(deviceType, identifier string) (*database.Device, error) {
	var device database.Device
	if err := s.db.Preload("User").Where("type = ? AND identifier = ?", deviceType, identifier).First(&device).Error; err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}
	return &device, nil
}

// ListDevices retrieves all devices or devices for a specific user
func (s *DeviceService) ListDevices(userID *uuid.UUID) ([]database.Device, error) {
	var devices []database.Device
	var err error

	if userID != nil {
		err = s.db.Preload("User").Where("user_id = ?", userID).Find(&devices).Error
	} else {
		err = s.db.Preload("User").Find(&devices).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}

	return devices, nil
}

// ListActiveDevices retrieves only active devices, optionally filtered by userID
func (s *DeviceService) ListActiveDevices(userID *uuid.UUID) ([]database.Device, error) {
	var devices []database.Device
	var err error

	if userID != nil {
		err = s.db.Preload("User").Where("user_id = ? AND active = ?", userID, true).Find(&devices).Error
	} else {
		err = s.db.Preload("User").Where("active = ?", true).Find(&devices).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch active devices: %w", err)
	}

	return devices, nil
}

// UpdateDevice updates a device
func (s *DeviceService) UpdateDevice(deviceID uuid.UUID, updates map[string]interface{}) (*database.Device, error) {
	var device database.Device
	if err := s.db.Where("id = ?", deviceID).First(&device).Error; err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Validate device type if it's being updated
	if deviceType, ok := updates["type"].(string); ok {
		validTypes := []string{"yubikey", "totp", "sms", "email"}
		validType := false
		for _, t := range validTypes {
			if deviceType == t {
				validType = true
				break
			}
		}
		if !validType {
			return nil, fmt.Errorf("device type must be one of: %v", validTypes)
		}
	}

	if err := s.db.Model(&device).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// Reload device with user
	if err := s.db.Preload("User").Where("id = ?", deviceID).First(&device).Error; err != nil {
		return nil, fmt.Errorf("failed to reload device: %w", err)
	}

	return &device, nil
}

// DeleteDevice deletes a device
func (s *DeviceService) DeleteDevice(deviceID uuid.UUID) error {
	var device database.Device
	if err := s.db.Preload("User").Where("id = ?", deviceID).First(&device).Error; err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	if err := s.db.Delete(&device).Error; err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	return nil
}

// UpdateDeviceLastUsed updates the last used timestamp for a device
func (s *DeviceService) UpdateDeviceLastUsed(deviceID uuid.UUID) error {
	return s.db.Model(&database.Device{}).Where("id = ?", deviceID).Update("last_used_at", time.Now()).Error
} 