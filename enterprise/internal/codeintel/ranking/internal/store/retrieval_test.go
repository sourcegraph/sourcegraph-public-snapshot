package store

import (
	"context"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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
	store := New(&observation.TestContext, db)

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

func TestDocumentRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)
	repoName := api.RepoName("foo")

	key := rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_definition_id, max_reference_id, max_path_id, mappers_started_at, reducer_completed_at)
		VALUES
			($1,  1000, 1000, 1000, NOW(), NOW())
	`,
		key,
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name, stars) VALUES ('foo', 1000)`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), repoName, map[string]float64{
		"cmd/main.go":        2, // no longer referenced
		"internal/secret.go": 3,
		"internal/util.go":   4,
		"README.md":          5, // no longer referenced
	}, key); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), repoName, map[string]float64{
		"cmd/args.go":        8, // new
		"internal/secret.go": 7, // edited
		"internal/util.go":   6, // edited
	}, key); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}

	ranks, _, err := store.GetDocumentRanks(ctx, repoName)
	if err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	expectedRanks := map[string]float64{
		"cmd/args.go":        8,
		"internal/secret.go": 7,
		"internal/util.go":   6,
	}
	if diff := cmp.Diff(expectedRanks, ranks); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}
}

func TestGetReferenceCountStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_definition_id, max_reference_id, max_path_id, mappers_started_at, reducer_completed_at)
		VALUES
			($1,  1000, 1000, 1000, NOW(), NOW())
	`,
		key,
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo'), ('bar'), ('baz')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), api.RepoName("foo"), map[string]float64{"foo": 18, "bar": 3985, "baz": 5260}, key); err != nil {
		t.Fatalf("failed to set document ranks: %s", err)
	}
	if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), api.RepoName("bar"), map[string]float64{"foo": 5712, "bar": 5902, "baz": 79}, key); err != nil {
		t.Fatalf("failed to set document ranks: %s", err)
	}
	if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), api.RepoName("baz"), map[string]float64{"foo": 86, "bar": 89, "baz": 9, "bonk": 918, "quux": 0}, key); err != nil {
		t.Fatalf("failed to set document ranks: %s", err)
	}

	logmean, err := store.GetReferenceCountStatistics(ctx)
	if err != nil {
		t.Fatalf("unexpected error getting reference count statistics: %s", err)
	}
	if expected := 7.8181; !cmpFloat(logmean, expected) {
		t.Errorf("unexpected logmean. want=%.5f have=%.5f", expected, logmean)
	}
}

func TestLastUpdatedAt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, db)

	key := rankingshared.NewDerivativeGraphKeyKey(mockRankingGraphKey, "", 123)

	if _, err := db.ExecContext(ctx, `
		INSERT INTO codeintel_ranking_progress(graph_key, max_definition_id, max_reference_id, max_path_id, mappers_started_at, reducer_completed_at)
		VALUES
			($1,  1000, 1000, 1000, NOW(), NOW())
	`,
		key,
	); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}

	idFoo := api.RepoID(1)
	idBar := api.RepoID(2)
	if _, err := db.ExecContext(ctx, `INSERT INTO repo (id, name) VALUES (1, 'foo'), (2, 'bar'), (3, 'baz')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}
	if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), "foo", nil, key); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	if err := setDocumentRanks(ctx, basestore.NewWithHandle(db.Handle()), "bar", nil, key); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}

	pairs, err := store.LastUpdatedAt(ctx, []api.RepoID{idFoo, idBar})
	if err != nil {
		t.Fatalf("unexpected error getting repo last update times: %s", err)
	}

	fooUpdatedAt, ok := pairs[idFoo]
	if !ok {
		t.Fatalf("expected 'foo' in result: %v", pairs)
	}
	barUpdatedAt, ok := pairs[idBar]
	if !ok {
		t.Fatalf("expected 'bar' in result: %v", pairs)
	}
	if _, ok := pairs[999]; ok {
		t.Fatalf("unexpected 'bonk' in result: %v", pairs)
	}

	if !fooUpdatedAt.Before(barUpdatedAt) {
		t.Errorf("unexpected timestamp ordering: %v and %v", fooUpdatedAt, barUpdatedAt)
	}
}

//
//

const epsilon = 0.0001

func cmpFloat(x, y float64) bool {
	return math.Abs(x-y) < epsilon
}
