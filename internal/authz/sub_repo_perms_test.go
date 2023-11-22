package authz

import (
	"context"
	"io/fs"
	"testing"

	"github.com/gobwas/glob"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/fileutil"
)

func TestFilterActorPaths(t *testing.T) {
	testPaths := []string{"file1", "file2", "file3"}
	checker := NewMockSubRepoPermissionChecker()
	ctx := context.Background()
	a := &actor.Actor{
		UID: 1,
	}
	ctx = actor.WithActor(ctx, a)
	repo := api.RepoName("foo")

	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.FilePermissionsFuncFunc.SetDefaultHook(func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
		return func(path string) (Perms, error) {
			if path == "file1" {
				return Read, nil
			}
			return None, nil
		}, nil
	})

	filtered, err := FilterActorPaths(ctx, checker, a, repo, testPaths)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"file1"}
	if diff := cmp.Diff(want, filtered); diff != "" {
		t.Fatal(diff)
	}
}

func TestCanReadAllPaths(t *testing.T) {
	testPaths := []string{"file1", "file2", "file3"}
	checker := NewMockSubRepoPermissionChecker()
	ctx := context.Background()
	a := &actor.Actor{
		UID: 1,
	}
	ctx = actor.WithActor(ctx, a)
	repo := api.RepoName("foo")

	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.FilePermissionsFuncFunc.SetDefaultHook(func(context.Context, int32, api.RepoName) (FilePermissionFunc, error) {
		return func(path string) (Perms, error) {
			switch path {
			case "file1", "file2", "file3":
				return Read, nil
			default:
				return None, nil
			}
		}, nil
	})
	checker.EnabledForRepoFunc.SetDefaultHook(func(ctx context.Context, rn api.RepoName) (bool, error) {
		if rn == repo {
			return true, nil
		}
		return false, nil
	})

	ok, err := CanReadAllPaths(ctx, checker, repo, testPaths)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("Should be allowed to read all paths")
	}
	ok, err = CanReadAnyPath(ctx, checker, repo, testPaths)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("CanReadyAnyPath should've returned true since the user can read all paths")
	}

	// Add path we can't read
	testPaths = append(testPaths, "file4")

	ok, err = CanReadAllPaths(ctx, checker, repo, testPaths)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("Should fail, not allowed to read file4")
	}
	ok, err = CanReadAnyPath(ctx, checker, repo, testPaths)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("user can read some of the testPaths, so CanReadAnyPath should return true")
	}
}

func TestSubRepoEnabled(t *testing.T) {
	t.Run("checker is nil", func(t *testing.T) {
		if SubRepoEnabled(nil) {
			t.Errorf("expected checker to be invalid since it is nil")
		}
	})
	t.Run("checker is not enabled", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return false
		})
		if SubRepoEnabled(checker) {
			t.Errorf("expected checker to be invalid since it is disabled")
		}
	})
	t.Run("checker is enabled", func(t *testing.T) {
		checker := NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		if !SubRepoEnabled(checker) {
			t.Errorf("expected checker to be valid since it is enabled")
		}
	})
}

func TestFileInfoPath(t *testing.T) {
	t.Run("adding trailing slash to directory", func(t *testing.T) {
		fi := &fileutil.FileInfo{
			Name_: "app",
			Mode_: fs.ModeDir,
		}
		assert.Equal(t, "app/", fileInfoPath(fi))
	})
	t.Run("doesn't add trailing slash if not directory", func(t *testing.T) {
		fi := &fileutil.FileInfo{
			Name_: "my-file.txt",
		}
		assert.Equal(t, "my-file.txt", fileInfoPath(fi))
	})
}

func TestGlobMatchOnlyDirectories(t *testing.T) {
	g, err := glob.Compile("**/", '/')
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, g.Match("foo/"))
	assert.True(t, g.Match("foo/thing/"))
	assert.False(t, g.Match("foo/thing"))
	assert.False(t, g.Match("/foo/thing"))
}
