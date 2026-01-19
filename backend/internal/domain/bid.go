package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type Bid struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	AuctionID  uuid.UUID       `json:"auction_id" db:"auction_id"`
	BidderID   uuid.UUID       `json:"bidder_id" db:"bidder_id"`
	Amount     decimal.Decimal `json:"amount" db:"amount"`
	IsAutoBid  bool            `json:"is_auto_bid" db:"is_auto_bid"`
	MaxAutoBid *decimal.Decimal `json:"max_auto_bid,omitempty" db:"max_auto_bid"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`

	// Joined fields
	Bidder *PublicUser `json:"bidder,omitempty"`
}

// Request/Response DTOs
type PlaceBidRequest struct {
	Amount     string  `json:"amount" validate:"required,numeric,gt=0"`
	MaxAutoBid *string `json:"max_auto_bid" validate:"omitempty,numeric,gtefield=Amount"`
}

type BidResponse struct {
	Bid            *Bid            `json:"bid"`
	Auction        *Auction        `json:"auction"`
	AuctionExtended bool           `json:"auction_extended"`
	NewEndTime     *time.Time      `json:"new_end_time,omitempty"`
}

type BidListParams struct {
	AuctionID *uuid.UUID `json:"auction_id"`
	BidderID  *uuid.UUID `json:"bidder_id"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

type BidListResponse struct {
	Bids       []Bid `json:"bids"`
	TotalCount int   `json:"total_count"`
	Page       int   `json:"page"`
	TotalPages int   `json:"total_pages"`
}

// WebSocket messages
type WSMessageType string

const (
	WSMessageNewBid          WSMessageType = "new_bid"
	WSMessageAuctionExtended WSMessageType = "auction_extended"
	WSMessageAuctionEnded    WSMessageType = "auction_ended"
	WSMessageError           WSMessageType = "error"
)

type WSMessage struct {
	Type    WSMessageType `json:"type"`
	Payload interface{}   `json:"payload"`
}

type WSNewBidPayload struct {
	BidID      uuid.UUID       `json:"bid_id"`
	AuctionID  uuid.UUID       `json:"auction_id"`
	BidderID   uuid.UUID       `json:"bidder_id"`
	BidderName string          `json:"bidder_name"`
	Amount     decimal.Decimal `json:"amount"`
	BidCount   int             `json:"bid_count"`
	Timestamp  time.Time       `json:"timestamp"`
}

type WSAuctionExtendedPayload struct {
	AuctionID  uuid.UUID `json:"auction_id"`
	NewEndTime time.Time `json:"new_end_time"`
}

type WSAuctionEndedPayload struct {
	AuctionID   uuid.UUID        `json:"auction_id"`
	WinnerID    *uuid.UUID       `json:"winner_id"`
	WinnerName  *string          `json:"winner_name"`
	FinalPrice  decimal.Decimal  `json:"final_price"`
	Status      AuctionStatus    `json:"status"`
}
