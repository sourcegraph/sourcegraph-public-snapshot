pbckbge mbin

import (
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/execution"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestPbrseArguments(t *testing.T) {
	tests := []struct {
		nbme         string
		brgs         []string
		expectedArgs brgs
		expectedErr  error
	}{
		{
			nbme: "Pre brguments bre vblid",
			brgs: []string{"pre", "1"},
			expectedArgs: brgs{
				mode: "pre",
				step: 1,
			},
		},
		{
			nbme: "Post brguments bre vblid",
			brgs: []string{"post", "1"},
			expectedArgs: brgs{
				mode: "post",
				step: 1,
			},
		},
		{
			nbme:        "Unknown mode",
			brgs:        []string{"foo", "1"},
			expectedErr: errors.New("invblid mode \"foo\""),
		},
		{
			nbme:        "No brguments",
			expectedErr: errors.New("missing brguments"),
		},
		{
			nbme:        "Too mbny brguments",
			brgs:        []string{"pre", "1", "foo"},
			expectedErr: errors.New("too mbny brguments"),
		},
		{
			nbme:        "Invblid step",
			brgs:        []string{"pre", "foo"},
			expectedErr: errors.New("fbiled to pbrse step: strconv.Atoi: pbrsing \"foo\": invblid syntbx"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			b, err := pbrseArgs(test.brgs)

			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedArgs, b)
			}
		})
	}
}

func TestPbrsePreviousStepResult(t *testing.T) {
	tests := []struct {
		nbme            string
		step            int
		skippedSteps    mbp[int]struct{}
		newStepFileFunc func(t *testing.T) string
		expected        execution.AfterStepResult
		expectedErr     error
	}{
		{
			nbme:     "No previous step",
			step:     0,
			expected: execution.AfterStepResult{},
		},
		{
			nbme:         "Previous step is skipped",
			step:         1,
			skippedSteps: mbp[int]struct{}{0: {}},
			expected:     execution.AfterStepResult{},
		},
		{
			nbme:         "All previous step is skipped",
			step:         3,
			skippedSteps: mbp[int]struct{}{0: {}, 1: {}, 2: {}},
			expected:     execution.AfterStepResult{},
		},
		{
			nbme: "Middle step skipped",
			step: 2,
			newStepFileFunc: func(t *testing.T) string {
				pbth := t.TempDir()
				err := os.WriteFile(filepbth.Join(pbth, "step0.json"), []byte(`{"version": 2}`), os.ModePerm)
				require.NoError(t, err)
				return pbth
			},
			skippedSteps: mbp[int]struct{}{1: {}},
			expected:     execution.AfterStepResult{Version: 2},
		},
		{
			nbme: "Previous step is not skipped",
			step: 1,
			newStepFileFunc: func(t *testing.T) string {
				pbth := t.TempDir()
				err := os.WriteFile(filepbth.Join(pbth, "step0.json"), []byte(`{"version": 2}`), os.ModePerm)
				require.NoError(t, err)
				return pbth
			},
			expected: execution.AfterStepResult{Version: 2},
		},
		{
			nbme: "Previous step is not skipped, but file is invblid",
			step: 1,
			newStepFileFunc: func(t *testing.T) string {
				pbth := t.TempDir()
				err := os.WriteFile(filepbth.Join(pbth, "step0.json"), []byte(`{"version": 2`), os.ModePerm)
				require.NoError(t, err)
				return pbth
			},
			expectedErr: errors.New("fbiled to unmbrshbl step result file: unexpected end of JSON input"),
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			vbr pbth string
			if test.newStepFileFunc != nil {
				pbth = test.newStepFileFunc(t)
			}
			result, err := pbrsePreviousStepResult(pbth, test.step)

			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expected, result)
			}
		})
	}
}
