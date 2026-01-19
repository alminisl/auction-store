package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	Email                  string     `json:"email" db:"email"`
	Username               string     `json:"username" db:"username"`
	PasswordHash           *string    `json:"-" db:"password_hash"`
	AvatarURL              *string    `json:"avatar_url" db:"avatar_url"`
	Bio                    *string    `json:"bio" db:"bio"`
	Phone                  *string    `json:"-" db:"phone"`
	Address                *string    `json:"-" db:"address"`
	Role                   UserRole   `json:"role" db:"role"`
	EmailVerified          bool       `json:"email_verified" db:"email_verified"`
	EmailVerificationToken *string    `json:"-" db:"email_verification_token"`
	PasswordResetToken     *string    `json:"-" db:"password_reset_token"`
	PasswordResetExpires   *time.Time `json:"-" db:"password_reset_expires"`
	IsBanned               bool       `json:"is_banned" db:"is_banned"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}

type PublicUser struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url"`
	Bio       *string   `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:        u.ID,
		Username:  u.Username,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		CreatedAt: u.CreatedAt,
	}
}

type OAuthAccount struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	Provider       string     `json:"provider" db:"provider"`
	ProviderUserID string     `json:"provider_user_id" db:"provider_user_id"`
	AccessToken    *string    `json:"-" db:"access_token"`
	RefreshToken   *string    `json:"-" db:"refresh_token"`
	ExpiresAt      *time.Time `json:"-" db:"expires_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

type RefreshToken struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Request/Response DTOs
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User        *User  `json:"user"`
	AccessToken string `json:"access_token"`
}

type UpdateProfileRequest struct {
	Username  *string `json:"username" validate:"omitempty,min=3,max=50,alphanum"`
	Bio       *string `json:"bio" validate:"omitempty,max=500"`
	Phone     *string `json:"phone" validate:"omitempty,max=20"`
	Address   *string `json:"address" validate:"omitempty,max=500"`
	AvatarURL *string `json:"avatar_url" validate:"omitempty,url,max=500"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}
