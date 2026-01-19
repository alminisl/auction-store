package domain

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
	Description *string    `json:"description,omitempty" db:"description"`
	ImageURL    *string    `json:"image_url,omitempty" db:"image_url"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`

	// Computed/joined fields
	AuctionCount int         `json:"auction_count,omitempty"`
	Children     []Category  `json:"children,omitempty"`
}

// Request DTOs
type CreateCategoryRequest struct {
	Name        string     `json:"name" validate:"required,min=2,max=100"`
	Slug        string     `json:"slug" validate:"required,min=2,max=100,alphanum"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	Description *string    `json:"description" validate:"omitempty,max=500"`
	ImageURL    *string    `json:"image_url" validate:"omitempty,url,max=500"`
}

type UpdateCategoryRequest struct {
	Name        *string    `json:"name" validate:"omitempty,min=2,max=100"`
	Slug        *string    `json:"slug" validate:"omitempty,min=2,max=100,alphanum"`
	ParentID    *uuid.UUID `json:"parent_id" validate:"omitempty,uuid"`
	Description *string    `json:"description" validate:"omitempty,max=500"`
	ImageURL    *string    `json:"image_url" validate:"omitempty,url,max=500"`
}
