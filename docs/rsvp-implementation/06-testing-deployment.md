# Phase 6: Testing and Deployment

## Overview
This phase covers comprehensive testing strategies, deployment procedures, and monitoring setup for the RSVP system.

## Prerequisites
- All previous phases completed
- AWS credentials configured
- Production domain verified

## Step 1: Unit Testing

### 1.1 Test Guest Repository
Create `internal/db/dynamodb/guest_repository_test.go`:

```go
package dynamodb

import (
    "context"
    "testing"
    "time"
    
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "github.com/apkiernan/thedrewzers/internal/models"
)

type mockDynamoDBClient struct {
    mock.Mock
}

func (m *mockDynamoDBClient) Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error) {
    args := m.Called(ctx, params)
    return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func TestGetGuestByInvitationCode(t *testing.T) {
    mockClient := new(mockDynamoDBClient)
    repo := &GuestRepository{
        client:    mockClient,
        tableName: "test-guests",
    }
    
    testGuest := models.Guest{
        GuestID:        "123",
        InvitationCode: "ABC12345",
        PrimaryGuest:   "Test Guest",
        MaxPartySize:   2,
    }
    
    // Mock successful query
    mockClient.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
        Items: []map[string]types.AttributeValue{
            {
                "guest_id":        &types.AttributeValueMemberS{Value: testGuest.GuestID},
                "invitation_code": &types.AttributeValueMemberS{Value: testGuest.InvitationCode},
                "primary_guest":   &types.AttributeValueMemberS{Value: testGuest.PrimaryGuest},
                "max_party_size":  &types.AttributeValueMemberN{Value: "2"},
            },
        },
    }, nil)
    
    // Test
    guest, err := repo.GetGuestByInvitationCode(context.Background(), "ABC12345")
    
    assert.NoError(t, err)
    assert.Equal(t, testGuest.GuestID, guest.GuestID)
    assert.Equal(t, testGuest.InvitationCode, guest.InvitationCode)
    mockClient.AssertExpectations(t)
}

func TestGetGuestByInvitationCode_NotFound(t *testing.T) {
    mockClient := new(mockDynamoDBClient)
    repo := &GuestRepository{
        client:    mockClient,
        tableName: "test-guests",
    }
    
    // Mock empty query result
    mockClient.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
        Items: []map[string]types.AttributeValue{},
    }, nil)
    
    // Test
    guest, err := repo.GetGuestByInvitationCode(context.Background(), "NOTFOUND")
    
    assert.Error(t, err)
    assert.Nil(t, guest)
    assert.Contains(t, err.Error(), "guest not found")
}
```

### 1.2 Test RSVP Handler
Create `internal/handlers/rsvp_test.go`:

```go
package handlers

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "github.com/apkiernan/thedrewzers/internal/models"
)

type mockGuestRepo struct {
    mock.Mock
}

func (m *mockGuestRepo) GetGuestByInvitationCode(ctx context.Context, code string) (*models.Guest, error) {
    args := m.Called(ctx, code)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.Guest), args.Error(1)
}

type mockRSVPRepo struct {
    mock.Mock
}

func (m *mockRSVPRepo) CreateRSVP(ctx context.Context, rsvp *models.RSVP) error {
    args := m.Called(ctx, rsvp)
    return args.Error(0)
}

func (m *mockRSVPRepo) GetRSVP(ctx context.Context, guestID string) (*models.RSVP, error) {
    args := m.Called(ctx, guestID)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.RSVP), args.Error(1)
}

func TestHandleRSVPSubmit_Success(t *testing.T) {
    // Setup mocks
    mockGuests := new(mockGuestRepo)
    mockRSVPs := new(mockRSVPRepo)
    handler := NewRSVPHandler(mockGuests, mockRSVPs)
    
    testGuest := &models.Guest{
        GuestID:        "123",
        InvitationCode: "ABC12345",
        PrimaryGuest:   "Test Guest",
        MaxPartySize:   2,
    }
    
    // Mock guest lookup
    mockGuests.On("GetGuestByInvitationCode", mock.Anything, "ABC12345").Return(testGuest, nil)
    
    // Mock no existing RSVP
    mockRSVPs.On("GetRSVP", mock.Anything, "123").Return(nil, nil)
    
    // Mock RSVP creation
    mockRSVPs.On("CreateRSVP", mock.Anything, mock.Anything).Return(nil)
    
    // Create request
    reqBody := models.RSVPRequest{
        GuestID:        "123",
        InvitationCode: "ABC12345",
        Attending:      true,
        PartySize:      2,
        AttendeeNames:  []string{"Test Guest", "Plus One"},
    }
    
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/rsvp/submit", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute
    w := httptest.NewRecorder()
    handler.HandleRSVPSubmit(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.True(t, response["success"].(bool))
    
    mockGuests.AssertExpectations(t)
    mockRSVPs.AssertExpectations(t)
}

func TestHandleRSVPSubmit_InvalidCode(t *testing.T) {
    mockGuests := new(mockGuestRepo)
    mockRSVPs := new(mockRSVPRepo)
    handler := NewRSVPHandler(mockGuests, mockRSVPs)
    
    // Mock guest not found
    mockGuests.On("GetGuestByInvitationCode", mock.Anything, "INVALID").Return(nil, errors.New("not found"))
    
    // Create request
    reqBody := models.RSVPRequest{
        InvitationCode: "INVALID",
        Attending:      true,
    }
    
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/rsvp/submit", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    // Execute
    w := httptest.NewRecorder()
    handler.HandleRSVPSubmit(w, req)
    
    // Assert
    assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

## Step 2: Integration Testing

### 2.1 Create Integration Test Suite
Create `tests/integration/rsvp_flow_test.go`:

```go
package integration

import (
    "context"
    "net/http"
    "testing"
    "time"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/stretchr/testify/suite"
    
    "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
    "github.com/apkiernan/thedrewzers/internal/models"
)

type RSVPIntegrationSuite struct {
    suite.Suite
    dynamoClient *dynamodb.Client
    guestRepo    *dynamodb.GuestRepository
    rsvpRepo     *dynamodb.RSVPRepository
    baseURL      string
}

func (s *RSVPIntegrationSuite) SetupSuite() {
    // Setup DynamoDB client
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
    )
    s.Require().NoError(err)
    
    s.dynamoClient = dynamodb.NewFromConfig(cfg)
    s.guestRepo = dynamodb.NewGuestRepository(s.dynamoClient, "wedding-guests-test")
    s.rsvpRepo = dynamodb.NewRSVPRepository(s.dynamoClient, "wedding-rsvps-test")
    s.baseURL = "http://localhost:8080"
}

func (s *RSVPIntegrationSuite) TestCompleteRSVPFlow() {
    ctx := context.Background()
    
    // Create test guest
    testGuest := &models.Guest{
        InvitationCode:   "TEST12345",
        PrimaryGuest:     "Integration Test Guest",
        MaxPartySize:     2,
        Email:           "test@example.com",
        HouseholdMembers: []string{},
    }
    
    err := s.guestRepo.CreateGuest(ctx, testGuest)
    s.Require().NoError(err)
    
    // Test QR code redirect
    resp, err := http.Get(s.baseURL + "/rsvp?code=TEST12345")
    s.Require().NoError(err)
    s.Equal(http.StatusOK, resp.StatusCode)
    
    // Submit RSVP
    rsvpData := map[string]interface{}{
        "guest_id":        testGuest.GuestID,
        "invitation_code": "TEST12345",
        "attending":       true,
        "party_size":      2,
        "attendee_names":  []string{"Test Guest", "Plus One"},
    }
    
    // ... continue with API submission test
    
    // Verify RSVP was saved
    savedRSVP, err := s.rsvpRepo.GetRSVP(ctx, testGuest.GuestID)
    s.Require().NoError(err)
    s.NotNil(savedRSVP)
    s.True(savedRSVP.Attending)
    s.Equal(2, savedRSVP.PartySize)
    
    // Cleanup
    s.cleanupTestData(testGuest.GuestID)
}

func (s *RSVPIntegrationSuite) cleanupTestData(guestID string) {
    // Delete test data
    ctx := context.Background()
    s.guestRepo.DeleteGuest(ctx, guestID)
    s.rsvpRepo.DeleteRSVP(ctx, guestID)
}

func TestRSVPIntegrationSuite(t *testing.T) {
    suite.Run(t, new(RSVPIntegrationSuite))
}
```

## Step 3: End-to-End Testing

### 3.1 Create E2E Test Script
Create `tests/e2e/rsvp_e2e.js`:

```javascript
// Using Playwright for E2E testing
const { test, expect } = require('@playwright/test');

test.describe('RSVP Flow', () => {
  test('Complete RSVP submission', async ({ page }) => {
    // Navigate to RSVP page with code
    await page.goto('https://thedrewzers.com/rsvp?code=E2ETEST1');
    
    // Wait for page load
    await expect(page.locator('h1')).toContainText('RSVP');
    
    // Verify guest name is displayed
    await expect(page.locator('text=Hello, Test Guest!')).toBeVisible();
    
    // Select attending
    await page.click('input[value="yes"]');
    
    // Wait for details section
    await expect(page.locator('#attending-details')).toBeVisible();
    
    // Select party size
    await page.selectOption('select[name="party_size"]', '2');
    
    // Fill attendee names
    const nameInputs = page.locator('input[name="attendee_names[]"]');
    await nameInputs.nth(0).fill('Test Guest');
    await nameInputs.nth(1).fill('Plus One');
    
    // Add dietary restrictions
    await page.fill('textarea[name="dietary_restrictions"]', 'Vegetarian');
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Verify success
    await expect(page).toHaveURL(/.*\/rsvp\/success/);
    await expect(page.locator('text=Thank You!')).toBeVisible();
  });
  
  test('QR code scanning', async ({ page, context }) => {
    // Test mobile viewport
    await page.setViewportSize({ width: 375, height: 812 });
    
    // Navigate to QR code URL
    await page.goto('https://thedrewzers.com/rsvp?code=QRTEST01');
    
    // Verify mobile-friendly layout
    await expect(page.locator('.rsvp-container')).toBeVisible();
    
    // Test form is usable on mobile
    await page.click('input[value="yes"]');
    await expect(page.locator('#attending-details')).toBeVisible();
  });
});

test.describe('Admin Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Login to admin
    await page.goto('https://admin.thedrewzers.com/login');
    await page.fill('input[name="email"]', 'test@admin.com');
    await page.fill('input[name="password"]', 'testpassword');
    await page.click('button[type="submit"]');
    
    // Wait for redirect to dashboard
    await expect(page).toHaveURL(/.*\/dashboard/);
  });
  
  test('View dashboard statistics', async ({ page }) => {
    // Verify stats are displayed
    await expect(page.locator('text=Total Invited')).toBeVisible();
    await expect(page.locator('text=Responses')).toBeVisible();
    await expect(page.locator('text=Attending')).toBeVisible();
    
    // Check chart is rendered
    await expect(page.locator('#responseChart')).toBeVisible();
  });
  
  test('Export RSVP data', async ({ page }) => {
    // Navigate to RSVPs
    await page.click('a[href="/rsvps"]');
    
    // Click export button
    const [download] = await Promise.all([
      page.waitForEvent('download'),
      page.click('a[href="/rsvps/export"]')
    ]);
    
    // Verify download
    expect(download.suggestedFilename()).toMatch(/rsvps_.*\.csv/);
  });
});
```

## Step 4: Load Testing

### 4.1 Create Load Test Script
Create `tests/load/rsvp_load_test.js`:

```javascript
// Using k6 for load testing
import http from 'k6/http';
import { check, sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

export const options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up to 50 users
    { duration: '5m', target: 50 },   // Stay at 50 users
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests under 500ms
    http_req_failed: ['rate<0.1'],    // Error rate under 10%
  },
};

const BASE_URL = 'https://thedrewzers.com';

export default function () {
  // Test RSVP page load
  const rsvpPageRes = http.get(`${BASE_URL}/rsvp?code=LOAD${randomString(5)}`);
  check(rsvpPageRes, {
    'RSVP page loads': (r) => r.status === 200 || r.status === 404,
    'Page loads quickly': (r) => r.timings.duration < 1000,
  });
  
  sleep(1);
  
  // Test RSVP submission
  const payload = JSON.stringify({
    guest_id: randomString(10),
    invitation_code: `LOAD${randomString(5)}`,
    attending: Math.random() > 0.2, // 80% attendance rate
    party_size: Math.floor(Math.random() * 4) + 1,
    attendee_names: ['Test User'],
  });
  
  const headers = { 'Content-Type': 'application/json' };
  const submitRes = http.post(`${BASE_URL}/api/rsvp/submit`, payload, { headers });
  
  check(submitRes, {
    'RSVP submission': (r) => r.status === 200 || r.status === 400,
    'API responds quickly': (r) => r.timings.duration < 300,
  });
  
  sleep(2);
}
```

### 4.2 Run Load Tests
```bash
# Install k6
brew install k6

# Run load test
k6 run tests/load/rsvp_load_test.js

# Run with cloud reporting
k6 cloud tests/load/rsvp_load_test.js
```

## Step 5: Deployment Process

### 5.1 Create Deployment Script
Create `scripts/deploy.sh`:

```bash
#!/bin/bash
set -e

echo "🚀 Starting RSVP System Deployment"

# Check required environment variables
required_vars=("AWS_REGION" "JWT_SECRET" "DOMAIN_NAME")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Error: $var is not set"
        exit 1
    fi
done

# 1. Run tests
echo "📋 Running tests..."
go test ./... -v
if [ $? -ne 0 ]; then
    echo "❌ Tests failed. Deployment aborted."
    exit 1
fi

# 2. Build Lambda functions
echo "🔨 Building Lambda functions..."
make lambda-build

# 3. Generate Templ files
echo "🎨 Generating templates..."
make tpl

# 4. Build CSS
echo "💅 Building CSS..."
npm run build

# 5. Deploy infrastructure
echo "🏗️ Deploying infrastructure..."
cd terraform
terraform plan -out=tfplan
terraform apply tfplan
cd ..

# 6. Deploy Lambda
echo "⚡ Deploying Lambda function..."
aws lambda update-function-code \
    --function-name wedding-rsvp-api \
    --zip-file fileb://rsvp-lambda.zip

# 7. Deploy static assets
echo "📦 Deploying static assets..."
make upload-static

# 8. Invalidate CloudFront cache
echo "🔄 Invalidating CloudFront cache..."
make invalidate-cache

# 9. Run smoke tests
echo "🧪 Running smoke tests..."
./scripts/smoke-test.sh

echo "✅ Deployment complete!"
```

### 5.2 Create Smoke Test Script
Create `scripts/smoke-test.sh`:

```bash
#!/bin/bash

echo "🔥 Running smoke tests..."

# Test public site
echo "Testing public site..."
response=$(curl -s -o /dev/null -w "%{http_code}" https://thedrewzers.com)
if [ $response -eq 200 ]; then
    echo "✅ Public site is up"
else
    echo "❌ Public site returned $response"
    exit 1
fi

# Test RSVP page
echo "Testing RSVP page..."
response=$(curl -s -o /dev/null -w "%{http_code}" https://thedrewzers.com/rsvp)
if [ $response -eq 200 ]; then
    echo "✅ RSVP page is up"
else
    echo "❌ RSVP page returned $response"
    exit 1
fi

# Test admin site
echo "Testing admin site..."
response=$(curl -s -o /dev/null -w "%{http_code}" https://admin.thedrewzers.com/login)
if [ $response -eq 200 ]; then
    echo "✅ Admin site is up"
else
    echo "❌ Admin site returned $response"
    exit 1
fi

# Test API endpoint
echo "Testing API endpoint..."
response=$(curl -s -o /dev/null -w "%{http_code}" -X POST https://thedrewzers.com/api/rsvp/lookup \
    -H "Content-Type: application/json" \
    -d '{"code":"SMOKETEST"}')
if [ $response -eq 404 ] || [ $response -eq 200 ]; then
    echo "✅ API is responding"
else
    echo "❌ API returned $response"
    exit 1
fi

echo "✨ All smoke tests passed!"
```

## Step 6: Monitoring and Alerts

### 6.1 CloudWatch Alarms
Add to `terraform/monitoring.tf`:

```hcl
# Lambda error alarm
resource "aws_cloudwatch_metric_alarm" "lambda_errors" {
  alarm_name          = "${var.project_name}-lambda-errors"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "Errors"
  namespace           = "AWS/Lambda"
  period              = "300"
  statistic           = "Sum"
  threshold           = "10"
  alarm_description   = "Lambda function errors"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    FunctionName = aws_lambda_function.rsvp_api.function_name
  }
}

# DynamoDB throttling alarm
resource "aws_cloudwatch_metric_alarm" "dynamodb_throttles" {
  alarm_name          = "${var.project_name}-dynamodb-throttles"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "SystemErrors"
  namespace           = "AWS/DynamoDB"
  period              = "300"
  statistic           = "Sum"
  threshold           = "5"
  alarm_description   = "DynamoDB throttling"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    TableName = aws_dynamodb_table.wedding_guests.name
  }
}

# CloudFront 5xx errors
resource "aws_cloudwatch_metric_alarm" "cloudfront_5xx" {
  alarm_name          = "${var.project_name}-cloudfront-5xx"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "5xxErrorRate"
  namespace           = "AWS/CloudFront"
  period              = "300"
  statistic           = "Average"
  threshold           = "5"
  alarm_description   = "CloudFront 5xx error rate"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    DistributionId = aws_cloudfront_distribution.main.id
  }
}

# SNS topic for alerts
resource "aws_sns_topic" "alerts" {
  name = "${var.project_name}-alerts"
}

resource "aws_sns_topic_subscription" "email_alerts" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}
```

### 6.2 Application Metrics
Create `internal/metrics/metrics.go`:

```go
package metrics

import (
    "context"
    "time"
    
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch"
    "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

type MetricsClient struct {
    cloudwatch *cloudwatch.Client
    namespace  string
}

func NewMetricsClient(cw *cloudwatch.Client) *MetricsClient {
    return &MetricsClient{
        cloudwatch: cw,
        namespace:  "WeddingRSVP",
    }
}

func (m *MetricsClient) RecordRSVPSubmission(attending bool, partySize int) {
    metrics := []types.MetricDatum{
        {
            MetricName: aws.String("RSVPSubmissions"),
            Value:      aws.Float64(1),
            Unit:       types.StandardUnitCount,
            Timestamp:  aws.Time(time.Now()),
        },
    }
    
    if attending {
        metrics = append(metrics, types.MetricDatum{
            MetricName: aws.String("AttendingGuests"),
            Value:      aws.Float64(float64(partySize)),
            Unit:       types.StandardUnitCount,
            Timestamp:  aws.Time(time.Now()),
        })
    }
    
    m.cloudwatch.PutMetricData(context.TODO(), &cloudwatch.PutMetricDataInput{
        Namespace:  aws.String(m.namespace),
        MetricData: metrics,
    })
}

func (m *MetricsClient) RecordAPILatency(operation string, duration time.Duration) {
    m.cloudwatch.PutMetricData(context.TODO(), &cloudwatch.PutMetricDataInput{
        Namespace: aws.String(m.namespace),
        MetricData: []types.MetricDatum{
            {
                MetricName: aws.String("APILatency"),
                Value:      aws.Float64(duration.Milliseconds()),
                Unit:       types.StandardUnitMilliseconds,
                Timestamp:  aws.Time(time.Now()),
                Dimensions: []types.Dimension{
                    {
                        Name:  aws.String("Operation"),
                        Value: aws.String(operation),
                    },
                },
            },
        },
    })
}
```

## Step 7: Rollback Plan

### 7.1 Create Rollback Script
Create `scripts/rollback.sh`:

```bash
#!/bin/bash

echo "🔄 Starting rollback..."

# Get previous Lambda version
PREV_VERSION=$(aws lambda list-versions-by-function \
    --function-name wedding-rsvp-api \
    --max-items 2 \
    --query 'Versions[-2].Version' \
    --output text)

if [ "$PREV_VERSION" != "None" ]; then
    echo "Rolling back to Lambda version $PREV_VERSION"
    
    # Update alias to point to previous version
    aws lambda update-alias \
        --function-name wedding-rsvp-api \
        --name production \
        --function-version "$PREV_VERSION"
    
    echo "✅ Lambda rolled back"
else
    echo "❌ No previous version found"
fi

# Invalidate CloudFront
make invalidate-cache

echo "🧪 Running smoke tests..."
./scripts/smoke-test.sh
```

## Deployment Checklist

### Pre-Deployment
- [ ] All tests passing
- [ ] Code reviewed and approved
- [ ] Database backups taken
- [ ] Rollback plan reviewed
- [ ] Monitoring alerts configured

### Deployment Steps
- [ ] Deploy infrastructure changes
- [ ] Deploy Lambda functions
- [ ] Deploy static assets
- [ ] Run smoke tests
- [ ] Monitor error rates

### Post-Deployment
- [ ] Verify all features working
- [ ] Check CloudWatch metrics
- [ ] Monitor for 30 minutes
- [ ] Update documentation
- [ ] Notify stakeholders

## Troubleshooting Guide

### Common Issues

1. **RSVP submissions failing**
   - Check Lambda logs in CloudWatch
   - Verify DynamoDB permissions
   - Check API Gateway integration

2. **Admin login not working**
   - Verify JWT_SECRET is set
   - Check admin user exists in DynamoDB
   - Verify cookie domain settings

3. **QR codes not scanning**
   - Test QR code with multiple apps
   - Verify URL format is correct
   - Check for HTTPS redirects

4. **High latency**
   - Check Lambda cold starts
   - Review DynamoDB capacity
   - Verify CloudFront caching

## Success Metrics
- RSVP submission success rate > 99%
- Page load time < 2 seconds
- API response time < 300ms
- Zero data loss
- 99.9% uptime