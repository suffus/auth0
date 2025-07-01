package services

import (
	"fmt"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocationService struct {
	db *gorm.DB
}

func NewLocationService(db *gorm.DB) *LocationService {
	return &LocationService{db: db}
}

// CreateLocation creates a new location
func (s *LocationService) CreateLocation(name, description, address, locationType string, active bool) (*database.Location, error) {
	// Validate location type
	validTypes := []string{"office", "home", "event", "other"}
	validType := false
	for _, t := range validTypes {
		if locationType == t {
			validType = true
			break
		}
	}
	if !validType {
		return nil, fmt.Errorf("location type must be one of: %v", validTypes)
	}

	location := database.Location{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Address:     address,
		Type:        locationType,
		Active:      active,
	}

	if err := s.db.Create(&location).Error; err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	return &location, nil
}

// GetLocationByID retrieves a location by ID
func (s *LocationService) GetLocationByID(locationID uuid.UUID) (*database.Location, error) {
	var location database.Location
	if err := s.db.Where("id = ?", locationID).First(&location).Error; err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}
	return &location, nil
}

// GetLocationByName retrieves a location by name
func (s *LocationService) GetLocationByName(name string) (*database.Location, error) {
	var location database.Location
	if err := s.db.Where("name = ?", name).First(&location).Error; err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}
	return &location, nil
}

// ListLocations retrieves all locations
func (s *LocationService) ListLocations() ([]database.Location, error) {
	var locations []database.Location
	if err := s.db.Find(&locations).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch locations: %w", err)
	}
	return locations, nil
}

// ListActiveLocations retrieves only active locations
func (s *LocationService) ListActiveLocations() ([]database.Location, error) {
	var locations []database.Location
	if err := s.db.Where("active = ?", true).Find(&locations).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch active locations: %w", err)
	}
	return locations, nil
}

// ListLocationsByType retrieves locations by type
func (s *LocationService) ListLocationsByType(locationType string) ([]database.Location, error) {
	var locations []database.Location
	if err := s.db.Where("type = ?", locationType).Find(&locations).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch locations by type: %w", err)
	}
	return locations, nil
}

// UpdateLocation updates a location
func (s *LocationService) UpdateLocation(locationID uuid.UUID, updates map[string]interface{}) (*database.Location, error) {
	var location database.Location
	if err := s.db.Where("id = ?", locationID).First(&location).Error; err != nil {
		return nil, fmt.Errorf("location not found: %w", err)
	}

	// Validate location type if it's being updated
	if locationType, ok := updates["type"].(string); ok {
		validTypes := []string{"office", "home", "event", "other"}
		validType := false
		for _, t := range validTypes {
			if locationType == t {
				validType = true
				break
			}
		}
		if !validType {
			return nil, fmt.Errorf("location type must be one of: %v", validTypes)
		}
	}

	if err := s.db.Model(&location).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update location: %w", err)
	}

	// Reload location
	if err := s.db.Where("id = ?", locationID).First(&location).Error; err != nil {
		return nil, fmt.Errorf("failed to reload location: %w", err)
	}

	return &location, nil
}

// DeleteLocation marks a location as inactive (soft delete)
func (s *LocationService) DeleteLocation(locationID uuid.UUID) error {
	var location database.Location
	if err := s.db.Where("id = ?", locationID).First(&location).Error; err != nil {
		return fmt.Errorf("location not found: %w", err)
	}

	// Soft delete by setting active to false
	if err := s.db.Model(&location).Update("active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate location: %w", err)
	}

	return nil
}

// HardDeleteLocation permanently deletes a location
func (s *LocationService) HardDeleteLocation(locationID uuid.UUID) error {
	var location database.Location
	if err := s.db.Where("id = ?", locationID).First(&location).Error; err != nil {
		return fmt.Errorf("location not found: %w", err)
	}

	if err := s.db.Delete(&location).Error; err != nil {
		return fmt.Errorf("failed to delete location: %w", err)
	}

	return nil
} 