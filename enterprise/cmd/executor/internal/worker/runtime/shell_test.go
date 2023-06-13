package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestShellRuntime_Name(t *testing.T) {
	r := shellRuntime{}
	assert.Equal(t, "shell", string(r.Name()))
}

func TestShellRuntime_NewRunnerSpecs(t *testing.T) {
	operations := command.NewOperations(&observation.TestContext)

	tests := []struct {
		name           string
		steps          []types.DockerStep
		mockFunc       func(ws *MockWorkspace)
		expected       []runner.Spec
		expectedErr    error
		assertMockFunc func(t *testing.T, ws *MockWorkspace)
	}{
		{
			name:     "No steps",
			steps:    []types.DockerStep{},
			expected: []runner.Spec{},
			assertMockFunc: func(t *testing.T, ws *MockWorkspace) {
				require.Len(t, ws.ScriptFilenamesFunc.History(), 0)
			},
		},
		{
			name: "Single step",
			steps: []types.DockerStep{
				{
					Key:      "key-1",
					Image:    "my-image",
					Commands: []string{"echo", "hello"},
					Dir:      ".",
					Env:      []string{"FOO=bar"},
				},
			},
			mockFunc: func(ws *MockWorkspace) {
				ws.ScriptFilenamesFunc.SetDefaultReturn([]string{"script.sh"})
			},
			expected: []runner.Spec{{
				CommandSpec: command.Spec{
					Key:       "step.docker.key-1",
					Command:   []string(nil),
					Dir:       ".",
					Env:       []string{"FOO=bar"},
					Operation: operations.Exec,
				},
				Image:      "my-image",
				ScriptPath: "script.sh",
			}},
			assertMockFunc: func(t *testing.T, ws *MockWorkspace) {
				require.Len(t, ws.ScriptFilenamesFunc.History(), 1)
			},
		},
		{
			name: "Multiple steps",
			steps: []types.DockerStep{
				{
					Key:      "key-1",
					Image:    "my-image",
					Commands: []string{"echo", "hello"},
					Dir:      ".",
					Env:      []string{"FOO=bar"},
				},
				{
					Key:      "key-2",
					Image:    "my-image",
					Commands: []string{"echo", "hello"},
					Dir:      ".",
					Env:      []string{"FOO=bar"},
				},
			},
			mockFunc: func(ws *MockWorkspace) {
				ws.ScriptFilenamesFunc.SetDefaultReturn([]string{"script1.sh", "script2.sh"})
			},
			expected: []runner.Spec{
				{
					CommandSpec: command.Spec{
						Key:       "step.docker.key-1",
						Command:   []string(nil),
						Dir:       ".",
						Env:       []string{"FOO=bar"},
						Operation: operations.Exec,
					},
					Image:      "my-image",
					ScriptPath: "script1.sh",
				},
				{
					CommandSpec: command.Spec{
						Key:       "step.docker.key-2",
						Command:   []string(nil),
						Dir:       ".",
						Env:       []string{"FOO=bar"},
						Operation: operations.Exec,
					},
					Image:      "my-image",
					ScriptPath: "script2.sh",
				},
			},
			assertMockFunc: func(t *testing.T, ws *MockWorkspace) {
				require.Len(t, ws.ScriptFilenamesFunc.History(), 2)
			},
		},
		{
			name: "Default key",
			steps: []types.DockerStep{
				{
					Image:    "my-image",
					Commands: []string{"echo", "hello"},
					Dir:      ".",
					Env:      []string{"FOO=bar"},
				},
			},
			mockFunc: func(ws *MockWorkspace) {
				ws.ScriptFilenamesFunc.SetDefaultReturn([]string{"script.sh"})
			},
			expected: []runner.Spec{{
				CommandSpec: command.Spec{
					Key:       "step.docker.0",
					Command:   []string(nil),
					Dir:       ".",
					Env:       []string{"FOO=bar"},
					Operation: operations.Exec,
				},
				Image:      "my-image",
				ScriptPath: "script.sh",
			}},
			assertMockFunc: func(t *testing.T, ws *MockWorkspace) {
				require.Len(t, ws.ScriptFilenamesFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ws := NewMockWorkspace()

			if test.mockFunc != nil {
				test.mockFunc(ws)
			}

			r := &shellRuntime{operations: operations}
			actual, err := r.NewRunnerSpecs(ws, test.steps)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expected, actual)
			}

			test.assertMockFunc(t, ws)
		})
	}
}
