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
			name: "simple",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
				StepIndex:  0,
			},
			expectedKey: "NxWM6tGwnsIG5EoaivFOsg-step-0",
		},
		{
			name: "multiple steps",
			keyer: &CacheKey{
				Repository: repo,
				Steps: []batches.Step{
					{Run: "foo"},
					{Run: "bar"},
				},
				StepIndex: 1,
			},
			expectedKey: "_zR95x8sdhauCYxjtOHCbA-step-1",
		},
		{
			name: "step env",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: singleStepEnv}},
				StepIndex:  0,
			},
			expectedKey: "wKBVeg3u99TBq2U0qxx6cA-step-0",
		},
		{
			name: "multiple step envs",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: multipleStepEnv}},
				StepIndex:  0,
			},
			expectedKey: "YO88Tvj7bzCjP5pRag-9bQ-step-0",
		},
		{
			name: "null step env",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: nullStepEnv}},
				StepIndex:  0,
			},
			expectedKey: "wFqInfgY7mpK9F4qpW2iew-step-0",
		},
		{
			name: "mount metadata",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
				MetadataRetriever: testM{
					m: []MountMetadata{{Path: "/foo/bar", Size: 100, Modified: modDate}},
				},
				StepIndex: 0,
			},
			expectedKey: "cpXzPMXfSM2ZmXYW_gAEyw-step-0",
		},
		{
			name: "multiple mount metadata",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
				MetadataRetriever: testM{
					m: []MountMetadata{
						{Path: "/foo/bar", Size: 100, Modified: modDate},
						{Path: "/faz/baz", Size: 100, Modified: modDate},
					},
				},
				StepIndex: 0,
			},
			expectedKey: "ZSeYkHFJFD0TYsbKsgVjqQ-step-0",
		},
		{
			name: "mount metadata error",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo"}},
				MetadataRetriever: testM{
					err: errors.New("failed to get mount metadata"),
				},
				StepIndex: 0,
			},
			expectedError: errors.New("failed to get mount metadata"),
		},
		{
			name: "with global env",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				GlobalEnv:  []string{"SOME_ENV=FOO", "FAZ=BAZ"},
				StepIndex:  0,
			},
			expectedKey: "sBrPdQ_SaeTHL5cNRtl0uA-step-0",
		},
		{
			name: "env var not in global env",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				GlobalEnv:  []string{"FAZ=BAZ"},
				StepIndex:  0,
			},
			expectedKey: "cs52Db_-N1MfDw2_0FFulw-step-0",
		},
		{
			name: "no global env but forwarded",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				StepIndex:  0,
			},
			expectedKey: "cs52Db_-N1MfDw2_0FFulw-step-0",
		},
		{
			name: "malformed global env",
			keyer: &CacheKey{
				Repository: repo,
				Steps:      []batches.Step{{Run: "foo", Env: stepEnv}},
				GlobalEnv:  []string{"SOME_ENV"},
				StepIndex:  0,
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
