package command_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/cmdlogger"
	"github.com/sourcegraph/sourcegraph/cmd/executor/internal/worker/command"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestCommand_Run(t *testing.T) {
	internalLogger := logtest.Scoped(t)
	operations := command.NewOperations(observation.TestContextTB(t))

	tests := []struct {
		name         string
		command      []string
		mockExitCode int
		mockStdout   string
		mockFunc     func(t *testing.T, cmdRunner *fakeCmdRunner, logger *mockLogger)
		expectedErr  error
	}{
		{
			name:         "Success",
			command:      []string{"git", "pull"},
			mockExitCode: 0,
			mockStdout:   "got the stuff",
			mockFunc: func(t *testing.T, cmdRunner *fakeCmdRunner, logger *mockLogger) {
				logEntry := new(mockLogEntry)
				logger.
					On("LogEntry", "some-key", []string{"git", "pull"}).
					Return(logEntry)
				logEntry.On("Write", mock.Anything).Run(func(args mock.Arguments) {
					// Use Run to see the actual output in the test output. Else we just get byte output.
					actual := args.Get(0).([]byte)
					assert.Equal(t, "stdout: got the stuff\n", string(actual))
				}).Return(0, nil)
				logEntry.On("Finalize", 0).Return()
				logEntry.On("Close").Return(nil)
			},
		},
		{
			name:        "Invalid Command",
			command:     []string{"echo", "hello"},
			expectedErr: command.ErrIllegalCommand,
		},
		{
			name:         "Bad exit code",
			command:      []string{"git", "pull"},
			mockExitCode: 1,
			mockStdout:   "something went wrong",
			mockFunc: func(t *testing.T, cmdRunner *fakeCmdRunner, logger *mockLogger) {
				logEntry := new(mockLogEntry)
				logger.
					On("LogEntry", "some-key", []string{"git", "pull"}).
					Return(logEntry)
				logEntry.On("Write", mock.Anything).Run(func(args mock.Arguments) {
					// Use Run to see the actual output in the test output. Else we just get byte output.
					actual := args.Get(0).([]byte)
					assert.Equal(t, "stdout: something went wrong\n", string(actual))
				}).Return(0, nil)
				logEntry.On("Finalize", 1).Return()
				logEntry.On("Close").Return(nil)
			},
			expectedErr: errors.New("command failed with exit code 1"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmdRunner := new(fakeCmdRunner)
			logger := new(mockLogger)

			if test.mockFunc != nil {
				test.mockFunc(t, cmdRunner, logger)
			}

			cmd := command.RealCommand{CmdRunner: cmdRunner, Logger: internalLogger}

			dir := t.TempDir()
			spec := command.Spec{
				Key:     "some-key",
				Command: test.command,
				Dir:     dir,
				Env: []string{
					"FOO=BAR",
					"GO_WANT_HELPER_PROCESS=1",
					fmt.Sprintf("EXIT_STATUS=%d", test.mockExitCode),
					fmt.Sprintf("STDOUT=%s", test.mockStdout),
				},
				Operation: operations.Exec,
			}
			err := cmd.Run(context.Background(), logger, spec)
			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, logger)
		})
	}
}

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockLogger) LogEntry(key string, cmd []string) cmdlogger.LogEntry {
	args := m.Called(key, cmd)
	return args.Get(0).(cmdlogger.LogEntry)
}

type mockLogEntry struct {
	mock.Mock
}

func (m *mockLogEntry) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockLogEntry) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockLogEntry) Finalize(exitCode int) {
	m.Called(exitCode)
}

func (m *mockLogEntry) CurrentLogEntry() executor.ExecutionLogEntry {
	args := m.Called()
	return args.Get(0).(executor.ExecutionLogEntry)
}

type fakeCmdRunner struct {
	mock.Mock
}

var _ util.CmdRunner = &fakeCmdRunner{}

func (f *fakeCmdRunner) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestExecCommandHelper", "--"}
	cs = append(cs, args...)
	return exec.Command(os.Args[0], cs...)
}

func (f *fakeCmdRunner) CombinedOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	panic("not needed")
}

func (f *fakeCmdRunner) LookPath(file string) (string, error) {
	panic("not needed")
}

func (f *fakeCmdRunner) Stat(filename string) (os.FileInfo, error) {
	panic("not needed")
}

// TestExecCommandHelper a fake test that fakeExecCommand will run instead of calling the actual exec.CommandContext.
func TestExecCommandHelper(t *testing.T) {
	// Since this function must be big T test. We don't want to actually test anything. So if GO_WANT_HELPER_PROCESS
	// is not set, just exit right away.
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	_, err := fmt.Fprint(os.Stdout, os.Getenv("STDOUT"))
	require.NoError(t, err)

	i, err := strconv.Atoi(os.Getenv("EXIT_STATUS"))
	require.NoError(t, err)

	os.Exit(i)
}
