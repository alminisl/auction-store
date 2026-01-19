package handler

import (
	"net/http"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/service"
	"github.com/shopspring/decimal"
)

type AuctionHandler struct {
	auctionService *service.AuctionService
}

func NewAuctionHandler(auctionService *service.AuctionService) *AuctionHandler {
	return &AuctionHandler{auctionService: auctionService}
}

func (h *AuctionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateAuctionRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	userID := getUserID(r)
	auction, err := h.auctionService.Create(r.Context(), userID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, auction)
}

func (h *AuctionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	auction, err := h.auctionService.GetByID(r.Context(), id, true)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, auction)
}

func (h *AuctionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	var req domain.UpdateAuctionRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	userID := getUserID(r)
	auction, err := h.auctionService.Update(r.Context(), id, userID, &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, auction)
}

func (h *AuctionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	userID := getUserID(r)
	if err := h.auctionService.Delete(r.Context(), id, userID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Auction deleted successfully",
	})
}

func (h *AuctionHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	userID := getUserID(r)
	auction, err := h.auctionService.Publish(r.Context(), id, userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, auction)
}

func (h *AuctionHandler) List(w http.ResponseWriter, r *http.Request) {
	params := &domain.AuctionListParams{
		Page:   getQueryParamInt(r, "page", 1),
		Limit:  getQueryParamInt(r, "limit", 20),
		SortBy: r.URL.Query().Get("sort"),
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.AuctionStatus(status)
		params.Status = &s
	} else {
		// Default to active auctions for public listing
		s := domain.AuctionStatusActive
		params.Status = &s
	}

	params.CategoryID = getQueryParamUUID(r, "category_id")
	params.SellerID = getQueryParamUUID(r, "seller_id")
	params.Search = getQueryParamString(r, "search")

	if minPrice := r.URL.Query().Get("min_price"); minPrice != "" {
		price, _ := decimal.NewFromString(minPrice)
		params.MinPrice = &price
	}
	if maxPrice := r.URL.Query().Get("max_price"); maxPrice != "" {
		price, _ := decimal.NewFromString(maxPrice)
		params.MaxPrice = &price
	}

	result, err := h.auctionService.List(r.Context(), params)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSONWithMeta(w, http.StatusOK, result.Auctions, &domain.APIMeta{
		Page:       result.Page,
		Limit:      params.Limit,
		TotalCount: result.TotalCount,
		TotalPages: result.TotalPages,
	})
}

func (h *AuctionHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	id, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_FORM", "Invalid form data")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		respondError(w, http.StatusBadRequest, "NO_FILE", "No image file provided")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")

	userID := getUserID(r)
	image, err := h.auctionService.UploadImage(r.Context(), id, userID, file, contentType, header.Size)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, image)
}

func (h *AuctionHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	imageID, err := getURLParamUUID(r, "imageId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_IMAGE_ID", "Invalid image ID")
		return
	}

	userID := getUserID(r)
	if err := h.auctionService.DeleteImage(r.Context(), auctionID, imageID, userID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Image deleted successfully",
	})
}

// Category handlers

func (h *AuctionHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.auctionService.GetCategories(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, categories)
}

func (h *AuctionHandler) GetCategoryBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		respondError(w, http.StatusBadRequest, "INVALID_SLUG", "Category slug is required")
		return
	}

	category, err := h.auctionService.GetCategoryBySlug(r.Context(), slug)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, category)
}
