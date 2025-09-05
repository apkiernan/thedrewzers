# Phase 4: Admin Authentication System

## Overview
This phase implements the admin authentication system with password-based login, JWT tokens, and secure session management on the admin subdomain.

## Prerequisites
- Phase 1-3 completed
- Admin subdomain configured (admin.thedrewzers.com)
- JWT secret configured in environment

## Step 1: Admin User Model and Repository

### 1.1 Create Admin User Model
Create `internal/models/admin.go`:

```go
package models

import (
    "time"
)

type AdminUser struct {
    Email        string    `json:"email" dynamodbav:"email"`
    PasswordHash string    `json:"-" dynamodbav:"password_hash"`
    Role         string    `json:"role" dynamodbav:"role"`
    Name         string    `json:"name" dynamodbav:"name"`
    CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" dynamodbav:"updated_at"`
    LastLogin    time.Time `json:"last_login" dynamodbav:"last_login"`
}

type AdminRole string

const (
    RoleAdmin  AdminRole = "admin"
    RoleViewer AdminRole = "viewer"
)

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

type LoginResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
    Token   string `json:"token,omitempty"`
}
```

### 1.2 Create Admin Repository
Create `internal/db/dynamodb/admin_repository.go`:

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
    
    "github.com/apkiernan/thedrewzers/internal/models"
)

type AdminRepository struct {
    client    *dynamodb.Client
    tableName string
}

func NewAdminRepository(client *dynamodb.Client, tableName string) *AdminRepository {
    return &AdminRepository{
        client:    client,
        tableName: tableName,
    }
}

func (r *AdminRepository) GetAdminUser(ctx context.Context, email string) (*models.AdminUser, error) {
    result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "email": &types.AttributeValueMemberS{Value: email},
        },
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to get admin user: %w", err)
    }
    
    if result.Item == nil {
        return nil, fmt.Errorf("admin user not found")
    }
    
    var admin models.AdminUser
    err = attributevalue.UnmarshalMap(result.Item, &admin)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal admin user: %w", err)
    }
    
    return &admin, nil
}

func (r *AdminRepository) UpdateLastLogin(ctx context.Context, email string) error {
    _, err := r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "email": &types.AttributeValueMemberS{Value: email},
        },
        UpdateExpression: aws.String("SET last_login = :now"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":now": &types.AttributeValueMemberS{
                Value: time.Now().Format(time.RFC3339),
            },
        },
    })
    
    return err
}

func (r *AdminRepository) CreateAdminUser(ctx context.Context, admin *models.AdminUser) error {
    admin.CreatedAt = time.Now()
    admin.UpdatedAt = time.Now()
    
    item, err := attributevalue.MarshalMap(admin)
    if err != nil {
        return fmt.Errorf("failed to marshal admin user: %w", err)
    }
    
    _, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
        TableName: aws.String(r.tableName),
        Item:      item,
    })
    
    return err
}
```

## Step 2: JWT Authentication

### 2.1 Create JWT Service
Create `internal/auth/jwt.go`:

```go
package auth

import (
    "fmt"
    "os"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    Email string `json:"email"`
    Name  string `json:"name"`
    Role  string `json:"role"`
    jwt.RegisteredClaims
}

type JWTService struct {
    secret []byte
}

func NewJWTService() *JWTService {
    secret := os.Getenv("JWT_SECRET")
    if secret == "" {
        panic("JWT_SECRET not set")
    }
    
    return &JWTService{
        secret: []byte(secret),
    }
}

func (s *JWTService) GenerateToken(email, name, role string) (string, error) {
    claims := &Claims{
        Email: email,
        Name:  name,
        Role:  role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secret)
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.secret, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid token")
}
```

### 2.2 Create Authentication Middleware
Create `internal/middleware/auth.go`:

```go
package middleware

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/apkiernan/thedrewzers/internal/auth"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func RequireAdmin(jwtService *auth.JWTService) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            // Check for JWT in cookie
            cookie, err := r.Cookie("admin_token")
            if err != nil {
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            }
            
            // Validate JWT
            claims, err := jwtService.ValidateToken(cookie.Value)
            if err != nil {
                // Clear invalid cookie
                http.SetCookie(w, &http.Cookie{
                    Name:     "admin_token",
                    Value:    "",
                    Path:     "/",
                    MaxAge:   -1,
                    HttpOnly: true,
                    Secure:   true,
                })
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            }
            
            // Check role
            if claims.Role != "admin" && claims.Role != "viewer" {
                http.Error(w, "Unauthorized", http.StatusForbidden)
                return
            }
            
            // Add claims to context
            ctx := context.WithValue(r.Context(), ClaimsKey, claims)
            next(w, r.WithContext(ctx))
        }
    }
}

func RequireRole(role string) func(http.HandlerFunc) http.HandlerFunc {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            claims, ok := r.Context().Value(ClaimsKey).(*auth.Claims)
            if !ok || claims.Role != role {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }
            next(w, r)
        }
    }
}

func AdminSecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "same-origin")
        w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        next.ServeHTTP(w, r)
    })
}
```

## Step 3: Admin Authentication Handlers

### 3.1 Create Login Handler
Create `internal/handlers/admin_auth.go`:

```go
package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "time"
    
    "golang.org/x/crypto/bcrypt"
    
    "github.com/apkiernan/thedrewzers/internal/auth"
    "github.com/apkiernan/thedrewzers/internal/db"
    "github.com/apkiernan/thedrewzers/internal/models"
    "github.com/apkiernan/thedrewzers/internal/views"
)

type AdminAuthHandler struct {
    adminRepo  db.AdminRepository
    jwtService *auth.JWTService
}

func NewAdminAuthHandler(adminRepo db.AdminRepository, jwtService *auth.JWTService) *AdminAuthHandler {
    return &AdminAuthHandler{
        adminRepo:  adminRepo,
        jwtService: jwtService,
    }
}

func (h *AdminAuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Check if already logged in
        if cookie, err := r.Cookie("admin_token"); err == nil {
            if _, err := h.jwtService.ValidateToken(cookie.Value); err == nil {
                http.Redirect(w, r, "/dashboard", http.StatusFound)
                return
            }
        }
        
        // Show login form
        views.AdminLogin().Render(r.Context(), w)
        return
    }
    
    // POST - process login
    var req models.LoginRequest
    
    // Check content type
    if r.Header.Get("Content-Type") == "application/json" {
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }
    } else {
        // Form data
        req.Email = r.FormValue("email")
        req.Password = r.FormValue("password")
    }
    
    // Validate input
    if req.Email == "" || req.Password == "" {
        views.AdminLogin("Please enter email and password").Render(r.Context(), w)
        return
    }
    
    // Get admin user
    admin, err := h.adminRepo.GetAdminUser(r.Context(), req.Email)
    if err != nil {
        log.Printf("Failed to get admin user %s: %v", req.Email, err)
        views.AdminLogin("Invalid email or password").Render(r.Context(), w)
        return
    }
    
    // Check password
    err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password))
    if err != nil {
        log.Printf("Invalid password for %s", req.Email)
        views.AdminLogin("Invalid email or password").Render(r.Context(), w)
        return
    }
    
    // Generate JWT token
    token, err := h.jwtService.GenerateToken(admin.Email, admin.Name, admin.Role)
    if err != nil {
        log.Printf("Failed to generate token: %v", err)
        http.Error(w, "Authentication failed", http.StatusInternalServerError)
        return
    }
    
    // Set secure cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "admin_token",
        Value:    token,
        Path:     "/",
        Domain:   ".thedrewzers.com",
        Expires:  time.Now().Add(24 * time.Hour),
        MaxAge:   86400, // 24 hours
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
    
    // Update last login
    h.adminRepo.UpdateLastLogin(r.Context(), admin.Email)
    
    // Return response based on content type
    if r.Header.Get("Content-Type") == "application/json" {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(models.LoginResponse{
            Success: true,
            Message: "Login successful",
        })
    } else {
        http.Redirect(w, r, "/dashboard", http.StatusFound)
    }
}

func (h *AdminAuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
    // Clear cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "admin_token",
        Value:    "",
        Path:     "/",
        Domain:   ".thedrewzers.com",
        MaxAge:   -1,
        HttpOnly: true,
        Secure:   true,
    })
    
    http.Redirect(w, r, "/login", http.StatusFound)
}
```

## Step 4: Admin Login Views

### 4.1 Create Admin Login Template
Create `internal/views/admin_login.templ`:

```templ
package views

templ AdminLogin(error ...string) {
    <!DOCTYPE html>
    <html lang="en">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>Admin Login - Wedding RSVP</title>
            <link href="/static/css/tailwind.css" rel="stylesheet"/>
            <style>
                .admin-gradient {
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                }
            </style>
        </head>
        <body class="bg-gray-100">
            <div class="min-h-screen flex items-center justify-center px-4">
                <div class="max-w-md w-full space-y-8">
                    <div class="bg-white shadow-2xl rounded-lg px-8 py-10">
                        <div class="text-center mb-8">
                            <h2 class="text-3xl font-bold text-gray-900">Admin Login</h2>
                            <p class="mt-2 text-sm text-gray-600">Wedding RSVP Management</p>
                        </div>
                        
                        <form class="space-y-6" action="/login" method="POST">
                            if len(error) > 0 && error[0] != "" {
                                <div class="rounded-md bg-red-50 border border-red-200 p-4">
                                    <p class="text-sm text-red-800">{ error[0] }</p>
                                </div>
                            }
                            
                            <div>
                                <label for="email" class="block text-sm font-medium text-gray-700">
                                    Email address
                                </label>
                                <input 
                                    id="email" 
                                    name="email" 
                                    type="email" 
                                    autocomplete="email"
                                    required 
                                    class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                    placeholder="admin@example.com"
                                />
                            </div>
                            
                            <div>
                                <label for="password" class="block text-sm font-medium text-gray-700">
                                    Password
                                </label>
                                <input 
                                    id="password" 
                                    name="password" 
                                    type="password" 
                                    autocomplete="current-password"
                                    required 
                                    class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                                    placeholder="••••••••"
                                />
                            </div>
                            
                            <div>
                                <button 
                                    type="submit" 
                                    class="group relative w-full flex justify-center py-3 px-4 border border-transparent text-sm font-medium rounded-md text-white admin-gradient hover:opacity-90 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 transition-opacity"
                                >
                                    Sign in
                                </button>
                            </div>
                        </form>
                        
                        <div class="mt-6 text-center">
                            <a href="/" class="text-sm text-gray-600 hover:text-gray-900">
                                ← Back to main site
                            </a>
                        </div>
                    </div>
                </div>
            </div>
        </body>
    </html>
}
```

## Step 5: Create Admin User Tool

### 5.1 Create Admin User CLI
Create `cmd/create-admin/main.go`:

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "syscall"
    
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "golang.org/x/crypto/bcrypt"
    "golang.org/x/term"
    
    dbRepo "github.com/apkiernan/thedrewzers/internal/db/dynamodb"
    "github.com/apkiernan/thedrewzers/internal/models"
)

func main() {
    email := flag.String("email", "", "Admin email address")
    name := flag.String("name", "", "Admin name")
    role := flag.String("role", "admin", "Admin role (admin or viewer)")
    tableName := flag.String("table", os.Getenv("ADMINS_TABLE"), "DynamoDB table name")
    flag.Parse()
    
    if *email == "" || *name == "" {
        log.Fatal("Please provide email and name")
    }
    
    // Validate role
    if *role != "admin" && *role != "viewer" {
        log.Fatal("Role must be 'admin' or 'viewer'")
    }
    
    // Get password securely
    fmt.Print("Enter password: ")
    password, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println()
    
    fmt.Print("Confirm password: ")
    confirmPassword, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println()
    
    if string(password) != string(confirmPassword) {
        log.Fatal("Passwords do not match")
    }
    
    if len(password) < 8 {
        log.Fatal("Password must be at least 8 characters")
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        log.Fatal("Failed to hash password:", err)
    }
    
    // Setup DynamoDB
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        log.Fatal(err)
    }
    
    client := dynamodb.NewFromConfig(cfg)
    repo := dbRepo.NewAdminRepository(client, *tableName)
    
    // Create admin user
    admin := &models.AdminUser{
        Email:        *email,
        PasswordHash: string(hashedPassword),
        Name:         *name,
        Role:         *role,
    }
    
    err = repo.CreateAdminUser(context.TODO(), admin)
    if err != nil {
        log.Fatal("Failed to create admin user:", err)
    }
    
    fmt.Printf("Admin user created successfully:\n")
    fmt.Printf("  Email: %s\n", admin.Email)
    fmt.Printf("  Name: %s\n", admin.Name)
    fmt.Printf("  Role: %s\n", admin.Role)
}
```

### 5.2 Create First Admin User
```bash
# Build the tool
go build -o create-admin cmd/create-admin/main.go

# Create admin user
./create-admin -email="apkiernan@gmail.com" -name="Andrew Kiernan" -role="admin"
# Enter password when prompted

# Create viewer user (optional)
./create-admin -email="viewer@example.com" -name="Viewer User" -role="viewer"
```

## Step 6: Admin Route Setup

### 6.1 Update Lambda for Subdomain Routing
Update `cmd/lambda/main.go`:

```go
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // Check if this is an admin subdomain request
    host := request.Headers["Host"]
    if host == "" {
        host = request.Headers["host"]
    }
    
    if strings.HasPrefix(host, "admin.") {
        return adminApp.ProxyWithContext(ctx, request)
    }
    
    return publicApp.ProxyWithContext(ctx, request)
}
```

### 6.2 Setup Admin Routes
Create `internal/routes/admin_routes.go`:

```go
package routes

import (
    "net/http"
    
    "github.com/apkiernan/thedrewzers/internal/auth"
    "github.com/apkiernan/thedrewzers/internal/handlers"
    "github.com/apkiernan/thedrewzers/internal/middleware"
)

func SetupAdminRoutes(
    mux *http.ServeMux,
    authHandler *handlers.AdminAuthHandler,
    adminHandler *handlers.AdminDashboardHandler,
    jwtService *auth.JWTService,
) {
    // Public routes (no auth required)
    mux.HandleFunc("/login", authHandler.HandleLogin)
    mux.HandleFunc("/logout", authHandler.HandleLogout)
    
    // Protected routes
    requireAdmin := middleware.RequireAdmin(jwtService)
    
    // Dashboard
    mux.HandleFunc("/", requireAdmin(func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/dashboard", http.StatusFound)
    }))
    mux.HandleFunc("/dashboard", requireAdmin(adminHandler.HandleDashboard))
    
    // Guest management
    mux.HandleFunc("/guests", requireAdmin(adminHandler.HandleGuests))
    mux.HandleFunc("/guests/add", requireAdmin(adminHandler.HandleAddGuest))
    mux.HandleFunc("/guests/edit", requireAdmin(adminHandler.HandleEditGuest))
    
    // RSVP management
    mux.HandleFunc("/rsvps", requireAdmin(adminHandler.HandleRSVPs))
    mux.HandleFunc("/rsvps/export", requireAdmin(adminHandler.HandleExportRSVPs))
    
    // API endpoints
    mux.HandleFunc("/api/stats", requireAdmin(adminHandler.HandleStats))
    mux.HandleFunc("/api/guests", requireAdmin(adminHandler.HandleGuestsAPI))
    
    // Static files (CSS, JS)
    mux.Handle("/static/", http.StripPrefix("/static/", 
        http.FileServer(http.Dir("./static"))))
}
```

## Step 7: Testing

### 7.1 Test Admin Login Locally
```bash
# Set environment variables
export ADMINS_TABLE=wedding-admins
export JWT_SECRET=$(openssl rand -base64 32)

# Run local admin server
go run cmd/local-admin/main.go

# Visit http://localhost:8081/login
```

### 7.2 Test Authentication Flow
```bash
# Test login endpoint
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"email":"apkiernan@gmail.com","password":"yourpassword"}'

# Test with cookie
curl -b "admin_token=YOUR_JWT_TOKEN" http://localhost:8081/dashboard
```

## Next Steps
- Phase 5: Build admin dashboard UI
- Implement guest and RSVP management
- Add export functionality

## Security Checklist
- [ ] JWT secret is strong and unique
- [ ] Passwords are hashed with bcrypt
- [ ] HTTPS only in production
- [ ] Secure cookie settings
- [ ] CSRF protection
- [ ] Rate limiting on login endpoint