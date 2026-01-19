package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/handler"
	"github.com/auction-cards/backend/internal/middleware"
	"github.com/auction-cards/backend/internal/service"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Mock auction repository
type mockAuctionRepo struct {
	auctions map[uuid.UUID]*domain.Auction
}

func newMockAuctionRepo() *mockAuctionRepo {
	return &mockAuctionRepo{
		auctions: make(map[uuid.UUID]*domain.Auction),
	}
}

func (r *mockAuctionRepo) Create(ctx context.Context, auction *domain.Auction) error {
	if auction.ID == uuid.Nil {
		auction.ID = uuid.New()
	}
	auction.CreatedAt = time.Now()
	auction.UpdatedAt = time.Now()
	auction.Version = 1
	r.auctions[auction.ID] = auction
	return nil
}

func (r *mockAuctionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Auction, error) {
	if auction, ok := r.auctions[id]; ok {
		return auction, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockAuctionRepo) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.Auction, error) {
	return r.GetByID(ctx, id)
}

func (r *mockAuctionRepo) Update(ctx context.Context, auction *domain.Auction) error {
	auction.UpdatedAt = time.Now()
	auction.Version++
	r.auctions[auction.ID] = auction
	return nil
}

func (r *mockAuctionRepo) UpdateWithVersion(ctx context.Context, auction *domain.Auction, expectedVersion int) error {
	existing, ok := r.auctions[auction.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if existing.Version != expectedVersion {
		return domain.ErrConcurrentBid
	}
	return r.Update(ctx, auction)
}

func (r *mockAuctionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.auctions, id)
	return nil
}

func (r *mockAuctionRepo) List(ctx context.Context, params *domain.AuctionListParams) ([]domain.Auction, int, error) {
	auctions := make([]domain.Auction, 0)
	for _, auction := range r.auctions {
		if params.Status != nil && auction.Status != *params.Status {
			continue
		}
		if params.SellerID != nil && auction.SellerID != *params.SellerID {
			continue
		}
		auctions = append(auctions, *auction)
	}
	return auctions, len(auctions), nil
}

func (r *mockAuctionRepo) GetEndingAuctions(ctx context.Context, before int64) ([]domain.Auction, error) {
	auctions := make([]domain.Auction, 0)
	for _, auction := range r.auctions {
		if auction.Status == domain.AuctionStatusActive && auction.EndTime.Unix() <= before {
			auctions = append(auctions, *auction)
		}
	}
	return auctions, nil
}

func (r *mockAuctionRepo) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	if auction, ok := r.auctions[id]; ok {
		auction.ViewsCount++
	}
	return nil
}

func (r *mockAuctionRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AuctionStatus, winnerID *uuid.UUID, winningBidID *uuid.UUID) error {
	if auction, ok := r.auctions[id]; ok {
		auction.Status = status
		auction.WinnerID = winnerID
		auction.WinningBidID = winningBidID
	}
	return nil
}

type mockAuctionImageRepo struct{}

func (r *mockAuctionImageRepo) Create(ctx context.Context, image *domain.AuctionImage) error {
	return nil
}

func (r *mockAuctionImageRepo) GetByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]domain.AuctionImage, error) {
	return nil, nil
}

func (r *mockAuctionImageRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (r *mockAuctionImageRepo) DeleteByAuctionID(ctx context.Context, auctionID uuid.UUID) error {
	return nil
}

func (r *mockAuctionImageRepo) UpdatePositions(ctx context.Context, auctionID uuid.UUID, positions map[uuid.UUID]int) error {
	return nil
}

type mockCategoryRepo struct {
	categories map[uuid.UUID]*domain.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	repo := &mockCategoryRepo{
		categories: make(map[uuid.UUID]*domain.Category),
	}
	// Add some default categories
	cat1 := &domain.Category{
		ID:   uuid.New(),
		Name: "Electronics",
		Slug: "electronics",
	}
	cat2 := &domain.Category{
		ID:   uuid.New(),
		Name: "Fashion",
		Slug: "fashion",
	}
	repo.categories[cat1.ID] = cat1
	repo.categories[cat2.ID] = cat2
	return repo
}

func (r *mockCategoryRepo) Create(ctx context.Context, category *domain.Category) error {
	if category.ID == uuid.Nil {
		category.ID = uuid.New()
	}
	r.categories[category.ID] = category
	return nil
}

func (r *mockCategoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	if cat, ok := r.categories[id]; ok {
		return cat, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockCategoryRepo) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	for _, cat := range r.categories {
		if cat.Slug == slug {
			return cat, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *mockCategoryRepo) Update(ctx context.Context, category *domain.Category) error {
	r.categories[category.ID] = category
	return nil
}

func (r *mockCategoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.categories, id)
	return nil
}

func (r *mockCategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	cats := make([]domain.Category, 0)
	for _, cat := range r.categories {
		cats = append(cats, *cat)
	}
	return cats, nil
}

func (r *mockCategoryRepo) GetWithAuctionCounts(ctx context.Context) ([]domain.Category, error) {
	return r.List(ctx)
}

func TestAuctionHandler_Create(t *testing.T) {
	auctionRepo := newMockAuctionRepo()
	categoryRepo := newMockCategoryRepo()
	jwtManager := newTestJWTManager()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	auctionService := service.NewAuctionService(
		auctionRepo,
		&mockAuctionImageRepo{},
		categoryRepo,
		nil, // no S3 for tests
	)

	r := createTestRouter()
	auctionHandler := handler.NewAuctionHandler(auctionService)

	r.With(authMiddleware.RequireAuth).Post("/api/auctions", auctionHandler.Create)

	// Create a test user token
	userID := uuid.New()
	token, _ := jwtManager.GenerateAccessToken(userID, "user")

	tests := []struct {
		name       string
		body       domain.CreateAuctionRequest
		token      string
		wantStatus int
		wantErr    bool
	}{
		{
			name: "successful auction creation",
			body: domain.CreateAuctionRequest{
				Title:         "Test Auction",
				Description:   stringPtr("A test auction"),
				StartingPrice: "100.00",
				StartTime:     time.Now().Add(1 * time.Hour),
				EndTime:       time.Now().Add(24 * time.Hour),
			},
			token:      token,
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "missing title",
			body: domain.CreateAuctionRequest{
				StartingPrice: "100.00",
				StartTime:     time.Now().Add(1 * time.Hour),
				EndTime:       time.Now().Add(24 * time.Hour),
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "invalid price",
			body: domain.CreateAuctionRequest{
				Title:         "Test Auction",
				StartingPrice: "invalid",
				StartTime:     time.Now().Add(1 * time.Hour),
				EndTime:       time.Now().Add(24 * time.Hour),
			},
			token:      token,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "no authentication",
			body: domain.CreateAuctionRequest{
				Title:         "Test Auction",
				StartingPrice: "100.00",
				StartTime:     time.Now().Add(1 * time.Hour),
				EndTime:       time.Now().Add(24 * time.Hour),
			},
			token:      "",
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "POST", "/api/auctions", tt.body, tt.token)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			response := parseResponse(t, rr)
			if tt.wantErr && response.Success {
				t.Errorf("expected error but got success")
			}
			if !tt.wantErr && !response.Success {
				t.Errorf("expected success but got error: %v", response.Error)
			}
		})
	}
}

func TestAuctionHandler_List(t *testing.T) {
	auctionRepo := newMockAuctionRepo()
	categoryRepo := newMockCategoryRepo()

	// Create some test auctions
	userID := uuid.New()
	for i := 0; i < 5; i++ {
		auction := &domain.Auction{
			SellerID:      userID,
			Title:         "Test Auction",
			StartingPrice: decimal.NewFromFloat(100),
			CurrentPrice:  decimal.NewFromFloat(100),
			BidIncrement:  decimal.NewFromFloat(1),
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(24 * time.Hour),
			Status:        domain.AuctionStatusActive,
		}
		auctionRepo.Create(context.Background(), auction)
	}

	auctionService := service.NewAuctionService(
		auctionRepo,
		&mockAuctionImageRepo{},
		categoryRepo,
		nil,
	)

	r := createTestRouter()
	auctionHandler := handler.NewAuctionHandler(auctionService)

	r.Get("/api/auctions", auctionHandler.List)

	tests := []struct {
		name         string
		queryParams  string
		wantStatus   int
		wantMinCount int
	}{
		{
			name:         "list all active auctions",
			queryParams:  "",
			wantStatus:   http.StatusOK,
			wantMinCount: 5,
		},
		{
			name:         "list with pagination",
			queryParams:  "?page=1&limit=2",
			wantStatus:   http.StatusOK,
			wantMinCount: 2,
		},
		{
			name:         "list with search",
			queryParams:  "?search=Test",
			wantStatus:   http.StatusOK,
			wantMinCount: 0, // Search not fully implemented in mock
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "GET", "/api/auctions"+tt.queryParams, nil, "")

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			response := parseResponse(t, rr)
			if !response.Success {
				t.Errorf("expected success but got error: %v", response.Error)
			}
		})
	}
}

func TestAuctionHandler_GetByID(t *testing.T) {
	auctionRepo := newMockAuctionRepo()
	categoryRepo := newMockCategoryRepo()

	// Create a test auction
	userID := uuid.New()
	auction := &domain.Auction{
		SellerID:      userID,
		Title:         "Test Auction",
		StartingPrice: decimal.NewFromFloat(100),
		CurrentPrice:  decimal.NewFromFloat(100),
		BidIncrement:  decimal.NewFromFloat(1),
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        domain.AuctionStatusActive,
	}
	auctionRepo.Create(context.Background(), auction)

	auctionService := service.NewAuctionService(
		auctionRepo,
		&mockAuctionImageRepo{},
		categoryRepo,
		nil,
	)

	r := createTestRouter()
	auctionHandler := handler.NewAuctionHandler(auctionService)

	r.Get("/api/auctions/{id}", auctionHandler.GetByID)

	tests := []struct {
		name       string
		auctionID  string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "get existing auction",
			auctionID:  auction.ID.String(),
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "get non-existent auction",
			auctionID:  uuid.New().String(),
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "invalid auction ID",
			auctionID:  "invalid-uuid",
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "GET", "/api/auctions/"+tt.auctionID, nil, "")

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			response := parseResponse(t, rr)
			if tt.wantErr && response.Success {
				t.Errorf("expected error but got success")
			}
			if !tt.wantErr && !response.Success {
				t.Errorf("expected success but got error: %v", response.Error)
			}
		})
	}
}

func TestAuctionHandler_GetCategories(t *testing.T) {
	categoryRepo := newMockCategoryRepo()

	auctionService := service.NewAuctionService(
		newMockAuctionRepo(),
		&mockAuctionImageRepo{},
		categoryRepo,
		nil,
	)

	r := createTestRouter()
	auctionHandler := handler.NewAuctionHandler(auctionService)

	r.Get("/api/categories", auctionHandler.GetCategories)

	rr := makeRequest(t, r, "GET", "/api/categories", nil, "")

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	response := parseResponse(t, rr)
	if !response.Success {
		t.Errorf("expected success but got error: %v", response.Error)
	}
}

func stringPtr(s string) *string {
	return &s
}
