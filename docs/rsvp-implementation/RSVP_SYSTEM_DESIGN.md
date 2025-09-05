# Wedding RSVP System with QR Codes - Technical Design Document

## Table of Contents
1. [System Overview](#system-overview)
2. [Technical Architecture](#technical-architecture)
3. [Database Design](#database-design)
4. [API Design](#api-design)
5. [QR Code Implementation](#qr-code-implementation)
6. [Guest Experience Flow](#guest-experience-flow)
7. [Admin Dashboard](#admin-dashboard)
8. [Infrastructure Updates](#infrastructure-updates)
9. [Security Considerations](#security-considerations)
10. [Implementation Timeline](#implementation-timeline)

## System Overview

The RSVP system will provide a seamless, personalized experience for wedding guests using unique QR codes on physical invitations. Each QR code will link directly to a pre-populated RSVP form, eliminating the need for guests to search for their names or enter invitation codes.

### Key Features
- **Unique QR codes** for each invitation/household
- **Mobile-first** responsive design
- **Real-time updates** to guest responses
- **Admin dashboard** for tracking RSVPs
- **Email confirmations** with calendar invites
- **Dietary restrictions** and special requests handling
- **Plus-one management** with name capture

## Technical Architecture

### Component Overview
```
┌─────────────────┐     ┌──────────────┐     ┌─────────────────┐
│  Physical       │     │  CloudFront  │     │  Lambda         │
│  Invitation     │────▶│  CDN         │────▶│  Functions      │
│  (QR Code)      │     │              │     │                 │
└─────────────────┘     └──────────────┘     └────────┬────────┘
                                                       │
                              ┌────────────────────────┼────────────────────────┐
                              │                        │                        │
                        ┌─────▼──────┐         ┌──────▼──────┐        ┌────────▼────────┐
                        │ DynamoDB   │         │ S3 Static   │        │ API Gateway     │
                        │ Tables     │         │ Assets      │        │ REST API        │
                        └────────────┘         └─────────────┘        └─────────────────┘
```

### Technology Stack
- **Backend**: Go 1.23 with AWS Lambda
- **Frontend**: Templ templates + Tailwind CSS
- **Database**: DynamoDB
- **QR Generation**: `github.com/skip2/go-qrcode`
- **Infrastructure**: Terraform + AWS

## Database Design

### DynamoDB Tables

#### 1. Guests Table
```json
{
  "TableName": "wedding-guests",
  "PartitionKey": "guest_id",
  "Attributes": {
    "guest_id": "UUID",
    "invitation_code": "string",
    "primary_guest": "string",
    "household_members": ["string"],
    "max_party_size": "number",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  },
  "GSI": {
    "invitation_code_index": {
      "PartitionKey": "invitation_code"
    }
  }
}
```

#### 2. RSVPs Table
```json
{
  "TableName": "wedding-rsvps",
  "PartitionKey": "rsvp_id",
  "SortKey": "guest_id",
  "Attributes": {
    "rsvp_id": "UUID",
    "guest_id": "UUID",
    "attending": "boolean",
    "party_size": "number",
    "attendee_names": ["string"],
    "dietary_restrictions": ["string"],
    "special_requests": "string",
    "submitted_at": "timestamp",
    "ip_address": "string",
    "user_agent": "string"
  }
}
```

#### 3. Admin Users Table
```json
{
  "TableName": "wedding-admins",
  "PartitionKey": "email",
  "Attributes": {
    "email": "string",
    "password_hash": "string",
    "role": "string",
    "last_login": "timestamp"
  }
}
```

## API Design

### Guest-Facing Endpoints

#### GET /api/rsvp/{invitation_code}
Retrieves guest information and any existing RSVP
```go
type GuestResponse struct {
    GuestID         string   `json:"guest_id"`
    PrimaryGuest    string   `json:"primary_guest"`
    HouseholdMembers []string `json:"household_members"`
    MaxPartySize    int      `json:"max_party_size"`
    ExistingRSVP    *RSVP    `json:"existing_rsvp,omitempty"`
}
```

#### POST /api/rsvp
Submits or updates an RSVP
```go
type RSVPRequest struct {
    GuestID            string   `json:"guest_id"`
    InvitationCode     string   `json:"invitation_code"`
    Attending          bool     `json:"attending"`
    PartySize          int      `json:"party_size"`
    AttendeeNames      []string `json:"attendee_names"`
    DietaryRestrictions []string `json:"dietary_restrictions"`
    SpecialRequests    string   `json:"special_requests"`
}
```

### Admin Endpoints

#### GET /api/admin/guests
Lists all guests with RSVP status

#### POST /api/admin/guests
Adds new guests to the system

#### GET /api/admin/rsvps/export
Exports RSVP data as CSV

#### GET /api/admin/dashboard/stats
Returns RSVP statistics for dashboard

## QR Code Implementation

### Generation Process

```go
package qrcode

import (
    "fmt"
    qr "github.com/skip2/go-qrcode"
)

func GenerateInvitationQR(baseURL, invitationCode string) ([]byte, error) {
    // Create URL with invitation code
    rsvpURL := fmt.Sprintf("%s/rsvp?code=%s", baseURL, invitationCode)
    
    // Generate QR code with high error correction
    qrCode, err := qr.Encode(rsvpURL, qr.High, 256)
    if err != nil {
        return nil, err
    }
    
    return qrCode, nil
}
```

### Batch Generation Script

```go
// cmd/generate-qr/main.go
func main() {
    guests := loadGuestsFromCSV("guests.csv")
    
    for _, guest := range guests {
        qrData, _ := qrcode.GenerateInvitationQR(
            "https://thedrewzers.com",
            guest.InvitationCode,
        )
        
        saveQRCode(qrData, fmt.Sprintf("qr_%s.png", guest.InvitationCode))
    }
}
```

## Guest Experience Flow

### 1. QR Code Scan
- Guest scans QR code on invitation
- Redirected to `https://thedrewzers.com/rsvp?code=ABC123`
- System validates invitation code

### 2. Personalized RSVP Page
```templ
templ RSVPForm(guest GuestInfo) {
    <div class="rsvp-container">
        <h1>Hello, { guest.PrimaryGuest }!</h1>
        <p>We'd love to celebrate with you and your party of up to { guest.MaxPartySize }</p>
        
        <form id="rsvp-form">
            <label>
                <input type="radio" name="attending" value="yes" /> 
                Joyfully Accept
            </label>
            <label>
                <input type="radio" name="attending" value="no" /> 
                Regretfully Decline
            </label>
            
            <div id="attending-details" class="hidden">
                <label>Number attending: 
                    <select name="party_size">
                        @for i := 1; i <= guest.MaxPartySize; i++ {
                            <option value={strconv.Itoa(i)}>{strconv.Itoa(i)}</option>
                        }
                    </select>
                </label>
                
                <div id="attendee-names">
                    <!-- Dynamic name inputs based on party size -->
                </div>
                
                <label>Dietary Restrictions:
                    <textarea name="dietary_restrictions"></textarea>
                </label>
            </div>
            
            <button type="submit">Submit RSVP</button>
        </form>
    </div>
}
```

### 3. Confirmation
- Success page with details
- Email confirmation sent
- Calendar invite attached

## Admin Dashboard

### Admin/Public Separation Strategy

We'll use **subdomain separation** with password-based JWT authentication for maximum security and clear separation:

- **Public Site**: `https://thedrewzers.com`
- **Admin Site**: `https://admin.thedrewzers.com`

This approach provides:
- Complete infrastructure separation
- Independent security policies
- Clear mental model for users
- Ability to use different CloudFront distributions

### Authentication Implementation

#### 1. Password-Based Login System
```go
// internal/handlers/admin_auth.go
package handlers

import (
    "golang.org/x/crypto/bcrypt"
    "github.com/golang-jwt/jwt/v5"
)

func HandleAdminLogin(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        views.AdminLogin().Render(r.Context(), w)
        return
    }
    
    // POST - process login
    email := r.FormValue("email")
    password := r.FormValue("password")
    
    // Verify credentials against DynamoDB
    admin, err := db.GetAdminUser(email)
    if err != nil || admin == nil {
        views.AdminLogin(LoginError{Message: "Invalid credentials"}).Render(r.Context(), w)
        return
    }
    
    // Check password
    err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password))
    if err != nil {
        views.AdminLogin(LoginError{Message: "Invalid credentials"}).Render(r.Context(), w)
        return
    }
    
    // Create JWT token
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
        Email: email,
        Role:  admin.Role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    })
    
    tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
    if err != nil {
        http.Error(w, "Authentication failed", http.StatusInternalServerError)
        return
    }
    
    // Set secure cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "admin_token",
        Value:    tokenString,
        Path:     "/",
        Domain:   ".thedrewzers.com", // Works for all subdomains
        Expires:  time.Now().Add(24 * time.Hour),
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
    
    // Update last login
    db.UpdateLastLogin(email)
    
    http.Redirect(w, r, "/dashboard", http.StatusFound)
}
```

#### 2. Admin Login Page
```templ
// internal/views/admin_login.templ
templ AdminLogin(error ...LoginError) {
    <html lang="en">
        <head>
            <title>Admin Login - Wedding RSVP</title>
            <link href="/static/css/tailwind.css" rel="stylesheet"/>
        </head>
        <body class="bg-gray-100">
            <div class="min-h-screen flex items-center justify-center">
                <div class="max-w-md w-full space-y-8">
                    <div>
                        <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                            Admin Login
                        </h2>
                        <p class="mt-2 text-center text-sm text-gray-600">
                            Wedding RSVP Management System
                        </p>
                    </div>
                    <form class="mt-8 space-y-6" action="/login" method="POST">
                        if len(error) > 0 {
                            <div class="rounded-md bg-red-50 p-4">
                                <p class="text-sm text-red-800">{ error[0].Message }</p>
                            </div>
                        }
                        <div class="rounded-md shadow-sm -space-y-px">
                            <div>
                                <label for="email" class="sr-only">Email address</label>
                                <input id="email" name="email" type="email" required 
                                    class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm" 
                                    placeholder="Email address"/>
                            </div>
                            <div>
                                <label for="password" class="sr-only">Password</label>
                                <input id="password" name="password" type="password" required 
                                    class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm" 
                                    placeholder="Password"/>
                            </div>
                        </div>
                        <div>
                            <button type="submit" class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                                Sign in
                            </button>
                        </div>
                    </form>
                </div>
            </div>
        </body>
    </html>
}
```

#### 3. JWT Authentication Middleware
```go
// internal/middleware/auth.go
package middleware

type Claims struct {
    Email string `json:"email"`
    Role  string `json:"role"`
    jwt.RegisteredClaims
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check for JWT in cookie
        cookie, err := r.Cookie("admin_token")
        if err != nil {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        
        // Validate JWT
        claims := &Claims{}
        token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        
        if err != nil || !token.Valid || claims.Role != "admin" {
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        
        // Add claims to context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next(w, r.WithContext(ctx))
    }
}

func AdminSecurityHeaders(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Referrer-Policy", "same-origin")
        next(w, r)
    }
}
```

#### 4. Route Structure with Subdomain Separation
```go
// cmd/lambda/main.go - Main Lambda handler
func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    // Determine which app to route to based on Host header
    host := request.Headers["Host"]
    
    if strings.HasPrefix(host, "admin.") {
        return adminApp.Handler(ctx, request)
    }
    
    return publicApp.Handler(ctx, request)
}

// internal/routes/public_routes.go
func SetupPublicRoutes(mux *http.ServeMux) {
    mux.HandleFunc("/", handlers.HandleHomePage)
    mux.HandleFunc("/rsvp", handlers.HandleRSVPPage)
    mux.HandleFunc("/api/rsvp/lookup", handlers.HandleRSVPLookup)
    mux.HandleFunc("/api/rsvp/submit", handlers.HandleRSVPSubmit)
}

// internal/routes/admin_routes.go
func SetupAdminRoutes(mux *http.ServeMux) {
    // Public admin routes
    mux.HandleFunc("/login", handlers.HandleAdminLogin)
    mux.HandleFunc("/logout", handlers.HandleAdminLogout)
    
    // All other routes require authentication
    mux.HandleFunc("/", middleware.RequireAdmin(handlers.HandleAdminDashboard))
    mux.HandleFunc("/dashboard", middleware.RequireAdmin(handlers.HandleAdminDashboard))
    mux.HandleFunc("/guests", middleware.RequireAdmin(handlers.HandleGuestManagement))
    mux.HandleFunc("/rsvps", middleware.RequireAdmin(handlers.HandleRSVPManagement))
    mux.HandleFunc("/reports", middleware.RequireAdmin(handlers.HandleReports))
    
    // API endpoints
    mux.HandleFunc("/api/stats", middleware.RequireAdmin(handlers.HandleAdminStats))
    mux.HandleFunc("/api/guests", middleware.RequireAdmin(handlers.HandleGuestsAPI))
    mux.HandleFunc("/api/rsvps/export", middleware.RequireAdmin(handlers.HandleExportRSVPs))
    
    // Apply security headers to all admin routes
    mux = middleware.AdminSecurityHeaders(mux)
}
```

### Admin UI Implementation

#### 1. Admin Layout Template
```templ
// internal/views/admin_layout.templ
package views

templ AdminLayout(title string, content templ.Component) {
    <html lang="en">
        <head>
            <meta charset="UTF-8"/>
            <meta name="viewport" content="width=device-width, initial-scale=1.0"/>
            <title>{ title } - Wedding Admin</title>
            <link href="/static/css/tailwind.css" rel="stylesheet"/>
            <link href="/static/css/admin.css" rel="stylesheet"/>
        </head>
        <body class="bg-gray-100">
            <nav class="bg-white shadow-lg">
                <div class="max-w-7xl mx-auto px-4">
                    <div class="flex justify-between h-16">
                        <div class="flex">
                            <div class="flex-shrink-0 flex items-center">
                                <h1 class="text-xl font-semibold">Wedding Admin</h1>
                            </div>
                            <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
                                <a href="/admin/dashboard" class="nav-link">Dashboard</a>
                                <a href="/admin/guests" class="nav-link">Guests</a>
                                <a href="/admin/rsvps" class="nav-link">RSVPs</a>
                                <a href="/admin/reports" class="nav-link">Reports</a>
                            </div>
                        </div>
                        <div class="flex items-center">
                            <a href="/admin/logout" class="text-gray-500 hover:text-gray-700">Logout</a>
                        </div>
                    </div>
                </div>
            </nav>
            <main class="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
                @content
            </main>
        </body>
    </html>
}
```

#### 2. Admin Dashboard
```templ
templ AdminDashboard(stats DashboardStats) {
    @AdminLayout("Dashboard", AdminDashboardContent(stats))
}

templ AdminDashboardContent(stats DashboardStats) {
    <div class="px-4 py-6 sm:px-0">
        <h2 class="text-2xl font-bold mb-6">RSVP Dashboard</h2>
        
        <!-- Stats Grid -->
        <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="px-4 py-5 sm:p-6">
                    <dt class="text-sm font-medium text-gray-500 truncate">Total Invited</dt>
                    <dd class="mt-1 text-3xl font-semibold text-gray-900">{ strconv.Itoa(stats.TotalInvited) }</dd>
                </div>
            </div>
            
            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="px-4 py-5 sm:p-6">
                    <dt class="text-sm font-medium text-gray-500 truncate">Responses</dt>
                    <dd class="mt-1 text-3xl font-semibold text-gray-900">
                        { strconv.Itoa(stats.TotalResponses) }
                        <span class="text-sm text-gray-500">({ stats.ResponseRate }%)</span>
                    </dd>
                </div>
            </div>
            
            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="px-4 py-5 sm:p-6">
                    <dt class="text-sm font-medium text-gray-500 truncate">Attending</dt>
                    <dd class="mt-1 text-3xl font-semibold text-green-600">{ strconv.Itoa(stats.TotalAttending) }</dd>
                </div>
            </div>
            
            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="px-4 py-5 sm:p-6">
                    <dt class="text-sm font-medium text-gray-500 truncate">Declined</dt>
                    <dd class="mt-1 text-3xl font-semibold text-red-600">{ strconv.Itoa(stats.TotalDeclined) }</dd>
                </div>
            </div>
        </div>
        
        <!-- Recent RSVPs -->
        <div class="mt-8">
            <h3 class="text-lg font-medium mb-4">Recent RSVPs</h3>
            <div class="bg-white shadow overflow-hidden sm:rounded-md">
                <!-- RSVP list here -->
            </div>
        </div>
    </div>
}
```

### Features
1. **Overview Stats**
   - Total invited
   - Responses received
   - Attending count
   - Dietary restriction summary

2. **Guest Management**
   - Search/filter guests
   - Edit guest information
   - Add last-minute invites
   - View RSVP history

3. **Reports**
   - Export guest list
   - Seating chart data
   - Dietary requirements
   - Contact information

## Infrastructure Updates

### Terraform Configuration

```hcl
# DynamoDB Tables
resource "aws_dynamodb_table" "guests" {
  name           = "wedding-guests"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "guest_id"

  attribute {
    name = "guest_id"
    type = "S"
  }

  attribute {
    name = "invitation_code"
    type = "S"
  }

  global_secondary_index {
    name            = "invitation_code_index"
    hash_key        = "invitation_code"
    projection_type = "ALL"
  }
}

resource "aws_dynamodb_table" "rsvps" {
  name           = "wedding-rsvps"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "rsvp_id"
  range_key      = "guest_id"

  attribute {
    name = "rsvp_id"
    type = "S"
  }

  attribute {
    name = "guest_id"
    type = "S"
  }
}

resource "aws_dynamodb_table" "admin_users" {
  name           = "wedding-admins"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "email"

  attribute {
    name = "email"
    type = "S"
  }
}

# Lambda Function for RSVP API
resource "aws_lambda_function" "rsvp_api" {
  filename      = "rsvp-lambda.zip"
  function_name = "wedding-rsvp-api"
  role          = aws_iam_role.lambda_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  
  environment {
    variables = {
      GUESTS_TABLE = aws_dynamodb_table.guests.name
      RSVPS_TABLE  = aws_dynamodb_table.rsvps.name
      ADMINS_TABLE = aws_dynamodb_table.admin_users.name
      JWT_SECRET   = var.jwt_secret
    }
  }
}

# API Gateway Routes - Public
resource "aws_apigatewayv2_route" "get_rsvp" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "GET /api/rsvp/{code}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_route" "post_rsvp" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /api/rsvp"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# API Gateway Routes - Admin
resource "aws_apigatewayv2_route" "admin_login" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /admin/login"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

resource "aws_apigatewayv2_route" "admin_routes" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "ANY /admin/{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}

# WAF Web ACL for CloudFront - Admin Protection
resource "aws_wafv2_web_acl" "admin_protection" {
  provider = aws.us_east_1
  name     = "wedding-admin-protection"
  scope    = "CLOUDFRONT"

  default_action {
    allow {}
  }

  # Rate limiting rule for admin paths
  rule {
    name     = "RateLimitAdminPaths"
    priority = 1

    action {
      block {}
    }

    statement {
      and_statement {
        statement {
          byte_match_statement {
            search_string = "/admin"
            field_to_match {
              uri_path {}
            }
            text_transformation {
              priority = 0
              type     = "LOWERCASE"
            }
            positional_constraint = "STARTS_WITH"
          }
        }
        statement {
          rate_based_statement {
            limit              = 100
            aggregate_key_type = "IP"
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "RateLimitAdminPaths"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "wedding-admin-protection"
    sampled_requests_enabled   = true
  }
}

# Separate CloudFront distribution for admin subdomain
resource "aws_cloudfront_distribution" "admin" {
  enabled = true
  aliases = ["admin.thedrewzers.com"]
  
  origin {
    domain_name = aws_apigatewayv2_api.main.api_endpoint
    origin_id   = aws_apigatewayv2_api.main.id
    
    custom_origin_config {
      http_port              = 80
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }
  
  default_cache_behavior {
    allowed_methods  = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = aws_apigatewayv2_api.main.id
    
    forwarded_values {
      query_string = true
      cookies {
        forward = "all"
      }
      headers = ["*"]
    }
    
    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 0
    max_ttl                = 0
  }
  
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }
  
  viewer_certificate {
    acm_certificate_arn = aws_acm_certificate_validation.admin_cert.certificate_arn
    ssl_support_method  = "sni-only"
  }
  
  web_acl_id = aws_wafv2_web_acl.admin_protection.arn
}

# ACM Certificate for admin subdomain
resource "aws_acm_certificate" "admin" {
  provider          = aws.us_east_1
  domain_name       = "admin.thedrewzers.com"
  validation_method = "DNS"
}

# Route53 record for admin subdomain
resource "aws_route53_record" "admin" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = "admin.thedrewzers.com"
  type    = "A"
  
  alias {
    name                   = aws_cloudfront_distribution.admin.domain_name
    zone_id                = aws_cloudfront_distribution.admin.hosted_zone_id
    evaluate_target_health = false
  }
}

# Update main CloudFront distribution (public site only)
resource "aws_cloudfront_distribution" "main" {
  # ... existing configuration ...
  # Remove admin-specific configurations
}
```

## Security Considerations

### 1. Invitation Code Security
- Use cryptographically secure random codes
- Minimum 8 characters (alphanumeric)
- One-way hashing in database
- Rate limiting on lookups

### 2. RSVP Submission
- CSRF protection
- Rate limiting per IP
- Validation of party size limits
- Sanitization of user inputs

### 3. Admin Access
- Password-based authentication with bcrypt hashing
- JWT tokens stored in secure HTTP-only cookies
- Separate subdomain with dedicated CloudFront distribution
- Role-based permissions (admin, viewer)
- Audit logging of all admin actions
- Session timeout after 24 hours

#### Creating Admin Users
```go
// cmd/create-admin/main.go
package main

import (
    "golang.org/x/crypto/bcrypt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
    email := flag.String("email", "", "Admin email")
    password := flag.String("password", "", "Admin password")
    role := flag.String("role", "admin", "Admin role")
    flag.Parse()
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    
    // Save to DynamoDB
    item := map[string]*dynamodb.AttributeValue{
        "email":         {S: aws.String(*email)},
        "password_hash": {S: aws.String(string(hashedPassword))},
        "role":          {S: aws.String(*role)},
        "created_at":    {S: aws.String(time.Now().Format(time.RFC3339))},
    }
    
    _, err = db.PutItem(&dynamodb.PutItemInput{
        TableName: aws.String("wedding-admins"),
        Item:      item,
    })
    
    fmt.Printf("Admin user created: %s\n", *email)
}
```

### 4. Data Privacy
- PII encryption at rest
- HTTPS everywhere
- GDPR compliance considerations
- Data retention policies

## Implementation Timeline

### Phase 1: Core Infrastructure (Week 1-2)
- [ ] Set up DynamoDB tables
- [ ] Create Lambda functions
- [ ] Implement basic API endpoints
- [ ] Deploy with Terraform

### Phase 2: Guest Experience (Week 3-4)
- [ ] Design RSVP form UI
- [ ] Implement form validation
- [ ] Create confirmation pages
- [ ] Set up email notifications

### Phase 3: QR Code System (Week 5)
- [ ] Implement QR generation
- [ ] Create batch generation tool
- [ ] Test with sample invitations
- [ ] Document printing guidelines

### Phase 4: Admin Dashboard (Week 6-7)
- [ ] Build dashboard UI
- [ ] Implement authentication
- [ ] Create reporting features
- [ ] Add export functionality

### Phase 5: Testing & Launch (Week 8)
- [ ] End-to-end testing
- [ ] Load testing
- [ ] Security audit
- [ ] Soft launch with family

## Additional Considerations

### Backup Plans
- Manual RSVP option via website
- Phone number for assistance
- Paper RSVP cards as fallback

### Analytics
- Track QR code scans
- Monitor form abandonment
- Response time metrics
- Geographic distribution

### Future Enhancements
- SMS reminders
- Seating chart integration
- Photo sharing portal
- Day-of event updates

## Conclusion

This RSVP system leverages your existing AWS infrastructure while adding powerful guest management capabilities. The QR code approach provides a modern, frictionless experience while maintaining the elegance of traditional invitations. The system is designed to scale, with room for future enhancements as needed.