package github

import (
	"io"
	"net/http"
	"strings"
	"testing"
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

func (s *mockHTTPResponseBody) Do(req *http.Request) (*http.Response, error) {
	s.count++
	status := s.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		Request:    req,
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(s.responseBody)),
	}, nil
}
