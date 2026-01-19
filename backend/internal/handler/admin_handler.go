package handler

import (
	"net/http"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/auction-cards/backend/internal/service"
)

type AdminHandler struct {
	userService    *service.UserService
	auctionService *service.AuctionService
	categoryRepo   repository.CategoryRepository
	reportRepo     repository.ReportRepository
	auctionRepo    repository.AuctionRepository
	bidRepo        repository.BidRepository
}

func NewAdminHandler(
	userService *service.UserService,
	auctionService *service.AuctionService,
	categoryRepo repository.CategoryRepository,
	reportRepo repository.ReportRepository,
	auctionRepo repository.AuctionRepository,
	bidRepo repository.BidRepository,
) *AdminHandler {
	return &AdminHandler{
		userService:    userService,
		auctionService: auctionService,
		categoryRepo:   categoryRepo,
		reportRepo:     reportRepo,
		auctionRepo:    auctionRepo,
		bidRepo:        bidRepo,
	}
}

func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get counts
	users, totalUsers, _ := h.userService.ListUsers(ctx, 1, 1)
	activeAuctions, activeCount, _ := h.auctionRepo.List(ctx, &domain.AuctionListParams{
		Status: ptrTo(domain.AuctionStatusActive),
		Page:   1,
		Limit:  1,
	})

	pendingReports, pendingCount, _ := h.reportRepo.List(ctx, &domain.ReportListParams{
		Status: ptrTo(domain.ReportStatusPending),
		Page:   1,
		Limit:  1,
	})

	_ = users
	_ = activeAuctions
	_ = pendingReports

	dashboard := map[string]interface{}{
		"total_users":      totalUsers,
		"active_auctions":  activeCount,
		"pending_reports":  pendingCount,
	}

	respondJSON(w, http.StatusOK, dashboard)
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page := getQueryParamInt(r, "page", 1)
	limit := getQueryParamInt(r, "limit", 20)

	users, totalCount, err := h.userService.ListUsers(r.Context(), page, limit)
	if err != nil {
		handleError(w, err)
		return
	}

	totalPages := (totalCount + limit - 1) / limit

	respondJSONWithMeta(w, http.StatusOK, users, &domain.APIMeta{
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

func (h *AdminHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	userID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	var req struct {
		Ban bool `json:"ban"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.userService.BanUser(r.Context(), userID, req.Ban); err != nil {
		handleError(w, err)
		return
	}

	action := "banned"
	if !req.Ban {
		action = "unbanned"
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "User " + action + " successfully",
	})
}

func (h *AdminHandler) ListAuctions(w http.ResponseWriter, r *http.Request) {
	params := &domain.AuctionListParams{
		Page:   getQueryParamInt(r, "page", 1),
		Limit:  getQueryParamInt(r, "limit", 20),
		SortBy: r.URL.Query().Get("sort"),
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.AuctionStatus(status)
		params.Status = &s
	}

	params.Search = getQueryParamString(r, "search")

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

func (h *AdminHandler) UpdateAuctionStatus(w http.ResponseWriter, r *http.Request) {
	auctionID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid auction ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	status := domain.AuctionStatus(req.Status)
	if err := h.auctionService.AdminUpdateStatus(r.Context(), auctionID, status); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Auction status updated successfully",
	})
}

// Category management

func (h *AdminHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	category := &domain.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		ParentID:    req.ParentID,
		Description: req.Description,
		ImageURL:    req.ImageURL,
	}

	if err := h.categoryRepo.Create(r.Context(), category); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, category)
}

func (h *AdminHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	var req domain.UpdateCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	category, err := h.categoryRepo.GetByID(r.Context(), categoryID)
	if err != nil {
		handleError(w, err)
		return
	}

	if req.Name != nil {
		category.Name = *req.Name
	}
	if req.Slug != nil {
		category.Slug = *req.Slug
	}
	if req.ParentID != nil {
		category.ParentID = req.ParentID
	}
	if req.Description != nil {
		category.Description = req.Description
	}
	if req.ImageURL != nil {
		category.ImageURL = req.ImageURL
	}

	if err := h.categoryRepo.Update(r.Context(), category); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, category)
}

func (h *AdminHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID")
		return
	}

	if err := h.categoryRepo.Delete(r.Context(), categoryID); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Category deleted successfully",
	})
}

// Reports management

func (h *AdminHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	params := &domain.ReportListParams{
		Page:  getQueryParamInt(r, "page", 1),
		Limit: getQueryParamInt(r, "limit", 20),
	}

	if status := r.URL.Query().Get("status"); status != "" {
		s := domain.ReportStatus(status)
		params.Status = &s
	}

	reports, totalCount, err := h.reportRepo.List(r.Context(), params)
	if err != nil {
		handleError(w, err)
		return
	}

	totalPages := (totalCount + params.Limit - 1) / params.Limit

	respondJSONWithMeta(w, http.StatusOK, reports, &domain.APIMeta{
		Page:       params.Page,
		Limit:      params.Limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
	})
}

func (h *AdminHandler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	reportID, err := getURLParamUUID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_ID", "Invalid report ID")
		return
	}

	var req domain.UpdateReportRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	report, err := h.reportRepo.GetByID(r.Context(), reportID)
	if err != nil {
		handleError(w, err)
		return
	}

	report.Status = req.Status

	if err := h.reportRepo.Update(r.Context(), report); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, report)
}

func ptrTo[T any](v T) *T {
	return &v
}
