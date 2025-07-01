package utils

import (
	"fmt"
	"log"

	"github.com/YubiApp/internal/config"
	"github.com/YubiApp/internal/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Global database connection
var DB *gorm.DB

// InitDatabase initializes the database connection without running migrations
func InitDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn), // Reduce logging verbosity
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connected successfully")
	DB = db
	return db, nil
}

// RunMigrations runs database migrations
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")
	
	if err := DB.AutoMigrate(
		&database.User{},
		&database.Role{},
		&database.Permission{},
		&database.Resource{},
		&database.Device{},
		&database.Action{},
		&database.AuthenticationLog{},
		&database.DeviceRegistration{},
		&database.Location{},
		&database.UserStatus{},
		&database.UserActivityHistory{},
	); err != nil {
		return fmt.Errorf("failed to auto-migrate database: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// GetDB returns the global database connection
func GetDB() *gorm.DB {
	return DB
} 