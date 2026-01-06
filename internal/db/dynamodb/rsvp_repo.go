package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/apkiernan/thedrewzers/internal/models"
)

// RSVPRepository implements db.RSVPRepository using DynamoDB
type RSVPRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewRSVPRepository creates a new RSVPRepository
func NewRSVPRepository(client *dynamodb.Client, tableName string) *RSVPRepository {
	return &RSVPRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetRSVP retrieves an RSVP by its unique ID
// Note: The table uses rsvp_id as hash key, so we query by it
func (r *RSVPRepository) GetRSVP(ctx context.Context, rsvpID string) (*models.RSVP, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		KeyConditionExpression: aws.String("rsvp_id = :rsvp_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":rsvp_id": &types.AttributeValueMemberS{Value: rsvpID},
		},
		Limit: aws.Int32(1),
	})
	if err != nil {
		return nil, fmt.Errorf("querying rsvp %s: %w", rsvpID, err)
	}

	if len(result.Items) == 0 {
		return nil, models.ErrRSVPNotFound
	}

	var rsvp models.RSVP
	if err := attributevalue.UnmarshalMap(result.Items[0], &rsvp); err != nil {
		return nil, fmt.Errorf("unmarshaling rsvp: %w", err)
	}

	return &rsvp, nil
}

// GetRSVPByGuestID retrieves an RSVP by the guest's ID
// Note: This requires a scan since guest_id is the range key, not hash key
func (r *RSVPRepository) GetRSVPByGuestID(ctx context.Context, guestID string) (*models.RSVP, error) {
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(r.tableName),
		FilterExpression: aws.String("guest_id = :guest_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":guest_id": &types.AttributeValueMemberS{Value: guestID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("scanning for rsvp by guest %s: %w", guestID, err)
	}

	if len(result.Items) == 0 {
		return nil, models.ErrRSVPNotFound
	}

	var rsvp models.RSVP
	if err := attributevalue.UnmarshalMap(result.Items[0], &rsvp); err != nil {
		return nil, fmt.Errorf("unmarshaling rsvp: %w", err)
	}

	return &rsvp, nil
}

// CreateRSVP creates a new RSVP record
func (r *RSVPRepository) CreateRSVP(ctx context.Context, rsvp *models.RSVP) error {
	if rsvp.RSVPID == "" {
		rsvp.RSVPID = uuid.New().String()
	}

	now := time.Now().UTC()
	rsvp.SubmittedAt = now
	rsvp.UpdatedAt = now

	item, err := attributevalue.MarshalMap(rsvp)
	if err != nil {
		return fmt.Errorf("marshaling rsvp: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(rsvp_id) AND attribute_not_exists(guest_id)"),
	})
	if err != nil {
		return fmt.Errorf("creating rsvp: %w", err)
	}

	return nil
}

// UpdateRSVP updates an existing RSVP record
func (r *RSVPRepository) UpdateRSVP(ctx context.Context, rsvp *models.RSVP) error {
	rsvp.UpdatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(rsvp)
	if err != nil {
		return fmt.Errorf("marshaling rsvp: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(rsvp_id) AND attribute_exists(guest_id)"),
	})
	if err != nil {
		return fmt.Errorf("updating rsvp: %w", err)
	}

	return nil
}

// ListRSVPs returns all RSVPs with pagination support for large datasets
func (r *RSVPRepository) ListRSVPs(ctx context.Context) ([]*models.RSVP, error) {
	var rsvps []*models.RSVP
	var lastKey map[string]types.AttributeValue

	for {
		result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(r.tableName),
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			return nil, fmt.Errorf("scanning rsvps: %w", err)
		}

		var batch []*models.RSVP
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshaling rsvps: %w", err)
		}

		rsvps = append(rsvps, batch...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	return rsvps, nil
}
