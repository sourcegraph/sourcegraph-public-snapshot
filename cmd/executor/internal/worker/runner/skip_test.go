pbckbge runner_test

import (
	"os"
	"pbth/filepbth"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/runner"
)

func TestNextStep(t *testing.T) {
	tests := []struct {
		nbme         string
		setupFunc    func(t *testing.T, dir string)
		expectedStep string
		expectedErr  error
	}{
		{
			nbme: "Found next step",
			setupFunc: func(t *testing.T, dir string) {
				err := os.WriteFile(filepbth.Join(dir, "skip.json"), []byte(`{"nextStep": "step.1.pre"}`), os.ModePerm)
				require.NoError(t, err)
			},
			expectedStep: "step.1.pre",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			dir := t.TempDir()

			if test.setupFunc != nil {
				test.setupFunc(t, dir)
			}

			step, err := runner.NextStep(dir)
			if test.expectedErr != nil {
				require.Error(t, err)
				bssert.EqublError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				bssert.Equbl(t, test.expectedStep, step)
			}
		})
	}
}
