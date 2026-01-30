package db

import (
	"context"

	"github.com/apkiernan/thedrewzers/internal/models"
)

// GuestRepository defines operations for guest data persistence
type GuestRepository interface {
	// GetGuest retrieves a guest by their unique ID
	GetGuest(ctx context.Context, guestID string) (*models.Guest, error)

	// GetGuestByInvitationCode retrieves a guest by their invitation code
	GetGuestByInvitationCode(ctx context.Context, code string) (*models.Guest, error)

	// CreateGuest creates a new guest record
	CreateGuest(ctx context.Context, guest *models.Guest) error

	// UpdateGuest updates an existing guest record
	UpdateGuest(ctx context.Context, guest *models.Guest) error

	// DeleteGuest removes a guest by their ID
	DeleteGuest(ctx context.Context, guestID string) error

	// ListGuests returns all guests
	ListGuests(ctx context.Context) ([]*models.Guest, error)

	// SearchGuestsByName finds guests matching the given name (case-insensitive, partial match on primary_guest or household_members)
	SearchGuestsByName(ctx context.Context, name string) ([]*models.Guest, error)
}

// RSVPRepository defines operations for RSVP data persistence
type RSVPRepository interface {
	// GetRSVP retrieves an RSVP by its unique ID
	GetRSVP(ctx context.Context, rsvpID string) (*models.RSVP, error)

	// GetRSVPByGuestID retrieves an RSVP by the guest's ID
	GetRSVPByGuestID(ctx context.Context, guestID string) (*models.RSVP, error)

	// CreateRSVP creates a new RSVP record
	CreateRSVP(ctx context.Context, rsvp *models.RSVP) error

	// UpdateRSVP updates an existing RSVP record
	UpdateRSVP(ctx context.Context, rsvp *models.RSVP) error

	// ListRSVPs returns all RSVPs
	ListRSVPs(ctx context.Context) ([]*models.RSVP, error)
}

// AdminRepository defines operations for admin user data persistence
type AdminRepository interface {
	// GetAdminByEmail retrieves an admin user by their email address
	GetAdminByEmail(ctx context.Context, email string) (*models.AdminUser, error)

	// CreateAdmin creates a new admin user
	CreateAdmin(ctx context.Context, admin *models.AdminUser) error

	// UpdateAdmin updates an existing admin user
	UpdateAdmin(ctx context.Context, admin *models.AdminUser) error

	// UpdateLastLogin updates the last login timestamp for an admin
	UpdateLastLogin(ctx context.Context, email string) error
}
