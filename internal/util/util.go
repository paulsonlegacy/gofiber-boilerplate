// Package util provides small, stateless helpers used across the application.
// No business logic here — only pure functions with zero external dependencies.
package util

import (
	"regexp"
	"strings"
)

// ContainsSQLInjection returns true if the input contains common SQL injection patterns.
// Not a substitute for parameterized queries — GORM already handles that.
// Use this as an extra guard on user-visible fields like names and descriptions.
func ContainsSQLInjection(input string) bool {
	patterns := []string{
		`'`, `"`, `;`, "--", "/*", "*/", "xp_", "EXEC", "DROP", "INSERT", "SELECT",
		"DELETE", "UPDATE", "UNION", "WHERE", "OR 1=1", "AND 1=1",
	}
	upper := strings.ToUpper(input)
	for _, p := range patterns {
		if strings.Contains(upper, strings.ToUpper(p)) {
			return true
		}
	}
	return false
}

// SanitizeInput removes leading/trailing whitespace and strips HTML tags.
func SanitizeInput(input string) string {
	input = strings.TrimSpace(input)
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

// TruncateString truncates a string to maxLen characters, appending "..." if truncated.
func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

// SliceContains reports whether slice contains target (case-insensitive for strings).
func SliceContains(slice []string, target string) bool {
	target = strings.ToLower(target)
	for _, s := range slice {
		if strings.ToLower(s) == target {
			return true
		}
	}
	return false
}
