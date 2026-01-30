package main

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"

	"github.com/apkiernan/thedrewzers/internal/auth"
	dbdynamo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
	"github.com/apkiernan/thedrewzers/internal/handlers"
	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/services"
	"github.com/apkiernan/thedrewzers/internal/views"
)

func main() {
	// Load .env file if present (ignored in production)
	_ = godotenv.Load()

	ctx := context.Background()

	// Initialize DynamoDB client
	var dynamoClient *dynamodb.Client
	var dbConfig dbdynamo.Config

	// Only initialize DynamoDB if AWS credentials are available
	if hasAWSCredentials() {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			logger.Warn("could not load AWS config", "error", err)
		} else {
			dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
				// Support local DynamoDB endpoint
				if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
					o.BaseEndpoint = &endpoint
				}
			})
			dbConfig = dbdynamo.ConfigFromEnv()
			logger.Info("dynamodb initialized")
		}
	} else {
		logger.Warn("AWS credentials not found - database functionality disabled")
	}

	server := http.NewServeMux()

	// Serve static files from dist/ directory (includes optimized images)
	fs := http.FileServer(http.Dir("./dist"))
	server.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Set up public routes
	setupPublicRoutes(server, dynamoClient, dbConfig)

	// Set up admin routes (if DynamoDB is available)
	if dynamoClient != nil {
		setupAdminRoutes(server, dynamoClient, dbConfig)
	}

	logger.Info("server started", "port", 8080, "static_dir", "./dist")
	if err := http.ListenAndServe(":8080", server); err != nil {
		logger.Error("server failed to start", "error", err)
		panic(err)
	}
}

// setupPublicRoutes configures routes for the public wedding website
func setupPublicRoutes(server *http.ServeMux, dynamoClient *dynamodb.Client, dbConfig dbdynamo.Config) {
	// Page routes
	server.HandleFunc("GET /", handlers.HandleHomePage)
	server.HandleFunc("GET /gallery", handlers.HandleGalleryPage)
	server.HandleFunc("GET /wedding-party", handlers.HandleWeddingPartyPage)

	// RSVP routes
	if dynamoClient != nil {
		guestRepo := dbdynamo.NewGuestRepository(dynamoClient, dbConfig.GuestsTable)
		rsvpRepo := dbdynamo.NewRSVPRepository(dynamoClient, dbConfig.RSVPsTable)
		rsvpHandler := handlers.NewRSVPHandler(guestRepo, rsvpRepo)

		server.HandleFunc("GET /rsvp", rsvpHandler.HandleRSVPPage)
		server.HandleFunc("GET /rsvp/form", rsvpHandler.HandleRSVPForm)
		server.HandleFunc("GET /rsvp/success", rsvpHandler.HandleRSVPSuccess)
		server.HandleFunc("POST /api/rsvp/search", rsvpHandler.HandleRSVPSearch)
		server.HandleFunc("POST /api/rsvp/submit", rsvpHandler.HandleRSVPSubmit)
		logger.Info("rsvp routes enabled", "guests_table", dbConfig.GuestsTable, "rsvps_table", dbConfig.RSVPsTable)
	} else {
		// Fallback handlers when DynamoDB is not available
		server.HandleFunc("GET /rsvp", func(w http.ResponseWriter, r *http.Request) {
			views.App(views.RSVPNameSearch()).Render(r.Context(), w)
		})
		server.HandleFunc("GET /rsvp/form", func(w http.ResponseWriter, r *http.Request) {
			views.App(views.RSVPNotFound()).Render(r.Context(), w)
		})
		server.HandleFunc("POST /api/rsvp/search", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error": "RSVP system not configured"}`, http.StatusServiceUnavailable)
		})
		server.HandleFunc("GET /rsvp/success", func(w http.ResponseWriter, r *http.Request) {
			views.App(views.RSVPSuccess()).Render(r.Context(), w)
		})
		server.HandleFunc("POST /api/rsvp/submit", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error": "RSVP system not configured"}`, http.StatusServiceUnavailable)
		})
		logger.Warn("rsvp routes disabled", "reason", "no database")
	}
}

// setupAdminRoutes configures routes for the admin dashboard
func setupAdminRoutes(server *http.ServeMux, dynamoClient *dynamodb.Client, dbConfig dbdynamo.Config) {
	// Initialize JWT service
	jwtService, err := auth.NewJWTService()
	if err != nil {
		logger.Error("failed to initialize JWT service", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	adminRepo := dbdynamo.NewAdminRepository(dynamoClient, dbConfig.AdminsTable)
	guestRepo := dbdynamo.NewGuestRepository(dynamoClient, dbConfig.GuestsTable)
	rsvpRepo := dbdynamo.NewRSVPRepository(dynamoClient, dbConfig.RSVPsTable)

	// Initialize services
	statsService := services.NewStatsService(guestRepo, rsvpRepo)

	// Initialize handlers
	authHandler := handlers.NewAdminAuthHandler(adminRepo, jwtService)
	dashboardHandler := handlers.NewAdminDashboardHandler(statsService)

	// Public admin routes (no auth required)
	server.HandleFunc("GET /login", authHandler.HandleLoginPage)
	server.HandleFunc("POST /login", authHandler.HandleLoginSubmit)
	server.HandleFunc("GET /logout", authHandler.HandleLogout)

	// Protected admin routes (auth required)
	requireAuth := auth.RequireAuth(jwtService)

	// Dashboard routes
	server.Handle("GET /dashboard", requireAuth(http.HandlerFunc(dashboardHandler.HandleDashboard)))
	server.Handle("GET /guests", requireAuth(http.HandlerFunc(dashboardHandler.HandleGuests)))
	server.Handle("GET /rsvps/export", requireAuth(http.HandlerFunc(dashboardHandler.HandleExportCSV)))

	logger.Info("admin routes enabled", "admins_table", dbConfig.AdminsTable)
}

// hasAWSCredentials checks if AWS credentials are likely available
func hasAWSCredentials() bool {
	// Check common environment variables
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		return true
	}
	// Check for AWS profile
	if os.Getenv("AWS_PROFILE") != "" {
		return true
	}
	// Check for default credentials file
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(home + "/.aws/credentials"); err == nil {
			return true
		}
	}
	return false
}
