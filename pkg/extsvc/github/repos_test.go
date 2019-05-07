package github

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/ratelimit"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
)

func TestSplitRepositoryNameWithOwner(t *testing.T) {
	owner, name, err := SplitRepositoryNameWithOwner("a/b")
	if err != nil {
		t.Fatal(err)
	}
	if want := "a"; owner != want {
		t.Errorf("got owner %q, want %q", owner, want)
	}
	if want := "b"; name != want {
		t.Errorf("got name %q, want %q", name, want)
	}
}

type mockHTTPResponseBody struct {
	count        int
	responseBody string
	status       int
}

func newMockHTTPResponseBody(responseBody string, status int) *mockHTTPResponseBody {
	return &mockHTTPResponseBody{
		responseBody: responseBody,
	}
}

func (s *mockHTTPResponseBody) Do(req *http.Request) (*http.Response, error) {
	s.count++
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		Request:    req,
		StatusCode: status,
		Body:       ioutil.NopCloser(strings.NewReader(s.responseBody)),
	}, nil
}

type mockHTTPEmptyResponse struct {
	statusCode int
}

func (s mockHTTPEmptyResponse) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Request:    req,
		StatusCode: s.statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
	}, nil
}

func newTestClient(t *testing.T, cli httpcli.Doer) *Client {
	rcache.SetupForTest(t)
	return &Client{
		apiURL:          &url.URL{Scheme: "https", Host: "example.com", Path: "/"},
		httpClient:      cli,
		RateLimit:       &ratelimit.Monitor{},
		repoCache:       map[string]*rcache.Cache{},
		repoCachePrefix: "__test__gh_repo",
		repoCacheTTL:    1000,
	}
}

// TestClient_GetRepository tests the behavior of GetRepository.
func TestClient_GetRepository(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
	"node_id": "i",
	"full_name": "o/r",
	"description": "d",
	"html_url": "https://github.example.com/o/r",
	"fork": true
}
`}
	c := newTestClient(t, &mock)

	want := Repository{
		ID:            "i",
		NameWithOwner: "o/r",
		Description:   "d",
		URL:           "https://github.example.com/o/r",
		IsFork:        true,
	}

	repo, err := c.GetRepository(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to miss cache once", mock.count)
	}
	if !reflect.DeepEqual(repo, &want) {
		t.Errorf("got repository %+v, want %+v", repo, &want)
	}

	// Test that repo is cached (and therefore NOT fetched) from client on second request.
	repo, err = c.GetRepository(context.Background(), "owner", "repo")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to hit cache", mock.count)
	}
	if !reflect.DeepEqual(repo, &want) {
		t.Errorf("got repository %+v, want %+v", repo, &want)
	}
}

func TestClient_GetRepositoriesByNodeFromAPI(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
  "data": {
    "nodes": [
      {
		"id": "i0",
		"nameWithOwner": "o/r0",
		"description": "d0",
		"url": "https://github.example.com/o/r0",
		"isFork": false
      },
      {
		"id": "i1",
		"nameWithOwner": "o/r1",
		"description": "d1",
		"url": "https://github.example.com/o/r1",
		"isFork": false
      },
      {
		"id": "i2",
		"nameWithOwner": "o/r2",
		"description": "d2",
		"url": "https://github.example.com/o/r2",
		"isFork": false
      }
    ]
  }
}
`}
	c := newTestClient(t, &mock)
	want := map[string]*Repository{
		"i0": {
			ID:            "i0",
			NameWithOwner: "o/r0",
			Description:   "d0",
			URL:           "https://github.example.com/o/r0",
		},
		"i1": {
			ID:            "i1",
			NameWithOwner: "o/r1",
			Description:   "d1",
			URL:           "https://github.example.com/o/r1",
		},
		"i2": {
			ID:            "i2",
			NameWithOwner: "o/r2",
			Description:   "d2",
			URL:           "https://github.example.com/o/r2",
		},
	}
	gotRepos, err := c.GetRepositoriesByNodeIDFromAPI(context.Background(), "", []string{"i0", "i1", "i2"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotRepos, want) {
		t.Errorf("got repos %+v, want %+v", gotRepos, want)
	}
}

// TestClient_GetRepository_nonexistent tests the behavior of GetRepository when called
// on a repository that does not exist.
func TestClient_GetRepository_nonexistent(t *testing.T) {
	mock := mockHTTPEmptyResponse{http.StatusNotFound}
	c := newTestClient(t, &mock)

	repo, err := c.GetRepository(context.Background(), "owner", "repo")
	if !IsNotFound(err) {
		t.Errorf("got err == %v, want IsNotFound(err) == true", err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}

// TestClient_GetRepositoryByNodeID tests the behavior of GetRepositoryByNodeID.
func TestClient_GetRepositoryByNodeID(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
	"data": {
		"node": {
			"id": "i",
			"nameWithOwner": "o/r",
			"description": "d",
			"url": "https://github.example.com/o/r",
			"isFork": true
		}
	}
}
`}
	c := newTestClient(t, &mock)

	want := Repository{
		ID:            "i",
		NameWithOwner: "o/r",
		Description:   "d",
		URL:           "https://github.example.com/o/r",
		IsFork:        true,
	}

	repo, err := c.GetRepositoryByNodeID(context.Background(), "", "i")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to miss cache once", mock.count)
	}
	if !reflect.DeepEqual(repo, &want) {
		t.Errorf("got repository %+v, want %+v", repo, &want)
	}

	// Test that repo is cached (and therefore NOT fetched) from client on second request.
	repo, err = c.GetRepositoryByNodeID(context.Background(), "", "i")
	if err != nil {
		t.Fatal(err)
	}
	if repo == nil {
		t.Error("repo == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to hit cache", mock.count)
	}
	if !reflect.DeepEqual(repo, &want) {
		t.Errorf("got repository %+v, want %+v", repo, &want)
	}
}

// TestClient_GetRepositoryByNodeID_nonexistent tests the behavior of GetRepositoryByNodeID when called
// on a repository that does not exist.
func TestClient_GetRepositoryByNodeID_nonexistent(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
	"data": {
		"node": null
	}
}
`}
	c := newTestClient(t, &mock)

	repo, err := c.GetRepositoryByNodeID(context.Background(), "", "i")
	if !IsNotFound(err) {
		t.Errorf("got err == %v, want IsNotFound(err) == true", err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}

func stringForRepoList(repos []*Repository) string {
	repoStrings := []string{}
	for _, repo := range repos {
		repoStrings = append(repoStrings, fmt.Sprintf("%#v", repo))
	}
	return "{\n" + strings.Join(repoStrings, ",\n") + "}\n"
}

func repoListsAreEqual(a []*Repository, b []*Repository) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if *a[i] != *b[i] {
			return false
		}
	}
	return true
}

func TestClient_ListRepositoriesForSearch(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
  "total_count": 2,
  "incomplete_results": false,
  "items": [
    {
      "node_id": "i",
      "full_name": "o/r",
      "description": "d",
      "html_url": "https://github.example.com/o/r",
      "fork": true
    },
    {
      "node_id": "j",
      "full_name": "a/b",
      "description": "c",
      "html_url": "https://github.example.com/a/b",
      "fork": false
    }
  ]
}
`}
	c := newTestClient(t, &mock)

	wantRepos := []*Repository{
		{
			ID:            "i",
			NameWithOwner: "o/r",
			Description:   "d",
			URL:           "https://github.example.com/o/r",
			IsFork:        true,
		},
		{
			ID:            "j",
			NameWithOwner: "a/b",
			Description:   "c",
			URL:           "https://github.example.com/a/b",
			IsFork:        false,
		},
	}

	reposPage, err := c.ListRepositoriesForSearch(context.Background(), "org:sourcegraph", 1)
	if err != nil {
		t.Fatal(err)
	}
	if !repoListsAreEqual(reposPage.Repos, wantRepos) {
		t.Errorf("got repositories:\n%s\nwant:\n%s", stringForRepoList(reposPage.Repos), stringForRepoList(wantRepos))
	}
}

func TestClient_ListRepositoriesForSearch_incomplete(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
{
  "total_count": 2,
  "incomplete_results": true,
  "items": [
    {
      "node_id": "i",
      "full_name": "o/r",
      "description": "d",
      "html_url": "https://github.example.com/o/r",
      "fork": true
    },
    {
      "node_id": "j",
      "full_name": "a/b",
      "description": "c",
      "html_url": "https://github.example.com/a/b",
      "fork": false
    }
  ]
}
`}
	c := newTestClient(t, &mock)

	// If we have incomplete results we want to fail. Our syncer requires all
	// repositories to be returned, otherwise it will delete the missing
	// repositories.
	want := `github repository search returned incomplete results. This is an ephemeral error: query="org:sourcegraph" page=1 total=2`
	_, err := c.ListRepositoriesForSearch(context.Background(), "org:sourcegraph", 1)

	if have := fmt.Sprint(err); want != have {
		t.Errorf("\nhave: %s\nwant: %s", have, want)
	}
}

// 🚨 SECURITY: test that cache entries are keyed by auth token
func TestClient_GetRepositoryByNodeID_security(t *testing.T) {
	c := newTestClient(t, newMockHTTPResponseBody(`{ "data": { "node": { "id": "i0" } } }`, http.StatusOK))

	got, err := c.GetRepositoryByNodeID(context.Background(), "tok0", "id0")
	if err != nil {
		t.Fatal(err)
	}
	if want := (&Repository{ID: "i0"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	c.httpClient = newMockHTTPResponseBody(`{ "data": { "node": { "id": "SHOULD NOT BE SEEN" } } }`, http.StatusOK)
	got, err = c.GetRepositoryByNodeID(context.Background(), "tok0", "id0")
	if err != nil {
		t.Fatal(err)
	}
	if want := (&Repository{ID: "i0"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	c.httpClient = newMockHTTPResponseBody(`{ "data": { "node": { "id": "i0-tok1" } } }`, http.StatusOK)
	got, err = c.GetRepositoryByNodeID(context.Background(), "tok1", "id0")
	if err != nil {
		t.Fatal(err)
	}
	if want := (&Repository{ID: "i0-tok1"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	c.httpClient = newMockHTTPResponseBody(`{}`, http.StatusNotFound)
	_, err = c.GetRepositoryByNodeID(context.Background(), "tok0", "id1")
	if err != ErrNotFound {
		t.Errorf("expected err %v, but got %v", ErrNotFound, err)
	}

	// "not found" should be cached
	c.httpClient = newMockHTTPResponseBody(`{ "data": { "node": { "id": "id1" } } }`, http.StatusOK)
	_, err = c.GetRepositoryByNodeID(context.Background(), "tok0", "id1")
	if err != ErrNotFound {
		t.Errorf("expected err %v, but got %v", ErrNotFound, err)
	}
	// "not found" not cached with different auth token
	got, err = c.GetRepositoryByNodeID(context.Background(), "tok1", "id1")
	if err != nil {
		t.Fatal(err)
	}
	if want := (&Repository{ID: "id1"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	// "not found" not cached with no auth token
	got, err = c.GetRepositoryByNodeID(context.Background(), "", "id1")
	if err != nil {
		t.Fatal(err)
	}
	if want := (&Repository{ID: "id1"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
