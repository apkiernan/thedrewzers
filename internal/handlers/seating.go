package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/apkiernan/thedrewzers/internal/auth"
	"github.com/apkiernan/thedrewzers/internal/logger"
	"github.com/apkiernan/thedrewzers/internal/models"
	"github.com/apkiernan/thedrewzers/internal/services"
	"github.com/apkiernan/thedrewzers/internal/views"
)

// SeatingHandler handles seating chart requests
type SeatingHandler struct {
	seatingService *services.SeatingService
}

// NewSeatingHandler creates a new SeatingHandler
func NewSeatingHandler(seatingService *services.SeatingService) *SeatingHandler {
	return &SeatingHandler{
		seatingService: seatingService,
	}
}

// HandleSeatingPage renders the seating chart management page
func (h *SeatingHandler) HandleSeatingPage(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	tables, err := h.seatingService.GetTablesWithGuests(r.Context())
	if err != nil {
		logger.Error("failed to get tables", "error", err)
		http.Error(w, "Failed to load seating chart", http.StatusInternalServerError)
		return
	}

	unassigned, err := h.seatingService.GetUnassignedGuests(r.Context())
	if err != nil {
		logger.Error("failed to get unassigned guests", "error", err)
		http.Error(w, "Failed to load unassigned guests", http.StatusInternalServerError)
		return
	}

	stats, err := h.seatingService.GetSeatingStats(r.Context())
	if err != nil {
		logger.Error("failed to get seating stats", "error", err)
		http.Error(w, "Failed to load seating stats", http.StatusInternalServerError)
		return
	}

	views.AdminSeating(claims.Name, tables, unassigned, stats).Render(r.Context(), w)
}

// HandleCreateTable handles table creation
func (h *SeatingHandler) HandleCreateTable(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Capacity int    `json:"capacity"`
		Shape    string `json:"shape"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate
	if strings.TrimSpace(req.Name) == "" {
		http.Error(w, "Table name is required", http.StatusBadRequest)
		return
	}
	if req.Capacity < 1 {
		http.Error(w, "Capacity must be at least 1", http.StatusBadRequest)
		return
	}
	if req.Shape == "" {
		req.Shape = "round" // Default shape
	}

	table := &models.Table{
		Name:     strings.TrimSpace(req.Name),
		Capacity: req.Capacity,
		Shape:    req.Shape,
	}

	if err := h.seatingService.CreateTable(r.Context(), table); err != nil {
		logger.Error("failed to create table", "error", err)
		http.Error(w, "Failed to create table", http.StatusInternalServerError)
		return
	}

	logger.Info("table created", "admin", claims.Email, "table", table.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(table)
}

// HandleUpdateTable handles table updates
func (h *SeatingHandler) HandleUpdateTable(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tableID := r.PathValue("id")
	if tableID == "" {
		http.Error(w, "Table ID required", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Capacity int    `json:"capacity"`
		Shape    string `json:"shape"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	table, err := h.seatingService.GetTable(r.Context(), tableID)
	if err != nil {
		if err == models.ErrTableNotFound {
			http.Error(w, "Table not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get table", http.StatusInternalServerError)
		}
		return
	}

	// Update fields
	if strings.TrimSpace(req.Name) != "" {
		table.Name = strings.TrimSpace(req.Name)
	}
	if req.Capacity > 0 {
		table.Capacity = req.Capacity
	}
	if req.Shape != "" {
		table.Shape = req.Shape
	}

	if err := h.seatingService.UpdateTable(r.Context(), table); err != nil {
		logger.Error("failed to update table", "error", err)
		http.Error(w, "Failed to update table", http.StatusInternalServerError)
		return
	}

	logger.Info("table updated", "admin", claims.Email, "table", table.Name)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(table)
}

// HandleDeleteTable handles table deletion
func (h *SeatingHandler) HandleDeleteTable(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tableID := r.PathValue("id")
	if tableID == "" {
		http.Error(w, "Table ID required", http.StatusBadRequest)
		return
	}

	if err := h.seatingService.DeleteTable(r.Context(), tableID); err != nil {
		logger.Error("failed to delete table", "error", err)
		http.Error(w, "Failed to delete table", http.StatusInternalServerError)
		return
	}

	logger.Info("table deleted", "admin", claims.Email, "table_id", tableID)

	w.WriteHeader(http.StatusNoContent)
}

// HandleAssignGuest handles guest table assignment
func (h *SeatingHandler) HandleAssignGuest(w http.ResponseWriter, r *http.Request) {
	claims := auth.GetClaims(r.Context())
	if claims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		GuestID string `json:"guest_id"`
		TableID string `json:"table_id"` // Empty string to unassign
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.GuestID == "" {
		http.Error(w, "Guest ID is required", http.StatusBadRequest)
		return
	}

	if err := h.seatingService.AssignGuestToTable(r.Context(), req.GuestID, req.TableID); err != nil {
		if err == models.ErrGuestNotFound {
			http.Error(w, "Guest not found", http.StatusNotFound)
		} else if err == models.ErrTableNotFound {
			http.Error(w, "Table not found", http.StatusNotFound)
		} else {
			logger.Error("failed to assign guest", "error", err)
			http.Error(w, "Failed to assign guest", http.StatusInternalServerError)
		}
		return
	}

	action := "assigned"
	if req.TableID == "" {
		action = "unassigned"
	}
	logger.Info("guest "+action, "admin", claims.Email, "guest_id", req.GuestID, "table_id", req.TableID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
