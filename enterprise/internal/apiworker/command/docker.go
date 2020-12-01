package command

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// formatRawOrDockerCommand constructs the command to run on the host in order to
// invoke the given spec. If the spec does not specify an image, then the command
// will be run _directly_ on the host. Otherwise, the command will be run inside
// of a one-shot docker container subject to the resource limits specified in the
// given options.
func formatRawOrDockerCommand(spec CommandSpec, dir string, options Options) command {
	if spec.Image == "" {
		return command{
			Key:      spec.Key,
			Commands: spec.Commands,
			Dir:      filepath.Join(dir, spec.Dir),
			Env:      spec.Env,
		}
	}

	return command{
		Key: spec.Key,
		Commands: flatten(
			"docker", "run", "--rm",
			dockerResourceFlags(options.ResourceOptions),
			dockerVolumeFlags(dir),
			dockerWorkingdirectoryFlags(spec.Dir),
			dockerEnvFlags(spec.Env),
			spec.Image,
			spec.Commands,
		),
	}
}

func dockerResourceFlags(options ResourceOptions) []string {
	return []string{
		"--cpus", strconv.Itoa(options.NumCPUs),
		"--memory", options.Memory,
	}
}

func dockerVolumeFlags(wd string) []string {
	return []string{"-v", fmt.Sprintf("%s:/data", wd)}
}

func dockerWorkingdirectoryFlags(dir string) []string {
	return []string{"-w", filepath.Join("/data", dir)}
}

func dockerEnvFlags(env []string) []string {
	return intersperse("-e", env)
}
