package server

import (
	"net/http"

	"github.com/YubiApp/internal/database"
	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// handleRegisterDevice handles POST /devices/register
func handleRegisterDevice(authService *services.AuthService, deviceRegService *services.DeviceRegistrationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Extract device code from Authorization header
		var deviceCode string
		if len(authHeader) > 8 && authHeader[:8] == "yubikey:" {
			deviceCode = authHeader[8:]
		} else {
			errorResponse(c, http.StatusUnauthorized, "Invalid authorization format. Expected: yubikey:<device_code>")
			return
		}

		// Authenticate the registrar using the device code
		registrarUser, _, err := authService.AuthenticateDevice("yubikey", deviceCode, "yubiapp:register-other")
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Authentication failed: "+err.Error())
			return
		}

		// Parse request body
		var req struct {
			TargetUserID     string `json:"target_user_id" binding:"required"`
			DeviceIdentifier string `json:"device_identifier" binding:"required"`
			DeviceType       string `json:"device_type" binding:"required"`
			Notes            string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate device type
		validTypes := []string{"yubikey", "totp", "sms", "email"}
		validType := false
		for _, t := range validTypes {
			if req.DeviceType == t {
				validType = true
				break
			}
		}
		if !validType {
			errorResponse(c, http.StatusBadRequest, "Invalid device type. Must be one of: yubikey, totp, sms, email")
			return
		}

		// Find target user
		targetUserID, err := uuid.Parse(req.TargetUserID)
		if err != nil {
			// Try to find user by email
			var targetUser database.User
			if err := authService.GetDB().Where("email = ?", req.TargetUserID).First(&targetUser).Error; err != nil {
				errorResponse(c, http.StatusNotFound, "Target user not found")
				return
			}
			targetUserID = targetUser.ID
		}

		// Register device
		registration, err := deviceRegService.RegisterDevice(
			registrarUser.ID,
			targetUserID,
			req.DeviceIdentifier,
			req.DeviceType,
			req.Notes,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Failed to register device: "+err.Error())
			return
		}

		// Return success response
		successResponse(c, gin.H{
			"success": true,
			"message": "Device registered successfully",
			"registration": gin.H{
				"id":        registration.ID,
				"device_id": registration.DeviceID,
				"registrar": gin.H{
					"id":    registrarUser.ID,
					"email": registrarUser.Email,
				},
				"target_user_id": targetUserID,
				"action_type":    registration.ActionType,
				"created_at":     registration.CreatedAt,
			},
		})
	}
}

// handleDeregisterDevice handles DELETE /devices/{device_id}/deregister
func handleDeregisterDevice(authService *services.AuthService, deviceRegService *services.DeviceRegistrationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get device ID from URL
		deviceIDStr := c.Param("device_id")
		deviceID, err := uuid.Parse(deviceIDStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid device ID")
			return
		}

		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Extract device code from Authorization header
		var deviceCode string
		if len(authHeader) > 8 && authHeader[:8] == "yubikey:" {
			deviceCode = authHeader[8:]
		} else {
			errorResponse(c, http.StatusUnauthorized, "Invalid authorization format. Expected: yubikey:<device_code>")
			return
		}

		// Authenticate the deregistrar using the device code
		registrarUser, _, err := authService.AuthenticateDevice("yubikey", deviceCode, "yubiapp:deregister-other")
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Authentication failed: "+err.Error())
			return
		}

		// Parse request body
		var req struct {
			Reason string `json:"reason" binding:"required"`
			Notes  string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate reason
		validReasons := []string{"user_left", "device_lost", "device_transfer", "administrative"}
		validReason := false
		for _, r := range validReasons {
			if req.Reason == r {
				validReason = true
				break
			}
		}
		if !validReason {
			errorResponse(c, http.StatusBadRequest, "Invalid reason. Must be one of: user_left, device_lost, device_transfer, administrative")
			return
		}

		// Deregister device
		registration, err := deviceRegService.DeregisterDevice(
			registrarUser.ID,
			deviceID,
			req.Reason,
			req.Notes,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Failed to deregister device: "+err.Error())
			return
		}

		// Return success response
		successResponse(c, gin.H{
			"success": true,
			"message": "Device deregistered successfully",
			"deregistration": gin.H{
				"id":        registration.ID,
				"device_id": registration.DeviceID,
				"registrar": gin.H{
					"id":    registrarUser.ID,
					"email": registrarUser.Email,
				},
				"action_type": registration.ActionType,
				"reason":      registration.Reason,
				"created_at":  registration.CreatedAt,
			},
		})
	}
}

// handleTransferDevice handles POST /devices/{device_id}/transfer
func handleTransferDevice(authService *services.AuthService, deviceRegService *services.DeviceRegistrationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get device ID from URL
		deviceIDStr := c.Param("device_id")
		deviceID, err := uuid.Parse(deviceIDStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid device ID")
			return
		}

		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Extract device code from Authorization header
		var deviceCode string
		if len(authHeader) > 8 && authHeader[:8] == "yubikey:" {
			deviceCode = authHeader[8:]
		} else {
			errorResponse(c, http.StatusUnauthorized, "Invalid authorization format. Expected: yubikey:<device_code>")
			return
		}

		// Authenticate the transferrer using the device code
		// Note: Transfer requires both register-other and deregister-other permissions
		registrarUser, _, err := authService.AuthenticateDevice("yubikey", deviceCode, "yubiapp:register-other")
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Authentication failed: "+err.Error())
			return
		}

		// Check if user also has deregister-other permission
		hasDeregisterPermission, err := authService.CheckUserPermissionByResourceAction(registrarUser.ID, "yubiapp", "deregister-other")
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Error checking permissions: "+err.Error())
			return
		}
		if !hasDeregisterPermission {
			errorResponse(c, http.StatusForbidden, "Transfer requires both register-other and deregister-other permissions")
			return
		}

		// Parse request body
		var req struct {
			TargetUserID string `json:"target_user_id" binding:"required"`
			Notes        string `json:"notes"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Find target user
		targetUserID, err := uuid.Parse(req.TargetUserID)
		if err != nil {
			// Try to find user by email
			var targetUser database.User
			if err := authService.GetDB().Where("email = ?", req.TargetUserID).First(&targetUser).Error; err != nil {
				errorResponse(c, http.StatusNotFound, "Target user not found")
				return
			}
			targetUserID = targetUser.ID
		}

		// Transfer device
		registration, err := deviceRegService.TransferDevice(
			registrarUser.ID,
			deviceID,
			targetUserID,
			req.Notes,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Failed to transfer device: "+err.Error())
			return
		}

		// Return success response
		successResponse(c, gin.H{
			"success": true,
			"message": "Device transferred successfully",
			"transfer": gin.H{
				"id":        registration.ID,
				"device_id": registration.DeviceID,
				"registrar": gin.H{
					"id":    registrarUser.ID,
					"email": registrarUser.Email,
				},
				"target_user_id": targetUserID,
				"action_type":    registration.ActionType,
				"created_at":     registration.CreatedAt,
			},
		})
	}
}

// handleGetDeviceHistory handles GET /devices/{device_id}/history
func handleGetDeviceHistory(authService *services.AuthService, deviceRegService *services.DeviceRegistrationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get device ID from URL
		deviceIDStr := c.Param("device_id")
		deviceID, err := uuid.Parse(deviceIDStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid device ID")
			return
		}

		// Get the authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errorResponse(c, http.StatusUnauthorized, "Authorization header is required")
			return
		}

		// Extract device code from Authorization header
		var deviceCode string
		if len(authHeader) > 8 && authHeader[:8] == "yubikey:" {
			deviceCode = authHeader[8:]
		} else {
			errorResponse(c, http.StatusUnauthorized, "Invalid authorization format. Expected: yubikey:<device_code>")
			return
		}

		// Authenticate the user (any authenticated user can view device history)
		_, _, err = authService.AuthenticateDevice("yubikey", deviceCode, "")
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Authentication failed: "+err.Error())
			return
		}

		// Get device history
		history, err := deviceRegService.GetDeviceHistory(deviceID)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to get device history: "+err.Error())
			return
		}

		// Convert to response format
		historyList := make([]gin.H, len(history))
		for i, reg := range history {
			historyList[i] = gin.H{
				"id":          reg.ID,
				"action_type": reg.ActionType,
				"registrar": gin.H{
					"id":    reg.RegistrarUser.ID,
					"email": reg.RegistrarUser.Email,
				},
				"target_user": func() gin.H {
					if reg.TargetUserID != nil && reg.TargetUser != nil {
						return gin.H{
							"id":    reg.TargetUser.ID,
							"email": reg.TargetUser.Email,
						}
					}
					return gin.H{"id": nil, "email": nil}
				}(),
				"reason":     reg.Reason,
				"notes":      reg.Notes,
				"ip_address": reg.IPAddress,
				"created_at": reg.CreatedAt,
			}
		}

		// Return success response
		successResponse(c, gin.H{
			"device_id": deviceID,
			"history":   historyList,
		})
	}
}
