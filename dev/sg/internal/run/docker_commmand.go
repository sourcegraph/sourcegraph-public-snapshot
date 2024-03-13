package run

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/rjeczalik/notify"
)

// A DockerCommand is a command definition for sg run/start that uses
// bazel under the hood. It will handle restarting itself autonomously,
// as long as iBazel is running and watch that specific target.
type DockerCommand struct {
	Config    SGConfigCommandOptions
	Container string `yaml:"container"`
	Target    string `yaml:"target"`
	Args      string `yaml:"args"`
	PreCmd    string `yaml:"precmd"`
	Cmd       string `yaml:"cmd"`
}

// UnmarshalYAML implements the Unmarshaler interface for DockerCommand.
// This allows us to parse the flat YAML configuration into nested struct.
func (dc *DockerCommand) UnmarshalYAML(unmarshal func(any) error) error {
	// In order to not recurse infinitely (calling UnmarshalYAML over and over) we create a
	// temporary type alias.
	// First parse the DockerCommand specific options
	type rawDocker DockerCommand
	if err := unmarshal((*rawDocker)(dc)); err != nil {
		return err
	}

	// Then parse the common options from the same list into a nested struct
	return unmarshal(&dc.Config)
}

func (dc DockerCommand) GetConfig() SGConfigCommandOptions {
	return dc.Config
}

func (dc DockerCommand) GetBinaryLocation() (string, error) {
	return binaryLocation(dc.Target)
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
	cleanup := fmt.Sprintf("docker inspect %s > /dev/null 2>&1 && docker rm -f %s", dc.Config.Name, dc.Config.Name)
	load := fmt.Sprintf("docker load -i %s", bin)
	cmd := fmt.Sprintf("%s\n%s\n%s", cleanup, load, dc.Cmd)
	return exec.CommandContext(ctx, "bash", "-c", cmd), nil
}

func (dc *DockerCommand) GetOptions() *SGConfigCommandOptions {
	return &dc.Config
}
