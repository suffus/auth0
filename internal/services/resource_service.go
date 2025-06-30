package services

import (
	"fmt"
	"strings"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResourceService struct {
	db *gorm.DB
}

func NewResourceService(db *gorm.DB) *ResourceService {
	return &ResourceService{db: db}
}

// CreateResource creates a new resource
func (s *ResourceService) CreateResource(name, resourceType, location, department string, active bool) (*database.Resource, error) {
	// Validate resource name - no colons allowed to avoid ambiguity in permission format
	if strings.Contains(name, ":") {
		return nil, fmt.Errorf("resource name cannot contain colons (':') to avoid ambiguity in permission format")
	}

	// Validate resource type
	validTypes := []string{"server", "service", "database", "application"}
	validType := false
	for _, t := range validTypes {
		if resourceType == t {
			validType = true
			break
		}
	}
	if !validType {
		return nil, fmt.Errorf("resource type must be one of: %v", validTypes)
	}

	resource := database.Resource{
		ID:         uuid.New(),
		Name:       name,
		Type:       resourceType,
		Location:   location,
		Department: department,
		Active:     active,
	}

	if err := s.db.Create(&resource).Error; err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	return &resource, nil
}

// GetResourceByID retrieves a resource by ID
func (s *ResourceService) GetResourceByID(resourceID uuid.UUID) (*database.Resource, error) {
	var resource database.Resource
	if err := s.db.Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}
	return &resource, nil
}

// GetResourceByName retrieves a resource by name
func (s *ResourceService) GetResourceByName(name string) (*database.Resource, error) {
	var resource database.Resource
	if err := s.db.Where("name = ?", name).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}
	return &resource, nil
}

// ListResources retrieves all resources
func (s *ResourceService) ListResources() ([]database.Resource, error) {
	var resources []database.Resource
	if err := s.db.Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %w", err)
	}
	return resources, nil
}

// UpdateResource updates a resource
func (s *ResourceService) UpdateResource(resourceID uuid.UUID, updates map[string]interface{}) (*database.Resource, error) {
	var resource database.Resource
	if err := s.db.Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("resource not found: %w", err)
	}

	// Validate resource name if it's being updated - no colons allowed
	if name, ok := updates["name"].(string); ok {
		if strings.Contains(name, ":") {
			return nil, fmt.Errorf("resource name cannot contain colons (':') to avoid ambiguity in permission format")
		}
	}

	// Validate resource type if it's being updated
	if resourceType, ok := updates["type"].(string); ok {
		validTypes := []string{"server", "service", "database", "application"}
		validType := false
		for _, t := range validTypes {
			if resourceType == t {
				validType = true
				break
			}
		}
		if !validType {
			return nil, fmt.Errorf("resource type must be one of: %v", validTypes)
		}
	}

	if err := s.db.Model(&resource).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update resource: %w", err)
	}

	// Reload resource
	if err := s.db.Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("failed to reload resource: %w", err)
	}

	return &resource, nil
}

// DeleteResource deletes a resource
func (s *ResourceService) DeleteResource(resourceID uuid.UUID) error {
	var resource database.Resource
	if err := s.db.Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return fmt.Errorf("resource not found: %w", err)
	}

	if err := s.db.Delete(&resource).Error; err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
} 