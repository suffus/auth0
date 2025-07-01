package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/YubiApp/internal/database"
	"github.com/YubiApp/internal/services"
)

// handleListUserStatuses handles GET /user-statuses
func handleListUserStatuses(userStatusService *services.UserStatusService) gin.HandlerFunc {
	return func(c *gin.Context) {
		activeOnly := c.Query("active") == "true"
		var userStatuses []database.UserStatus
		var err error
		if activeOnly {
			userStatuses, err = userStatusService.ListActiveUserStatuses()
		} else {
			userStatuses, err = userStatusService.ListUserStatuses()
		}
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Build response
		userStatusList := make([]gin.H, len(userStatuses))
		for i, userStatus := range userStatuses {
			userStatusList[i] = gin.H{
				"id":          userStatus.ID,
				"name":        userStatus.Name,
				"description": userStatus.Description,
				"type":        userStatus.Type,
				"active":      userStatus.Active,
				"created_at":  userStatus.CreatedAt,
				"updated_at":  userStatus.UpdatedAt,
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"items": userStatusList,
			"total": len(userStatusList),
		})
	}
}

// handleCreateUserStatus handles POST /user-statuses
func handleCreateUserStatus(userStatusService *services.UserStatusService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Type        string `json:"type"`
			Active      bool   `json:"active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Set default type if not provided
		if req.Type == "" {
			req.Type = "working"
		}

		userStatus, err := userStatusService.CreateUserStatus(req.Name, req.Description, req.Type, req.Active)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":          userStatus.ID,
			"name":        userStatus.Name,
			"description": userStatus.Description,
			"type":        userStatus.Type,
			"active":      userStatus.Active,
			"created_at":  userStatus.CreatedAt,
			"updated_at":  userStatus.UpdatedAt,
		})
	}
}

// handleGetUserStatus handles GET /user-statuses/{id}
func handleGetUserStatus(userStatusService *services.UserStatusService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user status ID")
			return
		}

		userStatus, err := userStatusService.GetUserStatusByID(id)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          userStatus.ID,
			"name":        userStatus.Name,
			"description": userStatus.Description,
			"type":        userStatus.Type,
			"active":      userStatus.Active,
			"created_at":  userStatus.CreatedAt,
			"updated_at":  userStatus.UpdatedAt,
		})
	}
}

// handleUpdateUserStatus handles PUT /user-statuses/{id}
func handleUpdateUserStatus(userStatusService *services.UserStatusService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user status ID")
			return
		}

		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Type        *string `json:"type"`
			Active      *bool   `json:"active"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		userStatus, err := userStatusService.UpdateUserStatus(id, req.Name, req.Description, req.Type, req.Active)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":          userStatus.ID,
			"name":        userStatus.Name,
			"description": userStatus.Description,
			"type":        userStatus.Type,
			"active":      userStatus.Active,
			"created_at":  userStatus.CreatedAt,
			"updated_at":  userStatus.UpdatedAt,
		})
	}
}

// handleDeleteUserStatus handles DELETE /user-statuses/{id}
func handleDeleteUserStatus(userStatusService *services.UserStatusService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user status ID")
			return
		}

		if err := userStatusService.DeleteUserStatus(id); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		c.Status(http.StatusNoContent)
	}
} 