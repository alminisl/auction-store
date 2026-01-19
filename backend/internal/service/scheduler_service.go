package service

import (
	"context"
	"log"
	"time"

	"github.com/auction-cards/backend/internal/cache"
	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/google/uuid"
)

type SchedulerService struct {
	auctionRepo     repository.AuctionRepository
	bidRepo         repository.BidRepository
	notificationSvc *NotificationService
	cache           *cache.RedisCache
	stopChan        chan struct{}
}

func NewSchedulerService(
	auctionRepo repository.AuctionRepository,
	bidRepo repository.BidRepository,
	notificationSvc *NotificationService,
	cache *cache.RedisCache,
) *SchedulerService {
	return &SchedulerService{
		auctionRepo:     auctionRepo,
		bidRepo:         bidRepo,
		notificationSvc: notificationSvc,
		cache:           cache,
		stopChan:        make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	go s.processEndingAuctions()
	go s.sendEndingSoonNotifications()
}

func (s *SchedulerService) Stop() {
	close(s.stopChan)
}

func (s *SchedulerService) processEndingAuctions() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkEndedAuctions()
		}
	}
}

func (s *SchedulerService) checkEndedAuctions() {
	ctx := context.Background()

	// Get auctions that have ended
	auctions, err := s.auctionRepo.GetEndingAuctions(ctx, time.Now().Unix())
	if err != nil {
		log.Printf("Error getting ending auctions: %v", err)
		return
	}

	for _, auction := range auctions {
		s.processAuctionEnd(ctx, &auction)
	}
}

func (s *SchedulerService) processAuctionEnd(ctx context.Context, auction *domain.Auction) {
	// Get highest bid
	highestBid, err := s.bidRepo.GetHighestBid(ctx, auction.ID)
	if err != nil {
		log.Printf("Error getting highest bid for auction %s: %v", auction.ID, err)
		return
	}

	var status domain.AuctionStatus
	var winnerID *uuid.UUID
	var winningBidID *uuid.UUID

	if highestBid != nil {
		// Check if reserve price was met
		if auction.ReservePrice != nil && highestBid.Amount.LessThan(*auction.ReservePrice) {
			status = domain.AuctionStatusUnsold
		} else {
			status = domain.AuctionStatusCompleted
			winnerID = &highestBid.BidderID
			winningBidID = &highestBid.ID
		}
	} else {
		status = domain.AuctionStatusUnsold
	}

	// Update auction status
	if err := s.auctionRepo.UpdateStatus(ctx, auction.ID, status, winnerID, winningBidID); err != nil {
		log.Printf("Error updating auction status %s: %v", auction.ID, err)
		return
	}

	// Publish auction ended message
	if s.cache != nil {
		var winnerName *string
		message := domain.WSMessage{
			Type: domain.WSMessageAuctionEnded,
			Payload: domain.WSAuctionEndedPayload{
				AuctionID:  auction.ID,
				WinnerID:   winnerID,
				WinnerName: winnerName,
				FinalPrice: auction.CurrentPrice,
				Status:     status,
			},
		}
		_ = s.cache.Publish(ctx, cache.AuctionChannel(auction.ID), message)
	}

	// Send notifications
	if s.notificationSvc != nil {
		if status == domain.AuctionStatusCompleted && winnerID != nil {
			// Notify winner
			s.notificationSvc.NotifyAuctionWon(ctx, *winnerID, auction)

			// Notify seller
			s.notificationSvc.NotifyAuctionSold(ctx, auction.SellerID, auction, *winnerID)

			// Notify other bidders they lost
			s.notifyLosingBidders(ctx, auction, *winnerID)
		}
	}

	log.Printf("Processed auction end: %s, status: %s", auction.ID, status)
}

func (s *SchedulerService) notifyLosingBidders(ctx context.Context, auction *domain.Auction, winnerID uuid.UUID) {
	// Get all bids and notify unique bidders (except winner)
	bids, _, err := s.bidRepo.GetByAuctionID(ctx, auction.ID, 1, 1000) // Get all bids
	if err != nil {
		return
	}

	notifiedBidders := make(map[uuid.UUID]bool)
	notifiedBidders[winnerID] = true // Don't notify winner

	for _, bid := range bids {
		if notifiedBidders[bid.BidderID] {
			continue
		}
		notifiedBidders[bid.BidderID] = true
		s.notificationSvc.NotifyAuctionLost(ctx, bid.BidderID, auction)
	}
}

func (s *SchedulerService) sendEndingSoonNotifications() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAuctionsEndingSoon()
		}
	}
}

func (s *SchedulerService) checkAuctionsEndingSoon() {
	ctx := context.Background()

	// Get auctions ending in the next hour
	oneHourFromNow := time.Now().Add(1 * time.Hour).Unix()

	auctions, err := s.auctionRepo.GetEndingAuctions(ctx, oneHourFromNow)
	if err != nil {
		log.Printf("Error getting auctions ending soon: %v", err)
		return
	}

	for _, auction := range auctions {
		// Only notify for auctions that haven't ended yet
		if auction.EndTime.After(time.Now()) && auction.Status == domain.AuctionStatusActive {
			s.notificationSvc.NotifyAuctionEnding(ctx, &auction)
		}
	}
}
