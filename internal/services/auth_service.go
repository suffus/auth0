package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/YubiApp/internal/config"
	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthService struct {
	db            *gorm.DB
	deviceService *DeviceService
	config        *config.Config
}

func NewAuthService(db *gorm.DB, config *config.Config) *AuthService {
	return &AuthService{
		db:            db,
		deviceService: NewDeviceService(db),
		config:        config,
	}
}

// AuthenticateDevice authenticates a user using a device and checks permissions
func (s *AuthService) AuthenticateDevice(deviceType, authCode, requiredPermission string) (*database.User, error) {
	var device *database.Device
	var err error

	switch deviceType {
	case "yubikey":
		device, err = s.authenticateYubikey(authCode)
	case "totp":
		device, err = s.authenticateTOTP(authCode)
	case "sms":
		device, err = s.authenticateSMS(authCode)
	case "email":
		device, err = s.authenticateEmail(authCode)
	default:
		return nil, fmt.Errorf("unsupported device type: %s", deviceType)
	}

	if err != nil {
		return nil, err
	}

	// Get user associated with the device
	var user database.User
	if err := s.db.Preload("Roles.Permissions.Resource").First(&user, device.UserID).Error; err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if user and device are active
	if !user.Active {
		return nil, fmt.Errorf("user is not active")
	}
	if !device.Active {
		return nil, fmt.Errorf("device is not active")
	}

	// If no permission required, just return the user
	if requiredPermission == "" {
		s.deviceService.UpdateDeviceLastUsed(device.ID)
		s.logAuthentication(device, &user, true, requiredPermission, "")
		return &user, nil
	}

	// Check if user has the required permission
	hasPermission := false
	
	// Try to parse as UUID first
	if permissionID, err := uuid.Parse(requiredPermission); err == nil {
		// It's a UUID, check if user has this specific permission
		hasPermission = s.checkUserHasPermissionByID(&user, permissionID)
	} else {
		// It's not a UUID, try to parse as resource:action format
		parts := strings.Split(requiredPermission, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid permission format: %s (expected 'resource:action' or permission UUID)", requiredPermission)
		}
		resourceName, action := parts[0], parts[1]
		hasPermission = s.checkUserHasPermissionByResourceAction(&user, resourceName, action)
	}

	if !hasPermission {
		s.logAuthentication(device, &user, false, requiredPermission, "permission denied")
		return nil, fmt.Errorf("permission denied: %s", requiredPermission)
	}

	// Update device last used timestamp
	s.deviceService.UpdateDeviceLastUsed(device.ID)

	// Log successful authentication
	s.logAuthentication(device, &user, true, requiredPermission, "")

	return &user, nil
}

// checkUserHasPermissionByID checks if a user has a specific permission by UUID
func (s *AuthService) checkUserHasPermissionByID(user *database.User, permissionID uuid.UUID) bool {
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.ID == permissionID && perm.Effect == "allow" {
				return true
			}
		}
	}
	return false
}

// checkUserHasPermissionByResourceAction checks if a user has a permission by resource name and action
func (s *AuthService) checkUserHasPermissionByResourceAction(user *database.User, resourceName, action string) bool {
	for _, role := range user.Roles {
		for _, perm := range role.Permissions {
			if perm.Resource.Name == resourceName && 
			   perm.Action == action && 
			   perm.Effect == "allow" {
				return true
			}
		}
	}
	return false
}

// authenticateYubikey authenticates using YubiKey OTP
func (s *AuthService) authenticateYubikey(otp string) (*database.Device, error) {
	// Extract device ID from OTP (first 12 characters)
	if len(otp) < 12 {
		return nil, fmt.Errorf("invalid YubiKey OTP format")
	}
	deviceID := otp[:12]

	// Verify OTP with Yubico servers
	if err := s.verifyYubikeyOTP(otp); err != nil {
		return nil, fmt.Errorf("OTP verification failed: %w", err)
	}

	// Find the device in our database
	return s.deviceService.GetDeviceByIdentifier("yubikey", deviceID)
}

// authenticateTOTP authenticates using TOTP
func (s *AuthService) authenticateTOTP(code string) (*database.Device, error) {
	// For now, we'll need the device ID to be provided separately
	// In a real implementation, you might encode the device ID in the code
	// or require it to be provided explicitly
	return nil, fmt.Errorf("TOTP authentication not yet implemented")
}

// authenticateSMS authenticates using SMS
func (s *AuthService) authenticateSMS(code string) (*database.Device, error) {
	// For now, we'll need the device ID to be provided separately
	return nil, fmt.Errorf("SMS authentication not yet implemented")
}

// authenticateEmail authenticates using Email
func (s *AuthService) authenticateEmail(code string) (*database.Device, error) {
	// For now, we'll need the device ID to be provided separately
	return nil, fmt.Errorf("Email authentication not yet implemented")
}

// verifyYubikeyOTP verifies the OTP with Yubico servers
func (s *AuthService) verifyYubikeyOTP(otp string) error {
	params := url.Values{}
	params.Add("id", s.config.Yubikey.ClientID)
	params.Add("otp", otp)
	
	// Generate alphanumeric nonce (16-40 characters, no hyphens)
	nonceBytes := make([]byte, 20)
	rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)
	params.Add("nonce", nonce)

	resp, err := http.Get(fmt.Sprintf("%s?%s", s.config.Yubikey.APIURL, params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to verify OTP with Yubico: %w", err)
	}
	defer resp.Body.Close()

	// Read the response as plain text
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Yubico response: %w", err)
	}

	// Parse key-value pairs
	lines := strings.Split(string(body), "\n")
	status := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "status=") {
			status = strings.TrimSpace(strings.TrimPrefix(line, "status="))
			break
		}
	}

	switch strings.ToLower(status) {
	case "ok":
		return nil
	case "replayed_otp":
		return fmt.Errorf("replayed OTP detected")
	case "bad_otp":
		return fmt.Errorf("invalid OTP format")
	case "missing_parameter":
		return fmt.Errorf("missing parameter in OTP verification")
	case "no_such_client":
		return fmt.Errorf("invalid client ID")
	case "operation_not_allowed":
		return fmt.Errorf("operation not allowed")
	case "backend_error":
		return fmt.Errorf("Yubico backend error")
	default:
		return fmt.Errorf("Yubico verification failed with status: %s", status)
	}
}

// logAuthentication logs the authentication attempt
func (s *AuthService) logAuthentication(device *database.Device, user *database.User, success bool, permissionChecked, errorMsg string) {
	authLog := database.AuthenticationLog{
		ID:        uuid.New(),
		UserID:    user.ID,
		DeviceID:  device.ID,
		Type:      "mfa", // Use 'mfa' to comply with DB constraint
		Success:   success,
		IPAddress: "", // Will be set by web handlers
		UserAgent: "", // Will be set by web handlers
		Details: map[string]interface{}{
			"device_type":        device.Type,
			"permission_checked": permissionChecked,
		},
	}

	if errorMsg != "" {
		authLog.Details["error"] = errorMsg
	}

	s.db.Create(&authLog)
} 