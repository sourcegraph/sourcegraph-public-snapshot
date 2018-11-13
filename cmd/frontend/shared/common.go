package shared

import (
	"fmt"
	"time"
)

func parseTTLOrDefault(ttl string, defaultVal time.Duration, warnings []string) (_ time.Duration, updatedWarnings []string) {
	if ttl == "" {
		return defaultVal, warnings
	}
	parsed, err := time.ParseDuration(ttl)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Could not parse time duration %q, falling back to %v.", ttl, defaultVal))
		return defaultVal, warnings
	}
	return parsed, warnings
}
