package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/apkiernan/thedrewzers/internal/models"
)

// AdminRepository implements db.AdminRepository using DynamoDB
type AdminRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewAdminRepository creates a new AdminRepository
func NewAdminRepository(client *dynamodb.Client, tableName string) *AdminRepository {
	return &AdminRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetAdminByEmail retrieves an admin user by their email address
func (r *AdminRepository) GetAdminByEmail(ctx context.Context, email string) (*models.AdminUser, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("getting admin %s: %w", email, err)
	}

	if result.Item == nil {
		return nil, models.ErrAdminNotFound
	}

	var admin models.AdminUser
	if err := attributevalue.UnmarshalMap(result.Item, &admin); err != nil {
		return nil, fmt.Errorf("unmarshaling admin: %w", err)
	}

	return &admin, nil
}

// CreateAdmin creates a new admin user
func (r *AdminRepository) CreateAdmin(ctx context.Context, admin *models.AdminUser) error {
	now := time.Now().UTC()
	admin.CreatedAt = now
	admin.UpdatedAt = now

	item, err := attributevalue.MarshalMap(admin)
	if err != nil {
		return fmt.Errorf("marshaling admin: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(email)"),
	})
	if err != nil {
		return fmt.Errorf("creating admin: %w", err)
	}

	return nil
}

// UpdateAdmin updates an existing admin user
func (r *AdminRepository) UpdateAdmin(ctx context.Context, admin *models.AdminUser) error {
	admin.UpdatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(admin)
	if err != nil {
		return fmt.Errorf("marshaling admin: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(email)"),
	})
	if err != nil {
		return fmt.Errorf("updating admin: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the last login timestamp for an admin
func (r *AdminRepository) UpdateLastLogin(ctx context.Context, email string) error {
	now := time.Now().UTC()

	_, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"email": &types.AttributeValueMemberS{Value: email},
		},
		UpdateExpression: aws.String("SET last_login = :now, updated_at = :now"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":now": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
		},
		ConditionExpression: aws.String("attribute_exists(email)"),
	})
	if err != nil {
		return fmt.Errorf("updating last login for %s: %w", email, err)
	}

	return nil
}
