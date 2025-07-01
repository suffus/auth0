package services

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/YubiApp/internal/database"
)

type UserStatusService struct {
	db *gorm.DB
}

func NewUserStatusService(db *gorm.DB) *UserStatusService {
	return &UserStatusService{db: db}
}

// CreateUserStatus creates a new user status
func (s *UserStatusService) CreateUserStatus(name, description, statusType string, active bool) (*database.UserStatus, error) {
	// Validate status type
	validTypes := []string{"working", "break", "leave", "travel", "other"}
	isValidType := false
	for _, validType := range validTypes {
		if statusType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, fmt.Errorf("invalid status type: %s. Valid types are: %s", statusType, strings.Join(validTypes, ", "))
	}

	// Check if name already exists
	var existing database.UserStatus
	if err := s.db.Where("name = ?", name).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("user status with name '%s' already exists", name)
	}

	userStatus := &database.UserStatus{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Type:        statusType,
		Active:      active,
	}

	if err := s.db.Create(userStatus).Error; err != nil {
		return nil, fmt.Errorf("failed to create user status: %w", err)
	}

	return userStatus, nil
}

// GetUserStatusByID retrieves a user status by ID
func (s *UserStatusService) GetUserStatusByID(id uuid.UUID) (*database.UserStatus, error) {
	var userStatus database.UserStatus
	if err := s.db.Where("id = ?", id).First(&userStatus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user status not found")
		}
		return nil, fmt.Errorf("failed to fetch user status: %w", err)
	}
	return &userStatus, nil
}

// GetUserStatusByName retrieves a user status by name
func (s *UserStatusService) GetUserStatusByName(name string) (*database.UserStatus, error) {
	var userStatus database.UserStatus
	if err := s.db.Where("name = ?", name).First(&userStatus).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user status not found")
		}
		return nil, fmt.Errorf("failed to fetch user status: %w", err)
	}
	return &userStatus, nil
}

// ListUserStatuses retrieves all user statuses
func (s *UserStatusService) ListUserStatuses() ([]database.UserStatus, error) {
	var userStatuses []database.UserStatus
	if err := s.db.Find(&userStatuses).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user statuses: %w", err)
	}
	return userStatuses, nil
}

// ListActiveUserStatuses retrieves only active user statuses
func (s *UserStatusService) ListActiveUserStatuses() ([]database.UserStatus, error) {
	var userStatuses []database.UserStatus
	if err := s.db.Where("active = ?", true).Find(&userStatuses).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active user statuses: %w", err)
	}
	return userStatuses, nil
}

// UpdateUserStatus updates a user status
func (s *UserStatusService) UpdateUserStatus(id uuid.UUID, name, description, statusType *string, active *bool) (*database.UserStatus, error) {
	userStatus, err := s.GetUserStatusByID(id)
	if err != nil {
		return nil, err
	}

	// Validate status type if provided
	if statusType != nil {
		validTypes := []string{"working", "break", "leave", "travel", "other"}
		isValidType := false
		for _, validType := range validTypes {
			if *statusType == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			return nil, fmt.Errorf("invalid status type: %s. Valid types are: %s", *statusType, strings.Join(validTypes, ", "))
		}
		userStatus.Type = *statusType
	}

	// Check if name already exists (if changing name)
	if name != nil && *name != userStatus.Name {
		var existing database.UserStatus
		if err := s.db.Where("name = ? AND id != ?", *name, id).First(&existing).Error; err == nil {
			return nil, fmt.Errorf("user status with name '%s' already exists", *name)
		}
		userStatus.Name = *name
	}

	if description != nil {
		userStatus.Description = *description
	}

	if active != nil {
		userStatus.Active = *active
	}

	if err := s.db.Save(userStatus).Error; err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	return userStatus, nil
}

// DeleteUserStatus performs a soft delete by setting active to false
func (s *UserStatusService) DeleteUserStatus(id uuid.UUID) error {
	userStatus, err := s.GetUserStatusByID(id)
	if err != nil {
		return err
	}

	userStatus.Active = false
	if err := s.db.Save(userStatus).Error; err != nil {
		return fmt.Errorf("failed to delete user status: %w", err)
	}

	return nil
} 