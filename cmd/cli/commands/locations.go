package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var createLocationCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new location",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		address, _ := cmd.Flags().GetString("address")
		active, _ := cmd.Flags().GetBool("active")

		location := database.Location{
			ID:          uuid.New(),
			Name:        name,
			Description: description,
			Address:     address,
			Active:      active,
		}

		if err := DB.Create(&location).Error; err != nil {
			return fmt.Errorf("failed to create location: %w", err)
		}

		fmt.Printf("Location created: %s (%s)\n", location.Name, location.ID)
		return nil
	},
}

var listLocationsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all locations",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeOnly, _ := cmd.Flags().GetBool("active-only")

		var locations []database.Location
		query := DB
		if activeOnly {
			query = query.Where("active = ?", true)
		}

		if err := query.Find(&locations).Error; err != nil {
			return fmt.Errorf("failed to fetch locations: %w", err)
		}

		fmt.Printf("Found %d locations:\n\n", len(locations))
		for _, location := range locations {
			fmt.Printf("ID: %s\n  Name: %s\n  Description: %s\n  Address: %s\n  Active: %t\n  Created: %s\n  Updated: %s\n\n",
				location.ID, location.Name, location.Description, location.Address, location.Active, location.CreatedAt.Format(time.RFC3339), location.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var updateLocationCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		address, _ := cmd.Flags().GetString("address")
		active, _ := cmd.Flags().GetBool("active")

		var location database.Location
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&location, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("location not found: %w", err)
			}
		} else {
			if err := DB.First(&location, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("location not found: %w", err)
			}
		}

		if name != "" {
			location.Name = name
		}
		if description != "" {
			location.Description = description
		}
		if address != "" {
			location.Address = address
		}
		if cmd.Flags().Changed("active") {
			location.Active = active
		}

		if err := DB.Save(&location).Error; err != nil {
			return fmt.Errorf("failed to update location: %w", err)
		}

		fmt.Printf("Location updated: %s (%s)\n", location.Name, location.ID)
		return nil
	},
}

var deleteLocationCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a location",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var location database.Location
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&location, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("location not found: %w", err)
			}
		} else {
			if err := DB.First(&location, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("location not found: %w", err)
			}
		}

		if err := DB.Delete(&location).Error; err != nil {
			return fmt.Errorf("failed to delete location: %w", err)
		}

		fmt.Printf("Location deleted: %s (%s)\n", location.Name, location.ID)
		return nil
	},
}

// LocationCmd represents the location command
var LocationCmd = &cobra.Command{
	Use:   "location",
	Short: "Manage locations",
	Long:  "Create, update, delete, and list locations",
}

// InitLocationCommands initializes the location commands and their flags
func InitLocationCommands() {
	// Add subcommands
	LocationCmd.AddCommand(createLocationCmd)
	LocationCmd.AddCommand(listLocationsCmd)
	LocationCmd.AddCommand(updateLocationCmd)
	LocationCmd.AddCommand(deleteLocationCmd)

	// Create location flags
	createLocationCmd.Flags().String("name", "", "Location name")
	createLocationCmd.Flags().String("description", "", "Location description")
	createLocationCmd.Flags().String("address", "", "Location address")
	createLocationCmd.Flags().Bool("active", true, "Whether the location is active")
	createLocationCmd.MarkFlagRequired("name")

	// Update location flags
	updateLocationCmd.Flags().String("name", "", "Location name")
	updateLocationCmd.Flags().String("description", "", "Location description")
	updateLocationCmd.Flags().String("address", "", "Location address")
	updateLocationCmd.Flags().Bool("active", true, "Whether the location is active")

	// List locations flags
	listLocationsCmd.Flags().Bool("active-only", false, "Show only active locations")
} 