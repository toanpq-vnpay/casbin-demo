package enforcer

import (
	"regexp"
	"strings"
)

func CustomKeyMatch(key1 string, key2 string) bool {
	// Split both paths into segments
	key1Parts := strings.Split(strings.Trim(key1, "/"), "/")
	key2Parts := strings.Split(strings.Trim(key2, "/"), "/")

	if len(key1Parts) != len(key2Parts) {
		return false
	}

	for i := 0; i < len(key1Parts); i++ {
		// Check if this is a parameter pattern
		if strings.HasPrefix(key2Parts[i], "{") && strings.HasSuffix(key2Parts[i], "}") {
			paramName := key2Parts[i][1 : len(key2Parts[i])-1]
			switch {
			case strings.HasSuffix(paramName, "ID"):
				// Must be numeric for ID fields
				if !regexp.MustCompile(`^[0-9]+$`).MatchString(key1Parts[i]) {
					return false
				}
			default:
				// For other parameters, allow alphanumeric and common special chars
				if !regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`).MatchString(key1Parts[i]) {
					return false
				}
			}
			continue
		}

		// Handle wildcard
		if key2Parts[i] == "*" {
			continue
		}

		// Exact match required for non-parameter segments
		if key1Parts[i] != key2Parts[i] {
			return false
		}
	}

	return true
}
