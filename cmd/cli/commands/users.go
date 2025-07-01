package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

// UserCmd represents the user command
var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  "Create, update, delete, and list users",
}

var createUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		active, _ := cmd.Flags().GetBool("active")

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		user := database.User{
			ID:        uuid.New(),
			Email:     email,
			Username:  username,
			Password:  string(hashedPassword),
			FirstName: firstName,
			LastName:  lastName,
			Active:    active,
		}

		if err := DB.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		fmt.Printf("User created: %s (%s)\n", user.Email, user.ID)
		return nil
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	RunE: func(cmd *cobra.Command, args []string) error {
		activeOnly, _ := cmd.Flags().GetBool("active-only")

		var users []database.User
		query := DB.Preload("Roles")
		if activeOnly {
			query = query.Where("active = ?", true)
		}

		if err := query.Find(&users).Error; err != nil {
			return fmt.Errorf("failed to fetch users: %w", err)
		}

		fmt.Printf("Found %d users:\n\n", len(users))
		for _, user := range users {
			roles := make([]string, len(user.Roles))
			for i, role := range user.Roles {
				roles[i] = role.Name
			}

			fmt.Printf("ID: %s\n  Email: %s\n  Username: %s\n  Name: %s %s\n  Active: %t\n  Roles: %v\n  Created: %s\n  Updated: %s\n\n",
				user.ID, user.Email, user.Username, user.FirstName, user.LastName, user.Active, roles, user.CreatedAt.Format(time.RFC3339), user.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var updateUserCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		email, _ := cmd.Flags().GetString("email")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		active, _ := cmd.Flags().GetBool("active")

		var user database.User
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&user, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("user not found: %w", err)
			}
		} else {
			if err := DB.First(&user, "email = ? OR username = ?", identifier, identifier).Error; err != nil {
				return fmt.Errorf("user not found: %w", err)
			}
		}

		if email != "" {
			user.Email = email
		}
		if username != "" {
			user.Username = username
		}
		if password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash password: %w", err)
			}
			user.Password = string(hashedPassword)
		}
		if firstName != "" {
			user.FirstName = firstName
		}
		if lastName != "" {
			user.LastName = lastName
		}
		if cmd.Flags().Changed("active") {
			user.Active = active
		}

		if err := DB.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		fmt.Printf("User updated: %s (%s)\n", user.Email, user.ID)
		return nil
	},
}

var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		identifier := args[0]
		var user database.User
		if _, err := uuid.Parse(identifier); err == nil {
			if err := DB.First(&user, "id = ?", identifier).Error; err != nil {
				return fmt.Errorf("user not found: %w", err)
			}
		} else {
			if err := DB.First(&user, "email = ? OR username = ?", identifier, identifier).Error; err != nil {
				return fmt.Errorf("user not found: %w", err)
			}
		}

		if err := DB.Delete(&user).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		fmt.Printf("User deleted: %s (%s)\n", user.Email, user.ID)
		return nil
	},
}

// InitUserCommands initializes the user commands and their flags
func InitUserCommands() {
	// Add subcommands
	UserCmd.AddCommand(createUserCmd)
	UserCmd.AddCommand(listUsersCmd)
	UserCmd.AddCommand(updateUserCmd)
	UserCmd.AddCommand(deleteUserCmd)

	// Create user flags
	createUserCmd.Flags().String("email", "", "User email address")
	createUserCmd.Flags().String("username", "", "Username")
	createUserCmd.Flags().String("password", "", "Password")
	createUserCmd.Flags().String("first-name", "", "First name")
	createUserCmd.Flags().String("last-name", "", "Last name")
	createUserCmd.Flags().Bool("active", true, "Whether the user is active")
	createUserCmd.MarkFlagRequired("email")
	createUserCmd.MarkFlagRequired("username")
	createUserCmd.MarkFlagRequired("password")

	// Update user flags
	updateUserCmd.Flags().String("email", "", "User email address")
	updateUserCmd.Flags().String("username", "", "Username")
	updateUserCmd.Flags().String("password", "", "Password")
	updateUserCmd.Flags().String("first-name", "", "First name")
	updateUserCmd.Flags().String("last-name", "", "Last name")
	updateUserCmd.Flags().Bool("active", true, "Whether the user is active")

	// List users flags
	listUsersCmd.Flags().Bool("active-only", false, "Show only active users")
} 