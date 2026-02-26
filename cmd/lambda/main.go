package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"

	"github.com/apkiernan/thedrewzers/internal/auth"
	dbdynamo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
	"github.com/apkiernan/thedrewzers/internal/handlers"
	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/services"
)

var publicAdapter *httpadapter.HandlerAdapter
var adminAdapter *httpadapter.HandlerAdapter

// init initializes the Lambda handler
func init() {
	ctx := context.Background()

	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("failed to load AWS config: " + err.Error())
	}

	// Initialize DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Get table names from environment variables
	guestsTable := os.Getenv("GUESTS_TABLE")
	if guestsTable == "" {
		guestsTable = "thedrewzers-wedding-guests"
	}
	rsvpsTable := os.Getenv("RSVPS_TABLE")
	if rsvpsTable == "" {
		rsvpsTable = "thedrewzers-wedding-rsvps"
	}
	adminsTable := os.Getenv("ADMINS_TABLE")
	if adminsTable == "" {
		adminsTable = "thedrewzers-wedding-admins"
	}

	// Initialize repositories
	guestRepo := dbdynamo.NewGuestRepository(dynamoClient, guestsTable)
	rsvpRepo := dbdynamo.NewRSVPRepository(dynamoClient, rsvpsTable)
	adminRepo := dbdynamo.NewAdminRepository(dynamoClient, adminsTable)

	// Initialize handlers
	rsvpHandler := handlers.NewRSVPHandler(guestRepo, rsvpRepo)

	// Create public server mux
	publicServer := http.NewServeMux()
	publicServer.HandleFunc("POST /api/rsvp/submit", rsvpHandler.HandleRSVPSubmit)
	publicServer.HandleFunc("GET /api/health", handleHealthCheck)
	publicAdapter = httpadapter.New(publicServer)

	// Create admin server mux
	adminServer := http.NewServeMux()
	statsService := services.NewStatsService(guestRepo, rsvpRepo)
	setupAdminRoutes(adminServer, adminRepo, guestRepo, statsService)
	adminAdapter = httpadapter.New(adminServer)
}

// setupAdminRoutes configures routes for the admin dashboard
func setupAdminRoutes(server *http.ServeMux, adminRepo *dbdynamo.AdminRepository, guestRepo *dbdynamo.GuestRepository, statsService *services.StatsService) {
	// Initialize JWT service
	jwtService, err := auth.NewJWTService()
	if err != nil {
		logger.Warn("JWT service not initialized", "error", err)
		server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Admin authentication not configured", http.StatusServiceUnavailable)
		})
		return
	}

	// Initialize handlers
	authHandler := handlers.NewAdminAuthHandler(adminRepo, jwtService)
	dashboardHandler := handlers.NewAdminDashboardHandler(statsService, guestRepo)

	// Public admin routes (no auth required)
	server.HandleFunc("GET /login", authHandler.HandleLoginPage)
	server.HandleFunc("POST /login", authHandler.HandleLoginSubmit)
	server.HandleFunc("GET /logout", authHandler.HandleLogout)
	server.HandleFunc("GET /api/health", handleHealthCheck)

	// Protected admin routes (auth required)
	requireAuth := auth.RequireAuth(jwtService)

	server.Handle("GET /dashboard", requireAuth(http.HandlerFunc(dashboardHandler.HandleDashboard)))
	server.Handle("GET /guests", requireAuth(http.HandlerFunc(dashboardHandler.HandleGuests)))
	server.Handle("GET /guests/add", requireAuth(http.HandlerFunc(dashboardHandler.HandleAddGuests)))
	server.Handle("GET /guests/{id}", requireAuth(http.HandlerFunc(dashboardHandler.HandleGuestDetail)))
	server.Handle("POST /guests/create", requireAuth(http.HandlerFunc(dashboardHandler.HandleCreateGuest)))
	server.Handle("POST /guests/import", requireAuth(http.HandlerFunc(dashboardHandler.HandleImportCSV)))
	server.Handle("GET /rsvps/export", requireAuth(http.HandlerFunc(dashboardHandler.HandleExportCSV)))
	server.Handle("GET /", requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})))
}

// handleHealthCheck provides a simple health check endpoint
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "thedrewzers-wedding-api",
	})
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get host header (check both cases for compatibility)
	host := req.Headers["Host"]
	if host == "" {
		host = req.Headers["host"]
	}

	// Route to admin handler if accessing admin subdomain.
	// Check custom header set by admin CloudFront distribution,
	// since CloudFront replaces the Host header with the origin domain.
	isAdmin := strings.HasPrefix(host, "admin.") ||
		req.Headers["x-admin-request"] == "true" ||
		req.Headers["X-Admin-Request"] == "true"
	if isAdmin {
		return adminAdapter.ProxyWithContext(ctx, req)
	}

	// For public site, only process API routes
	if !strings.HasPrefix(req.Path, "/api/") {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error": "Not Found"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}

	// Process the public API request
	return publicAdapter.ProxyWithContext(ctx, req)
}

func main() {
	// Start the Lambda handler
	lambda.Start(Handler)
}
