package command

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFormatRawOrDockerCommandRaw(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Commands: []string{"ls", "-a"},
			Dir:      "subdir",
			Env:      []string{"TEST=true"},
		},
		"/proj/src",
		Options{},
	)

	expected := command{
		Commands: []string{"ls", "-a"},
		Dir:      "/proj/src/subdir",
		Env:      []string{"TEST=true"},
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}

func TestFormatRawOrDockerCommandDocker(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Image:    "alpine:latest",
			Commands: []string{"ls", "-a"},
			Dir:      "subdir",
			Env:      []string{"TEST=true"},
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
		Commands: []string{
			"docker", "run", "--rm",
			"--cpus", "4",
			"--memory", "20G",
			"-v", "/proj/src:/data",
			"-w", "/data/subdir",
			"-e", "TEST=true",
			"alpine:latest",
			"ls", "-a",
		},
	}
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("unexpected command (-want +got):\n%s", diff)
	}
}
