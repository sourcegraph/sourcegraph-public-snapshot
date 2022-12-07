package httpcli

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRedisLoggerMiddleware(t *testing.T) {
	rcache.SetupForTest(t)

	normalReq, _ := http.NewRequest("GET", "http://dev/null", strings.NewReader("horse"))
	complexReq, _ := http.NewRequest("PATCH", "http://test.aa?a=2", strings.NewReader("graph"))
	complexReq.Header.Set("Cache-Control", "no-cache")

	for _, tc := range []struct {
		req  *http.Request
		name string
		cli  Doer
		err  string
		want *types.OutboundRequestLogItem
	}{
		{
			req:  normalReq,
			name: "normal response",
			cli:  newFakeClientWithHeaders(map[string][]string{"X-Test-Header": {"value"}}, http.StatusOK, []byte(`{"responseBody":true}`), nil),
			err:  "<nil>",
			want: &types.OutboundRequestLogItem{
				Method:          normalReq.Method,
				URL:             normalReq.URL.String(),
				RequestHeaders:  map[string][]string{},
				RequestBody:     "horse",
				StatusCode:      http.StatusOK,
				ResponseHeaders: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}, "X-Test-Header": {"value"}},
			},
		},
		{
			req:  complexReq,
			name: "complex request",
			cli:  newFakeClientWithHeaders(map[string][]string{"X-Test-Header": {"value1", "value2"}}, http.StatusForbidden, []byte(`{"permission":false}`), nil),
			err:  "<nil>",
			want: &types.OutboundRequestLogItem{
				Method:          complexReq.Method,
				URL:             complexReq.URL.String(),
				RequestHeaders:  map[string][]string{"Cache-Control": {"no-cache"}},
				RequestBody:     "graph",
				StatusCode:      http.StatusForbidden,
				ResponseHeaders: map[string][]string{"Content-Type": {"text/plain; charset=utf-8"}, "X-Test-Header": {"value1", "value2"}},
			},
		},
		{
			req:  normalReq,
			name: "no response",
			cli: DoerFunc(func(r *http.Request) (*http.Response, error) {
				return nil, errors.New("oh no")
			}),
			err: "oh no",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Enable the feature
			old := OutboundRequestLogLimit()
			SetOutboundRequestLogLimit(1)
			t.Cleanup(func() { SetOutboundRequestLogLimit(old) })

			// Build client with middleware
			cli := redisLoggerMiddleware()(tc.cli)

			// Send request
			_, err := cli.Do(tc.req)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("have error: %q\nwant error: %q", have, want)
			}

			// Check logged request
			logged, err := GetOutboundRequestLogItems(context.Background(), "")
			if err != nil {
				t.Fatalf("couldnt get logged requests: %s", err)
			}
			if len(logged) != 1 {
				t.Fatalf("request was not logged")
			}

			if tc.want == nil {
				return
			}

			if diff := cmp.Diff(tc.want, logged[0], cmpopts.IgnoreFields(
				types.OutboundRequestLogItem{},
				"ID",
				"StartedAt",
				"Duration",
				"CreationStackFrame",
				"CallStackFrame",
			)); diff != "" {
				t.Fatalf("wrong request logged: %s", diff)
			}
		})
	}

}

func TestRedisLoggerMiddleware_getAllValuesAfter(t *testing.T) {
	rcache.SetupForTest(t)
	c := rcache.NewWithTTL("some_prefix", 1)
	ctx := context.Background()

	var pairs = make([][2]string, 10)
	for i := 0; i < 10; i++ {
		pairs[i] = [2]string{"key" + strconv.Itoa(i), "value" + strconv.Itoa(i)}
	}
	c.SetMulti(pairs...)

	key := "key5"
	got, err := getAllValuesAfter(ctx, c, key, 10)

	assert.Nil(t, err)
	assert.Len(t, got, 4)

	got, err = getAllValuesAfter(ctx, c, key, 2)
	assert.Nil(t, err)
	assert.Len(t, got, 2)
}

func TestRedisLoggerMiddleware_redactSensitiveHeaders(t *testing.T) {
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
	want := make(http.Header)
	riskyKeys := []string{"Bearer", "GHP_XXXX", "GLPAT-XXXX", "X-Powered-By"}
	for key, value := range input {
		if slices.Contains(riskyKeys, key) {
			want[key] = value
		} else {
			want[key] = []string{"REDACTED"}
		}
	}

	cleanHeaders := redactSensitiveHeaders(input)

	if diff := cmp.Diff(cleanHeaders, want); diff != "" {
		t.Errorf("unexpected request headers (-have +want):\n%s", diff)
	}
}

func TestRedisLoggerMiddleware_DeleteFirstN(t *testing.T) {
	rcache.SetupForTest(t)
	c := rcache.NewWithTTL("some_prefix", 1)

	// Add 10 key-value pairs
	var pairs = make([][2]string, 10)
	for i := 0; i < 10; i++ {
		pairs[i] = [2]string{"key" + strconv.Itoa(i), "value" + strconv.Itoa(i)}
	}
	c.SetMulti(pairs...)

	// Delete the first 4 key-value pairs
	_ = deleteExcessItems(c, 4)

	got, listErr := c.ListKeys(context.Background())

	assert.Nil(t, listErr)

	assert.Len(t, got, 4)

	assert.NotContains(t, got, "key0") // 0 through 5 should be deleted
	assert.NotContains(t, got, "key5")

	assert.Contains(t, got, "key6") // 6 through 9 (4 items) should be kept
	assert.Contains(t, got, "key9")
}

func TestRedisLoggerMiddleware_formatStackFrame(t *testing.T) {
	tests := []struct {
		name     string
		function string
		file     string
		line     int
		want     string
	}{
		{
			name:     "Sourcegraph internal package",
			function: "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend.(*requestTracer).TraceQuery",
			file:     "/Users/x/github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlbackend.go",
			line:     51,
			want:     "cmd/frontend/graphqlbackend/graphqlbackend.go:51 (Function: (*requestTracer).TraceQuery)",
		},
		{
			name:     "third-party package",
			function: "third-party/library.f",
			file:     "/Users/x/github.com/third-party/library/file.go",
			line:     11,
			want:     "third-party/library/file.go:11 (Function: f)",
		},
		{
			name:     "main package",
			function: "main.f",
			file:     "/Users/x/file.go",
			line:     11,
			want:     "file.go:11 (Function: f)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := formatStackFrame(test.function, test.file, test.line)
			if got != test.want {
				t.Errorf("got %q, want %q", got, test.want)
			}
		})
	}
}
