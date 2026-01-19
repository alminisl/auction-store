package domain

import (
	"time"

	"github.com/google/uuid"
)

type RatingType string

const (
	RatingTypeBuyer  RatingType = "buyer"
	RatingTypeSeller RatingType = "seller"
)

type Rating struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	AuctionID   uuid.UUID  `json:"auction_id" db:"auction_id"`
	RaterID     uuid.UUID  `json:"rater_id" db:"rater_id"`
	RatedUserID uuid.UUID  `json:"rated_user_id" db:"rated_user_id"`
	Rating      int        `json:"rating" db:"rating"`
	Comment     *string    `json:"comment,omitempty" db:"comment"`
	Type        RatingType `json:"type" db:"type"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`

	// Joined fields
	Rater     *PublicUser `json:"rater,omitempty"`
	RatedUser *PublicUser `json:"rated_user,omitempty"`
	Auction   *Auction    `json:"auction,omitempty"`
}

type UserRatingSummary struct {
	UserID        uuid.UUID `json:"user_id"`
	AverageRating float64   `json:"average_rating"`
	TotalRatings  int       `json:"total_ratings"`
	SellerRating  float64   `json:"seller_rating"`
	SellerCount   int       `json:"seller_count"`
	BuyerRating   float64   `json:"buyer_rating"`
	BuyerCount    int       `json:"buyer_count"`
}

// Request DTOs
type CreateRatingRequest struct {
	Rating  int     `json:"rating" validate:"required,min=1,max=5"`
	Comment *string `json:"comment" validate:"omitempty,max=1000"`
}

type RatingListParams struct {
	RatedUserID *uuid.UUID  `json:"rated_user_id"`
	RaterID     *uuid.UUID  `json:"rater_id"`
	Type        *RatingType `json:"type"`
	Page        int         `json:"page"`
	Limit       int         `json:"limit"`
}

type RatingListResponse struct {
	Ratings    []Rating `json:"ratings"`
	TotalCount int      `json:"total_count"`
	Page       int      `json:"page"`
	TotalPages int      `json:"total_pages"`
}
