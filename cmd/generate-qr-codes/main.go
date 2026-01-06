package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	dbdynamo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
	"github.com/apkiernan/thedrewzers/internal/models"
	"github.com/apkiernan/thedrewzers/internal/qrcode"
)

func main() {
	outputDir := flag.String("output", "./qr-codes", "Output directory for QR code images")
	baseURL := flag.String("url", "https://thekiernan.wedding", "Base URL for RSVP links")
	tableName := flag.String("table", os.Getenv("GUESTS_TABLE"), "DynamoDB table name")
	styled := flag.Bool("styled", false, "Use wedding-themed styling (dark gray instead of black)")
	flag.Parse()

	if *tableName == "" {
		*tableName = "thedrewzers-wedding-guests"
	}

	ctx := context.Background()

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Setup DynamoDB client
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	repo := dbdynamo.NewGuestRepository(client, *tableName)

	// Get all guests
	guests, err := repo.ListGuests(ctx)
	if err != nil {
		log.Fatalf("Failed to list guests: %v", err)
	}

	if len(guests) == 0 {
		fmt.Println("No guests found in database")
		return
	}

	fmt.Printf("Found %d guests\n", len(guests))
	fmt.Printf("Generating QR codes to: %s\n", *outputDir)
	fmt.Printf("RSVP URL base: %s\n\n", *baseURL)

	// Generate QR codes
	generator := qrcode.NewGenerator(*baseURL)

	var generated int
	for _, guest := range guests {
		var png []byte
		var genErr error

		if *styled {
			png, genErr = generator.GenerateStyledQR(guest.InvitationCode, qrcode.WeddingStyle())
		} else {
			png, genErr = generator.GenerateInvitationQR(guest.InvitationCode)
		}

		if genErr != nil {
			log.Printf("Failed to generate QR for %s: %v", guest.PrimaryGuest, genErr)
			continue
		}

		filename := fmt.Sprintf("%s_%s.png",
			sanitizeFilename(guest.PrimaryGuest),
			guest.InvitationCode)

		path := filepath.Join(*outputDir, filename)
		if err := os.WriteFile(path, png, 0644); err != nil {
			log.Printf("Failed to save QR for %s: %v", guest.PrimaryGuest, err)
			continue
		}

		generated++
		fmt.Printf("Generated: %s\n", filename)
	}

	// Generate guest list CSV
	if err := generateGuestList(*outputDir, guests, *baseURL); err != nil {
		log.Printf("Warning: Failed to generate guest list CSV: %v", err)
	}

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Generated %d QR codes in %s\n", generated, *outputDir)
	fmt.Printf("Guest list CSV: %s/guest-list.csv\n", *outputDir)
}

func sanitizeFilename(name string) string {
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
		"'", "",
		",", "",
	)
	return replacer.Replace(name)
}

func generateGuestList(outputDir string, guests []*models.Guest, baseURL string) error {
	path := filepath.Join(outputDir, "guest-list.csv")
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header row
	if err := writer.Write([]string{
		"Primary Guest",
		"Invitation Code",
		"QR Filename",
		"RSVP URL",
		"Max Party Size",
		"Email",
		"Household Members",
	}); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Data rows
	for _, guest := range guests {
		filename := fmt.Sprintf("%s_%s.png",
			sanitizeFilename(guest.PrimaryGuest),
			guest.InvitationCode)

		rsvpURL := fmt.Sprintf("%s/rsvp?code=%s", baseURL, guest.InvitationCode)

		if err := writer.Write([]string{
			guest.PrimaryGuest,
			guest.InvitationCode,
			filename,
			rsvpURL,
			strconv.Itoa(guest.MaxPartySize),
			guest.Email,
			strings.Join(guest.HouseholdMembers, "; "),
		}); err != nil {
			return fmt.Errorf("writing row for %s: %w", guest.PrimaryGuest, err)
		}
	}

	return nil
}
