package services

import (
	"context"
	"sort"

	"github.com/apkiernan/thedrewzers/internal/db"
	"github.com/apkiernan/thedrewzers/internal/models"
)

// SeatingService provides seating chart management
type SeatingService struct {
	tableRepo db.TableRepository
	guestRepo db.GuestRepository
	rsvpRepo  db.RSVPRepository
}

// NewSeatingService creates a new SeatingService
func NewSeatingService(tableRepo db.TableRepository, guestRepo db.GuestRepository, rsvpRepo db.RSVPRepository) *SeatingService {
	return &SeatingService{
		tableRepo: tableRepo,
		guestRepo: guestRepo,
		rsvpRepo:  rsvpRepo,
	}
}

// GetTablesWithGuests returns all tables with their assigned guests
func (s *SeatingService) GetTablesWithGuests(ctx context.Context) ([]*models.TableWithGuests, error) {
	tables, err := s.tableRepo.ListTables(ctx)
	if err != nil {
		return nil, err
	}

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

	// Group guests by table
	guestsByTable := make(map[string][]*models.GuestWithRSVP)
	for _, guest := range guests {
		if guest.TableID != "" {
			guestsByTable[guest.TableID] = append(guestsByTable[guest.TableID], &models.GuestWithRSVP{
				Guest: guest,
				RSVP:  rsvpMap[guest.GuestID],
			})
		}
	}

	// Build result
	result := make([]*models.TableWithGuests, 0, len(tables))
	for _, table := range tables {
		tableGuests := guestsByTable[table.TableID]
		seatedCount := 0
		for _, gw := range tableGuests {
			if gw.RSVP != nil && gw.RSVP.Attending {
				seatedCount += gw.RSVP.PartySize
			}
		}
		result = append(result, &models.TableWithGuests{
			Table:       table,
			Guests:      tableGuests,
			SeatedCount: seatedCount,
		})
	}

	// Sort by table name
	sort.Slice(result, func(i, j int) bool {
		return result[i].Table.Name < result[j].Table.Name
	})

	return result, nil
}

// GetUnassignedGuests returns guests who are attending but not assigned to a table
func (s *SeatingService) GetUnassignedGuests(ctx context.Context) ([]*models.GuestWithRSVP, error) {
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

	// Find unassigned guests who are attending
	var unassigned []*models.GuestWithRSVP
	for _, guest := range guests {
		rsvp := rsvpMap[guest.GuestID]
		// Include if: no table assigned AND (has RSVP attending OR no RSVP yet)
		if guest.TableID == "" {
			if rsvp == nil || rsvp.Attending {
				unassigned = append(unassigned, &models.GuestWithRSVP{
					Guest: guest,
					RSVP:  rsvp,
				})
			}
		}
	}

	// Sort by primary guest name
	sort.Slice(unassigned, func(i, j int) bool {
		return unassigned[i].Guest.PrimaryGuest < unassigned[j].Guest.PrimaryGuest
	})

	return unassigned, nil
}

// AssignGuestToTable assigns a guest to a table
func (s *SeatingService) AssignGuestToTable(ctx context.Context, guestID, tableID string) error {
	guest, err := s.guestRepo.GetGuest(ctx, guestID)
	if err != nil {
		return err
	}

	// Verify table exists (if tableID is not empty)
	if tableID != "" {
		_, err = s.tableRepo.GetTable(ctx, tableID)
		if err != nil {
			return err
		}
	}

	guest.TableID = tableID
	return s.guestRepo.UpdateGuest(ctx, guest)
}

// CreateTable creates a new seating table
func (s *SeatingService) CreateTable(ctx context.Context, table *models.Table) error {
	return s.tableRepo.CreateTable(ctx, table)
}

// UpdateTable updates a seating table
func (s *SeatingService) UpdateTable(ctx context.Context, table *models.Table) error {
	return s.tableRepo.UpdateTable(ctx, table)
}

// DeleteTable deletes a seating table and unassigns all guests from it
func (s *SeatingService) DeleteTable(ctx context.Context, tableID string) error {
	// First, unassign all guests from this table
	guests, err := s.guestRepo.ListGuests(ctx)
	if err != nil {
		return err
	}

	for _, guest := range guests {
		if guest.TableID == tableID {
			guest.TableID = ""
			if err := s.guestRepo.UpdateGuest(ctx, guest); err != nil {
				return err
			}
		}
	}

	// Then delete the table
	return s.tableRepo.DeleteTable(ctx, tableID)
}

// GetTable gets a single table by ID
func (s *SeatingService) GetTable(ctx context.Context, tableID string) (*models.Table, error) {
	return s.tableRepo.GetTable(ctx, tableID)
}

// GetSeatingStats returns seating statistics
func (s *SeatingService) GetSeatingStats(ctx context.Context) (*SeatingStats, error) {
	tables, err := s.GetTablesWithGuests(ctx)
	if err != nil {
		return nil, err
	}

	unassigned, err := s.GetUnassignedGuests(ctx)
	if err != nil {
		return nil, err
	}

	stats := &SeatingStats{
		TotalTables:      len(tables),
		TotalCapacity:    0,
		TotalSeated:      0,
		UnassignedGuests: len(unassigned),
	}

	// Count unassigned guest party sizes
	for _, gw := range unassigned {
		if gw.RSVP != nil && gw.RSVP.Attending {
			stats.UnassignedAttendees += gw.RSVP.PartySize
		} else if gw.RSVP == nil {
			// Pending RSVP - count max party size
			stats.UnassignedAttendees += gw.Guest.MaxPartySize
		}
	}

	for _, tw := range tables {
		stats.TotalCapacity += tw.Table.Capacity
		stats.TotalSeated += tw.SeatedCount
	}

	return stats, nil
}

// SeatingStats holds seating chart statistics
type SeatingStats struct {
	TotalTables         int `json:"total_tables"`
	TotalCapacity       int `json:"total_capacity"`
	TotalSeated         int `json:"total_seated"`
	UnassignedGuests    int `json:"unassigned_guests"`    // Number of households
	UnassignedAttendees int `json:"unassigned_attendees"` // Number of people
}
