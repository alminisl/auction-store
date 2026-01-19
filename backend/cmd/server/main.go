package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/auction-cards/backend/internal/cache"
	"github.com/auction-cards/backend/internal/config"
	"github.com/auction-cards/backend/internal/handler"
	"github.com/auction-cards/backend/internal/middleware"
	"github.com/auction-cards/backend/internal/pkg/email"
	"github.com/auction-cards/backend/internal/pkg/jwt"
	"github.com/auction-cards/backend/internal/pkg/storage"
	"github.com/auction-cards/backend/internal/repository/postgres"
	"github.com/auction-cards/backend/internal/service"
	"github.com/auction-cards/backend/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}
}

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := postgres.NewDB(cfg.Database.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to PostgreSQL")

	// Connect to Redis
	redisCache, err := cache.NewRedisCache(cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		redisCache = nil
	} else {
		defer redisCache.Close()
		log.Println("Connected to Redis")
	}

	// Initialize S3 storage
	s3Storage, err := storage.NewS3Storage(&storage.Config{
		Endpoint:        cfg.S3.Endpoint,
		AccessKeyID:     cfg.S3.AccessKeyID,
		SecretAccessKey: cfg.S3.SecretAccessKey,
		BucketName:      cfg.S3.BucketName,
		UseSSL:          cfg.S3.UseSSL,
		PublicURL:       cfg.S3.PublicURL,
	})
	if err != nil {
		log.Printf("Warning: Failed to connect to S3: %v", err)
		s3Storage = nil
	} else {
		log.Println("Connected to S3 storage")
	}

	// Initialize email sender (mock for development)
	emailSender := email.NewMockSender()

	// Initialize JWT manager
	jwtManager := jwt.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiration,
		cfg.JWT.RefreshExpiration,
	)

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	oauthRepo := postgres.NewOAuthAccountRepository(db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db)
	auctionRepo := postgres.NewAuctionRepository(db)
	auctionImageRepo := postgres.NewAuctionImageRepository(db)
	bidRepo := postgres.NewBidRepository(db)
	categoryRepo := postgres.NewCategoryRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	watchlistRepo := postgres.NewWatchlistRepository(db)
	ratingRepo := postgres.NewRatingRepository(db)
	reportRepo := postgres.NewReportRepository(db)
	messageRepo := postgres.NewMessageRepository(db)

	// Initialize services
	frontendURL := cfg.Server.AllowOrigins[0]

	authService := service.NewAuthService(
		userRepo,
		oauthRepo,
		refreshTokenRepo,
		jwtManager,
		emailSender,
		frontendURL,
	)

	notificationService := service.NewNotificationService(
		notificationRepo,
		userRepo,
		watchlistRepo,
		emailSender,
		frontendURL,
	)

	auctionService := service.NewAuctionService(
		auctionRepo,
		auctionImageRepo,
		categoryRepo,
		s3Storage,
	)

	bidService := service.NewBidService(
		bidRepo,
		auctionRepo,
		nil, // bid transaction not needed with simpler implementation
		notificationService,
		redisCache,
	)

	userService := service.NewUserService(
		userRepo,
		watchlistRepo,
		ratingRepo,
		auctionRepo,
	)

	schedulerService := service.NewSchedulerService(
		auctionRepo,
		bidRepo,
		notificationService,
		redisCache,
	)

	// Initialize WebSocket hubs
	wsHub := websocket.NewHub(redisCache)
	go wsHub.Run()

	messageHub := websocket.NewMessageHub(redisCache)
	go messageHub.Run()

	// Initialize message service
	messageService, err := service.NewMessageService(
		messageRepo,
		userRepo,
		cfg.Messaging.EncryptionKey,
		messageHub,
	)
	if err != nil {
		log.Fatalf("Failed to initialize message service: %v", err)
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, cfg)
	auctionHandler := handler.NewAuctionHandler(auctionService)
	bidHandler := handler.NewBidHandler(bidService)
	userHandler := handler.NewUserHandler(userService, notificationService)
	adminHandler := handler.NewAdminHandler(
		userService,
		auctionService,
		categoryRepo,
		reportRepo,
		auctionRepo,
		bidRepo,
	)
	wsHandler := handler.NewWebSocketHandler(wsHub)
	messageHandler := handler.NewMessageHandler(messageService)
	messageWsHandler := handler.NewMessageWebSocketHandler(messageHub)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CORS(&middleware.CORSConfig{
		AllowedOrigins:   cfg.Server.AllowOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		AllowCredentials: true,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Apply global rate limiting
		r.Use(middleware.RateLimit(redisCache, middleware.DefaultRateLimitConfig()))

		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Use(middleware.RateLimit(redisCache, middleware.AuthRateLimitConfig()))
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Post("/refresh", authHandler.RefreshToken)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)
			r.Get("/google", authHandler.GoogleLogin)
			r.Get("/google/callback", authHandler.GoogleCallback)
		})

		// Categories (public)
		r.Get("/categories", auctionHandler.GetCategories)
		r.Get("/categories/{slug}", auctionHandler.GetCategoryBySlug)

		// Auctions (public read, auth write)
		r.Route("/auctions", func(r chi.Router) {
			r.With(authMiddleware.OptionalAuth).Get("/", auctionHandler.List)
			r.With(authMiddleware.OptionalAuth).Get("/{id}", auctionHandler.GetByID)
			r.Get("/{id}/bids", bidHandler.GetBidsByAuction)

			// Authenticated routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Post("/", auctionHandler.Create)
				r.Put("/{id}", auctionHandler.Update)
				r.Delete("/{id}", auctionHandler.Delete)
				r.Post("/{id}/publish", auctionHandler.Publish)
				r.Post("/{id}/images", auctionHandler.UploadImage)
				r.Delete("/{id}/images/{imageId}", auctionHandler.DeleteImage)

				// Bidding with rate limiting
				r.With(middleware.RateLimit(redisCache, middleware.BidRateLimitConfig())).
					Post("/{id}/bids", bidHandler.PlaceBid)
				r.Post("/{id}/buy-now", bidHandler.BuyNow)
			})
		})

		// Users
		r.Route("/users", func(r chi.Router) {
			// Authenticated routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequireAuth)
				r.Get("/me", authHandler.GetMe)
				r.Put("/me", userHandler.UpdateProfile)
				r.Get("/me/bids", bidHandler.GetMyBids)
			})

			// Public user profiles
			r.Get("/{id}", userHandler.GetPublicProfile)
			r.Get("/{id}/auctions", userHandler.GetUserAuctions)
			r.Get("/{id}/ratings", userHandler.GetUserRatings)
		})

		// Watchlist (authenticated)
		r.Route("/watchlist", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/", userHandler.GetWatchlist)
			r.Post("/{auctionId}", userHandler.AddToWatchlist)
			r.Delete("/{auctionId}", userHandler.RemoveFromWatchlist)
		})

		// Notifications (authenticated)
		r.Route("/notifications", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/", userHandler.GetNotifications)
			r.Put("/{id}/read", userHandler.MarkNotificationRead)
			r.Put("/read-all", userHandler.MarkAllNotificationsRead)
		})

		// Ratings (authenticated)
		r.Route("/ratings", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Post("/auction/{auctionId}", userHandler.CreateRating)
		})

		// Admin routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Use(authMiddleware.RequireAdmin)

			r.Get("/dashboard", adminHandler.GetDashboard)
			r.Get("/users", adminHandler.ListUsers)
			r.Put("/users/{id}/ban", adminHandler.BanUser)
			r.Get("/auctions", adminHandler.ListAuctions)
			r.Put("/auctions/{id}/status", adminHandler.UpdateAuctionStatus)
			r.Post("/categories", adminHandler.CreateCategory)
			r.Put("/categories/{id}", adminHandler.UpdateCategory)
			r.Delete("/categories/{id}", adminHandler.DeleteCategory)
			r.Get("/reports", adminHandler.ListReports)
			r.Put("/reports/{id}", adminHandler.UpdateReport)
		})

		// Messages (authenticated)
		r.Route("/messages", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Post("/", messageHandler.SendMessage)
			r.Get("/unread-count", messageHandler.GetUnreadCount)
		})

		// Conversations (authenticated)
		r.Route("/conversations", func(r chi.Router) {
			r.Use(authMiddleware.RequireAuth)
			r.Get("/", messageHandler.GetConversations)
			r.Get("/{id}", messageHandler.GetConversation)
			r.Get("/{id}/messages", messageHandler.GetMessages)
			r.Put("/{id}/read", messageHandler.MarkAsRead)
		})
	})

	// WebSocket routes
	r.With(authMiddleware.OptionalAuth).Get("/ws/auctions/{id}", wsHandler.HandleAuctionWS)
	r.With(authMiddleware.RequireAuth).Get("/ws/messages", messageWsHandler.HandleMessageWS)

	// Start scheduler
	schedulerService.Start()
	defer schedulerService.Stop()

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		wsHub.Stop()
		messageHub.Stop()
		server.Shutdown(ctx)
	}()

	// Start server
	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}

	log.Println("Server stopped")
}
