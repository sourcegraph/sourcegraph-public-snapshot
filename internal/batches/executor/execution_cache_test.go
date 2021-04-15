package executor

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/git"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"gopkg.in/yaml.v3"
)

const testExecutionCacheKeyEnv = "TEST_EXECUTION_CACHE_KEY_ENV"

func TestExecutionCacheKey(t *testing.T) {
	// Let's set up an array of steps that we can test with. One step will
	// depend on an environment variable outside the spec.
	var steps []batches.Step
	if err := yaml.Unmarshal([]byte(`
- run: foo
  env:
    FOO: BAR

- run: bar
  env:
    - FOO: BAR
    - `+testExecutionCacheKeyEnv+`
`), &steps); err != nil {
		t.Fatal(err)
	}

	// And now we can set up a key to work with.
	key := ExecutionCacheKey{&Task{Steps: steps}}

	// All righty. Let's get ourselves a baseline cache key here.
	initial, err := key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Let's set an unrelated environment variable and ensure we still have the
	// same key.
	if err := os.Setenv(testExecutionCacheKeyEnv+"_UNRELATED", "foo"); err != nil {
		t.Fatal(err)
	}
	have, err := key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(initial) != string(have) {
		t.Errorf("unexpected change in key: initial=%q have=%q", initial, have)
	}

	// Let's now set the environment variable referenced in the steps and verify
	// that the cache key does change.
	if err := os.Setenv(testExecutionCacheKeyEnv, "foo"); err != nil {
		t.Fatal(err)
	}
	have, err = key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(initial) == string(have) {
		t.Errorf("unexpected lack of change in key: %q", have)
	}

	// And, just to be sure, let's change it again.
	if err := os.Setenv(testExecutionCacheKeyEnv, "bar"); err != nil {
		t.Fatal(err)
	}
	again, err := key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(initial) == string(again) || string(have) == string(again) {
		t.Errorf("unexpected lack of change in key: %q", again)
	}

	// Finally, if we unset the environment variable again, we should get a key
	// that matches the initial key.
	if err := os.Unsetenv(testExecutionCacheKeyEnv); err != nil {
		t.Fatal(err)
	}
	have, err = key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(initial) != string(have) {
		t.Errorf("unexpected change in key: initial=%q have=%q", initial, have)
	}
}

const testDiff = `diff --git a/README.md b/README.md
new file mode 100644
index 0000000..3363c39
--- /dev/null
+++ b/README.md
@@ -0,0 +1,3 @@
+# README
+
+This is the readme
`

func TestExecutionDiskCache(t *testing.T) {
	ctx := context.Background()

	cacheTmpDir := func(t *testing.T) string {
		testTempDir, err := ioutil.TempDir("", "execution-disk-cache-test-*")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Remove(testTempDir) })

		return testTempDir
	}

	cacheKey1 := ExecutionCacheKey{Task: &Task{
		Repository: &graphql.Repository{Name: "src-cli"},
		Steps: []batches.Step{
			{Run: "echo 'Hello World'", Container: "alpine:3"},
		},
	}}

	cacheKey2 := ExecutionCacheKey{Task: &Task{
		Repository: &graphql.Repository{Name: "documentation"},
		Steps: []batches.Step{
			{Run: "echo 'Hello World'", Container: "alpine:3"},
		},
	}}

	value := executionResult{
		Diff: testDiff,
		ChangedFiles: &git.Changes{
			Added: []string{"README.md"},
		},
		Outputs: map[string]interface{}{},
	}

	t.Run("cache contains v3 cache file", func(t *testing.T) {
		cache := ExecutionDiskCache{Dir: cacheTmpDir(t)}

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
	})
}

func TestSortCacheFiles(t *testing.T) {
	tests := []struct {
		paths []string
		want  []string
	}{
		{
			paths: []string{"file.v3.json", "file.diff", "file.json"},
			want:  []string{"file.v3.json", "file.diff", "file.json"},
		},
		{
			paths: []string{"file.json", "file.diff", "file.v3.json"},
			want:  []string{"file.v3.json", "file.json", "file.diff"},
		},
		{
			paths: []string{"file.diff", "file.v3.json"},
			want:  []string{"file.v3.json", "file.diff"},
		},
		{
			paths: []string{"file1.v3.json", "file2.v3.json"},
			want:  []string{"file1.v3.json", "file2.v3.json"},
		},
	}

	for _, tt := range tests {
		sortCacheFiles(tt.paths)
		if diff := cmp.Diff(tt.paths, tt.want); diff != "" {
			t.Errorf("wrong cached result (-have +want):\n\n%s", diff)
		}
	}
}

func assertFileDeleted(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("file exists: %s", path)
	} else if os.IsNotExist(err) {
		// Seems to be deleted, all good
	} else {
		t.Fatalf("could not determine whether file exists: %s", err)
	}
}

func writeV1CacheFile(t *testing.T, c ExecutionDiskCache, k ExecutionCacheKey, diff string) (path string) {
	t.Helper()

	hashedKey, err := k.Key()
	if err != nil {
		t.Fatalf("failed to hash cacheKey: %s", err)
	}
	// The v1 file format ended in .json
	path = filepath.Join(c.Dir, hashedKey+".json")

	// v1 contained a fully serialized ChangesetSpec
	spec := batches.ChangesetSpec{CreatedChangeset: &batches.CreatedChangeset{
		Commits: []batches.GitCommitDescription{
			{Diff: testDiff},
		},
	}}

	raw, err := json.Marshal(&spec)
	if err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(path, raw, 0600); err != nil {
		t.Fatalf("writing the cache file failed: %s", err)
	}

	return path
}

func writeV2CacheFile(t *testing.T, c ExecutionDiskCache, k ExecutionCacheKey, diff string) (path string) {
	t.Helper()

	hashedKey, err := k.Key()
	if err != nil {
		t.Fatalf("failed to hash cacheKey: %s", err)
	}

	// The v2 file format ended in .json
	path = filepath.Join(c.Dir, hashedKey+".diff")

	// v2 contained only a diff
	if err := ioutil.WriteFile(path, []byte(diff), 0600); err != nil {
		t.Fatalf("writing the cache file failed: %s", err)
	}

	return path
}

func assertCacheHit(t *testing.T, c ExecutionDiskCache, k ExecutionCacheKey, want executionResult) {
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

func assertCacheMiss(t *testing.T, c ExecutionDiskCache, k ExecutionCacheKey) {
	t.Helper()

	_, found, err := c.Get(context.Background(), k)
	if err != nil {
		t.Fatalf("cache.Get returned unexpected error: %s", err)
	}
	if found {
		t.Fatalf("cache hit when miss was expected")
	}
}
