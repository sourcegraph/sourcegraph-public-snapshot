package worker

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestHandle(t *testing.T) {
	testDir := t.TempDir()
	workspace.MakeTempDirectory = func(string) (string, error) { return testDir, nil }
	t.Cleanup(func() {
		workspace.MakeTempDirectory = workspace.MakeTemporaryDirectory
	})

	if err := os.MkdirAll(filepath.Join(testDir, command.ScriptsPath), os.ModePerm); err != nil {
		t.Fatalf("unexpected error creating workspace: %s", err)
	}

	runner := NewMockRunner()

	job := executor.Job{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		VirtualMachineFiles: map[string]executor.VirtualMachineFile{
			"test.txt": {Content: "<file payload>"},
		},
		DockerSteps: []executor.DockerStep{
			{
				Image:    "go",
				Commands: []string{"go", "mod", "install"},
				Dir:      "",
				Env:      []string{"FOO=BAR"},
			},
			{
				Image:    "alpine",
				Commands: []string{"yarn", "install"},
				Dir:      "web",
				Env:      []string{},
			},
		},
		CliSteps: []executor.CliStep{
			{
				Commands: []string{"batch", "help"},
				Dir:      "",
				Env:      []string{},
			},
			{
				Commands: []string{"batch", "apply", "-f", "spec.yaml"},
				Dir:      "cmpg",
				Env:      []string{"BAR=BAZ"},
			},
		},
	}

	filesStore := NewMockFilesStore()

	h := &handler{
		store:      NewMockStore(),
		filesStore: filesStore,
		nameSet:    janitor.NewNameSet(),
		options:    Options{},
		operations: command.NewOperations(&observation.TestContext),
		runnerFactory: func(dir string, logger command.Logger, options command.Options, operations *command.Operations) command.Runner {
			if dir == "" {
				// The handler allocates a temporary runner to invoke the git commands,
				// which do not have a specific directory to run in. We don't need to
				// check those (again) as they were already confirmed in the workspace
				// specific unit tests. We'll just give it a blackhole runner so we don't
				// have to deal with more output during assertions.
				return NewMockRunner()
			}

			return runner
		},
	}

	if err := h.Handle(context.Background(), logtest.Scoped(t), job); err != nil {
		t.Fatalf("unexpected error handling record: %s", err)
	}

	if value := len(runner.SetupFunc.History()); value != 1 {
		t.Errorf("unexpected number of Setup calls. want=%d have=%d", 1, value)
	}
	if value := len(runner.TeardownFunc.History()); value != 1 {
		t.Errorf("unexpected number of Teardown calls. want=%d have=%d", 1, value)
	}
	if value := len(runner.RunFunc.History()); value != 4 {
		t.Fatalf("unexpected number of Run calls. want=%d have=%d", 4, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		if call.Arg1.Image != "" {
			commands = append(commands, []string{"/bin/sh", call.Arg1.ScriptPath})
		} else {
			commands = append(commands, call.Arg1.Command)
		}
	}

	// Ensure the files store was not called
	assert.Len(t, filesStore.GetFunc.History(), 0)

	expectedCommands := [][]string{
		{"/bin/sh", "42.0_linux@deadbeef.sh"},
		{"/bin/sh", "42.1_linux@deadbeef.sh"},
		{"src", "batch", "help"},
		{"src", "batch", "apply", "-f", "spec.yaml"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestHandle_WorkspaceFile(t *testing.T) {
	testDir := t.TempDir()
	workspace.MakeTempDirectory = func(string) (string, error) { return testDir, nil }
	t.Cleanup(func() {
		workspace.MakeTempDirectory = workspace.MakeTemporaryDirectory
	})

	if err := os.MkdirAll(filepath.Join(testDir, command.ScriptsPath), os.ModePerm); err != nil {
		t.Fatalf("unexpected error creating workspace: %s", err)
	}

	runner := NewMockRunner()

	virtualFileModifiedAt := time.Now()

	job := executor.Job{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		VirtualMachineFiles: map[string]executor.VirtualMachineFile{
			"test.txt":  {Content: "<file payload>"},
			"script.sh": {Bucket: "batch-changes", Key: "123/abc", ModifiedAt: virtualFileModifiedAt},
		},
		DockerSteps: []executor.DockerStep{
			{
				Image:    "go",
				Commands: []string{"go", "mod", "install"},
				Dir:      "",
				Env:      []string{"FOO=BAR"},
			},
			{
				Image:    "alpine",
				Commands: []string{"yarn", "install"},
				Dir:      "web",
				Env:      []string{},
			},
		},
		CliSteps: []executor.CliStep{
			{
				Commands: []string{"batch", "help"},
				Dir:      "",
				Env:      []string{},
			},
			{
				Commands: []string{"batch", "apply", "-f", "spec.yaml"},
				Dir:      "cmpg",
				Env:      []string{"BAR=BAZ"},
			},
		},
	}

	filesStore := NewMockFilesStore()
	// mock the store
	filesStore.ExistsFunc.SetDefaultReturn(true, nil)
	filesStore.GetFunc.SetDefaultReturn(io.NopCloser(bytes.NewReader([]byte("echo foo"))), nil)

	h := &handler{
		store:      NewMockStore(),
		filesStore: filesStore,
		nameSet:    janitor.NewNameSet(),
		options: Options{
			// Do not clean directory after handler has run. We want to check if files were actually written.
			// t.TempDir() will clean up after the test.
			KeepWorkspaces: true,
		},
		operations: command.NewOperations(&observation.TestContext),
		runnerFactory: func(dir string, logger command.Logger, options command.Options, operations *command.Operations) command.Runner {
			if dir == "" {
				// The handler allocates a temporary runner to invoke the git commands,
				// which do not have a specific directory to run in. We don't need to
				// check those (again) as they were already confirmed in the workspace
				// specific unit tests. We'll just give it a blackhole runner so we don't
				// have to deal with more output during assertions.
				return NewMockRunner()
			}

			return runner
		},
	}

	if err := h.Handle(context.Background(), logtest.Scoped(t), job); err != nil {
		t.Fatalf("unexpected error handling record: %s", err)
	}

	if value := len(runner.SetupFunc.History()); value != 1 {
		t.Errorf("unexpected number of Setup calls. want=%d have=%d", 1, value)
	}
	if value := len(runner.TeardownFunc.History()); value != 1 {
		t.Errorf("unexpected number of Teardown calls. want=%d have=%d", 1, value)
	}
	if value := len(runner.RunFunc.History()); value != 4 {
		t.Fatalf("unexpected number of Run calls. want=%d have=%d", 4, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		if call.Arg1.Image != "" {
			commands = append(commands, []string{"/bin/sh", call.Arg1.ScriptPath})
		} else {
			commands = append(commands, call.Arg1.Command)
		}
	}

	// Ensure the files store was called properly
	getHistory := filesStore.GetFunc.History()
	assert.Len(t, getHistory, 1)
	assert.Equal(t, "batch-changes", getHistory[0].Arg1)
	assert.Equal(t, "123/abc", getHistory[0].Arg2)

	expectedCommands := [][]string{
		{"/bin/sh", "42.0_linux@deadbeef.sh"},
		{"/bin/sh", "42.1_linux@deadbeef.sh"},
		{"src", "batch", "help"},
		{"src", "batch", "apply", "-f", "spec.yaml"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}

	// Ensure files were actually written
	scriptFile, err := os.Open(filepath.Join(testDir, "script.sh"))
	require.NoError(t, err)
	defer scriptFile.Close()
	scriptFileContent, err := io.ReadAll(scriptFile)
	require.NoError(t, err)
	assert.Equal(t, "echo foo", string(scriptFileContent))

	testFile, err := os.Open(filepath.Join(testDir, "test.txt"))
	require.NoError(t, err)
	defer testFile.Close()
	testFileContent, err := io.ReadAll(testFile)
	require.NoError(t, err)
	assert.Equal(t, "<file payload>", string(testFileContent))

	dockerScriptFile1, err := os.Open(filepath.Join(testDir, ".sourcegraph-executor", "42.0_linux@deadbeef.sh"))
	require.NoError(t, err)
	defer dockerScriptFile1.Close()
	dockerScriptFile1Content, err := io.ReadAll(dockerScriptFile1)
	require.NoError(t, err)
	assert.Equal(t, "\nset -x\n\n\ngo\nmod\ninstall\n", string(dockerScriptFile1Content))

	dockerScriptFile2, err := os.Open(filepath.Join(testDir, ".sourcegraph-executor", "42.1_linux@deadbeef.sh"))
	require.NoError(t, err)
	defer dockerScriptFile2.Close()
	dockerScriptFile2Content, err := io.ReadAll(dockerScriptFile2)
	require.NoError(t, err)
	assert.Equal(t, "\nset -x\n\n\nyarn\ninstall\n", string(dockerScriptFile2Content))
}
