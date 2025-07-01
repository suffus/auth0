package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	userActivityService *services.UserActivityService
}

// GetUserActivity handles GET /api/v1/user-activity
func (h *Handler) GetUserActivity(c *gin.Context) {
	// Parse query parameters
	filter := services.ActivityFilter{}

	// Parse datetime filters
	if fromStr := c.Query("from_datetime"); fromStr != "" {
		if fromTime, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.FromDateTime = &fromTime
		} else {
			errorResponse(c, http.StatusBadRequest, "Invalid from_datetime format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)")
			return
		}
	}

	if toStr := c.Query("to_datetime"); toStr != "" {
		if toTime, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.ToDateTime = &toTime
		} else {
			errorResponse(c, http.StatusBadRequest, "Invalid to_datetime format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)")
			return
		}
	}

	// Parse ID arrays
	if userIDsStr := c.Query("user_ids"); userIDsStr != "" {
		userIDs, err := parseUUIDArray(userIDsStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user_ids format")
			return
		}
		filter.UserIDs = userIDs
	}

	if locationIDsStr := c.Query("location_ids"); locationIDsStr != "" {
		locationIDs, err := parseUUIDArray(locationIDsStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid location_ids format")
			return
		}
		filter.LocationIDs = locationIDs
	}

	if statusIDsStr := c.Query("status_ids"); statusIDsStr != "" {
		statusIDs, err := parseUUIDArray(statusIDsStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid status_ids format")
			return
		}
		filter.StatusIDs = statusIDs
	}

	if actionIDsStr := c.Query("action_ids"); actionIDsStr != "" {
		actionIDs, err := parseUUIDArray(actionIDsStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid action_ids format")
			return
		}
		filter.ActionIDs = actionIDs
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		} else {
			filter.Limit = 50 // default limit
		}
	} else {
		filter.Limit = 50 // default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Get activities
	activities, total, err := h.userActivityService.GetUserActivity(filter)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get user activity: %v", err))
		return
	}

	// Build response
	response := gin.H{
		"data": activities,
		"meta": gin.H{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetUserActivitySummary handles GET /api/v1/user-activity/summary
func (h *Handler) GetUserActivitySummary(c *gin.Context) {
	// Parse query parameters
	var userIDs []uuid.UUID
	if userIDsStr := c.Query("user_ids"); userIDsStr != "" {
		parsedIDs, err := parseUUIDArray(userIDsStr)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user_ids format")
			return
		}
		userIDs = parsedIDs
	}

	// Parse datetime filters (required)
	fromStr := c.Query("from_datetime")
	if fromStr == "" {
		errorResponse(c, http.StatusBadRequest, "from_datetime is required")
		return
	}

	toStr := c.Query("to_datetime")
	if toStr == "" {
		errorResponse(c, http.StatusBadRequest, "to_datetime is required")
		return
	}

	fromTime, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid from_datetime format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)")
		return
	}

	toTime, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid to_datetime format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)")
		return
	}

	// Get summary
	summaries, err := h.userActivityService.GetActivitySummary(userIDs, fromTime, toTime)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get activity summary: %v", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": summaries})
}

// GetUserActivityByUser handles GET /api/v1/user-activity/{user_id}
func (h *Handler) GetUserActivityByUser(c *gin.Context) {
	// Parse user ID
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Parse query parameters
	filter := services.ActivityFilter{}

	// Parse datetime filters
	if fromStr := c.Query("from_datetime"); fromStr != "" {
		if fromTime, err := time.Parse(time.RFC3339, fromStr); err == nil {
			filter.FromDateTime = &fromTime
		} else {
			errorResponse(c, http.StatusBadRequest, "Invalid from_datetime format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)")
			return
		}
	}

	if toStr := c.Query("to_datetime"); toStr != "" {
		if toTime, err := time.Parse(time.RFC3339, toStr); err == nil {
			filter.ToDateTime = &toTime
		} else {
			errorResponse(c, http.StatusBadRequest, "Invalid to_datetime format. Use RFC3339 format (e.g., 2023-01-01T00:00:00Z)")
			return
		}
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		} else {
			filter.Limit = 50 // default limit
		}
	} else {
		filter.Limit = 50 // default limit
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	// Get activities for specific user
	activities, total, err := h.userActivityService.GetActivityByUser(userID, filter)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get user activity: %v", err))
		return
	}

	// Build response
	response := gin.H{
		"data": activities,
		"meta": gin.H{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetActivityByID handles GET /api/v1/user-activity/activity/{id}
func (h *Handler) GetActivityByID(c *gin.Context) {
	// Parse activity ID
	activityIDStr := c.Param("id")
	activityID, err := uuid.Parse(activityIDStr)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid activity ID")
		return
	}

	// Get activity
	activity, err := h.userActivityService.GetActivityByID(activityID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to get activity: %v", err))
		return
	}

	if activity == nil {
		errorResponse(c, http.StatusNotFound, "Activity not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": activity})
}

// parseUUIDArray parses a comma-separated string of UUIDs
func parseUUIDArray(uuidStr string) ([]uuid.UUID, error) {
	parts := strings.Split(uuidStr, ",")
	var uuids []uuid.UUID

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		parsedUUID, err := uuid.Parse(part)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, parsedUUID)
	}

	return uuids, nil
}

// Handler wrapper functions
func handleGetUserActivity(userActivityService *services.UserActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := &Handler{userActivityService: userActivityService}
		handler.GetUserActivity(c)
	}
}

func handleGetUserActivitySummary(userActivityService *services.UserActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := &Handler{userActivityService: userActivityService}
		handler.GetUserActivitySummary(c)
	}
}

func handleGetUserActivityByUser(userActivityService *services.UserActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := &Handler{userActivityService: userActivityService}
		handler.GetUserActivityByUser(c)
	}
}

func handleGetActivityByID(userActivityService *services.UserActivityService) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := &Handler{userActivityService: userActivityService}
		handler.GetActivityByID(c)
	}
} 