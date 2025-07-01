package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createPermissionCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new permission",
	RunE: func(cmd *cobra.Command, args []string) error {
		action, _ := cmd.Flags().GetString("action")
		resourceID, _ := cmd.Flags().GetString("resource-id")
		resourceName, _ := cmd.Flags().GetString("resource-name")

		// Find the resource
		var resource database.Resource
		if resourceID != "" {
			if _, err := uuid.Parse(resourceID); err != nil {
				return fmt.Errorf("invalid resource ID: %w", err)
			}
			if err := DB.First(&resource, "id = ?", resourceID).Error; err != nil {
				return fmt.Errorf("resource not found: %w", err)
			}
		} else if resourceName != "" {
			if err := DB.First(&resource, "name = ?", resourceName).Error; err != nil {
				return fmt.Errorf("resource not found: %w", err)
			}
		} else {
			return fmt.Errorf("either resource-id or resource-name must be provided")
		}

		permission := database.Permission{
			ID:         uuid.New(),
			Action:     action,
			ResourceID: resource.ID,
		}

		if err := DB.Create(&permission).Error; err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}

		fmt.Printf("Permission created: %s:%s (%s)\n", resource.Name, action, permission.ID)
		return nil
	},
}

var listPermissionsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		var permissions []database.Permission
		if err := DB.Preload("Resource").Find(&permissions).Error; err != nil {
			return fmt.Errorf("failed to fetch permissions: %w", err)
		}

		fmt.Printf("Found %d permissions:\n\n", len(permissions))
		for _, permission := range permissions {
			fmt.Printf("ID: %s\n  Action: %s\n  Resource: %s (%s)\n  Created: %s\n  Updated: %s\n\n",
				permission.ID, permission.Action, permission.Resource.Name, permission.ResourceID, permission.CreatedAt.Format(time.RFC3339), permission.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var deletePermissionCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a permission",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var permission database.Permission
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.Preload("Resource").First(&permission, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("permission not found: %w", err)
			}
		} else {
			// Try to parse as resource:action format
			if err := DB.Preload("Resource").First(&permission, "action = ? AND resource_id = (SELECT id FROM resources WHERE name = ?)", identifier, identifier).Error; err != nil {
				return fmt.Errorf("permission not found: %w", err)
			}
		}

		if err := DB.Delete(&permission).Error; err != nil {
			return fmt.Errorf("failed to delete permission: %w", err)
		}

		fmt.Printf("Permission deleted: %s:%s (%s)\n", permission.Resource.Name, permission.Action, permission.ID)
		return nil
	},
}

// PermissionCmd represents the permission command
var PermissionCmd = &cobra.Command{
	Use:   "permission",
	Short: "Manage permissions",
	Long:  "Create, delete, and list permissions",
}

// InitPermissionCommands initializes the permission commands and their flags
func InitPermissionCommands() {
	// Add subcommands
	PermissionCmd.AddCommand(createPermissionCmd)
	PermissionCmd.AddCommand(listPermissionsCmd)
	PermissionCmd.AddCommand(deletePermissionCmd)

	// Create permission flags
	createPermissionCmd.Flags().String("action", "", "Permission action (e.g., read, write, delete)")
	createPermissionCmd.Flags().String("resource-id", "", "Resource ID")
	createPermissionCmd.Flags().String("resource-name", "", "Resource name")
	createPermissionCmd.MarkFlagRequired("action")
} 