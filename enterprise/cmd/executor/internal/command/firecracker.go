package command

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/sync/errgroup"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/inconshreveable/log15"
	"github.com/weaveworks/ignite/cmd/ignite/run"
	"github.com/weaveworks/ignite/pkg/apis/ignite"
	"github.com/weaveworks/ignite/pkg/apis/ignite/validation"
	meta "github.com/weaveworks/ignite/pkg/apis/meta/v1alpha1"
	"github.com/weaveworks/ignite/pkg/config"
	"github.com/weaveworks/ignite/pkg/constants"
	igniteNetwork "github.com/weaveworks/ignite/pkg/network"
	"github.com/weaveworks/ignite/pkg/operations"
	"github.com/weaveworks/ignite/pkg/preflight/checkers"
	"github.com/weaveworks/ignite/pkg/providers"
	"github.com/weaveworks/ignite/pkg/providers/client"
	"github.com/weaveworks/ignite/pkg/providers/manifeststorage"
	"github.com/weaveworks/ignite/pkg/providers/storage"
	igniteRuntime "github.com/weaveworks/ignite/pkg/runtime"
	issh "github.com/weaveworks/ignite/pkg/ssh"
	"github.com/weaveworks/libgitops/pkg/filter"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const firecrackerContainerDir = "/work"

var igniteProviderSync sync.Once

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
func setupFirecracker(ctx context.Context, logger Logger, name, repoDir string, options Options, op *Operations) (err error) {
	// Start the VM and wait for the SSH server to become available

	err = func() (err error) {
		_, _, endObservation := op.SetupFirecrackerStart.With(ctx, &err, observation.Args{})
		defer endObservation(1, observation.Args{})

		handle := logger.Log("setup.firecracker.start", flatten(
			"ignite", "run",
			"--runtime", "docker",
			"--network-plugin", "cni",
			firecrackerResourceFlags(options.ResourceOptions),
			firecrackerCopyfileFlags(repoDir, options.FirecrackerOptions.VMStartupScriptPath),
			"--ssh",
			"--name", name,
			sanitizeImage(options.FirecrackerOptions.Image),
		))
		defer func() {
			if err != nil {
				handle.Finalize(1)
			} else {
				handle.Finalize(0)
			}
			handle.Close()
		}()

		igniteProviderSync.Do(func() {
			if err := manifeststorage.SetManifestStorage(); err != nil {
				panic(fmt.Sprintf("failed to set ignite manifest storage: %v", err))
			}
			if err := storage.SetGenericStorage(); err != nil {
				panic(fmt.Sprintf("failed to set generic ignite storage: %v", err))
			}
			if err := client.SetClient(); err != nil {
				panic(fmt.Sprintf("failed to set ignite client: %v", err))
			}
			if err := config.SetAndPopulateProviders(igniteRuntime.RuntimeDocker, igniteNetwork.PluginCNI); err != nil {
				panic(fmt.Sprintf("failed to populate ignite providers: %v", err))
			}
		})

		baseVM := providers.Client.VMs().New()

		baseVM.Name = name

		baseVM.Status.Runtime.Name = igniteRuntime.RuntimeDocker
		baseVM.Status.Network.Plugin = igniteNetwork.PluginCNI

		ociRef, err := meta.NewOCIImageRef(sanitizeImage(options.FirecrackerOptions.Image))
		if err != nil {
			return errors.Wrap(err, "failed to create OCI image ref")
		}
		baseVM.Spec.Image.OCI = ociRef
		baseVM.Spec.CPUs = uint64(options.ResourceOptions.NumCPUs)
		baseVM.Spec.Memory, _ = meta.NewSizeFromString(options.ResourceOptions.Memory)
		baseVM.Spec.DiskSize, _ = meta.NewSizeFromString(options.ResourceOptions.DiskSpace)

		if repoDir != "" {
			baseVM.Spec.CopyFiles = append(baseVM.Spec.CopyFiles, ignite.FileMapping{
				HostPath: repoDir,
				VMPath:   firecrackerContainerDir,
			})
		}
		if options.FirecrackerOptions.VMStartupScriptPath != "" {
			baseVM.Spec.CopyFiles = append(baseVM.Spec.CopyFiles, ignite.FileMapping{
				HostPath: options.FirecrackerOptions.VMStartupScriptPath,
				VMPath:   options.FirecrackerOptions.VMStartupScriptPath,
			})
		}

		baseVM.Spec.SSH = &ignite.SSH{Generate: true}

		if err := validation.ValidateVM(baseVM).ToAggregate(); err != nil {
			return errors.Wrap(err, "build invalid ignite vm")
		}

		if err := callWithInstrumentedLock(op, logger, func() error {
			img, err := operations.FindOrImportImage(ctx, providers.Client, baseVM.Spec.Image.OCI)
			if err != nil {
				return errors.Wrap(err, "failed to import OCI image")
			}
			baseVM.SetImage(img)

			kernel, err := operations.FindOrImportKernel(ctx, providers.Client, baseVM.Spec.Kernel.OCI)
			if err != nil {
				return errors.Wrap(err, "failed to import kernel")
			}
			baseVM.SetKernel(kernel)

			return nil
		}); err != nil {
			return errors.Wrap(err, "failed to start firecracker vm")
		}

		createOpts := &run.CreateOptions{CreateFlags: &run.CreateFlags{
			CopyFiles:   firecrackerCopyfileFlags(repoDir, options.FirecrackerOptions.VMStartupScriptPath),
			SSH:         ignite.SSH{Generate: true},
			VM:          baseVM,
			RequireName: true,
		}}

		if err := run.Create(ctx, createOpts); err != nil {
			return errors.Wrap(err, "failed to create vm")
		}

		if err := checkers.StartCmdChecks(baseVM, sets.String{}); err != nil {
			return errors.Wrap(err, "failed pre-start checks")
		}

		if err := operations.StartVM(ctx, baseVM, false); err != nil {
			return errors.Wrap(err, "failed to start ignite vm")
		}

		return errors.Wrap(issh.WaitForSSH(ctx, baseVM, constants.SSH_DEFAULT_TIMEOUT_SECONDS, constants.IGNITE_SPAWN_TIMEOUT), "ssh failed to start in ignite vm")
	}()
	if err != nil {
		return err
	}

	if options.FirecrackerOptions.VMStartupScriptPath != "" {
		cmd := command{
			Key:       "setup.startup-script",
			Command:   flatten("ignite", "exec", name, "--", options.FirecrackerOptions.VMStartupScriptPath),
			Operation: op.SetupStartupScript,
		}
		return errors.Wrap(execFirecracker(ctx, name, cmd, logger), "failed to run startup script")
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
func callWithInstrumentedLock(operations *Operations, logger Logger, f func() error) error {
	handle := logger.Log("setup.firecracker.runlock", nil)

	lockRequestedAt := time.Now()

	igniteRunLock.Lock()

	lockAcquiredAt := time.Now()

	handle.Finalize(0)
	handle.Close()

	err := f()

	lockReleasedAt := time.Now()

	igniteRunLock.Unlock()

	operations.RunLockWaitTotal.Add(float64(lockAcquiredAt.Sub(lockRequestedAt) / time.Millisecond))
	operations.RunLockHeldTotal.Add(float64(lockReleasedAt.Sub(lockAcquiredAt) / time.Millisecond))
	return err
}

func execFirecracker(ctx context.Context, name string, cmd command, logger Logger) (err error) {
	_, _, endObservation := cmd.Operation.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	outWriter, errWriter, cleanup := igniteExecLogger(cmd.Key, cmd.Command, logger)
	defer func() {
		if cerr := cleanup(err); err == nil && cerr != nil {
			err = cerr
		}
	}()

	// trim 'firecracker exec <name> --' from the command being passed in, but keep for display purposes
	execOpts, err := new(run.ExecFlags).NewExecOptions(name, outWriter, errWriter, nil, cmd.Command[4:]...)
	if err != nil {
		return errors.Wrap(err, "error creating exec options")
	}

	if err := run.Exec(ctx, execOpts); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// teardownFirecracker issues a stop and a remove request for the Firecracker VM with
// the given name.
func teardownFirecracker(ctx context.Context, logger Logger, name string, operations *Operations) (err error) {
	_, _, endObservation := operations.TeardownFirecrackerRemove.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	handle := logger.Log("teardown.firecracker.remove", flatten("ignite", "rm", "-f", name))
	defer func() {
		if err != nil {
			handle.Finalize(1)
		} else {
			handle.Finalize(0)
		}
		handle.Close()
	}()

	vm, err := providers.Client.VMs().Find(filter.NewIDNameFilter(name))
	if err != nil {
		return errors.Wrapf(err, "no vm found with name %s", name)
	}

	if derr := providers.Client.VMs().Delete(vm.UID); derr != nil {
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

func igniteExecLogger(key string, command []string, logger Logger) (stdout io.WriteCloser, stderr io.WriteCloser, cleanup func(error) error) {
	handle := logger.Log(key, command)

	outReader, outWriter := io.Pipe()
	errReader, errWriter := io.Pipe()

	eg := &errgroup.Group{}

	eg.Go(func() error {
		return logLineWriter(handle)("stdout", outReader)
	})
	eg.Go(func() error {
		return logLineWriter(handle)("stderr", errReader)
	})

	return outWriter, errWriter, func(err error) error {
		outWriter.Close()
		errWriter.Close()
		var sshErr *ssh.ExitError
		if errors.As(err, &sshErr) {
			handle.Finalize(sshErr.ExitStatus())
		} else if err != nil {
			handle.Finalize(1)
		} else {
			handle.Finalize(0)
		}
		handle.Close()
		return errors.Wrap(eg.Wait(), "error reading vm output pipes")
	}
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
