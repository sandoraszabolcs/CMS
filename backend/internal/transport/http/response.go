package http

import (
	"github.com/gin-gonic/gin"
)

// ErrorCode is a typed API error code string.
type ErrorCode string

const (
	CodePassengerNotFound ErrorCode = "PASSENGER_NOT_FOUND"
	CodePassengerInactive ErrorCode = "PASSENGER_INACTIVE"
	CodeOpenCheckinExists ErrorCode = "OPEN_CHECKIN_EXISTS"
	CodeValidationError   ErrorCode = "VALIDATION_ERROR"
	CodeInternalError     ErrorCode = "INTERNAL_ERROR"
)

type apiError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type errorResponse struct {
	Error apiError `json:"error"`
}

func respondError(c *gin.Context, status int, code ErrorCode, message string) {
	c.JSON(status, errorResponse{
		Error: apiError{Code: code, Message: message},
	})
}

func respondOK(c *gin.Context, data any) {
	c.JSON(200, data)
}
