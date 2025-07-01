package services

import (
	"errors"
	"fmt"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

type ActionService struct {
	db *gorm.DB
}

func NewActionService(db *gorm.DB) *ActionService {
	return &ActionService{db: db}
}

// GetActionByName retrieves an action by its name
func (s *ActionService) GetActionByName(name string) (*database.Action, error) {
	var action database.Action
	if err := s.db.Where("name = ?", name).First(&action).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("action '%s' not found", name)
		}
		return nil, err
	}
	return &action, nil
}

// GetActionByID retrieves an action by its ID
func (s *ActionService) GetActionByID(id uuid.UUID) (*database.Action, error) {
	var action database.Action
	if err := s.db.Where("id = ?", id).First(&action).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("action with ID '%s' not found", id)
		}
		return nil, err
	}
	return &action, nil
}

// ListActions retrieves all actions
func (s *ActionService) ListActions() ([]database.Action, error) {
	var actions []database.Action
	if err := s.db.Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
}

// CreateAction creates a new action
func (s *ActionService) CreateAction(name string, activityType string, requiredPermissions []string, details map[string]interface{}, active bool) (*database.Action, error) {
	// Validate activity type
	validTypes := []string{"user", "system", "automated", "other"}
	validType := false
	for _, t := range validTypes {
		if activityType == t {
			validType = true
			break
		}
	}
	if !validType {
		return nil, fmt.Errorf("invalid activity type. Must be one of: %v", validTypes)
	}

	// Convert []string to pgtype.JSONB for required permissions
	var permissionsJSONB pgtype.JSONB
	if err := permissionsJSONB.Set(requiredPermissions); err != nil {
		return nil, fmt.Errorf("failed to convert permissions to JSONB: %w", err)
	}

	// Convert details map to pgtype.JSONB
	var detailsJSONB pgtype.JSONB
	if details == nil {
		details = make(map[string]interface{})
	}
	if err := detailsJSONB.Set(details); err != nil {
		return nil, fmt.Errorf("failed to convert details to JSONB: %w", err)
	}

	action := &database.Action{
		Name:                name,
		ActivityType:        activityType,
		RequiredPermissions: permissionsJSONB,
		Details:             detailsJSONB,
		Active:              active,
	}

	if err := s.db.Create(action).Error; err != nil {
		return nil, err
	}

	return action, nil
}

// UpdateAction updates an existing action
func (s *ActionService) UpdateAction(id uuid.UUID, name string, activityType string, requiredPermissions []string, details map[string]interface{}, active *bool) (*database.Action, error) {
	action := &database.Action{}
	if err := s.db.Where("id = ?", id).First(action).Error; err != nil {
		return nil, err
	}

	action.Name = name
	
	// Validate activity type if provided
	if activityType != "" {
		validTypes := []string{"user", "system", "automated", "other"}
		validType := false
		for _, t := range validTypes {
			if activityType == t {
				validType = true
				break
			}
		}
		if !validType {
			return nil, fmt.Errorf("invalid activity type. Must be one of: %v", validTypes)
		}
		action.ActivityType = activityType
	}
	
	// Convert []string to pgtype.JSONB for required permissions
	var permissionsJSONB pgtype.JSONB
	if err := permissionsJSONB.Set(requiredPermissions); err != nil {
		return nil, fmt.Errorf("failed to convert permissions to JSONB: %w", err)
	}
	action.RequiredPermissions = permissionsJSONB

	// Convert details map to pgtype.JSONB
	if details != nil {
		var detailsJSONB pgtype.JSONB
		if err := detailsJSONB.Set(details); err != nil {
			return nil, fmt.Errorf("failed to convert details to JSONB: %w", err)
		}
		action.Details = detailsJSONB
	}

	// Update active status if provided
	if active != nil {
		action.Active = *active
	}

	if err := s.db.Save(action).Error; err != nil {
		return nil, err
	}

	return action, nil
}

// DeleteAction deletes an action
func (s *ActionService) DeleteAction(id uuid.UUID) error {
	return s.db.Delete(&database.Action{}, id).Error
}

// CheckUserPermissionsForAction checks if a user has the required permissions for an action
func (s *ActionService) CheckUserPermissionsForAction(userID uuid.UUID, actionName string) (bool, error) {
	// Get the action
	action, err := s.GetActionByName(actionName)
	if err != nil {
		return false, err
	}

	// Extract required permissions from JSONB
	var requiredPermissions []string
	if action.RequiredPermissions.Status == pgtype.Present {
		if err := action.RequiredPermissions.AssignTo(&requiredPermissions); err != nil {
			return false, fmt.Errorf("failed to read action permissions: %w", err)
		}
	}

	// If no permissions required, allow
	if len(requiredPermissions) == 0 {
		return true, nil
	}

	// Get user with roles and permissions
	var user database.User
	if err := s.db.Preload("Roles.Permissions.Resource").Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}

	// Check if user has all required permissions
	userPermissions := make(map[string]bool)
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			// Create permission key in format "resource:action"
			permissionKey := fmt.Sprintf("%s:%s", permission.Resource.Name, permission.Action)
			if permission.Effect == "allow" {
				userPermissions[permissionKey] = true
			} else if permission.Effect == "deny" {
				userPermissions[permissionKey] = false
			}
		}
	}

	// Check if user has all required permissions
	for _, requiredPermission := range requiredPermissions {
		if !userPermissions[requiredPermission] {
			return false, nil
		}
	}

	return true, nil
}

// ListActionsWithFilter retrieves actions with optional active filter
func (s *ActionService) ListActionsWithFilter(activeOnly *bool) ([]database.Action, error) {
	var actions []database.Action
	query := s.db

	if activeOnly != nil && *activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Find(&actions).Error; err != nil {
		return nil, err
	}
	return actions, nil
} 