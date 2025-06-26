package server

import (
	"net/http"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handlePerformAction handles POST /auth/action/${action_name}
func handlePerformAction(authService *services.AuthService, actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		actionName := c.Param("action_name")
		if actionName == "" {
			errorResponse(c, http.StatusBadRequest, "action name is required")
			return
		}

		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Extract device code from Authorization header
		// Expected format: "yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj"
		var deviceCode string
		if len(authHeader) > 8 && authHeader[:8] == "yubikey:" {
			deviceCode = authHeader[8:]
		} else {
			errorResponse(c, http.StatusUnauthorized, "Invalid authorization format. Expected: yubikey:<device_code>")
			return
		}

		// Authenticate the user using the device code
		user, err := authService.AuthenticateDevice("yubikey", deviceCode, "")
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Authentication failed: "+err.Error())
			return
		}

		// Check if the action exists
		action, err := actionService.GetActionByName(actionName)
		if err != nil {
			errorResponse(c, http.StatusNotFound, "Action '"+actionName+"' not found")
			return
		}

		// Check if user has required permissions for the action
		hasPermission, err := actionService.CheckUserPermissionsForAction(user.ID, actionName)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Error checking permissions: "+err.Error())
			return
		}

		if !hasPermission {
			errorResponse(c, http.StatusForbidden, "User does not have required permissions for action '"+actionName+"'")
			return
		}

		// Get the request body as JSON for json_detail
		var requestBody map[string]interface{}
		if err := c.ShouldBindJSON(&requestBody); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid JSON in request body: "+err.Error())
			return
		}

		// Get device ID from the authentication
		var deviceID uuid.UUID
		// Note: This would need to be implemented in the auth service to return device info
		// For now, we'll use a zero UUID as placeholder
		deviceID = uuid.Nil

		// Log the action in AuthenticationLog
		logEntry := map[string]interface{}{
			"user_id":     user.ID,
			"device_id":   deviceID,
			"action_id":   action.ID,
			"type":        "action",
			"success":     true,
			"ip_address":  c.ClientIP(),
			"user_agent":  c.GetHeader("User-Agent"),
			"json_detail": requestBody,
		}

		// Create authentication log entry
		if err := authService.LogAuthentication(logEntry); err != nil {
			// Log the error but don't fail the request
			// In a production system, you might want to handle this differently
			c.Error(err)
		}

		// Return success response
		successResponse(c, gin.H{
			"action": actionName,
			"user_id": user.ID,
			"success": true,
			"message": "Action performed successfully",
		})
	}
}

// handleListActions handles GET /actions
func handleListActions(actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		actions, err := actionService.ListActions()
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to list actions: "+err.Error())
			return
		}

		// Convert to response format
		actionList := make([]gin.H, len(actions))
		for i, action := range actions {
			actionList[i] = gin.H{
				"id":                   action.ID,
				"name":                 action.Name,
				"required_permissions": action.RequiredPermissions,
				"created_at":           action.CreatedAt,
				"updated_at":           action.UpdatedAt,
			}
		}

		successResponse(c, gin.H{
			"actions": actionList,
		})
	}
}

// handleGetAction handles GET /actions/:id
func handleGetAction(actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid action ID: "+err.Error())
			return
		}

		action, err := actionService.GetActionByID(id)
		if err != nil {
			errorResponse(c, http.StatusNotFound, "Action not found: "+err.Error())
			return
		}

		successResponse(c, gin.H{
			"id":                   action.ID,
			"name":                 action.Name,
			"required_permissions": action.RequiredPermissions,
			"created_at":           action.CreatedAt,
			"updated_at":           action.UpdatedAt,
		})
	}
}

// handleCreateAction handles POST /actions
func handleCreateAction(actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name                string   `json:"name" binding:"required"`
			RequiredPermissions []string `json:"required_permissions"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		action, err := actionService.CreateAction(req.Name, req.RequiredPermissions)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to create action: "+err.Error())
			return
		}

		successResponse(c, gin.H{
			"id":                   action.ID,
			"name":                 action.Name,
			"required_permissions": action.RequiredPermissions,
			"created_at":           action.CreatedAt,
			"updated_at":           action.UpdatedAt,
		})
	}
}

// handleUpdateAction handles PUT /actions/:id
func handleUpdateAction(actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid action ID: "+err.Error())
			return
		}

		var req struct {
			Name                string   `json:"name" binding:"required"`
			RequiredPermissions []string `json:"required_permissions"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		action, err := actionService.UpdateAction(id, req.Name, req.RequiredPermissions)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to update action: "+err.Error())
			return
		}

		successResponse(c, gin.H{
			"id":                   action.ID,
			"name":                 action.Name,
			"required_permissions": action.RequiredPermissions,
			"created_at":           action.CreatedAt,
			"updated_at":           action.UpdatedAt,
		})
	}
}

// handleDeleteAction handles DELETE /actions/:id
func handleDeleteAction(actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid action ID: "+err.Error())
			return
		}

		if err := actionService.DeleteAction(id); err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to delete action: "+err.Error())
			return
		}

		successResponse(c, gin.H{
			"message": "Action deleted successfully",
		})
	}
} 