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
	tablesTable := os.Getenv("TABLES_TABLE")
	if tablesTable == "" {
		tablesTable = "thedrewzers-wedding-tables"
	}

	// Initialize repositories
	guestRepo := dbdynamo.NewGuestRepository(dynamoClient, guestsTable)
	rsvpRepo := dbdynamo.NewRSVPRepository(dynamoClient, rsvpsTable)
	adminRepo := dbdynamo.NewAdminRepository(dynamoClient, adminsTable)
	tableRepo := dbdynamo.NewTableRepository(dynamoClient, tablesTable)

	// Initialize handlers
	rsvpHandler := handlers.NewRSVPHandler(guestRepo, rsvpRepo)

	// Create public server mux
	publicServer := http.NewServeMux()
	publicServer.HandleFunc("POST /api/rsvp/submit", rsvpHandler.HandleRSVPSubmit)
	publicServer.HandleFunc("GET /api/health", handleHealthCheck)
	publicAdapter = httpadapter.New(publicServer)

	// Create admin server mux
	adminServer := http.NewServeMux()
	setupAdminRoutes(adminServer, adminRepo, guestRepo, rsvpRepo, tableRepo)
	adminAdapter = httpadapter.New(adminServer)
}

// setupAdminRoutes configures routes for the admin dashboard
func setupAdminRoutes(server *http.ServeMux, adminRepo *dbdynamo.AdminRepository, guestRepo *dbdynamo.GuestRepository, rsvpRepo *dbdynamo.RSVPRepository, tableRepo *dbdynamo.TableRepository) {
	// Initialize JWT service
	jwtService, err := auth.NewJWTService()
	if err != nil {
		logger.Warn("JWT service not initialized", "error", err)
		// Return a handler that always returns an error
		server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Admin authentication not configured", http.StatusServiceUnavailable)
		})
		return
	}

	// Initialize auth handler
	authHandler := handlers.NewAdminAuthHandler(adminRepo, jwtService)

	// Initialize seating service and handler
	seatingService := services.NewSeatingService(tableRepo, guestRepo, rsvpRepo)
	seatingHandler := handlers.NewSeatingHandler(seatingService)

	// Public admin routes (no auth required)
	server.HandleFunc("GET /login", authHandler.HandleLoginPage)
	server.HandleFunc("POST /login", authHandler.HandleLoginSubmit)
	server.HandleFunc("GET /logout", authHandler.HandleLogout)
	server.HandleFunc("GET /api/health", handleHealthCheck)

	// Protected admin routes (auth required)
	requireAuth := auth.RequireAuth(jwtService)

	// Dashboard placeholder
	server.Handle("GET /dashboard", requireAuth(http.HandlerFunc(authHandler.HandleDashboardPlaceholder)))
	server.Handle("GET /", requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})))

	// Seating routes
	server.Handle("GET /seating", requireAuth(http.HandlerFunc(seatingHandler.HandleSeatingPage)))
	server.Handle("POST /api/seating/tables", requireAuth(http.HandlerFunc(seatingHandler.HandleCreateTable)))
	server.Handle("PUT /api/seating/tables/{id}", requireAuth(http.HandlerFunc(seatingHandler.HandleUpdateTable)))
	server.Handle("DELETE /api/seating/tables/{id}", requireAuth(http.HandlerFunc(seatingHandler.HandleDeleteTable)))
	server.Handle("POST /api/seating/assign", requireAuth(http.HandlerFunc(seatingHandler.HandleAssignGuest)))
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

	// Route to admin handler if accessing admin subdomain
	if strings.HasPrefix(host, "admin.") {
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
