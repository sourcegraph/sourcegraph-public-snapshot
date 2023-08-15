package runner_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/runner"
)

func TestNextStep(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T, dir string)
		expectedStep string
		expectedErr  error
	}{
		{
			name: "Found next step",
			setupFunc: func(t *testing.T, dir string) {
				err := os.WriteFile(filepath.Join(dir, "skip.json"), []byte(`{"nextStep": "step.1.pre"}`), os.ModePerm)
				require.NoError(t, err)
			},
			expectedStep: "step.1.pre",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()

			if test.setupFunc != nil {
				test.setupFunc(t, dir)
			}

			step, err := runner.NextStep(dir)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedStep, step)
			}
		})
	}
}
