package middleware

import (
	"log"
	"triply-server/internal/utils"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// ErrorHandler handles errors and returns appropriate responses
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default error
	code := fiber.StatusInternalServerError
	errResp := &ErrorResponse{
		Code:    "INTERNAL_ERROR",
		Message: "An unexpected error occurred",
	}

	// Check for custom AppError
	if appErr, ok := err.(*utils.AppError); ok {
		code = appErr.Status
		errResp.Code = appErr.Code
		errResp.Message = appErr.Message
		errResp.Details = appErr.Details
	} else if fiberErr, ok := err.(*fiber.Error); ok {
		// Fiber error
		code = fiberErr.Code
		errResp.Code = mapHTTPStatusToCode(code)
		errResp.Message = fiberErr.Message
	} else if err == gorm.ErrRecordNotFound {
		// GORM not found error
		code = fiber.StatusNotFound
		errResp.Code = "NOT_FOUND"
		errResp.Message = "Resource not found"
	} else {
		// Generic error - log it
		log.Printf("Unhandled error: %v", err)
		errResp.Message = err.Error()
	}

	// Don't leak internal errors in production
	if code >= 500 {
		log.Printf("Internal error: %v", err)
		// In production, you might want to hide the actual error message
		// errResp.Message = "An unexpected error occurred"
	}

	return c.Status(code).JSON(errResp)
}

func mapHTTPStatusToCode(status int) string {
	switch status {
	case 400:
		return "BAD_REQUEST"
	case 401:
		return "UNAUTHORIZED"
	case 403:
		return "FORBIDDEN"
	case 404:
		return "NOT_FOUND"
	case 409:
		return "CONFLICT"
	case 422:
		return "VALIDATION_ERROR"
	case 500:
		return "INTERNAL_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}
