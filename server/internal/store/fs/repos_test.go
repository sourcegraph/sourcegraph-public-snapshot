package fs

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func preCreateRepo(repo *sourcegraph.Repo) *sourcegraph.Repo {
	repo.VCS = "git"
	return repo
}

func createRepoDir(path, name string) error {
	gitDir := filepath.Join(path, name, ".git")
	if err := os.MkdirAll(gitDir, 0777); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(gitDir, "config"))
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func TestRepos_Get_existing(t *testing.T) {
	ctx := createTestContext(t)
	err := createRepoDir(reposAbsPath(ctx), "repo")
	if err != nil {
		t.Fatal(err)
	}
	testsuite.Repos_Get_existing(ctx, t, &Repos{}, "repo")
}

func TestRepos_Get_local(t *testing.T) {
	ctx := createTestContext(t)

	repoName := "repo"
	host := "example.com"
	err := createRepoDir(reposAbsPath(ctx), repoName)
	if err != nil {
		t.Fatal(err)
	}
	ctx = conf.WithAppURL(ctx, &url.URL{Scheme: "http", Host: host})
	expectedCloneURL := fmt.Sprintf("http://%s/%s", host, repoName)

	s := &Repos{}
	repo, err := s.Get(ctx, repoName)
	if err != nil {
		t.Fatal(err)
	}
	// The host of the clone URL should match the app's host.
	if repo.HTTPCloneURL != expectedCloneURL {
		t.Errorf("got HTTP Clone URL %q, want %q", repo.HTTPCloneURL, expectedCloneURL)
	}
}

func TestRepos_Get_nonexistent(t *testing.T) {
	ctx := createTestContext(t)
	testsuite.Repos_Get_nonexistent(ctx, t, &Repos{}, "owner/repo")
}

func TestRepos_List_query(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_query(ctx, t, &Repos{}, preCreateRepo)
}

func TestRepos_List_URIs(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_URIs(ctx, t, &Repos{}, preCreateRepo)
}

func TestRepos_Create(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create(ctx, t, &Repos{}, preCreateRepo)
}

func TestRepos_Create_dupe(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create_dupe(ctx, t, &Repos{}, preCreateRepo)
}

func TestRepos_Update_Description(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_Description(ctx, t, &Repos{}, preCreateRepo)
}

func TestRepos_Update_UpdatedAt(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_UpdatedAt(ctx, t, &Repos{}, preCreateRepo)
}

func TestRepos_Update_PushedAt(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_PushedAt(ctx, t, &Repos{}, preCreateRepo)
}
