# Phase 2: Guest Data Model and QR Code Generation

## Overview
This phase implements the guest data model, DynamoDB access layer, and QR code generation system for invitations.

## Prerequisites
- Phase 1 infrastructure deployed
- Go development environment set up
- Access to DynamoDB tables

## Step 1: Guest Data Models

### 1.1 Create Core Types
Create `internal/models/guest.go`:

```go
package models

import (
    "time"
)

type Guest struct {
    GuestID          string    `json:"guest_id" dynamodbav:"guest_id"`
    InvitationCode   string    `json:"invitation_code" dynamodbav:"invitation_code"`
    PrimaryGuest     string    `json:"primary_guest" dynamodbav:"primary_guest"`
    HouseholdMembers []string  `json:"household_members" dynamodbav:"household_members"`
    MaxPartySize     int       `json:"max_party_size" dynamodbav:"max_party_size"`
    Email           string    `json:"email,omitempty" dynamodbav:"email,omitempty"`
    Phone           string    `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
    Address         Address   `json:"address,omitempty" dynamodbav:"address,omitempty"`
    CreatedAt       time.Time `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt       time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

type Address struct {
    Street  string `json:"street" dynamodbav:"street"`
    City    string `json:"city" dynamodbav:"city"`
    State   string `json:"state" dynamodbav:"state"`
    Zip     string `json:"zip" dynamodbav:"zip"`
    Country string `json:"country" dynamodbav:"country"`
}

type RSVP struct {
    RSVPID              string    `json:"rsvp_id" dynamodbav:"rsvp_id"`
    GuestID             string    `json:"guest_id" dynamodbav:"guest_id"`
    Attending           bool      `json:"attending" dynamodbav:"attending"`
    PartySize           int       `json:"party_size" dynamodbav:"party_size"`
    AttendeeNames       []string  `json:"attendee_names" dynamodbav:"attendee_names"`
    DietaryRestrictions []string  `json:"dietary_restrictions" dynamodbav:"dietary_restrictions"`
    SpecialRequests     string    `json:"special_requests" dynamodbav:"special_requests"`
    SubmittedAt         time.Time `json:"submitted_at" dynamodbav:"submitted_at"`
    UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
    IPAddress           string    `json:"ip_address" dynamodbav:"ip_address"`
    UserAgent           string    `json:"user_agent" dynamodbav:"user_agent"`
}
```

## Step 2: DynamoDB Repository

### 2.1 Create Database Interface
Create `internal/db/interface.go`:

```go
package db

import (
    "context"
    "github.com/apkiernan/thedrewzers/internal/models"
)

type GuestRepository interface {
    GetGuest(ctx context.Context, guestID string) (*models.Guest, error)
    GetGuestByInvitationCode(ctx context.Context, code string) (*models.Guest, error)
    CreateGuest(ctx context.Context, guest *models.Guest) error
    UpdateGuest(ctx context.Context, guest *models.Guest) error
    ListGuests(ctx context.Context) ([]*models.Guest, error)
}

type RSVPRepository interface {
    GetRSVP(ctx context.Context, guestID string) (*models.RSVP, error)
    CreateRSVP(ctx context.Context, rsvp *models.RSVP) error
    UpdateRSVP(ctx context.Context, rsvp *models.RSVP) error
    ListRSVPs(ctx context.Context) ([]*models.RSVP, error)
}
```

### 2.2 Implement DynamoDB Repository
Create `internal/db/dynamodb/guest_repository.go`:

```go
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

type GuestRepository struct {
    client    *dynamodb.Client
    tableName string
}

func NewGuestRepository(client *dynamodb.Client, tableName string) *GuestRepository {
    return &GuestRepository{
        client:    client,
        tableName: tableName,
    }
}

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
        return nil, fmt.Errorf("failed to query guest: %w", err)
    }
    
    if len(result.Items) == 0 {
        return nil, fmt.Errorf("guest not found")
    }
    
    var guest models.Guest
    err = attributevalue.UnmarshalMap(result.Items[0], &guest)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal guest: %w", err)
    }
    
    return &guest, nil
}

func (r *GuestRepository) CreateGuest(ctx context.Context, guest *models.Guest) error {
    if guest.GuestID == "" {
        guest.GuestID = uuid.New().String()
    }
    
    guest.CreatedAt = time.Now()
    guest.UpdatedAt = time.Now()
    
    item, err := attributevalue.MarshalMap(guest)
    if err != nil {
        return fmt.Errorf("failed to marshal guest: %w", err)
    }
    
    _, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
        TableName: aws.String(r.tableName),
        Item:      item,
    })
    
    if err != nil {
        return fmt.Errorf("failed to create guest: %w", err)
    }
    
    return nil
}

func (r *GuestRepository) ListGuests(ctx context.Context) ([]*models.Guest, error) {
    result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
        TableName: aws.String(r.tableName),
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to scan guests: %w", err)
    }
    
    var guests []*models.Guest
    err = attributevalue.UnmarshalListOfMaps(result.Items, &guests)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal guests: %w", err)
    }
    
    return guests, nil
}
```

## Step 3: QR Code Generation

### 3.1 Install QR Code Library
Add to `go.mod`:
```bash
go get github.com/skip2/go-qrcode
```

### 3.2 Create QR Code Generator
Create `internal/qrcode/generator.go`:

```go
package qrcode

import (
    "fmt"
    "image"
    "image/color"
    "image/draw"
    "image/png"
    "io"
    
    qr "github.com/skip2/go-qrcode"
)

type Generator struct {
    baseURL string
}

func NewGenerator(baseURL string) *Generator {
    return &Generator{baseURL: baseURL}
}

// GenerateInvitationQR creates a QR code for a specific invitation
func (g *Generator) GenerateInvitationQR(invitationCode string) ([]byte, error) {
    // Create URL with invitation code
    rsvpURL := fmt.Sprintf("%s/rsvp?code=%s", g.baseURL, invitationCode)
    
    // Generate QR code with high error correction
    qrCode, err := qr.New(rsvpURL, qr.High)
    if err != nil {
        return nil, fmt.Errorf("failed to create QR code: %w", err)
    }
    
    // Set custom colors (optional)
    qrCode.BackgroundColor = color.White
    qrCode.ForegroundColor = color.Black
    
    // Generate PNG image
    png, err := qrCode.PNG(512)
    if err != nil {
        return nil, fmt.Errorf("failed to generate PNG: %w", err)
    }
    
    return png, nil
}

// GenerateStyledQR creates a QR code with custom styling
func (g *Generator) GenerateStyledQR(invitationCode string, style QRStyle) ([]byte, error) {
    rsvpURL := fmt.Sprintf("%s/rsvp?code=%s", g.baseURL, invitationCode)
    
    qrCode, err := qr.New(rsvpURL, qr.High)
    if err != nil {
        return nil, err
    }
    
    // Apply custom styling
    qrCode.BackgroundColor = style.BackgroundColor
    qrCode.ForegroundColor = style.ForegroundColor
    
    return qrCode.PNG(style.Size)
}

type QRStyle struct {
    Size            int
    BackgroundColor color.Color
    ForegroundColor color.Color
}
```

## Step 4: Guest Import Tool

### 4.1 Create CSV Import Tool
Create `cmd/import-guests/main.go`:

```go
package main

import (
    "context"
    "encoding/csv"
    "flag"
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    
    "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
    "github.com/apkiernan/thedrewzers/internal/models"
)

func main() {
    csvFile := flag.String("file", "", "CSV file to import")
    tableName := flag.String("table", os.Getenv("GUESTS_TABLE"), "DynamoDB table name")
    flag.Parse()
    
    if *csvFile == "" {
        log.Fatal("Please provide a CSV file")
    }
    
    // Open CSV file
    file, err := os.Open(*csvFile)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    
    // Parse CSV
    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        log.Fatal(err)
    }
    
    // Setup DynamoDB client
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }
    
    client := dynamodb.NewFromConfig(cfg)
    repo := dynamodb.NewGuestRepository(client, *tableName)
    
    // Skip header row
    for i, record := range records[1:] {
        if len(record) < 5 {
            log.Printf("Skipping row %d: insufficient columns", i+2)
            continue
        }
        
        maxPartySize, _ := strconv.Atoi(record[3])
        
        guest := &models.Guest{
            InvitationCode:   generateInvitationCode(),
            PrimaryGuest:     record[0],
            HouseholdMembers: parseHouseholdMembers(record[1]),
            Email:           record[2],
            MaxPartySize:    maxPartySize,
            Address: models.Address{
                Street:  record[4],
                City:    record[5],
                State:   record[6],
                Zip:     record[7],
                Country: "USA",
            },
        }
        
        err := repo.CreateGuest(context.TODO(), guest)
        if err != nil {
            log.Printf("Failed to create guest %s: %v", guest.PrimaryGuest, err)
            continue
        }
        
        fmt.Printf("Created guest: %s (Code: %s)\n", guest.PrimaryGuest, guest.InvitationCode)
    }
}

func generateInvitationCode() string {
    // Generate 8-character alphanumeric code
    const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    b := make([]byte, 8)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

func parseHouseholdMembers(members string) []string {
    if members == "" {
        return []string{}
    }
    return strings.Split(members, ";")
}
```

### 4.2 Create CSV Template
Create `templates/guests-import-template.csv`:

```csv
primary_guest,household_members,email,max_party_size,street,city,state,zip
"John & Jane Smith","","john@email.com",2,"123 Main St","Boston","MA","02101"
"Bob Johnson","Mary Johnson;Billy Johnson","bob@email.com",3,"456 Oak Ave","Cambridge","MA","02139"
```

## Step 5: QR Code Batch Generation

### 5.1 Create Batch Generator
Create `cmd/generate-qr-codes/main.go`:

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    
    dbRepo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
    "github.com/apkiernan/thedrewzers/internal/qrcode"
)

func main() {
    outputDir := flag.String("output", "./qr-codes", "Output directory for QR codes")
    baseURL := flag.String("url", "https://thedrewzers.com", "Base URL for RSVP links")
    tableName := flag.String("table", os.Getenv("GUESTS_TABLE"), "DynamoDB table name")
    flag.Parse()
    
    // Create output directory
    err := os.MkdirAll(*outputDir, 0755)
    if err != nil {
        log.Fatal(err)
    }
    
    // Setup DynamoDB client
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }
    
    client := dynamodb.NewFromConfig(cfg)
    repo := dbRepo.NewGuestRepository(client, *tableName)
    
    // Get all guests
    guests, err := repo.ListGuests(context.TODO())
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate QR codes
    generator := qrcode.NewGenerator(*baseURL)
    
    for _, guest := range guests {
        png, err := generator.GenerateInvitationQR(guest.InvitationCode)
        if err != nil {
            log.Printf("Failed to generate QR for %s: %v", guest.PrimaryGuest, err)
            continue
        }
        
        // Save QR code
        filename := fmt.Sprintf("%s_%s.png", 
            sanitizeFilename(guest.PrimaryGuest), 
            guest.InvitationCode)
        
        path := filepath.Join(*outputDir, filename)
        err = os.WriteFile(path, png, 0644)
        if err != nil {
            log.Printf("Failed to save QR for %s: %v", guest.PrimaryGuest, err)
            continue
        }
        
        fmt.Printf("Generated QR code for %s: %s\n", guest.PrimaryGuest, path)
    }
    
    // Generate guest list CSV
    generateGuestList(*outputDir, guests)
}

func sanitizeFilename(name string) string {
    // Replace problematic characters
    replacer := strings.NewReplacer(
        " ", "_",
        "&", "and",
        "/", "-",
        "\\", "-",
        ":", "-",
        "*", "-",
        "?", "-",
        "\"", "-",
        "<", "-",
        ">", "-",
        "|", "-",
    )
    return replacer.Replace(name)
}

func generateGuestList(outputDir string, guests []*models.Guest) {
    file, err := os.Create(filepath.Join(outputDir, "guest-list.csv"))
    if err != nil {
        log.Printf("Failed to create guest list: %v", err)
        return
    }
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // Write header
    writer.Write([]string{
        "Primary Guest",
        "Invitation Code", 
        "QR Filename",
        "Max Party Size",
        "Email",
    })
    
    // Write guest data
    for _, guest := range guests {
        filename := fmt.Sprintf("%s_%s.png",
            sanitizeFilename(guest.PrimaryGuest),
            guest.InvitationCode)
            
        writer.Write([]string{
            guest.PrimaryGuest,
            guest.InvitationCode,
            filename,
            strconv.Itoa(guest.MaxPartySize),
            guest.Email,
        })
    }
}
```

## Step 6: Testing

### 6.1 Test Guest Import
```bash
# Create test CSV
cat > test-guests.csv << EOF
primary_guest,household_members,email,max_party_size,street,city,state,zip
"Test Guest","","test@email.com",2,"123 Test St","Boston","MA","02101"
EOF

# Import guests
go run cmd/import-guests/main.go -file test-guests.csv
```

### 6.2 Generate QR Codes
```bash
# Generate QR codes for all guests
go run cmd/generate-qr-codes/main.go -output ./qr-output

# Check generated files
ls -la ./qr-output/
```

## Next Steps
- Phase 3: Implement public RSVP system
- Design QR code placement on invitations
- Test QR code scanning on various devices

## Useful Commands

### Query Guest by Code
```bash
aws dynamodb query \
  --table-name wedding-guests \
  --index-name invitation_code_index \
  --key-condition-expression "invitation_code = :code" \
  --expression-attribute-values '{":code":{"S":"ABC12345"}}'
```

### List All Guests
```bash
aws dynamodb scan --table-name wedding-guests
```