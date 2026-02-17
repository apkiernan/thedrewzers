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

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("local", "local", "")),
	)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	fmt.Printf("Clearing local DynamoDB data at %s\n", endpoint)

	guestsDeleted, err := clearTable(ctx, client, "thedrewzers-wedding-guests", []string{"guest_id"})
	if err != nil {
		log.Fatalf("Failed clearing guests table: %v", err)
	}
	fmt.Printf("  Deleted %d items from thedrewzers-wedding-guests\n", guestsDeleted)

	rsvpsDeleted, err := clearTable(ctx, client, "thedrewzers-wedding-rsvps", []string{"rsvp_id", "guest_id"})
	if err != nil {
		log.Fatalf("Failed clearing RSVPs table: %v", err)
	}
	fmt.Printf("  Deleted %d items from thedrewzers-wedding-rsvps\n", rsvpsDeleted)

	fmt.Println("Done.")
}

func clearTable(ctx context.Context, client *dynamodb.Client, table string, keyNames []string) (int, error) {
	var deleted int
	var lastKey map[string]types.AttributeValue

	projection := strings.Join(keyNames, ", ")

	for {
		result, err := client.Scan(ctx, &dynamodb.ScanInput{
			TableName:            aws.String(table),
			ProjectionExpression: aws.String(projection),
			ExclusiveStartKey:    lastKey,
		})
		if err != nil {
			return deleted, err
		}

		for _, item := range result.Items {
			key := make(map[string]types.AttributeValue, len(keyNames))
			for _, keyName := range keyNames {
				value, ok := item[keyName]
				if !ok {
					return deleted, fmt.Errorf("missing key field %s in table %s", keyName, table)
				}
				key[keyName] = value
			}

			if _, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
				TableName: aws.String(table),
				Key:       key,
			}); err != nil {
				return deleted, err
			}
			deleted++
		}

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	return deleted, nil
}
