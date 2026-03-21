package handlers

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/apkiernan/thedrewzers/internal/db"
	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/models"
	"github.com/apkiernan/thedrewzers/internal/views"
)

// RSVPHandler handles RSVP-related HTTP requests
type RSVPHandler struct {
	guestRepo db.GuestRepository
	rsvpRepo  db.RSVPRepository
}

var allowedMealOptions = []string{"Roasted Boneless Chicken Breast", "Grilled Brandt Farms 10z NY Strip", "Roasted Cauliflower Al Pastor (GF-V)"}

type guestSearchResult struct {
	GuestID          string   `json:"guest_id"`
	PrimaryGuest     string   `json:"primary_guest"`
	HouseholdMembers []string `json:"household_members"`
	MaxPartySize     int      `json:"max_party_size"`
}

// NewRSVPHandler creates a new RSVPHandler with the given repositories
func NewRSVPHandler(guestRepo db.GuestRepository, rsvpRepo db.RSVPRepository) *RSVPHandler {
	return &RSVPHandler{
		guestRepo: guestRepo,
		rsvpRepo:  rsvpRepo,
	}
}

// HandleRSVPPage displays the RSVP name search form
func (h *RSVPHandler) HandleRSVPPage(w http.ResponseWriter, r *http.Request) {
	views.App(views.RSVPNameSearch()).Render(r.Context(), w)
}

// HandleRSVPSearch handles name-based guest search
func (h *RSVPHandler) HandleRSVPSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeJSONError(w, "Name is required", http.StatusBadRequest)
		return
	}

	guests, err := h.guestRepo.SearchGuestsByName(r.Context(), name)
	if err != nil {
		logger.Error("failed to search guests", "name", name, "error", err)
		writeJSONError(w, "Search failed", http.StatusInternalServerError)
		return
	}

	results := make([]guestSearchResult, 0, len(guests))
	for _, guest := range guests {
		results = append(results, guestSearchResult{
			GuestID:          guest.GuestID,
			PrimaryGuest:     guest.PrimaryGuest,
			HouseholdMembers: guest.HouseholdMembers,
			MaxPartySize:     guest.MaxPartySize,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"guests": results,
		"count":  len(results),
	})
}

// HandleRSVPForm displays the RSVP form for a specific guest
func (h *RSVPHandler) HandleRSVPForm(w http.ResponseWriter, r *http.Request) {
	guestID := r.URL.Query().Get("id")
	if guestID == "" {
		http.Redirect(w, r, "/rsvp", http.StatusFound)
		return
	}

	guest, err := h.guestRepo.GetGuest(r.Context(), guestID)
	if err != nil {
		logger.Warn("guest not found for form", "guest_id", guestID, "error", err)
		views.App(views.RSVPNotFound()).Render(r.Context(), w)
		return
	}

	// Check for existing RSVP
	existingRSVP, _ := h.rsvpRepo.GetRSVPByGuestID(r.Context(), guest.GuestID)

	// Render RSVP form
	views.App(views.RSVPForm(guest, existingRSVP)).Render(r.Context(), w)
}

// HandleRSVPSuccess displays the success confirmation page
func (h *RSVPHandler) HandleRSVPSuccess(w http.ResponseWriter, r *http.Request) {
	attending := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("attending")))
	// Normalize to one of: "yes", "no", "partial"
	switch attending {
	case "yes", "no", "partial":
		// valid
	default:
		attending = "yes"
	}
	views.App(views.RSVPSuccess(attending)).Render(r.Context(), w)
}

// HandleRSVPSubmit processes RSVP form submissions
func (h *RSVPHandler) HandleRSVPSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req models.RSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode RSVP request", "error", err)
		writeJSONError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate guest ID
	req.GuestID = strings.TrimSpace(req.GuestID)
	if req.GuestID == "" {
		writeJSONError(w, "Guest ID is required", http.StatusBadRequest)
		return
	}

	// Validate guest exists
	guest, err := h.guestRepo.GetGuest(r.Context(), req.GuestID)
	if err != nil {
		logger.Warn("guest not found", "guest_id", req.GuestID, "error", err)
		writeJSONError(w, "Guest not found", http.StatusBadRequest)
		return
	}

	// Validate and normalize attendees with per-member attending flags.
	attendees, validationErr := validateAttendees(req, guest.MaxPartySize)
	if validationErr != "" {
		writeJSONError(w, validationErr, http.StatusBadRequest)
		return
	}

	// Derive top-level attending: true if any member is attending
	anyAttending := false
	attendingCount := 0
	for _, a := range attendees {
		if a.Attending {
			anyAttending = true
			attendingCount++
		}
	}

	// Determine attending status for response: "yes", "no", or "partial"
	attendingStatus := "no"
	if anyAttending {
		if attendingCount == len(attendees) {
			attendingStatus = "yes"
		} else {
			attendingStatus = "partial"
		}
	}

	// Build RSVP record
	now := time.Now().UTC()
	rsvp := &models.RSVP{
		RSVPID:              uuid.New().String(),
		GuestID:             guest.GuestID,
		Attending:           anyAttending,
		PartySize:           attendingCount,
		Attendees:           attendees,
		AttendeeNames:       attendingNames(attendees),
		DietaryRestrictions: req.DietaryRestrictions,
		SpecialRequests:     req.SpecialRequests,
		SubmittedAt:         now,
		UpdatedAt:           now,
		IPAddress:           getClientIP(r),
		UserAgent:           r.UserAgent(),
	}

	// Check if updating existing RSVP
	existingRSVP, _ := h.rsvpRepo.GetRSVPByGuestID(r.Context(), guest.GuestID)
	if existingRSVP != nil {
		// Update existing RSVP
		rsvp.RSVPID = existingRSVP.RSVPID
		rsvp.SubmittedAt = existingRSVP.SubmittedAt // Preserve original submission time
		err = h.rsvpRepo.UpdateRSVP(r.Context(), rsvp)
	} else {
		// Create new RSVP
		err = h.rsvpRepo.CreateRSVP(r.Context(), rsvp)
	}

	if err != nil {
		logger.Error("failed to save RSVP", "guest_id", guest.GuestID, "error", err)
		writeJSONError(w, "Failed to save RSVP", http.StatusInternalServerError)
		return
	}

	logger.Info("rsvp saved", "guest", guest.PrimaryGuest, "attending", rsvp.Attending, "party_size", rsvp.PartySize)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   "RSVP submitted successfully",
		"attending": attendingStatus,
	})
}

func validateAttendees(req models.RSVPRequest, maxPartySize int) ([]models.RSVPAttendee, string) {
	if len(req.Attendees) == 0 {
		return nil, "At least one household member must be included"
	}

	// Every member must have responded (have a name set)
	attendingCount := 0
	attendees := make([]models.RSVPAttendee, 0, len(req.Attendees))
	for _, attendee := range req.Attendees {
		name := strings.TrimSpace(attendee.Name)
		if name == "" {
			return nil, "Each household member must have a name"
		}

		a := models.RSVPAttendee{
			Name:      name,
			Attending: attendee.Attending,
		}

		if attendee.Attending {
			attendingCount++
			meal := strings.ToLower(strings.TrimSpace(attendee.Meal))
			if meal == "" {
				return nil, "Please select a meal for " + name
			}
			if !slices.ContainsFunc(allowedMealOptions, func(m string) bool {
				return strings.EqualFold(m, meal)
			}) {
				return nil, "Invalid meal selection for " + name
			}
			a.Meal = meal
		}
		// Non-attending members have no meal stored

		attendees = append(attendees, a)
	}

	if attendingCount > maxPartySize {
		return nil, "Number of attending guests exceeds maximum allowed"
	}

	return attendees, ""
}

// attendingNames returns the names of attending members only.
func attendingNames(attendees []models.RSVPAttendee) []string {
	names := make([]string, 0, len(attendees))
	for _, attendee := range attendees {
		if attendee.Attending {
			names = append(names, attendee.Name)
		}
	}
	return names
}

// getClientIP extracts the client IP from the request, checking X-Forwarded-For header first
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from CloudFront/load balancer)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	return r.RemoteAddr
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
