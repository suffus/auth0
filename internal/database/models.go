package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Email     string `gorm:"uniqueIndex"`
	Username  string `gorm:"uniqueIndex"`
	Password  string // Hashed password
	FirstName string
	LastName  string
	Active    bool `gorm:"default:true"`

	Roles    []Role    `gorm:"many2many:user_roles;"`
	Devices  []Device  `gorm:"foreignKey:UserID"`
	Sessions []Session `gorm:"foreignKey:UserID"`
}

type Role struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name        string `gorm:"uniqueIndex"`
	Description string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}

type Resource struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name       string `gorm:"uniqueIndex"`
	Type       string // "server", "service", "database", "application"
	Location   string
	Department string
	Active     bool `gorm:"default:true"`
}

type Permission struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time

	ResourceID uuid.UUID `gorm:"type:uuid"`
	Resource   Resource  `gorm:"foreignKey:ResourceID"`
	Action     string
	Effect     string // "allow" or "deny"
}

type Device struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID      uuid.UUID `gorm:"type:uuid"`
	User        User      `gorm:"foreignKey:UserID"`
	Type        string    // "yubikey", "totp", "sms", "email"
	Identifier  string    // Device identifier (e.g., Yubikey public ID, phone number)
	Secret      string    // For TOTP/device-specific secrets
	LastUsedAt  time.Time
	VerifiedAt  time.Time
	Active      bool
	Properties  map[string]interface{} `gorm:"type:jsonb"`
}

type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID       uuid.UUID `gorm:"type:uuid"`
	Token        string    `gorm:"uniqueIndex"`
	RefreshToken string    `gorm:"uniqueIndex"`
	ExpiresAt    time.Time
	LastUsedAt   time.Time
	UserAgent    string
	IPAddress    string
	Active       bool
}

type AuthenticationLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time

	UserID    uuid.UUID `gorm:"type:uuid"`
	DeviceID  uuid.UUID `gorm:"type:uuid"`
	Type      string    // "login", "logout", "refresh", "mfa"
	Success   bool
	IPAddress string
	UserAgent string
	Details   map[string]interface{} `gorm:"type:jsonb"`
} 