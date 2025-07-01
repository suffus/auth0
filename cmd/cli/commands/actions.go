package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/spf13/cobra"
)

var createActionCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new action",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		actionType, _ := cmd.Flags().GetString("type")
		details, _ := cmd.Flags().GetString("details")
		active, _ := cmd.Flags().GetBool("active")

		// Parse details JSON if provided
		var detailsJSON interface{}
		if details != "" {
			if err := json.Unmarshal([]byte(details), &detailsJSON); err != nil {
				return fmt.Errorf("invalid JSON in details: %w", err)
			}
		}

		// Convert details to JSONB
		var detailsMap map[string]interface{}
		if details != "" {
			if err := json.Unmarshal([]byte(details), &detailsMap); err != nil {
				return fmt.Errorf("invalid JSON in details: %w", err)
			}
		} else {
			detailsMap = make(map[string]interface{})
		}

		// Create JSONB from map
		var detailsJSONB pgtype.JSONB
		if err := detailsJSONB.Set(detailsMap); err != nil {
			return fmt.Errorf("failed to set details JSON: %w", err)
		}

		// Always set required_permissions to an empty array if not provided
		var requiredPermissionsJSONB pgtype.JSONB
		if err := requiredPermissionsJSONB.Set([]string{}); err != nil {
			return fmt.Errorf("failed to set required_permissions JSONB: %w", err)
		}

		action := database.Action{
			ID:                   uuid.New(),
			Name:                 name,
			ActivityType:         actionType,
			RequiredPermissions:  requiredPermissionsJSONB,
			Details:              detailsJSONB,
			Active:               active,
		}

		if err := DB.Create(&action).Error; err != nil {
			return fmt.Errorf("failed to create action: %w", err)
		}

		fmt.Printf("Action created: %s (%s)\n", action.Name, action.ID)
		return nil
	},
}

var listActionsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all actions",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeOnly, _ := cmd.Flags().GetBool("active-only")

		var actions []database.Action
		query := DB
		if activeOnly {
			query = query.Where("active = ?", true)
		}

		if err := query.Find(&actions).Error; err != nil {
			return fmt.Errorf("failed to fetch actions: %w", err)
		}

		fmt.Printf("Found %d actions:\n\n", len(actions))
		for _, action := range actions {
			detailsStr := "null"
			if action.Details.Status == pgtype.Present {
				if detailsBytes, err := json.Marshal(action.Details.Bytes); err == nil {
					detailsStr = string(detailsBytes)
				}
			}

			fmt.Printf("ID: %s\n  Name: %s\n  Type: %s\n  Active: %t\n  Details: %s\n  Created: %s\n  Updated: %s\n\n",
				action.ID, action.Name, action.ActivityType, action.Active, detailsStr, action.CreatedAt.Format(time.RFC3339), action.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var updateActionCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an action",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		name, _ := cmd.Flags().GetString("name")
		actionType, _ := cmd.Flags().GetString("type")
		details, _ := cmd.Flags().GetString("details")
		active, _ := cmd.Flags().GetBool("active")

		var action database.Action
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&action, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("action not found: %w", err)
			}
		} else {
			if err := DB.First(&action, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("action not found: %w", err)
			}
		}

		if name != "" {
			action.Name = name
		}
		if actionType != "" {
			action.ActivityType = actionType
		}
		if details != "" {
			var detailsJSON interface{}
			if err := json.Unmarshal([]byte(details), &detailsJSON); err != nil {
				return fmt.Errorf("invalid JSON in details: %w", err)
			}
			if err := action.Details.Set(detailsJSON); err != nil {
				return fmt.Errorf("failed to set details JSON: %w", err)
			}
		}
		if cmd.Flags().Changed("active") {
			action.Active = active
		}

		if err := DB.Save(&action).Error; err != nil {
			return fmt.Errorf("failed to update action: %w", err)
		}

		fmt.Printf("Action updated: %s (%s)\n", action.Name, action.ID)
		return nil
	},
}

var deleteActionCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an action",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var action database.Action
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&action, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("action not found: %w", err)
			}
		} else {
			if err := DB.First(&action, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("action not found: %w", err)
			}
		}

		if err := DB.Delete(&action).Error; err != nil {
			return fmt.Errorf("failed to delete action: %w", err)
		}

		fmt.Printf("Action deleted: %s (%s)\n", action.Name, action.ID)
		return nil
	},
}

var executeActionCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute an action",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var action database.Action
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&action, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("action not found: %w", err)
			}
		} else {
			if err := DB.First(&action, "name = ?", identifier).Error; err != nil {
				return fmt.Errorf("action not found: %w", err)
			}
		}

		if !action.Active {
			return fmt.Errorf("cannot execute inactive action: %s", action.Name)
		}

		fmt.Printf("Executing action: %s (%s)\n", action.Name, action.ID)
		fmt.Printf("Type: %s\n", action.ActivityType)
		if action.Details.Status == pgtype.Present {
			if detailsBytes, err := json.MarshalIndent(action.Details.Bytes, "", "  "); err == nil {
				fmt.Printf("Details: %s\n", string(detailsBytes))
			}
		}

		// Here you would implement the actual action execution logic
		fmt.Println("Action executed successfully!")
		return nil
	},
}

// ActionCmd represents the action command
var ActionCmd = &cobra.Command{
	Use:   "action",
	Short: "Manage actions",
	Long:  "Create, update, delete, list, and execute actions",
}

// InitActionCommands initializes the action commands and their flags
func InitActionCommands() {
	// Add subcommands
	ActionCmd.AddCommand(createActionCmd)
	ActionCmd.AddCommand(listActionsCmd)
	ActionCmd.AddCommand(updateActionCmd)
	ActionCmd.AddCommand(deleteActionCmd)
	ActionCmd.AddCommand(executeActionCmd)

	// Create action flags
	createActionCmd.Flags().String("name", "", "Action name")
	createActionCmd.Flags().String("description", "", "Action description")
	createActionCmd.Flags().String("type", "user", "Action type (user, system, automated, other)")
	createActionCmd.Flags().String("details", "", "Action details (JSON format)")
	createActionCmd.Flags().Bool("active", true, "Whether the action is active")
	createActionCmd.MarkFlagRequired("name")

	// Update action flags
	updateActionCmd.Flags().String("name", "", "Action name")
	updateActionCmd.Flags().String("description", "", "Action description")
	updateActionCmd.Flags().String("type", "", "Action type (user, system, automated, other)")
	updateActionCmd.Flags().String("details", "", "Action details (JSON format)")
	updateActionCmd.Flags().Bool("active", true, "Whether the action is active")

	// List actions flags
	listActionsCmd.Flags().Bool("active-only", false, "Show only active actions")
} 