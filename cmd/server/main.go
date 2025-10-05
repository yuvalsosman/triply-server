package main

import (
	"log"
	"time"
	"triply-server/internal/config"
	"triply-server/internal/handlers"
	"triply-server/internal/middleware"
	"triply-server/internal/models"
	"triply-server/internal/repository"
	"triply-server/internal/service"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := openDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := autoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Seed demo data
	if err := seedDemoData(db); err != nil {
		log.Printf("Failed to seed demo data: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tripRepo := repository.NewTripRepository(db)
	publicTripRepo := repository.NewPublicTripRepository(db)
	activityRepo := repository.NewActivityRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo)
	tripService := service.NewTripService(tripRepo)
	publicTripService := service.NewPublicTripService(publicTripRepo, tripRepo)
	activityService := service.NewActivityService(activityRepo)
	importService := service.NewImportService(publicTripRepo, tripRepo)

	// Initialize OAuth config
	oauthConfig := setupOAuth(cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, tripService, oauthConfig, cfg.JWT.Secret, cfg.Server.FrontendOrigin)
	tripHandler := handlers.NewTripHandler(tripService)
	publicTripHandler := handlers.NewPublicTripHandler(publicTripService)
	activityHandler := handlers.NewActivityHandler(activityService)
	importHandler := handlers.NewImportHandler(importService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWT.Secret)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		ErrorHandler:          middleware.ErrorHandler,
	})

	// Global middleware
	app.Use(middleware.Logger)
	app.Use(func(c *fiber.Ctx) error {
		c.Response().Header.Add("Cache-Control", "no-store")
		c.Response().Header.Add("Pragma", "no-cache")
		c.Response().Header.Add("Expires", "0")
		return c.Next()
	})
	app.Use(middleware.NewCORS(cfg.Server.FrontendOrigin))

	// Health check
	app.Get("/api/health", func(c *fiber.Ctx) error {
		// Check DB connection
		sqlDB, err := db.DB()
		if err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status":   "unhealthy",
				"database": "error",
			})
		}
		if err := sqlDB.Ping(); err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status":   "unhealthy",
				"database": "down",
			})
		}
		return c.JSON(fiber.Map{
			"status":   "healthy",
			"database": "up",
		})
	})

	// Auth routes (public)
	app.Get("/auth/google", authHandler.GoogleLogin)
	app.Get("/auth/google/callback", authHandler.GoogleCallback)
	app.Post("/auth/dev-login", authHandler.DevLogin)
	app.Post("/auth/logout", authHandler.Logout)

	// Auth routes (protected)
	app.Get("/auth/me", authMiddleware.OptionalAuth, authHandler.GetMe)
	app.Put("/api/user/profile", authMiddleware.OptionalAuth, authHandler.UpdateProfile)
	app.Post("/auth/migrate-shadow-trips", authMiddleware.OptionalAuth, authHandler.MigrateShadowTrips)

	// Trip routes (protected)
	apiRoutes := app.Group("/api")
	apiRoutes.Get("/users/:userId/trips", authMiddleware.OptionalAuth, tripHandler.ListTrips)
	apiRoutes.Post("/users/:userId/trips", authMiddleware.OptionalAuth, tripHandler.CreateTrip)
	apiRoutes.Put("/users/:userId/trips/:tripId", authMiddleware.OptionalAuth, tripHandler.UpdateTrip)
	apiRoutes.Delete("/users/:userId/trips/:tripId", authMiddleware.OptionalAuth, tripHandler.DeleteTrip)

	// Activity routes
	apiRoutes.Post("/activities/order", activityHandler.UpdateActivityOrder)

	// Public trips routes
	apiRoutes.Get("/public-trips", publicTripHandler.ListPublicTrips)
	apiRoutes.Get("/public-trips/:tripId", publicTripHandler.GetPublicTripDetail)
	apiRoutes.Post("/public-trips/:tripId/visibility", authMiddleware.OptionalAuth, publicTripHandler.ToggleVisibility)

	// Import routes (protected)
	apiRoutes.Post("/import-trip", authMiddleware.OptionalAuth, importHandler.ImportTripParts)

	// Start server
	log.Printf("🚀 Triply server listening on :%s", cfg.Server.Port)
	if err := app.Listen(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}
}

func openDB(cfg *config.Config) (*gorm.DB, error) {
	log.Printf("Connecting to PostgreSQL: %s", cfg.Database.URL)
	return gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{})
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Trip{},
		&models.Destination{},
		&models.DayPlan{},
		&models.Activity{},
		&models.PublicTrip{},
	)
}

func setupOAuth(cfg *config.Config) *oauth2.Config {
	if cfg.Auth.GoogleClientID == "" || cfg.Auth.GoogleClientSecret == "" {
		return nil
	}

	return &oauth2.Config{
		ClientID:     cfg.Auth.GoogleClientID,
		ClientSecret: cfg.Auth.GoogleClientSecret,
		RedirectURL:  cfg.Auth.OAuthRedirectURL,
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
}

func seedDemoData(db *gorm.DB) error {
	// Create demo user
	now := time.Now().UTC()
	demoUser := models.User{
		ID:        "user-sarah",
		Name:      "Sarah Levi",
		Email:     "sarah.levi@example.com",
		Locale:    "en",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&demoUser).Error; err != nil {
		return err
	}

	// Check if user already has trips
	var count int64
	if err := db.Model(&models.Trip{}).Where("user_id = ?", demoUser.ID).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // Already seeded
	}

	// Create demo trip
	desc := "Sample spring itinerary between Tokyo and Kyoto with a balance of culture and cuisine."
	tz := "Asia/Tokyo"
	cover := "https://images.unsplash.com/photo-1549692520-acc6669e2f0c?auto=format&fit=crop&w=1400&q=80"

	trip := models.Trip{
		ID:            "trip-001",
		UserID:        demoUser.ID,
		Name:          "My Japan Trip",
		Description:   &desc,
		TravelerCount: 2,
		StartDate:     "2025-03-28",
		EndDate:       "2025-04-10",
		Locale:        "en",
		Visibility:    "private",
		Status:        "active",
		Timezone:      &tz,
		CoverImage:    &cover,
		CreatedAt:     now,
		UpdatedAt:     now,
		Destinations: []models.Destination{
			{
				ID:        "dest-tokyo",
				City:      "Tokyo",
				Region:    ptr("Kanto"),
				HeroImage: ptr("https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=1400&q=80"),
				StartDate: "2025-03-28",
				EndDate:   "2025-04-02",
				DailyPlans: []models.DayPlan{
					{
						ID:    "day-tokyo-1",
						Date:  "2025-03-28",
						Notes: ptr("Arrival day with light activities to adjust to the time zone."),
						Activities: []models.Activity{
							{
								ID:        "act-1",
								Title:     "Narita Express to Tokyo Station",
								TimeOfDay: "start",
								Order:     0,
								Location:  "Narita Airport",
								Type:      "transportation",
								PlaceID:   ptr("ChIJN5X73rWMImARPA6C8I-g2NA"),
							},
							{
								ID:        "act-2",
								Title:     "Check-in at Nihonbashi boutique hotel",
								TimeOfDay: "start",
								Order:     1,
								Location:  "Chiyoda",
								Type:      "accommodation",
								Coordinates: &models.Coordinates{
									Lat: 35.6995,
									Lng: 139.7537,
								},
							},
							{
								ID:        "act-3",
								Title:     "Lunch at Tsukiji Outer Market",
								TimeOfDay: "mid",
								Order:     0,
								Location:  "Tsukiji",
								Type:      "meal",
								Coordinates: &models.Coordinates{
									Lat: 35.6655,
									Lng: 139.7708,
								},
							},
							{
								ID:        "act-4",
								Title:     "Evening stroll in Asakusa and Senso-ji",
								TimeOfDay: "end",
								Order:     0,
								Location:  "Asakusa",
								Type:      "culture",
								Coordinates: &models.Coordinates{
									Lat: 35.7148,
									Lng: 139.7967,
								},
							},
						},
					},
				},
			},
		},
	}

	return db.Session(&gorm.Session{FullSaveAssociations: true}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&trip).Error
}

func ptr[T any](v T) *T {
	return &v
}
