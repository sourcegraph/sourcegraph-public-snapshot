package command

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
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
			dockerHostGatewayFlag(),
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

// dockerHostGatewayFlag makes the Docker host accessible to the container (on the hostname
// `host.docker.internal`), which simplifies the use of executors when the Sourcegraph instance is
// running uncontainerized in the Docker host. This *only* takes effect if the site config
// `executors.frontendURL` is a URL with hostname `host.docker.internal`, to reduce the risk of
// unexpected compatibility or security issues with using --add-host=...  when it is not needed.
func dockerHostGatewayFlag() []string {
	const hostDockerInternal = "host.docker.internal"
	frontendURL, _ := url.Parse(conf.ExecutorsFrontendURL())
	if frontendURL != nil && frontendURL.Host == hostDockerInternal || strings.HasPrefix(frontendURL.Host, hostDockerInternal+":") {
		return []string{"--add-host=host.docker.internal:host-gateway"}
	}
	return nil
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
