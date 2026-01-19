package service

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/pkg/storage"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AuctionService struct {
	auctionRepo      repository.AuctionRepository
	auctionImageRepo repository.AuctionImageRepository
	categoryRepo     repository.CategoryRepository
	storage          *storage.S3Storage
}

func NewAuctionService(
	auctionRepo repository.AuctionRepository,
	auctionImageRepo repository.AuctionImageRepository,
	categoryRepo repository.CategoryRepository,
	storage *storage.S3Storage,
) *AuctionService {
	return &AuctionService{
		auctionRepo:      auctionRepo,
		auctionImageRepo: auctionImageRepo,
		categoryRepo:     categoryRepo,
		storage:          storage,
	}
}

func (s *AuctionService) Create(ctx context.Context, sellerID uuid.UUID, req *domain.CreateAuctionRequest) (*domain.Auction, error) {
	startingPrice, err := decimal.NewFromString(req.StartingPrice)
	if err != nil {
		return nil, domain.ErrBadRequest
	}

	auction := &domain.Auction{
		SellerID:      sellerID,
		CategoryID:    req.CategoryID,
		Title:         req.Title,
		Description:   req.Description,
		StartingPrice: startingPrice,
		CurrentPrice:  startingPrice,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Status:        domain.AuctionStatusDraft,
		BidIncrement:  decimal.NewFromFloat(1.00),
	}

	if req.Condition != nil {
		condition := domain.ItemCondition(*req.Condition)
		auction.Condition = &condition
	}

	if req.ReservePrice != nil {
		reservePrice, _ := decimal.NewFromString(*req.ReservePrice)
		auction.ReservePrice = &reservePrice
	}

	if req.BuyNowPrice != nil {
		buyNowPrice, _ := decimal.NewFromString(*req.BuyNowPrice)
		auction.BuyNowPrice = &buyNowPrice
	}

	if req.BidIncrement != nil {
		bidIncrement, _ := decimal.NewFromString(*req.BidIncrement)
		auction.BidIncrement = bidIncrement
	}

	if err := s.auctionRepo.Create(ctx, auction); err != nil {
		return nil, err
	}

	return auction, nil
}

func (s *AuctionService) GetByID(ctx context.Context, id uuid.UUID, incrementViews bool) (*domain.Auction, error) {
	auction, err := s.auctionRepo.GetByIDWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}

	if incrementViews {
		_ = s.auctionRepo.IncrementViewCount(ctx, id)
	}

	return auction, nil
}

func (s *AuctionService) Update(ctx context.Context, id, sellerID uuid.UUID, req *domain.UpdateAuctionRequest) (*domain.Auction, error) {
	auction, err := s.auctionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only seller can update
	if auction.SellerID != sellerID {
		return nil, domain.ErrForbidden
	}

	if req.CategoryID != nil {
		auction.CategoryID = req.CategoryID
	}
	if req.Title != nil {
		auction.Title = *req.Title
	}
	if req.Description != nil {
		auction.Description = req.Description
	}
	if req.Condition != nil {
		condition := domain.ItemCondition(*req.Condition)
		auction.Condition = &condition
	}
	if req.StartingPrice != nil {
		price, _ := decimal.NewFromString(*req.StartingPrice)
		auction.StartingPrice = price
		auction.CurrentPrice = price
	}
	if req.ReservePrice != nil {
		price, _ := decimal.NewFromString(*req.ReservePrice)
		auction.ReservePrice = &price
	}
	if req.BuyNowPrice != nil {
		price, _ := decimal.NewFromString(*req.BuyNowPrice)
		auction.BuyNowPrice = &price
	}
	if req.BidIncrement != nil {
		increment, _ := decimal.NewFromString(*req.BidIncrement)
		auction.BidIncrement = increment
	}
	if req.StartTime != nil {
		auction.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		auction.EndTime = *req.EndTime
	}

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return nil, err
	}

	return auction, nil
}

func (s *AuctionService) Delete(ctx context.Context, id, sellerID uuid.UUID) error {
	auction, err := s.auctionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Only seller can delete
	if auction.SellerID != sellerID {
		return domain.ErrForbidden
	}

	// Delete images from storage
	images, _ := s.auctionImageRepo.GetByAuctionID(ctx, id)
	for _, img := range images {
		_ = s.storage.Delete(ctx, img.URL)
	}

	return s.auctionRepo.Delete(ctx, id)
}

func (s *AuctionService) Publish(ctx context.Context, id, sellerID uuid.UUID) (*domain.Auction, error) {
	auction, err := s.auctionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only seller can publish
	if auction.SellerID != sellerID {
		return nil, domain.ErrForbidden
	}

	// Can only publish draft auctions
	if auction.Status != domain.AuctionStatusDraft {
		return nil, domain.ErrAuctionNotDraft
	}

	// Validate auction has required data
	if auction.StartTime.Before(time.Now()) {
		// If start time is in the past, set to now
		auction.StartTime = time.Now()
	}

	auction.Status = domain.AuctionStatusActive

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return nil, err
	}

	return auction, nil
}

func (s *AuctionService) List(ctx context.Context, params *domain.AuctionListParams) (*domain.AuctionListResponse, error) {
	auctions, totalCount, err := s.auctionRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	// Fetch first image for each auction
	if len(auctions) > 0 {
		auctionIDs := make([]uuid.UUID, len(auctions))
		for i, a := range auctions {
			auctionIDs[i] = a.ID
		}

		images, err := s.auctionImageRepo.GetFirstImageByAuctionIDs(ctx, auctionIDs)
		if err == nil {
			for i := range auctions {
				if img, ok := images[auctions[i].ID]; ok {
					auctions[i].Images = []domain.AuctionImage{img}
				}
			}
		}
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}

	totalPages := (totalCount + limit - 1) / limit

	return &domain.AuctionListResponse{
		Auctions:   auctions,
		TotalCount: totalCount,
		Page:       params.Page,
		TotalPages: totalPages,
	}, nil
}

func (s *AuctionService) UploadImage(ctx context.Context, auctionID, sellerID uuid.UUID, reader io.Reader, contentType string, size int64) (*domain.AuctionImage, error) {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Only seller can upload images
	if auction.SellerID != sellerID {
		return nil, domain.ErrForbidden
	}

	// Can only upload to draft auctions
	if auction.Status != domain.AuctionStatusDraft {
		return nil, domain.ErrAuctionNotDraft
	}

	// Validate content type
	if !storage.ValidateImageContentType(contentType) {
		return nil, errors.New("invalid image type")
	}

	// Validate size
	if size > storage.MaxImageSize {
		return nil, errors.New("image too large")
	}

	// Get current image count for position
	images, _ := s.auctionImageRepo.GetByAuctionID(ctx, auctionID)
	position := len(images)

	// Upload to S3
	folder := storage.GetImageFolder(auctionID)
	url, err := s.storage.Upload(ctx, reader, contentType, size, folder)
	if err != nil {
		return nil, err
	}

	// Save to database
	image := &domain.AuctionImage{
		AuctionID: auctionID,
		URL:       url,
		Position:  position,
	}

	if err := s.auctionImageRepo.Create(ctx, image); err != nil {
		// Try to delete uploaded file
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}

	return image, nil
}

func (s *AuctionService) DeleteImage(ctx context.Context, auctionID, imageID, sellerID uuid.UUID) error {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return err
	}

	// Only seller can delete images
	if auction.SellerID != sellerID {
		return domain.ErrForbidden
	}

	// Can only delete from draft auctions
	if auction.Status != domain.AuctionStatusDraft {
		return domain.ErrAuctionNotDraft
	}

	// Get image
	images, err := s.auctionImageRepo.GetByAuctionID(ctx, auctionID)
	if err != nil {
		return err
	}

	var imageToDelete *domain.AuctionImage
	for _, img := range images {
		if img.ID == imageID {
			imageToDelete = &img
			break
		}
	}

	if imageToDelete == nil {
		return domain.ErrNotFound
	}

	// Delete from storage
	_ = s.storage.Delete(ctx, imageToDelete.URL)

	// Delete from database
	return s.auctionImageRepo.Delete(ctx, imageID)
}

func (s *AuctionService) GetCategories(ctx context.Context) ([]domain.Category, error) {
	return s.categoryRepo.GetWithAuctionCounts(ctx)
}

func (s *AuctionService) GetCategoryBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return s.categoryRepo.GetBySlug(ctx, slug)
}

// Admin methods
func (s *AuctionService) AdminUpdateStatus(ctx context.Context, id uuid.UUID, status domain.AuctionStatus) error {
	auction, err := s.auctionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	auction.Status = status
	return s.auctionRepo.Update(ctx, auction)
}
