// Package response provides the standard JSON envelope for all HTTP responses.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard success/error envelope.
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// OK writes a 200 success response.
func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data})
}

// Created writes a 201 success response.
func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

// Error writes an error response. Internal (5xx) details are never exposed to
// the client to avoid leaking system information.
func Error(c *gin.Context, status int, message string, err error) {
	resp := Response{Success: false, Message: message}
	if err != nil && status < http.StatusInternalServerError {
		resp.Error = err.Error()
	}
	c.JSON(status, resp)
}
