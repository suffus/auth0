package server

import (
	"github.com/gin-gonic/gin"
)

// extractNonceFromRequest extracts nonce from request (JSON body for POST/PUT, URL param for GET)
func extractNonceFromRequest(c *gin.Context) string {
	// For GET requests, try URL parameter first
	if c.Request.Method == "GET" {
		if nonce := c.Query("nonce"); nonce != "" {
			return nonce
		}
	}
	
	// For POST/PUT requests, try to extract from the request context
	// The nonce should be stored in context by the handler after JSON binding
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		if nonce, exists := c.Get("request_nonce"); exists {
			if nonceStr, ok := nonce.(string); ok {
				return nonceStr
			}
		}
	}
	
	return ""
}

// setRequestNonce stores the nonce from the request in the context
func setRequestNonce(c *gin.Context, nonce string) {
	if nonce != "" {
		c.Set("request_nonce", nonce)
	}
}

// responseWithNonce wraps a response with the nonce from the request
func responseWithNonce(c *gin.Context, statusCode int, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	
	// Extract nonce from request and include it in response
	nonce := extractNonceFromRequest(c)
	if nonce != "" {
		data["nonce"] = nonce
	}
	
	c.JSON(statusCode, data)
}

// successResponse creates a success response with nonce from request
func successResponse(c *gin.Context, data gin.H) {
	responseWithNonce(c, 200, data)
}

// errorResponse creates an error response with nonce from request
func errorResponse(c *gin.Context, statusCode int, message string) {
	responseWithNonce(c, statusCode, gin.H{
		"error": message,
	})
}

// listResponse creates a list response with nonce from request
func listResponse(c *gin.Context, items interface{}, total int64) {
	responseWithNonce(c, 200, gin.H{
		"items": items,
		"total": total,
	})
}

// itemResponse creates a single item response with nonce from request
func itemResponse(c *gin.Context, item interface{}) {
	responseWithNonce(c, 200, gin.H{
		"item": item,
	})
}

// createdResponse creates a 201 response with nonce from request
func createdResponse(c *gin.Context, item interface{}) {
	responseWithNonce(c, 201, gin.H{
		"item": item,
	})
}

// deletedResponse creates a 204 response with nonce from request
func deletedResponse(c *gin.Context) {
	responseWithNonce(c, 204, gin.H{
		"message": "deleted",
	})
} 