package executor

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
)

var cacheRepo1 = batches.Repository{
	ID:          "src-cli",
	Name:        "github.com/sourcegraph/src-cli",
	BaseRef:     "refs/heads/main",
	BaseRev:     "d34db33f",
	FileMatches: []string{"README.md", "main.go"},
}

var cacheRepo2 = batches.Repository{
	ID:          "sourcegraph",
	Name:        "github.com/sourcegraph/sourcegraph",
	BaseRef:     "refs/heads/main-2",
	BaseRev:     "c0ff33",
	FileMatches: []string{"main.go"},
}

var testDiff = []byte(`diff --git a/README.md b/README.md
new file mode 100644
index 0000000..3363c39
--- /dev/null
+++ b/README.md
@@ -0,0 +1,3 @@
+# README
+
+This is the readme
`)

func TestExecutionDiskCache_GetSet(t *testing.T) {
	ctx := context.Background()

	cacheKey1 := &cache.CacheKey{
		Repository: cacheRepo1,
		Steps: []batcheslib.Step{
			{Run: "echo 'Hello World'", Container: "alpine:3"},
		},
	}

	cacheKey2 := &cache.CacheKey{
		Repository: cacheRepo2,
		Steps: []batcheslib.Step{
			{Run: "echo 'Hello World'", Container: "alpine:3"},
		},
	}

	value := execution.AfterStepResult{
		Version: 2,
		Diff:    testDiff,
		ChangedFiles: git.Changes{
			Added: []string{"README.md"},
		},
		Outputs: map[string]interface{}{},
	}

	cache := ExecutionDiskCache{Dir: t.TempDir()}

	// Empty cache, no hits
	assertCacheMiss(t, cache, cacheKey1)
	assertCacheMiss(t, cache, cacheKey2)

	// Set the cache
	if err := cache.Set(ctx, cacheKey1, value); err != nil {
		t.Fatalf("cache.Set returned unexpected error: %s", err)
	}

	// Cache hit
	assertCacheHit(t, cache, cacheKey1, value)

	// Cache miss due to different key
	assertCacheMiss(t, cache, cacheKey2)

	// Cache miss due to cleared cache
	if err := cache.Clear(ctx, cacheKey1); err != nil {
		t.Fatalf("cache.Get returned unexpected error: %s", err)
	}
	assertCacheMiss(t, cache, cacheKey1)
}

func assertCacheHit(t *testing.T, c ExecutionDiskCache, k cache.Keyer, want execution.AfterStepResult) {
	t.Helper()

	have, found, err := c.Get(context.Background(), k)
	if err != nil {
		t.Fatalf("cache.Get returned unexpected error: %s", err)
	}
	if !found {
		t.Fatalf("cache miss when hit was expected")
	}

	if diff := cmp.Diff(have, want); diff != "" {
		t.Errorf("wrong cached result (-have +want):\n\n%s", diff)
	}
}

func assertCacheMiss(t *testing.T, c ExecutionDiskCache, k cache.Keyer) {
	t.Helper()

	_, found, err := c.Get(context.Background(), k)
	if err != nil {
		t.Fatalf("cache.Get returned unexpected error: %s", err)
	}
	if found {
		t.Fatalf("cache hit when miss was expected")
	}
}
