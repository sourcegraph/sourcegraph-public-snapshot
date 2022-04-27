package git

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func TestRead(t *testing.T) {
	t.Parallel()

	db := database.NewMockDB()
	const wantData = "abcd\n"
	repo := MakeGitRepository(t,
		"echo abcd > file1",
		"git add file1",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	)
	const commitID = "3d689662de70f9e252d4f6f1d75284e23587d670"

	ctx := context.Background()

	tests := map[string]struct {
		file    string
		checkFn func(*testing.T, error, []byte)
	}{
		"all": {
			file: "file1",
			checkFn: func(t *testing.T, err error, data []byte) {
				if err != nil {
					t.Fatal(err)
				}
				if string(data) != wantData {
					t.Errorf("got %q, want %q", data, wantData)
				}
			},
		},

		"nonexistent": {
			file: "filexyz",
			checkFn: func(t *testing.T, err error, data []byte) {
				if err == nil {
					t.Fatal("err == nil")
				}
				if !os.IsNotExist(err) {
					t.Fatalf("got err %v, want os.IsNotExist", err)
				}
			},
		},
	}

	for name, test := range tests {
		checker := authz.NewMockSubRepoPermissionChecker()
		ctx = actor.WithActor(ctx, &actor.Actor{
			UID: 1,
		})
		t.Run(name+"-ReadFile", func(t *testing.T) {
			data, err := ReadFile(ctx, db, repo, commitID, test.file, nil)
			test.checkFn(t, err, data)
		})
		t.Run(name+"-ReadFile-with-sub-repo-permissions-no-op", func(t *testing.T) {
			checker.EnabledFunc.SetDefaultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
				if content.Path == test.file {
					return authz.Read, nil
				}
				return authz.None, nil
			})
			data, err := ReadFile(ctx, db, repo, commitID, test.file, checker)
			test.checkFn(t, err, data)
		})
		t.Run(name+"-ReadFile-with-sub-repo-permissions-filters-file", func(t *testing.T) {
			checker.EnabledFunc.SetDefaultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
				return authz.None, nil
			})
			data, err := ReadFile(ctx, db, repo, commitID, test.file, checker)
			if err != os.ErrNotExist {
				t.Errorf("unexpected error reading file: %s", err)
			}
			if string(data) != "" {
				t.Errorf("unexpected data: %s", data)
			}
		})
		t.Run(name+"-GetFileReader", func(t *testing.T) {
			runNewFileReaderTest(ctx, t, repo, commitID, test.file, nil, test.checkFn)
		})
		t.Run(name+"-GetFileReader-with-sub-repo-permissions-noop", func(t *testing.T) {
			checker.EnabledFunc.SetDefaultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
				if content.Path == test.file {
					return authz.Read, nil
				}
				return authz.None, nil
			})
			runNewFileReaderTest(ctx, t, repo, commitID, test.file, checker, test.checkFn)
		})
		t.Run(name+"-GetFileReader-with-sub-repo-permissions-filters-file", func(t *testing.T) {
			checker.EnabledFunc.SetDefaultHook(func() bool {
				return true
			})
			checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
				return authz.None, nil
			})
			rc, err := NewFileReader(ctx, database.NewMockDB(), repo, commitID, test.file, checker)
			if err != os.ErrNotExist {
				t.Fatalf("unexpected error: %s", err)
			}
			if rc != nil {
				t.Fatal("expected reader to be nil")
			}
		})
	}
}

func runNewFileReaderTest(ctx context.Context, t *testing.T, repo api.RepoName, commitID api.CommitID, file string,
	checker authz.SubRepoPermissionChecker, checkFn func(*testing.T, error, []byte)) {
	t.Helper()
	rc, err := NewFileReader(ctx, database.NewMockDB(), repo, commitID, file, checker)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	checkFn(t, err, data)
}
