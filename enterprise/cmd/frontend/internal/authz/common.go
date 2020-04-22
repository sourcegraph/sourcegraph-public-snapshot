package authz

import (
	"fmt"
	"time"
)

// ParseTTL parses ttl string to a valid time duration.
func ParseTTL(ttl string) (time.Duration, error) {
	defaultValue := 3 * time.Hour
	if ttl == "" {
		return defaultValue, nil
	}
	d, err := time.ParseDuration(ttl)
	if err != nil {
		return defaultValue, fmt.Errorf("authorization.ttl: %s", err)
	}
	return d, nil
}
