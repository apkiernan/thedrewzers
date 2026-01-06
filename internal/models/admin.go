package models

import (
	"time"
)

// AdminRole represents the role of an admin user
type AdminRole string

const (
	// RoleAdmin has full access to all admin features
	RoleAdmin AdminRole = "admin"
	// RoleViewer can view but not modify data
	RoleViewer AdminRole = "viewer"
)

// AdminUser represents an admin user for the wedding dashboard
type AdminUser struct {
	Email        string    `json:"email" dynamodbav:"email"`
	PasswordHash string    `json:"-" dynamodbav:"password_hash"` // Never expose in JSON
	Role         string    `json:"role" dynamodbav:"role"`
	Name         string    `json:"name" dynamodbav:"name"`
	CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" dynamodbav:"updated_at"`
	LastLogin    time.Time `json:"last_login,omitempty" dynamodbav:"last_login,omitempty"`
}

// LoginRequest represents a login form submission
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
