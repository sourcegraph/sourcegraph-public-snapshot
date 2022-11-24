package httpcli

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/strings/slices"
)

func TestRedisLoggerMiddleware_getAllValuesAfter(t *testing.T) {
	rcache.SetupForTest(t)
	c := rcache.NewWithTTL("some_prefix", 1)

	var pairs = make([][2]string, 10)
	for i := 0; i < 10; i++ {
		pairs[i] = [2]string{"keys:key" + strconv.Itoa(i), "value" + strconv.Itoa(i)}
	}
	c.SetMulti(pairs...)

	key := "keys:key5"
	got, err := getAllValuesAfter(c, "keys", key, 10)

	assert.Nil(t, err)
	assert.Len(t, got, 4)

	got, err = getAllValuesAfter(c, "keys", key, 2)
	assert.Nil(t, err)
	assert.Len(t, got, 2)
}

func TestRedisLoggerMiddleware_removeSensitiveHeaders(t *testing.T) {
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

	cleanHeaders := removeSensitiveHeaders(input)

	if diff := cmp.Diff(cleanHeaders, want); diff != "" {
		t.Errorf("unexpected request headers (-have +want):\n%s", diff)
	}
}
