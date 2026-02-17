package dynamodb

import "testing"

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
