package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type TestGuest struct {
	GuestID          string   `dynamodbav:"guest_id"`
	InvitationCode   string   `dynamodbav:"invitation_code"`
	PrimaryGuest     string   `dynamodbav:"primary_guest"`
	HouseholdMembers []string `dynamodbav:"household_members"`
	MaxPartySize     int      `dynamodbav:"max_party_size"`
	Email            string   `dynamodbav:"email"`
	CreatedAt        string   `dynamodbav:"created_at"`
	UpdatedAt        string   `dynamodbav:"updated_at"`
}

func main() {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a custom endpoint resolver for local DynamoDB
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	fmt.Println("Seeding local DynamoDB with test guests...")
	fmt.Println()

	testGuests := []TestGuest{
		{
			GuestID:          "test-guest-001",
			InvitationCode:   "TESTCODE",
			PrimaryGuest:     "John Smith",
			HouseholdMembers: []string{},
			MaxPartySize:     1,
			Email:            "john@example.com",
			CreatedAt:        "2024-01-01T00:00:00Z",
			UpdatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			GuestID:          "test-guest-002",
			InvitationCode:   "COUPLE23",
			PrimaryGuest:     "Jane & Bob Johnson",
			HouseholdMembers: []string{},
			MaxPartySize:     2,
			Email:            "jane@example.com",
			CreatedAt:        "2024-01-01T00:00:00Z",
			UpdatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			GuestID:          "test-guest-003",
			InvitationCode:   "FAMILY44",
			PrimaryGuest:     "The Martinez Family",
			HouseholdMembers: []string{"Carlos Martinez", "Maria Martinez", "Sofia Martinez"},
			MaxPartySize:     4,
			Email:            "martinez@example.com",
			CreatedAt:        "2024-01-01T00:00:00Z",
			UpdatedAt:        "2024-01-01T00:00:00Z",
		},
	}

	for _, guest := range testGuests {
		fmt.Printf("Adding test guest: %s (code: %s)\n", guest.PrimaryGuest, guest.InvitationCode)

		item, err := attributevalue.MarshalMap(guest)
		if err != nil {
			log.Fatalf("Failed to marshal guest: %v", err)
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String("thedrewzers-wedding-guests"),
			Item:      item,
		})
		if err != nil {
			log.Fatalf("Failed to put item: %v", err)
		}
	}

	fmt.Println()
	fmt.Println("Done! Test guests added.")
	fmt.Println()
	fmt.Println("Test invitation codes:")
	fmt.Println("  TESTCODE  - John Smith (1 guest max)")
	fmt.Println("  COUPLE23  - Jane & Bob Johnson (2 guests max)")
	fmt.Println("  FAMILY44  - The Martinez Family (4 guests max)")
	fmt.Println()
	fmt.Println("Try: http://localhost:8080/rsvp?code=TESTCODE")
}
