package command

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormatRawOrDockerCommandRaw(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Command: []string{"ls", "-a"},
			Dir:     "subdir",
			Env: []string{
				`TEST=true`,
				`CONTAINS_WHITESPACE=yes it does`,
			},
			Operation: makeTestOperation(),
		},
		"/proj/src",
		Options{},
	)

	expected := command{
		Command: []string{"ls", "-a"},
		Dir:     "/proj/src/subdir",
		Env:     []string{"TEST=true", "CONTAINS_WHITESPACE=yes it does"},
	}
	if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}

func TestFormatRawOrDockerCommandDockerScript(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Image:      "alpine:latest",
			ScriptPath: "myscript.sh",
			Dir:        "subdir",
			Env: []string{
				`TEST=true`,
				`CONTAINS_WHITESPACE=yes it does`,
			},
			Operation: makeTestOperation(),
		},
		"/proj/src",
		Options{
			ResourceOptions: ResourceOptions{
				NumCPUs: 4,
				Memory:  "20G",
			},
		},
	)

	expected := command{
		Command: []string{
			"docker", "run", "--rm",
			"--cpus", "4",
			"--memory", "20G",
			"-v", "/proj/src:/data",
			"-w", "/data/subdir",
			"-e", "TEST=true",
			"-e", `CONTAINS_WHITESPACE="yes it does"`,
			"--entrypoint",
			"/bin/sh",
			"alpine:latest",
			"/data/.sourcegraph-executor/myscript.sh",
		},
	}
	if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}
