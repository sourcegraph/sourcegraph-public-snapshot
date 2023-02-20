package worker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient/queue"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/workspace"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var ignorePort = cmpopts.IgnoreSliceElements(func(v string) bool {
	return strings.HasPrefix(v, "http://127.0.0.1:")
})

func TestPrepareWorkspace_Clone(t *testing.T) {
	testDir := t.TempDir()
	workspace.MakeTempDirectory = func(string) (string, error) { return testDir, nil }
	t.Cleanup(func() {
		workspace.MakeTempDirectory = workspace.MakeTemporaryDirectory
	})

	options := Options{
		QueueOptions: queue.Options{
			BaseClientOptions: apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL:   "https://test.io",
					Token: "hunter2",
				},
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	workspace, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName: "torvalds/linux",
		Commit:         "deadbeef",
		FetchTags:      true,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(workspace.Path())

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", workspace.Path(), "init"},
		{"git", "-C", workspace.Path(), "remote", "add", "origin", "http://127.0.0.1:port/torvalds/linux"},
		{"git", "-C", workspace.Path(), "config", "--local", "gc.auto", "0"},
		{"git", "-C", workspace.Path(), "-c", "protocol.version=2", "fetch", "--progress", "--no-recurse-submodules", "--tags", "origin", "deadbeef"},
		{"git", "-C", workspace.Path(), "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", workspace.Path(), "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands, ignorePort); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_Clone_Subdirectory(t *testing.T) {
	testDir := t.TempDir()
	workspace.MakeTempDirectory = func(string) (string, error) { return testDir, nil }
	t.Cleanup(func() {
		workspace.MakeTempDirectory = workspace.MakeTemporaryDirectory
	})

	options := Options{
		QueueOptions: queue.Options{
			BaseClientOptions: apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL:   "https://test.io",
					Token: "hunter2",
				},
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	workspace, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName:      "torvalds/linux",
		RepositoryDirectory: "subdirectory",
		Commit:              "deadbeef",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(workspace.Path())

	repoDir := filepath.Join(workspace.Path(), "subdirectory")

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", repoDir, "init"},
		{"git", "-C", repoDir, "remote", "add", "origin", "http://127.0.0.1:port/torvalds/linux"},
		{"git", "-C", repoDir, "config", "--local", "gc.auto", "0"},
		{"git", "-C", repoDir, "-c", "protocol.version=2", "fetch", "--progress", "--no-recurse-submodules", "origin", "deadbeef"},
		{"git", "-C", repoDir, "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", repoDir, "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands, ignorePort); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_ShallowClone(t *testing.T) {
	testDir := t.TempDir()
	workspace.MakeTempDirectory = func(string) (string, error) { return testDir, nil }
	t.Cleanup(func() {
		workspace.MakeTempDirectory = workspace.MakeTemporaryDirectory
	})

	options := Options{
		QueueOptions: queue.Options{
			BaseClientOptions: apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL:   "https://test.io",
					Token: "hunter2",
				},
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	workspace, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName: "torvalds/linux",
		Commit:         "deadbeef",
		ShallowClone:   true,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(workspace.Path())

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", workspace.Path(), "init"},
		{"git", "-C", workspace.Path(), "remote", "add", "origin", "http://127.0.0.1:port/torvalds/linux"},
		{"git", "-C", workspace.Path(), "config", "--local", "gc.auto", "0"},
		{"git", "-C", workspace.Path(), "-c", "protocol.version=2", "fetch", "--progress", "--no-recurse-submodules", "--no-tags", "--depth=1", "origin", "deadbeef"},
		{"git", "-C", workspace.Path(), "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", workspace.Path(), "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands, ignorePort); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_SparseCheckout(t *testing.T) {
	testDir := t.TempDir()
	workspace.MakeTempDirectory = func(string) (string, error) { return testDir, nil }
	t.Cleanup(func() {
		workspace.MakeTempDirectory = workspace.MakeTemporaryDirectory
	})

	options := Options{
		QueueOptions: queue.Options{
			BaseClientOptions: apiclient.BaseClientOptions{
				EndpointOptions: apiclient.EndpointOptions{
					URL:   "https://test.io",
					Token: "hunter2",
				},
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	workspace, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName: "torvalds/linux",
		Commit:         "deadbeef",
		ShallowClone:   true,
		SparseCheckout: []string{"kernel"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(workspace.Path())

	if value := len(runner.RunFunc.History()); value != 8 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 8, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", workspace.Path(), "init"},
		{"git", "-C", workspace.Path(), "remote", "add", "origin", "http://127.0.0.1:port/torvalds/linux"},
		{"git", "-C", workspace.Path(), "config", "--local", "gc.auto", "0"},
		{"git", "-C", workspace.Path(), "-c", "protocol.version=2", "fetch", "--progress", "--no-recurse-submodules", "--no-tags", "--depth=1", "--filter=blob:none", "origin", "deadbeef"},
		{"git", "-C", workspace.Path(), "config", "--local", "core.sparseCheckout", "1"},
		{"git", "-C", workspace.Path(), "sparse-checkout", "set", "--no-cone", "--", "kernel"},
		{"git", "-C", workspace.Path(), "-c", "protocol.version=2", "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", workspace.Path(), "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands, ignorePort); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_NoRepository(t *testing.T) {
	testDir := t.TempDir()
	workspace.MakeTempDirectory = func(string) (string, error) { return testDir, nil }
	t.Cleanup(func() {
		workspace.MakeTempDirectory = workspace.MakeTemporaryDirectory
	})

	options := Options{}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	workspace, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(workspace.Path())

	if value := len(runner.RunFunc.History()); value != 0 {
		t.Fatalf("unexpected call to Run")
	}
}
