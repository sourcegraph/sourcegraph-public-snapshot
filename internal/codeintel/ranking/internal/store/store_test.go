package store

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGetStarRank(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO repo (name, stars)
		VALUES
			('foo', 1000),
			('bar',  200),
			('baz',  300),
			('bonk',  50),
			('quux',   0),
			('honk',   0)
	`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	testCases := []struct {
		name     string
		expected float64
	}{
		{"foo", 1.0},  // 1000
		{"baz", 0.8},  // 300
		{"bar", 0.6},  // 200
		{"bonk", 0.4}, // 50
		{"quux", 0.0}, // 0
		{"honk", 0.0}, // 0
	}

	for _, testCase := range testCases {
		stars, err := store.GetStarRank(ctx, api.RepoName(testCase.name))
		if err != nil {
			t.Fatalf("unexpected error getting star rank: %s", err)
		}

		if stars != testCase.expected {
			t.Errorf("unexpected rank. want=%.2f have=%.2f", testCase.expected, stars)
		}
	}
}
