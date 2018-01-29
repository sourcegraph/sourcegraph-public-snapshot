package github

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
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

type mockGraphQLResponse struct {
	count        int
	responseBody string
}

func (s *mockGraphQLResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	s.count++
	return &http.Response{
		Request:    req,
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader(s.responseBody)),
	}, nil
}

type mockHTTPEmptyResponse struct {
	statusCode int
}

func (s mockHTTPEmptyResponse) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Request:    req,
		StatusCode: s.statusCode,
		Body:       ioutil.NopCloser(bytes.NewReader(nil)),
	}, nil
}

func newTestClient() *Client {
	const cachePrefix = "__test__gh_repo"
	rcache.SetupForTest(cachePrefix)
	return &Client{
		apiURL:     &url.URL{Scheme: "https", Host: "example.com", Path: "/"},
		repoCache:  rcache.NewWithTTL(cachePrefix, 1000),
		httpClient: &http.Client{},
	}
}

// TestClient_GetRepository_nocache tests the behavior of GetRepository.
func TestClient_GetRepository(t *testing.T) {
	mock := mockGraphQLResponse{
		responseBody: `
{
	"node_id": "i",
	"full_name": "o/r",
	"description": "d",
	"html_url": "https://github.example.com/o/r",
	"fork": true
}
`}
	c := newTestClient()
	c.httpClient.Transport = &mock

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

// TestClient_GetRepository_nonexistent tests the behavior of GetRepository when called
// on a repository that does not exist.
func TestClient_GetRepository_nonexistent(t *testing.T) {
	mock := mockHTTPEmptyResponse{http.StatusNotFound}
	c := newTestClient()
	c.httpClient.Transport = &mock

	repo, err := c.GetRepository(context.Background(), "owner", "repo")
	if !IsNotFound(err) {
		t.Errorf("got err == %v, want IsNotFound(err) == true", err)
	}
	if repo != nil {
		t.Error("repo != nil")
	}
}
