package command

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type commandRunner interface {
	RunCommand(ctx context.Context, command command, logger *Logger) error
}

const firecrackerContainerDir = "/work"

// formatFirecrackerCommand constructs the command to run on the host via a Firecracker
// virtual machine in order to invoke the given spec. If the spec specifies an image, then
// the command will be run inside of a container inside of the VM. Otherwise, the command
// will be run inside of the VM. The containers are one-shot and subject to the resource
// limits specified in the given options.
//
// The name value supplied here refers to the Firecracker virtual machine, which must have
// also been the name supplied to a successful invocation of setupFirecracker. Additionally,
// the virtual machine must not yet have been torn down (via teardownFirecracker).
func formatFirecrackerCommand(spec CommandSpec, name string, options Options) command {
	rawOrDockerCommand := formatRawOrDockerCommand(spec, firecrackerContainerDir, options)

	innerCommand := strings.Join(rawOrDockerCommand.Command, " ")
	if len(rawOrDockerCommand.Env) > 0 {
		// If we have env vars that are arguments to the command we need to escape them
		quotedEnv := quoteEnv(rawOrDockerCommand.Env)
		innerCommand = fmt.Sprintf("%s %s", strings.Join(quotedEnv, " "), innerCommand)
	}
	if rawOrDockerCommand.Dir != "" {
		innerCommand = fmt.Sprintf("cd %s && %s", rawOrDockerCommand.Dir, innerCommand)
	}

	return command{
		Key:       spec.Key,
		Command:   []string{"ignite", "exec", name, "--", innerCommand},
		Operation: spec.Operation,
	}
}

// setupFirecracker invokes a set of commands to provision and prepare a Firecracker virtual
// machine instance. If a startup script path (an executable file on the host) is supplied,
// it will be mounted into the new virtual machine instance and executed.
func setupFirecracker(ctx context.Context, runner commandRunner, logger *Logger, name, repoDir string, options Options, operations *Operations) error {
	// Start the VM and wait for the SSH server to become available
	startCommand := command{
		Key: "setup.firecracker.start",
		Command: flatten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "cni",
			firecrackerResourceFlags(options.ResourceOptions),
			firecrackerCopyfileFlags(repoDir, options.FirecrackerOptions.VMStartupScriptPath),
			"--ssh",
			"--name", name,
			sanitizeImage(options.FirecrackerOptions.Image),
		),
		Operation: operations.SetupFirecrackerStart,
	}

	if err := callWithInstrumentedLock(operations, func() error { return runner.RunCommand(ctx, startCommand, logger) }); err != nil {
		return errors.Wrap(err, "failed to start firecracker vm")
	}

	if options.FirecrackerOptions.VMStartupScriptPath != "" {
		startupScriptCommand := command{
			Key:       "setup.startup-script",
			Command:   flatten("ignite", "exec", name, "--", options.FirecrackerOptions.VMStartupScriptPath),
			Operation: operations.SetupStartupScript,
		}
		if err := runner.RunCommand(ctx, startupScriptCommand, logger); err != nil {
			return errors.Wrap(err, "failed to run startup script")
		}
	}

	return nil
}

// We've recently seen issues with concurent VM creation. It's likely we
// can do better here and run an empty VM at application startup, but I
// want to do this quick and dirty to see if we can raise our concurrency
// without other issues.
//
// https://github.com/weaveworks/ignite/issues/559
// Following up in https://github.com/sourcegraph/sourcegraph/issues/21377.
var igniteRunLock sync.Mutex

// callWithInstrumentedLock calls f while holding the igniteRunLock. The duration of the wait
// and active portions of this method are emitted as prometheus metrics.
func callWithInstrumentedLock(operations *Operations, f func() error) error {
	lockRequestedAt := time.Now()
	igniteRunLock.Lock()
	lockAcquiredAt := time.Now()
	err := f()
	lockReleasedAt := time.Now()
	igniteRunLock.Unlock()

	operations.RunLockWaitTotal.Add(float64(lockAcquiredAt.Sub(lockRequestedAt) / time.Millisecond))
	operations.RunLockHeldTotal.Add(float64(lockReleasedAt.Sub(lockAcquiredAt) / time.Millisecond))
	return err
}

// teardownFirecracker issues a stop and a remove request for the Firecracker VM with
// the given name.
func teardownFirecracker(ctx context.Context, runner commandRunner, logger *Logger, name string, operations *Operations) error {
	removeCommand := command{
		Key:       "teardown.firecracker.remove",
		Command:   flatten("ignite", "rm", "-f", name),
		Operation: operations.TeardownFirecrackerRemove,
	}
	if err := runner.RunCommand(ctx, removeCommand, logger); err != nil {
		log15.Error("Failed to remove firecracker vm", "name", name, "err", err)
	}

	return nil
}

func firecrackerResourceFlags(options ResourceOptions) []string {
	return []string{
		"--cpus", strconv.Itoa(options.NumCPUs),
		"--memory", options.Memory,
		"--size", options.DiskSpace,
	}
}

func firecrackerCopyfileFlags(dir, vmStartupScriptPath string) []string {
	copyfiles := make([]string, 0, 2)
	if dir != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", dir, firecrackerContainerDir))
	}
	if vmStartupScriptPath != "" {
		copyfiles = append(copyfiles, fmt.Sprintf("%s:%s", vmStartupScriptPath, vmStartupScriptPath))
	}

	sort.Strings(copyfiles)
	return intersperse("--copy-files", copyfiles)
}

var imagePattern = lazyregexp.New(`([^:@]+)(?::([^@]+))?(?:@sha256:([a-z0-9]{64}))?`)

// sanitizeImage sanitizes the given docker image for use by ignite. The ignite utility
// has some issue parsing docker tags that include a sha256 hash, so we try to remove it
// from any of the image references before passing it to the ignite command.
func sanitizeImage(image string) string {
	if matches := imagePattern.FindStringSubmatch(image); len(matches) == 4 {
		if matches[2] == "" {
			return matches[1]
		}

		return fmt.Sprintf("%s:%s", matches[1], matches[2])
	}

	return image
}
