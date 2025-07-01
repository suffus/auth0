package commands

import (
	"fmt"
	"time"

	"github.com/YubiApp/cmd/cli/utils"
	"github.com/YubiApp/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var assignRoleCmd = &cobra.Command{
	Use:   "role",
	Short: "Assign a role to a user",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userIdentifier := args[0]
		roleIdentifier := args[1]

		// Find the user
		user, err := utils.FindUserByString(userIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		// Find the role
		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := DB.First(&role, "id = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := DB.First(&role, "name = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		// Check if assignment already exists
		var existingAssignment database.UserRole
		if err := DB.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&existingAssignment).Error; err == nil {
			return fmt.Errorf("user %s already has role %s", user.Email, role.Name)
		}

		// Create the assignment
		assignment := database.UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}

		if err := DB.Create(&assignment).Error; err != nil {
			return fmt.Errorf("failed to assign role: %w", err)
		}

		fmt.Printf("Role %s assigned to user %s\n", role.Name, user.Email)
		return nil
	},
}

var unassignRoleCmd = &cobra.Command{
	Use:   "unassign-role",
	Short: "Remove a role from a user",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		userIdentifier := args[0]
		roleIdentifier := args[1]

		// Find the user
		user, err := utils.FindUserByString(userIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		// Find the role
		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := DB.First(&role, "id = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := DB.First(&role, "name = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		// Find and delete the assignment
		var assignment database.UserRole
		if err := DB.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&assignment).Error; err != nil {
			return fmt.Errorf("user %s does not have role %s", user.Email, role.Name)
		}

		if err := DB.Delete(&assignment).Error; err != nil {
			return fmt.Errorf("failed to remove role: %w", err)
		}

		fmt.Printf("Role %s removed from user %s\n", role.Name, user.Email)
		return nil
	},
}

var assignPermissionCmd = &cobra.Command{
	Use:   "permission",
	Short: "Assign a permission to a role",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		roleIdentifier := args[0]
		permissionIdentifier := args[1]

		// Find the role
		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := DB.First(&role, "id = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := DB.First(&role, "name = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		// Find the permission
		permission, err := utils.FindPermissionByString(permissionIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find permission: %w", err)
		}

		// Check if assignment already exists
		var existingAssignment database.RolePermission
		if err := DB.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).First(&existingAssignment).Error; err == nil {
			return fmt.Errorf("role %s already has permission %s:%s", role.Name, permission.Resource.Name, permission.Action)
		}

		// Create the assignment
		assignment := database.RolePermission{
			RoleID:       role.ID,
			PermissionID: permission.ID,
		}

		if err := DB.Create(&assignment).Error; err != nil {
			return fmt.Errorf("failed to assign permission: %w", err)
		}

		fmt.Printf("Permission %s:%s assigned to role %s\n", permission.Resource.Name, permission.Action, role.Name)
		return nil
	},
}

var unassignPermissionCmd = &cobra.Command{
	Use:   "unassign-permission",
	Short: "Remove a permission from a role",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		roleIdentifier := args[0]
		permissionIdentifier := args[1]

		// Find the role
		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := DB.First(&role, "id = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := DB.First(&role, "name = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		// Find the permission
		permission, err := utils.FindPermissionByString(permissionIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find permission: %w", err)
		}

		// Find and delete the assignment
		var assignment database.RolePermission
		if err := DB.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).First(&assignment).Error; err != nil {
			return fmt.Errorf("role %s does not have permission %s:%s", role.Name, permission.Resource.Name, permission.Action)
		}

		if err := DB.Delete(&assignment).Error; err != nil {
			return fmt.Errorf("failed to remove permission: %w", err)
		}

		fmt.Printf("Permission %s:%s removed from role %s\n", permission.Resource.Name, permission.Action, role.Name)
		return nil
	},
}

var listUserRolesCmd = &cobra.Command{
	Use:   "list-user-roles",
	Short: "List roles assigned to a user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		userIdentifier := args[0]

		// Find the user
		user, err := utils.FindUserByString(userIdentifier)
		if err != nil {
			return fmt.Errorf("failed to find user: %w", err)
		}

		var roles []database.Role
		if err := DB.Joins("JOIN user_roles ON roles.id = user_roles.role_id").Where("user_roles.user_id = ?", user.ID).Find(&roles).Error; err != nil {
			return fmt.Errorf("failed to fetch user roles: %w", err)
		}

		fmt.Printf("Roles assigned to user %s:\n\n", user.Email)
		for _, role := range roles {
			fmt.Printf("ID: %s\n  Name: %s\n  Description: %s\n  Active: %t\n  Created: %s\n  Updated: %s\n\n",
				role.ID, role.Name, role.Description, role.Active, role.CreatedAt.Format(time.RFC3339), role.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

var listRolePermissionsCmd = &cobra.Command{
	Use:   "list-role-permissions",
	Short: "List permissions assigned to a role",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		roleIdentifier := args[0]

		// Find the role
		var role database.Role
		if _, err := uuid.Parse(roleIdentifier); err == nil {
			if err := DB.First(&role, "id = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		} else {
			if err := DB.First(&role, "name = ?", roleIdentifier).Error; err != nil {
				return fmt.Errorf("role not found: %w", err)
			}
		}

		var permissions []database.Permission
		if err := DB.Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").Preload("Resource").Where("role_permissions.role_id = ?", role.ID).Find(&permissions).Error; err != nil {
			return fmt.Errorf("failed to fetch role permissions: %w", err)
		}

		fmt.Printf("Permissions assigned to role %s:\n\n", role.Name)
		for _, permission := range permissions {
			fmt.Printf("ID: %s\n  Resource: %s\n  Action: %s\n  Created: %s\n  Updated: %s\n\n",
				permission.ID, permission.Resource.Name, permission.Action, permission.CreatedAt.Format(time.RFC3339), permission.UpdatedAt.Format(time.RFC3339))
		}
		return nil
	},
}

// AssignmentCmd represents the assignment command
var AssignmentCmd = &cobra.Command{
	Use:   "assign",
	Short: "Manage role and permission assignments",
	Long:  "Assign and unassign roles to users and permissions to roles",
}

// InitAssignmentCommands initializes the assignment commands and their flags
func InitAssignmentCommands() {
	// Add subcommands
	AssignmentCmd.AddCommand(assignRoleCmd)
	AssignmentCmd.AddCommand(unassignRoleCmd)
	AssignmentCmd.AddCommand(assignPermissionCmd)
	AssignmentCmd.AddCommand(unassignPermissionCmd)
	AssignmentCmd.AddCommand(listUserRolesCmd)
	AssignmentCmd.AddCommand(listRolePermissionsCmd)
} 