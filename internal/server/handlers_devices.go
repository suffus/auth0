package server

import (
	"net/http"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Device API handlers

func handleCreateDevice(deviceService *services.DeviceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID     string `json:"user_id" binding:"required"`
			Type       string `json:"type" binding:"required"`
			Identifier string `json:"identifier" binding:"required"`
			Secret     string `json:"secret"`
			Active     bool   `json:"active"`
			Nonce      string `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		userID, err := uuid.Parse(req.UserID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		device, err := deviceService.CreateDevice(userID, req.Type, req.Identifier, req.Secret, req.Active)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		createdResponse(c, gin.H{
			"id":         device.ID,
			"user_id":    device.UserID,
			"type":       device.Type,
			"identifier": device.Identifier,
			"active":     device.Active,
			"verified_at": device.VerifiedAt,
			"created_at": device.CreatedAt,
		})
	}
}

func handleGetDevice(deviceService *services.DeviceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid device ID")
			return
		}

		device, err := deviceService.GetDeviceByID(deviceID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		itemResponse(c, gin.H{
			"id":         device.ID,
			"user": gin.H{
				"id":    device.User.ID,
				"email": device.User.Email,
				"username": device.User.Username,
			},
			"type":        device.Type,
			"identifier":  device.Identifier,
			"active":      device.Active,
			"verified_at": device.VerifiedAt,
			"last_used_at": device.LastUsedAt,
			"created_at":  device.CreatedAt,
			"updated_at":  device.UpdatedAt,
		})
	}
}

func handleListDevices(deviceService *services.DeviceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if filtering by user ID
		userIDParam := c.Query("user_id")
		activeOnly := c.Query("active") == "true"
		var userID *uuid.UUID
		if userIDParam != "" {
			parsedUserID, err := uuid.Parse(userIDParam)
			if err != nil {
				errorResponse(c, http.StatusBadRequest, "Invalid user ID")
				return
			}
			userID = &parsedUserID
		}

		var devices []database.Device
		var err error
		if activeOnly {
			devices, err = deviceService.ListActiveDevices(userID)
		} else {
			devices, err = deviceService.ListDevices(userID)
		}
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Build response
		deviceList := make([]gin.H, len(devices))
		for i, device := range devices {
			deviceList[i] = gin.H{
				"id":         device.ID,
				"user": gin.H{
					"id":    device.User.ID,
					"email": device.User.Email,
					"username": device.User.Username,
				},
				"type":        device.Type,
				"identifier":  device.Identifier,
				"active":      device.Active,
				"verified_at": device.VerifiedAt,
				"last_used_at": device.LastUsedAt,
				"created_at":  device.CreatedAt,
				"updated_at":  device.UpdatedAt,
			}
		}

		listResponse(c, deviceList, int64(len(deviceList)))
	}
}

func handleUpdateDevice(deviceService *services.DeviceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid device ID")
			return
		}

		var req struct {
			Type       *string `json:"type"`
			Identifier *string `json:"identifier"`
			Secret     *string `json:"secret"`
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
		if req.Type != nil {
			updates["type"] = *req.Type
		}
		if req.Identifier != nil {
			updates["identifier"] = *req.Identifier
		}
		if req.Secret != nil {
			updates["secret"] = *req.Secret
		}
		if req.Active != nil {
			updates["active"] = *req.Active
		}

		device, err := deviceService.UpdateDevice(deviceID, updates)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		itemResponse(c, gin.H{
			"id":         device.ID,
			"user": gin.H{
				"id":    device.User.ID,
				"email": device.User.Email,
				"username": device.User.Username,
			},
			"type":        device.Type,
			"identifier":  device.Identifier,
			"active":      device.Active,
			"verified_at": device.VerifiedAt,
			"last_used_at": device.LastUsedAt,
			"created_at":  device.CreatedAt,
			"updated_at":  device.UpdatedAt,
		})
	}
}

func handleDeleteDevice(deviceService *services.DeviceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		deviceID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid device ID")
			return
		}

		err = deviceService.DeleteDevice(deviceID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		deletedResponse(c)
	}
} 