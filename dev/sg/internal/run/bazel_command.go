package run

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rjeczalik/notify"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
)

// A BazelCommand is a command definition for sg run/start that uses
// bazel under the hood. It will handle restarting itself autonomously,
// as long as iBazel is running and watch that specific target.
type BazelCommand struct {
	Name                string
	Description         string            `yaml:"description"`
	Target              string            `yaml:"target"`
	Args                string            `yaml:"args"`
	PreCmd              string            `yaml:"precmd"`
	Env                 map[string]string `yaml:"env"`
	IgnoreStdout        bool              `yaml:"ignoreStdout"`
	IgnoreStderr        bool              `yaml:"ignoreStderr"`
	ContinueWatchOnExit bool              `yaml:"continueWatchOnExit"`
	// Preamble is a short and visible message, displayed when the command is launched.
	Preamble        string                            `yaml:"preamble"`
	ExternalSecrets map[string]secrets.ExternalSecret `yaml:"external_secrets"`

	// RunTarget specifies a target that should be run via `bazel run $RunTarget` instead of directly executing the binary.
	RunTarget string `yaml:"runTarget"`
}

func (bc BazelCommand) GetName() string {
	return bc.Name
}

func (bc BazelCommand) GetContinueWatchOnExit() bool {
	return bc.ContinueWatchOnExit
}

func (bc BazelCommand) GetEnv() map[string]string {
	return bc.Env
}

func (bc BazelCommand) GetIgnoreStdout() bool {
	return bc.IgnoreStdout
}

func (bc BazelCommand) GetIgnoreStderr() bool {
	return bc.IgnoreStderr
}

func (bc BazelCommand) GetPreamble() string {
	return bc.Preamble
}

func (bc BazelCommand) GetBinaryLocation() (string, error) {
	baseOutput, err := outputPath()
	if err != nil {
		return "", err
	}
	// Trim "bazel-out" because the next bazel query will include it.
	outputPath := strings.TrimSuffix(strings.TrimSpace(string(baseOutput)), "bazel-out")

	// Get the binary from the specific target.
	cmd := exec.Command("bazel", "cquery", bc.Target, "--output=files")
	baseOutput, err = cmd.Output()
	if err != nil {
		return "", err
	}
	binPath := strings.TrimSpace(string(baseOutput))

	return fmt.Sprintf("%s%s", outputPath, binPath), nil
}

func (bc BazelCommand) GetExternalSecrets() map[string]secrets.ExternalSecret {
	return bc.ExternalSecrets
}

func (bc BazelCommand) watchPaths() ([]string, error) {
	// If no target is defined, there is nothing to be built and watched
	if bc.Target == "" {
		return nil, nil
	}
	// Grab the location of the binary in bazel-out.
	binLocation, err := bc.GetBinaryLocation()
	if err != nil {
		return nil, err
	}
	return []string{binLocation}, nil

}

func (bc BazelCommand) StartWatch(ctx context.Context) (<-chan struct{}, error) {
	if watchPaths, err := bc.watchPaths(); err != nil {
		return nil, err
	} else {
		// skip remove events as we don't care about files being removed, we only
		// want to know when the binary has been rebuilt
		return WatchPaths(ctx, watchPaths, notify.Remove)
	}
}

func (bc BazelCommand) GetExecCmd(ctx context.Context) (*exec.Cmd, error) {
	var cmd string
	var err error
	if bc.RunTarget != "" {
		cmd = "bazel run " + bc.RunTarget
	} else {
		if cmd, err = bc.GetBinaryLocation(); err != nil {
			return nil, err
		}
	}

	return exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("%s\n%s", bc.PreCmd, cmd)), nil
}

func outputPath() ([]byte, error) {
	// Get the output directory from Bazel, which varies depending on which OS
	// we're running against.
	cmd := exec.Command("bazel", "info", "output_path")
	return cmd.Output()
}
