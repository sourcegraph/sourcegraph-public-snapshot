package github

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/github/githubcli"
)

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

const testGHCachePrefix = "__test__gh_pub"

func resetCache(t *testing.T) {
	rcache.SetupForTest(testGHCachePrefix)
	reposGithubPublicCache = rcache.NewWithTTL(testGHCachePrefix, 1000)
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
	ctx := newContext(context.Background(), github.NewClient(&http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calledGet = true
			body, err := json.Marshal(&github.Repository{
				ID:       github.Int(123),
				Name:     github.String("repo"),
				FullName: github.String("owner/repo"),
				Owner:    &github.User{ID: github.Int(1)},
				CloneURL: github.String("https://github.com/owner/repo.git"),
				Private:  github.Bool(private),
			})
			if err != nil {
				t.Fatal(err)
			}
			return &http.Response{
				Request:    req,
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
			}, nil
		}),
	}))

	repo, err := GetRepo(ctx, "github.com/owner/repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if !calledGet {
		t.Error("!calledGet")
	}

	// Test that repo is not cached and fetched from client on second request.
	calledGet = false
	repo, err = GetRepo(ctx, "github.com/owner/repo")
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
}

// TestRepos_Get_nonexistent tests the behavior of Repos.Get when called
// on a repo that does not exist.
func TestRepos_Get_nonexistent(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	resetCache(t)

	ctx := newContext(context.Background(), github.NewClient(&http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				Request:    req,
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}),
	}))

	nonexistentRepo := "github.com/owner/repo"
	repo, err := GetRepo(ctx, nonexistentRepo)
	if legacyerr.ErrCode(err) != legacyerr.NotFound {
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
	mockGetMissing := github.NewClient(&http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calledGetMissing = true
			return &http.Response{
				Request:    req,
				StatusCode: http.StatusNotFound,
				Body:       ioutil.NopCloser(bytes.NewReader(nil)),
			}, nil
		}),
	})
	mockGetPrivate := github.NewClient(&http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			body, err := json.Marshal(&github.Repository{
				ID:       github.Int(123),
				Name:     github.String("repo"),
				FullName: github.String("owner/repo"),
				Owner:    &github.User{ID: github.Int(1)},
				CloneURL: github.String("https://github.com/owner/repo.git"),
				Private:  github.Bool(true),
			})
			if err != nil {
				t.Fatal(err)
			}
			return &http.Response{
				Request:    req,
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
			}, nil
		}),
	})

	privateRepo := "github.com/owner/repo"

	// An unauthed user won't be able to see the repo
	ctx := newContext(context.Background(), mockGetMissing)
	ctx = auth.WithActor(ctx, &auth.Actor{})
	if _, err := GetRepo(ctx, privateRepo); legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Fatal(err)
	}
	if !calledGetMissing {
		t.Fatal("did not call repos.Get when it should not be cached")
	}

	// If we repeat the call, we should hit the cache
	calledGetMissing = false
	if _, err := GetRepo(ctx, privateRepo); legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Fatal(err)
	}
	if calledGetMissing {
		t.Fatal("should have hit cache")
	}

	// Now if we call as an authed user, we will hit the cache but not use
	// it since the repo may not 404 for us
	ctx = newContext(context.Background(), mockGetPrivate)
	ctx = auth.WithActor(ctx, &auth.Actor{UID: "1", Login: "test", GitHubToken: "test"})
	repo, err := GetRepo(ctx, privateRepo)
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Fatal("repo is nil")
	}

	// Ensure the repo is still missing for unauthed users
	calledGetMissing = false
	ctx = newContext(context.Background(), mockGetMissing)
	ctx = auth.WithActor(ctx, &auth.Actor{})
	if _, err := GetRepo(ctx, privateRepo); legacyerr.ErrCode(err) != legacyerr.NotFound {
		t.Fatal(err)
	}
	if calledGetMissing {
		t.Fatal("should have hit cache")
	}

	// Authed user should never use public cache. Do twice to ensure we do not
	// use the cached 404 response.
	for i := 0; i < 2; i++ {
		calledGetMissing = false
		ctx = newContext(context.Background(), mockGetMissing) // Pretend that privateRepo is deleted now, so even authed user can't see it. Do this to ensure cached 404 value isn't used by authed user.
		ctx = auth.WithActor(ctx, &auth.Actor{UID: "1", Login: "test", GitHubToken: "test"})
		if _, err := GetRepo(ctx, privateRepo); legacyerr.ErrCode(err) != legacyerr.NotFound {
			t.Fatal(err)
		}
		if !calledGetMissing {
			t.Fatal("should not hit cache")
		}
	}
}

// TestRepos_Get_authednocache tests authed users do add public repos to the cache
func TestRepos_Get_authednocache(t *testing.T) {
	githubcli.Config.GitHubHost = "github.com"
	resetCache(t)

	calledGet := false
	ctx := newContext(context.Background(), github.NewClient(&http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			calledGet = true
			body, err := json.Marshal(&github.Repository{
				ID:       github.Int(123),
				Name:     github.String("repo"),
				FullName: github.String("owner/repo"),
				Owner:    &github.User{ID: github.Int(1)},
				CloneURL: github.String("https://github.com/owner/repo.git"),
				Private:  github.Bool(false),
			})
			if err != nil {
				t.Fatal(err)
			}
			return &http.Response{
				Request:    req,
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewReader(body)),
			}, nil
		}),
	}))

	repo := "github.com/owner/repo"

	authedGet := func() bool {
		calledGet = false
		ctx = auth.WithActor(ctx, &auth.Actor{UID: "1", Login: "test", GitHubToken: "test"})
		_, err := GetRepo(ctx, repo)
		if err != nil {
			t.Fatal(err)
		}
		return calledGet
	}
	unauthedGet := func() bool {
		calledGet = false
		ctx = auth.WithActor(ctx, &auth.Actor{})
		_, err := GetRepo(ctx, repo)
		if err != nil {
			t.Fatal(err)
		}
		return calledGet
	}

	// An authed user will populate the empty cache
	if !authedGet() {
		t.Fatal("authed should populate empty cache")
	}

	// An unauthed user should now get from cache
	if unauthedGet() {
		t.Fatal("unauthed should get from cache")
	}

	// The authed user should also be able to get public repo from the cache
	if authedGet() {
		t.Fatal("authed should not get from cache")
	}
}
