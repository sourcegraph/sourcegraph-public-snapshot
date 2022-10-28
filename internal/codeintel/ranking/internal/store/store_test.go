package store

import (
	"context"
	"fmt"
	"testing"

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

	if err := store.SetDocumentRanks(ctx, repoName, map[string][]float64{
		"cmd/main.go":        {1, 2, 3},
		"internal/secret.go": {2, 3, 4},
		"internal/util.go":   {3, 4, 5},
		"README.md":          {4, 5, 6},
	}); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	if err := store.SetDocumentRanks(ctx, repoName, map[string][]float64{
		"cmd/args.go":        {7, 8, 9}, // new
		"internal/secret.go": {6, 7, 8}, // edited
		"internal/util.go":   {5, 6, 7}, // edited
	}); err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}

	ranks, _, err := store.GetDocumentRanks(ctx, repoName)
	if err != nil {
		t.Fatalf("unexpected error setting document ranks: %s", err)
	}
	expectedRanks := map[string][]float64{
		"cmd/main.go":        {1, 2, 3},
		"README.md":          {4, 5, 6},
		"cmd/args.go":        {7, 8, 9},
		"internal/secret.go": {6, 7, 8},
		"internal/util.go":   {5, 6, 7},
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

	if _, err := db.ExecContext(ctx, `INSERT INTO repo (name) VALUES ('r1'), ('r2'), ('r3'), ('r4'), ('r5'), ('r6'), ('r7'), ('r8'), ('r9'), ('r10'), ('r11'), ('r12'), ('r13'), ('r14'), ('r15')`); err != nil {
		t.Fatalf("failed to insert repos: %s", err)
	}

	for i := 0; i < 10; i++ {
		fi := float64(i)

		if err := store.BulkSetDocumentRanks(ctx, "test", fmt.Sprintf("f-%02d.csv", i+1), map[api.RepoName]map[string][]float64{
			api.RepoName(fmt.Sprintf("r%d", i+1)): {fmt.Sprintf("foo-%d.go", i): {fi + 1, fi + 2, fi + 3}, fmt.Sprintf("bar-%d.go", i): {fi + 3, fi + 4, fi + 5}},
			api.RepoName(fmt.Sprintf("r%d", i+2)): {fmt.Sprintf("foo-%d.go", i): {fi + 1, fi + 2, fi + 3}, fmt.Sprintf("bar-%d.go", i): {fi + 3, fi + 4, fi + 5}},
			api.RepoName(fmt.Sprintf("r%d", i+3)): {fmt.Sprintf("foo-%d.go", i): {fi + 1, fi + 2, fi + 3}, fmt.Sprintf("bar-%d.go", i): {fi + 3, fi + 4, fi + 5}},
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

	allRanks := map[string]map[string][]float64{}
	for i := 0; i < 12; i++ {
		repoName := fmt.Sprintf("r%d", i+1)
		ranks, _, err := store.GetDocumentRanks(ctx, api.RepoName(repoName))
		if err != nil {
			t.Fatalf("unexpected error getting ranks for repo %s: %s", repoName, err)
		}

		allRanks[repoName] = ranks
	}

	expectedRanks := map[string]map[string][]float64{
		"r1": {
			"foo-0.go": {1, 2, 3}, "bar-0.go": {3, 4, 5},
		},
		"r2": {
			"foo-0.go": {1, 2, 3}, "bar-0.go": {3, 4, 5},
			"foo-1.go": {2, 3, 4}, "bar-1.go": {4, 5, 6},
		},
		"r3": {
			"foo-0.go": {1, 2, 3}, "bar-0.go": {3, 4, 5},
			"foo-1.go": {2, 3, 4}, "bar-1.go": {4, 5, 6},
			"foo-2.go": {3, 4, 5}, "bar-2.go": {5, 6, 7},
		},
		"r4": {
			"foo-1.go": {2, 3, 4}, "bar-1.go": {4, 5, 6},
			"foo-2.go": {3, 4, 5}, "bar-2.go": {5, 6, 7},
			"foo-3.go": {4, 5, 6}, "bar-3.go": {6, 7, 8},
		},
		"r5": {
			"foo-2.go": {3, 4, 5}, "bar-2.go": {5, 6, 7},
			"foo-3.go": {4, 5, 6}, "bar-3.go": {6, 7, 8},
			"foo-4.go": {5, 6, 7}, "bar-4.go": {7, 8, 9},
		},
		"r6": {
			"foo-3.go": {4, 5, 6}, "bar-3.go": {6, 7, 8},
			"foo-4.go": {5, 6, 7}, "bar-4.go": {7, 8, 9},
			"foo-5.go": {6, 7, 8}, "bar-5.go": {8, 9, 10},
		},
		"r7": {
			"foo-4.go": {5, 6, 7}, "bar-4.go": {7, 8, 9},
			"foo-5.go": {6, 7, 8}, "bar-5.go": {8, 9, 10},
			"foo-6.go": {7, 8, 9}, "bar-6.go": {9, 10, 11},
		},
		"r8": {
			"foo-5.go": {6, 7, 8}, "bar-5.go": {8, 9, 10},
			"foo-6.go": {7, 8, 9}, "bar-6.go": {9, 10, 11},
			"foo-7.go": {8, 9, 10}, "bar-7.go": {10, 11, 12},
		},
		"r9": {
			"foo-6.go": {7, 8, 9}, "bar-6.go": {9, 10, 11},
			"foo-7.go": {8, 9, 10}, "bar-7.go": {10, 11, 12},
			"foo-8.go": {9, 10, 11}, "bar-8.go": {11, 12, 13},
		},
		"r10": {
			"foo-7.go": {8, 9, 10}, "bar-7.go": {10, 11, 12},
			"foo-8.go": {9, 10, 11}, "bar-8.go": {11, 12, 13},
			"foo-9.go": {10, 11, 12}, "bar-9.go": {12, 13, 14},
		},
		"r11": {
			"foo-8.go": {9, 10, 11}, "bar-8.go": {11, 12, 13},
			"foo-9.go": {10, 11, 12}, "bar-9.go": {12, 13, 14},
		},
		"r12": {
			"foo-9.go": {10, 11, 12}, "bar-9.go": {12, 13, 14},
		},
	}
	if diff := cmp.Diff(expectedRanks, allRanks); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}
}
