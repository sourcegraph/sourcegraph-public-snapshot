package command

import (
	"fmt"
	"path/filepath"
	"strconv"
)

// ScriptsPath is the location relative to the executor workspace where the executor
// will write scripts required for the execution of the job.
const ScriptsPath = ".sourcegraph-executor"

// formatRawOrDockerCommand constructs the command to run on the host in order to
// invoke the given spec. If the spec does not specify an image, then the command
// will be run _directly_ on the host. Otherwise, the command will be run inside
// of a one-shot docker container subject to the resource limits specified in the
// given options.
func formatRawOrDockerCommand(spec CommandSpec, dir string, options Options, dockerConfigPath string) command {
	// TODO - remove this once src-cli is not required anymore for SSBC.
	if spec.Image == "" {
		env := spec.Env
		if dockerConfigPath != "" {
			env = append(env, fmt.Sprintf("DOCKER_CONFIG=%s", dockerConfigPath))
		}
		return command{
			Key:       spec.Key,
			Command:   spec.Command,
			Dir:       filepath.Join(dir, spec.Dir),
			Env:       env,
			Operation: spec.Operation,
		}
	}

	hostDir := dir
	if options.ResourceOptions.DockerHostMountPath != "" {
		hostDir = filepath.Join(options.ResourceOptions.DockerHostMountPath, filepath.Base(dir))
	}

	return command{
		Key: spec.Key,
		Command: flatten(
			"docker",
			dockerConfigFlag(dockerConfigPath),
			"run", "--rm",
			dockerResourceFlags(options.ResourceOptions),
			dockerVolumeFlags(hostDir),
			dockerWorkingdirectoryFlags(spec.Dir),
			dockerEnvFlags(spec.Env),
			dockerEntrypointFlags(),
			spec.Image,
			filepath.Join("/data", ScriptsPath, spec.ScriptPath),
		),
		Operation: spec.Operation,
	}
}

func dockerResourceFlags(options ResourceOptions) []string {
	flags := make([]string, 0, 2)
	if options.NumCPUs != 0 {
		flags = append(flags, "--cpus", strconv.Itoa(options.NumCPUs))
	}
	if options.Memory != "0" && options.Memory != "" {
		flags = append(flags, "--memory", options.Memory)
	}

	return flags
}

func dockerVolumeFlags(wd string) []string {
	return []string{"-v", wd + ":/data"}
}

func dockerConfigFlag(dockerConfigPath string) []string {
	if dockerConfigPath == "" {
		return nil
	}
	return []string{"--config", dockerConfigPath}
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
