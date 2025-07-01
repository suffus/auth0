package main

import (
	"log"
	"os"

	"github.com/YubiApp/cmd/cli/commands"
	"github.com/YubiApp/cmd/cli/utils"
	"github.com/YubiApp/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := utils.InitDatabase(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Set dependencies for all command packages
	commands.SetDependencies(db, cfg)

	// Initialize all command groups
	commands.InitUserCommands()
	commands.InitRoleCommands()
	commands.InitPermissionCommands()
	commands.InitResourceCommands()
	commands.InitDeviceCommands()
	commands.InitActionCommands()
	commands.InitLocationCommands()
	commands.InitUserStatusCommands()
	commands.InitUserActivityCommands()
	commands.InitAssignmentCommands()
	commands.InitAuthenticationCommands()

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "yubiapp-cli",
		Short: "YubiApp CLI - Authentication and Management System",
		Long: `YubiApp CLI is a command-line interface for managing users, roles, permissions,
resources, devices, actions, locations, user statuses, and authentication logs.

The CLI supports YubiKey OTP authentication and provides comprehensive management
capabilities for the YubiApp system.`,
	}

	// Add migration command
	rootCmd.AddCommand(commands.InitMigrationCommand())

	// Add all command groups to root
	rootCmd.AddCommand(commands.UserCmd)
	rootCmd.AddCommand(commands.RoleCmd)
	rootCmd.AddCommand(commands.PermissionCmd)
	rootCmd.AddCommand(commands.ResourceCmd)
	rootCmd.AddCommand(commands.DeviceCmd)
	rootCmd.AddCommand(commands.ActionCmd)
	rootCmd.AddCommand(commands.LocationCmd)
	rootCmd.AddCommand(commands.UserStatusCmd)
	rootCmd.AddCommand(commands.UserActivityCmd)
	rootCmd.AddCommand(commands.AssignmentCmd)
	rootCmd.AddCommand(commands.AuthenticationCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
} 