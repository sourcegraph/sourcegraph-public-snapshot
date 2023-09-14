package run_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/batcheshelper/log"
	"github.com/sourcegraph/sourcegraph/cmd/batcheshelper/run"
	"github.com/sourcegraph/sourcegraph/cmd/batcheshelper/util"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
)

func TestPost(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, dir string, workspaceFileDir string, executionInput batcheslib.WorkspacesExecutionInput)
		mockFunc       func(runner *fakeCmdRunner)
		step           int
		executionInput batcheslib.WorkspacesExecutionInput
		previousResult execution.AfterStepResult
		stdoutLogs     string
		stderrLogs     string
		expectedErr    error
		assertFunc     func(t *testing.T, logEntries []batcheslib.LogEvent, dir string, runner *fakeCmdRunner)
	}{
		{
			name: "Success",
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("Git", mock.Anything, "", []string{"config", "--global", "--add", "safe.directory", "/job/repository"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"add", "--all"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"diff", "--cached", "--no-prefix", "--binary"}).
					Return("git diff", nil)
			},
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{Run: "echo hello world"},
				},
			},
			previousResult: execution.AfterStepResult{},
			stdoutLogs:     "hello world",
			stderrLogs:     "error",
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string, runner *fakeCmdRunner) {
				require.Len(t, logEntries, 2)

				assert.Equal(t, batcheslib.LogEventOperationTaskStep, logEntries[0].Operation)
				assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[0].Timestamp)
				assert.Equal(t, batcheslib.LogEventStatusSuccess, logEntries[0].Status)
				assert.IsType(t, &batcheslib.TaskStepMetadata{}, logEntries[0].Metadata)
				assert.Equal(t, []byte("git diff"), logEntries[0].Metadata.(*batcheslib.TaskStepMetadata).Diff)

				assert.Equal(t, batcheslib.LogEventOperationCacheAfterStepResult, logEntries[1].Operation)
				assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[1].Timestamp)
				assert.Equal(t, batcheslib.LogEventStatusSuccess, logEntries[1].Status)
				assert.IsType(t, &batcheslib.CacheAfterStepResultMetadata{}, logEntries[1].Metadata)
				assert.Equal(t, "deZzMP85HWs6lfhWRnMVBA-step-0", logEntries[1].Metadata.(*batcheslib.CacheAfterStepResultMetadata).Key)
				assert.Equal(t, "hello world", logEntries[1].Metadata.(*batcheslib.CacheAfterStepResultMetadata).Value.Stdout)
				assert.Equal(t, "error", logEntries[1].Metadata.(*batcheslib.CacheAfterStepResultMetadata).Value.Stderr)
				assert.Equal(t, []byte("git diff"), logEntries[1].Metadata.(*batcheslib.CacheAfterStepResultMetadata).Value.Diff)

				entries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 3)
				b, err := os.ReadFile(filepath.Join(dir, "step0.json"))
				require.NoError(t, err)
				var result execution.AfterStepResult
				err = json.Unmarshal(b, &result)
				require.NoError(t, err)
				assert.Equal(
					t,
					execution.AfterStepResult{
						Version: 2,
						Stdout:  "hello world",
						Stderr:  "error",
						Diff:    []byte("git diff"),
						Outputs: make(map[string]interface{}),
					},
					result,
				)
			},
		},
		{
			name: "File Mounts",
			setupFunc: func(t *testing.T, dir string, workspaceFileDir string, executionInput batcheslib.WorkspacesExecutionInput) {
				err := os.Mkdir(filepath.Join(dir, "step0files"), os.ModePerm)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(dir, "step0files", "file1.txt"), []byte("hello world"), os.ModePerm)
				require.NoError(t, err)
			},
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("Git", mock.Anything, "", []string{"config", "--global", "--add", "safe.directory", "/job/repository"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"add", "--all"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"diff", "--cached", "--no-prefix", "--binary"}).
					Return("git diff", nil)
			},
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{
						Run: "echo hello world",
						Files: map[string]string{
							"file1.txt": "hello world",
						},
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			stdoutLogs:     "hello world",
			stderrLogs:     "error",
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string, runner *fakeCmdRunner) {
				require.Len(t, logEntries, 2)
				assert.Equal(t, "4qXjs4-Arh1VpWWfWhqm3A-step-0", logEntries[1].Metadata.(*batcheslib.CacheAfterStepResultMetadata).Key)

				entries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 3)
				b, err := os.ReadFile(filepath.Join(dir, "step0.json"))
				require.NoError(t, err)
				var result execution.AfterStepResult
				err = json.Unmarshal(b, &result)
				require.NoError(t, err)
				assert.Equal(
					t,
					execution.AfterStepResult{
						Version: 2,
						Stdout:  "hello world",
						Stderr:  "error",
						Diff:    []byte("git diff"),
						Outputs: make(map[string]interface{}),
					},
					result,
				)

				_, err = os.Stat(filepath.Join(dir, "step0files"))
				require.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "Workspace Files",
			setupFunc: func(t *testing.T, dir string, workspaceFileDir string, executionInput batcheslib.WorkspacesExecutionInput) {
				path := filepath.Join(workspaceFileDir, "file1.txt")
				err := os.WriteFile(path, []byte("hello world"), os.ModePerm)
				require.NoError(t, err)
				executionInput.Steps[0].Mount = []batcheslib.Mount{
					{
						Mountpoint: "/foo/file1.txt",
						Path:       path,
					},
				}
			},
			mockFunc: func(runner *fakeCmdRunner) {
				runner.On("Git", mock.Anything, "", []string{"config", "--global", "--add", "safe.directory", "/job/repository"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"add", "--all"}).
					Return("", nil)
				runner.On("Git", mock.Anything, "repository", []string{"diff", "--cached", "--no-prefix", "--binary"}).
					Return("git diff", nil)
			},
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{Run: "echo hello world"},
				},
			},
			previousResult: execution.AfterStepResult{},
			stdoutLogs:     "hello world",
			stderrLogs:     "error",
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string, runner *fakeCmdRunner) {
				require.Len(t, logEntries, 2)
				assert.Regexp(t, ".*-step-0$", logEntries[1].Metadata.(*batcheslib.CacheAfterStepResultMetadata).Key)

				entries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, entries, 3)
				b, err := os.ReadFile(filepath.Join(dir, "step0.json"))
				require.NoError(t, err)
				var result execution.AfterStepResult
				err = json.Unmarshal(b, &result)
				require.NoError(t, err)
				assert.Equal(
					t,
					execution.AfterStepResult{
						Version: 2,
						Stdout:  "hello world",
						Stderr:  "error",
						Diff:    []byte("git diff"),
						Outputs: make(map[string]interface{}),
					},
					result,
				)

				_, err = os.Stat(filepath.Join(dir, "workspaceFiles"))
				require.Error(t, err)
				assert.True(t, os.IsNotExist(err))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := &log.Logger{Writer: &buf}

			dir := t.TempDir()
			workspaceFilesDir := filepath.Join(dir, "workspaceFiles")
			err := os.Mkdir(workspaceFilesDir, os.ModePerm)
			require.NoError(t, err)

			err = os.WriteFile(filepath.Join(dir, fmt.Sprintf("stdout%d.log", test.step)), []byte(test.stdoutLogs), os.ModePerm)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(dir, fmt.Sprintf("stderr%d.log", test.step)), []byte(test.stderrLogs), os.ModePerm)
			require.NoError(t, err)

			if test.setupFunc != nil {
				test.setupFunc(t, dir, workspaceFilesDir, test.executionInput)
			}

			runner := new(fakeCmdRunner)
			if test.mockFunc != nil {
				test.mockFunc(runner)
			}

			err = run.Post(context.Background(), logger, runner, test.step, test.executionInput, test.previousResult, dir, workspaceFilesDir, true)

			if test.expectedErr != nil {
				require.Error(t, err)
				assert.EqualError(t, err, test.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			logs := buf.String()
			logLines := strings.Split(logs, "\n")
			logEntries := make([]batcheslib.LogEvent, len(logLines)-1)
			for i, line := range logLines {
				if len(line) == 0 {
					break
				}
				var entry batcheslib.LogEvent
				err = json.Unmarshal([]byte(line), &entry)
				require.NoError(t, err)
				logEntries[i] = entry
			}

			if test.assertFunc != nil {
				test.assertFunc(t, logEntries, dir, runner)
			}
		})
	}
}

type fakeCmdRunner struct {
	mock.Mock
}

var _ util.CmdRunner = &fakeCmdRunner{}

func (f *fakeCmdRunner) Git(ctx context.Context, dir string, args ...string) ([]byte, error) {
	calledArgs := f.Called(ctx, dir, args)
	return []byte(calledArgs.String(0)), calledArgs.Error(1)
}
