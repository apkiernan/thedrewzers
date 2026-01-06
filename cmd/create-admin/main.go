package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"

	dbdynamo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
	"github.com/apkiernan/thedrewzers/internal/models"
)

// Whitelisted admin emails - must match the whitelist in handlers/admin_auth.go
var adminEmailWhitelist = map[string]bool{
	"apkiernan@gmail.com":     true,
	"mollysmith128@gmail.com": true,
}

func init() {
	// Allow additional emails via environment variable (comma-separated)
	if extra := os.Getenv("ADMIN_EMAIL_WHITELIST"); extra != "" {
		for _, email := range strings.Split(extra, ",") {
			email = strings.TrimSpace(strings.ToLower(email))
			if email != "" {
				adminEmailWhitelist[email] = true
			}
		}
	}
}

func main() {
	email := flag.String("email", "", "Admin email address (required, must be whitelisted)")
	name := flag.String("name", "", "Admin display name (required)")
	role := flag.String("role", "admin", "Admin role: admin or viewer")
	tableName := flag.String("table", os.Getenv("ADMINS_TABLE"), "DynamoDB table name")
	flag.Parse()

	// Validate required flags
	if *email == "" || *name == "" {
		fmt.Println("Usage: create-admin -email <email> -name <name> [-role <role>] [-table <table>]")
		fmt.Println("\nRequired:")
		fmt.Println("  -email    Admin email address (must be whitelisted)")
		fmt.Println("  -name     Admin display name")
		fmt.Println("\nOptional:")
		fmt.Println("  -role     Admin role: admin (default) or viewer")
		fmt.Println("  -table    DynamoDB table name (default: ADMINS_TABLE env var)")
		fmt.Println("\nWhitelisted emails:")
		for e := range adminEmailWhitelist {
			fmt.Printf("  - %s\n", e)
		}
		os.Exit(1)
	}

	// Validate role
	*role = strings.ToLower(*role)
	if *role != "admin" && *role != "viewer" {
		log.Fatalf("Invalid role: %s (must be 'admin' or 'viewer')", *role)
	}

	// Normalize and validate email against whitelist
	*email = strings.ToLower(strings.TrimSpace(*email))
	if !adminEmailWhitelist[*email] {
		log.Fatalf("Email %s is not whitelisted. Allowed emails:", *email)
		for e := range adminEmailWhitelist {
			fmt.Printf("  - %s\n", e)
		}
		os.Exit(1)
	}

	// Set default table name
	if *tableName == "" {
		*tableName = "thedrewzers-wedding-admins"
	}

	// Get password interactively
	password, err := promptPassword()
	if err != nil {
		log.Fatalf("Failed to read password: %v", err)
	}

	// Validate password length
	if len(password) < 8 {
		log.Fatal("Password must be at least 8 characters")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Setup AWS
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Check for local DynamoDB endpoint
	dynamoClient := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = &endpoint
		}
	})

	// Create repository
	repo := dbdynamo.NewAdminRepository(dynamoClient, *tableName)

	// Check if admin already exists
	existing, err := repo.GetAdminByEmail(ctx, *email)
	if err == nil && existing != nil {
		log.Fatalf("Admin with email %s already exists", *email)
	}

	// Create admin user
	admin := &models.AdminUser{
		Email:        *email,
		PasswordHash: string(hashedPassword),
		Role:         *role,
		Name:         *name,
	}

	if err := repo.CreateAdmin(ctx, admin); err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}

	fmt.Printf("Admin user created successfully!\n")
	fmt.Printf("  Email: %s\n", admin.Email)
	fmt.Printf("  Name:  %s\n", admin.Name)
	fmt.Printf("  Role:  %s\n", admin.Role)
}

func promptPassword() (string, error) {
	fmt.Print("Enter password: ")
	password1, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		// Fallback for non-terminal input
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(line), nil
	}
	fmt.Println()

	fmt.Print("Confirm password: ")
	password2, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()

	if string(password1) != string(password2) {
		return "", fmt.Errorf("passwords do not match")
	}

	return string(password1), nil
}
