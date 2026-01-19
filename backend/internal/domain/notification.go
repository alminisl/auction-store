package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationOutbid        NotificationType = "outbid"
	NotificationAuctionWon    NotificationType = "auction_won"
	NotificationAuctionLost   NotificationType = "auction_lost"
	NotificationAuctionEnding NotificationType = "auction_ending"
	NotificationNewBid        NotificationType = "new_bid"
	NotificationAuctionSold   NotificationType = "auction_sold"
)

type Notification struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	UserID    uuid.UUID        `json:"user_id" db:"user_id"`
	Type      NotificationType `json:"type" db:"type"`
	Title     string           `json:"title" db:"title"`
	Message   *string          `json:"message,omitempty" db:"message"`
	AuctionID *uuid.UUID       `json:"auction_id,omitempty" db:"auction_id"`
	IsRead    bool             `json:"is_read" db:"is_read"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`

	// Joined fields
	Auction *Auction `json:"auction,omitempty"`
}

type NotificationListParams struct {
	UserID   uuid.UUID `json:"user_id"`
	Unread   *bool     `json:"unread"`
	Page     int       `json:"page"`
	Limit    int       `json:"limit"`
}

type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	TotalCount    int            `json:"total_count"`
	UnreadCount   int            `json:"unread_count"`
	Page          int            `json:"page"`
	TotalPages    int            `json:"total_pages"`
}
