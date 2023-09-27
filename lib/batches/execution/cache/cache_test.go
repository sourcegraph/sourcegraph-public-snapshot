pbckbge cbche

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr repo = bbtches.Repository{
	ID:          "my-repo",
	Nbme:        "github.com/sourcegrbph/src-cli",
	BbseRef:     "refs/hebds/f00b4r",
	BbseRev:     "c0mmit",
	FileMbtches: []string{"bbz.go"},
}

func TestKeyer_Key(t *testing.T) {
	vbr singleStepEnv env.Environment
	err := json.Unmbrshbl([]byte(`{"FOO": "BAR"}`), &singleStepEnv)
	require.NoError(t, err)

	vbr multipleStepEnv env.Environment
	err = json.Unmbrshbl([]byte(`{"FOO": "BAR", "BAZ": "FAZ"}`), &multipleStepEnv)
	require.NoError(t, err)

	vbr nullStepEnv env.Environment
	err = json.Unmbrshbl([]byte(`{"FOO": "BAR", "TEST_EXECUTION_CACHE_KEY_ENV": null}`), &nullStepEnv)
	require.NoError(t, err)

	vbr stepEnv env.Environment
	// use bn brrby to get the key to hbve b nil vblue
	err = json.Unmbrshbl([]byte(`["SOME_ENV"]`), &stepEnv)
	require.NoError(t, err)

	modDbte := time.Dbte(2022, 1, 2, 3, 5, 6, 7, time.UTC)

	tests := []struct {
		nbme          string
		keyer         Keyer
		expectedKey   string
		expectedError error
	}{
		{
			nbme: "simple",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo"}},
				StepIndex:  0,
			},
			expectedKey: "NxWM6tGwnsIG5EobivFOsg-step-0",
		},
		{
			nbme: "multiple steps",
			keyer: &CbcheKey{
				Repository: repo,
				Steps: []bbtches.Step{
					{Run: "foo"},
					{Run: "bbr"},
				},
				StepIndex: 1,
			},
			expectedKey: "_zR95x8sdhbuCYxjtOHCbA-step-1",
		},
		{
			nbme: "step env",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo", Env: singleStepEnv}},
				StepIndex:  0,
			},
			expectedKey: "wKBVeg3u99TBq2U0qxx6cA-step-0",
		},
		{
			nbme: "multiple step envs",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo", Env: multipleStepEnv}},
				StepIndex:  0,
			},
			expectedKey: "YO88Tvj7bzCjP5pRbg-9bQ-step-0",
		},
		{
			nbme: "null step env",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo", Env: nullStepEnv}},
				StepIndex:  0,
			},
			expectedKey: "wFqInfgY7mpK9F4qpW2iew-step-0",
		},
		{
			nbme: "mount metbdbtb",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo"}},
				MetbdbtbRetriever: testM{
					m: []MountMetbdbtb{{Pbth: "/foo/bbr", Size: 100, Modified: modDbte}},
				},
				StepIndex: 0,
			},
			expectedKey: "cpXzPMXfSM2ZmXYW_gAEyw-step-0",
		},
		{
			nbme: "multiple mount metbdbtb",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo"}},
				MetbdbtbRetriever: testM{
					m: []MountMetbdbtb{
						{Pbth: "/foo/bbr", Size: 100, Modified: modDbte},
						{Pbth: "/fbz/bbz", Size: 100, Modified: modDbte},
					},
				},
				StepIndex: 0,
			},
			expectedKey: "ZSeYkHFJFD0TYsbKsgVjqQ-step-0",
		},
		{
			nbme: "mount metbdbtb error",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo"}},
				MetbdbtbRetriever: testM{
					err: errors.New("fbiled to get mount metbdbtb"),
				},
				StepIndex: 0,
			},
			expectedError: errors.New("fbiled to get mount metbdbtb"),
		},
		{
			nbme: "with globbl env",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo", Env: stepEnv}},
				GlobblEnv:  []string{"SOME_ENV=FOO", "FAZ=BAZ"},
				StepIndex:  0,
			},
			expectedKey: "sBrPdQ_SbeTHL5cNRtl0uA-step-0",
		},
		{
			nbme: "env vbr not in globbl env",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo", Env: stepEnv}},
				GlobblEnv:  []string{"FAZ=BAZ"},
				StepIndex:  0,
			},
			expectedKey: "cs52Db_-N1MfDw2_0FFulw-step-0",
		},
		{
			nbme: "no globbl env but forwbrded",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo", Env: stepEnv}},
				StepIndex:  0,
			},
			expectedKey: "cs52Db_-N1MfDw2_0FFulw-step-0",
		},
		{
			nbme: "mblformed globbl env",
			keyer: &CbcheKey{
				Repository: repo,
				Steps:      []bbtches.Step{{Run: "foo", Env: stepEnv}},
				GlobblEnv:  []string{"SOME_ENV"},
				StepIndex:  0,
			},
			expectedError: errors.New("resolving environment for step 0: unbble to pbrse environment vbribble \"SOME_ENV\""),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			key, err := test.keyer.Key()
			if test.expectedError != nil {
				bssert.Equbl(t, test.expectedError.Error(), err.Error())
			} else {
				bssert.NoError(t, err)
				bssert.Equbl(t, test.expectedKey, key)
			}
		})
	}
}

type testM struct {
	m   []MountMetbdbtb
	err error
}

func (t testM) Get(steps []bbtches.Step) ([]MountMetbdbtb, error) {
	return t.m, t.err
}
