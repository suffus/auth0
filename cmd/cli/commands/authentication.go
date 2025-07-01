package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/cmd/cli/utils"
	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var authenticateUserCmd = &cobra.Command{
	Use:   "user",
	Short: "Authenticate a user with YubiKey OTP",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userIdentifier := args[0]
		otp := args[1]

		// Find the user
		user, err := utils.FindUserByString(userIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		if !user.Active {
			return fmt.Errorf("user %s is not active", user.Email)
		}

		// Verify YubiKey OTP
		if err := utils.VerifyYubikeyOTP(otp, Cfg.Yubikey); err != nil {
			return fmt.Errorf("OTP verification failed: %w", err)
		}

		// Find the device by OTP prefix
		devicePrefix := otp[:12]
		var device database.Device
		if err := DB.Where("serial_number LIKE ?", devicePrefix+"%").First(&device).Error; err != nil {
			return fmt.Errorf("device not found for OTP prefix: %s", devicePrefix)
		}

		if !device.Active {
			return fmt.Errorf("device %s is not active", device.Name)
		}

		// Log the authentication
		authLog := database.AuthenticationLog{
			ID:        uuid.New(),
			UserID:    &user.ID,
			DeviceID:  device.ID,
			OTP:       otp,
			Success:   true,
			Timestamp: time.Now(),
		}

		if err := DB.Create(&authLog).Error; err != nil {
			return fmt.Errorf("failed to log authentication: %w", err)
		}

		fmt.Printf("Authentication successful for user %s using device %s\n", user.Email, device.Name)
		fmt.Printf("Authentication log ID: %s\n", authLog.ID)
		return nil
	},
}

var authenticateDeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Authenticate a device with YubiKey OTP",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceIdentifier := args[0]
		otp := args[1]

		// Find the device
		var device database.Device
		if _, err := uuid.Parse(deviceIdentifier); err == nil {
			if err := DB.First(&device, "id = ?", deviceIdentifier).Error; err != nil {
				return fmt.Errorf("device not found: %w", err)
			}
		} else {
			if err := DB.First(&device, "name = ? OR serial_number = ?", deviceIdentifier, deviceIdentifier).Error; err != nil {
				return fmt.Errorf("device not found: %w", err)
			}
		}

		if !device.Active {
			return fmt.Errorf("device %s is not active", device.Name)
		}

		// Verify YubiKey OTP
		if err := utils.VerifyYubikeyOTP(otp, Cfg.Yubikey); err != nil {
			return fmt.Errorf("OTP verification failed: %w", err)
		}

		// Verify the OTP prefix matches the device
		devicePrefix := otp[:12]
		if device.SerialNumber != "" && device.SerialNumber != devicePrefix {
			return fmt.Errorf("OTP prefix %s does not match device serial number %s", devicePrefix, device.SerialNumber)
		}

		// Log the authentication
		authLog := database.AuthenticationLog{
			ID:        uuid.New(),
			DeviceID:  device.ID,
			OTP:       otp,
			Success:   true,
			Timestamp: time.Now(),
		}

		if err := DB.Create(&authLog).Error; err != nil {
			return fmt.Errorf("failed to log authentication: %w", err)
		}

		fmt.Printf("Device authentication successful for device %s\n", device.Name)
		fmt.Printf("Authentication log ID: %s\n", authLog.ID)
		return nil
	},
}

var listAuthLogsCmd = &cobra.Command{
	Use:   "list-logs",
	Short: "List authentication logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		deviceID, _ := cmd.Flags().GetString("device-id")
		success, _ := cmd.Flags().GetBool("success")
		limit, _ := cmd.Flags().GetInt("limit")

		query := DB.Preload("User").Preload("Device")

		// Apply filters
		if userID != "" {
			if _, err := uuid.Parse(userID); err != nil {
				return fmt.Errorf("invalid user ID: %w", err)
			}
			query = query.Where("user_id = ?", userID)
		}
		if deviceID != "" {
			if _, err := uuid.Parse(deviceID); err != nil {
				return fmt.Errorf("invalid device ID: %w", err)
			}
			query = query.Where("device_id = ?", deviceID)
		}
		if cmd.Flags().Changed("success") {
			query = query.Where("success = ?", success)
		}

		// Apply limit
		if limit > 0 {
			query = query.Limit(limit)
		}

		// Order by most recent first
		query = query.Order("timestamp DESC")

		var logs []database.AuthenticationLog
		if err := query.Find(&logs).Error; err != nil {
			return fmt.Errorf("failed to fetch authentication logs: %w", err)
		}

		fmt.Printf("Found %d authentication logs:\n\n", len(logs))
		for _, log := range logs {
			userEmail := "N/A"
			if log.User != nil {
				userEmail = log.User.Email
			}
			deviceName := "N/A"
			if log.Device.ID != uuid.Nil {
				deviceName = log.Device.Name
			}

			fmt.Printf("ID: %s\n  User: %s\n  Device: %s\n  OTP: %s\n  Success: %t\n  Timestamp: %s\n\n",
				log.ID, userEmail, deviceName, log.OTP, log.Success, log.Timestamp.Format(time.RFC3339))
		}
		return nil
	},
}

// AuthenticationCmd represents the authentication command
var AuthenticationCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Perform authentication operations",
	Long:  "Authenticate users and devices with YubiKey OTP",
}

// InitAuthenticationCommands initializes the authentication commands and their flags
func InitAuthenticationCommands() {
	// Add subcommands
	AuthenticationCmd.AddCommand(authenticateUserCmd)
	AuthenticationCmd.AddCommand(authenticateDeviceCmd)
	AuthenticationCmd.AddCommand(listAuthLogsCmd)

	// List auth logs flags
	listAuthLogsCmd.Flags().String("user-id", "", "Filter by user ID")
	listAuthLogsCmd.Flags().String("device-id", "", "Filter by device ID")
	listAuthLogsCmd.Flags().Bool("success", true, "Filter by success status")
	listAuthLogsCmd.Flags().Int("limit", 0, "Limit number of results")
} 