package run

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/rjeczalik/notify"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
)

// A DockerCommand is a command definition for sg run/start that uses
// bazel under the hood. It will handle restarting itself autonomously,
// as long as iBazel is running and watch that specific target.
type DockerCommand struct {
	Name                string
	Container           string            `yaml:"container"`
	Description         string            `yaml:"description"`
	Target              string            `yaml:"target"`
	Args                string            `yaml:"args"`
	PreCmd              string            `yaml:"precmd"`
	Cmd                 string            `yaml:"cmd"`
	Env                 map[string]string `yaml:"env"`
	IgnoreStdout        bool              `yaml:"ignoreStdout"`
	IgnoreStderr        bool              `yaml:"ignoreStderr"`
	ContinueWatchOnExit bool              `yaml:"continueWatchOnExit"`
	// Preamble is a short and visible message, displayed when the command is launched.
	Preamble        string                            `yaml:"preamble"`
	ExternalSecrets map[string]secrets.ExternalSecret `yaml:"external_secrets"`
}

func (dc DockerCommand) GetName() string {
	return dc.Name
}

func (dc DockerCommand) GetContinueWatchOnExit() bool {
	return dc.ContinueWatchOnExit
}

func (dc DockerCommand) GetEnv() map[string]string {
	return dc.Env
}

func (dc DockerCommand) GetIgnoreStdout() bool {
	return dc.IgnoreStdout
}

func (dc DockerCommand) GetIgnoreStderr() bool {
	return dc.IgnoreStderr
}

func (dc DockerCommand) GetPreamble() string {
	return dc.Preamble
}

func (dc DockerCommand) GetBinaryLocation() (string, error) {
	return binaryLocation(dc.Target)
}

func (dc DockerCommand) GetExternalSecrets() map[string]secrets.ExternalSecret {
	return dc.ExternalSecrets
}

func (dc DockerCommand) watchPaths() ([]string, error) {
	// If no target is defined, there is nothing to be built and watched
	if dc.Target == "" {
		return nil, nil
	}
	// Grab the location of the binary in bazel-out.
	binLocation, err := dc.GetBinaryLocation()
	if err != nil {
		return nil, err
	}
	return []string{binLocation}, nil

}

func (dc DockerCommand) StartWatch(ctx context.Context) (<-chan struct{}, error) {
	if watchPaths, err := dc.watchPaths(); err != nil {
		return nil, err
	} else {
		// skip remove events as we don't care about files being removed, we only
		// want to know when the binary has been rebuilt
		return WatchPaths(ctx, watchPaths, notify.Remove)
	}
}

func (dc DockerCommand) GetExecCmd(ctx context.Context) (*exec.Cmd, error) {
	bin, err := dc.GetBinaryLocation()
	if err != nil {
		return nil, err
	}
	cleanup := fmt.Sprintf("docker inspect %s > /dev/null 2>&1 && docker rm -f %s", dc.Name, dc.Name)
	load := fmt.Sprintf("docker load -i %s", bin)
	cmd := fmt.Sprintf("%s\n%s\n%s", cleanup, load, dc.Cmd)
	return exec.CommandContext(ctx, "bash", "-c", cmd), nil
}
