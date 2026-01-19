package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NotificationRepository struct {
	db *DB
}

func NewNotificationRepository(db *DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, type, title, message, auction_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	if notification.ID == uuid.Nil {
		notification.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		notification.ID,
		notification.UserID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.AuctionID,
	).Scan(&notification.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *NotificationRepository) CreateBatch(ctx context.Context, notifications []domain.Notification) error {
	if len(notifications) == 0 {
		return nil
	}

	query := `
		INSERT INTO notifications (id, user_id, type, title, message, auction_id)
		VALUES ($1, $2, $3, $4, $5, $6)`

	q := r.db.GetQuerier(ctx)
	for _, n := range notifications {
		if n.ID == uuid.Nil {
			n.ID = uuid.New()
		}
		_, err := q.Exec(ctx, query, n.ID, n.UserID, n.Type, n.Title, n.Message, n.AuctionID)
		if err != nil {
			return fmt.Errorf("failed to create notification: %w", err)
		}
	}

	return nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, title, message, auction_id, is_read, created_at
		FROM notifications
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	n := &domain.Notification{}
	err := q.QueryRow(ctx, query, id).Scan(
		&n.ID,
		&n.UserID,
		&n.Type,
		&n.Title,
		&n.Message,
		&n.AuctionID,
		&n.IsRead,
		&n.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	return n, nil
}

func (r *NotificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, params *domain.NotificationListParams) ([]domain.Notification, int, int, error) {
	whereClause := "WHERE user_id = $1"
	args := []interface{}{userID}
	argIndex := 2

	if params.Unread != nil && *params.Unread {
		whereClause += " AND is_read = FALSE"
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications %s", whereClause)
	unreadQuery := "SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE"

	q := r.db.GetQuerier(ctx)

	var totalCount, unreadCount int
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count notifications: %w", err)
	}
	if err := q.QueryRow(ctx, unreadQuery, userID).Scan(&unreadCount); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	args = append(args, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, user_id, type, title, message, auction_id, is_read, created_at
		FROM notifications
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := q.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to list notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]domain.Notification, 0)
	for rows.Next() {
		var n domain.Notification
		err := rows.Scan(
			&n.ID,
			&n.UserID,
			&n.Type,
			&n.Title,
			&n.Message,
			&n.AuctionID,
			&n.IsRead,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, n)
	}

	return notifications, totalCount, unreadCount, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *NotificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`

	q := r.db.GetQuerier(ctx)
	_, err := q.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	return nil
}

func (r *NotificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = FALSE`

	q := r.db.GetQuerier(ctx)
	var count int
	if err := q.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// WatchlistRepository
type WatchlistRepository struct {
	db *DB
}

func NewWatchlistRepository(db *DB) *WatchlistRepository {
	return &WatchlistRepository{db: db}
}

func (r *WatchlistRepository) Add(ctx context.Context, item *domain.WatchlistItem) error {
	query := `
		INSERT INTO watchlist (id, user_id, auction_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, auction_id) DO NOTHING
		RETURNING created_at`

	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query, item.ID, item.UserID, item.AuctionID).Scan(&item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Already exists, not an error
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to add to watchlist: %w", err)
	}

	return nil
}

func (r *WatchlistRepository) Remove(ctx context.Context, userID, auctionID uuid.UUID) error {
	query := `DELETE FROM watchlist WHERE user_id = $1 AND auction_id = $2`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, userID, auctionID)
	if err != nil {
		return fmt.Errorf("failed to remove from watchlist: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *WatchlistRepository) GetByUser(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.WatchlistItem, int, error) {
	countQuery := `SELECT COUNT(*) FROM watchlist WHERE user_id = $1`

	q := r.db.GetQuerier(ctx)
	var totalCount int
	if err := q.QueryRow(ctx, countQuery, userID).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count watchlist: %w", err)
	}

	offset := (page - 1) * limit
	listQuery := `
		SELECT w.id, w.user_id, w.auction_id, w.created_at,
		       a.id, a.seller_id, a.category_id, a.title, a.description, a.condition,
		       a.starting_price, a.reserve_price, a.buy_now_price, a.current_price,
		       a.bid_increment, a.start_time, a.end_time, a.status, a.winner_id,
		       a.winning_bid_id, a.views_count, a.bid_count, a.version, a.created_at, a.updated_at
		FROM watchlist w
		JOIN auctions a ON w.auction_id = a.id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := q.Query(ctx, listQuery, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list watchlist: %w", err)
	}
	defer rows.Close()

	items := make([]domain.WatchlistItem, 0)
	for rows.Next() {
		var item domain.WatchlistItem
		auction := &domain.Auction{}
		err := rows.Scan(
			&item.ID, &item.UserID, &item.AuctionID, &item.CreatedAt,
			&auction.ID, &auction.SellerID, &auction.CategoryID, &auction.Title,
			&auction.Description, &auction.Condition, &auction.StartingPrice,
			&auction.ReservePrice, &auction.BuyNowPrice, &auction.CurrentPrice,
			&auction.BidIncrement, &auction.StartTime, &auction.EndTime, &auction.Status,
			&auction.WinnerID, &auction.WinningBidID, &auction.ViewsCount, &auction.BidCount,
			&auction.Version, &auction.CreatedAt, &auction.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan watchlist item: %w", err)
		}
		item.Auction = auction
		items = append(items, item)
	}

	return items, totalCount, nil
}

func (r *WatchlistRepository) Exists(ctx context.Context, userID, auctionID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM watchlist WHERE user_id = $1 AND auction_id = $2)`

	q := r.db.GetQuerier(ctx)
	var exists bool
	if err := q.QueryRow(ctx, query, userID, auctionID).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check watchlist: %w", err)
	}

	return exists, nil
}

func (r *WatchlistRepository) GetWatchersForAuction(ctx context.Context, auctionID uuid.UUID) ([]uuid.UUID, error) {
	query := `SELECT user_id FROM watchlist WHERE auction_id = $1`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query, auctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchers: %w", err)
	}
	defer rows.Close()

	userIDs := make([]uuid.UUID, 0)
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user id: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

// RatingRepository
type RatingRepository struct {
	db *DB
}

func NewRatingRepository(db *DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) Create(ctx context.Context, rating *domain.Rating) error {
	query := `
		INSERT INTO ratings (id, auction_id, rater_id, rated_user_id, rating, comment, type)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`

	if rating.ID == uuid.Nil {
		rating.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		rating.ID,
		rating.AuctionID,
		rating.RaterID,
		rating.RatedUserID,
		rating.Rating,
		rating.Comment,
		rating.Type,
	).Scan(&rating.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create rating: %w", err)
	}

	return nil
}

func (r *RatingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Rating, error) {
	query := `
		SELECT id, auction_id, rater_id, rated_user_id, rating, comment, type, created_at
		FROM ratings
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	rating := &domain.Rating{}
	err := q.QueryRow(ctx, query, id).Scan(
		&rating.ID,
		&rating.AuctionID,
		&rating.RaterID,
		&rating.RatedUserID,
		&rating.Rating,
		&rating.Comment,
		&rating.Type,
		&rating.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	return rating, nil
}

func (r *RatingRepository) GetByAuctionAndRater(ctx context.Context, auctionID, raterID uuid.UUID, ratingType domain.RatingType) (*domain.Rating, error) {
	query := `
		SELECT id, auction_id, rater_id, rated_user_id, rating, comment, type, created_at
		FROM ratings
		WHERE auction_id = $1 AND rater_id = $2 AND type = $3`

	q := r.db.GetQuerier(ctx)
	rating := &domain.Rating{}
	err := q.QueryRow(ctx, query, auctionID, raterID, ratingType).Scan(
		&rating.ID,
		&rating.AuctionID,
		&rating.RaterID,
		&rating.RatedUserID,
		&rating.Rating,
		&rating.Comment,
		&rating.Type,
		&rating.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get rating: %w", err)
	}

	return rating, nil
}

func (r *RatingRepository) GetByRatedUser(ctx context.Context, ratedUserID uuid.UUID, params *domain.RatingListParams) ([]domain.Rating, int, error) {
	whereClause := "WHERE r.rated_user_id = $1"
	args := []interface{}{ratedUserID}
	argIndex := 2

	if params.Type != nil {
		whereClause += fmt.Sprintf(" AND r.type = $%d", argIndex)
		args = append(args, *params.Type)
		argIndex++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ratings r %s", whereClause)

	q := r.db.GetQuerier(ctx)
	var totalCount int
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count ratings: %w", err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	args = append(args, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT r.id, r.auction_id, r.rater_id, r.rated_user_id, r.rating, r.comment, r.type, r.created_at,
		       u.id, u.username, u.avatar_url, u.bio, u.created_at
		FROM ratings r
		JOIN users u ON r.rater_id = u.id
		%s
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := q.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list ratings: %w", err)
	}
	defer rows.Close()

	ratings := make([]domain.Rating, 0)
	for rows.Next() {
		var rating domain.Rating
		rater := &domain.PublicUser{}
		err := rows.Scan(
			&rating.ID, &rating.AuctionID, &rating.RaterID, &rating.RatedUserID,
			&rating.Rating, &rating.Comment, &rating.Type, &rating.CreatedAt,
			&rater.ID, &rater.Username, &rater.AvatarURL, &rater.Bio, &rater.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan rating: %w", err)
		}
		rating.Rater = rater
		ratings = append(ratings, rating)
	}

	return ratings, totalCount, nil
}

func (r *RatingRepository) GetUserRatingSummary(ctx context.Context, userID uuid.UUID) (*domain.UserRatingSummary, error) {
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

// ReportRepository
type ReportRepository struct {
	db *DB
}

func NewReportRepository(db *DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) Create(ctx context.Context, report *domain.ReportedListing) error {
	query := `
		INSERT INTO reported_listings (id, auction_id, reporter_id, reason, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, status`

	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		report.ID,
		report.AuctionID,
		report.ReporterID,
		report.Reason,
		report.Description,
	).Scan(&report.CreatedAt, &report.Status)

	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	return nil
}

func (r *ReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ReportedListing, error) {
	query := `
		SELECT id, auction_id, reporter_id, reason, description, status, created_at
		FROM reported_listings
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	report := &domain.ReportedListing{}
	err := q.QueryRow(ctx, query, id).Scan(
		&report.ID,
		&report.AuctionID,
		&report.ReporterID,
		&report.Reason,
		&report.Description,
		&report.Status,
		&report.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	return report, nil
}

func (r *ReportRepository) Update(ctx context.Context, report *domain.ReportedListing) error {
	query := `UPDATE reported_listings SET status = $2 WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, report.ID, report.Status)
	if err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *ReportRepository) List(ctx context.Context, params *domain.ReportListParams) ([]domain.ReportedListing, int, error) {
	whereClause := ""
	args := []interface{}{}
	argIndex := 1

	if params.Status != nil {
		whereClause = fmt.Sprintf("WHERE status = $%d", argIndex)
		args = append(args, *params.Status)
		argIndex++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM reported_listings %s", whereClause)

	q := r.db.GetQuerier(ctx)
	var totalCount int
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 20
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	args = append(args, limit, offset)
	listQuery := fmt.Sprintf(`
		SELECT r.id, r.auction_id, r.reporter_id, r.reason, r.description, r.status, r.created_at
		FROM reported_listings r
		%s
		ORDER BY r.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	rows, err := q.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list reports: %w", err)
	}
	defer rows.Close()

	reports := make([]domain.ReportedListing, 0)
	for rows.Next() {
		var report domain.ReportedListing
		err := rows.Scan(
			&report.ID,
			&report.AuctionID,
			&report.ReporterID,
			&report.Reason,
			&report.Description,
			&report.Status,
			&report.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan report: %w", err)
		}
		reports = append(reports, report)
	}

	return reports, totalCount, nil
}

// OAuthAccountRepository
type OAuthAccountRepository struct {
	db *DB
}

func NewOAuthAccountRepository(db *DB) *OAuthAccountRepository {
	return &OAuthAccountRepository{db: db}
}

func (r *OAuthAccountRepository) Create(ctx context.Context, account *domain.OAuthAccount) error {
	query := `
		INSERT INTO oauth_accounts (id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`

	if account.ID == uuid.Nil {
		account.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		account.ID,
		account.UserID,
		account.Provider,
		account.ProviderUserID,
		account.AccessToken,
		account.RefreshToken,
		account.ExpiresAt,
	).Scan(&account.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create oauth account: %w", err)
	}

	return nil
}

func (r *OAuthAccountRepository) GetByProviderUserID(ctx context.Context, provider, providerUserID string) (*domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at
		FROM oauth_accounts
		WHERE provider = $1 AND provider_user_id = $2`

	q := r.db.GetQuerier(ctx)
	account := &domain.OAuthAccount{}
	err := q.QueryRow(ctx, query, provider, providerUserID).Scan(
		&account.ID,
		&account.UserID,
		&account.Provider,
		&account.ProviderUserID,
		&account.AccessToken,
		&account.RefreshToken,
		&account.ExpiresAt,
		&account.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth account: %w", err)
	}

	return account, nil
}

func (r *OAuthAccountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.OAuthAccount, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, access_token, refresh_token, expires_at, created_at
		FROM oauth_accounts
		WHERE user_id = $1`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.OAuthAccount, 0)
	for rows.Next() {
		var account domain.OAuthAccount
		err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.Provider,
			&account.ProviderUserID,
			&account.AccessToken,
			&account.RefreshToken,
			&account.ExpiresAt,
			&account.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan oauth account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *OAuthAccountRepository) Update(ctx context.Context, account *domain.OAuthAccount) error {
	query := `
		UPDATE oauth_accounts
		SET access_token = $2, refresh_token = $3, expires_at = $4
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, account.ID, account.AccessToken, account.RefreshToken, account.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to update oauth account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *OAuthAccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM oauth_accounts WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oauth account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// RefreshTokenRepository
type RefreshTokenRepository struct {
	db *DB
}

func NewRefreshTokenRepository(db *DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		token.ID,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(&token.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND expires_at > NOW()`

	q := r.db.GetQuerier(ctx)
	token := &domain.RefreshToken{}
	err := q.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return token, nil
}

func (r *RefreshTokenRepository) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`

	q := r.db.GetQuerier(ctx)
	_, err := q.Exec(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	q := r.db.GetQuerier(ctx)
	_, err := q.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}

	return nil
}

func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`

	q := r.db.GetQuerier(ctx)
	_, err := q.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}
