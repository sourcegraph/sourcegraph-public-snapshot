package authz

import (
	"fmt"
	"time"
)

func parseTTL(ttl string) (time.Duration, error) {
	defaultValue := 3 * time.Hour
	if ttl == "" {
		return defaultValue, nil
	}
	d, err := time.ParseDuration(ttl)
	if err != nil {
		return defaultValue, fmt.Errorf("Could not parse time duration %q.", ttl)
	}
	return d, nil
}
