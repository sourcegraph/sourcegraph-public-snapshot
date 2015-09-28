package fs

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"

	"golang.org/x/net/context"
)

func testContextWithRepo() (context.Context, string, func()) {
	ctx, done := testContext()

	repo := "test-org/test-repo"
	repoAbsPath := filepath.Join(reposAbsPath(ctx), repo)
	err := os.MkdirAll(repoAbsPath, 0700)
	if err != nil {
		log.Fatalf("could not mkdir test repo dir: %s", err)
	}

	cmd := exec.Command("git", "init")
	cmd.Dir = repoAbsPath
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("could not initialize test repository (%s): %s", err, out)
	}

	// done will clean up our repo as well since it is a subdir of the test context
	return ctx, repo, done
}

func TestDiscussionsCreate(t *testing.T) {
	ctx, repo, done := testContextWithRepo()
	defer done()

	store := &Discussions{}
	testsuite.Discussions_Create_ok(ctx, t, store, repo)
}

func TestDiscussionsGet(t *testing.T) {
	ctx, repo, done := testContextWithRepo()
	defer done()

	store := &Discussions{}
	testsuite.Discussions_Get_ok(ctx, t, store, repo)
}

func TestDiscussionsListDefKey(t *testing.T) {
	ctx, repo, done := testContextWithRepo()
	defer done()

	store := &Discussions{}
	testsuite.Discussions_List_DefKey(ctx, t, store, repo)
}

func TestDiscussionsListRepo(t *testing.T) {
	ctx, repo, done := testContextWithRepo()
	defer done()

	store := &Discussions{}
	testsuite.Discussions_List_Repo(ctx, t, store, repo)
}
