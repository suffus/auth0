package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/YubiApp/internal/database"
	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
)

// authMiddleware handles device-based authentication
func authMiddleware(authService *services.AuthService, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Parse Authorization header format: "device_type:auth_code"
		parts := strings.SplitN(authHeader, ":", 2)
		if len(parts) != 2 {
			errorResponse(c, http.StatusUnauthorized, "Invalid Authorization header format. Expected: 'device_type:auth_code'")
			c.Abort()
			return
		}

		deviceType := strings.TrimSpace(parts[0])
		authCode := strings.TrimSpace(parts[1])

		if deviceType == "" || authCode == "" {
			errorResponse(c, http.StatusUnauthorized, "Device type and auth code cannot be empty")
			c.Abort()
			return
		}

		// Authenticate user and check permissions
		user, device, err := authService.AuthenticateDevice(deviceType, authCode, requiredPermission)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, fmt.Sprintf("Authentication failed: %v", err))
			c.Abort()
			return
		}

		// Store user and device in context for handlers to use
		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("device", device)
		c.Set("device_id", device.ID)

		// Set IP address and user agent for logging
		c.Set("client_ip", c.ClientIP())
		c.Set("user_agent", c.GetHeader("User-Agent"))

		c.Next()
	}
}

// adminMiddleware handles admin role validation
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user from context (set by authMiddleware)
		userInterface, exists := c.Get("user")
		if !exists {
			errorResponse(c, http.StatusUnauthorized, "User not found in context")
			c.Abort()
			return
		}

		user, ok := userInterface.(*database.User)
		if !ok {
			errorResponse(c, http.StatusInternalServerError, "Invalid user type in context")
			c.Abort()
			return
		}

		// Check if user has admin role
		hasAdminRole := false
		for _, role := range user.Roles {
			if role.Name == "admin" {
				hasAdminRole = true
				break
			}
		}

		if !hasAdminRole {
			errorResponse(c, http.StatusForbidden, "Admin role required")
			c.Abort()
			return
		}

		c.Next()
	}
} 