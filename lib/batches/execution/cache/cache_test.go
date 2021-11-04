package cache

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/batches"
)

const testExecutionCacheKeyEnv = "TEST_EXECUTION_CACHE_KEY_ENV"

func TestExecutionKey(t *testing.T) {
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
	key := ExecutionKey{
		Repository: batches.Repository{
			ID:          "graphql-id",
			Name:        "github.com/sourcegraph/src-cli",
			BaseRef:     "refs/heads/f00b4r",
			BaseRev:     "c0mmit",
			FileMatches: []string{"aa.go"},
		},
		Steps: steps,
	}

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
	if initial != have {
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
	if initial == have {
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
	if initial == again || have == again {
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
	if initial != have {
		t.Errorf("unexpected change in key: initial=%q have=%q", initial, have)
	}
}
