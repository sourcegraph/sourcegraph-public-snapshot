package worker

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/runner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestHandler_PreDequeue(t *testing.T) {
	logger := logtest.Scoped(t)

	tests := []struct {
		name              string
		options           Options
		mockFunc          func(cmdRunner *MockCmdRunner)
		expectedDequeue   bool
		expectedExtraArgs any
		expectedErr       error
		assertMockFunc    func(t *testing.T, cmdRunner *MockCmdRunner)
	}{
		{
			name: "Firecracker not enabled",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrackerOptions: runner.FirecrackerOptions{Enabled: false},
				},
			},
			expectedDequeue: true,
			assertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
			},
		},
		{
			name: "Firecracker enabled",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrackerOptions: runner.FirecrackerOptions{Enabled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHandlers: 1},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn([]byte{}, nil)
			},
			expectedDequeue: true,
			assertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
				assert.Equal(t, "ignite", cmdRunner.CombinedOutputFunc.History()[0].Arg1)
				assert.Equal(
					t,
					[]string{"ps", "-t", "{{ .Name }}:{{ .UID }}"},
					cmdRunner.CombinedOutputFunc.History()[0].Arg2,
				)
			},
		},
		{
			name: "Orphaned VMs",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrackerOptions: runner.FirecrackerOptions{Enabled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHandlers: 1},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn([]byte("foo:bar\nfaz:baz"), nil)
			},
			assertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
			},
		},
		{
			name: "Less Orphaned VMs than Handlers",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrackerOptions: runner.FirecrackerOptions{Enabled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHandlers: 3},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn([]byte("foo:bar\nfaz:baz"), nil)
			},
			expectedDequeue: true,
			assertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
			},
		},
		{
			name: "Failed to get active VMs",
			options: Options{
				RunnerOptions: runner.Options{
					FirecrackerOptions: runner.FirecrackerOptions{Enabled: true},
				},
				WorkerOptions: workerutil.WorkerOptions{NumHandlers: 3},
			},
			mockFunc: func(cmdRunner *MockCmdRunner) {
				cmdRunner.CombinedOutputFunc.PushReturn(nil, errors.New("failed"))
			},
			expectedErr: errors.New("failed"),
			assertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 1)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmdRunner := NewMockCmdRunner()

			h := &handler{
				cmdRunner: cmdRunner,
				options:   test.options,
			}

			if test.mockFunc != nil {
				test.mockFunc(cmdRunner)
			}

			dequeueable, extraArgs, err := h.PreDequeue(context.Background(), logger)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
				assert.Equal(t, test.expectedDequeue, dequeueable)
				assert.Equal(t, test.expectedExtraArgs, extraArgs)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedDequeue, dequeueable)
				assert.Equal(t, test.expectedExtraArgs, extraArgs)
			}

			test.assertMockFunc(t, cmdRunner)
		})
	}
}

func TestHandler_Handle(t *testing.T) {
	internalLogger := logtest.Scoped(t)
	operations := command.NewOperations(&observation.TestContext)

	tests := []struct {
		name           string
		options        Options
		job            types.Job
		mockFunc       func(cmdRunner *MockCmdRunner, command *MockCommand, logStore *MockExecutionLogEntryStore, filesStore *MockFilesStore)
		expectedErr    error
		assertMockFunc func(t *testing.T, cmdRunner *MockCmdRunner, command *MockCommand, logStore *MockExecutionLogEntryStore, filesStore *MockFilesStore)
	}{
		{
			name:    "Legacy Success with no steps",
			options: Options{},
			job:     types.Job{ID: 42, RepositoryName: "my-repo", Commit: "cool-commit"},
			mockFunc: func(cmdRunner *MockCmdRunner, cmd *MockCommand, logStore *MockExecutionLogEntryStore, filesStore *MockFilesStore) {
				// Since things will get complicated, it will be much better to push returns instead of setting a default.
				// This will allow easier copy-paste for other tests/scenarios.
			},
			assertMockFunc: func(t *testing.T, cmdRunner *MockCmdRunner, cmd *MockCommand, logStore *MockExecutionLogEntryStore, filesStore *MockFilesStore) {
				require.Len(t, cmdRunner.CombinedOutputFunc.History(), 0)
				require.Len(t, cmd.RunFunc.History(), 0)
				require.Len(t, logStore.AddExecutionLogEntryFunc.History(), 0)
				require.Len(t, filesStore.GetFunc.History(), 0)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Used in prepareWorkspace
			cmdRunner := NewMockCmdRunner()
			// Used in prepareWorkspace, runner
			cmd := NewMockCommand()
			// Used in NewLogger
			logStore := NewMockExecutionLogEntryStore()
			// Used in prepareWorkspace
			filesStore := NewMockFilesStore()

			h := &handler{
				cmdRunner:  cmdRunner,
				cmd:        cmd,
				logStore:   logStore,
				filesStore: filesStore,
				options:    test.options,
				operations: operations,
			}

			if test.mockFunc != nil {
				test.mockFunc(cmdRunner, cmd, logStore, filesStore)
			}

			err := h.Handle(context.Background(), internalLogger, test.job)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			test.assertMockFunc(t, cmdRunner, cmd, logStore, filesStore)
		})
	}
}
