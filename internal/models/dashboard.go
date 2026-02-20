package models

import "time"

// DashboardStats contains aggregated RSVP statistics for the admin dashboard
type DashboardStats struct {
	TotalInvited       int            `json:"total_invited"`
	TotalHouseholds    int            `json:"total_households"`
	TotalInvitedGuests int            `json:"total_invited_guests"`
	TotalResponses     int            `json:"total_responses"`
	TotalAttending     int            `json:"total_attending"`
	TotalDeclined      int            `json:"total_declined"`
	TotalPending       int            `json:"total_pending"`
	ResponseRate       float64        `json:"response_rate"`
	AttendingGuests    int            `json:"attending_guests"`
	MealBreakdown      map[string]int `json:"meal_breakdown"`
	RecentRSVPs        []RecentRSVP   `json:"recent_rsvps"`
}

// RecentRSVP represents a recent RSVP for display on the dashboard
type RecentRSVP struct {
	GuestName   string    `json:"guest_name"`
	Attending   bool      `json:"attending"`
	PartySize   int       `json:"party_size"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// GuestWithRSVP combines a guest with their RSVP response (if any)
type GuestWithRSVP struct {
	Guest *Guest `json:"guest"`
	RSVP  *RSVP  `json:"rsvp,omitempty"`
}
