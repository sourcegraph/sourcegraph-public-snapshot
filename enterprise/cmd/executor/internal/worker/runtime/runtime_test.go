package runtime_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runtime"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
)

func TestNewRuntime(t *testing.T) {
	tests := []struct {
		name         string
		mockFunc     func(runner *fakeCmdRunner)
		expectedName runtime.Name
		hasError     bool
	}{
		{
			name: "Docker",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").Return("", nil)
				runner.On("LookPath", "git").Return("", nil)
				runner.On("LookPath", "src").Return("", nil)
			},
			expectedName: runtime.NameDocker,
		},
		{
			name: "No Runtime",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("LookPath", "docker").Return("", exec.ErrNotFound)
				runner.On("LookPath", "git").Return("", exec.ErrNotFound)
				runner.On("LookPath", "src").Return("", exec.ErrNotFound)
			},
			hasError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := new(fakeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}
			logger := logtest.Scoped(t)
			// Most of the arguments can be nil/empty since we are not doing anything with them
			r, err := runtime.New(
				logger,
				nil,
				nil,
				workspace.CloneOptions{},
				command.DockerOptions{},
				runner,
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
		})
	}
}

type fakeCmdRunner struct {
	mock.Mock
}

func (f *fakeCmdRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	panic("not needed")
}

var _ util.CmdRunner = &fakeCmdRunner{}

func (f *fakeCmdRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	calledArgs := f.Called(ctx, name, args)
	return calledArgs.Get(0).([]byte), calledArgs.Error(1)
}

func (f *fakeCmdRunner) LookPath(file string) (string, error) {
	args := f.Called(file)
	return args.String(0), args.Error(1)
}
