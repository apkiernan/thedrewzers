package invite

import (
	"crypto/rand"
	"fmt"
	"strings"
)

// GenerateCode creates a cryptographically secure 8-character invitation code.
// Uses a charset without ambiguous characters (no I/O/0/1) for better readability.
func GenerateCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 8)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}

	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}

	return string(b), nil
}

// ParseHouseholdMembers splits a semicolon-separated string into trimmed member names.
func ParseHouseholdMembers(members string) []string {
	members = strings.TrimSpace(members)
	if members == "" {
		return []string{}
	}

	parts := strings.Split(members, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
