package service

import (
	"context"
	"time"

	"github.com/auction-cards/backend/internal/cache"
	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/auction-cards/backend/internal/repository/postgres"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	AntiSnipingWindow   = 5 * time.Minute  // Extend if bid in last 5 minutes
	AntiSnipingExtend   = 2 * time.Minute  // Extend by 2 minutes
)

type BidService struct {
	bidRepo         repository.BidRepository
	auctionRepo     repository.AuctionRepository
	bidTransaction  *postgres.BidTransaction
	notificationSvc *NotificationService
	cache           *cache.RedisCache
}

func NewBidService(
	bidRepo repository.BidRepository,
	auctionRepo repository.AuctionRepository,
	bidTransaction *postgres.BidTransaction,
	notificationSvc *NotificationService,
	cache *cache.RedisCache,
) *BidService {
	return &BidService{
		bidRepo:         bidRepo,
		auctionRepo:     auctionRepo,
		bidTransaction:  bidTransaction,
		notificationSvc: notificationSvc,
		cache:           cache,
	}
}

func (s *BidService) PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, req *domain.PlaceBidRequest) (*domain.BidResponse, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, domain.ErrBadRequest
	}

	var maxAutoBid *decimal.Decimal
	if req.MaxAutoBid != nil {
		max, err := decimal.NewFromString(*req.MaxAutoBid)
		if err != nil {
			return nil, domain.ErrBadRequest
		}
		maxAutoBid = &max
	}

	// Use transaction for atomic bid placement
	result, err := s.placeBidWithTransaction(ctx, auctionID, bidderID, amount, maxAutoBid)
	if err != nil {
		return nil, err
	}

	// Publish bid to Redis for WebSocket broadcast
	s.publishBidUpdate(ctx, result)

	// Send notifications asynchronously
	go s.sendBidNotifications(context.Background(), result, bidderID)

	response := &domain.BidResponse{
		Bid:             result.Bid,
		Auction:         result.Auction,
		AuctionExtended: result.AuctionExtended,
	}

	if result.NewEndTime != nil {
		endTime := time.Unix(*result.NewEndTime, 0)
		response.NewEndTime = &endTime
	}

	return response, nil
}

func (s *BidService) placeBidWithTransaction(ctx context.Context, auctionID, bidderID uuid.UUID, amount decimal.Decimal, maxAutoBid *decimal.Decimal) (*postgres.PlaceBidResult, error) {
	// Get auction first to validate
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Validate auction is active
	if auction.Status != domain.AuctionStatusActive {
		return nil, domain.ErrAuctionNotActive
	}

	// Check auction hasn't ended
	if time.Now().After(auction.EndTime) {
		return nil, domain.ErrAuctionEnded
	}

	// Validate not self-bidding
	if auction.SellerID == bidderID {
		return nil, domain.ErrSelfBidding
	}

	// Validate bid amount
	minBid := auction.CurrentPrice.Add(auction.BidIncrement)
	if amount.LessThan(minBid) {
		return nil, domain.ErrBidTooLow
	}

	// Get previous high bidder for outbid notification
	prevBid, _ := s.bidRepo.GetHighestBid(ctx, auctionID)
	var prevBidderID *uuid.UUID
	if prevBid != nil && prevBid.BidderID != bidderID {
		prevBidderID = &prevBid.BidderID
	}

	// Create bid
	bid := &domain.Bid{
		ID:         uuid.New(),
		AuctionID:  auctionID,
		BidderID:   bidderID,
		Amount:     amount,
		IsAutoBid:  maxAutoBid != nil,
		MaxAutoBid: maxAutoBid,
		CreatedAt:  time.Now(),
	}

	// Check for anti-sniping (bid in last 5 minutes)
	auctionExtended := false
	var newEndTime *int64
	timeUntilEnd := auction.EndTime.Sub(time.Now())
	if timeUntilEnd < AntiSnipingWindow && timeUntilEnd > 0 {
		// Extend by 2 minutes
		extendedTime := auction.EndTime.Add(AntiSnipingExtend)
		auction.EndTime = extendedTime
		auctionExtended = true
		endTimeUnix := extendedTime.Unix()
		newEndTime = &endTimeUnix
	}

	// Update auction
	auction.CurrentPrice = amount
	auction.BidCount++
	expectedVersion := auction.Version

	// Save bid
	if err := s.bidRepo.Create(ctx, bid); err != nil {
		return nil, err
	}

	// Update auction with version check
	if err := s.auctionRepo.UpdateWithVersion(ctx, auction, expectedVersion); err != nil {
		return nil, err
	}

	return &postgres.PlaceBidResult{
		Bid:             bid,
		Auction:         auction,
		AuctionExtended: auctionExtended,
		NewEndTime:      newEndTime,
		PreviousBidder:  prevBidderID,
	}, nil
}

func (s *BidService) publishBidUpdate(ctx context.Context, result *postgres.PlaceBidResult) {
	if s.cache == nil {
		return
	}

	message := domain.WSMessage{
		Type: domain.WSMessageNewBid,
		Payload: domain.WSNewBidPayload{
			BidID:      result.Bid.ID,
			AuctionID:  result.Bid.AuctionID,
			BidderID:   result.Bid.BidderID,
			Amount:     result.Bid.Amount,
			BidCount:   result.Auction.BidCount,
			Timestamp:  result.Bid.CreatedAt,
		},
	}

	_ = s.cache.Publish(ctx, cache.AuctionChannel(result.Auction.ID), message)

	if result.AuctionExtended && result.NewEndTime != nil {
		extendMessage := domain.WSMessage{
			Type: domain.WSMessageAuctionExtended,
			Payload: domain.WSAuctionExtendedPayload{
				AuctionID:  result.Auction.ID,
				NewEndTime: time.Unix(*result.NewEndTime, 0),
			},
		}
		_ = s.cache.Publish(ctx, cache.AuctionChannel(result.Auction.ID), extendMessage)
	}
}

func (s *BidService) sendBidNotifications(ctx context.Context, result *postgres.PlaceBidResult, bidderID uuid.UUID) {
	if s.notificationSvc == nil {
		return
	}

	// Notify previous high bidder they've been outbid
	if result.PreviousBidder != nil {
		s.notificationSvc.NotifyOutbid(ctx, *result.PreviousBidder, result.Auction, result.Bid.Amount)
	}

	// Notify seller of new bid
	s.notificationSvc.NotifyNewBid(ctx, result.Auction.SellerID, result.Auction, result.Bid.Amount, bidderID)
}

func (s *BidService) GetBidsByAuction(ctx context.Context, auctionID uuid.UUID, page, limit int) (*domain.BidListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	bids, totalCount, err := s.bidRepo.GetByAuctionID(ctx, auctionID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + limit - 1) / limit

	return &domain.BidListResponse{
		Bids:       bids,
		TotalCount: totalCount,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *BidService) GetBidsByUser(ctx context.Context, userID uuid.UUID, page, limit int) (*domain.BidListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	bids, totalCount, err := s.bidRepo.GetByBidderID(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + limit - 1) / limit

	return &domain.BidListResponse{
		Bids:       bids,
		TotalCount: totalCount,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *BidService) BuyNow(ctx context.Context, auctionID, buyerID uuid.UUID) (*domain.BidResponse, error) {
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Validate auction is active
	if auction.Status != domain.AuctionStatusActive {
		return nil, domain.ErrAuctionNotActive
	}

	// Check auction hasn't ended
	if time.Now().After(auction.EndTime) {
		return nil, domain.ErrAuctionEnded
	}

	// Validate not self-bidding
	if auction.SellerID == buyerID {
		return nil, domain.ErrSelfBidding
	}

	// Check if Buy Now is available
	if auction.BuyNowPrice == nil {
		return nil, domain.ErrBadRequest
	}

	// Create bid at buy now price
	bid := &domain.Bid{
		ID:        uuid.New(),
		AuctionID: auctionID,
		BidderID:  buyerID,
		Amount:    *auction.BuyNowPrice,
		CreatedAt: time.Now(),
	}

	if err := s.bidRepo.Create(ctx, bid); err != nil {
		return nil, err
	}

	// End auction immediately
	auction.Status = domain.AuctionStatusCompleted
	auction.CurrentPrice = *auction.BuyNowPrice
	auction.WinnerID = &buyerID
	auction.WinningBidID = &bid.ID
	auction.EndTime = time.Now()
	auction.BidCount++

	if err := s.auctionRepo.Update(ctx, auction); err != nil {
		return nil, err
	}

	// Publish auction ended
	if s.cache != nil {
		message := domain.WSMessage{
			Type: domain.WSMessageAuctionEnded,
			Payload: domain.WSAuctionEndedPayload{
				AuctionID:  auction.ID,
				WinnerID:   auction.WinnerID,
				FinalPrice: auction.CurrentPrice,
				Status:     auction.Status,
			},
		}
		_ = s.cache.Publish(ctx, cache.AuctionChannel(auction.ID), message)
	}

	// Send notifications
	if s.notificationSvc != nil {
		go func() {
			s.notificationSvc.NotifyAuctionWon(context.Background(), buyerID, auction)
			s.notificationSvc.NotifyAuctionSold(context.Background(), auction.SellerID, auction, buyerID)
		}()
	}

	return &domain.BidResponse{
		Bid:     bid,
		Auction: auction,
	}, nil
}
