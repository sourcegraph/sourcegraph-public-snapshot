package database

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		`^github\.com/sourcegraph/(sourcegraph-atom|sourcegraph)$`: {exact: []string{"github.com/sourcegraph/sourcegraph", "github.com/sourcegraph/sourcegraph-atom"}},

		`(^github\.com/Microsoft/vscode$)|(^github\.com/sourcegraph/go-langserver$)`: {exact: []string{"github.com/Microsoft/vscode", "github.com/sourcegraph/go-langserver"}},

		// Avoid DoS when there are too many possible matches to enumerate.
		`^(a|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`: {regexp: `^(a|b)(c|d)(e|f)(g|h)(i|j)(k|l)(m|n)$`},
		`^[0-a]$`:                               {regexp: `^[0-a]$`},
		`sourcegraph|^github\.com/foo/bar$`: {
			like:    []string{`%sourcegraph%`},
			exact:   []string{"github.com/foo/bar"},
			pattern: []*sqlf.Query{sqlf.Sprintf(`(name IN (%s) OR lower(name) LIKE %s)`, "github.com/foo/bar", "%sourcegraph%")},
		},
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
		if qs, err := parsePattern(pattern); err != nil {
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
	return fmt.Sprintf("%s %v", q.Query(sqlf.PostgresBindVar), q.Args())
}

func TestRepos_Count(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := dbtest.NewDB(t, "")
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	if count, err := Repos(db).Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 0; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	if err := Repos(db).Upsert(ctx, InsertRepoOp{Name: "myrepo", Description: "", Fork: false}); err != nil {
		t.Fatal(err)
	}

	if count, err := Repos(db).Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	t.Run("order and limit options are ignored", func(t *testing.T) {
		opts := ReposListOptions{
			OrderBy:     []RepoListSort{{Field: RepoListID}},
			LimitOffset: &LimitOffset{Limit: 1},
		}
		if count, err := Repos(db).Count(ctx, opts); err != nil {
			t.Fatal(err)
		} else if want := 1; count != want {
			t.Errorf("got %d, want %d", count, want)
		}
	})

	repos, err := Repos(db).List(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Repos(db).Delete(ctx, repos[0].ID); err != nil {
		t.Fatal(err)
	}

	if count, err := Repos(db).Count(ctx, ReposListOptions{}); err != nil {
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
	db := dbtest.NewDB(t, "")
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	if err := Repos(db).Upsert(ctx, InsertRepoOp{Name: "myrepo", Description: "", Fork: false}); err != nil {
		t.Fatal(err)
	}

	if count, err := Repos(db).Count(ctx, ReposListOptions{}); err != nil {
		t.Fatal(err)
	} else if want := 1; count != want {
		t.Errorf("got %d, want %d", count, want)
	}

	repos, err := Repos(db).List(ctx, ReposListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if err := Repos(db).Delete(ctx, repos[0].ID); err != nil {
		t.Fatal(err)
	}

	if count, err := Repos(db).Count(ctx, ReposListOptions{}); err != nil {
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
	db := dbtest.NewDB(t, "")
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	if _, err := Repos(db).GetByName(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fatal("myrepo already present")
		} else {
			t.Fatal(err)
		}
	}

	if err := Repos(db).Upsert(ctx, InsertRepoOp{Name: "myrepo", Description: "", Fork: false}); err != nil {
		t.Fatal(err)
	}

	rp, err := Repos(db).GetByName(ctx, "myrepo")
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

	if err := Repos(db).Upsert(ctx, InsertRepoOp{Name: "myrepo", Description: "asdfasdf", Fork: false, ExternalRepo: ext}); err != nil {
		t.Fatal(err)
	}

	rp, err = Repos(db).GetByName(ctx, "myrepo")
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
	if err := Repos(db).Upsert(ctx, InsertRepoOp{Name: "myrepo/renamed", Description: "asdfasdf", Fork: false, ExternalRepo: ext}); err != nil {
		t.Fatal(err)
	}

	if _, err := Repos(db).GetByName(ctx, "myrepo"); !errcode.IsNotFound(err) {
		if err == nil {
			t.Fatal("myrepo should be renamed, but still present as myrepo")
		} else {
			t.Fatal(err)
		}
	}

	rp, err = Repos(db).GetByName(ctx, "myrepo/renamed")
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
	db := dbtest.NewDB(t, "")
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	i := 0
	for _, fork := range []bool{true, false} {
		for _, archived := range []bool{true, false} {
			i++
			name := api.RepoName(fmt.Sprintf("myrepo-%d", i))

			if err := Repos(db).Upsert(ctx, InsertRepoOp{Name: name, Fork: fork, Archived: archived}); err != nil {
				t.Fatal(err)
			}

			rp, err := Repos(db).GetByName(ctx, name)
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
	db := dbtest.NewDB(t, "")
	ctx := context.Background()
	ctx = actor.WithActor(ctx, &actor.Actor{UID: 1, Internal: true})

	svcs := types.MakeExternalServices()
	if err := ExternalServices(db).Upsert(ctx, svcs...); err != nil {
		t.Fatalf("Upsert error: %s", err)
	}

	msvcs := types.ExternalServicesToMap(svcs)

	repo1 := types.MakeGithubRepo(msvcs[extsvc.KindGitHub], msvcs[extsvc.KindBitbucketServer])
	repo2 := types.MakeGitlabRepo(msvcs[extsvc.KindGitLab])

	t.Run("no repos should not fail", func(t *testing.T) {
		if err := Repos(db).Create(ctx); err != nil {
			t.Fatalf("Create error: %s", err)
		}
	})

	t.Run("many repos", func(t *testing.T) {
		want := types.GenerateRepos(7, repo1, repo2)

		if err := Repos(db).Create(ctx, want...); err != nil {
			t.Fatalf("Create error: %s", err)
		}

		sort.Sort(want)

		if noID := want.Filter(hasNoID); len(noID) > 0 {
			t.Fatalf("Create didn't assign an ID to all repos: %v", noID.Names())
		}

		have, err := Repos(db).List(ctx, ReposListOptions{})
		if err != nil {
			t.Fatalf("List error: %s", err)
		}

		if diff := cmp.Diff(have, []*types.Repo(want), cmpopts.EquateEmpty()); diff != "" {
			t.Fatalf("List:\n%s", diff)
		}
	})
}

func TestListDefaultReposUncloned(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	db := dbtest.NewDB(t, "")

	reposToAdd := []types.RepoName{
		{
			ID:   api.RepoID(1),
			Name: "github.com/foo/bar1",
		},
		{
			ID:   api.RepoID(2),
			Name: "github.com/baz/bar2",
		},
		{
			ID:   api.RepoID(3),
			Name: "github.com/foo/bar3",
		},
	}

	ctx := context.Background()
	// Add an external service
	_, err := db.ExecContext(ctx, `INSERT INTO external_services(id, kind, display_name, config, cloud_default) VALUES (1, 'github', 'github', '{}', true);`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users(id, username) VALUES (1, 'bob')`); err != nil {
		t.Fatal(err)
	}
	for _, r := range reposToAdd {
		cloned := int(r.ID) > 1
		if _, err := db.ExecContext(ctx, `INSERT INTO repo(id, name, cloned) VALUES ($1, $2, $3)`, r.ID, r.Name, cloned); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO default_repos(repo_id) VALUES ($1)`, r.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO external_service_repos VALUES (1, $1, 'https://github.com/foo/bar13', 1);`, r.ID); err != nil {
			t.Fatal(err)
		}
		cloneStatus := types.CloneStatusCloned
		if !cloned {
			cloneStatus = types.CloneStatusNotCloned
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO gitserver_repos(repo_id, clone_status, shard_id) VALUES ($1, $2, 'test');`, r.ID, cloneStatus); err != nil {
			t.Fatal(err)
		}
	}

	repos, err := Repos(db).ListDefaultRepos(ctx, ListDefaultReposOptions{
		OnlyUncloned: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	sort.Sort(types.RepoNames(repos))
	sort.Sort(types.RepoNames(reposToAdd))
	if diff := cmp.Diff(reposToAdd[:1], repos, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
