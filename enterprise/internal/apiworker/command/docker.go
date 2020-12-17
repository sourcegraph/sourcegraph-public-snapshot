package command

import (
	"path/filepath"
	"strconv"
)

// formatRawOrDockerCommand constructs the command to run on the host in order to
// invoke the given spec. If the spec does not specify an image, then the command
// will be run _directly_ on the host. Otherwise, the command will be run inside
// of a one-shot docker container subject to the resource limits specified in the
// given options.
func formatRawOrDockerCommand(spec CommandSpec, dir string, options Options) command {
	// TODO - make this a non-special case
	if spec.Image == "" {
		return command{
			Key:       spec.Key,
			Command:   spec.Command,
			Dir:       filepath.Join(dir, spec.Dir),
			Env:       spec.Env,
			Operation: spec.Operation,
		}
	}

	return command{
		Key: spec.Key,
		Command: flatten(
			"docker", "run", "--rm",
			dockerResourceFlags(options.ResourceOptions),
			dockerVolumeFlags(dir, spec.ScriptPath),
			dockerWorkingdirectoryFlags(spec.Dir),
			dockerEnvFlags(spec.Env),
			dockerEntrypointFlags(),
			spec.Image,
			spec.ScriptPath,
		),
		Operation: spec.Operation,
	}
}

func dockerResourceFlags(options ResourceOptions) []string {
	return []string{
		"--cpus", strconv.Itoa(options.NumCPUs),
		"--memory", options.Memory,
	}
}

func dockerVolumeFlags(wd, scriptPath string) []string {
	return []string{"-v", wd + ":/data", "-v", scriptPath + ":" + scriptPath}
}

func dockerWorkingdirectoryFlags(dir string) []string {
	return []string{"-w", filepath.Join("/data", dir)}
}

func dockerEnvFlags(env []string) []string {
	return intersperse("-e", env)
}

func dockerEntrypointFlags() []string {
	return []string{"--entrypoint", "/bin/sh"}
}
