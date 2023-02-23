package command

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/util"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
)

// ScriptsPath is the location relative to the executor workspace where the executor
// will write scripts required for the execution of the job.
const ScriptsPath = ".sourcegraph-executor"

type DockerOptions struct {
	Key              string
	Image            string
	Command          []string
	Dir              string
	Env              []string
	ScriptPath       string
	ConfigPath       string
	DockerAuthConfig types.DockerAuthConfig
	AddHostGateway   bool
	Resources        ResourceOptions
}

type ResourceOptions struct {
	// NumCPUs is the number of virtual CPUs a container or VM can use.
	NumCPUs int
	// Memory is the maximum amount of memory a container or VM can use.
	Memory string
	// DiskSpace is the maximum amount of disk a container or VM can use.
	// Only available in firecracker.
	DiskSpace string
	// MaxIngressBandwidth configures the maximum permissible ingress bytes per second
	// per job. Only available in Firecracker.
	MaxIngressBandwidth int
	// MaxEgressBandwidth configures the maximum permissible egress bytes per second
	// per job. Only available in Firecracker.
	MaxEgressBandwidth int
	// DockerHostMountPath, if supplied, replaces the workspace parent directory in the
	// volume mounts of Docker containers. This option is used when running privileged
	// executors in k8s or docker-compose without requiring the host and node paths to
	// be identical.
	DockerHostMountPath string
}

// NewDockerCommand constructs the command to run on the host in order to
// invoke the given spec. If the spec does not specify an image, then the command
// will be run _directly_ on the host. Otherwise, the command will be run inside
// a one-shot docker container subject to the resource limits specified in the
// given options.
func NewDockerCommand(logger Logger, cmdRunner util.CmdRunner, workingDir string, options DockerOptions) Command {
	// TODO - remove this once src-cli is not required anymore for SSBC.
	if options.Image == "" {
		env := options.Env
		if options.ConfigPath != "" {
			env = append(env, fmt.Sprintf("DOCKER_CONFIG=%s", options.ConfigPath))
		}
		return Command{
			Key:     options.Key,
			Command: options.Command,
			Dir:     filepath.Join(workingDir, options.Dir),
			Env:     env,
		}
	}

	hostDir := workingDir
	if options.Resources.DockerHostMountPath != "" {
		hostDir = filepath.Join(options.Resources.DockerHostMountPath, filepath.Base(workingDir))
	}

	return Command{
		Key:       options.Key,
		Command:   formatDockerCommand(hostDir, options),
		CmdRunner: cmdRunner,
		Logger:    logger,
	}
}

func formatDockerCommand(hostDir string, options DockerOptions) []string {
	return Flatten(
		"docker",
		dockerConfigFlag(options.ConfigPath),
		"run",
		"--rm",
		dockerHostGatewayFlag(options.AddHostGateway),
		dockerResourceFlags(options.Resources),
		dockerVolumeFlags(hostDir),
		dockerWorkingDirectoryFlags(options.Dir),
		dockerEnvFlags(options.Env),
		dockerEntrypointFlags,
		options.Image,
		filepath.Join("/data", ScriptsPath, options.ScriptPath),
	)
}

// dockerHostGatewayFlag makes the Docker host accessible to the container (on the hostname
// `host.docker.internal`), which simplifies the use of executors when the Sourcegraph instance is
// running un-containerized in the Docker host. This *only* takes effect if the site config
// `executors.frontendURL` is a URL with hostname `host.docker.internal`, to reduce the risk of
// unexpected compatibility or security issues with using --add-host=...  when it is not needed.
func dockerHostGatewayFlag(shouldAdd bool) []string {
	if shouldAdd {
		return dockerGatewayHost
	}
	return nil
}

var dockerGatewayHost = []string{"--add-host=host.docker.internal:host-gateway"}

func dockerResourceFlags(options ResourceOptions) []string {
	flags := make([]string, 0, 4)
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

func dockerWorkingDirectoryFlags(dir string) []string {
	return []string{"-w", filepath.Join("/data", dir)}
}

func dockerEnvFlags(env []string) []string {
	return Intersperse("-e", env)
}

var dockerEntrypointFlags = []string{"--entrypoint", "/bin/sh"}
