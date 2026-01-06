package models

import (
	"time"
)

// Guest represents an invited guest or household
type Guest struct {
	GuestID          string    `json:"guest_id" dynamodbav:"guest_id"`
	InvitationCode   string    `json:"invitation_code" dynamodbav:"invitation_code"`
	PrimaryGuest     string    `json:"primary_guest" dynamodbav:"primary_guest"`
	HouseholdMembers []string  `json:"household_members" dynamodbav:"household_members"`
	MaxPartySize     int       `json:"max_party_size" dynamodbav:"max_party_size"`
	Email            string    `json:"email,omitempty" dynamodbav:"email,omitempty"`
	Phone            string    `json:"phone,omitempty" dynamodbav:"phone,omitempty"`
	Address          Address   `json:"address,omitempty" dynamodbav:"address,omitempty"`
	CreatedAt        time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// Address represents a mailing address
type Address struct {
	Street  string `json:"street,omitempty" dynamodbav:"street,omitempty"`
	City    string `json:"city,omitempty" dynamodbav:"city,omitempty"`
	State   string `json:"state,omitempty" dynamodbav:"state,omitempty"`
	Zip     string `json:"zip,omitempty" dynamodbav:"zip,omitempty"`
	Country string `json:"country,omitempty" dynamodbav:"country,omitempty"`
}

// RSVPRequest represents the incoming RSVP form submission
type RSVPRequest struct {
	InvitationCode      string   `json:"invitation_code"`
	Attending           bool     `json:"attending"`
	PartySize           int      `json:"party_size"`
	AttendeeNames       []string `json:"attendee_names"`
	DietaryRestrictions []string `json:"dietary_restrictions"`
	SpecialRequests     string   `json:"special_requests"`
}

// RSVP represents a guest's response to the wedding invitation
type RSVP struct {
	RSVPID              string    `json:"rsvp_id" dynamodbav:"rsvp_id"`
	GuestID             string    `json:"guest_id" dynamodbav:"guest_id"`
	Attending           bool      `json:"attending" dynamodbav:"attending"`
	PartySize           int       `json:"party_size" dynamodbav:"party_size"`
	AttendeeNames       []string  `json:"attendee_names" dynamodbav:"attendee_names"`
	DietaryRestrictions []string  `json:"dietary_restrictions" dynamodbav:"dietary_restrictions"`
	SpecialRequests     string    `json:"special_requests,omitempty" dynamodbav:"special_requests,omitempty"`
	SubmittedAt         time.Time `json:"submitted_at" dynamodbav:"submitted_at"`
	UpdatedAt           time.Time `json:"updated_at" dynamodbav:"updated_at"`
	IPAddress           string    `json:"ip_address,omitempty" dynamodbav:"ip_address,omitempty"`
	UserAgent           string    `json:"user_agent,omitempty" dynamodbav:"user_agent,omitempty"`
}
