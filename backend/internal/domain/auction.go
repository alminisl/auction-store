package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AuctionStatus string

const (
	AuctionStatusDraft     AuctionStatus = "draft"
	AuctionStatusActive    AuctionStatus = "active"
	AuctionStatusCompleted AuctionStatus = "completed"
	AuctionStatusCancelled AuctionStatus = "cancelled"
	AuctionStatusUnsold    AuctionStatus = "unsold"
)

type ItemCondition string

const (
	ConditionNew     ItemCondition = "new"
	ConditionLikeNew ItemCondition = "like_new"
	ConditionGood    ItemCondition = "good"
	ConditionFair    ItemCondition = "fair"
	ConditionPoor    ItemCondition = "poor"
)

type Auction struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	SellerID      uuid.UUID       `json:"seller_id" db:"seller_id"`
	CategoryID    *uuid.UUID      `json:"category_id" db:"category_id"`
	Title         string          `json:"title" db:"title"`
	Description   *string         `json:"description" db:"description"`
	Condition     *ItemCondition  `json:"condition" db:"condition"`
	StartingPrice decimal.Decimal `json:"starting_price" db:"starting_price"`
	ReservePrice  *decimal.Decimal `json:"reserve_price,omitempty" db:"reserve_price"`
	BuyNowPrice   *decimal.Decimal `json:"buy_now_price,omitempty" db:"buy_now_price"`
	CurrentPrice  decimal.Decimal `json:"current_price" db:"current_price"`
	BidIncrement  decimal.Decimal `json:"bid_increment" db:"bid_increment"`
	StartTime     time.Time       `json:"start_time" db:"start_time"`
	EndTime       time.Time       `json:"end_time" db:"end_time"`
	Status        AuctionStatus   `json:"status" db:"status"`
	WinnerID      *uuid.UUID      `json:"winner_id,omitempty" db:"winner_id"`
	WinningBidID  *uuid.UUID      `json:"winning_bid_id,omitempty" db:"winning_bid_id"`
	ViewsCount    int             `json:"views_count" db:"views_count"`
	BidCount      int             `json:"bid_count" db:"bid_count"`
	Version       int             `json:"-" db:"version"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at" db:"updated_at"`

	// Joined fields
	Seller   *PublicUser      `json:"seller,omitempty"`
	Category *Category        `json:"category,omitempty"`
	Images   []AuctionImage   `json:"images,omitempty"`
	Winner   *PublicUser      `json:"winner,omitempty"`
}

type AuctionImage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	AuctionID uuid.UUID `json:"auction_id" db:"auction_id"`
	URL       string    `json:"url" db:"url"`
	Position  int       `json:"position" db:"position"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Request/Response DTOs
type CreateAuctionRequest struct {
	CategoryID    *uuid.UUID `json:"category_id" validate:"omitempty,uuid"`
	Title         string     `json:"title" validate:"required,min=3,max=255"`
	Description   *string    `json:"description" validate:"omitempty,max=5000"`
	Condition     *string    `json:"condition" validate:"omitempty,oneof=new like_new good fair poor"`
	StartingPrice string     `json:"starting_price" validate:"required,numeric,gt=0"`
	ReservePrice  *string    `json:"reserve_price" validate:"omitempty,numeric,gtefield=StartingPrice"`
	BuyNowPrice   *string    `json:"buy_now_price" validate:"omitempty,numeric,gtefield=StartingPrice"`
	BidIncrement  *string    `json:"bid_increment" validate:"omitempty,numeric,gt=0"`
	StartTime     time.Time  `json:"start_time" validate:"required"`
	EndTime       time.Time  `json:"end_time" validate:"required,gtfield=StartTime"`
}

type UpdateAuctionRequest struct {
	CategoryID    *uuid.UUID `json:"category_id" validate:"omitempty,uuid"`
	Title         *string    `json:"title" validate:"omitempty,min=3,max=255"`
	Description   *string    `json:"description" validate:"omitempty,max=5000"`
	Condition     *string    `json:"condition" validate:"omitempty,oneof=new like_new good fair poor"`
	StartingPrice *string    `json:"starting_price" validate:"omitempty,numeric,gt=0"`
	ReservePrice  *string    `json:"reserve_price" validate:"omitempty,numeric"`
	BuyNowPrice   *string    `json:"buy_now_price" validate:"omitempty,numeric"`
	BidIncrement  *string    `json:"bid_increment" validate:"omitempty,numeric,gt=0"`
	StartTime     *time.Time `json:"start_time"`
	EndTime       *time.Time `json:"end_time"`
}

type AuctionListParams struct {
	Status     *AuctionStatus `json:"status"`
	CategoryID *uuid.UUID     `json:"category_id"`
	SellerID   *uuid.UUID     `json:"seller_id"`
	Search     *string        `json:"search"`
	MinPrice   *decimal.Decimal `json:"min_price"`
	MaxPrice   *decimal.Decimal `json:"max_price"`
	SortBy     string         `json:"sort_by"` // ending_soon, newest, price_low, price_high, most_bids
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
}

type AuctionListResponse struct {
	Auctions   []Auction `json:"auctions"`
	TotalCount int       `json:"total_count"`
	Page       int       `json:"page"`
	TotalPages int       `json:"total_pages"`
}
