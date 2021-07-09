package command

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func overwriteEnv(t *testing.T, k, v string) {
	old := os.Getenv(k)
	os.Setenv(k, v)
	t.Cleanup(func() { os.Setenv(k, old) })
}

func TestFormatRawOrDockerCommandRaw(t *testing.T) {
	overwriteEnv(t, "TEST2", "testing")

	actual := formatRawOrDockerCommand(
		CommandSpec{
			Command:         []string{"ls", "-a"},
			Dir:             "subdir",
			Env:             []string{"TEST=true"},
			InheritLocalEnv: []string{"TEST2"},
			Operation:       makeTestOperation(),
		},
		"/proj/src",
		Options{},
	)

	expected := command{
		Command: []string{"ls", "-a"},
		Dir:     "/proj/src/subdir",
		Env:     []string{"TEST=true", "TEST2=testing"},
	}
	if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}

func TestFormatRawOrDockerCommandDockerScript(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Image:           "alpine:latest",
			ScriptPath:      "myscript.sh",
			Dir:             "subdir",
			Env:             []string{"TEST=true"},
			InheritLocalEnv: []string{"TEST2"},
			Operation:       makeTestOperation(),
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
			"-e", "TEST2",
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

func TestFormatRawOrDockerCommandDockerCommand(t *testing.T) {
	overwriteEnv(t, "TEST2", "testing")

	actual := formatRawOrDockerCommand(
		CommandSpec{
			Command: []string{"ls", "-a"},
			Dir:     "subdir",
			Env:     []string{"TEST=true"},
			InheritLocalEnv: []string{"TEST2"},
		},
		"/proj/src",
		Options{},
	)

	expected := command{
		Command: []string{
			"ls", "-a",
		},
		Env: []string{"TEST=true", "TEST2=testing"},
		Dir: "/proj/src/subdir",
	}
	if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}
