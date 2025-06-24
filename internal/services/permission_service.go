package services

import (
	"fmt"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PermissionService struct {
	db *gorm.DB
}

func NewPermissionService(db *gorm.DB) *PermissionService {
	return &PermissionService{db: db}
}

// CreatePermission creates a new permission
func (s *PermissionService) CreatePermission(resourceID uuid.UUID, action, effect string) (*database.Permission, error) {
	if effect != "allow" && effect != "deny" {
		return nil, fmt.Errorf("effect must be 'allow' or 'deny'")
	}

	// Check if resource exists
	var resource database.Resource
	if err := s.db.Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	permission := database.Permission{
		ID:         uuid.New(),
		ResourceID: resourceID,
		Action:     action,
		Effect:     effect,
	}

	if err := s.db.Create(&permission).Error; err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return &permission, nil
}

// GetPermissionByID retrieves a permission by ID
func (s *PermissionService) GetPermissionByID(permissionID uuid.UUID) (*database.Permission, error) {
	var permission database.Permission
	if err := s.db.Preload("Resource").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		return nil, fmt.Errorf("permission not found: %w", err)
	}
	return &permission, nil
}

// ListPermissions retrieves all permissions
func (s *PermissionService) ListPermissions() ([]database.Permission, error) {
	var permissions []database.Permission
	if err := s.db.Preload("Resource").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch permissions: %w", err)
	}
	return permissions, nil
}

// DeletePermission deletes a permission
func (s *PermissionService) DeletePermission(permissionID uuid.UUID) error {
	var permission database.Permission
	if err := s.db.Preload("Resource").Where("id = ?", permissionID).First(&permission).Error; err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if err := s.db.Delete(&permission).Error; err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// CheckUserPermission checks if a user has a specific permission
func (s *PermissionService) CheckUserPermission(userID uuid.UUID, resourceName, action string) (bool, error) {
	var user database.User
	if err := s.db.Preload("Roles.Permissions.Resource").First(&user, userID).Error; err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource.Name == resourceName && 
			   perm.Action == action && 
			   perm.Effect == "allow" {
				return true, nil
			}
		}
	}

	return false, nil
} 