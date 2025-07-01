package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createResourceCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		active, _ := cmd.Flags().GetBool("active")

		// Check for colons in resource name
		if strings.Contains(name, ":") {
			return fmt.Errorf("resource name cannot contain colons")
		}

		resource := database.Resource{
			ID:     uuid.New(),
			Name:   name,
			Active: active,
		}

		if err := DB.Create(&resource).Error; err != nil {
			return fmt.Errorf("failed to create resource: %w", err)
		}

		fmt.Printf("Resource created: %s (%s)\n", resource.Name, resource.ID)
		return nil
	},
}

var listResourcesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeOnly, _ := cmd.Flags().GetBool("active-only")

		var resources []database.Resource
		query := DB.Preload("Permissions")
		if activeOnly {
			query = query.Where("active = ?", true)
		}

		if err := query.Find(&resources).Error; err != nil {
			return fmt.Errorf("failed to fetch resources: %w", err)
		}

		fmt.Printf("Found %d resources:\n\n", len(resources))
		for _, resource := range resources {
			fmt.Printf("ID: %s\n  Name: %s\n  Active: %t\n  Created: %s\n  Updated: %s\n\n",
				resource.ID, resource.Name, resource.Active, resource.CreatedAt.Format(time.RFC3339), resource.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var updateResourceCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		name, _ := cmd.Flags().GetString("name")
		active, _ := cmd.Flags().GetBool("active")

		var resource database.Resource
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&resource, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("resource not found: %w", err)
			}
		} else {
			if err := DB.First(&resource, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("resource not found: %w", err)
			}
		}

		if name != "" {
			// Check for colons in resource name
			if strings.Contains(name, ":") {
				return fmt.Errorf("resource name cannot contain colons")
			}
			resource.Name = name
		}
		if cmd.Flags().Changed("active") {
			resource.Active = active
		}

		if err := DB.Save(&resource).Error; err != nil {
			return fmt.Errorf("failed to update resource: %w", err)
		}

		fmt.Printf("Resource updated: %s (%s)\n", resource.Name, resource.ID)
		return nil
	},
}

var deleteResourceCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var resource database.Resource
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&resource, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("resource not found: %w", err)
			}
		} else {
			if err := DB.First(&resource, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("resource not found: %w", err)
			}
		}

		if err := DB.Delete(&resource).Error; err != nil {
			return fmt.Errorf("failed to delete resource: %w", err)
		}

		fmt.Printf("Resource deleted: %s (%s)\n", resource.Name, resource.ID)
		return nil
	},
}

// ResourceCmd represents the resource command
var ResourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage resources",
	Long:  "Create, update, delete, and list resources",
}

// InitResourceCommands initializes the resource commands and their flags
func InitResourceCommands() {
	// Add subcommands
	ResourceCmd.AddCommand(createResourceCmd)
	ResourceCmd.AddCommand(listResourcesCmd)
	ResourceCmd.AddCommand(updateResourceCmd)
	ResourceCmd.AddCommand(deleteResourceCmd)

	// Create resource flags
	createResourceCmd.Flags().String("name", "", "Resource name")
	createResourceCmd.Flags().Bool("active", true, "Whether the resource is active")
	createResourceCmd.MarkFlagRequired("name")

	// Update resource flags
	updateResourceCmd.Flags().String("name", "", "Resource name")
	updateResourceCmd.Flags().Bool("active", true, "Whether the resource is active")

	// List resources flags
	listResourcesCmd.Flags().Bool("active-only", false, "Show only active resources")
} 