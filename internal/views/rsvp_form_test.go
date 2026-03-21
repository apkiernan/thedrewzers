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
		{Name: "A", Attending: false, Meal: ""},
		{Name: "B", Attending: false, Meal: ""},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("initialAttendees(%+v, nil) = %#v, want %#v", guest, got, want)
	}
}

func TestInitialAttendeesRestoresExistingRSVP(t *testing.T) {
	guest := &models.Guest{
		PrimaryGuest:     "Jess & Evan",
		HouseholdMembers: []string{"Jess", "Evan"},
		MaxPartySize:     2,
	}
	existing := &models.RSVP{
		Attendees: []models.RSVPAttendee{
			{Name: "Jess", Attending: true, Meal: "chicken"},
			{Name: "Evan", Attending: false, Meal: ""},
		},
	}

	got := initialAttendees(guest, existing)
	if !reflect.DeepEqual(got, existing.Attendees) {
		t.Fatalf("initialAttendees with existing RSVP = %#v, want %#v", got, existing.Attendees)
	}
}
