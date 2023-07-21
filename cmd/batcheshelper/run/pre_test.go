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
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/batcheshelper/log"
	"github.com/sourcegraph/sourcegraph/cmd/batcheshelper/run"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/env"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
)

func TestPre(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(t *testing.T, dir string, executionInput batcheslib.WorkspacesExecutionInput)
		step           int
		executionInput batcheslib.WorkspacesExecutionInput
		previousResult execution.AfterStepResult
		expectedErr    error
		assertFunc     func(t *testing.T, logEntries []batcheslib.LogEvent, dir string)
	}{
		{
			name: "Step skipped",
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{If: false},
				},
			},
			previousResult: execution.AfterStepResult{},
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)
				assert.Equal(t, batcheslib.LogEventOperationTaskStepSkipped, logEntries[0].Operation)
				assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[0].Timestamp)
				assert.Equal(t, batcheslib.LogEventStatusProgress, logEntries[0].Status)
				assert.IsType(t, &batcheslib.TaskStepSkippedMetadata{}, logEntries[0].Metadata)
				assert.Equal(t, 1, logEntries[0].Metadata.(*batcheslib.TaskStepSkippedMetadata).Step)

				dirEntries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 2)
				for _, entry := range dirEntries {
					if entry.Name() == "step0.json" {
						assert.Equal(t, "step0.json", entry.Name())
						b, err := os.ReadFile(filepath.Join(dir, entry.Name()))
						require.NoError(t, err)
						var result execution.AfterStepResult
						err = json.Unmarshal(b, &result)
						require.NoError(t, err)
						assert.Equal(t, execution.AfterStepResult{Version: 2, Skipped: true}, result)
					} else if entry.Name() == "skip.json" {
						assert.Equal(t, "skip.json", entry.Name())
						b, err := os.ReadFile(filepath.Join(dir, entry.Name()))
						require.NoError(t, err)
						var data map[string]interface{}
						err = json.Unmarshal(b, &data)
						require.NoError(t, err)
						assert.Equal(t, "step.1.pre", data["nextStep"])
					}
				}
			},
		},
		{
			name: "Simple step",
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{Run: "echo hello"},
				},
			},
			previousResult: execution.AfterStepResult{},
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)
				assert.Equal(t, batcheslib.LogEventOperationTaskStep, logEntries[0].Operation)
				assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d+ \+\d{4} UTC$`, logEntries[0].Timestamp)
				assert.Equal(t, batcheslib.LogEventStatusStarted, logEntries[0].Status)
				assert.IsType(t, &batcheslib.TaskStepMetadata{}, logEntries[0].Metadata)
				assert.Equal(t, 1, logEntries[0].Metadata.(*batcheslib.TaskStepMetadata).Step)

				dirEntries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 1)
				assert.Equal(t, "step0.sh", dirEntries[0].Name())
				b, err := os.ReadFile(filepath.Join(dir, dirEntries[0].Name()))
				require.NoError(t, err)
				assert.Equal(t, "echo hello", string(b))

				f, err := os.Open(filepath.Join(dir, dirEntries[0].Name()))
				require.NoError(t, err)
				defer f.Close()
				stat, err := f.Stat()
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0755), stat.Mode().Perm())
			},
		},
		{
			name: "File mounts",
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{
						Run: "echo hello",
						Files: map[string]string{
							"file1.sh": "echo file1",
							"file2.sh": "echo file2",
						},
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)

				dirEntries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 2)

				stepFiles, err := os.ReadDir(filepath.Join(dir, "step0files"))
				require.NoError(t, err)
				require.Len(t, stepFiles, 2)
			},
		},
		{
			name: "Workspace files",
			setupFunc: func(t *testing.T, dir string, executionInput batcheslib.WorkspacesExecutionInput) {
				err := os.Mkdir(filepath.Join(dir, "workspaceFiles"), os.ModePerm)
				require.NoError(t, err)

				file1 := filepath.Join(dir, "workspaceFiles", "file1.sh")
				err = os.WriteFile(file1, []byte("echo file1"), os.ModePerm)
				require.NoError(t, err)

				file2 := filepath.Join(dir, "workspaceFiles", "file2.sh")
				err = os.WriteFile(file2, []byte("echo file2"), os.ModePerm)
				require.NoError(t, err)

				executionInput.Steps[0].Mount = []batcheslib.Mount{
					{
						Mountpoint: "/foo/file1.sh",
						Path:       file1,
					},
					{
						Mountpoint: "/bar/file2.sh",
						Path:       file2,
					},
				}
			},
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{
						Run: "echo hello",
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)

				dirEntries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 2)

				for _, entry := range dirEntries {
					if entry.Name() == "step0.sh" {
						b, err := os.ReadFile(filepath.Join(dir, entry.Name()))
						require.NoError(t, err)
						assert.Equal(
							t,
							fmt.Sprintf("cp -r %s /foo/file1.sh\n", filepath.Join(dir, "workspaceFiles", "file1.sh"))+
								"chmod -R +x /foo/file1.sh\n"+
								fmt.Sprintf("cp -r %s /bar/file2.sh\n", filepath.Join(dir, "workspaceFiles", "file2.sh"))+
								"chmod -R +x /bar/file2.sh\n"+
								"echo hello",
							string(b),
						)
						break
					}
				}
			},
		},
		{
			name: "Environment Variables",
			setupFunc: func(t *testing.T, dir string, executionInput batcheslib.WorkspacesExecutionInput) {
				var envVars env.Environment
				err := json.Unmarshal([]byte(`{"FOO": "BAR"}`), &envVars)
				require.NoError(t, err)

				executionInput.Steps[0].Env = envVars
			},
			step: 0,
			executionInput: batcheslib.WorkspacesExecutionInput{
				Steps: []batcheslib.Step{
					{
						Run: "echo hello",
					},
				},
			},
			previousResult: execution.AfterStepResult{},
			assertFunc: func(t *testing.T, logEntries []batcheslib.LogEvent, dir string) {
				require.Len(t, logEntries, 1)

				dirEntries, err := os.ReadDir(dir)
				require.NoError(t, err)
				require.Len(t, dirEntries, 1)
				b, err := os.ReadFile(filepath.Join(dir, "step0.sh"))
				require.NoError(t, err)
				assert.Equal(
					t,
					"export FOO=BAR\necho hello",
					string(b),
				)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := &log.Logger{Writer: &buf}

			dir := t.TempDir()

			if test.setupFunc != nil {
				test.setupFunc(t, dir, test.executionInput)
			}

			err := run.Pre(context.Background(), logger, test.step, test.executionInput, test.previousResult, dir, dir)

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
				test.assertFunc(t, logEntries, dir)
			}
		})
	}
}
