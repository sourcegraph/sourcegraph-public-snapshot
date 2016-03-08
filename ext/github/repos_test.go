package github

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sourcegraph/go-github/github"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
)

// TestRepos_Get_existing tests the behavior of Repos.Get when called on a
// repo that exists (i.e., the successful outcome).
func TestRepos_Get_existing(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
				return &github.Repository{
					ID:       github.Int(123),
					Name:     github.String("repo"),
					FullName: github.String("owner/repo"),
					Owner:    &github.User{ID: github.Int(1)},
					CloneURL: github.String("https://github.com/owner/repo.git"),
				}, nil, nil
			},
		},
	})

	repo, err := (&Repos{}).Get(ctx, "github.com/owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if want := int32(123); repo.GitHubID != want {
		t.Errorf("got %d, want %d", repo.GitHubID, want)
	}
}

// TestRepos_Get_nonexistent tests the behavior of Repos.Get when called
// on a repo that does not exist.
func TestRepos_Get_nonexistent(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
				resp := &http.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader(nil))}
				return nil, &github.Response{Response: resp}, github.CheckResponse(resp)
			},
		},
	})

	s := &Repos{}
	nonexistentRepo := "github.com/owner/repo"
	repo, err := s.Get(ctx, nonexistentRepo)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}

// TestRepos_GetByID_existing tests the behavior of Repos.GetByID when
// called on a repo that exists (i.e., the successful outcome).
func TestRepos_GetByID_existing(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			GetByID_: func(id int) (*github.Repository, *github.Response, error) {
				return &github.Repository{
					ID:       github.Int(123),
					Name:     github.String("repo"),
					FullName: github.String("owner/repo"),
					Owner:    &github.User{ID: github.Int(1)},
					CloneURL: github.String("https://github.com/owner/repo.git"),
				}, nil, nil
			},
		},
	})

	repo, err := (&Repos{}).GetByID(ctx, 123)
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if want := int32(123); repo.GitHubID != want {
		t.Errorf("got %d, want %d", repo.GitHubID, want)
	}
}

// TestRepos_GetByID_nonexistent tests the behavior of Repos.GetByID
// when called on a repo that does not exist.
func TestRepos_GetByID_nonexistent(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			GetByID_: func(id int) (*github.Repository, *github.Response, error) {
				resp := &http.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader(nil))}
				return nil, &github.Response{Response: resp}, github.CheckResponse(resp)
			},
		},
	})

	s := &Repos{}
	repo, err := s.GetByID(ctx, 456)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}
