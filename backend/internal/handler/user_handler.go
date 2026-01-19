package handler

import (
	"net/http"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/service"
)

type UserHandler struct {
	userService        *service.UserService
	notificationService *service.NotificationService
}

func NewUserHandler(userService *service.UserService, notificationService *service.NotificationService) *UserHandler {
	return &UserHandler{
		userService:        userService,
		notificationService: notificationService,
	}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	user, err := h.userService.GetProfile(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req domain.UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	userID := getUserID(r)
	user, err := h.userService.UpdateProfile(r.Context(), userID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	profile, ratingSummary, err := h.userService.GetPublicProfile(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user":   profile,
		"rating": ratingSummary,
	})
}

func (h *UserHandler) GetUserAuctions(w http.ResponseWriter, r *http.Request) {
	userID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	page := getQueryParamInt(r, "page", 1)
	limit := getQueryParamInt(r, "limit", 20)

	result, err := h.userService.GetUserAuctions(r.Context(), userID, page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSONWithMeta(w, http.StatusOK, result.Auctions, &domain.APIMeta{
		Page:       result.Page,
		Limit:      limit,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *UserHandler) GetUserRatings(w http.ResponseWriter, r *http.Request) {
	userID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	params := &domain.RatingListParams{
		Page:  getQueryParamInt(r, "page", 1),
		Limit: getQueryParamInt(r, "limit", 20),
	}

	if ratingType := r.URL.Query().Get("type"); ratingType != "" {
		t := domain.RatingType(ratingType)
		params.Type = &t
	}

	result, err := h.userService.GetUserRatings(r.Context(), userID, params)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSONWithMeta(w, http.StatusOK, result.Ratings, &domain.APIMeta{
		Page:       result.Page,
		Limit:      params.Limit,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

// Watchlist handlers

func (h *UserHandler) GetWatchlist(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	page := getQueryParamInt(r, "page", 1)
	limit := getQueryParamInt(r, "limit", 20)

	result, err := h.userService.GetWatchlist(r.Context(), userID, page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSONWithMeta(w, http.StatusOK, result.Items, &domain.APIMeta{
		Page:       result.Page,
		Limit:      limit,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *UserHandler) AddToWatchlist(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "auctionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	userID := getUserID(r)
	if err := h.userService.AddToWatchlist(r.Context(), userID, auctionID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{
		"message": "Added to watchlist",
	})
}

func (h *UserHandler) RemoveFromWatchlist(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "auctionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	userID := getUserID(r)
	if err := h.userService.RemoveFromWatchlist(r.Context(), userID, auctionID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Removed from watchlist",
	})
}

// Notification handlers

func (h *UserHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	params := &domain.NotificationListParams{
		Page:  getQueryParamInt(r, "page", 1),
		Limit: getQueryParamInt(r, "limit", 20),
	}

	if unread := r.URL.Query().Get("unread"); unread == "true" {
		b := true
		params.Unread = &b
	}

	result, err := h.notificationService.GetUserNotifications(r.Context(), userID, params)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSONWithMeta(w, http.StatusOK, map[string]interface{}{
		"notifications": result.Notifications,
		"unread_count":  result.UnreadCount,
	}, &domain.APIMeta{
		Page:       result.Page,
		Limit:      params.Limit,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *UserHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	notificationID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid notification ID")
		return
	}

	userID := getUserID(r)
	if err := h.notificationService.MarkAsRead(r.Context(), userID, notificationID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Notification marked as read",
	})
}

func (h *UserHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if err := h.notificationService.MarkAllAsRead(r.Context(), userID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "All notifications marked as read",
	})
}

// Rating handlers

func (h *UserHandler) CreateRating(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "auctionId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	var req domain.CreateRatingRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	userID := getUserID(r)
	rating, err := h.userService.CreateRating(r.Context(), auctionID, userID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, rating)
}
