package services

import (
	"fmt"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(email, username, password, firstName, lastName string, active bool) (*database.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := database.User{
		ID:        uuid.New(),
		Email:     email,
		Username:  username,
		Password:  string(hashedPassword),
		FirstName: firstName,
		LastName:  lastName,
		Active:    active,
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*database.User, error) {
	var user database.User
	if err := s.db.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*database.User, error) {
	var user database.User
	if err := s.db.Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// ListUsers retrieves all users
func (s *UserService) ListUsers() ([]database.User, error) {
	var users []database.User
	if err := s.db.Preload("Roles").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}
	return users, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(userID uuid.UUID, updates map[string]interface{}) (*database.User, error) {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Hash password if it's being updated
	if password, ok := updates["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		updates["password"] = string(hashedPassword)
	}

	if err := s.db.Model(&user).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Reload user with roles
	if err := s.db.Preload("Roles").First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("failed to reload user: %w", err)
	}

	return &user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.db.Delete(&user).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// AssignUserToRole assigns a user to a role
func (s *UserService) AssignUserToRole(userID, roleID uuid.UUID) error {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	var role database.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// Check if assignment already exists
	var count int64
	s.db.Model(&database.User{}).Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("users.id = ? AND user_roles.role_id = ?", user.ID, role.ID).Count(&count)
	
	if count > 0 {
		return fmt.Errorf("user is already assigned to role %s", role.Name)
	}

	if err := s.db.Model(&user).Association("Roles").Append(&role); err != nil {
		return fmt.Errorf("failed to assign user to role: %w", err)
	}

	return nil
}

// RemoveUserFromRole removes a user from a role
func (s *UserService) RemoveUserFromRole(userID, roleID uuid.UUID) error {
	var user database.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	var role database.Role
	if err := s.db.First(&role, roleID).Error; err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if err := s.db.Model(&user).Association("Roles").Delete(&role); err != nil {
		return fmt.Errorf("failed to remove user from role: %w", err)
	}

	return nil
} 