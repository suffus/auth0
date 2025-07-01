package database

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
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
}

type Role struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name        string `gorm:"uniqueIndex"`
	Description string
	Active      bool `gorm:"default:true"`
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

type Action struct {
	ID                  uuid.UUID     `gorm:"type:uuid;primary_key;"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Name                string        `gorm:"uniqueIndex"`
	ActivityType        string        `gorm:"type:varchar(20);default:'other';check:activity_type IN ('user', 'system', 'automated', 'other')"`
	RequiredPermissions pgtype.JSONB  `gorm:"type:jsonb"`
	Details             pgtype.JSONB  `gorm:"type:jsonb;default:'{}'::jsonb"`
	Active              bool          `gorm:"default:true"`
}

type Device struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	UserID      uuid.UUID `gorm:"type:uuid"`
	User        User      `gorm:"foreignKey:UserID"`
	Name        string    // Device name
	Type        string    // "yubikey", "totp", "sms", "email"
	SerialNumber string   // Device serial number
	Identifier  string    // Device identifier (e.g., Yubikey public ID, phone number)
	Secret      string    // For TOTP/device-specific secrets
	LastUsedAt  time.Time
	VerifiedAt  time.Time
	Active      bool
	Properties  map[string]interface{} `gorm:"type:jsonb"`
}

// Session represents a user session stored in Redis (not in PostgreSQL)
type Session struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	UserID       uuid.UUID `json:"user_id"`
	DeviceID     uuid.UUID `json:"device_id"`
	AccessCount  int       `json:"access_count"`
	RefreshCount int       `json:"refresh_count"`
	CreatedAt    time.Time `json:"created_at"`
	ExpiresAt    time.Time `json:"expires_at"`
	IsValid      bool      `json:"is_valid"`
}

// SessionToken represents JWT token claims for sessions
type SessionToken struct {
	SessionID    string `json:"session_id"`
	UserID       string `json:"user_id"`
	DeviceID     string `json:"device_id"`
	AccessCount  int    `json:"access_count"`
	RefreshCount int    `json:"refresh_count"`
	jwt.RegisteredClaims
}

// RefreshToken represents JWT token claims for refresh tokens
type RefreshToken struct {
	SessionID    string `json:"session_id"`
	UserID       string `json:"user_id"`
	DeviceID     string `json:"device_id"`
	RefreshCount int    `json:"refresh_count"`
	jwt.RegisteredClaims
}

type AuthenticationLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time

	UserID     *uuid.UUID `gorm:"type:uuid"`
	User       *User      `gorm:"foreignKey:UserID"`
	DeviceID   uuid.UUID  `gorm:"type:uuid"`
	Device     Device     `gorm:"foreignKey:DeviceID"`
	ActionID   *uuid.UUID `gorm:"type:uuid"`
	Type       string     // "login", "logout", "refresh", "mfa", "action"
	Success    bool
	IPAddress  string
	UserAgent  string
	OTP        string     // YubiKey OTP
	Timestamp  time.Time  // Authentication timestamp
	Details    pgtype.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"`
}

type DeviceRegistration struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt       time.Time

	RegistrarUserID uuid.UUID `gorm:"type:uuid"`
	RegistrarUser   User      `gorm:"foreignKey:RegistrarUserID"`

	DeviceID        uuid.UUID `gorm:"type:uuid"`
	Device          Device    `gorm:"foreignKey:DeviceID"`

	TargetUserID    *uuid.UUID `gorm:"type:uuid"` // NULL for deregistration
	TargetUser      *User      `gorm:"foreignKey:TargetUserID"`

	ActionType      string `gorm:"type:varchar(20);check:action_type IN ('register', 'deregister')"`
	Reason          string
	IPAddress       string
	UserAgent       string
	Notes           string
}

type Location struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name        string `gorm:"uniqueIndex"`
	Description string
	Address     string
	Type        string `gorm:"type:varchar(20);default:'office';check:type IN ('office', 'home', 'event', 'other')"`
	Active      bool   `gorm:"default:true"`
}

type UserStatus struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name        string `gorm:"uniqueIndex"`
	Description string
	Type        string `gorm:"type:varchar(30);default:'working';check:type IN ('working', 'break', 'leave', 'travel', 'other')"`
	Active      bool   `gorm:"default:true"`
}

type UserActivityHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time

	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	User         User      `gorm:"foreignKey:UserID"`
	ActionID     uuid.UUID `gorm:"type:uuid;not null"`
	Action       Action    `gorm:"foreignKey:ActionID"`
	FromDateTime time.Time `gorm:"not null"`
	ToDateTime   *time.Time `gorm:"type:timestamp"`
	LocationID   *uuid.UUID `gorm:"type:uuid"`
	Location     *Location `gorm:"foreignKey:LocationID"`
	StatusID     *uuid.UUID `gorm:"type:uuid"`
	Status       *UserStatus `gorm:"foreignKey:StatusID"`
	Details      pgtype.JSONB `gorm:"type:jsonb;default:'{}'::jsonb"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleID uuid.UUID `gorm:"type:uuid;primaryKey"`
	User   User      `gorm:"foreignKey:UserID"`
	Role   Role      `gorm:"foreignKey:RoleID"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role         Role      `gorm:"foreignKey:RoleID"`
	Permission   Permission `gorm:"foreignKey:PermissionID"`
} 