package httptestutil

import (
	"net/http"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-cmp/cmp"
)

func TestRiskyHeaderFilter(t *testing.T) {
	input := http.Header{
		"Authorization":   []string{"all values", "should be", "removed"},
		"Bearer":          []string{"this should be kept as the risky value is only in the name"},
		"GHP_XXXX":        []string{"this should be kept"},
		"GLPAT-XXXX":      []string{"this should also be kept"},
		"GitHub-PAT":      []string{"this should be removed: ghp_XXXX"},
		"GitLab-PAT":      []string{"this should be removed", "glpat-XXXX"},
		"Innocent-Header": []string{"this should be removed as it includes", "the word bearer"},
		"Set-Cookie":      []string{"this is verboten"},
		"Token":           []string{"a token should be removed"},
		"X-Powered-By":    []string{"PHP"},
		"X-Token":         []string{"something that smells like a token should also be removed"},
	}

	// Build the expected output.
	want := http.Header{}
	for _, k := range []string{"Bearer", "GHP_XXXX", "GLPAT-XXXX", "X-Powered-By"} {
		want[k] = input[k]
	}

	i := cassette.Interaction{
		Request:  cassette.Request{Headers: input},
		Response: cassette.Response{Headers: input},
	}

	err := riskyHeaderFilter(&i)
	if err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}

	if diff := cmp.Diff(i.Request.Headers, want); diff != "" {
		t.Errorf("unexpected request headers (-have +want):\n%s", diff)
	}
	if diff := cmp.Diff(i.Response.Headers, want); diff != "" {
		t.Errorf("unexpected response headers (-have +want):\n%s", diff)
	}
}
