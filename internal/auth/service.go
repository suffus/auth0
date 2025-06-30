package auth

import (
	"context"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
)

type AuthenticationRequest struct {
	Username string
	Password string
	DeviceID uuid.UUID
	OTP      string
	Type     string // "yubikey", "totp", "sms", "email"
}

type AuthenticationResponse struct {
	User         *database.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type Service interface {
	// User authentication
	Authenticate(ctx context.Context, req AuthenticationRequest) (*AuthenticationResponse, error)
	ValidateToken(ctx context.Context, token string) (*database.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthenticationResponse, error)
	Logout(ctx context.Context, token string) error

	// Device management
	RegisterDevice(ctx context.Context, userID uuid.UUID, deviceType string, identifier string) (*database.Device, error)
	VerifyDevice(ctx context.Context, deviceID uuid.UUID, otp string) error
	ListDevices(ctx context.Context, userID uuid.UUID) ([]database.Device, error)
	RemoveDevice(ctx context.Context, deviceID uuid.UUID) error

	// MFA
	InitiateMFA(ctx context.Context, userID uuid.UUID, deviceType string) error
	VerifyMFA(ctx context.Context, userID uuid.UUID, deviceID uuid.UUID, otp string) error
}

// Implementation will be provided in separate files for each authentication method:
// - yubikey.go
// - totp.go
// - sms.go
// - email.go 