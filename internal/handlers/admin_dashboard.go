package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apkiernan/thedrewzers/internal/auth"
	"github.com/apkiernan/thedrewzers/internal/db"
	"github.com/apkiernan/thedrewzers/internal/invite"
	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/models"
	"github.com/apkiernan/thedrewzers/internal/services"
	"github.com/apkiernan/thedrewzers/internal/views"
)

// AdminDashboardHandler handles admin dashboard requests
type AdminDashboardHandler struct {
	statsService *services.StatsService
	guestRepo    db.GuestRepository
}

// NewAdminDashboardHandler creates a new AdminDashboardHandler
func NewAdminDashboardHandler(statsService *services.StatsService, guestRepo db.GuestRepository) *AdminDashboardHandler {
	return &AdminDashboardHandler{
		statsService: statsService,
		guestRepo:    guestRepo,
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

// HandleGuestDetail renders a detail page for a specific guest household.
func (h *AdminDashboardHandler) HandleGuestDetail(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	guestID := strings.TrimSpace(r.PathValue("id"))
	if guestID == "" {
		http.Redirect(w, r, "/guests", http.StatusFound)
		return
	}

	guestWithRSVP, err := h.statsService.GetGuestWithRSVP(r.Context(), guestID)
	if err != nil {
		logger.Warn("failed to load guest detail", "guest_id", guestID, "error", err)
		views.AdminGuestDetailNotFound(claims.Name).Render(r.Context(), w)
		return
	}

	views.AdminGuestDetail(claims.Name, guestWithRSVP).Render(r.Context(), w)
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
		"Attendee Meals",
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
				formatAttendeeMeals(gwRSVP.RSVP),
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
			)
		}

		writer.Write(row)
	}

	logger.Info("rsvp data exported", "admin", claims.Email, "count", len(guestsWithRSVPs))
}

// HandleAddGuests renders the add guests page
func (h *AdminDashboardHandler) HandleAddGuests(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	success := r.URL.Query().Get("success") == "1"
	imported := r.URL.Query().Get("imported")
	errorMsg := r.URL.Query().Get("error")

	views.AdminAddGuests(claims.Name, success, imported, errorMsg).Render(r.Context(), w)
}

// HandleCreateGuest processes the single guest creation form
func (h *AdminDashboardHandler) HandleCreateGuest(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/guests/add?error=Failed+to+parse+form", http.StatusFound)
		return
	}

	primaryGuest := strings.TrimSpace(r.FormValue("primary_guest"))
	if primaryGuest == "" {
		http.Redirect(w, r, "/guests/add?error=Primary+guest+name+is+required", http.StatusFound)
		return
	}

	maxPartySize := normalizeMaxPartySize(r.FormValue("max_party_size"))

	code, err := invite.GenerateCode()
	if err != nil {
		logger.Error("failed to generate invitation code", "error", err)
		http.Redirect(w, r, "/guests/add?error=Failed+to+generate+invitation+code", http.StatusFound)
		return
	}

	guest := &models.Guest{
		InvitationCode:   code,
		PrimaryGuest:     primaryGuest,
		HouseholdMembers: invite.ParseHouseholdMembers(r.FormValue("household_members")),
		Email:            strings.TrimSpace(r.FormValue("email")),
		MaxPartySize:     maxPartySize,
		Address: models.Address{
			Street:  strings.TrimSpace(r.FormValue("street")),
			City:    strings.TrimSpace(r.FormValue("city")),
			State:   strings.TrimSpace(r.FormValue("state")),
			Zip:     strings.TrimSpace(r.FormValue("zip")),
			Country: "USA",
		},
	}

	if err := h.guestRepo.CreateGuest(r.Context(), guest); err != nil {
		logger.Error("failed to create guest", "error", err)
		http.Redirect(w, r, "/guests/add?error=Failed+to+create+guest", http.StatusFound)
		return
	}

	logger.Info("guest created", "admin", claims.Email, "guest", primaryGuest, "code", code)
	http.Redirect(w, r, "/guests/add?success=1", http.StatusFound)
}

// HandleImportCSV processes a CSV file upload for bulk guest import
func (h *AdminDashboardHandler) HandleImportCSV(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 10MB max
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Redirect(w, r, "/guests/add?error=Failed+to+parse+upload", http.StatusFound)
		return
	}

	file, _, err := r.FormFile("csv_file")
	if err != nil {
		http.Redirect(w, r, "/guests/add?error=No+file+uploaded", http.StatusFound)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		http.Redirect(w, r, "/guests/add?error=Failed+to+read+CSV+file", http.StatusFound)
		return
	}

	if len(records) < 2 {
		http.Redirect(w, r, "/guests/add?error=CSV+must+have+a+header+and+at+least+one+data+row", http.StatusFound)
		return
	}

	headerIndex := csvHeaderIndex(records[0])
	if _, ok := headerIndex["primary_guest"]; !ok {
		http.Redirect(w, r, "/guests/add?error=CSV+must+include+primary_guest+column", http.StatusFound)
		return
	}

	var imported int
	for i, record := range records[1:] {
		primaryGuest := csvCell(record, headerIndex, "primary_guest")
		if primaryGuest == "" {
			continue
		}

		maxPartySize := normalizeMaxPartySize(csvCell(record, headerIndex, "max_party_size"))

		code, err := invite.GenerateCode()
		if err != nil || maxPartySize < 1 {
			logger.Error("failed to generate invitation code during import", "error", err, "row", i+2)
			continue
		}

		guest := &models.Guest{
			InvitationCode:   code,
			PrimaryGuest:     primaryGuest,
			HouseholdMembers: invite.ParseHouseholdMembers(csvCell(record, headerIndex, "household_members")),
			Email:            csvCell(record, headerIndex, "email"),
			MaxPartySize:     maxPartySize,
			Address: models.Address{
				Street:  csvCell(record, headerIndex, "street"),
				City:    csvCell(record, headerIndex, "city"),
				State:   csvCell(record, headerIndex, "state"),
				Zip:     csvCell(record, headerIndex, "zip"),
				Country: "USA",
			},
		}

		if err := h.guestRepo.CreateGuest(r.Context(), guest); err != nil {
			logger.Error("failed to import guest", "error", err, "guest", guest.PrimaryGuest)
			continue
		}
		imported++
	}

	logger.Info("csv import completed", "admin", claims.Email, "imported", imported)

	if imported == 0 {
		http.Redirect(w, r, "/guests/add?error=No+guests+were+imported+from+CSV", http.StatusFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/guests/add?imported=%d", imported), http.StatusFound)
}

func formatAttendeeMeals(rsvp *models.RSVP) string {
	if rsvp == nil {
		return ""
	}
	if len(rsvp.Attendees) == 0 {
		return strings.Join(rsvp.AttendeeNames, "; ")
	}

	parts := make([]string, 0, len(rsvp.Attendees))
	for _, attendee := range rsvp.Attendees {
		name := strings.TrimSpace(attendee.Name)
		meal := strings.TrimSpace(attendee.Meal)
		if name == "" && meal == "" {
			continue
		}
		if meal == "" {
			parts = append(parts, name)
			continue
		}
		if name == "" {
			parts = append(parts, meal)
			continue
		}
		parts = append(parts, fmt.Sprintf("%s (%s)", name, meal))
	}
	return strings.Join(parts, "; ")
}

func csvHeaderIndex(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, column := range header {
		normalized := strings.ToLower(strings.TrimSpace(column))
		if normalized != "" {
			index[normalized] = i
		}
	}
	return index
}

func csvCell(row []string, header map[string]int, column string) string {
	idx, ok := header[column]
	if !ok || idx < 0 || idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

func normalizeMaxPartySize(raw string) int {
	parsed, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 1
	}
	if parsed <= 1 {
		return 1
	}
	return 2
}
