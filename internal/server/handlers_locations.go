package server

import (
	"net/http"

	"github.com/YubiApp/internal/database"
	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Location API handlers

func handleCreateLocation(locationService *services.LocationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			Description string `json:"description"`
			Address     string `json:"address"`
			Type        string `json:"type"`
			Active      bool   `json:"active"`
			Nonce       string `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		// Set default type if not provided
		if req.Type == "" {
			req.Type = "office"
		}

		location, err := locationService.CreateLocation(req.Name, req.Description, req.Address, req.Type, req.Active)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		createdResponse(c, gin.H{
			"id":          location.ID,
			"name":        location.Name,
			"description": location.Description,
			"address":     location.Address,
			"type":        location.Type,
			"active":      location.Active,
			"created_at":  location.CreatedAt,
		})
	}
}

func handleGetLocation(locationService *services.LocationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locationID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid location ID")
			return
		}

		location, err := locationService.GetLocationByID(locationID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		itemResponse(c, gin.H{
			"id":          location.ID,
			"name":        location.Name,
			"description": location.Description,
			"address":     location.Address,
			"type":        location.Type,
			"active":      location.Active,
			"created_at":  location.CreatedAt,
			"updated_at":  location.UpdatedAt,
		})
	}
}

func handleListLocations(locationService *services.LocationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for query parameters
		activeOnly := c.Query("active") == "true"
		locationType := c.Query("type")

		var locations []database.Location
		var err error

		if locationType != "" {
			locations, err = locationService.ListLocationsByType(locationType)
		} else if activeOnly {
			locations, err = locationService.ListActiveLocations()
		} else {
			locations, err = locationService.ListLocations()
		}

		if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Build response
		locationList := make([]gin.H, len(locations))
		for i, location := range locations {
			locationList[i] = gin.H{
				"id":          location.ID,
				"name":        location.Name,
				"description": location.Description,
				"address":     location.Address,
				"type":        location.Type,
				"active":      location.Active,
				"created_at":  location.CreatedAt,
				"updated_at":  location.UpdatedAt,
			}
		}

		listResponse(c, locationList, int64(len(locationList)))
	}
}

func handleUpdateLocation(locationService *services.LocationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locationID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid location ID")
			return
		}

		var req struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Address     *string `json:"address"`
			Type        *string `json:"type"`
			Active      *bool   `json:"active"`
			Nonce       string  `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		// Build updates map
		updates := make(map[string]interface{})
		if req.Name != nil {
			updates["name"] = *req.Name
		}
		if req.Description != nil {
			updates["description"] = *req.Description
		}
		if req.Address != nil {
			updates["address"] = *req.Address
		}
		if req.Type != nil {
			updates["type"] = *req.Type
		}
		if req.Active != nil {
			updates["active"] = *req.Active
		}

		location, err := locationService.UpdateLocation(locationID, updates)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		itemResponse(c, gin.H{
			"id":          location.ID,
			"name":        location.Name,
			"description": location.Description,
			"address":     location.Address,
			"type":        location.Type,
			"active":      location.Active,
			"created_at":  location.CreatedAt,
			"updated_at":  location.UpdatedAt,
		})
	}
}

func handleDeleteLocation(locationService *services.LocationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		locationID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid location ID")
			return
		}

		err = locationService.DeleteLocation(locationID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		deletedResponse(c)
	}
} 