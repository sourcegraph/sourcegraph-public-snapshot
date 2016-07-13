package github

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"
)

const testGHCachePrefix = "__test__gh_pub"

func resetCache(t *testing.T) {
	if err := rcache.ClearAllForTest(testGHCachePrefix); err != nil {
		t.Fatal(err)
	}
	reposGithubPublicCache = rcache.NewByteCache(testGHCachePrefix, 1000)
}

// TestRepos_Get_existing_public tests the behavior of Repos.Get when called on a
// repo that exists (i.e., successful outcome, cache hit).
func TestRepos_Get_existing_public(t *testing.T) {
	testRepos_Get(t, false)
}

// TestRepos_Get_existing_private tests the behavior of Repos.Get when called on a
// repo that exists (i.e., successful outcome, cache miss).
func TestRepos_Get_existing_private(t *testing.T) {
	testRepos_Get(t, true)
}

// TestRepos_Get_nocache tests the behavior of Repos.Get when called on a
// repo that is private (i.e., it shouldn't be cached).
func testRepos_Get(t *testing.T, private bool) {
	githubcli.Config.GitHubHost = "github.com"
	resetCache(t)

	var calledGet bool
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
				calledGet = true
				return &github.Repository{
					ID:       github.Int(123),
					Name:     github.String("repo"),
					FullName: github.String("owner/repo"),
					Owner:    &github.User{ID: github.Int(1)},
					CloneURL: github.String("https://github.com/owner/repo.git"),
					Private:  github.Bool(private),
				}, nil, nil
			},
		},
	})

	repo, err := (&repos{}).Get(ctx, "github.com/owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if want := "123"; repo.Origin.ID != want {
		t.Errorf("got %s, want %s", repo.Origin.ID, want)
	}

	// Test that repo is not cached and fetched from client on second request.
	calledGet = false
	repo, err = (&repos{}).Get(ctx, "github.com/owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if private && !calledGet {
		t.Error("!calledGet, expected to miss cache")
	}
	if !private && calledGet {
		t.Error("calledGet, expected to hit cache")
	}
	if want := "123"; repo.Origin.ID != want {
		t.Errorf("got %s, want %s", repo.Origin.ID, want)
	}
}

// TestRepos_Get_nonexistent tests the behavior of Repos.Get when called
// on a repo that does not exist.
func TestRepos_Get_nonexistent(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	resetCache(t)

	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader(nil)),
					Request:    &http.Request{},
				}
				return nil, &github.Response{Response: resp}, github.CheckResponse(resp)
			},
		},
	})

	s := &repos{}
	nonexistentRepo := "github.com/owner/repo"
	repo, err := s.Get(ctx, nonexistentRepo)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}

// TestRepos_Get_publicnotfound tests we correctly cache 404 responses. If we
// are not authed as a user, all private repos 404 which counts towards our
// rate limit. This test will ensure unauthed clients cache 404, but authed
// users do not use the cache
func TestRepos_Get_publicnotfound(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	resetCache(t)

	calledGetMissing := false
	mockGetMissing := mockGitHubRepos{
		Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
			calledGetMissing = true
			resp := &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
				Request:    &http.Request{},
			}
			return nil, &github.Response{Response: resp}, github.CheckResponse(resp)
		},
	}
	mockGetPrivate := mockGitHubRepos{
		Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
			return &github.Repository{
				ID:       github.Int(123),
				Name:     github.String("repo"),
				FullName: github.String("owner/repo"),
				Owner:    &github.User{ID: github.Int(1)},
				CloneURL: github.String("https://github.com/owner/repo.git"),
				Private:  github.Bool(true),
			}, nil, nil
		},
	}

	mock := &minimalClient{}
	ctx := testContext(mock)

	s := &repos{}
	privateRepo := "github.com/owner/repo"

	// An unauthed user won't be able to see the repo
	mock.isAuthedUser = false
	mock.repos = mockGetMissing
	repo, err := s.Get(ctx, privateRepo)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if !calledGetMissing {
		t.Fatal("did not call repos.Get when it should not be cached")
	}

	// If we repeat the call, we should hit the cache
	calledGetMissing = false
	repo, err = s.Get(ctx, privateRepo)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if calledGetMissing {
		t.Fatal("should have hit cache")
	}

	// Now if we call as an authed user, we will hit the cache but not use
	// it since the repo may not 404 for us
	mock.isAuthedUser = true
	mock.repos = mockGetPrivate
	repo, err = s.Get(ctx, privateRepo)
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Fatal("repo is nil")
	}

	// Ensure the repo is still missing for unauthed users
	calledGetMissing = false
	mock.isAuthedUser = false
	mock.repos = mockGetMissing
	repo, err = s.Get(ctx, privateRepo)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if calledGetMissing {
		t.Fatal("should have hit cache")
	}

	// Authed user should never use public cache. Do twice to ensure we do not
	// use the cached 404 response.
	for i := 0; i < 2; i++ {
		calledGetMissing = false
		mock.isAuthedUser = true
		mock.repos = mockGetMissing // Pretend that privateRepo is deleted now, so even authed user can't see it. Do this to ensure cached 404 value isn't used by authed user.
		repo, err = s.Get(ctx, privateRepo)
		if grpc.Code(err) != codes.NotFound {
			t.Fatal(err)
		}
		if !calledGetMissing {
			t.Fatal("should not hit cache")
		}
	}
}

// TestRepos_Get_autheddoescache tests not found responses from authed
// clients are cached.
func TestRepos_Get_autheddoescache(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	resetCache(t)

	calledGetMissing := false
	mockGetMissing := mockGitHubRepos{
		Get_: func(owner, repo string) (*github.Repository, *github.Response, error) {
			calledGetMissing = true
			resp := &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
				Request:    &http.Request{},
			}
			return nil, &github.Response{Response: resp}, github.CheckResponse(resp)
		},
	}

	mock := &minimalClient{}
	ctx := testContext(mock)

	s := &repos{}
	missingRepo := "github.com/owner/doesnotexist"
	mock.repos = mockGetMissing

	// An authed user should not get from the cache, but it should store a 404 in the cache.
	calledGetMissing = false
	mock.isAuthedUser = true
	_, err := s.Get(ctx, missingRepo)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if !calledGetMissing {
		t.Fatal("should not hit cache")
	}

	// An unauthed user will hit cache, because authed user's 404 response should have gotten cached.
	calledGetMissing = false
	mock.isAuthedUser = false
	_, err = s.Get(ctx, missingRepo)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if calledGetMissing {
		t.Fatal("should have hit cache; if this fails, that means authed user's 404 response wasn't cached")
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

	repo, err := (&repos{}).GetByID(ctx, 123)
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if want := "123"; repo.Origin.ID != want {
		t.Errorf("got %s, want %s", repo.Origin.ID, want)
	}
}

// TestRepos_GetByID_nonexistent tests the behavior of Repos.GetByID
// when called on a repo that does not exist.
func TestRepos_GetByID_nonexistent(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	ctx := testContext(&minimalClient{
		repos: mockGitHubRepos{
			GetByID_: func(id int) (*github.Repository, *github.Response, error) {
				resp := &http.Response{
					StatusCode: http.StatusNotFound,
					Body:       ioutil.NopCloser(bytes.NewReader(nil)),
					Request:    &http.Request{},
				}
				return nil, &github.Response{Response: resp}, github.CheckResponse(resp)
			},
		},
	})

	s := &repos{}
	repo, err := s.GetByID(ctx, 456)
	if grpc.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}
