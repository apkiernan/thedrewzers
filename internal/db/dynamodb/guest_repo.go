package dynamodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"github.com/apkiernan/thedrewzers/internal/models"
)

// GuestRepository implements db.GuestRepository using DynamoDB
type GuestRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewGuestRepository creates a new GuestRepository
func NewGuestRepository(client *dynamodb.Client, tableName string) *GuestRepository {
	return &GuestRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetGuest retrieves a guest by their unique ID
func (r *GuestRepository) GetGuest(ctx context.Context, guestID string) (*models.Guest, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"guest_id": &types.AttributeValueMemberS{Value: guestID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("getting guest %s: %w", guestID, err)
	}

	if result.Item == nil {
		return nil, models.ErrGuestNotFound
	}

	var guest models.Guest
	if err := attributevalue.UnmarshalMap(result.Item, &guest); err != nil {
		return nil, fmt.Errorf("unmarshaling guest: %w", err)
	}

	return &guest, nil
}

// GetGuestByInvitationCode retrieves a guest by their invitation code using the GSI
func (r *GuestRepository) GetGuestByInvitationCode(ctx context.Context, code string) (*models.Guest, error) {
	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("invitation_code_index"),
		KeyConditionExpression: aws.String("invitation_code = :code"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":code": &types.AttributeValueMemberS{Value: code},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("querying by invitation code: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, models.ErrGuestNotFound
	}

	var guest models.Guest
	if err := attributevalue.UnmarshalMap(result.Items[0], &guest); err != nil {
		return nil, fmt.Errorf("unmarshaling guest: %w", err)
	}

	return &guest, nil
}

// CreateGuest creates a new guest record
func (r *GuestRepository) CreateGuest(ctx context.Context, guest *models.Guest) error {
	if guest.GuestID == "" {
		guest.GuestID = uuid.New().String()
	}

	now := time.Now().UTC()
	guest.CreatedAt = now
	guest.UpdatedAt = now

	item, err := attributevalue.MarshalMap(guest)
	if err != nil {
		return fmt.Errorf("marshaling guest: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(guest_id)"),
	})
	if err != nil {
		return fmt.Errorf("creating guest: %w", err)
	}

	return nil
}

// UpdateGuest updates an existing guest record
func (r *GuestRepository) UpdateGuest(ctx context.Context, guest *models.Guest) error {
	guest.UpdatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(guest)
	if err != nil {
		return fmt.Errorf("marshaling guest: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(guest_id)"),
	})
	if err != nil {
		return fmt.Errorf("updating guest: %w", err)
	}

	return nil
}

// DeleteGuest removes a guest by their ID
func (r *GuestRepository) DeleteGuest(ctx context.Context, guestID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"guest_id": &types.AttributeValueMemberS{Value: guestID},
		},
	})
	if err != nil {
		return fmt.Errorf("deleting guest %s: %w", guestID, err)
	}
	return nil
}

// ListGuests returns all guests with pagination support for large datasets
func (r *GuestRepository) ListGuests(ctx context.Context) ([]*models.Guest, error) {
	var guests []*models.Guest
	var lastKey map[string]types.AttributeValue

	for {
		result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(r.tableName),
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			return nil, fmt.Errorf("scanning guests: %w", err)
		}

		var batch []*models.Guest
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshaling guests: %w", err)
		}

		guests = append(guests, batch...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	return guests, nil
}

// SearchGuestsByName finds guests matching the given name (case-insensitive, partial match)
func (r *GuestRepository) SearchGuestsByName(ctx context.Context, name string) ([]*models.Guest, error) {
	// Scan all guests (small dataset, scan is fine)
	allGuests, err := r.ListGuests(ctx)
	if err != nil {
		return nil, fmt.Errorf("searching guests by name: %w", err)
	}

	nameLower := strings.ToLower(strings.TrimSpace(name))
	if nameLower == "" {
		return nil, nil
	}

	var matches []*models.Guest
	for _, guest := range allGuests {
		// Check primary guest name
		if strings.Contains(strings.ToLower(guest.PrimaryGuest), nameLower) {
			matches = append(matches, guest)
			continue
		}

		// Check household members
		matched := false
		for _, member := range guest.HouseholdMembers {
			if strings.Contains(strings.ToLower(member), nameLower) {
				matched = true
				break
			}
		}
		if matched {
			matches = append(matches, guest)
		}
	}

	return matches, nil
}
