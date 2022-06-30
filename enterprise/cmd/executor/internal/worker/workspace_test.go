package worker

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
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

	dir, err := handler.prepareWorkspace(context.Background(), runner, "torvalds/linux", "", "deadbeef", true, false, []string{})
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir)

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "remote", "add", "origin", "https://executor@test.io/internal/git/torvalds/linux"},
		{"git", "-C", dir, "config", "--local", "gc.auto", "0"},
		{"git", "-C", dir, "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "fetch", "--progress", "--no-recurse-submodules", "--tags", "origin", "deadbeef"},
		{"git", "-C", dir, "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", dir, "remote", "set-url", "origin", "torvalds/linux"},
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

	dir, err := handler.prepareWorkspace(context.Background(), runner, "torvalds/linux", "subdirectory", "deadbeef", false, false, []string{})
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir)

	repoDir := filepath.Join(dir, "subdirectory")

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

	dir, err := handler.prepareWorkspace(context.Background(), runner, "torvalds/linux", "", "deadbeef", false, true, []string{})
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir)

	if value := len(runner.RunFunc.History()); value != 6 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 6, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "remote", "add", "origin", "https://executor@test.io/internal/git/torvalds/linux"},
		{"git", "-C", dir, "config", "--local", "gc.auto", "0"},
		{"git", "-C", dir, "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "fetch", "--progress", "--no-recurse-submodules", "--no-tags", "--depth=1", "origin", "deadbeef"},
		{"git", "-C", dir, "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", dir, "remote", "set-url", "origin", "torvalds/linux"},
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

	dir, err := handler.prepareWorkspace(context.Background(), runner, "torvalds/linux", "", "deadbeef", false, true, []string{"kernel"})
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir)

	if value := len(runner.RunFunc.History()); value != 8 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 8, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Command)
	}

	expectedCommands := [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "remote", "add", "origin", "https://executor@test.io/internal/git/torvalds/linux"},
		{"git", "-C", dir, "config", "--local", "gc.auto", "0"},
		{"git", "-C", dir, "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "fetch", "--progress", "--no-recurse-submodules", "--no-tags", "--depth=1", "--filter=blob:none", "origin", "deadbeef"},
		{"git", "-C", dir, "config", "--local", "core.sparseCheckout", "1"},
		{"git", "-C", dir, "sparse-checkout", "set", "--no-cone", "--", "kernel"},
		{"git", "-C", dir, "-c", "protocol.version=2", "-c", "http.extraHeader=Authorization: token-executor hunter2", "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal", "checkout", "--progress", "--force", "deadbeef"},
		{"git", "-C", dir, "remote", "set-url", "origin", "torvalds/linux"},
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

	dir, err := handler.prepareWorkspace(context.Background(), runner, "", "", "", false, false, []string{})
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir)

	if value := len(runner.RunFunc.History()); value != 0 {
		t.Fatalf("unexpected call to Run")
	}
}
