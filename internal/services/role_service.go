package services

import (
	"fmt"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleService struct {
	db *gorm.DB
}

func NewRoleService(db *gorm.DB) *RoleService {
	return &RoleService{db: db}
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(name, description string) (*database.Role, error) {
	role := database.Role{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
	}

	if err := s.db.Create(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return &role, nil
}

// GetRoleByID retrieves a role by ID
func (s *RoleService) GetRoleByID(roleID uuid.UUID) (*database.Role, error) {
	var role database.Role
	if err := s.db.Preload("Permissions.Resource").First(&role, roleID).Error; err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}
	return &role, nil
}

// GetRoleByName retrieves a role by name
func (s *RoleService) GetRoleByName(name string) (*database.Role, error) {
	var role database.Role
	if err := s.db.Preload("Permissions.Resource").Where("name = ?", name).First(&role).Error; err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}
	return &role, nil
}

// ListRoles retrieves all roles
func (s *RoleService) ListRoles() ([]database.Role, error) {
	var roles []database.Role
	if err := s.db.Preload("Permissions.Resource").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %w", err)
	}
	return roles, nil
}

// UpdateRole updates a role
func (s *RoleService) UpdateRole(roleID uuid.UUID, updates map[string]interface{}) (*database.Role, error) {
	var role database.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	if err := s.db.Model(&role).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Reload role with permissions
	if err := s.db.Preload("Permissions.Resource").First(&role, roleID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload role: %w", err)
	}

	return &role, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(roleID uuid.UUID) error {
	var role database.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if err := s.db.Delete(&role).Error; err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// AssignPermissionToRole assigns a permission to a role
func (s *RoleService) AssignPermissionToRole(roleID, permissionID uuid.UUID) error {
	var role database.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	var permission database.Permission
	if err := s.db.Preload("Resource").First(&permission, permissionID).Error; err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Check if assignment already exists
	var count int64
	s.db.Model(&database.Role{}).Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
		Where("roles.id = ? AND role_permissions.permission_id = ?", role.ID, permission.ID).Count(&count)
	
	if count > 0 {
		return fmt.Errorf("permission %s:%s is already assigned to role %s", 
			permission.Resource.Name, permission.Action, role.Name)
	}

	if err := s.db.Model(&role).Association("Permissions").Append(&permission); err != nil {
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (s *RoleService) RemovePermissionFromRole(roleID, permissionID uuid.UUID) error {
	var role database.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	var permission database.Permission
	if err := s.db.Preload("Resource").First(&permission, permissionID).Error; err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	if err := s.db.Model(&role).Association("Permissions").Delete(&permission); err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	return nil
} 