package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/auction-cards/backend/internal/config"
	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/handler"
	"github.com/auction-cards/backend/internal/middleware"
	"github.com/auction-cards/backend/internal/pkg/email"
	"github.com/auction-cards/backend/internal/pkg/jwt"
	"github.com/auction-cards/backend/internal/pkg/password"
	"github.com/auction-cards/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Test helpers
func newTestJWTManager() *jwt.Manager {
	return jwt.NewManager(
		"test-access-secret",
		"test-refresh-secret",
		15*time.Minute,
		7*24*time.Hour,
	)
}

func createTestRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	return r
}

func makeRequest(t *testing.T, r *chi.Mux, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	return rr
}

func parseResponse(t *testing.T, rr *httptest.ResponseRecorder) *domain.APIResponse {
	var response domain.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	return &response
}

// Mock implementations for testing

type mockUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[uuid.UUID]*domain.User),
	}
}

func (r *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *mockUserRepo) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *mockUserRepo) GetByVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	for _, user := range r.users {
		if user.EmailVerificationToken != nil && *user.EmailVerificationToken == token {
			return user, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *mockUserRepo) GetByPasswordResetToken(ctx context.Context, token string) (*domain.User, error) {
	for _, user := range r.users {
		if user.PasswordResetToken != nil && *user.PasswordResetToken == token {
			return user, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	user.UpdatedAt = time.Now()
	r.users[user.ID] = user
	return nil
}

func (r *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

func (r *mockUserRepo) List(ctx context.Context, page, limit int) ([]domain.User, int, error) {
	users := make([]domain.User, 0)
	for _, user := range r.users {
		users = append(users, *user)
	}
	return users, len(users), nil
}

func (r *mockUserRepo) GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.UserRatingSummary, error) {
	return &domain.UserRatingSummary{UserID: userID}, nil
}

type mockOAuthRepo struct{}

func (r *mockOAuthRepo) Create(ctx context.Context, account *domain.OAuthAccount) error {
	return nil
}

func (r *mockOAuthRepo) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error) {
	return nil, domain.ErrNotFound
}

func (r *mockOAuthRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.OAuthAccount, error) {
	return nil, nil
}

func (r *mockOAuthRepo) Update(ctx context.Context, account *domain.OAuthAccount) error {
	return nil
}

func (r *mockOAuthRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockRefreshTokenRepo struct {
	tokens map[string]*domain.RefreshToken
}

func newMockRefreshTokenRepo() *mockRefreshTokenRepo {
	return &mockRefreshTokenRepo{
		tokens: make(map[string]*domain.RefreshToken),
	}
}

func (r *mockRefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	r.tokens[token.TokenHash] = token
	return nil
}

func (r *mockRefreshTokenRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	if token, ok := r.tokens[tokenHash]; ok {
		return token, nil
	}
	return nil, domain.ErrNotFound
}

func (r *mockRefreshTokenRepo) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	delete(r.tokens, tokenHash)
	return nil
}

func (r *mockRefreshTokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	for hash, token := range r.tokens {
		if token.UserID == userID {
			delete(r.tokens, hash)
		}
	}
	return nil
}

func (r *mockRefreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return nil
}

type mockEmailSender struct {
	sentEmails []string
}

func (s *mockEmailSender) Send(data *email.EmailData) error {
	s.sentEmails = append(s.sentEmails, data.To)
	return nil
}

// Tests

func TestAuthHandler_Register(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtManager := newTestJWTManager()
	emailSender := &mockEmailSender{}

	authService := service.NewAuthService(
		userRepo,
		&mockOAuthRepo{},
		newMockRefreshTokenRepo(),
		jwtManager,
		emailSender,
		"http://localhost:5173",
	)

	r := createTestRouter()
	cfg := &config.Config{
		Server: config.ServerConfig{
			AllowOrigins: []string{"http://localhost:5173"},
		},
	}
	authHandler := handler.NewAuthHandler(authService, cfg)

	r.Post("/api/auth/register", authHandler.Register)

	tests := []struct {
		name       string
		body       domain.RegisterRequest
		wantStatus int
		wantErr    bool
	}{
		{
			name: "successful registration",
			body: domain.RegisterRequest{
				Email:    "test@example.com",
				Username: "testuser",
				Password: "Password123",
			},
			wantStatus: http.StatusCreated,
			wantErr:    false,
		},
		{
			name: "invalid email",
			body: domain.RegisterRequest{
				Email:    "invalid-email",
				Username: "testuser2",
				Password: "Password123",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "short password",
			body: domain.RegisterRequest{
				Email:    "test2@example.com",
				Username: "testuser3",
				Password: "short",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "missing username",
			body: domain.RegisterRequest{
				Email:    "test3@example.com",
				Password: "Password123",
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "POST", "/api/auth/register", tt.body, "")

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			response := parseResponse(t, rr)
			if tt.wantErr && response.Success {
				t.Errorf("expected error but got success")
			}
			if !tt.wantErr && !response.Success {
				t.Errorf("expected success but got error: %v", response.Error)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	userRepo := newMockUserRepo()
	jwtManager := newTestJWTManager()
	refreshTokenRepo := newMockRefreshTokenRepo()

	// Create a test user with hashed password
	hashedPassword, err := password.Hash("Admin123!")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	testUser := &domain.User{
		ID:            uuid.New(),
		Email:         "test@example.com",
		Username:      "testuser",
		PasswordHash:  &hashedPassword,
		Role:          domain.RoleUser,
		EmailVerified: true,
	}
	userRepo.Create(context.Background(), testUser)

	authService := service.NewAuthService(
		userRepo,
		&mockOAuthRepo{},
		refreshTokenRepo,
		jwtManager,
		&mockEmailSender{},
		"http://localhost:5173",
	)

	r := createTestRouter()
	cfg := &config.Config{
		Server: config.ServerConfig{
			AllowOrigins: []string{"http://localhost:5173"},
		},
	}
	authHandler := handler.NewAuthHandler(authService, cfg)
	r.Post("/api/auth/login", authHandler.Login)

	tests := []struct {
		name       string
		body       domain.LoginRequest
		wantStatus int
		wantErr    bool
	}{
		{
			name: "successful login",
			body: domain.LoginRequest{
				Email:    "test@example.com",
				Password: "Admin123!",
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "wrong password",
			body: domain.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name: "non-existent user",
			body: domain.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "Password123",
			},
			wantStatus: http.StatusUnauthorized,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "POST", "/api/auth/login", tt.body, "")

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}

			response := parseResponse(t, rr)
			if tt.wantErr && response.Success {
				t.Errorf("expected error but got success")
			}
			if !tt.wantErr && !response.Success {
				t.Errorf("expected success but got error: %v", response.Error)
			}
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	jwtManager := newTestJWTManager()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Create a test token
	userID := uuid.New()
	token, _ := jwtManager.GenerateAccessToken(userID, "user")

	r := createTestRouter()

	// Protected endpoint
	r.With(authMiddleware.RequireAuth).Get("/protected", func(w http.ResponseWriter, r *http.Request) {
		ctxUserID := middleware.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"user_id": ctxUserID.String()})
	})

	tests := []struct {
		name       string
		token      string
		wantStatus int
	}{
		{
			name:       "valid token",
			token:      token,
			wantStatus: http.StatusOK,
		},
		{
			name:       "no token",
			token:      "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid token",
			token:      "invalid-token",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := makeRequest(t, r, "GET", "/protected", nil, tt.token)

			if rr.Code != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.wantStatus)
			}
		})
	}
}
