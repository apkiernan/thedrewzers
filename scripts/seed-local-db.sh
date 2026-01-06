#!/bin/bash
# Seed local DynamoDB with test guest data
# Usage: ./scripts/seed-local-db.sh

set -e

# Disable AWS CLI pager
export AWS_PAGER=""

# Force dummy credentials for local DynamoDB (bypasses real AWS credential resolution)
export AWS_ACCESS_KEY_ID="local"
export AWS_SECRET_ACCESS_KEY="local"
export AWS_DEFAULT_REGION="us-east-1"

ENDPOINT="http://localhost:8000"
REGION="us-east-1"
TABLE="thedrewzers-wedding-guests"

echo "Seeding local DynamoDB with test guests..."
echo ""

# Check if DynamoDB Local is running
if ! curl -s "$ENDPOINT" > /dev/null 2>&1; then
    echo "Error: DynamoDB Local is not running at $ENDPOINT"
    echo "Start it with: docker-compose up -d"
    exit 1
fi

# Test Guest 1: Single guest
echo "Adding test guest: John Smith (code: TESTCODE)"
aws dynamodb put-item \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --table-name "$TABLE" \
    --item '{
        "guest_id": {"S": "test-guest-001"},
        "invitation_code": {"S": "TESTCODE"},
        "primary_guest": {"S": "John Smith"},
        "household_members": {"L": []},
        "max_party_size": {"N": "1"},
        "email": {"S": "john@example.com"},
        "created_at": {"S": "2024-01-01T00:00:00Z"},
        "updated_at": {"S": "2024-01-01T00:00:00Z"}
    }'

# Test Guest 2: Couple
echo "Adding test guest: Jane & Bob Johnson (code: COUPLE23)"
aws dynamodb put-item \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --table-name "$TABLE" \
    --item '{
        "guest_id": {"S": "test-guest-002"},
        "invitation_code": {"S": "COUPLE23"},
        "primary_guest": {"S": "Jane & Bob Johnson"},
        "household_members": {"L": []},
        "max_party_size": {"N": "2"},
        "email": {"S": "jane@example.com"},
        "created_at": {"S": "2024-01-01T00:00:00Z"},
        "updated_at": {"S": "2024-01-01T00:00:00Z"}
    }'

# Test Guest 3: Family
echo "Adding test guest: The Martinez Family (code: FAMILY44)"
aws dynamodb put-item \
    --endpoint-url "$ENDPOINT" \
    --region "$REGION" \
    --table-name "$TABLE" \
    --item '{
        "guest_id": {"S": "test-guest-003"},
        "invitation_code": {"S": "FAMILY44"},
        "primary_guest": {"S": "The Martinez Family"},
        "household_members": {"L": [{"S": "Carlos Martinez"}, {"S": "Maria Martinez"}, {"S": "Sofia Martinez"}]},
        "max_party_size": {"N": "4"},
        "email": {"S": "martinez@example.com"},
        "created_at": {"S": "2024-01-01T00:00:00Z"},
        "updated_at": {"S": "2024-01-01T00:00:00Z"}
    }'

echo ""
echo "Done! Test guests added."
echo ""
echo "Test invitation codes:"
echo "  TESTCODE  - John Smith (1 guest max)"
echo "  COUPLE23  - Jane & Bob Johnson (2 guests max)"
echo "  FAMILY44  - The Martinez Family (4 guests max)"
echo ""
echo "Try: http://localhost:8080/rsvp?code=TESTCODE"
