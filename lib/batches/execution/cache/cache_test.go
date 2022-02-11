package cache

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

const testExecutionCacheKeyEnv = "TEST_EXECUTION_CACHE_KEY_ENV"

func TestExecutionKey_RegressionTest(t *testing.T) {
	// This test is a regression that should fail when we change something that
	// influences the cache key generation, which would lead to busted caches.
	//
	// If this test fails and you're sure about the change, update the `want`
	// value below.

	var steps []batches.Step
	if err := yaml.Unmarshal([]byte(`
- run:  if [[ -f "package.json" ]]; then cat package.json | jq -j .name; fi
  container: jiapantw/jq-alpine:latest
  outputs:
    projectName:
      value: ${{ step.stdout }}
- run:  echo "This only runs in automation-testing" >> message.txt
  container: alpine:3
  if: ${{ eq repository.name "github.com/sourcegraph/automation-testing" }}
- run: bar
  container: alpine:3
  env:
    - FILE_TO_CHECK: .tool-versions
    - `+testExecutionCacheKeyEnv+`
`), &steps); err != nil {
		t.Fatal(err)
	}

	key := ExecutionKey{
		Repository: batches.Repository{
			ID:          "graphql-id",
			Name:        "github.com/sourcegraph/src-cli",
			BaseRef:     "refs/heads/f00b4r",
			BaseRev:     "c0mmit",
			FileMatches: []string{"aa.go"},
		},
		Path:               "path/to/workspace",
		OnlyFetchWorkspace: true,
		Steps:              steps,
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        "Batch Change Name",
			Description: "Batch Change Description",
		},
	}

	have, err := key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	want := "fsDKj1Uf1jNMhRXCJXE6nQ"
	if have != want {
		t.Fatalf("regression detected! cache key changed. have=%q, want=%q", have, want)
	}
}

func TestExecutionKeyWithEnvResolution(t *testing.T) {
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
	key := ExecutionKeyWithGlobalEnv{
		ExecutionKey: &ExecutionKey{
			Repository: batches.Repository{
				ID:          "graphql-id",
				Name:        "github.com/sourcegraph/src-cli",
				BaseRef:     "refs/heads/f00b4r",
				BaseRev:     "c0mmit",
				FileMatches: []string{"aa.go"},
			},
			Steps: steps,
		},
		GlobalEnv: os.Environ(),
	}

	// All righty. Let's get ourselves a baseline cache key here.
	initial, err := key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Let's set an unrelated environment variable and ensure we still have the
	// same key.
	key.GlobalEnv = append(key.GlobalEnv, fmt.Sprintf("%s=%s", testExecutionCacheKeyEnv+"_UNRELATED", "foo"))
	have, err := key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if initial != have {
		t.Errorf("unexpected change in key: initial=%q have=%q", initial, have)
	}

	// Let's now set the environment variable referenced in the steps and verify
	// that the cache key does change.
	key.GlobalEnv = append(key.GlobalEnv, fmt.Sprintf("%s=%s", testExecutionCacheKeyEnv, "foo"))
	have, err = key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if initial == have {
		t.Errorf("unexpected lack of change in key: %q", have)
	}

	// And, just to be sure, let's change it again.
	key.GlobalEnv[len(key.GlobalEnv)-1] = fmt.Sprintf("%s=%s", testExecutionCacheKeyEnv, "bar")
	again, err := key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if initial == again || have == again {
		t.Errorf("unexpected lack of change in key: %q", again)
	}

	// Finally, if we unset the environment variable again, we should get a key
	// that matches the initial key.
	key.GlobalEnv = key.GlobalEnv[:len(key.GlobalEnv)-1]
	have, err = key.Key()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if initial != have {
		t.Errorf("unexpected change in key: initial=%q have=%q", initial, have)
	}
}
