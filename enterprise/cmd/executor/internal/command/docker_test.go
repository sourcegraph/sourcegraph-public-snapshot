package command

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestFormatRawOrDockerCommandRaw(t *testing.T) {
	t.Run("docker", func(t *testing.T) {
		actual := formatRawOrDockerCommand(
			CommandSpec{
				Image:   "sourcegraph/sourcegraph",
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
			"/tmp/docker-config",
		)

		expected := command{
			Command: []string{
				"docker",
				"--config", "/tmp/docker-config",
				"run",
				"--rm",
				"-v", "/proj/src:/data",
				"-w", "/data/subdir",
				"-e", "TEST=true",
				"-e", "CONTAINS_WHITESPACE=yes it does",
				"--entrypoint", "/bin/sh",
				"sourcegraph/sourcegraph",
				"/data/.sourcegraph-executor",
			},
		}
		if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
			t.Errorf("unexpected command (-want +got):\n%s", diff)
		}
	})

	t.Run("docker with host gateway", func(t *testing.T) {
		conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{
			ExecutorsFrontendURL: "http://host.docker.internal:3080",
		}})
		defer conf.Mock(nil)

		actual := formatRawOrDockerCommand(
			CommandSpec{
				Image:     "sourcegraph/sourcegraph",
				Env:       []string{},
				Operation: makeTestOperation(),
			},
			"/proj/src",
			Options{},
			"",
		)

		expected := command{
			Command: []string{
				"docker",
				"run",
				"--rm",
				"--add-host=host.docker.internal:host-gateway",
				"-v", "/proj/src:/data",
				"-w", "/data",
				"--entrypoint", "/bin/sh",
				"sourcegraph/sourcegraph",
				"/data/.sourcegraph-executor",
			},
		}
		if diff := cmp.Diff(expected, actual, commandComparer); diff != "" {
			t.Errorf("unexpected command (-want +got):\n%s", diff)
		}
	})
}

func TestFormatRawOrDockerCommandRaw_SrcCli(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Image:   "",
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
		"/tmp/docker-config",
	)

	expected := command{
		Command: []string{"ls", "-a"},
		Dir:     "/proj/src/subdir",
		Env:     []string{"TEST=true", "CONTAINS_WHITESPACE=yes it does", "DOCKER_CONFIG=/tmp/docker-config"},
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
		"/tmp/docker-config",
	)

	expected := command{
		Command: []string{
			"docker",
			"--config", "/tmp/docker-config",
			"run", "--rm",
			"--cpus", "4",
			"--memory", "20G",
			"-v", "/proj/src:/data",
			"-w", "/data/subdir",
			"-e", "TEST=true",
			// Note: This does NOT need to be quoted, as exec.Command
			// properly passes each string in this slice as a separate argument.
			"-e", `CONTAINS_WHITESPACE=yes it does`,
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

func TestFormatRawOrDockerCommandDockerScriptWithoutResourceAllocation(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Image:      "alpine:latest",
			ScriptPath: "myscript.sh",
			Dir:        "subdir",
			Operation:  makeTestOperation(),
		},
		"/proj/src",
		Options{
			ResourceOptions: ResourceOptions{
				NumCPUs: 0,
				Memory:  "0",
			},
		},
		"/tmp/docker-config",
	)

	expected := command{
		Command: []string{
			"docker",
			"--config", "/tmp/docker-config",
			"run", "--rm",
			"-v", "/proj/src:/data",
			"-w", "/data/subdir",
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

func TestFormatRawOrDockerCommandDockerScriptWithDockerHostMountPath(t *testing.T) {
	actual := formatRawOrDockerCommand(
		CommandSpec{
			Image:      "alpine:latest",
			ScriptPath: "myscript.sh",
			Dir:        "subdir",
			Operation:  makeTestOperation(),
		},
		"/my/local/workspace",
		Options{
			ResourceOptions: ResourceOptions{
				NumCPUs:             4,
				Memory:              "20G",
				DockerHostMountPath: "/containers/rootfs/mount_fs",
			},
		},
		"/tmp/docker-config",
	)

	expected := command{
		Command: []string{
			"docker",
			"--config", "/tmp/docker-config",
			"run", "--rm",
			"--cpus", "4",
			"--memory", "20G",
			"-v", "/containers/rootfs/mount_fs/workspace:/data",
			"-w", "/data/subdir",
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
