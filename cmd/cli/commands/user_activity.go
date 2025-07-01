package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/jackc/pgtype"
)

var listUserActivityCmd = &cobra.Command{
	Use:   "list",
	Short: "List user activity history",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		userEmail, _ := cmd.Flags().GetString("user-email")
		actionID, _ := cmd.Flags().GetString("action-id")
		locationID, _ := cmd.Flags().GetString("location-id")
		userStatusID, _ := cmd.Flags().GetString("user-status-id")
		fromDate, _ := cmd.Flags().GetString("from-date")
		toDate, _ := cmd.Flags().GetString("to-date")
		limit, _ := cmd.Flags().GetInt("limit")

		query := DB.Preload("User").Preload("Action").Preload("Location").Preload("UserStatus")

		// Apply filters
		if userID != "" {
			if _, err := uuid.Parse(userID); err != nil {
				return fmt.Errorf("invalid user ID: %w", err)
			}
			query = query.Where("user_id = ?", userID)
		}
		if userEmail != "" {
			query = query.Joins("JOIN users ON user_activity_history.user_id = users.id").Where("users.email = ?", userEmail)
		}
		if actionID != "" {
			if _, err := uuid.Parse(actionID); err != nil {
				return fmt.Errorf("invalid action ID: %w", err)
			}
			query = query.Where("action_id = ?", actionID)
		}
		if locationID != "" {
			if _, err := uuid.Parse(locationID); err != nil {
				return fmt.Errorf("invalid location ID: %w", err)
			}
			query = query.Where("location_id = ?", locationID)
		}
		if userStatusID != "" {
			if _, err := uuid.Parse(userStatusID); err != nil {
				return fmt.Errorf("invalid user status ID: %w", err)
			}
			query = query.Where("user_status_id = ?", userStatusID)
		}
		if fromDate != "" {
			fromTime, err := time.Parse("2006-01-02", fromDate)
			if err != nil {
				return fmt.Errorf("invalid from date format (use YYYY-MM-DD): %w", err)
			}
			query = query.Where("from_date_time >= ?", fromTime)
		}
		if toDate != "" {
			toTime, err := time.Parse("2006-01-02", toDate)
			if err != nil {
				return fmt.Errorf("invalid to date format (use YYYY-MM-DD): %w", err)
			}
			// Add one day to include the entire day
			toTime = toTime.Add(24 * time.Hour)
			query = query.Where("from_date_time < ?", toTime)
		}

		// Apply limit
		if limit > 0 {
			query = query.Limit(limit)
		}

		// Order by most recent first
		query = query.Order("from_date_time DESC")

		var activities []database.UserActivityHistory
		if err := query.Find(&activities).Error; err != nil {
			return fmt.Errorf("failed to fetch user activity: %w", err)
		}

		fmt.Printf("Found %d activity records:\n\n", len(activities))
		for _, activity := range activities {
			detailsStr := "null"
			if activity.Details.Status == pgtype.Present {
				if detailsBytes, err := json.Marshal(activity.Details.Bytes); err == nil {
					detailsStr = string(detailsBytes)
				}
			}

			fmt.Printf("ID: %s\n  User: %s (%s)\n  Action: %s (%s)\n  User Status: %s (%s)\n  Location: %s (%s)\n  From: %s\n  To: %s\n  Details: %s\n  Created: %s\n\n",
				activity.ID,
				activity.User.Email, activity.UserID,
				activity.Action.Name, activity.ActionID,
				activity.Status.Name, activity.StatusID,
				activity.Location.Name, activity.LocationID,
				activity.FromDateTime.Format(time.RFC3339),
				formatTime(activity.ToDateTime),
				detailsStr,
				activity.CreatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var getUserActivityCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific user activity record",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		activityID := args[0]
		if _, err := uuid.Parse(activityID); err != nil {
			return fmt.Errorf("invalid activity ID: %w", err)
		}

		var activity database.UserActivityHistory
		if err := DB.Preload("User").Preload("Action").Preload("Location").Preload("UserStatus").First(&activity, "id = ?", activityID).Error; err != nil {
			return fmt.Errorf("activity not found: %w", err)
		}

		detailsStr := "null"
		if activity.Details.Status == pgtype.Present {
			if detailsBytes, err := json.MarshalIndent(activity.Details.Bytes, "", "  "); err == nil {
				detailsStr = string(detailsBytes)
			}
		}

		fmt.Printf("Activity ID: %s\n", activity.ID)
		fmt.Printf("User: %s (%s)\n", activity.User.Email, activity.UserID)
		fmt.Printf("Action: %s (%s)\n", activity.Action.Name, activity.ActionID)
		fmt.Printf("User Status: %s (%s)\n", activity.Status.Name, activity.StatusID)
		fmt.Printf("Location: %s (%s)\n", activity.Location.Name, activity.LocationID)
		fmt.Printf("From: %s\n", activity.FromDateTime.Format(time.RFC3339))
		fmt.Printf("To: %s\n", formatTime(activity.ToDateTime))
		fmt.Printf("Details: %s\n", detailsStr)
		fmt.Printf("Created: %s\n", activity.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated: %s\n", activity.UpdatedAt.Format(time.RFC3339))

		return nil
	},
}

// Helper function to format time, handling nil values
func formatTime(t *time.Time) string {
	if t == nil {
		return "null"
	}
	return t.Format(time.RFC3339)
}

// UserActivityCmd represents the user activity command
var UserActivityCmd = &cobra.Command{
	Use:   "user-activity",
	Short: "Query user activity history",
	Long:  "List and get user activity history records",
}

// InitUserActivityCommands initializes the user activity commands and their flags
func InitUserActivityCommands() {
	// Add subcommands
	UserActivityCmd.AddCommand(listUserActivityCmd)
	UserActivityCmd.AddCommand(getUserActivityCmd)

	// List user activity flags
	listUserActivityCmd.Flags().String("user-id", "", "Filter by user ID")
	listUserActivityCmd.Flags().String("user-email", "", "Filter by user email")
	listUserActivityCmd.Flags().String("action-id", "", "Filter by action ID")
	listUserActivityCmd.Flags().String("location-id", "", "Filter by location ID")
	listUserActivityCmd.Flags().String("user-status-id", "", "Filter by user status ID")
	listUserActivityCmd.Flags().String("from-date", "", "Filter from date (YYYY-MM-DD)")
	listUserActivityCmd.Flags().String("to-date", "", "Filter to date (YYYY-MM-DD)")
	listUserActivityCmd.Flags().Int("limit", 0, "Limit number of results")
} 