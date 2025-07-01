package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createDeviceCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new device",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		deviceType, _ := cmd.Flags().GetString("type")
		serialNumber, _ := cmd.Flags().GetString("serial-number")
		active, _ := cmd.Flags().GetBool("active")

		device := database.Device{
			ID:           uuid.New(),
			Name:         name,
			Type:         deviceType,
			SerialNumber: serialNumber,
			Active:       active,
		}

		if err := DB.Create(&device).Error; err != nil {
			return fmt.Errorf("failed to create device: %w", err)
		}

		fmt.Printf("Device created: %s (%s)\n", device.Name, device.ID)
		return nil
	},
}

var listDevicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeOnly, _ := cmd.Flags().GetBool("active-only")

		var devices []database.Device
		query := DB
		if activeOnly {
			query = query.Where("active = ?", true)
		}

		if err := query.Find(&devices).Error; err != nil {
			return fmt.Errorf("failed to fetch devices: %w", err)
		}

		fmt.Printf("Found %d devices:\n\n", len(devices))
		for _, device := range devices {
			fmt.Printf("ID: %s\n  Name: %s\n  Type: %s\n  Serial Number: %s\n  Active: %t\n  Created: %s\n  Updated: %s\n\n",
				device.ID, device.Name, device.Type, device.SerialNumber, device.Active, device.CreatedAt.Format(time.RFC3339), device.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var updateDeviceCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		name, _ := cmd.Flags().GetString("name")
		deviceType, _ := cmd.Flags().GetString("type")
		serialNumber, _ := cmd.Flags().GetString("serial-number")
		active, _ := cmd.Flags().GetBool("active")

		var device database.Device
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&device, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("device not found: %w", err)
			}
		} else {
			if err := DB.First(&device, "name = ? OR serial_number = ?", identifier, identifier).Error; err != nil {
				return fmt.Errorf("device not found: %w", err)
			}
		}

		if name != "" {
			device.Name = name
		}
		if deviceType != "" {
			device.Type = deviceType
		}
		if serialNumber != "" {
			device.SerialNumber = serialNumber
		}
		if cmd.Flags().Changed("active") {
			device.Active = active
		}

		if err := DB.Save(&device).Error; err != nil {
			return fmt.Errorf("failed to update device: %w", err)
		}

		fmt.Printf("Device updated: %s (%s)\n", device.Name, device.ID)
		return nil
	},
}

var deleteDeviceCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a device",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var device database.Device
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&device, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("device not found: %w", err)
			}
		} else {
			if err := DB.First(&device, "name = ? OR serial_number = ?", identifier, identifier).Error; err != nil {
				return fmt.Errorf("device not found: %w", err)
			}
		}

		if err := DB.Delete(&device).Error; err != nil {
			return fmt.Errorf("failed to delete device: %w", err)
		}

		fmt.Printf("Device deleted: %s (%s)\n", device.Name, device.ID)
		return nil
	},
}

// DeviceCmd represents the device command
var DeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Manage devices",
	Long:  "Create, update, delete, and list devices",
}

// InitDeviceCommands initializes the device commands and their flags
func InitDeviceCommands() {
	// Add subcommands
	DeviceCmd.AddCommand(createDeviceCmd)
	DeviceCmd.AddCommand(listDevicesCmd)
	DeviceCmd.AddCommand(updateDeviceCmd)
	DeviceCmd.AddCommand(deleteDeviceCmd)

	// Create device flags
	createDeviceCmd.Flags().String("name", "", "Device name")
	createDeviceCmd.Flags().String("type", "", "Device type")
	createDeviceCmd.Flags().String("serial-number", "", "Device serial number")
	createDeviceCmd.Flags().Bool("active", true, "Whether the device is active")
	createDeviceCmd.MarkFlagRequired("name")

	// Update device flags
	updateDeviceCmd.Flags().String("name", "", "Device name")
	updateDeviceCmd.Flags().String("type", "", "Device type")
	updateDeviceCmd.Flags().String("serial-number", "", "Device serial number")
	updateDeviceCmd.Flags().Bool("active", true, "Whether the device is active")

	// List devices flags
	listDevicesCmd.Flags().Bool("active-only", false, "Show only active devices")
} 