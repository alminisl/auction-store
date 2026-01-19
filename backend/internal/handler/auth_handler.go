package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/auction-cards/backend/internal/config"
	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/service"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthHandler struct {
	authService  *service.AuthService
	oauthConfig  *oauth2.Config
	frontendURL  string
}

func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	var oauthConfig *oauth2.Config
	if cfg.OAuth.GoogleClientID != "" {
		oauthConfig = &oauth2.Config{
			ClientID:     cfg.OAuth.GoogleClientID,
			ClientSecret: cfg.OAuth.GoogleClientSecret,
			RedirectURL:  cfg.OAuth.GoogleRedirectURL,
			Scopes:       []string{"email", "profile"},
			Endpoint:     google.Endpoint,
		}
	}

	return &AuthHandler{
		authService:  authService,
		oauthConfig:  oauthConfig,
		frontendURL:  cfg.Server.AllowOrigins[0],
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	user, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "Registration successful. Please check your email to verify your account.",
		"user":    user.ToPublic(),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	authResponse, refreshToken, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		handleError(w, err)
		return
	}

	// Set refresh token as httpOnly cookie
	h.setRefreshTokenCookie(w, refreshToken)

	respondJSON(w, http.StatusOK, authResponse)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err == nil && refreshToken.Value != "" {
		_ = h.authService.Logout(r.Context(), refreshToken.Value)
	}

	// Clear the cookie
	h.clearRefreshTokenCookie(w)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil || refreshToken.Value == "" {
		respondError(w, http.StatusUnauthorized, "NO_REFRESH_TOKEN", "No refresh token provided")
		return
	}

	accessToken, err := h.authService.RefreshAccessToken(r.Context(), refreshToken.Value)
	if err != nil {
		h.clearRefreshTokenCookie(w)
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req domain.VerifyEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.authService.VerifyEmail(r.Context(), req.Token); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req domain.ForgotPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	// Always return success to prevent email enumeration
	_ = h.authService.ForgotPassword(r.Context(), &req)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "If the email exists, a password reset link has been sent",
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req domain.ResetPasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if errors := validateRequest(&req); errors != nil {
		respondValidationError(w, errors)
		return
	}

	if err := h.authService.ResetPassword(r.Context(), &req); err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Password reset successfully",
	})
}

// Google OAuth handlers

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig == nil {
		respondError(w, http.StatusNotImplemented, "NOT_CONFIGURED", "Google OAuth not configured")
		return
	}

	state := generateOAuthState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if h.oauthConfig == nil {
		respondError(w, http.StatusNotImplemented, "NOT_CONFIGURED", "Google OAuth not configured")
		return
	}

	// Verify state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Redirect(w, r, h.frontendURL+"/login?error=invalid_state", http.StatusTemporaryRedirect)
		return
	}

	// Clear state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, h.frontendURL+"/login?error=no_code", http.StatusTemporaryRedirect)
		return
	}

	// Exchange code for token
	token, err := h.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=exchange_failed", http.StatusTemporaryRedirect)
		return
	}

	// Get user info from Google
	client := h.oauthConfig.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=userinfo_failed", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()

	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=decode_failed", http.StatusTemporaryRedirect)
		return
	}

	// Create or get user
	user, err := h.authService.GetOrCreateOAuthUser(r.Context(), "google", googleUser.ID, googleUser.Email, googleUser.Name)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=create_user_failed", http.StatusTemporaryRedirect)
		return
	}

	// Generate tokens
	authResponse, refreshToken, err := h.authService.GenerateTokens(r.Context(), user)
	if err != nil {
		http.Redirect(w, r, h.frontendURL+"/login?error=token_failed", http.StatusTemporaryRedirect)
		return
	}

	// Set refresh token cookie
	h.setRefreshTokenCookie(w, refreshToken)

	// Redirect to frontend with access token
	redirectURL := h.frontendURL + "/oauth/callback?access_token=" + url.QueryEscape(authResponse.AccessToken)
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// Helper methods

func (h *AuthHandler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *AuthHandler) clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func generateOAuthState() string {
	return time.Now().Format("20060102150405")
}
