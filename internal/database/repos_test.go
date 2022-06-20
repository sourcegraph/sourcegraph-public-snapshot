package database

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
)

func TestParseIncludePattern(t *testing.T) {
	tests := map[string]struct {
		exact  []string
		like   []string
		regexp string

		pattern []*sqlf.Query // only tested if non-nil
	}{
		`^$`:              {exact: []string{""}},
		`(^$)`:            {exact: []string{""}},
		`((^$))`:          {exact: []string{""}},
		`^((^$))$`:        {exact: []string{""}},
		`^()$`:            {exact: []string{""}},
		`^(())$`:          {exact: []string{""}},
		`^a$`:             {exact: []string{"a"}},
		`(^a$)`:           {exact: []string{"a"}},
		`((^a$))`:         {exact: []string{"a"}},
		`^((^a$))$`:       {exact: []string{"a"}},
		`^(a)$`:           {exact: []string{"a"}},
		`^((a))$`:         {exact: []string{"a"}},
		`^a|b$`:           {like: []string{"a%", "%b"}}, // "|" has higher precedence than "^" or "$"
		`^(a)|(b)$`:       {like: []string{"a%", "%b"}}, // "|" has higher precedence than "^" or "$"
		`^(a|b)$`:         {exact: []string{"a", "b"}},
		`(^a$)|(^b$)`:     {exact: []string{"a", "b"}},
		`((^a$)|(^b$))`:   {exact: []string{"a", "b"}},
		`^((^a$)|(^b$))$`: {exact: []string{"a", "b"}},
		`^((a)|(b))$`:     {exact: []string{"a", "b"}},
		`abc`:             {like: []string{"%abc%"}},
		`a|b`:             {like: []string{"%a%", "%b%"}},
		`^a(b|c)$`:        {exact: []string{"ab", "ac"}},

		`^github\.com/foo/bar`: {like: []string{"github.com/foo/bar%"}},

		`github.com`:  {regexp: `github.com`},
		`github\.com`: {like: []string{`%github.com%`}},

		// https://github.com/sourcegraph/sourcegraph/issues/9146
		`github.com/.*/ini$`:      {regexp: `github.com/.*/ini$`},
		`github\.com/.*/ini$`:     {regexp: `github\.com/.*/ini$`},
		`github\.com/go-ini/ini$`: {like: []string{`%github.com/go-ini/ini`}},

		// https://github.com/sourcegraph/sourcegraph/issues/4166
		`golang/oauth.*`:                    {like: []string{"%golang/oauth%"}},
		`^golang/oauth.*`:                   {like: []string{"golang/oauth%"}},
		`golang/(oauth.*|bla)`:              {like: []string{"%golang/oauth%", "%golang/bla%"}},
		`golang/(oauth|bla)`:                {like: []string{"%golang/oauth%", "%golang/bla%"}},
		`^github.com/(golang|go-.*)/oauth$`: {regexp: `^github.com/(golang|go-.*)/oauth$`},
		`^github.com/(go.*lang|go)/oauth$`:  {regexp: `^github.com/(go.*lang|go)/oauth$`},

		// https://github.com/sourcegraph/sourcegraph/issues/20389
		`^github\.com/sourcegraph/(sourcegraph-atom|sourcegraph)$`: {
			exact: []string{"github.com/sourcegraph/sourcegraph", "github.com/sourcegraph/sourcegraph-atom"},
		},

		`(^github\.com/Microsoft/vscode$)|(^github\.com/sourcegraph/go-langserver$)`: {
			exact: []string{"github.com/Microsoft/vscode", "github.com/sourcegraph/go-langserver"},
		},

		// Avoid DoS when there are too many possible matches to enumerate.
		`^(a|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`: {regexp: `^(a|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`},
		`^[0-a]$`:                               {regexp: `^[0-a]$`},
		`sourcegraph|^github\.com/foo/bar$`: {
			like:  []string{`%sourcegraph%`},
			exact: []string{"github.com/foo/bar"},
			pattern: []*sqlf.Query{
				sqlf.Sprintf(`(name = ANY (%s) OR lower(name) LIKE %s)`, "%!s(*pq.StringArray=&[github.com/foo/bar])", "%sourcegraph%"),
			},
		},

		// Recognize perl character class shorthand syntax.
		`\s`: {regexp: `\s`},
	}
	for pattern, want := range tests {
		exact, like, regexp, err := parseIncludePattern(pattern)
		if err != nil {
			t.Fatal(pattern, err)
		}
		if !reflect.DeepEqual(exact, want.exact) {
			t.Errorf("got exact %q, want %q for %s", exact, want.exact, pattern)
		}
		if !reflect.DeepEqual(like, want.like) {
			t.Errorf("got like %q, want %q for %s", like, want.like, pattern)
		}
		if regexp != want.regexp {
			t.Errorf("got regexp %q, want %q for %s", regexp, want.regexp, pattern)
		}
		if qs, err := parsePattern(pattern, false); err != nil {
			t.Fatal(pattern, err)
		} else {
			if testing.Verbose() {
				q := sqlf.Join(qs, "AND")
				t.Log(pattern, q.Query(sqlf.PostgresBindVar), q.Args())
			}

			if want.pattern != nil {
				want := queriesToString(want.pattern)
				q := queriesToString(qs)
				if want != q {
					t.Errorf("got pattern %q, want %q for %s", q, want, pattern)
				}
			}
		}
	}
}

func queriesToString(qs []*sqlf.Query) string {
	q := sqlf.Join(qs, "AND")
	return fmt.Sprintf("%s %s", q.Query(sqlf.PostgresBindVar), q.Args())
}

func TestRepos_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	if err := upsertRepo(ctx, db, InsertRepoOp{Name: "myrepo", Description: "", Fork: false}); err != nil {
		t.Fatal(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	t.Run("order and limit options are ignored", func(t *testing.T) {
		opts := ReposListOptions{
			OrderBy:     []RepoListSort{{Field: RepoListID}},
			LimitOffset: &LimitOffset{Limit: 1},
		}
		if count, err := db.Repos().Count(ctx, opts); err != nil {
			t.Fatal(err)
		} else if want := 1; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	})

	repos, err := db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Repos().Delete(ctx, repos[0].ID); err != nil {
		t.Fatal(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestRepos_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	if err := upsertRepo(ctx, db, InsertRepoOp{Name: "myrepo", Description: "", Fork: false}); err != nil {
		t.Fatal(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	repos, err := db.Repos().List(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Repos().Delete(ctx, repos[0].ID); err != nil {
		t.Fatal(err)
	}

	if count, err := db.Repos().Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}
}

func TestRepos_Upsert(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	if _, err := db.Repos().GetByName(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fatal("myrepo already present")
		} else {
			t.Fatal(err)
		}
	}

	if err := upsertRepo(ctx, db, InsertRepoOp{Name: "myrepo", Description: "", Fork: false}); err != nil {
		t.Fatal(err)
	}

	rp, err := db.Repos().GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if rp.Name != "myrepo" {
		t.Fatalf("rp.Name: %s != %s", rp.Name, "myrepo")
	}

	ext := api.ExternalRepoSpec{
		ID:          "ext:id",
		ServiceType: "test",
		ServiceID:   "ext:test",
	}

	if err := upsertRepo(ctx, db, InsertRepoOp{Name: "myrepo", Description: "asdfasdf", Fork: false, ExternalRepo: ext}); err != nil {
		t.Fatal(err)
	}

	rp, err = db.Repos().GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	if rp.Name != "myrepo" {
		t.Fatalf("rp.Name: %s != %s", rp.Name, "myrepo")
	}
	if rp.Description != "asdfasdf" {
		t.Fatalf("rp.Name: %q != %q", rp.Description, "asdfasdf")
	}
	if !reflect.DeepEqual(rp.ExternalRepo, ext) {
		t.Fatalf("rp.ExternalRepo: %s != %s", rp.ExternalRepo, ext)
	}

	// Rename. Detected by external repo
	if err := upsertRepo(ctx, db, InsertRepoOp{Name: "myrepo/renamed", Description: "asdfasdf", Fork: false, ExternalRepo: ext}); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Repos().GetByName(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fatal("myrepo should be renamed, but still present as myrepo")
		} else {
			t.Fatal(err)
		}
	}

	rp, err = db.Repos().GetByName(ctx, "myrepo/renamed")
	if err != nil {
		t.Fatal(err)
	}
	if rp.Name != "myrepo/renamed" {
		t.Fatalf("rp.Name: %s != %s", rp.Name, "myrepo/renamed")
	}
	if rp.Description != "asdfasdf" {
		t.Fatalf("rp.Name: %q != %q", rp.Description, "asdfasdf")
	}
	if !reflect.DeepEqual(rp.ExternalRepo, ext) {
		t.Fatalf("rp.ExternalRepo: %s != %s", rp.ExternalRepo, ext)
	}
}

func TestRepos_UpsertForkAndArchivedFields(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	i := 0
	for _, fork := range []bool{true, false} {
		for _, archived := range []bool{true, false} {
			i++
			name := api.RepoName(fmt.Sprintf("myrepo-%d", i))

			if err := upsertRepo(ctx, db, InsertRepoOp{Name: name, Fork: fork, Archived: archived}); err != nil {
				t.Fatal(err)
			}

			rp, err := db.Repos().GetByName(ctx, name)
			if err != nil {
				t.Fatal(err)
			}

			if rp.Fork != fork {
				t.Fatalf("rp.Fork: %v != %v", rp.Fork, fork)
			}
			if rp.Archived != archived {
				t.Fatalf("rp.Archived: %v != %v", rp.Archived, archived)
			}
		}
	}
}

func hasNoID(r *types.Repo) bool {
	return r.ID == 0
}

func TestRepos_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := NewDB(dbtest.NewDB(t))
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	svcs := typestest.MakeExternalServices()
	if err := db.ExternalServices().Upsert(ctx, svcs...); err != nil {
		t.Fatalf("Upsert error: %s", err)
	}

	msvcs := typestest.ExternalServicesToMap(svcs)

	repo1 := typestest.MakeGithubRepo(msvcs[extsvc.KindGitHub], msvcs[extsvc.KindBitbucketServer])
	repo2 := typestest.MakeGitlabRepo(msvcs[extsvc.KindGitLab])

	t.Run("no repos should not fail", func(t *testing.T) {
		if err := db.Repos().Create(ctx); err != nil {
			t.Fatalf("Create error: %s", err)
		}
	})

	t.Run("many repos", func(t *testing.T) {
		want := typestest.GenerateRepos(7, repo1, repo2)

		if err := db.Repos().Create(ctx, want...); err != nil {
			t.Fatalf("Create error: %s", err)
		}

		sort.Sort(want)

		if noID := want.Filter(hasNoID); len(noID) > 0 {
			t.Fatalf("Create didn't assign an ID to all repos: %v", noID.Names())
		}

		have, err := db.Repos().List(ctx, ReposListOptions{})
		if err != nil {
			t.Fatalf("List error: %s", err)
		}

		if diff := cmp.Diff(have, []*types.Repo(want), cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("List:\n%s", diff)
		}
	})
}

func TestListIndexableRepos(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))

	reposToAdd := []types.Repo{
		{
			ID:    api.RepoID(1),
			Name:  "github.com/foo/bar1",
			Stars: 20,
		},
		{
			ID:    api.RepoID(2),
			Name:  "github.com/baz/bar2",
			Stars: 30,
		},
		{
			ID:      api.RepoID(3),
			Name:    "github.com/foo/bar3",
			Private: true,
			Stars:   0, // Will still be returned because it gets added by a user.
		},
		{
			ID:    api.RepoID(4),
			Name:  "github.com/foo/bar4",
			Stars: 1, // Not enough stars
		},
		{
			ID:    api.RepoID(5),
			Name:  "github.com/foo/bar5",
			Stars: 400,
			Blocked: &types.RepoBlock{
				At:     time.Now().UTC().Unix(),
				Reason: "Failed to index too many times.",
			},
		},
	}

	ctx := context.Background()
	// Add an external service
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO external_services(id, kind, display_name, config, cloud_default) VALUES (1, 'github', 'github', '{}', true);`,
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users(id, username) VALUES (1, 'bob')`); err != nil {
		t.Fatal(err)
	}
	for _, r := range reposToAdd {
		blocked, err := json.Marshal(r.Blocked)
		if err != nil {
			t.Fatal(err)
		}
		_, err = db.ExecContext(ctx,
			`INSERT INTO repo(id, name, stars, private, blocked) VALUES ($1, $2, $3, $4, NULLIF($5, 'null'::jsonb))`,
			r.ID, r.Name, r.Stars, r.Private, blocked,
		)
		if err != nil {
			t.Fatal(err)
		}

		if r.Private {
			if _, err := db.ExecContext(ctx, `INSERT INTO external_service_repos VALUES (1, $1, $2, 1);`, r.ID, r.Name); err != nil {
				t.Fatal(err)
			}
		}

		cloned := int(r.ID) > 1
		cloneStatus := types.CloneStatusCloned
		if !cloned {
			cloneStatus = types.CloneStatusNotCloned
		}
		if _, err := db.ExecContext(ctx, `UPDATE gitserver_repos SET clone_status = $2, shard_id = 'test' WHERE repo_id = $1;`, r.ID, cloneStatus); err != nil {
			t.Fatal(err)
		}
	}

	for _, tc := range []struct {
		name string
		opts ListIndexableReposOptions
		want []api.RepoID
	}{
		{
			name: "no opts",
			want: []api.RepoID{2, 1}, // No private repos returned by default
		},
		{
			name: "only uncloned",
			opts: ListIndexableReposOptions{OnlyUncloned: true},
			want: []api.RepoID{1},
		},
		{
			name: "include private",
			opts: ListIndexableReposOptions{IncludePrivate: true},
			want: []api.RepoID{2, 1, 3},
		},
		{
			name: "limit 1",
			opts: ListIndexableReposOptions{LimitOffset: &LimitOffset{Limit: 1}},
			want: []api.RepoID{2},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repos, err := db.Repos().ListIndexableRepos(ctx, tc.opts)
			if err != nil {
				t.Fatal(err)
			}

			have := make([]api.RepoID, 0, len(repos))
			for _, r := range repos {
				have = append(have, r.ID)
			}

			if diff := cmp.Diff(tc.want, have, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("mismatch (-want +have):\n%s", diff)
			}
		})
	}
}

func TestRepoStore_Metadata(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := NewDB(dbtest.NewDB(t))

	ctx := context.Background()

	repos := []*types.Repo{
		{
			ID:          1,
			Name:        "foo",
			Description: "foo 1",
			Fork:        false,
			Archived:    false,
			Private:     false,
			Stars:       10,
			URI:         "foo-uri",
			Sources:     map[string]*types.SourceInfo{},
		},
		{
			ID:          2,
			Name:        "bar",
			Description: "bar 2",
			Fork:        true,
			Archived:    true,
			Private:     true,
			Stars:       20,
			URI:         "bar-uri",
			Sources:     map[string]*types.SourceInfo{},
		},
	}

	r := db.Repos()
	require.NoError(t, r.Create(ctx, repos...))

	d1 := time.Unix(1627945150, 0).UTC()
	d2 := time.Unix(1628945150, 0).UTC()
	gitserverRepos := []*types.GitserverRepo{
		{
			RepoID:      1,
			LastFetched: d1,
			ShardID:     "abc",
		},
		{
			RepoID:      2,
			LastFetched: d2,
			ShardID:     "abc",
		},
	}

	gr := db.GitserverRepos()
	require.NoError(t, gr.Upsert(ctx, gitserverRepos...))

	expected := []*types.SearchedRepo{
		{
			ID:          1,
			Name:        "foo",
			Description: "foo 1",
			Fork:        false,
			Archived:    false,
			Private:     false,
			Stars:       10,
			LastFetched: &d1,
		},
		{
			ID:          2,
			Name:        "bar",
			Description: "bar 2",
			Fork:        true,
			Archived:    true,
			Private:     true,
			Stars:       20,
			LastFetched: &d2,
		},
	}

	md, err := r.Metadata(ctx, 1, 2)
	require.NoError(t, err)
	require.ElementsMatch(t, expected, md)
}
