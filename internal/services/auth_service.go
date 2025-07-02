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
	"github.com/jackc/pgtype"
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
// Returns both user and device information
func (s *AuthService) AuthenticateDevice(deviceType, authCode string, requiredPermissions ...string) (*database.User, *database.Device, error) {
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
		return nil, nil, fmt.Errorf("unsupported device type: %s", deviceType)
	}

	if err != nil {
		return nil, nil, err
	}

	// Get user associated with the device
	var user database.User
	if err := s.db.Preload("Roles.Permissions.Resource").Where("id = ?", device.UserID).First(&user).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to find user: %w", err)
	}

	details := map[string]interface{}{
		"user": user.Username,
		"device_id": device.ID,
		"device_type": device.Type,
		"auth_code": authCode,
		"type": "mfa",
		"permissions_checked": requiredPermissions,
	}

	// Check if user and device are active
	if !user.Active {
		return nil, nil, fmt.Errorf("user is not active")
	}
	if !device.Active {
		return nil, nil, fmt.Errorf("device is not active")
	}

	// Filter out empty permission strings
	var validPermissions []string
	for _, perm := range requiredPermissions {
		if strings.TrimSpace(perm) != "" {
			validPermissions = append(validPermissions, perm)
		}
	}

	// If no permissions required, just return the user and device
	if len(validPermissions) == 0 {
		s.deviceService.UpdateDeviceLastUsed(device.ID)
		s.logAuthentication(device, &user, true, "", "", details)
		return &user, device, nil
	}

	// Check if user has ALL required permissions
	for _, requiredPermission := range validPermissions {
		hasPermission := false
		
		// Try to parse as UUID first
		if permissionID, err := uuid.Parse(requiredPermission); err == nil {
			// It's a UUID, check if user has this specific permission
			hasPermission = s.checkUserHasPermissionByID(&user, permissionID)
		} else {
			// It's not a UUID, try to parse as resource:action format
			parts := strings.Split(requiredPermission, ":")
			if len(parts) != 2 {
				return nil, nil, fmt.Errorf("invalid permission format: %s (expected 'resource:action' or permission UUID)", requiredPermission)
			}
			resourceName, action := parts[0], parts[1]
			hasPermission = s.checkUserHasPermissionByResourceAction(&user, resourceName, action)
		}

		if !hasPermission {
			s.logAuthentication(device, &user, false, requiredPermission, "permission denied", details)
			return nil, nil, fmt.Errorf("permission denied: %s", requiredPermission)
		}
	}

	// Update device last used timestamp
	s.deviceService.UpdateDeviceLastUsed(device.ID)

	// Log successful authentication
	s.logAuthentication(device, &user, true, strings.Join(validPermissions, ","), "", details)

	return &user, device, nil
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
func (s *AuthService) logAuthentication(device *database.Device, user *database.User, success bool, permissionChecked, errorMsg string, details map[string]interface{}) {
	s.LogAuthentication(map[string]interface{}{
		"user_id": user.ID,
		"device_id": device.ID,
		"type": "mfa",
		"success": success,
		"permission_checked": permissionChecked,
		"error_msg": errorMsg,
		"details": details,
	})
}

// LogAuthentication logs an authentication event with custom data
func (s *AuthService) LogAuthentication(logData map[string]interface{}) error {
	authLog := database.AuthenticationLog{
		ID:        uuid.New(),
		Type:      "action", // Use 'action' for action events
		Success:   true,
		IPAddress: "",
		UserAgent: "",
	}

	// Extract fields from logData
	if userID, ok := logData["user_id"].(uuid.UUID); ok {
		authLog.UserID = &userID
	}
	if deviceID, ok := logData["device_id"].(uuid.UUID); ok {
		authLog.DeviceID = deviceID
	}
	if actionID, ok := logData["action_id"].(uuid.UUID); ok {
		authLog.ActionID = &actionID
	}
	if success, ok := logData["success"].(bool); ok {
		authLog.Success = success
	}
	if ipAddress, ok := logData["ip_address"].(string); ok {
		authLog.IPAddress = ipAddress
	}
	if userAgent, ok := logData["user_agent"].(string); ok {
		authLog.UserAgent = userAgent
	}
	var detailsJSONB pgtype.JSONB
	// Set Details as JSONB only if we have data
	if details, ok := logData["details"].(map[string]interface{}); ok && len(details) > 0 {
		if err := detailsJSONB.Set(details); err != nil {
			return fmt.Errorf("failed to convert details to JSONB: %w", err)
		}
	}
	if detailsJSONB.Status != pgtype.Present {
		detailsJSONB = pgtype.JSONB{
			Bytes: []byte("{}"),
			Status: pgtype.Present,
		}
	}
	authLog.Details = detailsJSONB
	// Set type to "action" for action events
	authLog.Type = logData["type"].(string)

	return s.db.Create(&authLog).Error
}

// CheckUserPermissionByResourceAction checks if a user has a specific permission by resource name and action
func (s *AuthService) CheckUserPermissionByResourceAction(userID uuid.UUID, resourceName, action string) (bool, error) {
	var user database.User
	if err := s.db.Preload("Roles.Permissions.Resource").Where("id = ?", userID).First(&user).Error; err != nil {
		return false, err
	}

	return s.checkUserHasPermissionByResourceAction(&user, resourceName, action), nil
}

// GetDB returns the database instance (for use in handlers)
func (s *AuthService) GetDB() *gorm.DB {
	return s.db
} 