package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/YubiApp/internal/database"
	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

// ActionRequest represents the request body for action execution
type ActionRequest struct {
	Location  *string                 `json:"location"`  // Optional location name
	Status    *string                 `json:"status"`    // Optional status name
	StartTime *time.Time              `json:"start_time"` // Optional start time
	EndTime   *time.Time              `json:"end_time"`   // Optional end time
	Details   map[string]interface{}  `json:"details"`   // Additional details (merged with action details)
}

// handlePerformAction handles POST /auth/action/${action_name}
func handlePerformAction(
	authService *services.AuthService, 
	actionService *services.ActionService,
	userActivityService *services.UserActivityService,
	locationService *services.LocationService,
	userStatusService *services.UserStatusService,
) gin.HandlerFunc {
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

		// Extract device type and auth code from Authorization header
		// Expected format: "yubikey:cccccbvjbvdbijlrttlkfugllrrutgighrlnuibkbllj"
		// or "totp:123456", "sms:123456", etc.
		var deviceType, deviceCode string
		if colonIndex := strings.Index(authHeader, ":"); colonIndex > 0 {
			deviceType = authHeader[:colonIndex]
			deviceCode = authHeader[colonIndex+1:]
		} else {
			errorResponse(c, http.StatusUnauthorized, "Invalid authorization format. Expected: <device_type>:<auth_code>")
			return
		}

		// Check if the action exists
		action, err := actionService.GetActionByName(actionName)
		if err != nil {
			errorResponse(c, http.StatusNotFound, "Action '"+actionName+"' not found")
			return
		}

		// Check if the action is active
		if !action.Active {
			errorResponse(c, http.StatusForbidden, "Action '"+actionName+"' is inactive and cannot be executed")
			return
		}

		// Extract required permissions from action
		var requiredPermissions []string
		if action.RequiredPermissions.Status == pgtype.Present {
			if err := action.RequiredPermissions.AssignTo(&requiredPermissions); err != nil {
				errorResponse(c, http.StatusInternalServerError, "Error reading action permissions: "+err.Error())
				return
			}
		}

		// Authenticate the user using the device code and check permissions
		user, device, err := authService.AuthenticateDevice(deviceType, deviceCode, requiredPermissions...)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Authentication failed: "+err.Error())
			return
		}



		// Parse the request body
		var req ActionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid JSON in request body: "+err.Error())
			return
		}

		// Get device ID from the authentication
		deviceID := device.ID

		// Log the action in AuthenticationLog (middleware still logs here)
		details := map[string]interface{}{
			"action": actionName,
		}
		// Merge request details into authentication log details
		for key, value := range req.Details {
			details[key] = value
		}
		
		logEntry := map[string]interface{}{
			"user_id":     user.ID,
			"device_id":   deviceID,
			"action_id":   action.ID,
			"type":        "action",
			"success":     true,
			"ip_address":  c.ClientIP(),
			"user_agent":  c.GetHeader("User-Agent"),
			"details":     details,
		}

		// Create authentication log entry
		if err := authService.LogAuthentication(logEntry); err != nil {
			// Log the error but don't fail the request
			c.Error(err)
		}

		// If action has ActivityType 'user', create UserActivityHistory entry
		if action.ActivityType == "user" {
			// Find location if provided
			var location *database.Location
			if req.Location != nil {
				location, err = locationService.GetLocationByName(*req.Location)
				if err != nil {
					errorResponse(c, http.StatusBadRequest, "Location '"+*req.Location+"' not found: "+err.Error())
					return
				}
			}

			// Find status - either from request or from action details
			var status *database.UserStatus
			if req.Status != nil {
				// Use status from request
				status, err = userStatusService.GetUserStatusByName(*req.Status)
				if err != nil {
					errorResponse(c, http.StatusBadRequest, "Status '"+*req.Status+"' not found: "+err.Error())
					return
				}
			} else {
				// Try to get status from action details
				if action.Details.Status == pgtype.Present {
					var actionDetails map[string]interface{}
					if err := action.Details.AssignTo(&actionDetails); err == nil {
						if statusName, ok := actionDetails["default_status"].(string); ok {
							status, err = userStatusService.GetUserStatusByName(statusName)
							if err != nil {
								// Log the error but continue with nil status
								c.Error(err)
							}
						}
					}
				}
			}

			// Prepare activity details
			activityDetails := make(map[string]interface{})
			
			// Start with action details if present
			if action.Details.Status == pgtype.Present {
				var actionDetails map[string]interface{}
				if err := action.Details.AssignTo(&actionDetails); err == nil {
					for key, value := range actionDetails {
						activityDetails[key] = value
					}
				}
			}
			
			// Merge request details (overriding action details if needed)
			for key, value := range req.Details {
				activityDetails[key] = value
			}
			
			// Add request metadata
			activityDetails["request_location"] = req.Location
			activityDetails["request_status"] = req.Status
			activityDetails["request_start_time"] = req.StartTime
			activityDetails["request_end_time"] = req.EndTime
			activityDetails["device_id"] = deviceID
			activityDetails["ip_address"] = c.ClientIP()
			activityDetails["user_agent"] = c.GetHeader("User-Agent")

			// Create user activity history entry
			userActivity, err := userActivityService.CreateUserActivity(
				user,
				status,
				action,
				location,
				activityDetails,
				true, // Close previous activity
			)
			if err != nil {
				errorResponse(c, http.StatusInternalServerError, "Failed to create user activity: "+err.Error())
				return
			}

			// Return success response with activity information
			successResponse(c, gin.H{
				"action": actionName,
				"user_id": user.ID,
				"success": true,
				"message": "Action performed successfully",
				"user_activity": gin.H{
					"id": userActivity.ID,
					"from_datetime": userActivity.FromDateTime,
					"status": status,
					"location": location,
				},
			})
		} else {
			// For non-user actions, return standard success response
			successResponse(c, gin.H{
				"action": actionName,
				"user_id": user.ID,
				"success": true,
				"message": "Action performed successfully",
			})
		}
	}
}

// handleListActions handles GET /actions
func handleListActions(actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for active filter
		var activeOnly *bool
		if activeStr := c.Query("active"); activeStr != "" {
			if activeStr == "true" {
				active := true
				activeOnly = &active
			} else if activeStr == "false" {
				active := false
				activeOnly = &active
			}
		}

		actions, err := actionService.ListActionsWithFilter(activeOnly)
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
				"activity_type":        action.ActivityType,
				"required_permissions": action.RequiredPermissions,
				"details":              action.Details,
				"active":               action.Active,
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
			"activity_type":        action.ActivityType,
			"required_permissions": action.RequiredPermissions,
			"details":              action.Details,
			"active":               action.Active,
			"created_at":           action.CreatedAt,
			"updated_at":           action.UpdatedAt,
		})
	}
}

// handleCreateAction handles POST /actions
func handleCreateAction(actionService *services.ActionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name                string                 `json:"name" binding:"required"`
			ActivityType        string                 `json:"activity_type" binding:"required"`
			RequiredPermissions []string               `json:"required_permissions"`
			Details             map[string]interface{} `json:"details"`
			Active              bool                   `json:"active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		action, err := actionService.CreateAction(req.Name, req.ActivityType, req.RequiredPermissions, req.Details, req.Active)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to create action: "+err.Error())
			return
		}

		successResponse(c, gin.H{
			"id":                   action.ID,
			"name":                 action.Name,
			"activity_type":        action.ActivityType,
			"required_permissions": action.RequiredPermissions,
			"details":              action.Details,
			"active":               action.Active,
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
			Name                string                 `json:"name" binding:"required"`
			ActivityType        string                 `json:"activity_type"`
			RequiredPermissions []string               `json:"required_permissions"`
			Details             map[string]interface{} `json:"details"`
			Active              *bool                  `json:"active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		action, err := actionService.UpdateAction(id, req.Name, req.ActivityType, req.RequiredPermissions, req.Details, req.Active)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to update action: "+err.Error())
			return
		}

		successResponse(c, gin.H{
			"id":                   action.ID,
			"name":                 action.Name,
			"activity_type":        action.ActivityType,
			"required_permissions": action.RequiredPermissions,
			"details":              action.Details,
			"active":               action.Active,
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