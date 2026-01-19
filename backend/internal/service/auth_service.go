package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/pkg/email"
	"github.com/auction-cards/backend/internal/pkg/jwt"
	"github.com/auction-cards/backend/internal/pkg/password"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo         repository.UserRepository
	oauthRepo        repository.OAuthAccountRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtManager       *jwt.Manager
	emailSender      email.Sender
	baseURL          string
}

func NewAuthService(
	userRepo repository.UserRepository,
	oauthRepo repository.OAuthAccountRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtManager *jwt.Manager,
	emailSender email.Sender,
	baseURL string,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		oauthRepo:        oauthRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtManager:       jwtManager,
		emailSender:      emailSender,
		baseURL:          baseURL,
	}
}

func (s *AuthService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.User, error) {
	// Check if email exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, domain.ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// Check if username exists
	existingUser, err = s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, domain.ErrUsernameExists
	}
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	verificationToken := generateToken()

	// Create user
	user := &domain.User{
		Email:                  req.Email,
		Username:               req.Username,
		PasswordHash:           &hashedPassword,
		Role:                   domain.RoleUser,
		EmailVerified:          false,
		EmailVerificationToken: &verificationToken,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send verification email
	emailData := email.NewVerificationEmail(user.Email, verificationToken, s.baseURL)
	_ = s.emailSender.Send(emailData)

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, string, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, "", domain.ErrInvalidCredentials
		}
		return nil, "", err
	}

	// Check if user has a password (might be OAuth-only user)
	if user.PasswordHash == nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Verify password
	if !password.Verify(req.Password, *user.PasswordHash) {
		return nil, "", domain.ErrInvalidCredentials
	}

	// Check if banned
	if user.IsBanned {
		return nil, "", domain.ErrUserBanned
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	refreshToken, expiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	// Store refresh token
	tokenHash := hashToken(refreshToken)
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, "", err
	}

	return &domain.AuthResponse{
		User:        user,
		AccessToken: accessToken,
	}, refreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := hashToken(refreshToken)
	return s.refreshTokenRepo.DeleteByTokenHash(ctx, tokenHash)
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	// Validate refresh token
	userID, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Check if token exists in database
	tokenHash := hashToken(refreshToken)
	storedToken, err := s.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return "", domain.ErrTokenInvalid
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil {
		return "", err
	}

	if user.IsBanned {
		return "", domain.ErrUserBanned
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(userID, string(user.Role))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, token string) error {
	user, err := s.userRepo.GetByVerificationToken(ctx, token)
	if err != nil {
		return domain.ErrTokenInvalid
	}

	user.EmailVerified = true
	user.EmailVerificationToken = nil

	return s.userRepo.Update(ctx, user)
}

func (s *AuthService) ForgotPassword(ctx context.Context, req *domain.ForgotPasswordRequest) error {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists
		return nil
	}

	// Generate reset token
	resetToken := generateToken()
	expires := time.Now().Add(1 * time.Hour)

	user.PasswordResetToken = &resetToken
	user.PasswordResetExpires = &expires

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Send reset email
	emailData := email.NewPasswordResetEmail(user.Email, resetToken, s.baseURL)
	_ = s.emailSender.Send(emailData)

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, req *domain.ResetPasswordRequest) error {
	user, err := s.userRepo.GetByPasswordResetToken(ctx, req.Token)
	if err != nil {
		return domain.ErrTokenInvalid
	}

	// Hash new password
	hashedPassword, err := password.Hash(req.Password)
	if err != nil {
		return err
	}

	user.PasswordHash = &hashedPassword
	user.PasswordResetToken = nil
	user.PasswordResetExpires = nil

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Revoke all refresh tokens
	return s.refreshTokenRepo.DeleteByUserID(ctx, user.ID)
}

func (s *AuthService) GetOrCreateOAuthUser(ctx context.Context, provider, providerUserID, email, username string) (*domain.User, error) {
	// Check if OAuth account exists
	oauthAccount, err := s.oauthRepo.GetByProviderUserID(ctx, provider, providerUserID)
	if err == nil {
		// Get existing user
		return s.userRepo.GetByID(ctx, oauthAccount.UserID)
	}

	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	// Check if user with email exists
	user, err := s.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		// Create new user
		user = &domain.User{
			Email:         email,
			Username:      username,
			Role:          domain.RoleUser,
			EmailVerified: true, // OAuth emails are pre-verified
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	// Create OAuth account link
	oauthAccount = &domain.OAuthAccount{
		UserID:         user.ID,
		Provider:       provider,
		ProviderUserID: providerUserID,
	}

	if err := s.oauthRepo.Create(ctx, oauthAccount); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) GenerateTokens(ctx context.Context, user *domain.User) (*domain.AuthResponse, string, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", err
	}

	refreshToken, expiresAt, err := s.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	// Store refresh token
	tokenHash := hashToken(refreshToken)
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, "", err
	}

	return &domain.AuthResponse{
		User:        user,
		AccessToken: accessToken,
	}, refreshToken, nil
}

func (s *AuthService) ValidateAccessToken(tokenString string) (*jwt.Claims, error) {
	return s.jwtManager.ValidateAccessToken(tokenString)
}

func (s *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// Helper functions
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
