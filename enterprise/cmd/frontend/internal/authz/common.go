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
		return defaultValue, fmt.Errorf("authorization.ttl: %s", err)
	}
	return d, nil
}
