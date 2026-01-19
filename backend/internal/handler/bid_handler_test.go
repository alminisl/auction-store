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

// Mock bid repository
type mockBidRepo struct {
	bids map[uuid.UUID]*domain.Bid
}

func newMockBidRepo() *mockBidRepo {
	return &mockBidRepo{
		bids: make(map[uuid.UUID]*domain.Bid),
	}
}

func (r *mockBidRepo) Create(ctx context.Context, bid *domain.Bid) error {
	if bid.ID == uuid.Nil {
		bid.ID = uuid.New()
	}
	bid.CreatedAt = time.Now()
	r.bids[bid.ID] = bid
	return nil
}

func (r *mockBidRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bid, error) {
	if bid, ok := r.bids[id]; ok {
		return bid, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockBidRepo) GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*domain.Bid, error) {
	var highest *domain.Bid
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			if highest == nil || bid.Amount.GreaterThan(highest.Amount) {
				highest = bid
			}
		}
	}
	return highest, nil
}

func (r *mockBidRepo) GetByAuctionID(ctx context.Context, auctionID uuid.UUID, page, limit int) ([]domain.Bid, int, error) {
	bids := make([]domain.Bid, 0)
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			bids = append(bids, *bid)
		}
	}
	return bids, len(bids), nil
}

func (r *mockBidRepo) GetByBidderID(ctx context.Context, bidderID uuid.UUID, page, limit int) ([]domain.Bid, int, error) {
	bids := make([]domain.Bid, 0)
	for _, bid := range r.bids {
		if bid.BidderID == bidderID {
			bids = append(bids, *bid)
		}
	}
	return bids, len(bids), nil
}

func (r *mockBidRepo) GetBidCount(ctx context.Context, auctionID uuid.UUID) (int, error) {
	count := 0
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID {
			count++
		}
	}
	return count, nil
}

func (r *mockBidRepo) GetPreviousHighBidder(ctx context.Context, auctionID uuid.UUID, excludeBidderID uuid.UUID) (*domain.Bid, error) {
	var highest *domain.Bid
	for _, bid := range r.bids {
		if bid.AuctionID == auctionID && bid.BidderID != excludeBidderID {
			if highest == nil || bid.Amount.GreaterThan(highest.Amount) {
				highest = bid
			}
		}
	}
	return highest, nil
}

func TestBidHandler_PlaceBid(t *testing.T) {
	auctionRepo := newMockAuctionRepo()
	bidRepo := newMockBidRepo()
	jwtManager := newTestJWTManager()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Create a test auction
	sellerID := uuid.New()
	bidderID := uuid.New()

	auction := &domain.Auction{
		SellerID:      sellerID,
		Title:         "Test Auction",
		StartingPrice: decimal.NewFromFloat(100),
		CurrentPrice:  decimal.NewFromFloat(100),
		BidIncrement:  decimal.NewFromFloat(5),
		StartTime:     time.Now().Add(-1 * time.Hour),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        domain.AuctionStatusActive,
	}
	auctionRepo.Create(context.Background(), auction)

	bidService := service.NewBidService(
		bidRepo,
		auctionRepo,
		nil,
		nil, // no notification service for tests
		nil, // no redis for tests
	)

	r := createTestRouter()
	bidHandler := handler.NewBidHandler(bidService)

	r.With(authMiddleware.RequireAuth).Post("/api/auctions/{id}/bids", bidHandler.PlaceBid)

	// Create tokens
	bidderToken, _ := jwtManager.GenerateAccessToken(bidderID, "user")
	sellerToken, _ := jwtManager.GenerateAccessToken(sellerID, "user")

	tests := []struct {
		name       string
		auctionID  string
		body       domain.PlaceBidRequest
		token      string
		wantStatus int
		wantErr    bool
	}{
		{
			name:      "successful bid",
			auctionID: auction.ID.String(),
			body: domain.PlaceBidRequest{
				Amount: "110.00",
			},
			token:      bidderToken,
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name:      "bid too low",
			auctionID: auction.ID.String(),
			body: domain.PlaceBidRequest{
				Amount: "101.00", // Less than current + increment
			},
			token:      bidderToken,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:      "seller cannot bid on own auction",
			auctionID: auction.ID.String(),
			body: domain.PlaceBidRequest{
				Amount: "150.00",
			},
			token:      sellerToken,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:      "no authentication",
			auctionID: auction.ID.String(),
			body: domain.PlaceBidRequest{
				Amount: "110.00",
			},
			token:      "",
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:      "non-existent auction",
			auctionID: uuid.New().String(),
			body: domain.PlaceBidRequest{
				Amount: "110.00",
			},
			token:      bidderToken,
			wantStatus: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "POST", "/api/auctions/"+tt.auctionID+"/bids", tt.body, tt.token)

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

func TestBidHandler_GetBidsByAuction(t *testing.T) {
	auctionRepo := newMockAuctionRepo()
	bidRepo := newMockBidRepo()

	// Create a test auction
	sellerID := uuid.New()
	auction := &domain.Auction{
		SellerID:      sellerID,
		Title:         "Test Auction",
		StartingPrice: decimal.NewFromFloat(100),
		CurrentPrice:  decimal.NewFromFloat(150),
		BidIncrement:  decimal.NewFromFloat(5),
		StartTime:     time.Now().Add(-1 * time.Hour),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        domain.AuctionStatusActive,
	}
	auctionRepo.Create(context.Background(), auction)

	// Create some test bids
	for i := 0; i < 5; i++ {
		bid := &domain.Bid{
			AuctionID: auction.ID,
			BidderID:  uuid.New(),
			Amount:    decimal.NewFromFloat(float64(110 + i*10)),
		}
		bidRepo.Create(context.Background(), bid)
	}

	bidService := service.NewBidService(
		bidRepo,
		auctionRepo,
		nil,
		nil,
		nil,
	)

	r := createTestRouter()
	bidHandler := handler.NewBidHandler(bidService)

	r.Get("/api/auctions/{id}/bids", bidHandler.GetBidsByAuction)

	tests := []struct {
		name       string
		auctionID  string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "get bids for existing auction",
			auctionID:  auction.ID.String(),
			wantStatus: http.StatusOK,
			wantErr:    false,
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
			rr := makeRequest(t, r, "GET", "/api/auctions/"+tt.auctionID+"/bids", nil, "")

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

func TestBidHandler_GetMyBids(t *testing.T) {
	auctionRepo := newMockAuctionRepo()
	bidRepo := newMockBidRepo()
	jwtManager := newTestJWTManager()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Create test user and bids
	userID := uuid.New()
	auction := &domain.Auction{
		SellerID:      uuid.New(),
		Title:         "Test Auction",
		StartingPrice: decimal.NewFromFloat(100),
		CurrentPrice:  decimal.NewFromFloat(150),
		BidIncrement:  decimal.NewFromFloat(5),
		StartTime:     time.Now().Add(-1 * time.Hour),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        domain.AuctionStatusActive,
	}
	auctionRepo.Create(context.Background(), auction)

	// Create bids for the user
	for i := 0; i < 3; i++ {
		bid := &domain.Bid{
			AuctionID: auction.ID,
			BidderID:  userID,
			Amount:    decimal.NewFromFloat(float64(110 + i*10)),
		}
		bidRepo.Create(context.Background(), bid)
	}

	bidService := service.NewBidService(
		bidRepo,
		auctionRepo,
		nil,
		nil,
		nil,
	)

	r := createTestRouter()
	bidHandler := handler.NewBidHandler(bidService)

	r.With(authMiddleware.RequireAuth).Get("/api/users/me/bids", bidHandler.GetMyBids)

	token, _ := jwtManager.GenerateAccessToken(userID, "user")

	tests := []struct {
		name       string
		token      string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "get my bids with auth",
			token:      token,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "get my bids without auth",
			token:      "",
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "GET", "/api/users/me/bids", nil, tt.token)

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

func TestBidHandler_BuyNow(t *testing.T) {
	auctionRepo := newMockAuctionRepo()
	bidRepo := newMockBidRepo()
	jwtManager := newTestJWTManager()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Create a test auction with buy now price
	sellerID := uuid.New()
	buyerID := uuid.New()
	buyNowPrice := decimal.NewFromFloat(500)

	auction := &domain.Auction{
		SellerID:      sellerID,
		Title:         "Test Auction with Buy Now",
		StartingPrice: decimal.NewFromFloat(100),
		CurrentPrice:  decimal.NewFromFloat(100),
		BuyNowPrice:   &buyNowPrice,
		BidIncrement:  decimal.NewFromFloat(5),
		StartTime:     time.Now().Add(-1 * time.Hour),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        domain.AuctionStatusActive,
	}
	auctionRepo.Create(context.Background(), auction)

	// Create auction without buy now
	auctionNoBuyNow := &domain.Auction{
		SellerID:      sellerID,
		Title:         "Test Auction No Buy Now",
		StartingPrice: decimal.NewFromFloat(100),
		CurrentPrice:  decimal.NewFromFloat(100),
		BidIncrement:  decimal.NewFromFloat(5),
		StartTime:     time.Now().Add(-1 * time.Hour),
		EndTime:       time.Now().Add(24 * time.Hour),
		Status:        domain.AuctionStatusActive,
	}
	auctionRepo.Create(context.Background(), auctionNoBuyNow)

	bidService := service.NewBidService(
		bidRepo,
		auctionRepo,
		nil,
		nil,
		nil,
	)

	r := createTestRouter()
	bidHandler := handler.NewBidHandler(bidService)

	r.With(authMiddleware.RequireAuth).Post("/api/auctions/{id}/buy-now", bidHandler.BuyNow)

	buyerToken, _ := jwtManager.GenerateAccessToken(buyerID, "user")
	sellerToken, _ := jwtManager.GenerateAccessToken(sellerID, "user")

	tests := []struct {
		name       string
		auctionID  string
		token      string
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "successful buy now",
			auctionID:  auction.ID.String(),
			token:      buyerToken,
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "seller cannot buy own auction",
			auctionID:  auctionNoBuyNow.ID.String(), // Use a different auction since first one is now completed
			token:      sellerToken,
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "no authentication",
			auctionID:  auctionNoBuyNow.ID.String(),
			token:      "",
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "POST", "/api/auctions/"+tt.auctionID+"/buy-now", nil, tt.token)

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
