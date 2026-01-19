package domain

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrNotFound           = errors.New("resource not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrBadRequest         = errors.New("bad request")
	ErrConflict           = errors.New("conflict")
	ErrInternalServer     = errors.New("internal server error")
	ErrValidation         = errors.New("validation error")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBanned         = errors.New("user is banned")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")

	// Auction errors
	ErrAuctionNotActive   = errors.New("auction is not active")
	ErrAuctionEnded       = errors.New("auction has ended")
	ErrSelfBidding        = errors.New("cannot bid on own auction")
	ErrBidTooLow          = errors.New("bid amount too low")
	ErrAuctionNotDraft    = errors.New("auction is not in draft status")
	ErrConcurrentBid      = errors.New("concurrent bid detected, please retry")
)

// AppError is a custom error type that includes HTTP status code
type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Err     error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// API Response types
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type APIMeta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	TotalCount int `json:"total_count,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func SuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
	}
}

func SuccessResponseWithMeta(data interface{}, meta *APIMeta) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}

func ErrorResponse(code string, message string, details map[string]string) *APIResponse {
	return &APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}
