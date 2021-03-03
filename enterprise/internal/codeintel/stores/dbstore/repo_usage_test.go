package dbstore

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func TestRepoUsageStatistics(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	store := testStore()

	insertEvent := func(url, name string, count int) {
		query := sqlf.Sprintf(`
			INSERT INTO event_logs (user_id, anonymous_user_id, source, argument, version, timestamp, name, url)
			VALUES (1, '', 'test', '{}', 'dev', NOW(), %s, %s)
		`, name, url)

		for i := 0; i < count; i++ {
			if _, err := dbconn.Global.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
				t.Fatalf("unexpected error inserting event record: %s", err)
			}
		}
	}

	for _, data := range []struct {
		URL              string
		NumSearchEvents  int
		NumPreciseEvents int
	}{
		{"http://localhost:3080/github.com/foo/baz/-/remainder_of_path", 10, 10},
		{"https://sourcegraph.com/github.com/foo/bar/-/remainder_of_path", 25, 20},
		{"http://localhost:3080/gitlab.com/bar/baz/-/remainder_of_path", 15, 30},
		{"https://sourcegraph.com/github.com/bar/bonk/-/remainder_of_path", 5, 40},
		{"http://srcgraph.org/github.com/bonk/quux/-/remainder_of_path", 4, 50},
		{"https://sourcegraph.com/github.com/baz/honk/-/remainder_of_path", 10, 60},  // deleted repo
		{"https://sourcegraph.com/github.com/bonk/honk/-/remainder_of_path", 10, 60}, // no such repo
	} {
		insertEvent(data.URL, "codeintel.searchHover", data.NumSearchEvents)
		insertEvent(data.URL, "codeintel.lsifHover", data.NumPreciseEvents)
	}

	repos := []string{
		"github.com/foo/baz",
		"github.com/foo/bar",
		"gitlab.com/bar/baz",
		"github.com/bar/bonk",
		"github.com/bonk/quux",
	}
	for i, name := range repos {
		query := sqlf.Sprintf(`INSERT INTO repo (id, name, uri) VALUES (%s, %s, %s)`, i+1, name, name)

		if _, err := dbconn.Global.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
			t.Fatalf("unexpected error inserting repo: %s", err)
		}
	}

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name, uri, deleted_at) VALUES (%s, %s, %s, %s)`,
		len(repos)+1,
		"DELETED-github.com/baz/honk",
		"github.com/baz/honk",
		time.Now(),
	)
	if _, err := dbconn.Global.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	stats, err := store.RepoUsageStatistics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error getting repo counts: %s", err)
	}

	expectedStatistics := []RepoUsageStatistics{
		{RepositoryID: 2, SearchCount: 25, PreciseCount: 20},
		{RepositoryID: 3, SearchCount: 15, PreciseCount: 30},
		{RepositoryID: 1, SearchCount: 10, PreciseCount: 10},
		{RepositoryID: 4, SearchCount: 5, PreciseCount: 40},
		{RepositoryID: 5, SearchCount: 4, PreciseCount: 50},
	}
	if diff := cmp.Diff(expectedStatistics, stats); diff != "" {
		t.Errorf("unexpected repo counts (-want +got):\n%s", diff)
	}
}
