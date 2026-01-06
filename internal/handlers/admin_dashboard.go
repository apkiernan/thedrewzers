package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apkiernan/thedrewzers/internal/auth"
	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/services"
	"github.com/apkiernan/thedrewzers/internal/views"
)

// AdminDashboardHandler handles admin dashboard requests
type AdminDashboardHandler struct {
	statsService *services.StatsService
}

// NewAdminDashboardHandler creates a new AdminDashboardHandler
func NewAdminDashboardHandler(statsService *services.StatsService) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		statsService: statsService,
	}
}

// HandleDashboard renders the main dashboard with statistics
func (h *AdminDashboardHandler) HandleDashboard(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	stats, err := h.statsService.GetDashboardStats(r.Context())
	if err != nil {
		logger.Error("failed to get dashboard stats", "error", err)
		http.Error(w, "Failed to load dashboard", http.StatusInternalServerError)
		return
	}

	views.AdminDashboard(claims.Name, stats).Render(r.Context(), w)
}

// HandleGuests renders the guest list page
func (h *AdminDashboardHandler) HandleGuests(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	guestsWithRSVPs, err := h.statsService.GetGuestsWithRSVPs(r.Context())
	if err != nil {
		logger.Error("failed to get guests", "error", err)
		http.Error(w, "Failed to load guests", http.StatusInternalServerError)
		return
	}

	views.AdminGuestList(claims.Name, guestsWithRSVPs).Render(r.Context(), w)
}

// HandleExportCSV exports all guest and RSVP data as CSV
func (h *AdminDashboardHandler) HandleExportCSV(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	guestsWithRSVPs, err := h.statsService.GetGuestsWithRSVPs(r.Context())
	if err != nil {
		logger.Error("failed to export data", "error", err)
		http.Error(w, "Failed to export data", http.StatusInternalServerError)
		return
	}

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=rsvps_%s.csv",
		time.Now().Format("2006-01-02")))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write headers
	headers := []string{
		"Primary Guest",
		"Email",
		"Invitation Code",
		"Max Party Size",
		"RSVP Status",
		"Attending",
		"Party Size",
		"Attendee Names",
		"Dietary Restrictions",
		"Special Requests",
		"Submitted At",
	}
	writer.Write(headers)

	// Write data rows
	for _, gwRSVP := range guestsWithRSVPs {
		row := []string{
			gwRSVP.Guest.PrimaryGuest,
			gwRSVP.Guest.Email,
			gwRSVP.Guest.InvitationCode,
			strconv.Itoa(gwRSVP.Guest.MaxPartySize),
		}

		if gwRSVP.RSVP != nil {
			row = append(row,
				"Responded",
				strconv.FormatBool(gwRSVP.RSVP.Attending),
				strconv.Itoa(gwRSVP.RSVP.PartySize),
				strings.Join(gwRSVP.RSVP.AttendeeNames, "; "),
				strings.Join(gwRSVP.RSVP.DietaryRestrictions, "; "),
				gwRSVP.RSVP.SpecialRequests,
				gwRSVP.RSVP.SubmittedAt.Format("2006-01-02 15:04:05"),
			)
		} else {
			row = append(row,
				"Pending",
				"",
				"",
				"",
				"",
				"",
				"",
			)
		}

		writer.Write(row)
	}

	logger.Info("rsvp data exported", "admin", claims.Email, "count", len(guestsWithRSVPs))
}
