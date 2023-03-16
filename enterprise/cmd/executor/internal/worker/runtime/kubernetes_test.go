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

func TestKubernetesRuntime_Name(t *testing.T) {
	r := kubernetesRuntime{}
	assert.Equal(t, "kubernetes", string(r.Name()))
}

func TestKubernetesRuntime_NewRunnerSpecs(t *testing.T) {
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
					Key:       "step.kubernetes.key-1",
					Command:   []string{"/bin/sh", "-c", "/data/.sourcegraph-executor/script.sh"},
					Dir:       ".",
					Env:       []string{"FOO=bar"},
					Operation: operations.Exec,
				},
				Image: "my-image",
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
						Key:       "step.kubernetes.key-1",
						Command:   []string{"/bin/sh", "-c", "/data/.sourcegraph-executor/script1.sh"},
						Dir:       ".",
						Env:       []string{"FOO=bar"},
						Operation: operations.Exec,
					},
					Image: "my-image",
				},
				{
					CommandSpec: command.Spec{
						Key:       "step.kubernetes.key-2",
						Command:   []string{"/bin/sh", "-c", "/data/.sourcegraph-executor/script2.sh"},
						Dir:       ".",
						Env:       []string{"FOO=bar"},
						Operation: operations.Exec,
					},
					Image: "my-image",
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
					Key:       "step.kubernetes.0",
					Command:   []string{"/bin/sh", "-c", "/data/.sourcegraph-executor/script.sh"},
					Dir:       ".",
					Env:       []string{"FOO=bar"},
					Operation: operations.Exec,
				},
				Image: "my-image",
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

			r := &kubernetesRuntime{operations: operations}
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
