package server

import (
	"net/http"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
)

// Session API handlers

// handleCreateSession handles session creation after device authentication
func handleCreateSession(authService *services.AuthService, sessionService *services.SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			DeviceType string `json:"device_type" binding:"required"`
			AuthCode   string `json:"auth_code" binding:"required"`
			Permission string `json:"permission"` // Optional permission to check
			Nonce      string `json:"nonce"`      // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		// Authenticate the device first
		user, device, err := authService.AuthenticateDevice(req.DeviceType, req.AuthCode, req.Permission)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, err.Error())
			return
		}

		// Create a new session
		session, err := sessionService.CreateSession(user.ID, device.ID)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to create session: "+err.Error())
			return
		}

		// Generate access and refresh tokens
		accessToken, err := sessionService.GenerateAccessToken(session)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to generate access token: "+err.Error())
			return
		}

		refreshToken, err := sessionService.GenerateRefreshToken(session)
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, "Failed to generate refresh token: "+err.Error())
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
			"session_id":    session.ID,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
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

// handleRefreshSession handles session token refresh
func handleRefreshSession(sessionService *services.SessionService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("session_id")
		if sessionID == "" {
			errorResponse(c, http.StatusBadRequest, "Session ID is required")
			return
		}

		var req struct {
			RefreshToken string `json:"refresh_token" binding:"required"`
			Nonce        string `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		// Refresh the session and get new tokens
		session, accessToken, refreshToken, err := sessionService.RefreshSession(req.RefreshToken)
		if err != nil {
			errorResponse(c, http.StatusUnauthorized, "Failed to refresh session: "+err.Error())
			return
		}

		// Verify the session ID matches the URL parameter
		if session.ID != sessionID {
			errorResponse(c, http.StatusBadRequest, "Session ID mismatch")
			return
		}

		successResponse(c, gin.H{
			"session_id":    session.ID,
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	}
} 