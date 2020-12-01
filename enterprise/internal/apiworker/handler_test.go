package apiworker

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/apiclient"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/apiworker/command"
)

func TestHandle(t *testing.T) {
	store := NewMockStore()
	runner := NewMockRunner()

	job := apiclient.Job{
		ID:             42,
		Commit:         "deadbeef",
		RepositoryName: "linux",
		VirtualMachineFiles: map[string]string{
			"test.txt": "<file payload>",
		},
		DockerSteps: []apiclient.DockerStep{
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
		CliSteps: []apiclient.CliStep{
			{
				Commands: []string{"campaigns", "help"},
				Dir:      "",
				Env:      []string{},
			},
			{
				Commands: []string{"campaigns", "apply", "-f", "spec.yaml"},
				Dir:      "cmpg",
				Env:      []string{"BAR=BAZ"},
			},
		},
	}

	handler := &handler{
		idSet:   newIDSet(),
		options: Options{},
		runnerFactory: func(dir string, logger *command.Logger, options command.Options) command.Runner {
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

	if err := handler.Handle(context.Background(), store, job); err != nil {
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
		commands = append(commands, call.Arg1.Commands)
	}

	expectedCommands := [][]string{
		{"go", "mod", "install"},
		{"yarn", "install"},
		{"src", "campaigns", "help"},
		{"src", "campaigns", "apply", "-f", "spec.yaml"},
	}
	if diff := cmp.Diff(expectedCommands, commands); diff != "" {
		t.Errorf("unexpected commands (-want +got):\n%s", diff)
	}
}
