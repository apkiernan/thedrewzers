package services

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/apkiernan/thedrewzers/internal/db"
	"github.com/apkiernan/thedrewzers/internal/models"
)

// StatsService provides dashboard statistics and data aggregation
type StatsService struct {
	guestRepo db.GuestRepository
	rsvpRepo  db.RSVPRepository
}

// NewStatsService creates a new StatsService
func NewStatsService(guestRepo db.GuestRepository, rsvpRepo db.RSVPRepository) *StatsService {
	return &StatsService{
		guestRepo: guestRepo,
		rsvpRepo:  rsvpRepo,
	}
}

// GetDashboardStats calculates and returns dashboard statistics
func (s *StatsService) GetDashboardStats(ctx context.Context) (*models.DashboardStats, error) {
	guests, err := s.guestRepo.ListGuests(ctx)
	if err != nil {
		return nil, err
	}

	rsvps, err := s.rsvpRepo.ListRSVPs(ctx)
	if err != nil {
		return nil, err
	}

	// Create guest map for quick lookup
	guestMap := make(map[string]*models.Guest)
	for _, guest := range guests {
		guestMap[guest.GuestID] = guest
	}

	stats := &models.DashboardStats{
		TotalInvited:   len(guests),
		TotalResponses: len(rsvps),
		MealBreakdown:  make(map[string]int),
		RecentRSVPs:    make([]models.RecentRSVP, 0),
	}

	// Calculate attending/declined and meal selections
	for _, rsvp := range rsvps {
		if rsvp.Attending {
			stats.TotalAttending++
			stats.AttendingGuests += rsvp.PartySize

			for _, attendee := range rsvp.Attendees {
				normalized := strings.TrimSpace(strings.ToLower(attendee.Meal))
				if normalized != "" {
					stats.MealBreakdown[normalized]++
				}
			}
		} else {
			stats.TotalDeclined++
		}
	}

	stats.TotalPending = stats.TotalInvited - stats.TotalResponses

	if stats.TotalInvited > 0 {
		stats.ResponseRate = float64(stats.TotalResponses) / float64(stats.TotalInvited) * 100
	}

	// Get recent RSVPs (last 10, sorted by submission time)
	stats.RecentRSVPs = s.getRecentRSVPs(rsvps, guestMap, 10)

	return stats, nil
}

// getRecentRSVPs returns the most recent RSVPs
func (s *StatsService) getRecentRSVPs(rsvps []*models.RSVP, guestMap map[string]*models.Guest, limit int) []models.RecentRSVP {
	// Sort RSVPs by submission time (newest first)
	sorted := make([]*models.RSVP, len(rsvps))
	copy(sorted, rsvps)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].SubmittedAt.After(sorted[j].SubmittedAt)
	})

	recent := make([]models.RecentRSVP, 0, limit)
	for i := 0; i < len(sorted) && len(recent) < limit; i++ {
		rsvp := sorted[i]
		guest, ok := guestMap[rsvp.GuestID]
		if !ok {
			continue
		}
		recent = append(recent, models.RecentRSVP{
			GuestName:   guest.PrimaryGuest,
			Attending:   rsvp.Attending,
			PartySize:   rsvp.PartySize,
			SubmittedAt: rsvp.SubmittedAt,
		})
	}

	return recent
}

// GetGuestsWithRSVPs returns all guests with their RSVP data joined
func (s *StatsService) GetGuestsWithRSVPs(ctx context.Context) ([]*models.GuestWithRSVP, error) {
	guests, err := s.guestRepo.ListGuests(ctx)
	if err != nil {
		return nil, err
	}

	rsvps, err := s.rsvpRepo.ListRSVPs(ctx)
	if err != nil {
		return nil, err
	}

	// Create RSVP map by guest ID
	rsvpMap := make(map[string]*models.RSVP)
	for _, rsvp := range rsvps {
		rsvpMap[rsvp.GuestID] = rsvp
	}

	// Combine data
	result := make([]*models.GuestWithRSVP, 0, len(guests))
	for _, guest := range guests {
		result = append(result, &models.GuestWithRSVP{
			Guest: guest,
			RSVP:  rsvpMap[guest.GuestID],
		})
	}

	return result, nil
}

// GetGuestWithRSVP returns one guest joined with their RSVP (if present).
func (s *StatsService) GetGuestWithRSVP(ctx context.Context, guestID string) (*models.GuestWithRSVP, error) {
	guest, err := s.guestRepo.GetGuest(ctx, guestID)
	if err != nil {
		return nil, err
	}

	rsvp, err := s.rsvpRepo.GetRSVPByGuestID(ctx, guestID)
	if err != nil && !errors.Is(err, models.ErrRSVPNotFound) {
		return nil, err
	}
	if errors.Is(err, models.ErrRSVPNotFound) {
		rsvp = nil
	}

	return &models.GuestWithRSVP{
		Guest: guest,
		RSVP:  rsvp,
	}, nil
}
