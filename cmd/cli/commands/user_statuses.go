package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createUserStatusCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user status",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		active, _ := cmd.Flags().GetBool("active")

		userStatus := database.UserStatus{
			ID:          uuid.New(),
			Name:        name,
			Description: description,
			Active:      active,
		}

		if err := DB.Create(&userStatus).Error; err != nil {
			return fmt.Errorf("failed to create user status: %w", err)
		}

		fmt.Printf("User status created: %s (%s)\n", userStatus.Name, userStatus.ID)
		return nil
	},
}

var listUserStatusesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all user statuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeOnly, _ := cmd.Flags().GetBool("active-only")

		var userStatuses []database.UserStatus
		query := DB
		if activeOnly {
			query = query.Where("active = ?", true)
		}

		if err := query.Find(&userStatuses).Error; err != nil {
			return fmt.Errorf("failed to fetch user statuses: %w", err)
		}

		fmt.Printf("Found %d user statuses:\n\n", len(userStatuses))
		for _, userStatus := range userStatuses {
			fmt.Printf("ID: %s\n  Name: %s\n  Description: %s\n  Active: %t\n  Created: %s\n  Updated: %s\n\n",
				userStatus.ID, userStatus.Name, userStatus.Description, userStatus.Active, userStatus.CreatedAt.Format(time.RFC3339), userStatus.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var updateUserStatusCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		active, _ := cmd.Flags().GetBool("active")

		var userStatus database.UserStatus
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&userStatus, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("user status not found: %w", err)
			}
		} else {
			if err := DB.First(&userStatus, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("user status not found: %w", err)
			}
		}

		if name != "" {
			userStatus.Name = name
		}
		if description != "" {
			userStatus.Description = description
		}
		if cmd.Flags().Changed("active") {
			userStatus.Active = active
		}

		if err := DB.Save(&userStatus).Error; err != nil {
			return fmt.Errorf("failed to update user status: %w", err)
		}

		fmt.Printf("User status updated: %s (%s)\n", userStatus.Name, userStatus.ID)
		return nil
	},
}

var deleteUserStatusCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var userStatus database.UserStatus
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&userStatus, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("user status not found: %w", err)
			}
		} else {
			if err := DB.First(&userStatus, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("user status not found: %w", err)
			}
		}

		if err := DB.Delete(&userStatus).Error; err != nil {
			return fmt.Errorf("failed to delete user status: %w", err)
		}

		fmt.Printf("User status deleted: %s (%s)\n", userStatus.Name, userStatus.ID)
		return nil
	},
}

// UserStatusCmd represents the user status command
var UserStatusCmd = &cobra.Command{
	Use:   "user-status",
	Short: "Manage user statuses",
	Long:  "Create, update, delete, and list user statuses",
}

// InitUserStatusCommands initializes the user status commands and their flags
func InitUserStatusCommands() {
	// Add subcommands
	UserStatusCmd.AddCommand(createUserStatusCmd)
	UserStatusCmd.AddCommand(listUserStatusesCmd)
	UserStatusCmd.AddCommand(updateUserStatusCmd)
	UserStatusCmd.AddCommand(deleteUserStatusCmd)

	// Create user status flags
	createUserStatusCmd.Flags().String("name", "", "User status name")
	createUserStatusCmd.Flags().String("description", "", "User status description")
	createUserStatusCmd.Flags().Bool("active", true, "Whether the user status is active")
	createUserStatusCmd.MarkFlagRequired("name")

	// Update user status flags
	updateUserStatusCmd.Flags().String("name", "", "User status name")
	updateUserStatusCmd.Flags().String("description", "", "User status description")
	updateUserStatusCmd.Flags().Bool("active", true, "Whether the user status is active")

	// List user statuses flags
	listUserStatusesCmd.Flags().Bool("active-only", false, "Show only active user statuses")
} 