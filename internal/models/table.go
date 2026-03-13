package models

import (
	"time"
)

// Table represents a seating table at the venue
type Table struct {
	TableID   string    `json:"table_id" dynamodbav:"table_id"`
	Name      string    `json:"name" dynamodbav:"name"`           // "Table 1", "Head Table"
	Capacity  int       `json:"capacity" dynamodbav:"capacity"`   // Max seats at this table
	Shape     string    `json:"shape" dynamodbav:"shape"`         // "round", "rectangle"
	PositionX float64   `json:"position_x" dynamodbav:"position_x"` // For visual editor (Phase 2)
	PositionY float64   `json:"position_y" dynamodbav:"position_y"` // For visual editor (Phase 2)
	CreatedAt time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// TableWithGuests combines a table with its assigned guests
type TableWithGuests struct {
	Table  *Table           `json:"table"`
	Guests []*GuestWithRSVP `json:"guests"`
	// SeatedCount is the total number of attending guests at this table
	SeatedCount int `json:"seated_count"`
}
