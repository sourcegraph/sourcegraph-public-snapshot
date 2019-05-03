package db

import "context"

type MockRecentSearches struct {
	Add              func(ctx context.Context, q string) error
	DeleteExcessRows func(ctx context.Context, limit int) error
	Get              func(ctx context.Context) ([]string, error)
}
