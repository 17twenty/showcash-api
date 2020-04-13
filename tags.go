package showcash

import (
	"strings"
)

func isAlphaNumeric(s string) bool {
	if len(s) < 2 {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') &&
			(r < 'A' || r > 'Z') &&
			(r < '0' || r > '9') &&
			(r != '_') &&
			(r != '-') {
			return false
		}
	}
	return true
}

func cleanTags(tags []string) []string {
	var ln int
	for i := range tags {
		if !isAlphaNumeric(tags[i]) || !isAllowed(tags[i]) || len(tags) > 30 {
			continue // drop tag
		}
		tags[ln] = strings.ToLower(tags[i])
		ln++
	}
	return tags[:ln]
}
