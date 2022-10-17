package command

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ScriptsPath is the location relative to the executor workspace where the executor
// will write scripts required for the execution of the job.
const ScriptsPath = ".sourcegraph-executor"

// formatRawOrDockerCommand constructs the command to run on the host in order to
// invoke the given spec. If the spec does not specify an image, then the command
// will be run _directly_ on the host. Otherwise, the command will be run inside
// of a one-shot docker container subject to the resource limits specified in the
// given options.
func formatRawOrDockerCommand(ctx context.Context, spec CommandSpec, dir string, options Options) command {
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

	hostDir := dir
	if options.ResourceOptions.DockerHostMountPath != "" {
		hostDir = filepath.Join(options.ResourceOptions.DockerHostMountPath, filepath.Base(dir))
	}

	cidFile, cleanup, err := createCidFile(ctx)
	if err != nil {
		// TODO
		panic(err)
	}

	return command{
		Key: spec.Key,
		Command: flatten(
			"docker",
			"run",
			"--rm",
			// Use tini to make sure all zombie processes are always cleaned up
			// properly. This is to prevent any host pollution and potentially
			// undesirable side-effects from those.
			"--init",
			dockerCIDFlags(cidFile),
			dockerResourceFlags(options.ResourceOptions),
			dockerVolumeFlags(hostDir),
			dockerWorkingdirectoryFlags(spec.Dir),
			// If the env vars will be part of the command line args, we need to quote them
			dockerEnvFlags(quoteEnv(spec.Env)),
			dockerEntrypointFlags(),
			spec.Image,
			filepath.Join("/data", ScriptsPath, spec.ScriptPath),
		),
		Cleanup:   cleanup,
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

func dockerCIDFlags(cidFile string) []string {
	return []string{"--cidfile", cidFile}
}

func dockerVolumeFlags(wd string) []string {
	return []string{"-v", wd + ":/data"}
}

func dockerWorkingdirectoryFlags(dir string) []string {
	return []string{"--workdir", filepath.Join("/data", dir)}
}

func dockerEnvFlags(env []string) []string {
	return intersperse("-e", env)
}

func dockerEntrypointFlags() []string {
	return []string{"--entrypoint", "/bin/sh"}
}

// createCidFile creates a temporary file that will contain the container ID
// when executing steps.
// It returns the location of the file and a function that cleans up the
// file.
func createCidFile(ctx context.Context) (string, func(ctx context.Context) error, error) {
	// Find a location that we can use for a cidfile, which will contain the
	// container ID that is used below. We can then use this to remove the
	// container on a successful run, rather than leaving it dangling.
	cidFile, err := os.CreateTemp("", "executor-docker-container-id")
	if err != nil {
		return "", nil, errors.Wrap(err, "Creating a CID file failed")
	}

	// However, Docker will fail if the cidfile actually exists, so we need
	// to remove it. Because Windows can't remove open files, we'll first
	// close it, even though that's unnecessary elsewhere.
	cidFile.Close()
	if err = os.Remove(cidFile.Name()); err != nil {
		return "", nil, errors.Wrap(err, "removing cidfile")
	}

	// Since we went to all that effort, we can now defer a function that
	// uses the cidfile to clean up after this function is done.
	cleanup := func(ctx context.Context) error {
		var errs error
		cid, readFileErr := os.ReadFile(cidFile.Name())
		errs = errors.Append(errs, readFileErr)
		rmFileErr := os.Remove(cidFile.Name())
		errs = errors.Append(errs, rmFileErr)
		if readFileErr == nil {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
			errs = errors.Append(errs, exec.CommandContext(ctx, "docker", "rm", "-f", "--", string(cid)).Run())
		}
		return errs
	}

	return cidFile.Name(), cleanup, nil
}
