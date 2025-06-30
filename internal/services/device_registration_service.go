package services

import (
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceRegistrationService struct {
	db *gorm.DB
}

func NewDeviceRegistrationService(db *gorm.DB) *DeviceRegistrationService {
	return &DeviceRegistrationService{
		db: db,
	}
}

// RegisterDevice registers a device to a target user
func (s *DeviceRegistrationService) RegisterDevice(
	registrarUserID uuid.UUID,
	targetUserID uuid.UUID,
	deviceIdentifier string,
	deviceType string,
	notes string,
	ipAddress string,
	userAgent string,
) (*database.DeviceRegistration, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Find target user
	var targetUser database.User
	if err := tx.Where("id = ?", targetUserID).First(&targetUser).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("target user not found: %w", err)
	}

	if !targetUser.Active {
		tx.Rollback()
		return nil, fmt.Errorf("target user is not active")
	}

	// 2. Find or create device
	var device database.Device
	err := tx.Where("type = ? AND identifier = ?", deviceType, deviceIdentifier).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new device
			device = database.Device{
				ID:         uuid.New(),
				Type:       deviceType,
				Identifier: deviceIdentifier,
				Active:     true,
				VerifiedAt: time.Now(),
			}
			if err := tx.Create(&device).Error; err != nil {
				tx.Rollback()
				return nil, fmt.Errorf("failed to create device: %w", err)
			}
		} else {
			tx.Rollback()
			return nil, fmt.Errorf("failed to find device: %w", err)
		}
	} else {
		// Check if device is already registered to another user
		if device.UserID != uuid.Nil && device.UserID != targetUserID {
			tx.Rollback()
			return nil, fmt.Errorf("device is already registered to another user")
		}
	}

	// 3. Update device ownership
	device.UserID = targetUserID
	device.Active = true
	device.VerifiedAt = time.Now()
	if err := tx.Save(&device).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// 4. Create registration record
	registration := database.DeviceRegistration{
		ID:              uuid.New(),
		RegistrarUserID: registrarUserID,
		DeviceID:        device.ID,
		TargetUserID:    &targetUserID,
		ActionType:      "register",
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Notes:           notes,
	}

	if err := tx.Create(&registration).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create registration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &registration, nil
}

// DeregisterDevice deregisters a device from its current user
func (s *DeviceRegistrationService) DeregisterDevice(
	registrarUserID uuid.UUID,
	deviceID uuid.UUID,
	reason string,
	notes string,
	ipAddress string,
	userAgent string,
) (*database.DeviceRegistration, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Find device
	var device database.Device
	if err := tx.Preload("User").Where("id = ?", deviceID).First(&device).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// 2. Check if device is currently registered
	if device.UserID == uuid.Nil {
		tx.Rollback()
		return nil, fmt.Errorf("device is not currently registered to any user")
	}

	// 3. Deregister device
	device.UserID = uuid.Nil
	device.Active = false
	if err := tx.Save(&device).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to deregister device: %w", err)
	}

	// 4. Create deregistration record
	registration := database.DeviceRegistration{
		ID:              uuid.New(),
		RegistrarUserID: registrarUserID,
		DeviceID:        device.ID,
		TargetUserID:    nil, // NULL for deregistration
		ActionType:      "deregister",
		Reason:          reason,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Notes:           notes,
	}

	if err := tx.Create(&registration).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create deregistration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &registration, nil
}

// TransferDevice transfers a device from one user to another
func (s *DeviceRegistrationService) TransferDevice(
	registrarUserID uuid.UUID,
	deviceID uuid.UUID,
	targetUserID uuid.UUID,
	notes string,
	ipAddress string,
	userAgent string,
) (*database.DeviceRegistration, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Find target user
	var targetUser database.User
	if err := tx.Where("id = ?", targetUserID).First(&targetUser).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("target user not found: %w", err)
	}

	if !targetUser.Active {
		tx.Rollback()
		return nil, fmt.Errorf("target user is not active")
	}

	// 2. Find device
	var device database.Device
	if err := tx.Preload("User").Where("id = ?", deviceID).First(&device).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// 3. Check if device is currently registered
	if device.UserID == uuid.Nil {
		tx.Rollback()
		return nil, fmt.Errorf("device is not currently registered to any user")
	}

	// 4. Check if device is already registered to target user
	if device.UserID == targetUserID {
		tx.Rollback()
		return nil, fmt.Errorf("device is already registered to the target user")
	}

	// 5. Transfer device
	previousUserID := device.UserID
	device.UserID = targetUserID
	device.Active = true
	device.VerifiedAt = time.Now()
	if err := tx.Save(&device).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to transfer device: %w", err)
	}

	// 6. Create transfer record (deregister from previous user)
	deregRecord := database.DeviceRegistration{
		ID:              uuid.New(),
		RegistrarUserID: registrarUserID,
		DeviceID:        device.ID,
		TargetUserID:    nil,
		ActionType:      "deregister",
		Reason:          "transfer",
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Notes:           fmt.Sprintf("Transferred from user %s to user %s. %s", previousUserID, targetUserID, notes),
	}

	if err := tx.Create(&deregRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create deregistration record: %w", err)
	}

	// 7. Create registration record (register to new user)
	regRecord := database.DeviceRegistration{
		ID:              uuid.New(),
		RegistrarUserID: registrarUserID,
		DeviceID:        device.ID,
		TargetUserID:    &targetUserID,
		ActionType:      "register",
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		Notes:           fmt.Sprintf("Transferred from user %s. %s", previousUserID, notes),
	}

	if err := tx.Create(&regRecord).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create registration record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &regRecord, nil
}

// GetDeviceHistory returns the registration history for a device
func (s *DeviceRegistrationService) GetDeviceHistory(deviceID uuid.UUID) ([]database.DeviceRegistration, error) {
	var registrations []database.DeviceRegistration
	err := s.db.Preload("RegistrarUser").
		Preload("TargetUser").
		Preload("Device").
		Where("device_id = ?", deviceID).
		Order("created_at DESC").
		Find(&registrations).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get device history: %w", err)
	}

	return registrations, nil
} 