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

	dbdynamo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
	"github.com/apkiernan/thedrewzers/internal/invite"
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

	headerIndex := csvHeaderIndex(records[0])
	if _, ok := headerIndex["primary_guest"]; !ok {
		return nil, fmt.Errorf("CSV must include primary_guest column")
	}

	var guests []*models.Guest
	for i, record := range records[1:] { // Skip header
		rowNum := i + 2 // 1-indexed, accounting for header

		primaryGuest := csvCell(record, headerIndex, "primary_guest")
		if primaryGuest == "" {
			continue
		}

		maxPartySize := normalizeMaxPartySize(csvCell(record, headerIndex, "max_party_size"))

		code, err := invite.GenerateCode()
		if err != nil {
			return nil, fmt.Errorf("generating invitation code for row %d: %w", rowNum, err)
		}

		guest := &models.Guest{
			InvitationCode:   code,
			PrimaryGuest:     primaryGuest,
			HouseholdMembers: invite.ParseHouseholdMembers(csvCell(record, headerIndex, "household_members")),
			Email:            csvCell(record, headerIndex, "email"),
			MaxPartySize:     maxPartySize,
			Address: models.Address{
				Street:  csvCell(record, headerIndex, "street"),
				City:    csvCell(record, headerIndex, "city"),
				State:   csvCell(record, headerIndex, "state"),
				Zip:     csvCell(record, headerIndex, "zip"),
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

func csvHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, column := range header {
		normalized := strings.ToLower(strings.TrimSpace(column))
		if normalized != "" {
			index[normalized] = i
		}
	}
	return index
}

func csvCell(row []string, header map[string]int, column string) string {
	idx, ok := header[column]
	if !ok || idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func normalizeMaxPartySize(raw string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 1
	}
	if parsed <= 1 {
		return 1
	}
	return 2
}
