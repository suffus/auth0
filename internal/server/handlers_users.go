package server

import (
	"net/http"

	"github.com/YubiApp/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// User API handlers

func handleCreateUser(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email     string `json:"email" binding:"required,email"`
			Username  string `json:"username" binding:"required"`
			Password  string `json:"password" binding:"required,min=8"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Active    bool   `json:"active"`
			Nonce     string `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		user, err := userService.CreateUser(req.Email, req.Username, req.Password, req.FirstName, req.LastName, req.Active)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		createdResponse(c, gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"active":     user.Active,
			"created_at": user.CreatedAt,
		})
	}
}

func handleGetUser(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		user, err := userService.GetUserByID(userID)
		if err != nil {
			errorResponse(c, http.StatusNotFound, err.Error())
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

		itemResponse(c, gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"active":     user.Active,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
			"roles":      roles,
		})
	}
}

func handleListUsers(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := userService.ListUsers()
		if err != nil {
			errorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Build response
		userList := make([]gin.H, len(users))
		for i, user := range users {
			// Build roles list for each user
			roles := make([]gin.H, len(user.Roles))
			for j, role := range user.Roles {
				roles[j] = gin.H{
					"id":          role.ID,
					"name":        role.Name,
					"description": role.Description,
				}
			}

			userList[i] = gin.H{
				"id":         user.ID,
				"email":      user.Email,
				"username":   user.Username,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
				"active":     user.Active,
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
				"roles":      roles,
			}
		}

		listResponse(c, userList, int64(len(userList)))
	}
}

func handleUpdateUser(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		var req struct {
			Email     *string `json:"email"`
			Username  *string `json:"username"`
			Password  *string `json:"password"`
			FirstName *string `json:"first_name"`
			LastName  *string `json:"last_name"`
			Active    *bool   `json:"active"`
			Nonce     string  `json:"nonce"` // Optional nonce for response signing
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		// Store nonce in context for response functions to use
		setRequestNonce(c, req.Nonce)

		// Build updates map
		updates := make(map[string]interface{})
		if req.Email != nil {
			updates["email"] = *req.Email
		}
		if req.Username != nil {
			updates["username"] = *req.Username
		}
		if req.Password != nil {
			updates["password"] = *req.Password
		}
		if req.FirstName != nil {
			updates["first_name"] = *req.FirstName
		}
		if req.LastName != nil {
			updates["last_name"] = *req.LastName
		}
		if req.Active != nil {
			updates["active"] = *req.Active
		}

		user, err := userService.UpdateUser(userID, updates)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
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

		itemResponse(c, gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"username":   user.Username,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"active":     user.Active,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
			"roles":      roles,
		})
	}
}

func handleDeleteUser(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		err = userService.DeleteUser(userID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		deletedResponse(c)
	}
}

func handleAssignUserToRole(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		roleID, err := uuid.Parse(c.Param("role_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid role ID")
			return
		}

		err = userService.AssignUserToRole(userID, roleID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		successResponse(c, gin.H{
			"message": "User assigned to role successfully",
		})
	}
}

func handleRemoveUserFromRole(userService *services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := uuid.Parse(c.Param("user_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid user ID")
			return
		}

		roleID, err := uuid.Parse(c.Param("role_id"))
		if err != nil {
			errorResponse(c, http.StatusBadRequest, "Invalid role ID")
			return
		}

		err = userService.RemoveUserFromRole(userID, roleID)
		if err != nil {
			errorResponse(c, http.StatusBadRequest, err.Error())
			return
		}

		successResponse(c, gin.H{
			"message": "User removed from role successfully",
		})
	}
} 