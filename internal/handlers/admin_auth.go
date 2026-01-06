package handlers

import (
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/apkiernan/thedrewzers/internal/auth"
	"github.com/apkiernan/thedrewzers/internal/db"
	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/views"
)

// adminEmailWhitelist contains the only emails allowed to log in as admins.
// This is checked BEFORE database lookup for additional security.
var adminEmailWhitelist = map[string]bool{
	"apkiernan@gmail.com": true,
	// Add other allowed admin emails here
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

// isEmailWhitelisted checks if an email is allowed to log in
func isEmailWhitelisted(email string) bool {
	return adminEmailWhitelist[strings.ToLower(email)]
}

// AdminAuthHandler handles admin authentication requests
type AdminAuthHandler struct {
	adminRepo  db.AdminRepository
	jwtService *auth.JWTService
}

// NewAdminAuthHandler creates a new AdminAuthHandler
func NewAdminAuthHandler(adminRepo db.AdminRepository, jwtService *auth.JWTService) *AdminAuthHandler {
	return &AdminAuthHandler{
		adminRepo:  adminRepo,
		jwtService: jwtService,
	}
}

// HandleLoginPage displays the admin login form
func (h *AdminAuthHandler) HandleLoginPage(w http.ResponseWriter, r *http.Request) {
	// If already authenticated, redirect to dashboard
	cookie, err := r.Cookie("admin_token")
	if err == nil && cookie.Value != "" {
		if _, err := h.jwtService.ValidateToken(cookie.Value); err == nil {
			http.Redirect(w, r, "/dashboard", http.StatusFound)
			return
		}
	}

	views.AdminLoginPage("").Render(r.Context(), w)
}

// HandleLoginSubmit processes the login form submission
func (h *AdminAuthHandler) HandleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		logger.Error("failed to parse login form", "error", err)
		views.AdminLoginPage("Invalid request").Render(r.Context(), w)
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")

	// Validate input
	if email == "" || password == "" {
		views.AdminLoginPage("Email and password are required").Render(r.Context(), w)
		return
	}

	// Check email whitelist BEFORE database lookup
	if !isEmailWhitelisted(email) {
		logger.Warn("admin login rejected", "reason", "email not whitelisted", "email", email)
		views.AdminLoginPage("Invalid email or password").Render(r.Context(), w)
		return
	}

	// Look up admin user
	admin, err := h.adminRepo.GetAdminByEmail(r.Context(), email)
	if err != nil {
		logger.Warn("admin login failed", "reason", "user not found", "email", email)
		views.AdminLoginPage("Invalid email or password").Render(r.Context(), w)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		logger.Warn("admin login failed", "reason", "invalid password", "email", email)
		views.AdminLoginPage("Invalid email or password").Render(r.Context(), w)
		return
	}

	// Generate JWT token
	token, err := h.jwtService.GenerateToken(admin.Email, admin.Name, admin.Role)
	if err != nil {
		logger.Error("failed to generate token", "email", email, "error", err)
		views.AdminLoginPage("Authentication failed").Render(r.Context(), w)
		return
	}

	// Set auth cookie
	auth.SetAuthCookie(w, r, token)

	// Update last login
	if err := h.adminRepo.UpdateLastLogin(r.Context(), email); err != nil {
		logger.Warn("failed to update last login", "email", email, "error", err)
		// Non-fatal error, continue
	}

	logger.Info("admin login successful", "email", email)

	// Redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// HandleLogout logs out the admin user
func (h *AdminAuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	auth.ClearAuthCookie(w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// HandleDashboardPlaceholder provides a temporary dashboard placeholder
func (h *AdminAuthHandler) HandleDashboardPlaceholder(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Admin Dashboard</title>
	<link href="/static/css/tailwind.css" rel="stylesheet"/>
</head>
<body class="bg-gray-100 min-h-screen p-8">
	<div class="max-w-4xl mx-auto">
		<div class="bg-white rounded-lg shadow p-6">
			<h1 class="text-2xl font-bold text-gray-900 mb-4">Admin Dashboard</h1>
			<p class="text-gray-600 mb-4">Welcome, ` + claims.Name + `!</p>
			<p class="text-sm text-gray-500 mb-6">Role: ` + claims.Role + `</p>
			<a href="/logout" class="text-blue-600 hover:underline">Logout</a>
		</div>
		<p class="text-center text-sm text-gray-400 mt-4">
			Dashboard UI coming in Phase 5
		</p>
	</div>
</body>
</html>`))
}
