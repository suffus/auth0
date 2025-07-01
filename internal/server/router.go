package server

import (
	"fmt"
	"strings"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
)

func setupRouter(
	authService *services.AuthService,
	userService *services.UserService,
	roleService *services.RoleService,
	resourceService *services.ResourceService,
	permissionService *services.PermissionService,
	deviceService *services.DeviceService,
	actionService *services.ActionService,
	deviceRegService *services.DeviceRegistrationService,
	sessionService *services.SessionService,
	locationService *services.LocationService,
	userStatusService *services.UserStatusService,
	userActivityService *services.UserActivityService,
) *gin.Engine {
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Authentication endpoints
		api.POST("/auth/device", handleDeviceAuth(authService))
		api.POST("/auth/session", handleCreateSession(authService, sessionService))
		api.POST("/auth/session/refresh/:session_id", handleRefreshSession(sessionService))

		// Action endpoint - POST /auth/action/${action_name}
		api.POST("/auth/action/:action_name", handlePerformAction(authService, actionService))

		// User management - GET methods accept both device and session auth, write methods require device auth
		users := api.Group("/users")
		{
			users.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListUsers(userService))
			users.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreateUser(userService))
			users.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetUser(userService))
			users.PUT("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleUpdateUser(userService))
			users.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeleteUser(userService))
		}

		// User-role assignments (separate group to avoid conflicts) - write operations only
		userRoles := api.Group("/user-roles")
		userRoles.Use(authMiddlewareWrite(authService, "yubiapp:write"))
		{
			userRoles.POST("/:user_id/:role_id", handleAssignUserToRole(userService))
			userRoles.DELETE("/:user_id/:role_id", handleRemoveUserFromRole(userService))
		}

		// Role management - GET methods accept both device and session auth, write methods require device auth
		roles := api.Group("/roles")
		{
			roles.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListRoles(roleService))
			roles.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreateRole(roleService))
			roles.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetRole(roleService))
			roles.PUT("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleUpdateRole(roleService))
			roles.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeleteRole(roleService))
		}

		// Role-permission assignments (separate group to avoid conflicts) - write operations only
		rolePermissions := api.Group("/role-permissions")
		rolePermissions.Use(authMiddlewareWrite(authService, "yubiapp:write"))
		{
			rolePermissions.POST("/:role_id/:permission_id", handleAssignPermissionToRole(roleService))
			rolePermissions.DELETE("/:role_id/:permission_id", handleRemovePermissionFromRole(roleService))
		}

		// Resource management - GET methods accept both device and session auth, write methods require device auth
		resources := api.Group("/resources")
		{
			resources.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListResources(resourceService))
			resources.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreateResource(resourceService))
			resources.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetResource(resourceService))
			resources.PUT("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleUpdateResource(resourceService))
			resources.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeleteResource(resourceService))
		}

		// Permission management - GET methods accept both device and session auth, write methods require device auth
		permissions := api.Group("/permissions")
		{
			permissions.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListPermissions(permissionService))
			permissions.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreatePermission(permissionService))
			permissions.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetPermission(permissionService))
			permissions.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeletePermission(permissionService))
		}

		// Device management - GET methods accept both device and session auth, write methods require device auth
		devices := api.Group("/devices")
		{
			devices.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListDevices(deviceService))
			devices.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreateDevice(deviceService))

			// Device registration endpoints (action first, then ID) - write operations only
			devices.POST("/register", handleRegisterDevice(authService, deviceRegService))
			devices.POST("/deregister/:device_id", handleDeregisterDevice(authService, deviceRegService))
			devices.POST("/transfer/:device_id", handleTransferDevice(authService, deviceRegService))
			devices.GET("/history/:device_id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetDeviceHistory(authService, deviceRegService))

			// Generic :id routes
			devices.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetDevice(deviceService))
			devices.PUT("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleUpdateDevice(deviceService))
			devices.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeleteDevice(deviceService))
		}

		// Action management - GET methods accept both device and session auth, write methods require device auth
		actions := api.Group("/actions")
		{
			actions.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListActions(actionService))
			actions.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreateAction(actionService))
			actions.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetAction(actionService))
			actions.PUT("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleUpdateAction(actionService))
			actions.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeleteAction(actionService))
		}

		// Location management - GET methods accept both device and session auth, write methods require device auth
		locations := api.Group("/locations")
		{
			locations.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListLocations(locationService))
			locations.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreateLocation(locationService))
			locations.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetLocation(locationService))
			locations.PUT("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleUpdateLocation(locationService))
			locations.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeleteLocation(locationService))
		}

		// User status management - GET methods accept both device and session auth, write methods require device auth
		userStatuses := api.Group("/user-statuses")
		{
			userStatuses.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleListUserStatuses(userStatusService))
			userStatuses.POST("", authMiddlewareWrite(authService, "yubiapp:write"), handleCreateUserStatus(userStatusService))
			userStatuses.GET("/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetUserStatus(userStatusService))
			userStatuses.PUT("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleUpdateUserStatus(userStatusService))
			userStatuses.DELETE("/:id", authMiddlewareWrite(authService, "yubiapp:write"), handleDeleteUserStatus(userStatusService))
		}

		// User activity history - read-only operations, accept both device and session auth
		userActivity := api.Group("/user-activity")
		{
			userActivity.GET("", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetUserActivity(userActivityService))
			userActivity.GET("/summary", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetUserActivitySummary(userActivityService))
			userActivity.GET("/:user_id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetUserActivityByUser(userActivityService))
			userActivity.GET("/activity/:id", authMiddlewareRead(authService, sessionService, "yubiapp:read"), handleGetActivityByID(userActivityService))
		}
	}

	return router
}

// handleDeviceAuth handles device-based authentication
func handleDeviceAuth(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			DeviceType string `json:"device_type" binding:"required"`
			AuthCode   string `json:"auth_code" binding:"required"`
			Permission string `json:"permission"` // Optional permission to check
			Nonce      string `json:"nonce"`      // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, 400, err.Error())
			return
		}

		// Validate device-specific requirements
		if req.DeviceType == "yubikey" {
			if len(req.AuthCode) != 44 {
				errorResponse(c, 400, fmt.Sprintf("Invalid YubiKey OTP length. Expected 44 characters, got %d. Please ensure your YubiKey is properly inserted and tap the button to generate a complete OTP.", len(req.AuthCode)))
				return
			}

			// Validate that it contains only valid modhex characters
			validModhexChars := "cbdefghijklnrtuvCBDEFGHIJKLNRTUV"
			for _, char := range req.AuthCode {
				if !strings.ContainsRune(validModhexChars, char) {
					errorResponse(c, 400, "Invalid YubiKey OTP format. OTP should contain only modhex characters (c, b, d, e, f, g, h, i, j, k, l, n, r, t, u, v).")
					return
				}
			}
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		user, device, err := authService.AuthenticateDevice(req.DeviceType, req.AuthCode, req.Permission)
		if err != nil {
			errorResponse(c, 401, err.Error())
			return
		}

		// Build roles list
		roles := make([]gin.H, len(user.Roles))
		for i, role := range user.Roles {
			roles[i] = gin.H{
				"id":          role.ID,
				"name":        role.Name,
				"description": role.Description,
			}
		}

		successResponse(c, gin.H{
			"authenticated": true,
			"user": gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
				"active":     user.Active,
				"roles":      roles,
			},
			"device": gin.H{
				"id":         device.ID,
				"type":       device.Type,
				"identifier": device.Identifier,
			},
		})
	}
}

// Middleware and handlers will be implemented in separate files:
// - middleware.go
// - handlers.go
