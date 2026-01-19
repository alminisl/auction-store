package service

import (
	"context"

	"github.com/auction-cards/backend/internal/domain"
	"github.com/auction-cards/backend/internal/repository"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo      repository.UserRepository
	watchlistRepo repository.WatchlistRepository
	ratingRepo    repository.RatingRepository
	auctionRepo   repository.AuctionRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	watchlistRepo repository.WatchlistRepository,
	ratingRepo repository.RatingRepository,
	auctionRepo repository.AuctionRepository,
) *UserService {
	return &UserService{
		userRepo:      userRepo,
		watchlistRepo: watchlistRepo,
		ratingRepo:    ratingRepo,
		auctionRepo:   auctionRepo,
	}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *UserService) GetPublicProfile(ctx context.Context, userID uuid.UUID) (*domain.PublicUser, *domain.UserRatingSummary, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	ratingSummary, err := s.ratingRepo.GetUserRatingSummary(ctx, userID)
	if err != nil {
		ratingSummary = &domain.UserRatingSummary{UserID: userID}
	}

	return user.ToPublic(), ratingSummary, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Username != nil {
		// Check if username is taken by another user
		existing, err := s.userRepo.GetByUsername(ctx, *req.Username)
		if err == nil && existing.ID != userID {
			return nil, domain.ErrUsernameExists
		}
		user.Username = *req.Username
	}

	if req.Bio != nil {
		user.Bio = req.Bio
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if req.Address != nil {
		user.Address = req.Address
	}

	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Watchlist methods

func (s *UserService) GetWatchlist(ctx context.Context, userID uuid.UUID, page, limit int) (*domain.WatchlistResponse, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	items, totalCount, err := s.watchlistRepo.GetByUser(ctx, userID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + limit - 1) / limit

	return &domain.WatchlistResponse{
		Items:      items,
		TotalCount: totalCount,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *UserService) AddToWatchlist(ctx context.Context, userID, auctionID uuid.UUID) error {
	// Verify auction exists
	_, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return err
	}

	item := &domain.WatchlistItem{
		UserID:    userID,
		AuctionID: auctionID,
	}

	return s.watchlistRepo.Add(ctx, item)
}

func (s *UserService) RemoveFromWatchlist(ctx context.Context, userID, auctionID uuid.UUID) error {
	return s.watchlistRepo.Remove(ctx, userID, auctionID)
}

func (s *UserService) IsInWatchlist(ctx context.Context, userID, auctionID uuid.UUID) (bool, error) {
	return s.watchlistRepo.Exists(ctx, userID, auctionID)
}

// Rating methods

func (s *UserService) GetUserRatings(ctx context.Context, userID uuid.UUID, params *domain.RatingListParams) (*domain.RatingListResponse, error) {
	params.RatedUserID = &userID

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	ratings, totalCount, err := s.ratingRepo.GetByRatedUser(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + params.Limit - 1) / params.Limit

	return &domain.RatingListResponse{
		Ratings:    ratings,
		TotalCount: totalCount,
		Page:       params.Page,
		TotalPages: totalPages,
	}, nil
}

func (s *UserService) CreateRating(ctx context.Context, auctionID, raterID uuid.UUID, req *domain.CreateRatingRequest) (*domain.Rating, error) {
	// Get auction
	auction, err := s.auctionRepo.GetByID(ctx, auctionID)
	if err != nil {
		return nil, err
	}

	// Verify auction is completed
	if auction.Status != domain.AuctionStatusCompleted {
		return nil, domain.ErrBadRequest
	}

	// Determine rating type and rated user
	var ratingType domain.RatingType
	var ratedUserID uuid.UUID

	if raterID == auction.SellerID {
		// Seller rating buyer
		if auction.WinnerID == nil {
			return nil, domain.ErrBadRequest
		}
		ratingType = domain.RatingTypeBuyer
		ratedUserID = *auction.WinnerID
	} else if auction.WinnerID != nil && raterID == *auction.WinnerID {
		// Buyer rating seller
		ratingType = domain.RatingTypeSeller
		ratedUserID = auction.SellerID
	} else {
		return nil, domain.ErrForbidden
	}

	// Check if already rated
	_, err = s.ratingRepo.GetByAuctionAndRater(ctx, auctionID, raterID, ratingType)
	if err == nil {
		return nil, domain.ErrConflict
	}

	rating := &domain.Rating{
		AuctionID:   auctionID,
		RaterID:     raterID,
		RatedUserID: ratedUserID,
		Rating:      req.Rating,
		Comment:     req.Comment,
		Type:        ratingType,
	}

	if err := s.ratingRepo.Create(ctx, rating); err != nil {
		return nil, err
	}

	return rating, nil
}

// Admin methods

func (s *UserService) ListUsers(ctx context.Context, page, limit int) ([]domain.User, int, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	return s.userRepo.List(ctx, page, limit)
}

func (s *UserService) BanUser(ctx context.Context, userID uuid.UUID, ban bool) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.IsBanned = ban
	return s.userRepo.Update(ctx, user)
}

func (s *UserService) GetUserAuctions(ctx context.Context, userID uuid.UUID, page, limit int) (*domain.AuctionListResponse, error) {
	params := &domain.AuctionListParams{
		SellerID: &userID,
		Page:     page,
		Limit:    limit,
	}

	auctions, totalCount, err := s.auctionRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	totalPages := (totalCount + limit - 1) / limit

	return &domain.AuctionListResponse{
		Auctions:   auctions,
		TotalCount: totalCount,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}
