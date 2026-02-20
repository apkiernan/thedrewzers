package dynamodb

import (
	"testing"

	"github.com/apkiernan/thedrewzers/internal/models"
)

func TestNormalizeSearchText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "lowercases text",
			input: "JoHn SmItH",
			want:  "john smith",
		},
		{
			name:  "trims surrounding spaces",
			input: "  Jane Doe  ",
			want:  "jane doe",
		},
		{
			name:  "collapses repeated whitespace",
			input: "Alex   and    Sam",
			want:  "alex and sam",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSearchText(tt.input)
			if got != tt.want {
				t.Fatalf("normalizeSearchText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchTextMatches(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		query     string
		want      bool
	}{
		{
			name:      "matches contiguous substring",
			candidate: "John and Jane Smith",
			query:     "jane smith",
			want:      true,
		},
		{
			name:      "matches tokenized household name",
			candidate: "Jess & Evan Sahagian",
			query:     "jess sahagian",
			want:      true,
		},
		{
			name:      "matches tokenized household with secondary first name",
			candidate: "Jess & Evan Sahagian",
			query:     "evan sahagian",
			want:      true,
		},
		{
			name:      "does not match unrelated names",
			candidate: "Jess & Evan Sahagian",
			query:     "maria lopez",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := searchTextMatches(tt.candidate, normalizeSearchText(tt.query))
			if got != tt.want {
				t.Fatalf("searchTextMatches(%q, %q) = %v, want %v", tt.candidate, tt.query, got, tt.want)
			}
		})
	}
}

func TestGuestMatchesQuery(t *testing.T) {
	tests := []struct {
		name  string
		guest *models.Guest
		query string
		want  bool
	}{
		{
			name: "matches primary household label",
			guest: &models.Guest{
				PrimaryGuest: "Jess & Evan Sahagian",
			},
			query: "jess sahagian",
			want:  true,
		},
		{
			name: "matches household members",
			guest: &models.Guest{
				PrimaryGuest:     "The Sahagian Family",
				HouseholdMembers: []string{"Jess Sahagian", "Evan Sahagian"},
			},
			query: "evan sahagian",
			want:  true,
		},
		{
			name: "no match for unrelated name",
			guest: &models.Guest{
				PrimaryGuest:     "Jess & Evan Sahagian",
				HouseholdMembers: []string{"Jess Sahagian", "Evan Sahagian"},
			},
			query: "maria lopez",
			want:  false,
		},
		{
			name:  "nil guest does not match",
			guest: nil,
			query: "jess",
			want:  false,
		},
		{
			name: "empty query does not match",
			guest: &models.Guest{
				PrimaryGuest: "Jess Sahagian",
			},
			query: "",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := guestMatchesQuery(tt.guest, normalizeSearchText(tt.query))
			if got != tt.want {
				t.Fatalf("guestMatchesQuery(%v, %q) = %v, want %v", tt.guest, tt.query, got, tt.want)
			}
		})
	}
}
