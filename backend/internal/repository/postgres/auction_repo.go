package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type AuctionRepository struct {
	db *DB
}

func NewAuctionRepository(db *DB) *AuctionRepository {
	return &AuctionRepository{db: db}
}

func (r *AuctionRepository) Create(ctx context.Context, auction *domain.Auction) error {
	query := `
		INSERT INTO auctions (id, seller_id, category_id, title, description, condition, starting_price,
		                      reserve_price, buy_now_price, current_price, bid_increment, start_time,
		                      end_time, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at, version`

	if auction.ID == uuid.Nil {
		auction.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		auction.ID,
		auction.SellerID,
		auction.CategoryID,
		auction.Title,
		auction.Description,
		auction.Condition,
		auction.StartingPrice,
		auction.ReservePrice,
		auction.BuyNowPrice,
		auction.CurrentPrice,
		auction.BidIncrement,
		auction.StartTime,
		auction.EndTime,
		auction.Status,
	).Scan(&auction.CreatedAt, &auction.UpdatedAt, &auction.Version)

	if err != nil {
		return fmt.Errorf("failed to create auction: %w", err)
	}

	return nil
}

func (r *AuctionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Auction, error) {
	query := `
		SELECT id, seller_id, category_id, title, description, condition, starting_price,
		       reserve_price, buy_now_price, current_price, bid_increment, start_time, end_time,
		       status, winner_id, winning_bid_id, views_count, bid_count, version, created_at, updated_at
		FROM auctions
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	auction := &domain.Auction{}
	err := q.QueryRow(ctx, query, id).Scan(
		&auction.ID,
		&auction.SellerID,
		&auction.CategoryID,
		&auction.Title,
		&auction.Description,
		&auction.Condition,
		&auction.StartingPrice,
		&auction.ReservePrice,
		&auction.BuyNowPrice,
		&auction.CurrentPrice,
		&auction.BidIncrement,
		&auction.StartTime,
		&auction.EndTime,
		&auction.Status,
		&auction.WinnerID,
		&auction.WinningBidID,
		&auction.ViewsCount,
		&auction.BidCount,
		&auction.Version,
		&auction.CreatedAt,
		&auction.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get auction by id: %w", err)
	}

	return auction, nil
}

func (r *AuctionRepository) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.Auction, error) {
	// Get auction
	auction, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	q := r.db.GetQuerier(ctx)

	// Get seller
	sellerQuery := `SELECT id, username, avatar_url, bio, created_at FROM users WHERE id = $1`
	seller := &domain.PublicUser{}
	err = q.QueryRow(ctx, sellerQuery, auction.SellerID).Scan(
		&seller.ID, &seller.Username, &seller.AvatarURL, &seller.Bio, &seller.CreatedAt,
	)
	if err == nil {
		auction.Seller = seller
	}

	// Get category
	if auction.CategoryID != nil {
		categoryQuery := `SELECT id, name, slug, parent_id, description, image_url, created_at FROM categories WHERE id = $1`
		category := &domain.Category{}
		err = q.QueryRow(ctx, categoryQuery, *auction.CategoryID).Scan(
			&category.ID, &category.Name, &category.Slug, &category.ParentID,
			&category.Description, &category.ImageURL, &category.CreatedAt,
		)
		if err == nil {
			auction.Category = category
		}
	}

	// Get images
	imagesQuery := `SELECT id, auction_id, url, position, created_at FROM auction_images WHERE auction_id = $1 ORDER BY position`
	rows, err := q.Query(ctx, imagesQuery, id)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var img domain.AuctionImage
			if err := rows.Scan(&img.ID, &img.AuctionID, &img.URL, &img.Position, &img.CreatedAt); err == nil {
				auction.Images = append(auction.Images, img)
			}
		}
	}

	// Get winner if exists
	if auction.WinnerID != nil {
		winner := &domain.PublicUser{}
		err = q.QueryRow(ctx, sellerQuery, *auction.WinnerID).Scan(
			&winner.ID, &winner.Username, &winner.AvatarURL, &winner.Bio, &winner.CreatedAt,
		)
		if err == nil {
			auction.Winner = winner
		}
	}

	return auction, nil
}

func (r *AuctionRepository) Update(ctx context.Context, auction *domain.Auction) error {
	query := `
		UPDATE auctions
		SET category_id = $2, title = $3, description = $4, condition = $5, starting_price = $6,
		    reserve_price = $7, buy_now_price = $8, current_price = $9, bid_increment = $10,
		    start_time = $11, end_time = $12, status = $13, winner_id = $14, winning_bid_id = $15,
		    bid_count = $16, version = version + 1
		WHERE id = $1
		RETURNING updated_at, version`

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		auction.ID,
		auction.CategoryID,
		auction.Title,
		auction.Description,
		auction.Condition,
		auction.StartingPrice,
		auction.ReservePrice,
		auction.BuyNowPrice,
		auction.CurrentPrice,
		auction.BidIncrement,
		auction.StartTime,
		auction.EndTime,
		auction.Status,
		auction.WinnerID,
		auction.WinningBidID,
		auction.BidCount,
	).Scan(&auction.UpdatedAt, &auction.Version)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to update auction: %w", err)
	}

	return nil
}

func (r *AuctionRepository) UpdateWithVersion(ctx context.Context, auction *domain.Auction, expectedVersion int) error {
	query := `
		UPDATE auctions
		SET current_price = $2, bid_count = $3, end_time = $4, version = version + 1
		WHERE id = $1 AND version = $5
		RETURNING updated_at, version`

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		auction.ID,
		auction.CurrentPrice,
		auction.BidCount,
		auction.EndTime,
		expectedVersion,
	).Scan(&auction.UpdatedAt, &auction.Version)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrConcurrentBid
	}
	if err != nil {
		return fmt.Errorf("failed to update auction with version: %w", err)
	}

	return nil
}

func (r *AuctionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM auctions WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete auction: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *AuctionRepository) List(ctx context.Context, params *domain.AuctionListParams) ([]domain.Auction, int, error) {
	baseQuery := `FROM auctions a`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if params.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, *params.Status)
		argIndex++
	}

	if params.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.category_id = $%d", argIndex))
		args = append(args, *params.CategoryID)
		argIndex++
	}

	if params.SellerID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.seller_id = $%d", argIndex))
		args = append(args, *params.SellerID)
		argIndex++
	}

	if params.Search != nil && *params.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("to_tsvector('english', a.title || ' ' || COALESCE(a.description, '')) @@ plainto_tsquery('english', $%d)", argIndex))
		args = append(args, *params.Search)
		argIndex++
	}

	if params.MinPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.current_price >= $%d", argIndex))
		args = append(args, *params.MinPrice)
		argIndex++
	}

	if params.MaxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("a.current_price <= $%d", argIndex))
		args = append(args, *params.MaxPrice)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = " WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + baseQuery + whereClause
	q := r.db.GetQuerier(ctx)

	var totalCount int
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count auctions: %w", err)
	}

	// Sort
	orderBy := " ORDER BY "
	switch params.SortBy {
	case "ending_soon":
		orderBy += "a.end_time ASC"
	case "newest":
		orderBy += "a.created_at DESC"
	case "price_low":
		orderBy += "a.current_price ASC"
	case "price_high":
		orderBy += "a.current_price DESC"
	case "most_bids":
		orderBy += "a.bid_count DESC"
	default:
		orderBy += "a.created_at DESC"
	}

	// Pagination
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
		SELECT a.id, a.seller_id, a.category_id, a.title, a.description, a.condition, a.starting_price,
		       a.reserve_price, a.buy_now_price, a.current_price, a.bid_increment, a.start_time, a.end_time,
		       a.status, a.winner_id, a.winning_bid_id, a.views_count, a.bid_count, a.version, a.created_at, a.updated_at
		%s%s%s LIMIT $%d OFFSET $%d`, baseQuery, whereClause, orderBy, argIndex, argIndex+1)

	rows, err := q.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list auctions: %w", err)
	}
	defer rows.Close()

	auctions := make([]domain.Auction, 0)
	for rows.Next() {
		var auction domain.Auction
		err := rows.Scan(
			&auction.ID,
			&auction.SellerID,
			&auction.CategoryID,
			&auction.Title,
			&auction.Description,
			&auction.Condition,
			&auction.StartingPrice,
			&auction.ReservePrice,
			&auction.BuyNowPrice,
			&auction.CurrentPrice,
			&auction.BidIncrement,
			&auction.StartTime,
			&auction.EndTime,
			&auction.Status,
			&auction.WinnerID,
			&auction.WinningBidID,
			&auction.ViewsCount,
			&auction.BidCount,
			&auction.Version,
			&auction.CreatedAt,
			&auction.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan auction: %w", err)
		}
		auctions = append(auctions, auction)
	}

	return auctions, totalCount, nil
}

func (r *AuctionRepository) GetEndingAuctions(ctx context.Context, beforeUnix int64) ([]domain.Auction, error) {
	query := `
		SELECT id, seller_id, category_id, title, description, condition, starting_price,
		       reserve_price, buy_now_price, current_price, bid_increment, start_time, end_time,
		       status, winner_id, winning_bid_id, views_count, bid_count, version, created_at, updated_at
		FROM auctions
		WHERE status = 'active' AND end_time <= to_timestamp($1)`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query, beforeUnix)
	if err != nil {
		return nil, fmt.Errorf("failed to get ending auctions: %w", err)
	}
	defer rows.Close()

	auctions := make([]domain.Auction, 0)
	for rows.Next() {
		var auction domain.Auction
		err := rows.Scan(
			&auction.ID,
			&auction.SellerID,
			&auction.CategoryID,
			&auction.Title,
			&auction.Description,
			&auction.Condition,
			&auction.StartingPrice,
			&auction.ReservePrice,
			&auction.BuyNowPrice,
			&auction.CurrentPrice,
			&auction.BidIncrement,
			&auction.StartTime,
			&auction.EndTime,
			&auction.Status,
			&auction.WinnerID,
			&auction.WinningBidID,
			&auction.ViewsCount,
			&auction.BidCount,
			&auction.Version,
			&auction.CreatedAt,
			&auction.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan auction: %w", err)
		}
		auctions = append(auctions, auction)
	}

	return auctions, nil
}

func (r *AuctionRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE auctions SET views_count = views_count + 1 WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	_, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment view count: %w", err)
	}

	return nil
}

func (r *AuctionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.AuctionStatus, winnerID *uuid.UUID, winningBidID *uuid.UUID) error {
	query := `
		UPDATE auctions
		SET status = $2, winner_id = $3, winning_bid_id = $4
		WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, id, status, winnerID, winningBidID)
	if err != nil {
		return fmt.Errorf("failed to update auction status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// AuctionImageRepository
type AuctionImageRepository struct {
	db *DB
}

func NewAuctionImageRepository(db *DB) *AuctionImageRepository {
	return &AuctionImageRepository{db: db}
}

func (r *AuctionImageRepository) Create(ctx context.Context, image *domain.AuctionImage) error {
	query := `
		INSERT INTO auction_images (id, auction_id, url, position)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`

	if image.ID == uuid.Nil {
		image.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query, image.ID, image.AuctionID, image.URL, image.Position).Scan(&image.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create auction image: %w", err)
	}

	return nil
}

func (r *AuctionImageRepository) GetByAuctionID(ctx context.Context, auctionID uuid.UUID) ([]domain.AuctionImage, error) {
	query := `SELECT id, auction_id, url, position, created_at FROM auction_images WHERE auction_id = $1 ORDER BY position`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query, auctionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction images: %w", err)
	}
	defer rows.Close()

	images := make([]domain.AuctionImage, 0)
	for rows.Next() {
		var img domain.AuctionImage
		if err := rows.Scan(&img.ID, &img.AuctionID, &img.URL, &img.Position, &img.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

func (r *AuctionImageRepository) GetFirstImageByAuctionIDs(ctx context.Context, auctionIDs []uuid.UUID) (map[uuid.UUID]domain.AuctionImage, error) {
	if len(auctionIDs) == 0 {
		return make(map[uuid.UUID]domain.AuctionImage), nil
	}

	// Build query with DISTINCT ON to get first image per auction
	query := `
		SELECT DISTINCT ON (auction_id) id, auction_id, url, position, created_at
		FROM auction_images
		WHERE auction_id = ANY($1)
		ORDER BY auction_id, position ASC`

	q := r.db.GetQuerier(ctx)
	rows, err := q.Query(ctx, query, auctionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get auction images: %w", err)
	}
	defer rows.Close()

	images := make(map[uuid.UUID]domain.AuctionImage)
	for rows.Next() {
		var img domain.AuctionImage
		if err := rows.Scan(&img.ID, &img.AuctionID, &img.URL, &img.Position, &img.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images[img.AuctionID] = img
	}

	return images, nil
}

func (r *AuctionImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM auction_images WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	result, err := q.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete auction image: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *AuctionImageRepository) DeleteByAuctionID(ctx context.Context, auctionID uuid.UUID) error {
	query := `DELETE FROM auction_images WHERE auction_id = $1`

	q := r.db.GetQuerier(ctx)
	_, err := q.Exec(ctx, query, auctionID)
	if err != nil {
		return fmt.Errorf("failed to delete auction images: %w", err)
	}

	return nil
}

func (r *AuctionImageRepository) UpdatePositions(ctx context.Context, auctionID uuid.UUID, positions map[uuid.UUID]int) error {
	for imageID, position := range positions {
		query := `UPDATE auction_images SET position = $1 WHERE id = $2 AND auction_id = $3`
		q := r.db.GetQuerier(ctx)
		_, err := q.Exec(ctx, query, position, imageID, auctionID)
		if err != nil {
			return fmt.Errorf("failed to update image position: %w", err)
		}
	}
	return nil
}

// BidRepository
type BidRepository struct {
	db *DB
}

func NewBidRepository(db *DB) *BidRepository {
	return &BidRepository{db: db}
}

func (r *BidRepository) Create(ctx context.Context, bid *domain.Bid) error {
	query := `
		INSERT INTO bids (id, auction_id, bidder_id, amount, is_auto_bid, max_auto_bid)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at`

	if bid.ID == uuid.Nil {
		bid.ID = uuid.New()
	}

	q := r.db.GetQuerier(ctx)
	err := q.QueryRow(ctx, query,
		bid.ID, bid.AuctionID, bid.BidderID, bid.Amount, bid.IsAutoBid, bid.MaxAutoBid,
	).Scan(&bid.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create bid: %w", err)
	}

	return nil
}

func (r *BidRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bid, error) {
	query := `SELECT id, auction_id, bidder_id, amount, is_auto_bid, max_auto_bid, created_at FROM bids WHERE id = $1`

	q := r.db.GetQuerier(ctx)
	bid := &domain.Bid{}
	err := q.QueryRow(ctx, query, id).Scan(
		&bid.ID, &bid.AuctionID, &bid.BidderID, &bid.Amount, &bid.IsAutoBid, &bid.MaxAutoBid, &bid.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bid: %w", err)
	}

	return bid, nil
}

func (r *BidRepository) GetHighestBid(ctx context.Context, auctionID uuid.UUID) (*domain.Bid, error) {
	query := `
		SELECT id, auction_id, bidder_id, amount, is_auto_bid, max_auto_bid, created_at
		FROM bids
		WHERE auction_id = $1
		ORDER BY amount DESC, created_at ASC
		LIMIT 1`

	q := r.db.GetQuerier(ctx)
	bid := &domain.Bid{}
	err := q.QueryRow(ctx, query, auctionID).Scan(
		&bid.ID, &bid.AuctionID, &bid.BidderID, &bid.Amount, &bid.IsAutoBid, &bid.MaxAutoBid, &bid.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get highest bid: %w", err)
	}

	return bid, nil
}

func (r *BidRepository) GetByAuctionID(ctx context.Context, auctionID uuid.UUID, page, limit int) ([]domain.Bid, int, error) {
	countQuery := `SELECT COUNT(*) FROM bids WHERE auction_id = $1`
	listQuery := `
		SELECT b.id, b.auction_id, b.bidder_id, b.amount, b.is_auto_bid, b.max_auto_bid, b.created_at,
		       u.id, u.username, u.avatar_url, u.bio, u.created_at
		FROM bids b
		JOIN users u ON b.bidder_id = u.id
		WHERE b.auction_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3`

	q := r.db.GetQuerier(ctx)

	var totalCount int
	if err := q.QueryRow(ctx, countQuery, auctionID).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count bids: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := q.Query(ctx, listQuery, auctionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bids: %w", err)
	}
	defer rows.Close()

	bids := make([]domain.Bid, 0)
	for rows.Next() {
		var bid domain.Bid
		bidder := &domain.PublicUser{}
		err := rows.Scan(
			&bid.ID, &bid.AuctionID, &bid.BidderID, &bid.Amount, &bid.IsAutoBid, &bid.MaxAutoBid, &bid.CreatedAt,
			&bidder.ID, &bidder.Username, &bidder.AvatarURL, &bidder.Bio, &bidder.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan bid: %w", err)
		}
		bid.Bidder = bidder
		bids = append(bids, bid)
	}

	return bids, totalCount, nil
}

func (r *BidRepository) GetByBidderID(ctx context.Context, bidderID uuid.UUID, page, limit int) ([]domain.Bid, int, error) {
	countQuery := `SELECT COUNT(*) FROM bids WHERE bidder_id = $1`
	listQuery := `
		SELECT id, auction_id, bidder_id, amount, is_auto_bid, max_auto_bid, created_at
		FROM bids
		WHERE bidder_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	q := r.db.GetQuerier(ctx)

	var totalCount int
	if err := q.QueryRow(ctx, countQuery, bidderID).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count bids: %w", err)
	}

	offset := (page - 1) * limit
	rows, err := q.Query(ctx, listQuery, bidderID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list bids: %w", err)
	}
	defer rows.Close()

	bids := make([]domain.Bid, 0)
	for rows.Next() {
		var bid domain.Bid
		err := rows.Scan(
			&bid.ID, &bid.AuctionID, &bid.BidderID, &bid.Amount, &bid.IsAutoBid, &bid.MaxAutoBid, &bid.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan bid: %w", err)
		}
		bids = append(bids, bid)
	}

	return bids, totalCount, nil
}

func (r *BidRepository) GetBidCount(ctx context.Context, auctionID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM bids WHERE auction_id = $1`

	q := r.db.GetQuerier(ctx)
	var count int
	if err := q.QueryRow(ctx, query, auctionID).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to get bid count: %w", err)
	}

	return count, nil
}

func (r *BidRepository) GetPreviousHighBidder(ctx context.Context, auctionID uuid.UUID, excludeBidderID uuid.UUID) (*domain.Bid, error) {
	query := `
		SELECT id, auction_id, bidder_id, amount, is_auto_bid, max_auto_bid, created_at
		FROM bids
		WHERE auction_id = $1 AND bidder_id != $2
		ORDER BY amount DESC, created_at ASC
		LIMIT 1`

	q := r.db.GetQuerier(ctx)
	bid := &domain.Bid{}
	err := q.QueryRow(ctx, query, auctionID, excludeBidderID).Scan(
		&bid.ID, &bid.AuctionID, &bid.BidderID, &bid.Amount, &bid.IsAutoBid, &bid.MaxAutoBid, &bid.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get previous high bidder: %w", err)
	}

	return bid, nil
}

// BidTransaction implements atomic bid placement
type BidTransaction struct {
	db          *DB
	auctionRepo *AuctionRepository
	bidRepo     *BidRepository
}

func NewBidTransaction(db *DB, auctionRepo *AuctionRepository, bidRepo *BidRepository) *BidTransaction {
	return &BidTransaction{
		db:          db,
		auctionRepo: auctionRepo,
		bidRepo:     bidRepo,
	}
}

func (t *BidTransaction) PlaceBid(ctx context.Context, auctionID, bidderID uuid.UUID, amount decimal.Decimal, maxAutoBid *decimal.Decimal) (*PlaceBidResult, error) {
	var result *PlaceBidResult

	err := t.db.WithTx(ctx, func(txCtx context.Context) error {
		// Get auction with lock
		auction, err := t.auctionRepo.GetByID(txCtx, auctionID)
		if err != nil {
			return err
		}

		// Validate auction is active
		if auction.Status != domain.AuctionStatusActive {
			return domain.ErrAuctionNotActive
		}

		// Check auction hasn't ended
		if auction.EndTime.Unix() < getCurrentUnixTime() {
			return domain.ErrAuctionEnded
		}

		// Validate not self-bidding
		if auction.SellerID == bidderID {
			return domain.ErrSelfBidding
		}

		// Validate bid amount
		minBid := auction.CurrentPrice.Add(auction.BidIncrement)
		if amount.LessThan(minBid) {
			return domain.ErrBidTooLow
		}

		// Get previous high bidder for outbid notification
		prevBid, _ := t.bidRepo.GetHighestBid(txCtx, auctionID)
		var prevBidderID *uuid.UUID
		if prevBid != nil && prevBid.BidderID != bidderID {
			prevBidderID = &prevBid.BidderID
		}

		// Create bid
		bid := &domain.Bid{
			AuctionID:  auctionID,
			BidderID:   bidderID,
			Amount:     amount,
			IsAutoBid:  maxAutoBid != nil,
			MaxAutoBid: maxAutoBid,
		}
		if err := t.bidRepo.Create(txCtx, bid); err != nil {
			return err
		}

		// Check for anti-sniping (bid in last 5 minutes)
		auctionExtended := false
		var newEndTime *int64
		fiveMinutesFromNow := getCurrentUnixTime() + 300
		if auction.EndTime.Unix() < fiveMinutesFromNow {
			// Extend by 2 minutes
			extendedTime := auction.EndTime.Add(2 * 60 * 1000000000) // 2 minutes in nanoseconds
			auction.EndTime = extendedTime
			auctionExtended = true
			endTimeUnix := extendedTime.Unix()
			newEndTime = &endTimeUnix
		}

		// Update auction
		auction.CurrentPrice = amount
		auction.BidCount++
		expectedVersion := auction.Version

		if err := t.auctionRepo.UpdateWithVersion(txCtx, auction, expectedVersion); err != nil {
			return err
		}

		result = &PlaceBidResult{
			Bid:             bid,
			Auction:         auction,
			AuctionExtended: auctionExtended,
			NewEndTime:      newEndTime,
			PreviousBidder:  prevBidderID,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func getCurrentUnixTime() int64 {
	return 0 // Will be replaced with actual time in service
}

// PlaceBidResult for the repository package
type PlaceBidResult struct {
	Bid             *domain.Bid
	Auction         *domain.Auction
	AuctionExtended bool
	NewEndTime      *int64
	PreviousBidder  *uuid.UUID
}
