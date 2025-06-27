package server

import (
	"github.com/gin-gonic/gin"
	"github.com/YubiApp/internal/services"
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
) *gin.Engine {
	router := gin.Default()

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Authentication endpoint
		api.POST("/auth/device", handleDeviceAuth(authService))

		// Action endpoint - POST /auth/action/${action_name}
		api.POST("/auth/action/:action_name", handlePerformAction(authService, actionService))

		// User management (requires yubiapp:read permission)
		users := api.Group("/users")
		users.Use(authMiddleware(authService, "yubiapp:read"))
		{
			users.GET("", handleListUsers(userService))
			users.POST("", authMiddleware(authService, "yubiapp:write"), handleCreateUser(userService))
			users.GET("/:id", handleGetUser(userService))
			users.PUT("/:id", authMiddleware(authService, "yubiapp:write"), handleUpdateUser(userService))
			users.DELETE("/:id", authMiddleware(authService, "yubiapp:write"), handleDeleteUser(userService))
		}

		// User-role assignments (separate group to avoid conflicts)
		userRoles := api.Group("/user-roles")
		userRoles.Use(authMiddleware(authService, "yubiapp:write"))
		{
			userRoles.POST("/:user_id/:role_id", handleAssignUserToRole(userService))
			userRoles.DELETE("/:user_id/:role_id", handleRemoveUserFromRole(userService))
		}

		// Role management (requires yubiapp:read permission)
		roles := api.Group("/roles")
		roles.Use(authMiddleware(authService, "yubiapp:read"))
		{
			roles.GET("", handleListRoles(roleService))
			roles.POST("", authMiddleware(authService, "yubiapp:write"), handleCreateRole(roleService))
			roles.GET("/:id", handleGetRole(roleService))
			roles.PUT("/:id", authMiddleware(authService, "yubiapp:write"), handleUpdateRole(roleService))
			roles.DELETE("/:id", authMiddleware(authService, "yubiapp:write"), handleDeleteRole(roleService))
		}

		// Role-permission assignments (separate group to avoid conflicts)
		rolePermissions := api.Group("/role-permissions")
		rolePermissions.Use(authMiddleware(authService, "yubiapp:write"))
		{
			rolePermissions.POST("/:role_id/:permission_id", handleAssignPermissionToRole(roleService))
			rolePermissions.DELETE("/:role_id/:permission_id", handleRemovePermissionFromRole(roleService))
		}

		// Resource management (requires yubiapp:read permission)
		resources := api.Group("/resources")
		resources.Use(authMiddleware(authService, "yubiapp:read"))
		{
			resources.GET("", handleListResources(resourceService))
			resources.POST("", authMiddleware(authService, "yubiapp:write"), handleCreateResource(resourceService))
			resources.GET("/:id", handleGetResource(resourceService))
			resources.PUT("/:id", authMiddleware(authService, "yubiapp:write"), handleUpdateResource(resourceService))
			resources.DELETE("/:id", authMiddleware(authService, "yubiapp:write"), handleDeleteResource(resourceService))
		}

		// Permission management (requires yubiapp:read permission)
		permissions := api.Group("/permissions")
		permissions.Use(authMiddleware(authService, "yubiapp:read"))
		{
			permissions.GET("", handleListPermissions(permissionService))
			permissions.POST("", authMiddleware(authService, "yubiapp:write"), handleCreatePermission(permissionService))
			permissions.GET("/:id", handleGetPermission(permissionService))
			permissions.DELETE("/:id", authMiddleware(authService, "yubiapp:write"), handleDeletePermission(permissionService))
		}

		// Device management (requires yubiapp:read permission)
		devices := api.Group("/devices")
		devices.Use(authMiddleware(authService, "yubiapp:read"))
		{
			devices.GET("", handleListDevices(deviceService))
			devices.POST("", authMiddleware(authService, "yubiapp:write"), handleCreateDevice(deviceService))
			devices.GET("/:id", handleGetDevice(deviceService))
			devices.PUT("/:id", authMiddleware(authService, "yubiapp:write"), handleUpdateDevice(deviceService))
			devices.DELETE("/:id", authMiddleware(authService, "yubiapp:write"), handleDeleteDevice(deviceService))
		}

		// Device registration endpoints
		deviceReg := api.Group("/devices")
		{
			deviceReg.POST("/register", handleRegisterDevice(authService, deviceRegService))
			deviceReg.POST("/:device_id/deregister", handleDeregisterDevice(authService, deviceRegService))
			deviceReg.POST("/:device_id/transfer", handleTransferDevice(authService, deviceRegService))
			deviceReg.GET("/:device_id/history", handleGetDeviceHistory(authService, deviceRegService))
		}

		// Action management (requires yubiapp:read permission)
		actions := api.Group("/actions")
		actions.Use(authMiddleware(authService, "yubiapp:read"))
		{
			actions.GET("", handleListActions(actionService))
			actions.POST("", authMiddleware(authService, "yubiapp:write"), handleCreateAction(actionService))
			actions.GET("/:id", handleGetAction(actionService))
			actions.PUT("/:id", authMiddleware(authService, "yubiapp:write"), handleUpdateAction(actionService))
			actions.DELETE("/:id", authMiddleware(authService, "yubiapp:write"), handleDeleteAction(actionService))
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