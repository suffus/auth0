package server

import (
	"net/http"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Resource API handlers

func handleCreateResource(resourceService *services.ResourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name       string `json:"name" binding:"required"`
			Type       string `json:"type" binding:"required"`
			Location   string `json:"location"`
			Department string `json:"department"`
			Active     bool   `json:"active"`
			Nonce      string `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		resource, err := resourceService.CreateResource(req.Name, req.Type, req.Location, req.Department, req.Active)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		createdResponse(c, gin.H{
			"id":         resource.ID,
			"name":       resource.Name,
			"type":       resource.Type,
			"location":   resource.Location,
			"department": resource.Department,
			"active":     resource.Active,
			"created_at": resource.CreatedAt,
		})
	}
}

func handleGetResource(resourceService *services.ResourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid resource ID")
			return
		}

		resource, err := resourceService.GetResourceByID(resourceID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		itemResponse(c, gin.H{
			"id":         resource.ID,
			"name":       resource.Name,
			"type":       resource.Type,
			"location":   resource.Location,
			"department": resource.Department,
			"active":     resource.Active,
			"created_at": resource.CreatedAt,
			"updated_at": resource.UpdatedAt,
		})
	}
}

func handleListResources(resourceService *services.ResourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		resources, err := resourceService.ListResources()
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Build response
		resourceList := make([]gin.H, len(resources))
		for i, resource := range resources {
			resourceList[i] = gin.H{
				"id":         resource.ID,
				"name":       resource.Name,
				"type":       resource.Type,
				"location":   resource.Location,
				"department": resource.Department,
				"active":     resource.Active,
				"created_at": resource.CreatedAt,
				"updated_at": resource.UpdatedAt,
			}
		}

		listResponse(c, resourceList, int64(len(resourceList)))
	}
}

func handleUpdateResource(resourceService *services.ResourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid resource ID")
			return
		}

		var req struct {
			Name       *string `json:"name"`
			Type       *string `json:"type"`
			Location   *string `json:"location"`
			Department *string `json:"department"`
			Active     *bool   `json:"active"`
			Nonce      string  `json:"nonce"` // Optional nonce for response signing
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
		if req.Type != nil {
			updates["type"] = *req.Type
		}
		if req.Location != nil {
			updates["location"] = *req.Location
		}
		if req.Department != nil {
			updates["department"] = *req.Department
		}
		if req.Active != nil {
			updates["active"] = *req.Active
		}

		resource, err := resourceService.UpdateResource(resourceID, updates)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		itemResponse(c, gin.H{
			"id":         resource.ID,
			"name":       resource.Name,
			"type":       resource.Type,
			"location":   resource.Location,
			"department": resource.Department,
			"active":     resource.Active,
			"created_at": resource.CreatedAt,
			"updated_at": resource.UpdatedAt,
		})
	}
}

func handleDeleteResource(resourceService *services.ResourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid resource ID")
			return
		}

		err = resourceService.DeleteResource(resourceID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		deletedResponse(c)
	}
}

// Permission API handlers

func handleCreatePermission(permissionService *services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ResourceID string `json:"resource_id" binding:"required"`
			Action     string `json:"action" binding:"required"`
			Effect     string `json:"effect" binding:"required"`
			Nonce      string `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		resourceID, err := uuid.Parse(req.ResourceID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid resource ID")
			return
		}

		permission, err := permissionService.CreatePermission(resourceID, req.Action, req.Effect)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		createdResponse(c, gin.H{
			"id":         permission.ID,
			"resource":   permission.Resource.Name,
			"action":     permission.Action,
			"effect":     permission.Effect,
			"created_at": permission.CreatedAt,
		})
	}
}

func handleGetPermission(permissionService *services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
			return
		}

		permission, err := permissionService.GetPermissionByID(permissionID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		itemResponse(c, gin.H{
			"id":         permission.ID,
			"resource":   permission.Resource.Name,
			"action":     permission.Action,
			"effect":     permission.Effect,
			"created_at": permission.CreatedAt,
			"updated_at": permission.UpdatedAt,
		})
	}
}

func handleListPermissions(permissionService *services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, err := permissionService.ListPermissions()
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Build response
		permissionList := make([]gin.H, len(permissions))
		for i, permission := range permissions {
			permissionList[i] = gin.H{
				"id":         permission.ID,
				"resource":   permission.Resource.Name,
				"action":     permission.Action,
				"effect":     permission.Effect,
				"created_at": permission.CreatedAt,
				"updated_at": permission.UpdatedAt,
			}
		}

		listResponse(c, permissionList, int64(len(permissionList)))
	}
}

func handleDeletePermission(permissionService *services.PermissionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissionID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid permission ID")
			return
		}

		err = permissionService.DeletePermission(permissionID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		deletedResponse(c)
	}
} 