package service

import (
	"context"
	"fmt"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/pkg/email"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type NotificationService struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
	watchlistRepo    repository.WatchlistRepository
	emailSender      email.Sender
	baseURL          string
}

func NewNotificationService(
	notificationRepo repository.NotificationRepository,
	userRepo repository.UserRepository,
	watchlistRepo repository.WatchlistRepository,
	emailSender email.Sender,
	baseURL string,
) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		watchlistRepo:    watchlistRepo,
		emailSender:      emailSender,
		baseURL:          baseURL,
	}
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID, params *domain.NotificationListParams) (*domain.NotificationListResponse, error) {
	params.UserID = userID

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	notifications, totalCount, unreadCount, err := s.notificationRepo.GetByUserID(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + params.Limit - 1) / params.Limit

	return &domain.NotificationListResponse{
		Notifications: notifications,
		TotalCount:    totalCount,
		UnreadCount:   unreadCount,
		Page:          params.Page,
		TotalPages:    totalPages,
	}, nil
}

func (s *NotificationService) MarkAsRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		return err
	}

	// Verify ownership
	if notification.UserID != userID {
		return domain.ErrForbidden
	}

	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.notificationRepo.GetUnreadCount(ctx, userID)
}

// Notification creators

func (s *NotificationService) NotifyOutbid(ctx context.Context, userID uuid.UUID, auction *domain.Auction, newBidAmount decimal.Decimal) {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationOutbid,
		Title:     fmt.Sprintf("You've been outbid on %s", auction.Title),
		Message:   strPtr(fmt.Sprintf("A new bid of $%s has been placed. Place a higher bid to win!", newBidAmount.StringFixed(2))),
		AuctionID: &auction.ID,
	}

	_ = s.notificationRepo.Create(ctx, notification)

	// Send email
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		auctionURL := fmt.Sprintf("%s/auctions/%s", s.baseURL, auction.ID)
		emailData := email.NewOutbidEmail(user.Email, auction.Title, "$"+newBidAmount.StringFixed(2), auctionURL)
		_ = s.emailSender.Send(emailData)
	}
}

func (s *NotificationService) NotifyNewBid(ctx context.Context, sellerID uuid.UUID, auction *domain.Auction, bidAmount decimal.Decimal, bidderID uuid.UUID) {
	notification := &domain.Notification{
		UserID:    sellerID,
		Type:      domain.NotificationNewBid,
		Title:     fmt.Sprintf("New bid on %s", auction.Title),
		Message:   strPtr(fmt.Sprintf("A bid of $%s has been placed on your auction.", bidAmount.StringFixed(2))),
		AuctionID: &auction.ID,
	}

	_ = s.notificationRepo.Create(ctx, notification)

	// Send email
	seller, err := s.userRepo.GetByID(ctx, sellerID)
	if err == nil {
		bidder, _ := s.userRepo.GetByID(ctx, bidderID)
		bidderName := "Anonymous"
		if bidder != nil {
			bidderName = bidder.Username
		}
		auctionURL := fmt.Sprintf("%s/auctions/%s", s.baseURL, auction.ID)
		emailData := email.NewNewBidEmail(seller.Email, auction.Title, "$"+bidAmount.StringFixed(2), bidderName, auctionURL)
		_ = s.emailSender.Send(emailData)
	}
}

func (s *NotificationService) NotifyAuctionWon(ctx context.Context, winnerID uuid.UUID, auction *domain.Auction) {
	notification := &domain.Notification{
		UserID:    winnerID,
		Type:      domain.NotificationAuctionWon,
		Title:     fmt.Sprintf("Congratulations! You won %s", auction.Title),
		Message:   strPtr(fmt.Sprintf("You won the auction with a bid of $%s. The seller will contact you shortly.", auction.CurrentPrice.StringFixed(2))),
		AuctionID: &auction.ID,
	}

	_ = s.notificationRepo.Create(ctx, notification)

	// Send email
	user, err := s.userRepo.GetByID(ctx, winnerID)
	if err == nil {
		auctionURL := fmt.Sprintf("%s/auctions/%s", s.baseURL, auction.ID)
		emailData := email.NewAuctionWonEmail(user.Email, auction.Title, "$"+auction.CurrentPrice.StringFixed(2), auctionURL)
		_ = s.emailSender.Send(emailData)
	}
}

func (s *NotificationService) NotifyAuctionLost(ctx context.Context, userID uuid.UUID, auction *domain.Auction) {
	notification := &domain.Notification{
		UserID:    userID,
		Type:      domain.NotificationAuctionLost,
		Title:     fmt.Sprintf("Auction ended: %s", auction.Title),
		Message:   strPtr(fmt.Sprintf("The auction ended with a winning bid of $%s. Better luck next time!", auction.CurrentPrice.StringFixed(2))),
		AuctionID: &auction.ID,
	}

	_ = s.notificationRepo.Create(ctx, notification)

	// Send email
	user, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		auctionURL := fmt.Sprintf("%s/auctions/%s", s.baseURL, auction.ID)
		emailData := email.NewAuctionLostEmail(user.Email, auction.Title, "$"+auction.CurrentPrice.StringFixed(2), auctionURL)
		_ = s.emailSender.Send(emailData)
	}
}

func (s *NotificationService) NotifyAuctionSold(ctx context.Context, sellerID uuid.UUID, auction *domain.Auction, buyerID uuid.UUID) {
	notification := &domain.Notification{
		UserID:    sellerID,
		Type:      domain.NotificationAuctionSold,
		Title:     fmt.Sprintf("Your auction sold: %s", auction.Title),
		Message:   strPtr(fmt.Sprintf("Your item sold for $%s.", auction.CurrentPrice.StringFixed(2))),
		AuctionID: &auction.ID,
	}

	_ = s.notificationRepo.Create(ctx, notification)
}

func (s *NotificationService) NotifyAuctionEnding(ctx context.Context, auction *domain.Auction) {
	// Get all watchers
	watchers, err := s.watchlistRepo.GetWatchersForAuction(ctx, auction.ID)
	if err != nil {
		return
	}

	notifications := make([]domain.Notification, 0, len(watchers))
	for _, watcherID := range watchers {
		notifications = append(notifications, domain.Notification{
			UserID:    watcherID,
			Type:      domain.NotificationAuctionEnding,
			Title:     fmt.Sprintf("Auction ending soon: %s", auction.Title),
			Message:   strPtr(fmt.Sprintf("Current bid: $%s. Don't miss out!", auction.CurrentPrice.StringFixed(2))),
			AuctionID: &auction.ID,
		})
	}

	if len(notifications) > 0 {
		_ = s.notificationRepo.CreateBatch(ctx, notifications)
	}

	// Send emails to watchers
	for _, watcherID := range watchers {
		user, err := s.userRepo.GetByID(ctx, watcherID)
		if err != nil {
			continue
		}
		auctionURL := fmt.Sprintf("%s/auctions/%s", s.baseURL, auction.ID)
		emailData := email.NewAuctionEndingEmail(
			user.Email,
			auction.Title,
			"less than 1 hour",
			"$"+auction.CurrentPrice.StringFixed(2),
			auctionURL,
		)
		_ = s.emailSender.Send(emailData)
	}
}

func strPtr(s string) *string {
	return &s
}
