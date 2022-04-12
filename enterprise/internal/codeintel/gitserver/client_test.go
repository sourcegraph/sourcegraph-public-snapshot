package gitserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func newTestClient() *Client {
	db := database.NewMockDB()
	client := New(db, store.NewWithDB(db, &observation.TestContext), &observation.TestContext)
	return client
}

func TestParseDirectoryChildrenRoot(t *testing.T) {
	dirnames := []string{""}
	paths := []string{
		".github",
		".gitignore",
		"LICENSE",
		"README.md",
		"cmd",
		"go.mod",
		"go.sum",
		"internal",
		"protocol",
	}

	expected := map[string][]string{
		"": paths,
	}

	client := newTestClient()
	if diff := cmp.Diff(expected, client.parseDirectoryChildren(dirnames, paths)); diff != "" {
		t.Errorf("unexpected directory children result (-want +got):\n%s", diff)
	}
}

func TestParseDirectoryChildrenNonRoot(t *testing.T) {
	dirnames := []string{"cmd/", "protocol/", "cmd/protocol/"}
	paths := []string{
		"cmd/lsif-go",
		"protocol/protocol.go",
		"protocol/writer.go",
	}

	expected := map[string][]string{
		"cmd/":          {"cmd/lsif-go"},
		"protocol/":     {"protocol/protocol.go", "protocol/writer.go"},
		"cmd/protocol/": nil,
	}

	client := newTestClient()
	if diff := cmp.Diff(expected, client.parseDirectoryChildren(dirnames, paths)); diff != "" {
		t.Errorf("unexpected directory children result (-want +got):\n%s", diff)
	}
}

func TestCleanDirectoriesForLsTree(t *testing.T) {
	client := newTestClient()
	args := []string{"", "foo", "bar/", "baz"}
	actual := client.cleanDirectoriesForLsTree(args)
	expected := []string{".", "foo/", "bar/", "baz/"}

	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected ls-tree args (-want +got):\n%s", diff)
	}
}

// InitGitRepository initializes a new Git repository and runs cmds in a new
// temporary directory (returned as dir).
func InitGitRepository(t testing.TB, cmds ...string) string {
	t.Helper()
	remotes := filepath.Join(root, "remotes")
	if err := os.MkdirAll(remotes, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(remotes, strings.ReplaceAll(t.Name(), "/", "__"))
	if err != nil {
		t.Fatal(err)
	}
	cmds = append([]string{"git init"}, cmds...)
	for _, cmd := range cmds {
		out, err := GitCommand(dir, "bash", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fatalf("Command %q failed. Output was:\n\n%s", cmd, out)
		}
	}
	return dir
}

// MakeGitRepository calls initGitRepository to create a new Git repository and returns a handle to
// it.
func MakeGitRepository(t testing.TB, cmds ...string) api.RepoName {
	t.Helper()
	dir := InitGitRepository(t, cmds...)
	repo := api.RepoName(filepath.Base(dir))
	if resp, err := testGitserverClient.RequestRepoUpdate(context.Background(), repo, 0); err != nil {
		t.Fatal(err)
	} else if resp.Error != "" {
		t.Fatal(resp.Error)
	}
	return repo
}

func TestListDirectoryChildren(t *testing.T) {
	gitCommands := []string{
		"mkdir -p dir{1..3}/sub{1..3}",
		"touch dir1/sub1/file",
		"touch dir1/sub2/file",
		"touch dir2/sub1/file",
		"touch dir2/sub2/file",
		"touch dir3/sub1/file",
		"touch dir3/sub3/file",
		"git add .",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}

	db := database.NewMockDB()
	repo := MakeGitRepository(t, gitCommands...)

	ctx := context.Background()

	checker := authz.NewMockSubRepoPermissionChecker()
	// Start disabled
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return false
	})

	dirnames := []string{"dir1/", "dir2/", "dir3/"}
	children, err := ListDirectoryChildren(ctx, db, checker, repo, "HEAD", dirnames)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string][]string{
		"dir1/": {"dir1/sub1", "dir1/sub2"},
		"dir2/": {"dir2/sub1", "dir2/sub2"},
		"dir3/": {"dir3/sub1", "dir3/sub3"},
	}
	if diff := cmp.Diff(expected, children); diff != "" {
		t.Fatal(diff)
	}

	// With filtering
	checker.EnabledFunc.SetDefaultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefaultHook(func(ctx context.Context, i int32, content authz.RepoContent) (authz.Perms, error) {
		if strings.Contains(content.Path, "dir1/") {
			return authz.Read, nil
		}
		return authz.None, nil
	})
	ctx = actor.WithActor(ctx, &actor.Actor{
		UID: 1,
	})
	children, err = ListDirectoryChildren(ctx, db, checker, repo, "HEAD", dirnames)
	if err != nil {
		t.Fatal(err)
	}
	expected = map[string][]string{
		"dir1/": {"dir1/sub1", "dir1/sub2"},
		"dir2/": nil,
		"dir3/": nil,
	}
	if diff := cmp.Diff(expected, children); diff != "" {
		t.Fatal(diff)
	}
}
