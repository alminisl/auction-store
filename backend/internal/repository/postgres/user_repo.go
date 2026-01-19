package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, avatar_url, bio, phone, address, role, email_verified, email_verification_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.AvatarURL,
		user.Bio,
		user.Phone,
		user.Address,
		user.Role,
		user.EmailVerified,
		user.EmailVerificationToken,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, bio, phone, address, role,
		       email_verified, email_verification_token, password_reset_token, password_reset_expires,
		       is_banned, created_at, updated_at
		FROM users
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	user := &domain.User{}
	err := q.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Bio,
		&user.Phone,
		&user.Address,
		&user.Role,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, bio, phone, address, role,
		       email_verified, email_verification_token, password_reset_token, password_reset_expires,
		       is_banned, created_at, updated_at
		FROM users
		WHERE email = $1`

	q := r.db.GetQuerier(ctx)
	user := &domain.User{}
	err := q.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Bio,
		&user.Phone,
		&user.Address,
		&user.Role,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, bio, phone, address, role,
		       email_verified, email_verification_token, password_reset_token, password_reset_expires,
		       is_banned, created_at, updated_at
		FROM users
		WHERE username = $1`

	q := r.db.GetQuerier(ctx)
	user := &domain.User{}
	err := q.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Bio,
		&user.Phone,
		&user.Address,
		&user.Role,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByVerificationToken(ctx context.Context, token string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, bio, phone, address, role,
		       email_verified, email_verification_token, password_reset_token, password_reset_expires,
		       is_banned, created_at, updated_at
		FROM users
		WHERE email_verification_token = $1`

	q := r.db.GetQuerier(ctx)
	user := &domain.User{}
	err := q.QueryRow(ctx, query, token).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Bio,
		&user.Phone,
		&user.Address,
		&user.Role,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by verification token: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByPasswordResetToken(ctx context.Context, token string) (*domain.User, error) {
	query := `
		SELECT id, email, username, password_hash, avatar_url, bio, phone, address, role,
		       email_verified, email_verification_token, password_reset_token, password_reset_expires,
		       is_banned, created_at, updated_at
		FROM users
		WHERE password_reset_token = $1 AND password_reset_expires > NOW()`

	q := r.db.GetQuerier(ctx)
	user := &domain.User{}
	err := q.QueryRow(ctx, query, token).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.AvatarURL,
		&user.Bio,
		&user.Phone,
		&user.Address,
		&user.Role,
		&user.EmailVerified,
		&user.EmailVerificationToken,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.IsBanned,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by password reset token: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, username = $3, password_hash = $4, avatar_url = $5, bio = $6,
		    phone = $7, address = $8, role = $9, email_verified = $10, email_verification_token = $11,
		    password_reset_token = $12, password_reset_expires = $13, is_banned = $14
		WHERE id = $1
		RETURNING updated_at`

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.AvatarURL,
		user.Bio,
		user.Phone,
		user.Address,
		user.Role,
		user.EmailVerified,
		user.EmailVerificationToken,
		user.PasswordResetToken,
		user.PasswordResetExpires,
		user.IsBanned,
	).Scan(&user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *UserRepository) List(ctx context.Context, page, limit int) ([]domain.User, int, error) {
	countQuery := `SELECT COUNT(*) FROM users`
	listQuery := `
		SELECT id, email, username, password_hash, avatar_url, bio, phone, address, role,
		       email_verified, email_verification_token, password_reset_token, password_reset_expires,
		       is_banned, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	q := r.db.GetQuerier(ctx)

	var totalCount int
	if err := q.QueryRow(ctx, countQuery).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := q.Query(ctx, listQuery, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Username,
			&user.PasswordHash,
			&user.AvatarURL,
			&user.Bio,
			&user.Phone,
			&user.Address,
			&user.Role,
			&user.EmailVerified,
			&user.EmailVerificationToken,
			&user.PasswordResetToken,
			&user.PasswordResetExpires,
			&user.IsBanned,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, totalCount, nil
}

func (r *UserRepository) GetRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.UserRatingSummary, error) {
	query := `
		SELECT
			$1::uuid as user_id,
			COALESCE(AVG(rating)::float, 0) as average_rating,
			COUNT(*) as total_ratings,
			COALESCE(AVG(CASE WHEN type = 'seller' THEN rating END)::float, 0) as seller_rating,
			COUNT(CASE WHEN type = 'seller' THEN 1 END) as seller_count,
			COALESCE(AVG(CASE WHEN type = 'buyer' THEN rating END)::float, 0) as buyer_rating,
			COUNT(CASE WHEN type = 'buyer' THEN 1 END) as buyer_count
		FROM ratings
		WHERE rated_user_id = $1`

	q := r.db.GetQuerier(ctx)
	summary := &domain.UserRatingSummary{}
	err := q.QueryRow(ctx, query, userID).Scan(
		&summary.UserID,
		&summary.AverageRating,
		&summary.TotalRatings,
		&summary.SellerRating,
		&summary.SellerCount,
		&summary.BuyerRating,
		&summary.BuyerCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating summary: %w", err)
	}

	return summary, nil
}
