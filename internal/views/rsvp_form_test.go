package views

import (
	"reflect"
	"testing"

	"github.com/apkiernan/thedrewzers/internal/models"
)

func TestNamedInvitees(t *testing.T) {
	tests := []struct {
		name  string
		guest *models.Guest
		want  []string
	}{
		{
			name: "uses explicit household members",
			guest: &models.Guest{
				PrimaryGuest:     "Jess & Evan Sahagian",
				HouseholdMembers: []string{"Jess Sahagian", "Evan Sahagian"},
			},
			want: []string{"Jess Sahagian", "Evan Sahagian"},
		},
		{
			name: "falls back to primary label when no members",
			guest: &models.Guest{
				PrimaryGuest: "Sarah Williams",
			},
			want: []string{"Sarah Williams"},
		},
		{
			name: "trims blank members",
			guest: &models.Guest{
				PrimaryGuest:     "Ignored Label",
				HouseholdMembers: []string{"  ", "Jess Sahagian", ""},
			},
			want: []string{"Jess Sahagian"},
		},
		{
			name:  "nil guest returns empty placeholder",
			guest: nil,
			want:  []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := namedInvitees(tt.guest)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("namedInvitees(%+v) = %v, want %v", tt.guest, got, tt.want)
			}
		})
	}
}

func TestInitialAttendeesClampsToMaxPartySize(t *testing.T) {
	guest := &models.Guest{
		PrimaryGuest:     "Family",
		HouseholdMembers: []string{"A", "B", "C"},
		MaxPartySize:     2,
	}

	got := initialAttendees(guest, nil)
	want := []models.RSVPAttendee{
		{Name: "A", Meal: ""},
		{Name: "B", Meal: ""},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("initialAttendees(%+v, nil) = %#v, want %#v", guest, got, want)
	}
}

func TestDefaultPartySize(t *testing.T) {
	tests := []struct {
		name  string
		guest *models.Guest
		want  int
	}{
		{
			name: "defaults to number of named invitees",
			guest: &models.Guest{
				PrimaryGuest:     "Jess & Evan Sahagian",
				HouseholdMembers: []string{"Jess Sahagian", "Evan Sahagian"},
				MaxPartySize:     2,
			},
			want: 2,
		},
		{
			name: "caps default at max party size",
			guest: &models.Guest{
				PrimaryGuest:     "Family",
				HouseholdMembers: []string{"A", "B", "C"},
				MaxPartySize:     2,
			},
			want: 2,
		},
		{
			name: "generic plus one defaults to one",
			guest: &models.Guest{
				PrimaryGuest:     "Jess Sahagian + Guest",
				HouseholdMembers: []string{"Jess Sahagian"},
				MaxPartySize:     2,
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := defaultPartySize(tt.guest)
			if got != tt.want {
				t.Fatalf("defaultPartySize(%+v) = %d, want %d", tt.guest, got, tt.want)
			}
		})
	}
}
