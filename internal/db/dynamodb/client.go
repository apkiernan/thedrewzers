package dynamodb

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Config holds DynamoDB configuration
type Config struct {
	Region      string
	GuestsTable string
	RSVPsTable  string
	AdminsTable string
	Endpoint    string // For local testing with DynamoDB Local
}

// ConfigFromEnv creates Config from environment variables with sensible defaults
func ConfigFromEnv() Config {
	return Config{
		Region:      getEnvOrDefault("AWS_REGION", "us-east-1"),
		GuestsTable: getEnvOrDefault("GUESTS_TABLE", "thedrewzers-wedding-guests"),
		RSVPsTable:  getEnvOrDefault("RSVPS_TABLE", "thedrewzers-wedding-rsvps"),
		AdminsTable: getEnvOrDefault("ADMINS_TABLE", "thedrewzers-wedding-admins"),
		Endpoint:    os.Getenv("DYNAMODB_ENDPOINT"), // Empty for AWS, set for local
	}
}

// NewClient creates a DynamoDB client with the provided configuration
func NewClient(ctx context.Context, cfg Config) (*dynamodb.Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	opts := []func(*dynamodb.Options){}

	// Use custom endpoint for local development (DynamoDB Local)
	if cfg.Endpoint != "" {
		opts = append(opts, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		})
	}

	return dynamodb.NewFromConfig(awsCfg, opts...), nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
