package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/middleware"
	"github.com/auction-cards/backend/internal/pkg/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

var validate = validator.New()

// Response helpers

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(domain.SuccessResponse(data))
}

func respondJSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *domain.APIMeta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(domain.SuccessResponseWithMeta(data, meta))
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(domain.ErrorResponse(code, message, nil))
}

func respondValidationError(w http.ResponseWriter, errors map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(domain.ErrorResponse("VALIDATION_ERROR", "Validation failed", errors))
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		respondError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, domain.ErrUnauthorized):
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	case errors.Is(err, domain.ErrForbidden):
		respondError(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
	case errors.Is(err, domain.ErrConflict):
		respondError(w, http.StatusConflict, "CONFLICT", "Resource already exists")
	case errors.Is(err, domain.ErrInvalidCredentials):
		respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
	case errors.Is(err, domain.ErrUserBanned):
		respondError(w, http.StatusForbidden, "USER_BANNED", "Account has been suspended")
	case errors.Is(err, domain.ErrEmailAlreadyExists):
		respondError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
	case errors.Is(err, domain.ErrUsernameExists):
		respondError(w, http.StatusConflict, "USERNAME_EXISTS", "Username already taken")
	case errors.Is(err, domain.ErrTokenExpired):
		respondError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token has expired")
	case errors.Is(err, domain.ErrTokenInvalid):
		respondError(w, http.StatusUnauthorized, "TOKEN_INVALID", "Invalid token")
	case errors.Is(err, domain.ErrAuctionNotActive):
		respondError(w, http.StatusBadRequest, "AUCTION_NOT_ACTIVE", "Auction is not active")
	case errors.Is(err, domain.ErrAuctionEnded):
		respondError(w, http.StatusBadRequest, "AUCTION_ENDED", "Auction has ended")
	case errors.Is(err, domain.ErrSelfBidding):
		respondError(w, http.StatusBadRequest, "SELF_BIDDING", "Cannot bid on your own auction")
	case errors.Is(err, domain.ErrBidTooLow):
		respondError(w, http.StatusBadRequest, "BID_TOO_LOW", "Bid amount is too low")
	case errors.Is(err, domain.ErrAuctionNotDraft):
		respondError(w, http.StatusBadRequest, "AUCTION_NOT_DRAFT", "Can only modify draft auctions")
	case errors.Is(err, domain.ErrConcurrentBid):
		respondError(w, http.StatusConflict, "CONCURRENT_BID", "Another bid was placed, please retry")
	case errors.Is(err, domain.ErrValidation):
		respondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data")
	default:
		respondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
	}
}

// Request parsing helpers

func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func validateRequest(v interface{}) map[string]string {
	return validate.Validate(v)
}

func getURLParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	param := chi.URLParam(r, key)
	return uuid.Parse(param)
}

func getQueryParamInt(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

func getQueryParamString(r *http.Request, key string) *string {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil
	}
	return &val
}

func getQueryParamUUID(r *http.Request, key string) *uuid.UUID {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return nil
	}
	return &id
}

func getUserID(r *http.Request) uuid.UUID {
	return middleware.GetUserID(r.Context())
}

func isAdmin(r *http.Request) bool {
	return middleware.IsAdmin(r.Context())
}
