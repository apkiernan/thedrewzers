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
	var attending *bool
	switch strings.ToLower(strings.TrimSpace(r.URL.Query().Get("attending"))) {
	case "yes", "true", "1":
		v := true
		attending = &v
	case "no", "false", "0":
		v := false
		attending = &v
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

	// Validate and normalize attendees for attending guests.
	attendees, partySize, validationErr := validateAttendees(req, guest.MaxPartySize)
	if validationErr != "" {
		writeJSONError(w, validationErr, http.StatusBadRequest)
		return
	}

	// Build RSVP record
	now := time.Now().UTC()
	rsvp := &models.RSVP{
		RSVPID:              uuid.New().String(),
		GuestID:             guest.GuestID,
		Attending:           req.Attending,
		PartySize:           partySize,
		Attendees:           attendees,
		AttendeeNames:       attendeeNames(attendees),
		DietaryRestrictions: req.DietaryRestrictions,
		SpecialRequests:     req.SpecialRequests,
		SubmittedAt:         now,
		UpdatedAt:           now,
		IPAddress:           getClientIP(r),
		UserAgent:           r.UserAgent(),
	}

	// If not attending, clear party details
	if !req.Attending {
		rsvp.PartySize = 0
		rsvp.Attendees = nil
		rsvp.AttendeeNames = nil
		rsvp.DietaryRestrictions = nil
		rsvp.SpecialRequests = ""
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
		"attending": rsvp.Attending,
	})
}

func validateAttendees(req models.RSVPRequest, maxPartySize int) ([]models.RSVPAttendee, int, string) {
	if !req.Attending {
		return nil, 0, ""
	}

	attendees := make([]models.RSVPAttendee, 0, len(req.Attendees))
	for _, attendee := range req.Attendees {
		name := strings.TrimSpace(attendee.Name)
		meal := strings.ToLower(strings.TrimSpace(attendee.Meal))
		if name == "" {
			return nil, 0, "Each attending guest must include a name"
		}
		if meal == "" {
			return nil, 0, "Each attending guest must select a meal"
		}
		if !slices.Contains(allowedMealOptions, meal) {
			return nil, 0, "One or more meal selections are invalid"
		}
		attendees = append(attendees, models.RSVPAttendee{
			Name: name,
			Meal: meal,
		})
	}

	if len(attendees) == 0 && len(req.AttendeeNames) > 0 {
		return nil, 0, "Each attending guest must select a meal"
	}
	if len(attendees) == 0 {
		return nil, 0, "At least one attending guest is required"
	}

	partySize := req.PartySize
	if partySize == 0 {
		partySize = len(attendees)
	}
	if partySize < 1 {
		return nil, 0, "Party size must be at least 1"
	}
	if partySize > maxPartySize {
		return nil, 0, "Party size exceeds maximum allowed"
	}
	if len(attendees) != partySize {
		return nil, 0, "Guest count and party size must match"
	}

	return attendees, partySize, ""
}

func attendeeNames(attendees []models.RSVPAttendee) []string {
	names := make([]string, 0, len(attendees))
	for _, attendee := range attendees {
		names = append(names, attendee.Name)
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
