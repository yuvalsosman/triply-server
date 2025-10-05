package utils

import "fmt"

// AppError represents a custom application error
type AppError struct {
	Code    string
	Message string
	Details interface{}
	Status  int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    "NOT_FOUND",
		Message: fmt.Sprintf("%s not found", resource),
		Status:  404,
	}
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError() *AppError {
	return &AppError{
		Code:    "UNAUTHORIZED",
		Message: "Authentication required",
		Status:  401,
	}
}

// NewValidationError creates a validation error
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Status:  400,
	}
}

// NewInternalError creates an internal server error
func NewInternalError(message string) *AppError {
	return &AppError{
		Code:    "INTERNAL_ERROR",
		Message: message,
		Status:  500,
	}
}
