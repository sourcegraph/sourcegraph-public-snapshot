package command

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormatRawOrDockerCommandRaw(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Command:   []string{"ls", "-a"},
			Dir:       "subdir",
			Env:       []string{"TEST=true"},
			Operation: makeTestOperation(),
		},
		"/proj/src",
		Options{},
	)

	expected := command{
		Command: []string{"ls", "-a"},
		Dir:     "/proj/src/subdir",
		Env:     []string{"TEST=true"},
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
			Env:        []string{"TEST=true"},
			Operation:  makeTestOperation(),
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
			"-v", "myscript.sh:myscript.sh",
			"-w", "/data/subdir",
			"-e", "TEST=true",
			"--entrypoint",
			"/bin/sh",
			"alpine:latest",
			"myscript.sh",
		},
	}
	if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}

func TestFormatRawOrDockerCommandDockerCommand(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Command: []string{"ls", "-a"},
			Dir:     "subdir",
			Env:     []string{"TEST=true"},
		},
		"/proj/src",
		Options{},
	)

	expected := command{
		Command: []string{
			"ls", "-a",
		},
		Env: []string{"TEST=true"},
		Dir: "/proj/src/subdir",
	}
	if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}
