package pgsql

import (
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/dbutil"
)

func init() {
	Schema.Map.AddTableWithName(hit{}, "repo_hit").SetKeys(false)
	Schema.CreateSQL = append(Schema.CreateSQL,
		`CREATE INDEX repo_hit_repo ON repo_hit(repo);`,
		`CREATE INDEX repo_hit_repo_at ON repo_hit(repo,at);`,
	)
}

// hit represents a hit to a repository counter.
type hit struct {
	Repo string // URI of repository
	At   time.Time
}

// RepoCounters is a DB-backed implementation of the Repos store.
type RepoCounters struct{}

var _ store.RepoCounters = (*RepoCounters)(nil)

func (s *RepoCounters) RecordHit(ctx context.Context, repo string) error {
	return dbh(ctx).Insert(&hit{Repo: repo, At: time.Now().In(time.UTC)})
}

func (s *RepoCounters) CountHits(ctx context.Context, repo string, since time.Time) (int, error) {
	sql := `SELECT COUNT(*) FROM "repo_hit" WHERE repo=$1`
	args := []interface{}{repo}
	if !since.IsZero() {
		sql += ` AND "at" > $2`
		args = append(args, since)
	}
	n, err := dbutil.SelectInt(dbh(ctx), sql, args...)
	return int(n), err
}
