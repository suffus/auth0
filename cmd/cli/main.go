package main

import (
	"crypto/rand"
	"encoding/hex"
	//"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/YubiApp/internal/config"
	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	cfg *config.Config
)

// verifyYubikeyOTP verifies the OTP with Yubico servers
func verifyYubikeyOTP(otp string, config config.YubikeyConfig) error {
	params := url.Values{}
	params.Add("id", config.ClientID)
	params.Add("otp", otp)
	
	// Generate alphanumeric nonce (16-40 characters, no hyphens)
	nonceBytes := make([]byte, 20)
	rand.Read(nonceBytes)
	nonce := hex.EncodeToString(nonceBytes)
	params.Add("nonce", nonce)

	resp, err := http.Get(fmt.Sprintf("%s?%s", config.APIURL, params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to verify OTP with Yubico: %w", err)
	}
	defer resp.Body.Close()

	// Read the response as plain text
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Yubico response: %w", err)
	}

	// Parse key-value pairs
	lines := strings.Split(string(body), "\n")
	status := ""
	for _, line := range lines {
		if strings.HasPrefix(line, "status=") {
			status = strings.TrimSpace(strings.TrimPrefix(line, "status="))
			break
		}
	}

	switch strings.ToLower(status) {
	case "ok":
		return nil
	case "replayed_otp":
		return fmt.Errorf("replayed OTP detected")
	case "bad_otp":
		return fmt.Errorf("invalid OTP format")
	case "missing_parameter":
		return fmt.Errorf("missing parameter in OTP verification")
	case "no_such_client":
		return fmt.Errorf("invalid client ID")
	case "operation_not_allowed":
		return fmt.Errorf("operation not allowed")
	case "backend_error":
		return fmt.Errorf("Yubico backend error")
	default:
		return fmt.Errorf("Yubico verification failed with status: %s", status)
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "yubiapp-cli",
		Short: "YubiApp Command Line Management Tool",
		Long: `A comprehensive command-line tool for managing YubiApp users, roles, permissions, and devices.`,
		SilenceUsage:  true,  // Don't show usage for runtime errors
		SilenceErrors: true,  // Don't show error messages (we'll handle them ourselves)
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		db, err = initDatabase(cfg.Database)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		return nil
	}

	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(roleCmd)
	rootCmd.AddCommand(permissionCmd)
	rootCmd.AddCommand(resourceCmd)
	rootCmd.AddCommand(deviceCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(authenticateCmd)

	if err := rootCmd.Execute(); err != nil {
		// Check if this is a usage error by looking at the error message
		errMsg := err.Error()
		isUsageError := strings.Contains(errMsg, "required") || 
		               strings.Contains(errMsg, "invalid") ||
		               strings.Contains(errMsg, "unknown") ||
		               strings.Contains(errMsg, "not found") ||
		               strings.Contains(errMsg, "accepts") ||
		               strings.Contains(errMsg, "arguments")

		if isUsageError {
			// For usage errors, show the error and usage
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			rootCmd.Usage()
		} else {
			// For runtime errors, just show the error
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}

func initDatabase(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.AutoMigrate(
		&database.User{},
		&database.Role{},
		&database.Resource{},
		&database.Permission{},
		&database.Device{},
		&database.Session{},
		&database.AuthenticationLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// FindUserByString searches for a user by ID, username, or email
// Returns the user if exactly one match is found, otherwise returns an error
func FindUserByString(db *gorm.DB, identifier string) (*database.User, error) {
	var users []database.User
	
	// Try to parse as UUID first
	if _, err := uuid.Parse(identifier); err == nil {
		// Search by ID
		if err := db.Where("id = ?", identifier).Find(&users).Error; err != nil {
			return nil, fmt.Errorf("failed to search by ID: %w", err)
		}
	} else {
		// Search by username or email
		if err := db.Where("username = ? OR email = ?", identifier, identifier).Find(&users).Error; err != nil {
			return nil, fmt.Errorf("failed to search by username/email: %w", err)
		}
	}
	
	if len(users) == 0 {
		return nil, fmt.Errorf("no user found matching '%s'", identifier)
	}
	
	if len(users) > 1 {
		// Build list of matching users for error message
		matches := make([]string, len(users))
		for i, user := range users {
			matches[i] = fmt.Sprintf("%s (%s)", user.Email, user.ID)
		}
		return nil, fmt.Errorf("multiple users found matching '%s': %s", identifier, strings.Join(matches, ", "))
	}
	
	return &users[0], nil
}

// FindResourceByString searches for a resource by ID or name
// Returns the resource if exactly one match is found, otherwise returns an error
func FindResourceByString(db *gorm.DB, identifier string) (*database.Resource, error) {
	var resources []database.Resource
	
	// Try to parse as UUID first
	if _, err := uuid.Parse(identifier); err == nil {
		// Search by ID
		if err := db.Where("id = ?", identifier).Find(&resources).Error; err != nil {
			return nil, fmt.Errorf("failed to search by ID: %w", err)
		}
	} else {
		// Search by name
		if err := db.Where("name = ?", identifier).Find(&resources).Error; err != nil {
			return nil, fmt.Errorf("failed to search by name: %w", err)
		}
	}
	
	if len(resources) == 0 {
		return nil, fmt.Errorf("no resource found matching '%s'", identifier)
	}
	
	if len(resources) > 1 {
		// Build list of matching resources for error message
		matches := make([]string, len(resources))
		for i, resource := range resources {
			matches[i] = fmt.Sprintf("%s (%s)", resource.Name, resource.ID)
		}
		return nil, fmt.Errorf("multiple resources found matching '%s': %s", identifier, strings.Join(matches, ", "))
	}
	
	return &resources[0], nil
}

// FindPermissionByString searches for a permission by ID or resource:action format
// Returns the permission if exactly one match is found, otherwise returns an error
func FindPermissionByString(db *gorm.DB, identifier string) (*database.Permission, error) {
	var permissions []database.Permission
	
	// Try to parse as UUID first
	if _, err := uuid.Parse(identifier); err == nil {
		// Search by ID
		if err := db.Preload("Resource").Where("id = ?", identifier).Find(&permissions).Error; err != nil {
			return nil, fmt.Errorf("failed to search by ID: %w", err)
		}
	} else {
		// Check if it's in resource:action format
		if strings.Contains(identifier, ":") {
			parts := strings.SplitN(identifier, ":", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid permission format '%s', expected 'resource:action'", identifier)
			}
			resourceName := parts[0]
			action := parts[1]
			
			// Search by resource name and action
			if err := db.Preload("Resource").Joins("JOIN resources ON permissions.resource_id = resources.id").
				Where("resources.name = ? AND permissions.action = ?", resourceName, action).Find(&permissions).Error; err != nil {
				return nil, fmt.Errorf("failed to search by resource:action: %w", err)
			}
		} else {
			return nil, fmt.Errorf("invalid permission identifier '%s', expected UUID or 'resource:action' format", identifier)
		}
	}
	
	if len(permissions) == 0 {
		return nil, fmt.Errorf("no permission found matching '%s'", identifier)
	}
	
	if len(permissions) > 1 {
		// Build list of matching permissions for error message
		matches := make([]string, len(permissions))
		for i, perm := range permissions {
			matches[i] = fmt.Sprintf("%s:%s (%s)", perm.Resource.Name, perm.Action, perm.ID)
		}
		return nil, fmt.Errorf("multiple permissions found matching '%s': %s", identifier, strings.Join(matches, ", "))
	}
	
	return &permissions[0], nil
}

// User management commands
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  "Create, update, delete, and list users",
}

var createUserCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Long:  "Create a new user with the specified details",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		firstName, _ := cmd.Flags().GetString("first-name")
		lastName, _ := cmd.Flags().GetString("last-name")
		active, _ := cmd.Flags().GetBool("active")

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

		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		fmt.Printf("User created successfully:\n")
		fmt.Printf("  ID: %s\n", user.ID)
		fmt.Printf("  Email: %s\n", user.Email)
		fmt.Printf("  Username: %s\n", user.Username)
		fmt.Printf("  Name: %s %s\n", user.FirstName, user.LastName)
		fmt.Printf("  Active: %t\n", user.Active)

		return nil
	},
}

var listUsersCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Long:  "Display all users in the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		var users []database.User
		if err := db.Preload("Roles").Find(&users).Error; err != nil {
			return fmt.Errorf("failed to fetch users: %w", err)
		}

		fmt.Printf("Found %d users:\n\n", len(users))
		for _, user := range users {
			fmt.Printf("ID: %s\n", user.ID)
			fmt.Printf("  Email: %s\n", user.Email)
			fmt.Printf("  Username: %s\n", user.Username)
			fmt.Printf("  Name: %s %s\n", user.FirstName, user.LastName)
			fmt.Printf("  Active: %t\n", user.Active)
			
			if len(user.Roles) > 0 {
				roleNames := make([]string, len(user.Roles))
				for i, role := range user.Roles {
					roleNames[i] = role.Name
				}
				fmt.Printf("  Roles: %s\n", strings.Join(roleNames, ", "))
			}
			fmt.Println()
		}

		return nil
	},
}

var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a user",
	Long:  "Delete a user by ID or email",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a user ID or email")
		}

		identifier := args[0]
		user, err := FindUserByString(db, identifier)
		if err != nil {
			return err
		}

		if err := db.Delete(&user).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		fmt.Printf("User deleted successfully: %s (%s)\n", user.Email, user.ID)
		return nil
	},
}

// Role management commands
var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Manage roles",
	Long:  "Create, update, delete, and list roles",
}

var createRoleCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new role",
	Long:  "Create a new role with the specified details",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		role := database.Role{
			ID:          uuid.New(),
			Name:        name,
			Description: description,
		}

		if err := db.Create(&role).Error; err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}

		fmt.Printf("Role created successfully:\n")
		fmt.Printf("  ID: %s\n", role.ID)
		fmt.Printf("  Name: %s\n", role.Name)
		fmt.Printf("  Description: %s\n", role.Description)

		return nil
	},
}

var listRolesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all roles",
	Long:  "Display all roles in the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		var roles []database.Role
		if err := db.Preload("Permissions.Resource").Find(&roles).Error; err != nil {
			return fmt.Errorf("failed to fetch roles: %w", err)
		}

		fmt.Printf("Found %d roles:\n\n", len(roles))
		for _, role := range roles {
			fmt.Printf("ID: %s\n", role.ID)
			fmt.Printf("  Name: %s\n", role.Name)
			fmt.Printf("  Description: %s\n", role.Description)
			
			if len(role.Permissions) > 0 {
				fmt.Printf("  Permissions:\n")
				for _, perm := range role.Permissions {
					fmt.Printf("    - %s:%s (%s)\n", perm.Resource.Name, perm.Action, perm.Effect)
				}
			}
			fmt.Println()
		}

		return nil
	},
}

var deleteRoleCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a role",
	Long:  "Delete a role by ID or name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a role ID or name")
		}

		identifier := args[0]
		var role database.Role

		if _, err := uuid.Parse(identifier); err == nil {
			if err := db.Where("id = ?", identifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := db.Where("name = ?", identifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		if err := db.Delete(&role).Error; err != nil {
			return fmt.Errorf("failed to delete role: %w", err)
		}

		fmt.Printf("Role deleted successfully: %s (%s)\n", role.Name, role.ID)
		return nil
	},
}

// Permission management commands
var permissionCmd = &cobra.Command{
	Use:   "permission",
	Short: "Manage permissions",
	Long:  "Create, update, delete, and list permissions",
}

var createPermissionCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new permission",
	Long:  "Create a new permission with the specified details",
	RunE: func(cmd *cobra.Command, args []string) error {
		resourceID, _ := cmd.Flags().GetString("resource-id")
		action, _ := cmd.Flags().GetString("action")
		effect, _ := cmd.Flags().GetString("effect")

		if effect != "allow" && effect != "deny" {
			return fmt.Errorf("effect must be 'allow' or 'deny'")
		}

		// Find the resource
		resource, err := FindResourceByString(db, resourceID)
		if err != nil {
			return err
		}

		permission := database.Permission{
			ID:         uuid.New(),
			ResourceID: resource.ID,
			Action:     action,
			Effect:     effect,
		}

		if err := db.Create(&permission).Error; err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}

		fmt.Printf("Permission created successfully:\n")
		fmt.Printf("  ID: %s\n", permission.ID)
		fmt.Printf("  Resource: %s (%s)\n", resource.Name, permission.ResourceID)
		fmt.Printf("  Action: %s\n", permission.Action)
		fmt.Printf("  Effect: %s\n", permission.Effect)

		return nil
	},
}

var listPermissionsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all permissions",
	Long:  "Display all permissions in the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		var permissions []database.Permission
		if err := db.Preload("Resource").Find(&permissions).Error; err != nil {
			return fmt.Errorf("failed to fetch permissions: %w", err)
		}

		fmt.Printf("Found %d permissions:\n\n", len(permissions))
		for _, perm := range permissions {
			fmt.Printf("ID: %s\n", perm.ID)
			fmt.Printf("  Resource: %s (%s)\n", perm.Resource.Name, perm.ResourceID)
			fmt.Printf("  Action: %s\n", perm.Action)
			fmt.Printf("  Effect: %s\n", perm.Effect)
			fmt.Println()
		}

		return nil
	},
}

var deletePermissionCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a permission",
	Long:  "Delete a permission by ID or resource:action format (e.g., 'yubiapp:read')",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a permission ID or resource:action format")
		}

		identifier := args[0]
		permission, err := FindPermissionByString(db, identifier)
		if err != nil {
			return err
		}

		if err := db.Delete(&permission).Error; err != nil {
			return fmt.Errorf("failed to delete permission: %w", err)
		}

		fmt.Printf("Permission deleted successfully: %s:%s (%s)\n", 
			permission.Resource.Name, permission.Action, permission.Effect)
		return nil
	},
}

// Resource management commands
var resourceCmd = &cobra.Command{
	Use:   "resource",
	Short: "Manage resources",
	Long:  "Create, update, delete, and list resources",
}

var createResourceCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new resource",
	Long:  "Create a new resource with the specified details",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		resourceType, _ := cmd.Flags().GetString("type")
		location, _ := cmd.Flags().GetString("location")
		department, _ := cmd.Flags().GetString("department")
		active, _ := cmd.Flags().GetBool("active")

		// Validate resource type
		validTypes := []string{"server", "service", "database", "application"}
		validType := false
		for _, t := range validTypes {
			if resourceType == t {
				validType = true
				break
			}
		}
		if !validType {
			return fmt.Errorf("resource type must be one of: %s", strings.Join(validTypes, ", "))
		}

		resource := database.Resource{
			ID:         uuid.New(),
			Name:       name,
			Type:       resourceType,
			Location:   location,
			Department: department,
			Active:     active,
		}

		if err := db.Create(&resource).Error; err != nil {
			return fmt.Errorf("failed to create resource: %w", err)
		}

		fmt.Printf("Resource created successfully:\n")
		fmt.Printf("  ID: %s\n", resource.ID)
		fmt.Printf("  Name: %s\n", resource.Name)
		fmt.Printf("  Type: %s\n", resource.Type)
		fmt.Printf("  Location: %s\n", resource.Location)
		fmt.Printf("  Department: %s\n", resource.Department)
		fmt.Printf("  Active: %t\n", resource.Active)

		return nil
	},
}

var listResourcesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resources",
	Long:  "Display all resources in the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		var resources []database.Resource
		if err := db.Find(&resources).Error; err != nil {
			return fmt.Errorf("failed to fetch resources: %w", err)
		}

		fmt.Printf("Found %d resources:\n\n", len(resources))
		for _, resource := range resources {
			fmt.Printf("ID: %s\n", resource.ID)
			fmt.Printf("  Name: %s\n", resource.Name)
			fmt.Printf("  Type: %s\n", resource.Type)
			fmt.Printf("  Location: %s\n", resource.Location)
			fmt.Printf("  Department: %s\n", resource.Department)
			fmt.Printf("  Active: %t\n", resource.Active)
			fmt.Println()
		}

		return nil
	},
}

var deleteResourceCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource",
	Long:  "Delete a resource by ID or name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a resource ID or name")
		}

		identifier := args[0]
		resource, err := FindResourceByString(db, identifier)
		if err != nil {
			return err
		}

		if err := db.Delete(&resource).Error; err != nil {
			return fmt.Errorf("failed to delete resource: %w", err)
		}

		fmt.Printf("Resource deleted successfully: %s (%s)\n", resource.Name, resource.ID)
		return nil
	},
}

// Device management commands
var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Manage devices",
	Long:  "Create, update, delete, and list devices",
}

var createDeviceCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new device",
	Long:  "Create a new device (YubiKey, TOTP, SMS, Email) for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, _ := cmd.Flags().GetString("user-id")
		deviceType, _ := cmd.Flags().GetString("type")
		identifier, _ := cmd.Flags().GetString("identifier")
		secret, _ := cmd.Flags().GetString("secret")
		active, _ := cmd.Flags().GetBool("active")

		validTypes := []string{"yubikey", "totp", "sms", "email"}
		validType := false
		for _, t := range validTypes {
			if deviceType == t {
				validType = true
				break
			}
		}
		if !validType {
			return fmt.Errorf("device type must be one of: %s", strings.Join(validTypes, ", "))
		}

		user, err := FindUserByString(db, userID)
		if err != nil {
			return err
		}

		if secret == "" && deviceType == "totp" {
			secretBytes := make([]byte, 32)
			if _, err := rand.Read(secretBytes); err != nil {
				return fmt.Errorf("failed to generate secret: %w", err)
			}
			secret = hex.EncodeToString(secretBytes)
		}

		device := database.Device{
			ID:         uuid.New(),
			UserID:     user.ID,
			Type:       deviceType,
			Identifier: identifier,
			Secret:     secret,
			Active:     active,
			VerifiedAt: time.Now(),
		}

		if err := db.Create(&device).Error; err != nil {
			return fmt.Errorf("failed to create device: %w", err)
		}

		fmt.Printf("Device created successfully:\n")
		fmt.Printf("  ID: %s\n", device.ID)
		fmt.Printf("  User: %s (%s)\n", user.Email, user.ID)
		fmt.Printf("  Type: %s\n", device.Type)
		fmt.Printf("  Identifier: %s\n", device.Identifier)
		if device.Secret != "" {
			fmt.Printf("  Secret: %s\n", device.Secret)
		}
		fmt.Printf("  Active: %t\n", device.Active)

		return nil
	},
}

var listDevicesCmd = &cobra.Command{
	Use:   "list",
	Short: "List devices",
	Long:  "List all devices or devices for a specific user",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID, _ := cmd.Flags().GetString("user-id")

		var devices []database.Device
		var err error

		if userID != "" {
			user, err := FindUserByString(db, userID)
			if err != nil {
				return err
			}
			err = db.Preload("User").Where("user_id = ?", user.ID).Find(&devices).Error
		} else {
			err = db.Preload("User").Find(&devices).Error
		}

		if err != nil {
			return fmt.Errorf("failed to fetch devices: %w", err)
		}

		fmt.Printf("Found %d devices:\n\n", len(devices))
		for _, device := range devices {
			fmt.Printf("ID: %s\n", device.ID)
			fmt.Printf("  User: %s (%s)\n", device.User.Email, device.UserID)
			fmt.Printf("  Type: %s\n", device.Type)
			fmt.Printf("  Identifier: %s\n", device.Identifier)
			fmt.Printf("  Active: %t\n", device.Active)
			fmt.Printf("  Verified: %s\n", device.VerifiedAt.Format(time.RFC3339))
			if !device.LastUsedAt.IsZero() {
				fmt.Printf("  Last Used: %s\n", device.LastUsedAt.Format(time.RFC3339))
			}
			fmt.Println()
		}

		return nil
	},
}

var deleteDeviceCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a device",
	Long:  "Delete a device by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please provide a device ID")
		}

		deviceID := args[0]
		if _, err := uuid.Parse(deviceID); err != nil {
			return fmt.Errorf("invalid device ID format")
		}

		var device database.Device
		if err := db.Preload("User").Where("id = ?", deviceID).First(&device).Error; err != nil {
			return fmt.Errorf("device not found: %w", err)
		}

		if err := db.Delete(&device).Error; err != nil {
			return fmt.Errorf("failed to delete device: %w", err)
		}

		fmt.Printf("Device deleted successfully: %s (%s) for user %s\n", 
			device.Type, device.ID, device.User.Email)
		return nil
	},
}

// Assignment commands
var assignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign users to roles",
	Long:  "Assign or remove users from roles",
}

var assignUserToRoleCmd = &cobra.Command{
	Use:   "user-role",
	Short: "Assign a user to a role",
	Long:  "Assign a user to a role by user ID/email and role ID/name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("please provide user identifier and role identifier")
		}

		userIdentifier := args[0]
		roleIdentifier := args[1]

		user, err := FindUserByString(db, userIdentifier)
		if err != nil {
			return err
		}

		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := db.Where("id = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := db.Where("name = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		var count int64
		db.Model(&database.User{}).Joins("JOIN user_roles ON users.id = user_roles.user_id").
			Where("users.id = ? AND user_roles.role_id = ?", user.ID, role.ID).Count(&count)
		
		if count > 0 {
			return fmt.Errorf("user %s is already assigned to role %s", user.Email, role.Name)
		}

		if err := db.Model(&user).Association("Roles").Append(&role); err != nil {
			return fmt.Errorf("failed to assign user to role: %w", err)
		}

		fmt.Printf("Successfully assigned user %s to role %s\n", user.Email, role.Name)
		return nil
	},
}

var removeUserFromRoleCmd = &cobra.Command{
	Use:   "remove-user-role",
	Short: "Remove a user from a role",
	Long:  "Remove a user from a role by user ID/email and role ID/name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("please provide user identifier and role identifier")
		}

		userIdentifier := args[0]
		roleIdentifier := args[1]

		user, err := FindUserByString(db, userIdentifier)
		if err != nil {
			return err
		}

		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := db.Where("id = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := db.Where("name = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		if err := db.Model(&user).Association("Roles").Delete(&role); err != nil {
			return fmt.Errorf("failed to remove user from role: %w", err)
		}

		fmt.Printf("Successfully removed user %s from role %s\n", user.Email, role.Name)
		return nil
	},
}

var assignPermissionToRoleCmd = &cobra.Command{
	Use:   "permission-role",
	Short: "Assign a permission to a role",
	Long:  "Assign a permission to a role by permission ID/resource:action and role ID/name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("please provide permission ID/resource:action and role identifier")
		}

		permissionID := args[0]
		roleIdentifier := args[1]

		permission, err := FindPermissionByString(db, permissionID)
		if err != nil {
			return err
		}

		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := db.Where("id = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := db.Where("name = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		// Check if assignment already exists
		var count int64
		db.Model(&database.Role{}).Joins("JOIN role_permissions ON roles.id = role_permissions.role_id").
			Where("roles.id = ? AND role_permissions.permission_id = ?", role.ID, permission.ID).Count(&count)
		
		if count > 0 {
			return fmt.Errorf("permission %s:%s is already assigned to role %s", 
				permission.Resource.Name, permission.Action, role.Name)
		}

		// Assign permission to role
		if err := db.Model(&role).Association("Permissions").Append(&permission); err != nil {
			return fmt.Errorf("failed to assign permission to role: %w", err)
		}

		fmt.Printf("Successfully assigned permission %s:%s to role %s\n", 
			permission.Resource.Name, permission.Action, role.Name)
		return nil
	},
}

var removePermissionFromRoleCmd = &cobra.Command{
	Use:   "remove-permission-role",
	Short: "Remove a permission from a role",
	Long:  "Remove a permission from a role by permission ID/resource:action and role ID/name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("please provide permission ID/resource:action and role identifier")
		}

		permissionID := args[0]
		roleIdentifier := args[1]

		permission, err := FindPermissionByString(db, permissionID)
		if err != nil {
			return err
		}

		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := db.Where("id = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := db.Where("name = ?", roleIdentifier).First(&role).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		// Remove permission from role
		if err := db.Model(&role).Association("Permissions").Delete(&permission); err != nil {
			return fmt.Errorf("failed to remove permission from role: %w", err)
		}

		fmt.Printf("Successfully removed permission %s:%s from role %s\n", 
			permission.Resource.Name, permission.Action, role.Name)
		return nil
	},
}

// Authentication commands
var authenticateCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Authenticate users and check permissions",
	Long:  "Authenticate users using various device types and check their permissions",
}

var authenticateYubikeyCmd = &cobra.Command{
	Use:   "yubikey",
	Short: "Authenticate using YubiKey OTP",
	Long:  "Authenticate a user using YubiKey OTP and check if they have a specific permission (ID or resource:action format)",
	RunE: func(cmd *cobra.Command, args []string) error {
		permissionID, _ := cmd.Flags().GetString("permission")
		code, _ := cmd.Flags().GetString("code")

		if code == "" {
			return fmt.Errorf("please provide a YubiKey OTP code")
		}

		// Extract device ID from OTP (first 12 characters)
		if len(code) < 12 {
			return fmt.Errorf("invalid YubiKey OTP format")
		}
		deviceID := code[:12]

		// Verify OTP with Yubico servers first
		fmt.Printf("Verifying OTP with Yubico API...\n")
		if err := verifyYubikeyOTP(code, cfg.Yubikey); err != nil {
			return fmt.Errorf("OTP verification failed: %w", err)
		}
		fmt.Printf("✅ OTP verified successfully with Yubico\n")

		// Find the device in our database
		var device database.Device
		if err := db.Where("type = ? AND identifier = ?", "yubikey", deviceID).First(&device).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("YubiKey device not found: %s", deviceID)
			}
			return fmt.Errorf("failed to find device: %w", err)
		}

		// Get user associated with the device
		var user database.User
		if err := db.Preload("Roles.Permissions.Resource").First(&user, device.UserID).Error; err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		fmt.Printf("Device found:\n")
		fmt.Printf("  Device ID: %s\n", device.ID)
		fmt.Printf("  Device Identifier: %s\n", device.Identifier)
		fmt.Printf("  User: %s (%s)\n", user.Email, user.ID)
		fmt.Printf("  User Active: %t\n", user.Active)
		fmt.Printf("  Device Active: %t\n", device.Active)

		// Check if user and device are active
		if !user.Active {
			return fmt.Errorf("user is not active")
		}
		if !device.Active {
			return fmt.Errorf("device is not active")
		}

		// If no permission specified, just authenticate
		if permissionID == "" {
			fmt.Printf("\n✅ Authentication successful: User %s is authenticated\n", user.Email)
			return nil
		}

		// Find the permission
		permission, err := FindPermissionByString(db, permissionID)
		if err != nil {
			return err
		}

		fmt.Printf("\nChecking permission: %s:%s (%s)\n", permission.Resource.Name, permission.Action, permission.Effect)

		// Check if user has the required permission
		hasPermission := false
		for _, role := range user.Roles {
			for _, userPerm := range role.Permissions {
				if userPerm.Resource.Name == permission.Resource.Name && 
				   userPerm.Action == permission.Action && 
				   userPerm.Effect == "allow" {
					hasPermission = true
					break
				}
			}
			if hasPermission {
				break
			}
		}

		if hasPermission {
			fmt.Printf("✅ Authentication successful: User %s has permission %s:%s\n", 
				user.Email, permission.Resource.Name, permission.Action)
		} else {
			fmt.Printf("❌ Authentication failed: User %s does not have permission %s:%s\n", 
				user.Email, permission.Resource.Name, permission.Action)
			return fmt.Errorf("permission denied")
		}

		// Update device last used timestamp
		db.Model(&device).Update("last_used_at", time.Now())

		// Log authentication
		authLog := database.AuthenticationLog{
			ID:        uuid.New(),
			UserID:    user.ID,
			DeviceID:  device.ID,
			Type:      "mfa",
			Success:   hasPermission,
			IPAddress: "", // CLI doesn't have IP context
			UserAgent: "yubiapp-cli",
			Details: map[string]interface{}{
				"permission_checked": fmt.Sprintf("%s:%s", permission.Resource.Name, permission.Action),
				"permission_id":      permissionID,
				"device_type":        "yubikey",
			},
		}
		db.Create(&authLog)

		return nil
	},
}

func init() {
	// User command flags
	createUserCmd.Flags().String("email", "", "User email address")
	createUserCmd.Flags().String("username", "", "Username")
	createUserCmd.Flags().String("password", "", "Password")
	createUserCmd.Flags().String("first-name", "", "First name")
	createUserCmd.Flags().String("last-name", "", "Last name")
	createUserCmd.Flags().Bool("active", true, "Whether the user is active")
	createUserCmd.MarkFlagRequired("email")
	createUserCmd.MarkFlagRequired("username")
	createUserCmd.MarkFlagRequired("password")

	// Role command flags
	createRoleCmd.Flags().String("name", "", "Role name")
	createRoleCmd.Flags().String("description", "", "Role description")
	createRoleCmd.MarkFlagRequired("name")

	// Permission command flags
	createPermissionCmd.Flags().String("resource-id", "", "Resource ID")
	createPermissionCmd.Flags().String("action", "", "Action name (e.g., 'read', 'write')")
	createPermissionCmd.Flags().String("effect", "allow", "Effect ('allow' or 'deny')")
	createPermissionCmd.MarkFlagRequired("resource-id")
	createPermissionCmd.MarkFlagRequired("action")

	// Resource command flags
	createResourceCmd.Flags().String("name", "", "Resource name")
	createResourceCmd.Flags().String("type", "", "Resource type (server, service, database, application)")
	createResourceCmd.Flags().String("location", "", "Resource location")
	createResourceCmd.Flags().String("department", "", "Resource department")
	createResourceCmd.Flags().Bool("active", true, "Whether the resource is active")
	createResourceCmd.MarkFlagRequired("name")
	createResourceCmd.MarkFlagRequired("type")

	// Device command flags
	createDeviceCmd.Flags().String("user-id", "", "User ID")
	createDeviceCmd.Flags().String("type", "", "Device type (yubikey, totp, sms, email)")
	createDeviceCmd.Flags().String("identifier", "", "Device identifier (e.g., YubiKey public ID, phone number)")
	createDeviceCmd.Flags().String("secret", "", "Device secret (optional, auto-generated for TOTP)")
	createDeviceCmd.Flags().Bool("active", true, "Whether the device is active")
	createDeviceCmd.MarkFlagRequired("user-id")
	createDeviceCmd.MarkFlagRequired("type")
	createDeviceCmd.MarkFlagRequired("identifier")

	listDevicesCmd.Flags().String("user-id", "", "Filter devices by user ID")

	// Authentication command flags
	authenticateYubikeyCmd.Flags().String("permission", "", "Permission ID to check (optional)")
	authenticateYubikeyCmd.Flags().String("code", "", "YubiKey OTP code")
	authenticateYubikeyCmd.MarkFlagRequired("code")

	// Add subcommands
	userCmd.AddCommand(createUserCmd)
	userCmd.AddCommand(listUsersCmd)
	userCmd.AddCommand(deleteUserCmd)

	roleCmd.AddCommand(createRoleCmd)
	roleCmd.AddCommand(listRolesCmd)
	roleCmd.AddCommand(deleteRoleCmd)

	permissionCmd.AddCommand(createPermissionCmd)
	permissionCmd.AddCommand(listPermissionsCmd)
	permissionCmd.AddCommand(deletePermissionCmd)

	deviceCmd.AddCommand(createDeviceCmd)
	deviceCmd.AddCommand(listDevicesCmd)
	deviceCmd.AddCommand(deleteDeviceCmd)

	assignCmd.AddCommand(assignUserToRoleCmd)
	assignCmd.AddCommand(removeUserFromRoleCmd)
	assignCmd.AddCommand(assignPermissionToRoleCmd)
	assignCmd.AddCommand(removePermissionFromRoleCmd)

	resourceCmd.AddCommand(createResourceCmd)
	resourceCmd.AddCommand(listResourcesCmd)
	resourceCmd.AddCommand(deleteResourceCmd)

	authenticateCmd.AddCommand(authenticateYubikeyCmd)
}
