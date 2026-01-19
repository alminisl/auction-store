package repository

import (
	"context"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByVerificationToken(ctx context.Context, token string) (*domain.User, error)
	GetByPasswordResetToken(ctx context.Context, token string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, limit int) ([]domain.User, int, error)
	GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.UserRatingSummary, error)
}

type OAuthAccountRepository interface {
	Create(ctx context.Context, account *domain.OAuthAccount) error
	GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.OAuthAccount, error)
	Update(ctx context.Context, account *domain.OAuthAccount) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

type AuctionRepository interface {
	Create(ctx context.Context, auction *domain.Auction) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Auction, error)
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.Auction, error)
	Update(ctx context.Context, auction *domain.Auction) error
	UpdateWithVersion(ctx context.Context, auction *domain.Auction, expectedVersion int) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, params *domain.AuctionListParams) ([]domain.Auction, int, error)
	GetEndingAuctions(ctx context.Context, before int64) ([]domain.Auction, error)
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AuctionStatus, winnerID *uuid.UUID, winningBidID *uuid.UUID) error
}

type AuctionImageRepository interface {
	Create(ctx context.Context, image *domain.AuctionImage) error
	GetByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]domain.AuctionImage, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByAuctionID(ctx context.Context, auctionID uuid.UUID) error
	UpdatePositions(ctx context.Context, auctionID uuid.UUID, positions map[uuid.UUID]int) error
}

type BidRepository interface {
	Create(ctx context.Context, bid *domain.Bid) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Bid, error)
	GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*domain.Bid, error)
	GetByAuctionID(ctx context.Context, auctionID uuid.UUID, page, limit int) ([]domain.Bid, int, error)
	GetByBidderID(ctx context.Context, bidderID uuid.UUID, page, limit int) ([]domain.Bid, int, error)
	GetBidCount(ctx context.Context, auctionID uuid.UUID) (int, error)
	GetPreviousHighBidder(ctx context.Context, auctionID uuid.UUID, excludeBidderID uuid.UUID) (*domain.Bid, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]domain.Category, error)
	GetWithAuctionCounts(ctx context.Context) ([]domain.Category, error)
}

type WatchlistRepository interface {
	Add(ctx context.Context, item *domain.WatchlistItem) error
	Remove(ctx context.Context, userID, auctionID uuid.UUID) error
	GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.WatchlistItem, int, error)
	Exists(ctx context.Context, userID, auctionID uuid.UUID) (bool, error)
	GetWatchersForAuction(ctx context.Context, auctionID uuid.UUID) ([]uuid.UUID, error)
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	CreateBatch(ctx context.Context, notifications []domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, params *domain.NotificationListParams) ([]domain.Notification, int, int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
}

type RatingRepository interface {
	Create(ctx context.Context, rating *domain.Rating) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Rating, error)
	GetByAuctionAndRater(ctx context.Context, auctionID, raterID uuid.UUID, ratingType domain.RatingType) (*domain.Rating, error)
	GetByRatedUser(ctx context.Context, ratedUserID uuid.UUID, params *domain.RatingListParams) ([]domain.Rating, int, error)
	GetUserRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.UserRatingSummary, error)
}

type ReportRepository interface {
	Create(ctx context.Context, report *domain.ReportedListing) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportedListing, error)
	Update(ctx context.Context, report *domain.ReportedListing) error
	List(ctx context.Context, params *domain.ReportListParams) ([]domain.ReportedListing, int, error)
}

// Transaction support
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// PlaceBidResult holds the result of placing a bid
type PlaceBidResult struct {
	Bid             *domain.Bid
	Auction         *domain.Auction
	AuctionExtended bool
	NewEndTime      *int64
	PreviousBidder  *uuid.UUID
}

// BidTransaction handles atomic bid placement
type BidTransaction interface {
	PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount decimal.Decimal, maxAutoBid *decimal.Decimal) (*PlaceBidResult, error)
}
