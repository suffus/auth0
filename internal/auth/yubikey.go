package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrYubikeyInvalidOTP        = errors.New("invalid yubikey OTP")
	ErrYubikeyReplayedOTP       = errors.New("replayed OTP")
	ErrYubikeyDeviceNotFound    = errors.New("yubikey device not registered")
	ErrYubikeyVerificationError = errors.New("yubikey verification failed")
	ErrPermissionDenied         = errors.New("permission denied")
)

type YubikeyConfig struct {
	ClientID  string
	SecretKey string
	APIURL    string
}

type YubikeyService struct {
	db     *gorm.DB
	config YubikeyConfig
}

type YubikeyResponse struct {
	Status     string `json:"status"`
	Timestamp  string `json:"timestamp"`
	SessionID  string `json:"sessionid"`
	OTP        string `json:"otp"`
	Nonce      string `json:"nonce"`
	H          string `json:"h"`  // HMAC signature
}

func NewYubikeyService(db *gorm.DB, config YubikeyConfig) *YubikeyService {
	return &YubikeyService{
		db:     db,
		config: config,
	}
}

// AuthenticateYubikey handles Yubikey OTP authentication and permission verification
func (s *YubikeyService) AuthenticateYubikey(ctx context.Context, otp, requiredPermission string) (*database.User, error) {
	// Extract device ID from OTP (first 12 characters)
	if len(otp) < 12 {
		return nil, ErrYubikeyInvalidOTP
	}
	deviceID := otp[:12]

	// Find the device in our database
	var device database.Device
	if err := s.db.Where("type = ? AND identifier = ?", "yubikey", deviceID).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrYubikeyDeviceNotFound
		}
		return nil, err
	}

	// Verify OTP with Yubico servers
	if err := s.verifyOTP(otp); err != nil {
		return nil, err
	}

	// Get user associated with the device
	var user database.User
	if err := s.db.Preload("Roles.Permissions.Resource").First(&user, device.UserID).Error; err != nil {
		return nil, err
	}

	// Verify user has required permission
	if !s.hasPermission(&user, requiredPermission) {
		return nil, ErrPermissionDenied
	}

	// Update device last used timestamp
	s.db.Model(&device).Update("last_used_at", "NOW()")

	// Log authentication
	authLog := database.AuthenticationLog{
		ID:        uuid.New(),
		UserID:    user.ID,
		DeviceID:  device.ID,
		Type:      "yubikey",
		Success:   true,
		IPAddress: "", // Should be set by handler
		UserAgent: "", // Should be set by handler
		Details: map[string]interface{}{
			"permission_checked": requiredPermission,
		},
	}
	s.db.Create(&authLog)

	return &user, nil
}

// verifyOTP verifies the OTP with Yubico servers
func (s *YubikeyService) verifyOTP(otp string) error {
	params := url.Values{}
	params.Add("id", s.config.ClientID)
	params.Add("otp", otp)
	params.Add("nonce", uuid.New().String())

	resp, err := http.Get(fmt.Sprintf("%s?%s", s.config.APIURL, params.Encode()))
	if err != nil {
		return ErrYubikeyVerificationError
	}
	defer resp.Body.Close()

	var yResp YubikeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&yResp); err != nil {
		return ErrYubikeyVerificationError
	}

	switch strings.ToLower(yResp.Status) {
	case "ok":
		return nil
	case "replayed_otp":
		return ErrYubikeyReplayedOTP
	default:
		return ErrYubikeyVerificationError
	}
}

// hasPermission checks if the user has the required permission
func (s *YubikeyService) hasPermission(user *database.User, requiredPermission string) bool {
	parts := strings.Split(requiredPermission, ":")
	if len(parts) != 2 {
		return false
	}
	resourceName, action := parts[0], parts[1]

	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource.Name == resourceName && perm.Action == action && perm.Effect == "allow" {
				return true
			}
		}
	}
	return false
} 