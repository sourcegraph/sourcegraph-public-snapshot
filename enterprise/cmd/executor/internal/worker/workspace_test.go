package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestPrepareWorkspace_Clone(t *testing.T) {
	options := Options{
		ClientOptions: apiclient.Options{
			EndpointOptions: apiclient.EndpointOptions{
				URL:   "https://test.io",
				Token: "hunter2",
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	dir, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName: "torvalds/linux",
		Commit:         "deadbeef",
		FetchTags:      true,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir.Path())

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", dir.Path(), "init"},
		{"git", "-C", dir.Path(), "remote", "add", "origin", "https://executor@test.io/internal/git/torvalds/linux"},
		{"git", "-C", dir.Path(), "config", "--local", "gc.auto", "0"},
		{"git", "-C", dir.Path(), "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "fetch", "--progress", "--no-recurse-submodules", "--tags", "origin", "deadbeef"},
		{"git", "-C", dir.Path(), "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", dir.Path(), "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_Clone_Subdirectory(t *testing.T) {
	options := Options{
		ClientOptions: apiclient.Options{
			EndpointOptions: apiclient.EndpointOptions{
				URL:   "https://test.io",
				Token: "hunter2",
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	// dir, err := handler.prepareWorkspace(context.Background(), runner, "torvalds/linux", "subdirectory", "deadbeef", false, false, []string{})
	dir, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName:      "torvalds/linux",
		RepositoryDirectory: "subdirectory",
		Commit:              "deadbeef",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir.Path())

	repoDir := filepath.Join(dir.Path(), "subdirectory")

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", repoDir, "init"},
		{"git", "-C", repoDir, "remote", "add", "origin", "https://executor@test.io/internal/git/torvalds/linux"},
		{"git", "-C", repoDir, "config", "--local", "gc.auto", "0"},
		{"git", "-C", repoDir, "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "fetch", "--progress", "--no-recurse-submodules", "origin", "deadbeef"},
		{"git", "-C", repoDir, "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", repoDir, "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_ShallowClone(t *testing.T) {
	options := Options{
		ClientOptions: apiclient.Options{
			EndpointOptions: apiclient.EndpointOptions{
				URL:   "https://test.io",
				Token: "hunter2",
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	// dir, err := handler.prepareWorkspace(context.Background(), runner, "torvalds/linux", "", "deadbeef", false, true, []string{})
	dir, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName: "torvalds/linux",
		Commit:         "deadbeef",
		ShallowClone:   true,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir.Path())

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", dir.Path(), "init"},
		{"git", "-C", dir.Path(), "remote", "add", "origin", "https://executor@test.io/internal/git/torvalds/linux"},
		{"git", "-C", dir.Path(), "config", "--local", "gc.auto", "0"},
		{"git", "-C", dir.Path(), "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "fetch", "--progress", "--no-recurse-submodules", "--no-tags", "--depth=1", "origin", "deadbeef"},
		{"git", "-C", dir.Path(), "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", dir.Path(), "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_SparseCheckout(t *testing.T) {
	options := Options{
		ClientOptions: apiclient.Options{
			EndpointOptions: apiclient.EndpointOptions{
				URL:   "https://test.io",
				Token: "hunter2",
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	dir, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{
		RepositoryName: "torvalds/linux",
		Commit:         "deadbeef",
		ShallowClone:   true,
		SparseCheckout: []string{"kernel"},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir.Path())

	if value := len(runner.RunFunc.History()); value != 8 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 8, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", dir.Path(), "init"},
		{"git", "-C", dir.Path(), "remote", "add", "origin", "https://executor@test.io/internal/git/torvalds/linux"},
		{"git", "-C", dir.Path(), "config", "--local", "gc.auto", "0"},
		{"git", "-C", dir.Path(), "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "fetch", "--progress", "--no-recurse-submodules", "--no-tags", "--depth=1", "--filter=blob:none", "origin", "deadbeef"},
		{"git", "-C", dir.Path(), "config", "--local", "core.sparseCheckout", "1"},
		{"git", "-C", dir.Path(), "sparse-checkout", "set", "--no-cone", "--", "kernel"},
		{"git", "-C", dir.Path(), "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", dir.Path(), "remote", "set-url", "origin", "torvalds/linux"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspace_NoRepository(t *testing.T) {
	options := Options{}
	runner := NewMockRunner()
	handler := &handler{
		options:    options,
		operations: command.NewOperations(&observation.TestContext),
	}

	dir, err := handler.prepareWorkspace(context.Background(), runner, executor.Job{}, nil)
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir.Path())

	if value := len(runner.RunFunc.History()); value != 0 {
		t.Fatalf("unexpected call to Run")
	}
}
