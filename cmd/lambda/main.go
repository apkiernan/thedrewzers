package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

var httpAdapter *httpadapter.HandlerAdapter

// init initializes the Lambda handler
func init() {
	// Create a new HTTP server mux
	server := http.NewServeMux()

	// API routes only - no static file handling needed
	server.HandleFunc("POST /api/rsvp", handleRSVPSubmit)
	server.HandleFunc("GET /api/rsvp/{id}", handleRSVPGet)
	server.HandleFunc("GET /api/health", handleHealthCheck)

	// Create the adapter
	httpAdapter = httpadapter.New(server)
}

// handleHealthCheck provides a simple health check endpoint
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
		"service": "thedrewzers-wedding-api",
	})
}

// handleRSVPSubmit handles RSVP form submissions (placeholder for future implementation)
func handleRSVPSubmit(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement RSVP submission logic
	// - Parse form data
	// - Validate input
	// - Store in DynamoDB
	// - Send confirmation email via SES
	// - Return success response
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "RSVP functionality coming soon",
	})
}

// handleRSVPGet retrieves an RSVP by ID (placeholder for future implementation)
func handleRSVPGet(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	rsvpID := parts[3]
	
	// TODO: Implement RSVP retrieval logic
	// - Query DynamoDB
	// - Return RSVP details
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "RSVP functionality coming soon",
		"id": rsvpID,
	})
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Only process API routes
	if !strings.HasPrefix(req.Path, "/api/") {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       `{"error": "Not Found"}`,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
	
	// Process the request
	return httpAdapter.ProxyWithContext(ctx, req)
}

func main() {
	// Start the Lambda handler
	lambda.Start(Handler)
}