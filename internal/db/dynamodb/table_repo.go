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

// TableRepository implements db.TableRepository using DynamoDB
type TableRepository struct {
	client    *dynamodb.Client
	tableName string
}

// NewTableRepository creates a new TableRepository
func NewTableRepository(client *dynamodb.Client, tableName string) *TableRepository {
	return &TableRepository{
		client:    client,
		tableName: tableName,
	}
}

// GetTable retrieves a table by its unique ID
func (r *TableRepository) GetTable(ctx context.Context, tableID string) (*models.Table, error) {
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"table_id": &types.AttributeValueMemberS{Value: tableID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("getting table %s: %w", tableID, err)
	}

	if result.Item == nil {
		return nil, models.ErrTableNotFound
	}

	var table models.Table
	if err := attributevalue.UnmarshalMap(result.Item, &table); err != nil {
		return nil, fmt.Errorf("unmarshaling table: %w", err)
	}

	return &table, nil
}

// CreateTable creates a new table record
func (r *TableRepository) CreateTable(ctx context.Context, table *models.Table) error {
	if table.TableID == "" {
		table.TableID = uuid.New().String()
	}

	now := time.Now().UTC()
	table.CreatedAt = now
	table.UpdatedAt = now

	item, err := attributevalue.MarshalMap(table)
	if err != nil {
		return fmt.Errorf("marshaling table: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(table_id)"),
	})
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}

	return nil
}

// UpdateTable updates an existing table record
func (r *TableRepository) UpdateTable(ctx context.Context, table *models.Table) error {
	table.UpdatedAt = time.Now().UTC()

	item, err := attributevalue.MarshalMap(table)
	if err != nil {
		return fmt.Errorf("marshaling table: %w", err)
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_exists(table_id)"),
	})
	if err != nil {
		return fmt.Errorf("updating table: %w", err)
	}

	return nil
}

// DeleteTable removes a table by its ID
func (r *TableRepository) DeleteTable(ctx context.Context, tableID string) error {
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"table_id": &types.AttributeValueMemberS{Value: tableID},
		},
	})
	if err != nil {
		return fmt.Errorf("deleting table %s: %w", tableID, err)
	}
	return nil
}

// ListTables returns all tables with pagination support
func (r *TableRepository) ListTables(ctx context.Context) ([]*models.Table, error) {
	var tables []*models.Table
	var lastKey map[string]types.AttributeValue

	for {
		result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
			TableName:         aws.String(r.tableName),
			ExclusiveStartKey: lastKey,
		})
		if err != nil {
			return nil, fmt.Errorf("scanning tables: %w", err)
		}

		var batch []*models.Table
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &batch); err != nil {
			return nil, fmt.Errorf("unmarshaling tables: %w", err)
		}

		tables = append(tables, batch...)

		if result.LastEvaluatedKey == nil {
			break
		}
		lastKey = result.LastEvaluatedKey
	}

	return tables, nil
}
