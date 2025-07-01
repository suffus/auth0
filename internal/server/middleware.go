package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/YubiApp/internal/database"
	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
)

// authMiddlewareRead handles authentication for read operations (GET methods)
// Accepts both device-based and session-based authentication
func authMiddlewareRead(authService *services.AuthService, sessionService *services.SessionService, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Check if it's a Bearer token (session auth) or device auth
		if strings.HasPrefix(authHeader, "Bearer ") {
			// Session-based authentication
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			
			// Validate the access token
			claims, err := sessionService.ValidateAccessToken(tokenString)
			if err != nil {
				errorResponse(c, http.StatusUnauthorized, fmt.Sprintf("Invalid access token: %v", err))
				c.Abort()
				return
			}

			// Get the session from Redis
			session, err := sessionService.GetSession(claims.SessionID)
			if err != nil {
				errorResponse(c, http.StatusUnauthorized, fmt.Sprintf("Session not found: %v", err))
				c.Abort()
				return
			}

			// Check if session is still valid (not invalidated by logout, etc.)
			if !session.IsValid {
				errorResponse(c, http.StatusUnauthorized, "Session has been invalidated")
				c.Abort()
				return
			}

			// Verify refresh count matches (prevents use of access tokens from before a refresh)
			if session.RefreshCount != claims.RefreshCount {
				errorResponse(c, http.StatusUnauthorized, "Access token is invalid (refresh count mismatch)")
				c.Abort()
				return
			}

			// Get user from database
			var user database.User
			if err := authService.GetDB().Preload("Roles.Permissions.Resource").Where("id = ?", claims.UserID).First(&user).Error; err != nil {
				errorResponse(c, http.StatusUnauthorized, "User not found")
				c.Abort()
				return
			}

			// Store session info in context
			c.Set("session", session)
			c.Set("user", &user)
			c.Set("user_id", user.ID)
			c.Set("device_id", claims.DeviceID)
			c.Set("auth_method", "session")

		} else {
			// Device-based authentication
			// Parse Authorization header format: "device_type:auth_code"
			parts := strings.SplitN(authHeader, ":", 2)
			if len(parts) != 2 {
				errorResponse(c, http.StatusUnauthorized, "Invalid Authorization header format. Expected: 'device_type:auth_code' or 'Bearer <token>'")
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

			// Store user and device in context
			c.Set("user", user)
			c.Set("user_id", user.ID)
			c.Set("device", device)
			c.Set("device_id", device.ID)
			c.Set("auth_method", "device")
		}

		// Set IP address and user agent for logging
		c.Set("client_ip", c.ClientIP())
		c.Set("user_agent", c.GetHeader("User-Agent"))

		c.Next()
	}
}

// authMiddlewareWrite handles authentication for write operations (POST, PUT, DELETE methods)
// Only accepts device-based authentication
func authMiddlewareWrite(authService *services.AuthService, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Check if it's a Bearer token (session auth) - not allowed for write operations
		if strings.HasPrefix(authHeader, "Bearer ") {
			errorResponse(c, http.StatusForbidden, "Session-based authentication not allowed for write operations. Use device authentication.")
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
		c.Set("auth_method", "device")

		// Set IP address and user agent for logging
		c.Set("client_ip", c.ClientIP())
		c.Set("user_agent", c.GetHeader("User-Agent"))

		c.Next()
	}
}

// Legacy authMiddleware for backward compatibility (device-only auth)
func authMiddleware(authService *services.AuthService, requiredPermission string) gin.HandlerFunc {
	return authMiddlewareWrite(authService, requiredPermission)
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