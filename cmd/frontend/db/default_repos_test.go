package db

import (
	"fmt"
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

//
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_40(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
