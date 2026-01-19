package domain

import (
	"time"

	"github.com/google/uuid"
)

type WatchlistItem struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	AuctionID uuid.UUID `json:"auction_id" db:"auction_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Joined fields
	Auction *Auction `json:"auction,omitempty"`
}

type WatchlistResponse struct {
	Items      []WatchlistItem `json:"items"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	TotalPages int             `json:"total_pages"`
}

type ReportReason string

const (
	ReportReasonFraud        ReportReason = "fraud"
	ReportReasonProhibited   ReportReason = "prohibited"
	ReportReasonCounterfeit  ReportReason = "counterfeit"
	ReportReasonMisleading   ReportReason = "misleading"
	ReportReasonInappropriate ReportReason = "inappropriate"
	ReportReasonOther        ReportReason = "other"
)

type ReportStatus string

const (
	ReportStatusPending  ReportStatus = "pending"
	ReportStatusReviewed ReportStatus = "reviewed"
	ReportStatusResolved ReportStatus = "resolved"
)

type ReportedListing struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	AuctionID   uuid.UUID    `json:"auction_id" db:"auction_id"`
	ReporterID  uuid.UUID    `json:"reporter_id" db:"reporter_id"`
	Reason      ReportReason `json:"reason" db:"reason"`
	Description *string      `json:"description,omitempty" db:"description"`
	Status      ReportStatus `json:"status" db:"status"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`

	// Joined fields
	Auction  *Auction    `json:"auction,omitempty"`
	Reporter *PublicUser `json:"reporter,omitempty"`
}

type CreateReportRequest struct {
	Reason      string  `json:"reason" validate:"required,oneof=fraud prohibited counterfeit misleading inappropriate other"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
}

type UpdateReportRequest struct {
	Status ReportStatus `json:"status" validate:"required,oneof=pending reviewed resolved"`
}

type ReportListParams struct {
	Status *ReportStatus `json:"status"`
	Page   int           `json:"page"`
	Limit  int           `json:"limit"`
}

type ReportListResponse struct {
	Reports    []ReportedListing `json:"reports"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	TotalPages int               `json:"total_pages"`
}
