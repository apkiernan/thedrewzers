package models

import "errors"

var (
	// ErrGuestNotFound is returned when a guest cannot be found by ID or invitation code
	ErrGuestNotFound = errors.New("guest not found")

	// ErrRSVPNotFound is returned when an RSVP cannot be found
	ErrRSVPNotFound = errors.New("rsvp not found")

	// ErrInvalidCode is returned when an invitation code is invalid or malformed
	ErrInvalidCode = errors.New("invalid invitation code")

	// ErrDuplicateCode is returned when attempting to create a guest with a code that already exists
	ErrDuplicateCode = errors.New("duplicate invitation code")

	// ErrDuplicateRSVP is returned when an RSVP already exists for a guest
	ErrDuplicateRSVP = errors.New("rsvp already exists for this guest")

	// ErrAdminNotFound is returned when an admin user cannot be found
	ErrAdminNotFound = errors.New("admin user not found")

	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
)
