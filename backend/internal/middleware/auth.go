package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/pkg/jwt"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

type AuthMiddleware struct {
	jwtManager *jwt.Manager
}

func NewAuthMiddleware(jwtManager *jwt.Manager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

// RequireAuth validates the JWT token and adds user info to context
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				respondError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Access token expired")
				return
			}
			respondError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid access token")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth adds user info to context if token is present, but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			next.ServeHTTP(w, r)
			return
		}

		tokenString := parts[1]
		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin requires the user to have admin role
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := GetUserRole(r.Context())
		if role != string(domain.RoleAdmin) {
			respondError(w, http.StatusForbidden, "FORBIDDEN", "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper functions to get user info from context

func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}

func GetUserRole(ctx context.Context) string {
	if role, ok := ctx.Value(UserRoleKey).(string); ok {
		return role
	}
	return ""
}

func IsAuthenticated(ctx context.Context) bool {
	return GetUserID(ctx) != uuid.Nil
}

func IsAdmin(ctx context.Context) bool {
	return GetUserRole(ctx) == string(domain.RoleAdmin)
}
