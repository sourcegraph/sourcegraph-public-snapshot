package graphql

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func ExtractRestID(globalID string) (int, error) {
	parts := strings.Split(globalID, "/")
	if len(parts) == 1 {
		return 0, errors.New("global ID is not a URI")
	}

	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return 0, errors.New("cannot parse numeric part of global ID")
	}
	return id, nil
}

func OptionalTime(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
