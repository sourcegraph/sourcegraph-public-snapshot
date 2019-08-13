package db

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func Test_defaultRepos_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	tcs := []struct {
		name  string
		repos []*types.Repo
	}{
		{
			name:  "empty case",
			repos: nil,
		},
		{
			name: "one repo",
			repos: []*types.Repo{
				{
					ID:   api.RepoID(0),
					Name: "github.com/foo/bar",
				},
			},
		},
		{
			name: "a few repos",
			repos: []*types.Repo{
				{
					ID:   api.RepoID(0),
					Name: "github.com/foo/bar",
				},
				{
					ID:   api.RepoID(1),
					Name: "github.com/baz/qux",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := dbtesting.TestContext(t)
			for _, r := range tc.repos {
				if _, err := dbconn.Global.ExecContext(ctx, `INSERT INTO repo(id, name) VALUES ($1, $2)`, r.ID, r.Name); err != nil {
					t.Fatal(err)
				}
				if _, err := dbconn.Global.ExecContext(ctx, `INSERT INTO default_repos(repo_id) VALUES ($1)`, r.ID); err != nil {
					t.Fatal(err)
				}
			}

			repos, err := DefaultRepos.List(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(repos, tc.repos) {
				t.Errorf("repos = %v, want %v", repos, tc.repos)
			}
		})
	}
}

func BenchmarkDefaultRepos_List_Empty(b *testing.B) {
	ctx := dbtesting.TestContext(b)
	select {
	case <-ctx.Done():
		b.Fatal("context already canceled")
	default:
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, err := DefaultRepos.List(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
