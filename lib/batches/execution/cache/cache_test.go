package cache

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var repo = batches.Repository{
	ID:          "my-repo",
	Name:        "github.com/sourcegraph/src-cli",
	BaseRef:     "refs/heads/f00b4r",
	BaseRev:     "c0mmit",
	FileMatches: []string{"baz.go"},
}

func TestKeyer_Key(t *testing.T) {
	var singleStepEnv env.Environment
	err := json.Unmarshal([]byte(`{"FOO": "BAR"}`), &singleStepEnv)
	require.NoError(t, err)

	var multipleStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "BAZ": "FAZ"}`), &multipleStepEnv)
	require.NoError(t, err)

	var nullStepEnv env.Environment
	err = json.Unmarshal([]byte(`{"FOO": "BAR", "TEST_EXECUTION_CACHE_KEY_ENV": null}`), &nullStepEnv)
	require.NoError(t, err)

	var stepEnv env.Environment
	// use an array to get the key to have a nil value
	err = json.Unmarshal([]byte(`["SOME_ENV"]`), &stepEnv)
	require.NoError(t, err)

	modDate := time.Date(2022, 1, 2, 3, 5, 6, 7, time.UTC)

	tests := []struct {
		name          string
		keyer         Keyer
		expectedKey   string
		expectedError error
	}{
		{
			name: "ExecutionKey simple",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "ExecutionKey multiple steps",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps: []batches.Step{
					{Run: "foo"},
					{Run: "bar"},
				},
			},
			expectedKey: "nXrDA5Sv3jE2wGVTrixgJw",
		},
		{
			name: "ExecutionKey step env",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: singleStepEnv}},
			},
			expectedKey: "Ye3eFDmvvADzZuz-TWEA2g",
		},
		{
			name: "ExecutionKey multiple step envs",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: multipleStepEnv}},
			},
			expectedKey: "mZk8q7zjJioxI2nTwrt7XQ",
		},
		{
			name: "ExecutionKey null step env",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: nullStepEnv}},
			},
			expectedKey: "_txGuv3XrkWWVQz6hGsKhw",
		},
		{
			name: "ExecutionKey mount metadata",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
				MetadataRetriever: testM{
					m: []MountMetadata{{Path: "/foo/bar", Size: 100, Modified: modDate}},
				},
			},
			expectedKey: "DFPxThKpLG4BAqW_wMGLTQ",
		},
		{
			name: "ExecutionKey multiple mount metadata",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
				MetadataRetriever: testM{
					m: []MountMetadata{
						{Path: "/foo/bar", Size: 100, Modified: modDate},
						{Path: "/faz/baz", Size: 100, Modified: modDate},
					},
				},
			},
			expectedKey: "FhHlnyPvsOe9__15wfuQYQ",
		},
		{
			name: "ExecutionKey mount metadata error",
			keyer: &ExecutionKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
				MetadataRetriever: testM{
					err: errors.New("failed to get mount metadata"),
				},
			},
			expectedError: errors.New("failed to get mount metadata"),
		},
		{
			name: "ExecutionKeyWithGlobalEnv simple",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo"}},
				},
				GlobalEnv: []string{},
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g",
		},
		{
			name: "ExecutionKeyWithGlobalEnv has global env",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedKey: "UWaad_y5HkY90tPkgBO7og",
		},
		{
			name: "ExecutionKeyWithGlobalEnv env not updated",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"FAZ=BAZ"},
			},
			expectedKey: "tq9NsiMdvoKqMpgxE00XGQ",
		},
		{
			name: "ExecutionKeyWithGlobalEnv malformed global env",
			keyer: &ExecutionKeyWithGlobalEnv{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				},
				GlobalEnv: []string{"SOME_ENV"},
			},
			expectedError: errors.New("resolving environment for step 0: unable to parse environment variable \"SOME_ENV\""),
		},
		{
			name: "StepsCacheKey simple",
			keyer: StepsCacheKey{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps:      []batches.Step{{Run: "foo"}},
				},
				StepIndex: 0,
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g-step-0",
		},
		{
			name: "StepsCacheKey multiple steps",
			keyer: StepsCacheKey{
				ExecutionKey: &ExecutionKey{
					Repository: repo,
					Steps: []batches.Step{
						{Run: "foo"},
						{Run: "bar"},
					},
				},
				StepIndex: 0,
			},
			expectedKey: "cu8r-xdguU4s0kn9_uxL5g-step-0",
		},
		{
			name: "StepsCacheKeyWithGlobalEnv env set",
			keyer: &StepsCacheKeyWithGlobalEnv{
				StepsCacheKey: &StepsCacheKey{
					ExecutionKey: &ExecutionKey{
						Repository: repo,
						Steps: []batches.Step{
							{
								Run: "foo",
								Env: stepEnv,
							},
						},
					},
					StepIndex: 0,
				},
				GlobalEnv: []string{"SOME_ENV=FOO", "FAZ=BAZ"},
			},
			expectedKey: "UWaad_y5HkY90tPkgBO7og-step-0",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			key, err := test.keyer.Key()
			if test.expectedError != nil {
				assert.Equal(t, test.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expectedKey, key)
			}
		})
	}
}

type testM struct {
	m   []MountMetadata
	err error
}

func (t testM) Get(steps []batches.Step) ([]MountMetadata, error) {
	return t.m, t.err
}
