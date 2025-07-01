package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createRoleCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new role",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		active, _ := cmd.Flags().GetBool("active")

		role := database.Role{
			ID:          uuid.New(),
			Name:        name,
			Description: description,
			Active:      active,
		}

		if err := DB.Create(&role).Error; err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}

		fmt.Printf("Role created: %s (%s)\n", role.Name, role.ID)
		return nil
	},
}

var listRolesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all roles",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeOnly, _ := cmd.Flags().GetBool("active-only")

		var roles []database.Role
		query := DB.Preload("Permissions.Resource")
		if activeOnly {
			query = query.Where("active = ?", true)
		}

		if err := query.Find(&roles).Error; err != nil {
			return fmt.Errorf("failed to fetch roles: %w", err)
		}

		fmt.Printf("Found %d roles:\n\n", len(roles))
		for _, role := range roles {
			permissions := make([]string, len(role.Permissions))
			for i, perm := range role.Permissions {
				permissions[i] = fmt.Sprintf("%s:%s", perm.Resource.Name, perm.Action)
			}

			fmt.Printf("ID: %s\n  Name: %s\n  Description: %s\n  Active: %t\n  Permissions: %v\n  Created: %s\n  Updated: %s\n\n",
				role.ID, role.Name, role.Description, role.Active, permissions, role.CreatedAt.Format(time.RFC3339), role.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var updateRoleCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		active, _ := cmd.Flags().GetBool("active")

		var role database.Role
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&role, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := DB.First(&role, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		if name != "" {
			role.Name = name
		}
		if description != "" {
			role.Description = description
		}
		if cmd.Flags().Changed("active") {
			role.Active = active
		}

		if err := DB.Save(&role).Error; err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}

		fmt.Printf("Role updated: %s (%s)\n", role.Name, role.ID)
		return nil
	},
}

var deleteRoleCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var role database.Role
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&role, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := DB.First(&role, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		if err := DB.Delete(&role).Error; err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}

		fmt.Printf("Role deleted: %s (%s)\n", role.Name, role.ID)
		return nil
	},
}

// RoleCmd represents the role command
var RoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage roles",
	Long:  "Create, update, delete, and list roles",
}

// InitRoleCommands initializes the role commands and their flags
func InitRoleCommands() {
	// Add subcommands
	RoleCmd.AddCommand(createRoleCmd)
	RoleCmd.AddCommand(listRolesCmd)
	RoleCmd.AddCommand(updateRoleCmd)
	RoleCmd.AddCommand(deleteRoleCmd)

	// Create role flags
	createRoleCmd.Flags().String("name", "", "Role name")
	createRoleCmd.Flags().String("description", "", "Role description")
	createRoleCmd.Flags().Bool("active", true, "Whether the role is active")
	createRoleCmd.MarkFlagRequired("name")

	// Update role flags
	updateRoleCmd.Flags().String("name", "", "Role name")
	updateRoleCmd.Flags().String("description", "", "Role description")
	updateRoleCmd.Flags().Bool("active", true, "Whether the role is active")

	// List roles flags
	listRolesCmd.Flags().Bool("active-only", false, "Show only active roles")
} 