package github

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/store/testsuite"
)

func TestRepos_Get_existing(t *testing.T) {
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
				return &github.Repository{
					ID:       github.Int(1),
					Name:     github.String("repo"),
					FullName: github.String("owner/repo"),
					Owner:    &github.User{ID: github.Int(1)},
					CloneURL: github.String("https://github.com/owner/repo.git"),
				}, nil, nil
			},
		},
	})
	testsuite.Repos_Get_existing(ctx, t, &Repos{}, "github.com/owner/repo")
}

func TestRepos_Get_nonexistent(t *testing.T) {
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
				resp := &http.Response{StatusCode: http.StatusNotFound, Body: ioutil.NopCloser(bytes.NewReader(nil))}
				return nil, &github.Response{Response: resp}, github.CheckResponse(resp)
			},
		},
	})
	testsuite.Repos_Get_nonexistent(ctx, t, &Repos{}, "github.com/owner/repo")
}
