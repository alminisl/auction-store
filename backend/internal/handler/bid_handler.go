package handler

import (
	"net/http"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/service"
)

type BidHandler struct {
	bidService *service.BidService
}

func NewBidHandler(bidService *service.BidService) *BidHandler {
	return &BidHandler{bidService: bidService}
}

func (h *BidHandler) PlaceBid(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	var req domain.PlaceBidRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	userID := getUserID(r)
	response, err := h.bidService.PlaceBid(r.Context(), auctionID, userID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, response)
}

func (h *BidHandler) GetBidsByAuction(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	page := getQueryParamInt(r, "page", 1)
	limit := getQueryParamInt(r, "limit", 20)

	result, err := h.bidService.GetBidsByAuction(r.Context(), auctionID, page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSONWithMeta(w, http.StatusOK, result.Bids, &domain.APIMeta{
		Page:       result.Page,
		Limit:      limit,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *BidHandler) GetMyBids(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	page := getQueryParamInt(r, "page", 1)
	limit := getQueryParamInt(r, "limit", 20)

	result, err := h.bidService.GetBidsByUser(r.Context(), userID, page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSONWithMeta(w, http.StatusOK, result.Bids, &domain.APIMeta{
		Page:       result.Page,
		Limit:      limit,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *BidHandler) BuyNow(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	userID := getUserID(r)
	response, err := h.bidService.BuyNow(r.Context(), auctionID, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, response)
}
