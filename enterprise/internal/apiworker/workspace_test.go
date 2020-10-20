package apiworker

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
)

func TestPrepareWorkspace(t *testing.T) {
	options := Options{
		ClientOptions: apiclient.Options{
			EndpointOptions: apiclient.EndpointOptions{
				URL:      "https://test.io",
				Username: "test",
				Password: "hunter2",
			},
		},
		GitServicePath: "/internal/git",
	}
	runner := NewMockRunner()
	handler := &handler{options: options}

	dir, err := handler.prepareWorkspace(context.Background(), runner, "linux", "deadbeef")
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir)

	if value := len(runner.RunFunc.History()); value != 3 {
		t.Fatalf("unexpected number of calls to Run. want=%d have=%d", 3, value)
	}

	var commands [][]string
	for _, call := range runner.RunFunc.History() {
		commands = append(commands, call.Arg1.Commands)
	}

	expectedCommands := [][]string{
		{"git", "-C", dir, "init"},
		{"git", "-C", dir, "-c", "protocol.version=2", "fetch", "https://test:hunter2@test.io/internal/git/linux", "deadbeef"},
		{"git", "-C", dir, "checkout", "deadbeef"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}

func TestPrepareWorkspaceNoRepository(t *testing.T) {
	options := Options{}
	runner := NewMockRunner()
	handler := &handler{options: options}

	dir, err := handler.prepareWorkspace(context.Background(), runner, "", "")
	if err != nil {
		t.Fatalf("unexpected error preparing workspace: %s", err)
	}
	defer os.RemoveAll(dir)

	if value := len(runner.RunFunc.History()); value != 0 {
		t.Fatalf("unexpected call to Run")
	}
}
