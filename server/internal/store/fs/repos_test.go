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
	testsuite.Repos_Get_existing(ctx, t, &repos{}, "repo")
}

func TestRepos_Get_local(t *testing.T) {
	ctx := createTestContext(t)

	repoName := "repo"
	host := "example.com"
	err := createRepoDir(reposAbsPath(ctx), repoName)
	if err != nil {
		t.Fatal(err)
	}
	ctx = conf.WithURL(ctx, &url.URL{Scheme: "http", Host: host}, &url.URL{Scheme: "ssh", Host: host, User: url.User("git")})
	expectedCloneURL := fmt.Sprintf("http://%s/%s", host, repoName)
	expectedSSHCloneURL := fmt.Sprintf("ssh://git@%s/%s", host, repoName)

	s := &repos{}
	repo, err := s.Get(ctx, repoName)
	if err != nil {
		t.Fatal(err)
	}
	// The host of the clone URL should match the app's host.
	if repo.HTTPCloneURL != expectedCloneURL {
		t.Errorf("got HTTP Clone URL %q, want %q", repo.HTTPCloneURL, expectedCloneURL)
	}
	// The host of the clone SSH URL should match the app's host.
	if repo.SSHCloneURL != expectedSSHCloneURL {
		t.Errorf("got SSH Clone URL %q, want %q", repo.SSHCloneURL, expectedSSHCloneURL)
	}
}

func TestRepos_Get_nonexistent(t *testing.T) {
	ctx := createTestContext(t)
	testsuite.Repos_Get_nonexistent(ctx, t, &repos{}, "owner/repo")
}

func TestRepos_List_query(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_query(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_List_URIs(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_List_URIs(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Create(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Create_dupe(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Create_dupe(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Update_Description(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_Description(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Update_UpdatedAt(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_UpdatedAt(ctx, t, &repos{}, preCreateRepo)
}

func TestRepos_Update_PushedAt(t *testing.T) {
	ctx, done := testContext()
	defer done()
	testsuite.Repos_Update_PushedAt(ctx, t, &repos{}, preCreateRepo)
}
