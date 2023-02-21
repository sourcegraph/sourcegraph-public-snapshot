package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ScriptsPath is the location relative to the executor workspace where the executor
// will write scripts required for the execution of the job.
const ScriptsPath = ".sourcegraph-executor"

type dockerRunner struct {
	dir       string
	logger    log.Logger
	cmdLogger Logger
	options   Options
	// tmpDir is used to store temporary files used for docker execution.
	tmpDir           string
	dockerConfigPath string
}

var _ Runner = &dockerRunner{}

func (r *dockerRunner) Setup(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "executor-docker-runner")
	if err != nil {
		return errors.Wrap(err, "failed to create tmp dir for docker runner")
	}
	r.tmpDir = dir

	// If docker auth config is present, write it.
	if len(r.options.DockerOptions.DockerAuthConfig.Auths) > 0 {
		d, err := json.Marshal(r.options.DockerOptions.DockerAuthConfig)
		if err != nil {
			return err
		}
		r.dockerConfigPath, err = os.MkdirTemp(r.tmpDir, "docker_auth")
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(r.dockerConfigPath, "config.json"), d, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

func (r *dockerRunner) Teardown(ctx context.Context) error {
	if err := os.RemoveAll(r.tmpDir); err != nil {
		r.logger.Error("Failed to remove docker state tmp dir", log.String("tmpDir", r.tmpDir), log.Error(err))
	}

	return nil
}

func (r *dockerRunner) Run(ctx context.Context, command Spec) error {
	return runCommand(ctx, r.logger, formatRawOrDockerCommand(command, r.dir, r.options, r.dockerConfigPath), r.cmdLogger)
}

// formatRawOrDockerCommand constructs the command to run on the host in order to
// invoke the given spec. If the spec does not specify an image, then the command
// will be run _directly_ on the host. Otherwise, the command will be run inside
// of a one-shot docker container subject to the resource limits specified in the
// given options.
func formatRawOrDockerCommand(spec Spec, dir string, options Options, dockerConfigPath string) command {
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
			"run",
			"--rm",
			dockerHostGatewayFlag(options.DockerOptions.AddHostGateway),
			dockerResourceFlags(options.ResourceOptions),
			dockerVolumeFlags(hostDir),
			dockerWorkingDirectoryFlags(spec.Dir),
			dockerEnvFlags(spec.Env),
			dockerEntrypointFlags,
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
	return intersperse("-e", env)
}

var dockerEntrypointFlags = []string{"--entrypoint", "/bin/sh"}
