package handlers

import (
	"encoding/json"
	"net/http"
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"guests": guests,
		"count":  len(guests),
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
	views.App(views.RSVPSuccess()).Render(r.Context(), w)
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

	// Validate party size
	if req.Attending && req.PartySize > guest.MaxPartySize {
		writeJSONError(w, "Party size exceeds maximum allowed", http.StatusBadRequest)
		return
	}

	// Validate party size is at least 1 if attending
	if req.Attending && req.PartySize < 1 {
		req.PartySize = 1
	}

	// Build RSVP record
	now := time.Now().UTC()
	rsvp := &models.RSVP{
		RSVPID:              uuid.New().String(),
		GuestID:             guest.GuestID,
		Attending:           req.Attending,
		PartySize:           req.PartySize,
		AttendeeNames:       req.AttendeeNames,
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
		"success": true,
		"message": "RSVP submitted successfully",
	})
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
