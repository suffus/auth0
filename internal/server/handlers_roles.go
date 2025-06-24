package server

import (
	"net/http"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Role API handlers

func handleCreateRole(roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Nonce       string `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		role, err := roleService.CreateRole(req.Name, req.Description)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		createdResponse(c, gin.H{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"created_at":  role.CreatedAt,
		})
	}
}

func handleGetRole(roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid role ID")
			return
		}

		role, err := roleService.GetRoleByID(roleID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		// Build permissions list
		permissions := make([]gin.H, len(role.Permissions))
		for i, perm := range role.Permissions {
			permissions[i] = gin.H{
				"id":         perm.ID,
				"resource":   perm.Resource.Name,
				"action":     perm.Action,
				"effect":     perm.Effect,
				"created_at": perm.CreatedAt,
			}
		}

		itemResponse(c, gin.H{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"created_at":  role.CreatedAt,
			"updated_at":  role.UpdatedAt,
			"permissions": permissions,
		})
	}
}

func handleListRoles(roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := roleService.ListRoles()
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Build response
		roleList := make([]gin.H, len(roles))
		for i, role := range roles {
			// Build permissions list for each role
			permissions := make([]gin.H, len(role.Permissions))
			for j, perm := range role.Permissions {
				permissions[j] = gin.H{
					"id":         perm.ID,
					"resource":   perm.Resource.Name,
					"action":     perm.Action,
					"effect":     perm.Effect,
					"created_at": perm.CreatedAt,
				}
			}

			roleList[i] = gin.H{
				"id":          role.ID,
				"name":        role.Name,
				"description": role.Description,
				"created_at":  role.CreatedAt,
				"updated_at":  role.UpdatedAt,
				"permissions": permissions,
			}
		}

		listResponse(c, roleList, int64(len(roleList)))
	}
}

func handleUpdateRole(roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid role ID")
			return
		}

		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Nonce       string  `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		// Build updates map
		updates := make(map[string]interface{})
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}

		role, err := roleService.UpdateRole(roleID, updates)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Build permissions list
		permissions := make([]gin.H, len(role.Permissions))
		for i, perm := range role.Permissions {
			permissions[i] = gin.H{
				"id":         perm.ID,
				"resource":   perm.Resource.Name,
				"action":     perm.Action,
				"effect":     perm.Effect,
				"created_at": perm.CreatedAt,
			}
		}

		itemResponse(c, gin.H{
			"id":          role.ID,
			"name":        role.Name,
			"description": role.Description,
			"created_at":  role.CreatedAt,
			"updated_at":  role.UpdatedAt,
			"permissions": permissions,
		})
	}
}

func handleDeleteRole(roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid role ID")
			return
		}

		err = roleService.DeleteRole(roleID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		deletedResponse(c)
	}
}

func handleAssignPermissionToRole(roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := uuid.Parse(c.Param("role_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid role ID")
			return
		}

		permissionID, err := uuid.Parse(c.Param("permission_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
			return
		}

		err = roleService.AssignPermissionToRole(roleID, permissionID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		successResponse(c, gin.H{
			"message": "Permission assigned to role successfully",
		})
	}
}

func handleRemovePermissionFromRole(roleService *services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleID, err := uuid.Parse(c.Param("role_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid role ID")
			return
		}

		permissionID, err := uuid.Parse(c.Param("permission_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
			return
		}

		err = roleService.RemovePermissionFromRole(roleID, permissionID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		successResponse(c, gin.H{
			"message": "Permission removed from role successfully",
		})
	}
} 