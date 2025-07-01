package utils

import (
	"fmt"
	"strings"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
)

// FindUserByString finds a user by email, username, or UUID
func FindUserByString(identifier string) (*database.User, error) {
	var user database.User
	
	// Try to parse as UUID first
	if _, err := uuid.Parse(identifier); err == nil {
		if err := DB.Preload("Roles.Permissions.Resource").First(&user, "id = ?", identifier).Error; err != nil {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return &user, nil
	}
	
	// Try to find by email
	if err := DB.Preload("Roles.Permissions.Resource").First(&user, "email = ?", identifier).Error; err == nil {
		return &user, nil
	}
	
	// Try to find by username
	if err := DB.Preload("Roles.Permissions.Resource").First(&user, "username = ?", identifier).Error; err == nil {
		return &user, nil
	}
	
	return nil, fmt.Errorf("user not found with identifier: %s", identifier)
}

// FindResourceByString finds a resource by name or UUID
func FindResourceByString(identifier string) (*database.Resource, error) {
	var resource database.Resource
	
	// Try to parse as UUID first
	if _, err := uuid.Parse(identifier); err == nil {
		if err := DB.First(&resource, "id = ?", identifier).Error; err != nil {
			return nil, fmt.Errorf("resource not found: %w", err)
		}
		return &resource, nil
	}
	
	// Try to find by name
	if err := DB.First(&resource, "name = ?", identifier).Error; err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}
	
	return &resource, nil
}

// FindPermissionByString finds a permission by ID or resource:action format
func FindPermissionByString(identifier string) (*database.Permission, error) {
	var permission database.Permission
	
	// Try to parse as UUID first
	if _, err := uuid.Parse(identifier); err == nil {
		if err := DB.Preload("Resource").First(&permission, "id = ?", identifier).Error; err != nil {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		return &permission, nil
	}
	
	// Try to parse as resource:action format
	if strings.Contains(identifier, ":") {
		parts := strings.Split(identifier, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid permission format, expected 'resource:action'")
		}
		
		resourceName := parts[0]
		action := parts[1]
		
		if err := DB.Preload("Resource").First(&permission, "action = ? AND resource_id = (SELECT id FROM resources WHERE name = ?)", action, resourceName).Error; err != nil {
			return nil, fmt.Errorf("permission not found: %w", err)
		}
		return &permission, nil
	}
	
	return nil, fmt.Errorf("permission not found with identifier: %s", identifier)
}

// ParseUUIDArray parses a comma-separated string of UUIDs into a slice
func ParseUUIDArray(uuidStr string) ([]uuid.UUID, error) {
	if uuidStr == "" {
		return []uuid.UUID{}, nil
	}
	
	parts := strings.Split(uuidStr, ",")
	uuids := make([]uuid.UUID, 0, len(parts))
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		id, err := uuid.Parse(part)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID '%s': %w", part, err)
		}
		uuids = append(uuids, id)
	}
	
	return uuids, nil
} 