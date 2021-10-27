package database

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"

	"github.com/keegancsmith/sqlf"
)

func TestSubRepoPermsInsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	s := SubRepoPerms(db, clock)
	prepareSubRepo(ctx, t, s)

	userID := int32(1)
	repoID := int32(1)
	rules := Rules{
		PathIncludes: []string{"/src/foo/*"},
		PathExcludes: []string{"/src/bar/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, rules); err != nil {
		t.Fatal(err)
	}

	have, err := s.GetRules(ctx, userID, repoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&rules, have); diff != "" {
		t.Fatal(diff)
	}
}

func TestSubRepoPermsUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	db := dbtest.NewDB(t)

	ctx := context.Background()
	s := SubRepoPerms(db, clock)
	prepareSubRepo(ctx, t, s)

	userID := int32(1)
	repoID := int32(1)
	rules := Rules{
		PathIncludes: []string{"/src/foo/*"},
		PathExcludes: []string{"/src/bar/*"},
	}
	// Insert initial data
	if err := s.Upsert(ctx, userID, repoID, rules); err != nil {
		t.Fatal(err)
	}

	// Upsert to change rules
	rules = Rules{
		PathIncludes: []string{"/src/foo_upsert/*"},
		PathExcludes: []string{"/src/bar_upsert/*"},
	}
	if err := s.Upsert(ctx, userID, repoID, rules); err != nil {
		t.Fatal(err)
	}

	have, err := s.GetRules(ctx, userID, repoID)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(&rules, have); diff != "" {
		t.Fatal(diff)
	}
}

func prepareSubRepo(ctx context.Context, t *testing.T, s *SubRepoPermsStore) {
	t.Helper()

	// Prepare data
	qs := []*sqlf.Query{
		// ID=1, with newer code host connection sync
		sqlf.Sprintf(`INSERT INTO users(username) VALUES ('alice')`),
		sqlf.Sprintf(`INSERT INTO external_services(id, display_name, kind, config, namespace_user_id, last_sync_at) VALUES(1, 'GitHub #1', 'GITHUB', '{}', 1, NOW() + INTERVAL '10min')`),
		sqlf.Sprintf(`INSERT INTO repo(id, name) VALUES(1, 'github.com/foo/bar')`),
	}
	for _, q := range qs {
		if err := s.Exec(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
}
