package db

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"reflect"
	"testing"
)

func Test_defaultRepos_List(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	for N := 0; N <= 2; N++ {
		t.Run(fmt.Sprintf("with %d repo ids", N), func(t *testing.T) {
			ctx := dbtesting.TestContext(t)
			var wantRepos []*types.Repo
			for n := 0; n < N; n++ {
				name := fmt.Sprintf("repo-%d", n)
				if _, err := dbconn.Global.ExecContext(ctx, `INSERT INTO repo(id, name) VALUES ($1, $2)`, n, name); err != nil {
					t.Fatal(err)
				}
				if _, err := dbconn.Global.ExecContext(ctx, `INSERT INTO default_repos(repo_id) VALUES ($1)`, n); err != nil {
					t.Fatal(err)
				}

				wantRepos = append(wantRepos, &types.Repo{
					ID:   api.RepoID(n),
					Name: api.RepoName(name),
				})
			}

			repos, err := DefaultRepos.List(ctx)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(repos, wantRepos) {
				t.Errorf("repos = %v, want %v", repos, wantRepos)
			}
		})
	}
}
