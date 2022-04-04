package git

import (
	"context"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestDiffPath(t *testing.T) {
	testDiff := `
diff --git a/foo.md b/foo.md
index 51a59ef1c..493090958 100644
--- a/foo.md
+++ b/foo.md
@@ -1 +1 @@
-this is my file content
+this is my file contnent
`
	db := database.NewMockDB()
	t.Run("basic", func(t *testing.T) {
		Mocks.ExecReader = func(args []string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(testDiff)), nil
		}
		ctx := context.Background()
		checker := authz.NewMockSubRepoPermissionChecker()
		ctx = actor.WithActor(ctx, &actor.Actor{
			UID: 1,
		})
		hunks, err := DiffPath(ctx, db, "", "sourceCommit", "", "file", checker)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if len(hunks) != 1 {
			t.Errorf("unexpected hunks returned: %d", len(hunks))
		}
	})
	t.Run("with sub-repo permissions enabled", func(t *testing.T) {
		Mocks.ExecReader = func(args []string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(testDiff)), nil
		}
		ctx := context.Background()
		checker := authz.NewMockSubRepoPermissionChecker()
		ctx = actor.WithActor(ctx, &actor.Actor{
			UID: 1,
		})
		fileName := "foo"
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		// User doesn't have access to this file
		checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
			if content.Path == fileName {
				return authz.None, nil
			}
			return authz.Read, nil
		})
		hunks, err := DiffPath(ctx, db, "", "sourceCommit", "", fileName, checker)
		if !reflect.DeepEqual(err, os.ErrNotExist) {
			t.Errorf("unexpected error: %s", err)
		}
		if hunks != nil {
			t.Errorf("expected DiffPath to return no results, got %v", hunks)
		}
	})
}

func TestDiffFileIterator(t *testing.T) {
	t.Run("Close", func(t *testing.T) {
		c := new(closer)
		i := &DiffFileIterator{rdr: c}
		i.Close()
		if *c != true {
			t.Errorf("iterator did not close the underlying reader: have: %v; want: %v", *c, true)
		}
	})
}

type closer bool

func (c *closer) Read(p []byte) (int, error) {
	return 0, errors.New("testing only; this should never be invoked")
}

func (c *closer) Close() error {
	*c = true
	return nil
}
