package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create client with BaseEndpoint option (the new recommended way)
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	fmt.Println("Setting up local DynamoDB tables...")
	fmt.Printf("Endpoint: %s\n\n", endpoint)

	// Create guests table
	if err := createGuestsTable(ctx, client); err != nil {
		log.Fatalf("Failed to create guests table: %v", err)
	}

	// Create RSVPs table
	if err := createRSVPsTable(ctx, client); err != nil {
		log.Fatalf("Failed to create RSVPs table: %v", err)
	}

	// Create admins table
	if err := createAdminsTable(ctx, client); err != nil {
		log.Fatalf("Failed to create admins table: %v", err)
	}

	// List tables
	fmt.Println("\nTables created:")
	result, err := client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatalf("Failed to list tables: %v", err)
	}
	for _, table := range result.TableNames {
		fmt.Printf("  - %s\n", table)
	}

	fmt.Println("\nLocal DynamoDB setup complete!")
}

func createGuestsTable(ctx context.Context, client *dynamodb.Client) error {
	fmt.Println("Creating guests table...")

	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("thedrewzers-wedding-guests"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("guest_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("invitation_code"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("guest_id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("invitation_code_index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("invitation_code"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		// Check if table already exists
		if isResourceInUseException(err) {
			fmt.Println("  Table already exists, skipping...")
			return nil
		}
		return err
	}

	fmt.Println("  Done")
	return nil
}

func createRSVPsTable(ctx context.Context, client *dynamodb.Client) error {
	fmt.Println("Creating RSVPs table...")

	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("thedrewzers-wedding-rsvps"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("rsvp_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("guest_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("rsvp_id"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("guest_id"),
				KeyType:       types.KeyTypeRange,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		if isResourceInUseException(err) {
			fmt.Println("  Table already exists, skipping...")
			return nil
		}
		return err
	}

	fmt.Println("  Done")
	return nil
}

func createAdminsTable(ctx context.Context, client *dynamodb.Client) error {
	fmt.Println("Creating admins table...")

	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("thedrewzers-wedding-admins"),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("email"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("email"),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	})

	if err != nil {
		if isResourceInUseException(err) {
			fmt.Println("  Table already exists, skipping...")
			return nil
		}
		return err
	}

	fmt.Println("  Done")
	return nil
}

func isResourceInUseException(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "ResourceInUseException") ||
		strings.Contains(errStr, "Table already exists")
}
