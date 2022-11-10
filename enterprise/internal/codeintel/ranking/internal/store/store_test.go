package store

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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

func TestDocumentRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)
	repoName := api.RepoName("foo")

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name, stars) VALUES ('foo', 1000)`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	if err := store.SetDocumentRanks(ctx, repoName, 0.25, map[string]float64{
		"cmd/main.go":        2, // no longer referenced
		"internal/secret.go": 3,
		"internal/util.go":   4,
		"README.md":          5, // no longer referenced
	}); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	if err := store.SetDocumentRanks(ctx, repoName, 0.25, map[string]float64{
		"cmd/args.go":        8, // new
		"internal/secret.go": 7, // edited
		"internal/util.go":   6, // edited
	}); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}

	ranks, _, err := store.GetDocumentRanks(ctx, repoName)
	if err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	expectedRanks := map[string][2]float64{
		"cmd/args.go":        {0.25, 8},
		"internal/secret.go": {0.25, 7},
		"internal/util.go":   {0.25, 6},
	}
	if diff := cmp.Diff(expectedRanks, ranks); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}
}

func TestBulkSetAndMergeDocumentRanks(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	for i := 1; i <= 15; i++ {
		if _, err := db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO repo (name) VALUES ('r%d')`, i)); err != nil {
			t.Fatalf("failed to insert repos: %s", err)
		}
	}

	{
		for i := 0; i < 10; i++ {
			fi := float64(i)

			if err := store.BulkSetDocumentRanks(ctx, "old-scores", fmt.Sprintf("f-%02d.csv", i+1), 0.25, map[api.RepoName]map[string]float64{
				api.RepoName(fmt.Sprintf("r%d", i+1)): {fmt.Sprintf("baz-%d.go", i): fi + 5},
				api.RepoName(fmt.Sprintf("r%d", i+2)): {fmt.Sprintf("baz-%d.go", i): fi + 5},
				api.RepoName(fmt.Sprintf("r%d", i+3)): {fmt.Sprintf("baz-%d.go", i): fi + 5},
			}); err != nil {
				t.Fatalf("unexpected error setting document ranks: %s", err)
			}
		}

		// Create scores that will need to be overwritten with a newer graph key
		if _, _, err := store.MergeDocumentRanks(ctx, "old-scores", 500); err != nil {
			t.Fatalf("Unexpected error merging document ranks: %s", err)
		}
	}

	for i := 0; i < 10; i++ {
		fi := float64(i)

		if err := store.BulkSetDocumentRanks(ctx, "test", fmt.Sprintf("f-%02d.csv", i+1), 0.25, map[api.RepoName]map[string]float64{
			api.RepoName(fmt.Sprintf("r%d", i+1)): {fmt.Sprintf("foo-%d.go", i): fi + 2, fmt.Sprintf("bar-%d.go", i): fi + 4},
			api.RepoName(fmt.Sprintf("r%d", i+2)): {fmt.Sprintf("foo-%d.go", i): fi + 2, fmt.Sprintf("bar-%d.go", i): fi + 4},
			api.RepoName(fmt.Sprintf("r%d", i+3)): {fmt.Sprintf("foo-%d.go", i): fi + 2, fmt.Sprintf("bar-%d.go", i): fi + 4},
		}); err != nil {
			t.Fatalf("unexpected error setting document ranks: %s", err)
		}
	}

	inputFilenames := []string{
		"f-07.csv", "f-08.csv", "f-09.csv", "f-10.csv", // known
		"f-11.csv", "f-12.csv", "f-13.csv", "f-14.csv", // unknown
	}
	filenames, err := store.HasInputFilename(ctx, "test", inputFilenames)
	if err != nil {
		t.Fatalf("unexpected error checking if filename inputs exist: %s", err)
	}
	sort.Strings(filenames)
	expectedFilenames := []string{
		"f-07.csv",
		"f-08.csv",
		"f-09.csv",
		"f-10.csv",
	}
	if diff := cmp.Diff(expectedFilenames, filenames); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}

	if numRepositoriesUpdated, numInputsProcessed, err := store.MergeDocumentRanks(ctx, "test", 500); err != nil {
		t.Fatalf("Unexpected error merging document ranks: %s", err)
	} else if expected := 12; numRepositoriesUpdated != expected {
		t.Fatalf("unexpected numRepositoriesUpdated. want=%d have=%d", expected, numRepositoriesUpdated)
	} else if expected := 30; numInputsProcessed != expected {
		t.Fatalf("unexpected numInputsProcessed. want=%d have=%d", expected, numInputsProcessed)
	}

	allRanks := map[string]map[string][2]float64{}
	for i := 0; i < 12; i++ {
		repoName := fmt.Sprintf("r%d", i+1)
		ranks, _, err := store.GetDocumentRanks(ctx, api.RepoName(repoName))
		if err != nil {
			t.Fatalf("unexpected error getting ranks for repo %s: %s", repoName, err)
		}

		allRanks[repoName] = ranks
	}

	expectedRanks := map[string]map[string][2]float64{
		"r1": {
			"foo-0.go": {0.25, 2}, "bar-0.go": {0.25, 4},
		},
		"r2": {
			"foo-0.go": {0.25, 2}, "bar-0.go": {0.25, 4},
			"foo-1.go": {0.25, 3}, "bar-1.go": {0.25, 5},
		},
		"r3": {
			"foo-0.go": {0.25, 2}, "bar-0.go": {0.25, 4},
			"foo-1.go": {0.25, 3}, "bar-1.go": {0.25, 5},
			"foo-2.go": {0.25, 4}, "bar-2.go": {0.25, 6},
		},
		"r4": {
			"foo-1.go": {0.25, 3}, "bar-1.go": {0.25, 5},
			"foo-2.go": {0.25, 4}, "bar-2.go": {0.25, 6},
			"foo-3.go": {0.25, 5}, "bar-3.go": {0.25, 7},
		},
		"r5": {
			"foo-2.go": {0.25, 4}, "bar-2.go": {0.25, 6},
			"foo-3.go": {0.25, 5}, "bar-3.go": {0.25, 7},
			"foo-4.go": {0.25, 6}, "bar-4.go": {0.25, 8},
		},
		"r6": {
			"foo-3.go": {0.25, 5}, "bar-3.go": {0.25, 7},
			"foo-4.go": {0.25, 6}, "bar-4.go": {0.25, 8},
			"foo-5.go": {0.25, 7}, "bar-5.go": {0.25, 9},
		},
		"r7": {
			"foo-4.go": {0.25, 6}, "bar-4.go": {0.25, 8},
			"foo-5.go": {0.25, 7}, "bar-5.go": {0.25, 9},
			"foo-6.go": {0.25, 8}, "bar-6.go": {0.25, 10},
		},
		"r8": {
			"foo-5.go": {0.25, 7}, "bar-5.go": {0.25, 9},
			"foo-6.go": {0.25, 8}, "bar-6.go": {0.25, 10},
			"foo-7.go": {0.25, 9}, "bar-7.go": {0.25, 11},
		},
		"r9": {
			"foo-6.go": {0.25, 8}, "bar-6.go": {0.25, 10},
			"foo-7.go": {0.25, 9}, "bar-7.go": {0.25, 11},
			"foo-8.go": {0.25, 10}, "bar-8.go": {0.25, 12},
		},
		"r10": {
			"foo-7.go": {0.25, 9}, "bar-7.go": {0.25, 11},
			"foo-8.go": {0.25, 10}, "bar-8.go": {0.25, 12},
			"foo-9.go": {0.25, 11}, "bar-9.go": {0.25, 13},
		},
		"r11": {
			"foo-8.go": {0.25, 10}, "bar-8.go": {0.25, 12},
			"foo-9.go": {0.25, 11}, "bar-9.go": {0.25, 13},
		},
		"r12": {
			"foo-9.go": {0.25, 11}, "bar-9.go": {0.25, 13},
		},
	}
	if diff := cmp.Diff(expectedRanks, allRanks); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}
}

func TestLastUpdatedAt(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	idFoo := api.RepoID(1)
	idBar := api.RepoID(2)
	if _, err := db.ExecContext(ctx, `INSERT INTO repo (id, name) VALUES (1, 'foo'), (2, 'bar'), (3, 'baz')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}
	if err := store.SetDocumentRanks(ctx, "foo", 0.25, nil); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	if err := store.SetDocumentRanks(ctx, "bar", 0.25, nil); err != nil {
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

func TestUpdatedAfter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Background()
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := New(db, &observation.TestContext)

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('foo'), ('bar'), ('baz')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}
	if err := store.SetDocumentRanks(ctx, "foo", 0.25, nil); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	if err := store.SetDocumentRanks(ctx, "bar", 0.25, nil); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}

	// past
	{
		repoNames, err := store.UpdatedAfter(ctx, time.Now().Add(-time.Hour*24))
		if err != nil {
			t.Fatalf("unexpected error getting updated repos: %s", err)
		}
		if diff := cmp.Diff([]api.RepoName{"bar", "foo"}, repoNames); diff != "" {
			t.Errorf("unexpected repository names (-want +got):\n%s", diff)
		}
	}

	// future
	{
		repoNames, err := store.UpdatedAfter(ctx, time.Now().Add(time.Hour*24))
		if err != nil {
			t.Fatalf("unexpected error getting updated repos: %s", err)
		}
		if len(repoNames) != 0 {
			t.Fatal("expected no repos")
		}
	}
}
