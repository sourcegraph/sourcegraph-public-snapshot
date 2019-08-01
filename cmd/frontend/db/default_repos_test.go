package db

import (
	"context"
	"fmt"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"testing"
)

func Test_defaultRepos_List(t *testing.T) {
	for N := 0; N <= 2; N++ {
		t.Run(fmt.Sprintf("with %d repo ids", N), func(t *testing.T) {
			ctx := context.Background()
			var wantRepoIds []int
			for n := 0; n < N; n++ {
				q := fmt.Sprintf(`INSERT INTO default_repos(repo_id) VALUES (%d)`, n)
				if _, err := dbconn.Global.ExecContext(ctx, q); err != nil {
					t.Fatal(err)
				}
			}

			repoIds, err := DefaultRepos.List(ctx)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
