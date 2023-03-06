package runtime_test

import (
	"os/exec"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runtime"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
)

func TestNewRuntime(t *testing.T) {
	tests := []struct {
		name           string
		mockFunc       func(cmdRunner *runtime.MockCmdRunner)
		expectedName   runtime.Name
		hasError       bool
		assertMockFunc func(t *testing.T, cmdRunner *runtime.MockCmdRunner)
	}{
		{
			name: "Docker",
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPathFunc.SetDefaultReturn("", nil)
			},
			expectedName: runtime.NameDocker,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 3)
				assert.Equal(t, "docker", cmdRunner.LookPathFunc.History()[0].Arg0)
				assert.Equal(t, "git", cmdRunner.LookPathFunc.History()[1].Arg0)
				assert.Equal(t, "src", cmdRunner.LookPathFunc.History()[2].Arg0)
			},
		},
		{
			name: "No Runtime",
			mockFunc: func(cmdRunner *runtime.MockCmdRunner) {
				cmdRunner.LookPathFunc.PushReturn("", exec.ErrNotFound)
			},
			hasError: true,
			assertMockFunc: func(t *testing.T, cmdRunner *runtime.MockCmdRunner) {
				require.Len(t, cmdRunner.LookPathFunc.History(), 3)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmdRunner := runtime.NewMockCmdRunner()
			if test.mockFunc != nil {
				test.mockFunc(cmdRunner)
			}
			logger := logtest.Scoped(t)
			// Most of the arguments can be nil/empty since we are not doing anything with them
			r, err := runtime.New(
				logger,
				nil,
				nil,
				workspace.CloneOptions{},
				runner.Options{},
				cmdRunner,
				nil,
			)
			if test.hasError {
				require.Error(t, err)
				assert.Nil(t, r)
				assert.ErrorIs(t, err, runtime.ErrNoRuntime)
			} else {
				require.NoError(t, err)
				require.NotNil(t, r)
				assert.Equal(t, test.expectedName, r.Name())
			}

			test.assertMockFunc(t, cmdRunner)
		})
	}
}
