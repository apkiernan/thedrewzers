package main

import (
	"context"
	"crypto/rand"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	dbdynamo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
	"github.com/apkiernan/thedrewzers/internal/models"
)

func main() {
	csvFile := flag.String("file", "", "CSV file to import (required)")
	tableName := flag.String("table", os.Getenv("GUESTS_TABLE"), "DynamoDB table name")
	dryRun := flag.Bool("dry-run", false, "Validate CSV without importing to database")
	flag.Parse()

	if *csvFile == "" {
		fmt.Println("Usage: import-guests -file <csv-file> [-table <table-name>] [-dry-run]")
		fmt.Println("\nRequired:")
		fmt.Println("  -file    CSV file with guest data to import")
		fmt.Println("\nOptional:")
		fmt.Println("  -table   DynamoDB table name (default: GUESTS_TABLE env var or thedrewzers-wedding-guests)")
		fmt.Println("  -dry-run Validate CSV and show what would be imported without writing to database")
		os.Exit(1)
	}

	if *tableName == "" {
		*tableName = "thedrewzers-wedding-guests"
	}

	ctx := context.Background()

	// Parse CSV file
	guests, err := parseCSV(*csvFile)
	if err != nil {
		log.Fatalf("Failed to parse CSV: %v", err)
	}

	fmt.Printf("Parsed %d guests from CSV\n", len(guests))

	if *dryRun {
		fmt.Println("\nDry run - guests that would be imported:")
		fmt.Println(strings.Repeat("-", 80))
		for _, g := range guests {
			fmt.Printf("  Name: %s\n", g.PrimaryGuest)
			fmt.Printf("  Code: %s\n", g.InvitationCode)
			fmt.Printf("  Party Size: %d\n", g.MaxPartySize)
			if len(g.HouseholdMembers) > 0 {
				fmt.Printf("  Household: %s\n", strings.Join(g.HouseholdMembers, ", "))
			}
			if g.Email != "" {
				fmt.Printf("  Email: %s\n", g.Email)
			}
			fmt.Println()
		}
		fmt.Printf("Total: %d guests would be imported\n", len(guests))
		return
	}

	// Setup DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	repo := dbdynamo.NewGuestRepository(client, *tableName)

	// Import guests
	var success, failed int
	for _, guest := range guests {
		if err := repo.CreateGuest(ctx, guest); err != nil {
			log.Printf("Failed to create guest %s: %v", guest.PrimaryGuest, err)
			failed++
			continue
		}
		fmt.Printf("Created: %s (Code: %s)\n", guest.PrimaryGuest, guest.InvitationCode)
		success++
	}

	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("Import complete: %d succeeded, %d failed\n", success, failed)
}

func parseCSV(path string) ([]*models.Guest, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV must have header row and at least one data row")
	}

	// Validate header
	header := records[0]
	expectedCols := []string{"primary_guest", "household_members", "email", "max_party_size", "street", "city", "state", "zip"}
	if len(header) < len(expectedCols) {
		return nil, fmt.Errorf("CSV must have at least %d columns: %s", len(expectedCols), strings.Join(expectedCols, ", "))
	}

	var guests []*models.Guest
	for i, record := range records[1:] { // Skip header
		rowNum := i + 2 // 1-indexed, accounting for header

		if len(record) < 8 {
			log.Printf("Warning: Skipping row %d - only %d columns (need 8)", rowNum, len(record))
			continue
		}

		// Skip empty rows
		if strings.TrimSpace(record[0]) == "" {
			continue
		}

		maxPartySize, err := strconv.Atoi(strings.TrimSpace(record[3]))
		if err != nil || maxPartySize < 1 {
			maxPartySize = 1
		}

		code, err := generateInvitationCode()
		if err != nil {
			return nil, fmt.Errorf("generating invitation code for row %d: %w", rowNum, err)
		}

		guest := &models.Guest{
			InvitationCode:   code,
			PrimaryGuest:     strings.TrimSpace(record[0]),
			HouseholdMembers: parseHouseholdMembers(record[1]),
			Email:            strings.TrimSpace(record[2]),
			MaxPartySize:     maxPartySize,
			Address: models.Address{
				Street:  strings.TrimSpace(record[4]),
				City:    strings.TrimSpace(record[5]),
				State:   strings.TrimSpace(record[6]),
				Zip:     strings.TrimSpace(record[7]),
				Country: "USA",
			},
		}

		guests = append(guests, guest)
	}

	if len(guests) == 0 {
		return nil, fmt.Errorf("no valid guest rows found in CSV")
	}

	return guests, nil
}

// generateInvitationCode creates a cryptographically secure 8-character code
// Uses charset without ambiguous characters (no I/O/0/1) for better readability
func generateInvitationCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}

func parseHouseholdMembers(members string) []string {
	members = strings.TrimSpace(members)
	if members == "" {
		return []string{}
	}

	parts := strings.Split(members, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
