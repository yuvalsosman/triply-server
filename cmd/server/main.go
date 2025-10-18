package main

import (
	"fmt"
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

	// Seed public trips
	if err := seedPublicTrips(db); err != nil {
		log.Printf("Failed to seed public trips: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	tripRepo := repository.NewTripRepository(db)
	publicTripRepo := repository.NewPublicTripRepository(db)
	activityRepo := repository.NewActivityRepository(db)
	tripLikeRepo := repository.NewTripLikeRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo)
	tripService := service.NewTripService(tripRepo, publicTripRepo)
	publicTripService := service.NewPublicTripService(publicTripRepo, tripRepo, tripLikeRepo)
	activityService := service.NewActivityService(activityRepo)
	importService := service.NewImportService(publicTripRepo, tripRepo)
	tripLikeService := service.NewTripLikeService(tripLikeRepo)

	// Initialize OAuth config
	oauthConfig := setupOAuth(cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, tripService, oauthConfig, cfg.JWT.Secret, cfg.Server.FrontendOrigin)
	tripHandler := handlers.NewTripHandler(tripService)
	publicTripHandler := handlers.NewPublicTripHandler(publicTripService)
	activityHandler := handlers.NewActivityHandler(activityService)
	importHandler := handlers.NewImportHandler(importService)
	tripLikeHandler := handlers.NewTripLikeHandler(tripLikeService)
	mapsHandler := handlers.NewMapsHandler(cfg.Maps.APIKey)

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
	apiRoutes.Get("/public-trips", authMiddleware.OptionalAuth, publicTripHandler.ListPublicTrips)
	apiRoutes.Get("/public-trips/:tripId", authMiddleware.OptionalAuth, publicTripHandler.GetPublicTripDetail)
	apiRoutes.Post("/public-trips/:tripId/visibility", authMiddleware.OptionalAuth, publicTripHandler.ToggleVisibility)
	apiRoutes.Post("/public-trips/:tripId/like", authMiddleware.RequireAuth, tripLikeHandler.ToggleLike)

	// Clone trip route (requires authentication)
	apiRoutes.Post("/trips/clone/:tripId", authMiddleware.RequireAuth, tripHandler.CloneTrip)

	// Import routes (protected)
	apiRoutes.Post("/import-trip", authMiddleware.OptionalAuth, importHandler.ImportTripParts)

	// Maps API routes (public - protected by HTTP referrer restrictions in Google Cloud Console)
	apiRoutes.Get("/maps/config", mapsHandler.GetMapConfig)
	apiRoutes.Get("/photo", mapsHandler.ProxyPhoto) // Proxy for Google Maps/Places photos

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
	// Drop all tables if they exist (both old and new names)
	log.Println("🗑️  Dropping old tables...")
	tablesToDrop := []string{
		"activity_imports", "day_plan_activities", "day_plan_destinations",
		"day_plans", "trip_destinations", "activities", "destinations",
		"trips", "users", "public_trips", "daily_plans", "trip_likes",
	}
	for _, table := range tablesToDrop {
		if db.Migrator().HasTable(table) {
			if err := db.Migrator().DropTable(table); err != nil {
				log.Printf("Warning: failed to drop table %s: %v", table, err)
			}
		}
	}

	log.Println("📦 Migrating new schema...")
	// Migrate in order to respect foreign key dependencies
	err := db.AutoMigrate(
		&models.User{},
		&models.Trip{},
		&models.Destination{},
		&models.Activity{},
		&models.DayPlan{},
		&models.TripDestination{},
		&models.DayPlanDestination{},
		&models.DayPlanActivity{},
		&models.ActivityImport{},
		&models.TripLike{},
	)
	if err != nil {
		return err
	}

	// Add unique constraint for user-trip likes
	err = db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_user_trip_like 
		ON trip_likes (user_id, trip_id)
	`).Error
	if err != nil {
		log.Printf("Warning: Failed to create unique index: %v", err)
	}

	return nil
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

	// Create destinations
	tokyoDest := models.Destination{
		ID:              "dest-tokyo",
		City:            "Tokyo",
		Region:          ptr("Kanto"),
		Country:         "Japan",
		Latitude:        ptr(35.6762),
		Longitude:       ptr(139.6503),
		Timezone:        ptr("Asia/Tokyo"),
		HeroImage:       ptr("https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=1400&q=80"),
		TripCount:       0,
		PopularityScore: 0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tokyoDest).Error; err != nil {
		return err
	}

	// Create activities
	activities := []models.Activity{
		{
			ID:              "act-narita-express",
			Title:           "Narita Express to Tokyo Station",
			Type:            "transportation",
			Location:        ptr("Narita Airport"),
			PlaceID:         ptr("ChIJN5X73rWMImARPA6C8I-g2NA"),
			DurationMinutes: ptr(90),
			UsageCount:      0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:         "act-hotel-checkin",
			Title:      "Check-in at Nihonbashi boutique hotel",
			Type:       "accommodation",
			Location:   ptr("Chiyoda"),
			Latitude:   ptr(35.6995),
			Longitude:  ptr(139.7537),
			UsageCount: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "act-tsukiji-lunch",
			Title:      "Lunch at Tsukiji Outer Market",
			Type:       "meal",
			Location:   ptr("Tsukiji"),
			Latitude:   ptr(35.6655),
			Longitude:  ptr(139.7708),
			UsageCount: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "act-asakusa-stroll",
			Title:      "Evening stroll in Asakusa and Senso-ji",
			Type:       "culture",
			Location:   ptr("Asakusa"),
			Latitude:   ptr(35.7148),
			Longitude:  ptr(139.7967),
			UsageCount: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	for _, act := range activities {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&act).Error; err != nil {
			return err
		}
	}

	// Create demo trip
	cover := "https://images.unsplash.com/photo-1549692520-acc6669e2f0c?auto=format&fit=crop&w=1400&q=80"

	trip := models.Trip{
		ID:            "trip-001",
		UserID:        demoUser.ID,
		Name:          "הטיול שלי ליפן",
		TravelerCount: 2,
		StartDate:     "2025-03-28",
		EndDate:       "2025-04-10",
		CoverImage:    cover,
		Visibility:    "private",
		Status:        "active",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := db.Create(&trip).Error; err != nil {
		return err
	}

	// Link trip to destination
	tripDest := models.TripDestination{
		ID:            "td-1",
		TripID:        trip.ID,
		DestinationID: tokyoDest.ID,
		OrderIndex:    0,
		StartDate:     ptr("2025-03-28"),
		EndDate:       ptr("2025-04-02"),
		CreatedAt:     now,
	}
	if err := db.Create(&tripDest).Error; err != nil {
		return err
	}

	// Create day plan
	dayPlan := models.DayPlan{
		ID:        "day-tokyo-1",
		TripID:    trip.ID,
		Date:      "2025-03-28",
		DayNumber: 1,
		Notes:     ptr("Arrival day with light activities to adjust to the time zone."),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&dayPlan).Error; err != nil {
		return err
	}

	// Link destination to day plan
	dayPlanDest := models.DayPlanDestination{
		ID:            "dpd-1",
		DayPlanID:     dayPlan.ID,
		DestinationID: tokyoDest.ID,
		OrderIndex:    0,
		PartOfDay:     ptr("all-day"),
		CreatedAt:     now,
	}
	if err := db.Create(&dayPlanDest).Error; err != nil {
		return err
	}

	// Link activities to day plan
	dayPlanActivities := []models.DayPlanActivity{
		{
			ID:              "dpa-1",
			DayPlanID:       dayPlan.ID,
			ActivityID:      "act-narita-express",
			TimeOfDay:       "start",
			OrderWithinTime: 0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "dpa-2",
			DayPlanID:       dayPlan.ID,
			ActivityID:      "act-hotel-checkin",
			TimeOfDay:       "start",
			OrderWithinTime: 1,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "dpa-3",
			DayPlanID:       dayPlan.ID,
			ActivityID:      "act-tsukiji-lunch",
			TimeOfDay:       "mid",
			OrderWithinTime: 0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "dpa-4",
			DayPlanID:       dayPlan.ID,
			ActivityID:      "act-asakusa-stroll",
			TimeOfDay:       "end",
			OrderWithinTime: 0,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	for _, dpa := range dayPlanActivities {
		if err := db.Create(&dpa).Error; err != nil {
			return err
		}
	}

	log.Println("✅ Seeded demo data successfully")
	return nil
}

func seedPublicTrips(db *gorm.DB) error {
	// Check if we already have public trips
	var count int64
	if err := db.Model(&models.Trip{}).Where("visibility = ?", "public").Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil // Already seeded
	}

	// Get or create a demo author user for public trips
	now := time.Now().UTC()
	authorUser := models.User{
		ID:        "user-public-author",
		Name:      "אוצר טיולים",
		Email:     "curator@triply.com",
		Locale:    "en",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&authorUser).Error; err != nil {
		return err
	}

	// Get destinations (or create if they don't exist)
	destinations := map[string]models.Destination{
		"Tokyo": {
			ID:        "dest-tokyo-public",
			City:      "Tokyo",
			Region:    ptr("Kanto"),
			Country:   "Japan",
			Latitude:  ptr(35.6762),
			Longitude: ptr(139.6503),
			Timezone:  ptr("Asia/Tokyo"),
			HeroImage: ptr("https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=1400&q=80"),
			CreatedAt: now,
			UpdatedAt: now,
		},
		"Kyoto": {
			ID:        "dest-kyoto-public",
			City:      "Kyoto",
			Region:    ptr("Kansai"),
			Country:   "Japan",
			Latitude:  ptr(35.0116),
			Longitude: ptr(135.7681),
			Timezone:  ptr("Asia/Tokyo"),
			HeroImage: ptr("https://images.unsplash.com/photo-1493976040374-85c8e12f0c0e?auto=format&fit=crop&w=1400&q=80"),
			CreatedAt: now,
			UpdatedAt: now,
		},
		"Osaka": {
			ID:        "dest-osaka-public",
			City:      "Osaka",
			Region:    ptr("Kansai"),
			Country:   "Japan",
			Latitude:  ptr(34.6937),
			Longitude: ptr(135.5023),
			Timezone:  ptr("Asia/Tokyo"),
			HeroImage: ptr("https://images.unsplash.com/photo-1528360983277-13d401cdc186?auto=format&fit=crop&w=1400&q=80"),
			CreatedAt: now,
			UpdatedAt: now,
		},
		"Nagoya": {
			ID:        "dest-nagoya-public",
			City:      "Nagoya",
			Region:    ptr("Chubu"),
			Country:   "Japan",
			Latitude:  ptr(35.1815),
			Longitude: ptr(136.9066),
			Timezone:  ptr("Asia/Tokyo"),
			HeroImage: ptr("https://images.unsplash.com/photo-1624993590528-4ee743c9896e?auto=format&fit=crop&w=1400&q=80"),
			CreatedAt: now,
			UpdatedAt: now,
		},
		"Hiroshima": {
			ID:        "dest-hiroshima-public",
			City:      "Hiroshima",
			Region:    ptr("Chugoku"),
			Country:   "Japan",
			Latitude:  ptr(34.3853),
			Longitude: ptr(132.4553),
			Timezone:  ptr("Asia/Tokyo"),
			HeroImage: ptr("https://images.unsplash.com/photo-1590559899731-a382839e5549?auto=format&fit=crop&w=1400&q=80"),
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, dest := range destinations {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&dest).Error; err != nil {
			return err
		}
	}

	// Create public trips
	publicTrips := []struct {
		trip         models.Trip
		destinations []string
	}{
		{
			trip: models.Trip{
				ID:            "pt-tokyo-week-discovery",
				UserID:        authorUser.ID,
				Name:          "גילוי טוקיו בשבוע",
				Slug:          ptr("tokyo-week-discovery"),
				TravelerCount: 2,
				StartDate:     "2025-05-15",
				EndDate:       "2025-05-21",
				CoverImage:    "https://images.unsplash.com/photo-1540959733332-eab4deabeeaf?auto=format&fit=crop&w=1200&q=80",
				Visibility:    "public",
				Status:        "completed",
				Summary:       ptr("חוו את השילוב המושלם של טוקיו בין מסורת עתיקה למודרניות חדשנית בשבוע מרגש."),
				TravelerType:  string(models.TravelerTypeCouple),
				Likes:         312,
				CreatedAt:     now.AddDate(0, -1, -25),
				UpdatedAt:     now.AddDate(0, 0, -20),
			},
			destinations: []string{"Tokyo"},
		},
		{
			trip: models.Trip{
				ID:            "pt-tokyo-kyoto-10days",
				UserID:        authorUser.ID,
				Name:          "מסע תרבותי בטוקיו וקיוטו",
				Slug:          ptr("tokyo-kyoto-10-day-cultural-journey"),
				TravelerCount: 2,
				StartDate:     "2025-10-10",
				EndDate:       "2025-10-19",
				CoverImage:    "https://images.unsplash.com/photo-1493976040374-85c8e12f0c0e?auto=format&fit=crop&w=1200&q=80",
				Visibility:    "public",
				Status:        "completed",
				Summary:       ptr("מסע מושלם של 10 ימים המשלב את האנרגיה המודרנית של טוקיו עם היופי והמסורת הנצחית של קיוטו."),
				TravelerType:  string(models.TravelerTypeFriends),
				Likes:         428,
				CreatedAt:     now.AddDate(0, -1, -13),
				UpdatedAt:     now.AddDate(0, 0, -17),
			},
			destinations: []string{"Tokyo", "Kyoto"},
		},
		{
			trip: models.Trip{
				ID:            "pt-kansai-grand-tour-14days",
				UserID:        authorUser.ID,
				Name:          "סיור קנסאי הגדול האולטימטיבי",
				Slug:          ptr("ultimate-kansai-tokyo-kyoto-osaka-14-days"),
				TravelerCount: 2,
				StartDate:     "2025-03-20",
				EndDate:       "2025-04-02",
				CoverImage:    "https://images.unsplash.com/photo-1528360983277-13d401cdc186?auto=format&fit=crop&w=1200&q=80",
				Visibility:    "public",
				Status:        "completed",
				Summary:       ptr("הרפתקת יפן האולטימטיבית של שבועיים המשלבת את האנרגיה של טוקיו, התרבות של קיוטו והקולינריה של אוסקה."),
				TravelerType:  string(models.TravelerTypeFamily),
				Likes:         567,
				CreatedAt:     now.AddDate(0, 0, -34),
				UpdatedAt:     now.AddDate(0, 0, -15),
			},
			destinations: []string{"Tokyo", "Kyoto", "Osaka"},
		},
		{
			trip: models.Trip{
				ID:            "pt-japan-grand-adventure-28days",
				UserID:        authorUser.ID,
				Name:          "הרפתקה גדולה ביפן - חודש מלא",
				Slug:          ptr("japan-grand-adventure-28-days-tokyo-kyoto-osaka-nagoya-hiroshima"),
				TravelerCount: 1,
				StartDate:     "2025-04-05",
				EndDate:       "2025-05-02",
				CoverImage:    "https://images.unsplash.com/photo-1590559899731-a382839e5549?auto=format&fit=crop&w=1200&q=80",
				Visibility:    "public",
				Status:        "completed",
				Summary:       ptr("מסע אפי של 28 ימים לחקור את יפן מטוקיו להירושימה, לחוות את המגוון המלא של התרבות, המטבח והנופים היפניים."),
				TravelerType:  string(models.TravelerTypeSolo),
				Likes:         892,
				CreatedAt:     now.AddDate(0, 0, -45),
				UpdatedAt:     now.AddDate(0, 0, -3),
			},
			destinations: []string{"Tokyo", "Kyoto", "Osaka", "Nagoya", "Hiroshima"},
		},
	}

	for _, pt := range publicTrips {
		if err := db.Create(&pt.trip).Error; err != nil {
			return err
		}

		// Link destinations to trip
		for idx, destName := range pt.destinations {
			dest := destinations[destName]
			tripDest := models.TripDestination{
				ID:            "td-" + pt.trip.ID + "-" + dest.ID,
				TripID:        pt.trip.ID,
				DestinationID: dest.ID,
				OrderIndex:    idx,
				CreatedAt:     now,
			}
			if err := db.Create(&tripDest).Error; err != nil {
				return err
			}
		}

		// Create sample day plans and activities for each trip
		if err := seedTripItinerary(db, &pt.trip, destinations, now); err != nil {
			return err
		}
	}

	log.Printf("✅ Seeded %d public trips with itineraries", len(publicTrips))
	return nil
}

func seedTripItinerary(db *gorm.DB, trip *models.Trip, destinations map[string]models.Destination, now time.Time) error {
	// Create sample activities (these will be reused across trips)
	sampleActivities := []models.Activity{
		{
			ID:              "act-senso-ji",
			Title:           "Visit Senso-ji Temple",
			Type:            "culture",
			Location:        ptr("Asakusa"),
			Latitude:        ptr(35.7148),
			Longitude:       ptr(139.7967),
			Description:     ptr("Tokyo's oldest and most famous temple"),
			DurationMinutes: ptr(90),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-tsukiji-market",
			Title:           "Tsukiji Outer Market Food Tour",
			Type:            "meal",
			Location:        ptr("Tsukiji"),
			Latitude:        ptr(35.6655),
			Longitude:       ptr(139.7708),
			Description:     ptr("Fresh sushi and street food experience"),
			DurationMinutes: ptr(120),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-shibuya-crossing",
			Title:           "Shibuya Crossing Experience",
			Type:            "experience",
			Location:        ptr("Shibuya"),
			Latitude:        ptr(35.6595),
			Longitude:       ptr(139.7004),
			Description:     ptr("World's busiest pedestrian crossing"),
			DurationMinutes: ptr(60),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-fushimi-inari",
			Title:           "Fushimi Inari Shrine",
			Type:            "culture",
			Location:        ptr("Fushimi"),
			Latitude:        ptr(34.9671),
			Longitude:       ptr(135.7727),
			Description:     ptr("Famous shrine with thousands of red torii gates"),
			DurationMinutes: ptr(120),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-arashiyama-bamboo",
			Title:           "Arashiyama Bamboo Grove",
			Type:            "culture",
			Location:        ptr("Arashiyama"),
			Latitude:        ptr(35.0094),
			Longitude:       ptr(135.6686),
			Description:     ptr("Stunning bamboo forest path"),
			DurationMinutes: ptr(90),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-osaka-castle",
			Title:           "Osaka Castle Visit",
			Type:            "culture",
			Location:        ptr("Chuo Ward"),
			Latitude:        ptr(34.6873),
			Longitude:       ptr(135.5262),
			Description:     ptr("Historic castle with panoramic city views"),
			DurationMinutes: ptr(120),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-dotonbori",
			Title:           "Dotonbori Food Street",
			Type:            "meal",
			Location:        ptr("Namba"),
			Latitude:        ptr(34.6686),
			Longitude:       ptr(135.5006),
			Description:     ptr("Osaka's famous entertainment and food district"),
			DurationMinutes: ptr(150),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-nagoya-castle",
			Title:           "Nagoya Castle",
			Type:            "culture",
			Location:        ptr("Nagoya"),
			Latitude:        ptr(35.1856),
			Longitude:       ptr(136.8998),
			Description:     ptr("Historic castle with golden shachihoko"),
			DurationMinutes: ptr(120),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-atsuta-shrine",
			Title:           "Atsuta Shrine",
			Type:            "culture",
			Location:        ptr("Nagoya"),
			Latitude:        ptr(35.1280),
			Longitude:       ptr(136.9083),
			Description:     ptr("One of Japan's most important Shinto shrines"),
			DurationMinutes: ptr(90),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-peace-memorial",
			Title:           "Hiroshima Peace Memorial Park",
			Type:            "culture",
			Location:        ptr("Hiroshima"),
			Latitude:        ptr(34.3955),
			Longitude:       ptr(132.4536),
			Description:     ptr("Memorial dedicated to peace and the atomic bombing victims"),
			DurationMinutes: ptr(120),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              "act-miyajima",
			Title:           "Miyajima Island & Itsukushima Shrine",
			Type:            "culture",
			Location:        ptr("Miyajima"),
			Latitude:        ptr(34.2959),
			Longitude:       ptr(132.3197),
			Description:     ptr("Famous floating torii gate and sacred island"),
			DurationMinutes: ptr(240),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}

	// Create activities if they don't exist
	for _, act := range sampleActivities {
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&act).Error; err != nil {
			return err
		}
	}

	// Parse trip dates
	startDate, _ := time.Parse("2006-01-02", trip.StartDate)
	endDate, _ := time.Parse("2006-01-02", trip.EndDate)

	// Calculate number of days
	durationDays := int(endDate.Sub(startDate).Hours()/24) + 1

	// Create day plans based on trip duration
	for i := 0; i < durationDays; i++ {
		currentDate := startDate.AddDate(0, 0, i)
		dayPlan := models.DayPlan{
			ID:        fmt.Sprintf("%s-day-%d", trip.ID, i+1),
			TripID:    trip.ID,
			Date:      currentDate.Format("2006-01-02"),
			DayNumber: i + 1,
			Notes:     ptr(fmt.Sprintf("Day %d of the journey", i+1)),
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := db.Create(&dayPlan).Error; err != nil {
			return err
		}

		// Determine which destination for this day
		var destID string
		if trip.ID == "pt-tokyo-week-discovery" {
			destID = destinations["Tokyo"].ID
		} else if trip.ID == "pt-tokyo-kyoto-10days" {
			if i < 5 {
				destID = destinations["Tokyo"].ID
			} else {
				destID = destinations["Kyoto"].ID
			}
		} else if trip.ID == "pt-kansai-grand-tour-14days" {
			if i < 5 {
				destID = destinations["Tokyo"].ID
			} else if i < 10 {
				destID = destinations["Kyoto"].ID
			} else {
				destID = destinations["Osaka"].ID
			}
		} else if trip.ID == "pt-japan-grand-adventure-28days" {
			// 28 days across 5 cities: Tokyo (6 days), Kyoto (7 days), Osaka (5 days), Nagoya (5 days), Hiroshima (5 days)
			if i < 6 {
				destID = destinations["Tokyo"].ID
			} else if i < 13 {
				destID = destinations["Kyoto"].ID
			} else if i < 18 {
				destID = destinations["Osaka"].ID
			} else if i < 23 {
				destID = destinations["Nagoya"].ID
			} else {
				destID = destinations["Hiroshima"].ID
			}
		}

		// Link destination to day
		dayPlanDest := models.DayPlanDestination{
			ID:            fmt.Sprintf("dpd-%s-%d", trip.ID, i+1),
			DayPlanID:     dayPlan.ID,
			DestinationID: destID,
			OrderIndex:    0,
			PartOfDay:     ptr("all-day"),
			CreatedAt:     now,
		}
		if err := db.Create(&dayPlanDest).Error; err != nil {
			return err
		}

		// Add 2-3 activities per day
		activityCount := 2 + (i % 2) // Alternates between 2 and 3 activities
		for j := 0; j < activityCount; j++ {
			var activityID string
			timeOfDay := "start"

			// Select appropriate activity based on destination and time
			if destID == destinations["Tokyo"].ID {
				activityID = []string{"act-senso-ji", "act-tsukiji-market", "act-shibuya-crossing"}[j%3]
			} else if destID == destinations["Kyoto"].ID {
				activityID = []string{"act-fushimi-inari", "act-arashiyama-bamboo"}[j%2]
			} else if destID == destinations["Osaka"].ID {
				activityID = []string{"act-osaka-castle", "act-dotonbori"}[j%2]
			} else if destID == destinations["Nagoya"].ID {
				activityID = []string{"act-nagoya-castle", "act-atsuta-shrine"}[j%2]
			} else if destID == destinations["Hiroshima"].ID {
				activityID = []string{"act-peace-memorial", "act-miyajima"}[j%2]
			} else {
				// Fallback for any other destination
				activityID = []string{"act-senso-ji", "act-tsukiji-market"}[j%2]
			}

			if j == 0 {
				timeOfDay = "start"
			} else if j == 1 {
				timeOfDay = "mid"
			} else {
				timeOfDay = "end"
			}

			dayPlanActivity := models.DayPlanActivity{
				ID:              fmt.Sprintf("dpa-%s-%d-%d", trip.ID, i+1, j+1),
				DayPlanID:       dayPlan.ID,
				ActivityID:      activityID,
				TimeOfDay:       timeOfDay,
				OrderWithinTime: 0,
				CreatedAt:       now,
				UpdatedAt:       now,
			}
			if err := db.Create(&dayPlanActivity).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func ptr[T any](v T) *T {
	return &v
}
