#!/bin/bash
# Setup local DynamoDB tables for development/testing
# Usage: ./scripts/setup-local-db.sh

set -e

# Disable AWS CLI pager
export AWS_PAGER=""

ENDPOINT="http://localhost:8000"
REGION="us-east-1"

# Force dummy credentials for local DynamoDB (bypasses real AWS credential resolution)
export AWS_ACCESS_KEY_ID="local"
export AWS_SECRET_ACCESS_KEY="local"
export AWS_DEFAULT_REGION="us-east-1"

echo "Setting up local DynamoDB tables..."
echo "Endpoint: $ENDPOINT"
echo ""

# Check if DynamoDB Local is running
if ! curl -s "$ENDPOINT" > /dev/null 2>&1; then
    echo "Error: DynamoDB Local is not running at $ENDPOINT"
    echo "Start it with: docker-compose up -d"
    exit 1
fi

# Create guests table
echo "Creating guests table..."
aws dynamodb create-table \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --table-name "thedrewzers-wedding-guests" \
    --attribute-definitions \
        AttributeName=guest_id,AttributeType=S \
        AttributeName=invitation_code,AttributeType=S \
    --key-schema \
        AttributeName=guest_id,KeyType=HASH \
    --global-secondary-indexes \
        '[{"IndexName":"invitation_code_index","KeySchema":[{"AttributeName":"invitation_code","KeyType":"HASH"}],"Projection":{"ProjectionType":"ALL"}}]' \
    --billing-mode PAY_PER_REQUEST \
    2>&1 | grep -v "ResourceInUseException" || true
echo "  Done"

# Create RSVPs table
echo "Creating rsvps table..."
aws dynamodb create-table \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --table-name "thedrewzers-wedding-rsvps" \
    --attribute-definitions \
        AttributeName=rsvp_id,AttributeType=S \
        AttributeName=guest_id,AttributeType=S \
    --key-schema \
        AttributeName=rsvp_id,KeyType=HASH \
        AttributeName=guest_id,KeyType=RANGE \
    --billing-mode PAY_PER_REQUEST \
    2>&1 | grep -v "ResourceInUseException" || true
echo "  Done"

# Create admins table
echo "Creating admins table..."
aws dynamodb create-table \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --table-name "thedrewzers-wedding-admins" \
    --attribute-definitions \
        AttributeName=email,AttributeType=S \
    --key-schema \
        AttributeName=email,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    2>&1 | grep -v "ResourceInUseException" || true
echo "  Done"

echo ""
echo "Done! Tables created:"
aws dynamodb list-tables --endpoint-url "$ENDPOINT" --region "$REGION"

echo ""
echo "To use local DynamoDB, set this environment variable:"
echo "  export DYNAMODB_ENDPOINT=http://localhost:8000"
