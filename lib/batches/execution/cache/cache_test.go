package cache

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/env"
)

func TestExecutionKey_Key(t *testing.T) {
	var singleStepEnv env.Environment
	err := json.Unmarshal([]byte(`{"FOO": "BAR"}`), &singleStepEnv)
	require.NoError(t, err)

	var multipleStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "BAZ": "FAZ"}`), &multipleStepEnv)
	require.NoError(t, err)

	var nullStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "TEST_EXECUTION_CACHE_KEY_ENV": null}`), &nullStepEnv)
	require.NoError(t, err)

	tests := []struct {
		name          string
		keyer         ExecutionKey
		expectedKey   string
		expectedError error
	}{
		{
			name: "simple key",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo"}},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "multiple steps",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{
					{Run: "foo"},
					{Run: "bar"},
				},
			},
			expectedKey: "nXrDA5Sv3jE2wGVTrixgJw",
		},
		{
			name: "step env",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo", Env: singleStepEnv}},
			},
			expectedKey: "Ye3eFDmvvADzZuz-TWEA2g",
		},
		{
			name: "multiple step envs",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo", Env: multipleStepEnv}},
			},
			expectedKey: "mZk8q7zjJioxI2nTwrt7XQ",
		},
		{
			name: "null step env",
			keyer: ExecutionKey{
				Repository: batches.Repository{
					ID:          "my-repo",
					Name:        "github.com/sourcegraph/src-cli",
					BaseRef:     "refs/heads/f00b4r",
					BaseRev:     "c0mmit",
					FileMatches: []string{"baz.go"},
				},
				Steps: []batches.Step{{Run: "foo", Env: nullStepEnv}},
			},
			expectedKey: "_txGuv3XrkWWVQz6hGsKhw",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := test.keyer.Key()
			assert.ErrorIs(t, err, test.expectedError)
			assert.Equal(t, test.expectedKey, key)
		})
	}
}

func TestExecutionKeyWithGlobalEnv_Key(t *testing.T) {
	var stepEnv env.Environment
	// use an array to get the key to have a nil value
	err := json.Unmarshal([]byte(`["SOME_ENV"]`), &stepEnv)
	require.NoError(t, err)

	tests := []struct {
		name          string
		keyer         ExecutionKeyWithGlobalEnv
		expectedKey   string
		expectedError error
	}{
		{
			name: "simple key",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo"}},
				},
				GlobalEnv: []string{},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "has global env",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedKey: "UWaad_y5HkY90tPkgBO7og",
		},
		{
			name: "env not updated",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"FAZ=BAZ"},
			},
			expectedKey: "tq9NsiMdvoKqMpgxE00XGQ",
		},
		{
			name: "malformed global env",
			keyer: ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: batches.Repository{
						ID:          "my-repo",
						Name:        "github.com/sourcegraph/src-cli",
						BaseRef:     "refs/heads/f00b4r",
						BaseRev:     "c0mmit",
						FileMatches: []string{"baz.go"},
					},
					Steps: []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV"},
			},
			expectedError: errors.New("resolving environment for step 0: unable to parse environment variable \"SOME_ENV\""),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := test.keyer.Key()
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.Equal(t, test.expectedKey, key)
			}
		})
	}
}
