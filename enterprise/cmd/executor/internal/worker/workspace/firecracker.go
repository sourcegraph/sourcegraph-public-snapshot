package workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"

	"github.com/c2h5oh/datasize"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/worker/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFirecrackerWorkspace creates a new workspace for firecracker-based execution.
// A block device will be created on the host disk, with an ext4 file system. It
// is exposed through a loopback device. To set up the workspace, this device will
// be mounted and clone the repo and put script files in it. Then, the executor
// VM can mount this loopback device. This prevents host file system access.
func NewFirecrackerWorkspace(
	ctx context.Context,
	filesStore store.FilesStore,
	job executor.Job,
	diskSpace string,
	keepWorkspace bool,
	commandRunner command.Runner,
	logger command.Logger,
	cloneOpts CloneOptions,
	operations *command.Operations,
) (Workspace, error) {
	blockDeviceFile, tmpMountDir, blockDevice, err := setupLoopDevice(
		ctx,
		job.ID,
		diskSpace,
		keepWorkspace,
		logger,
	)
	if err != nil {
		return nil, err
	}

	// Unmount the workspace volume when done, we finished writing to it from the host.
	defer func() {
		if err2 := syscall.Unmount(tmpMountDir, 0); err2 != nil {
			err = errors.Append(err, err2)
			return
		}
		if err2 := os.RemoveAll(tmpMountDir); err2 != nil {
			err = errors.Append(err, err2)
		}
	}()

	if job.RepositoryName != "" {
		if err := cloneRepo(ctx, tmpMountDir, job, commandRunner, cloneOpts, operations); err != nil {
			return nil, err
		}
	}

	scriptPaths, err := prepareScripts(ctx, filesStore, job, tmpMountDir, logger)
	if err != nil {
		return nil, err
	}

	return &firecrackerWorkspace{
		scriptFilenames: scriptPaths,
		blockDeviceFile: blockDeviceFile,
		blockDevice:     blockDevice,
		logger:          logger,
	}, err
}

type firecrackerWorkspace struct {
	scriptFilenames []string
	blockDeviceFile string
	blockDevice     string
	logger          command.Logger
}

func (w firecrackerWorkspace) Path() string {
	return w.blockDevice
}

func (w firecrackerWorkspace) ScriptFilenames() []string {
	return w.scriptFilenames
}

func (w firecrackerWorkspace) Remove(ctx context.Context, keepWorkspace bool) {
	handle := w.logger.Log("teardown.fs", nil)
	defer func() {
		// We always finish this with exit code 0 even if it errored, because workspace
		// cleanup doesn't fail the execution job. We can deal with it separately.
		handle.Finalize(0)
		handle.Close()
	}()

	if keepWorkspace {
		fmt.Fprintf(handle, "Preserving workspace files (block device: %s, loop file: %s) as per config", w.blockDevice, w.blockDeviceFile)
		// Remount the workspace, so that it can be inspected.
		mountDir, err := mountLoopDevice(ctx, w.blockDevice, handle)
		if err != nil {
			fmt.Fprintf(handle, "Failed to mount workspace device %q, mount manually to inspect the contents: %s\n", w.blockDevice, err)
			return
		}
		fmt.Fprintf(handle, "Inspect the workspace contents at: %s\n", mountDir)
		return
	}

	fmt.Fprintf(handle, "Removing loop device %s\n", w.blockDevice)
	if err := detachLoopDevice(ctx, w.blockDevice, handle); err != nil {
		fmt.Fprintf(handle, "stderr: Failed to detach loop device: %s\n", err)
	}

	fmt.Fprintf(handle, "Removing block device file %s\n", w.blockDeviceFile)
	if err := os.Remove(w.blockDeviceFile); err != nil {
		fmt.Fprintf(handle, "stderr: Failed to remove block device: %s\n", err)
	}
}

// setupLoopDevice is used in firecracker mode. It creates a block device on disk,
// creates a loop device pointing to it, and mounts it so that it can be written to.
// The loop device will be given to ignite and mounted into the guest VM.
func setupLoopDevice(
	ctx context.Context,
	jobID int,
	diskSpace string,
	keepWorkspace bool,
	logger command.Logger,
) (blockDeviceFile, tmpMountDir, blockDevice string, err error) {
	handle := logger.Log("setup.fs.workspace", nil)
	defer func() {
		if err != nil {
			// add the error to the bottom of the step's log output,
			// but only if this isnt from exec.Command, as those get added
			// by our logging wrapper
			if !errors.HasType(err, &exec.ExitError{}) {
				fmt.Fprint(handle, err.Error())
			}
			handle.Finalize(1)
		} else {
			handle.Finalize(0)
		}
		handle.Close()
	}()

	// Create a temp file to hold the block device on disk.
	loopFile, err := MakeTempFile("workspace-loop-" + strconv.Itoa(jobID))
	if err != nil {
		return "", "", "", err
	}
	defer func() {
		if err != nil && !keepWorkspace {
			os.Remove(loopFile.Name())
		}
	}()
	blockDeviceFile = loopFile.Name()
	fmt.Fprintf(handle, "Created backing workspace file at %q\n", blockDeviceFile)

	// Truncate the file to be of the size of the maximum permissible disk space.
	diskSize, err := datasize.ParseString(diskSpace)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "invalid disk size provided: %q", diskSpace)
	}
	if err := loopFile.Truncate(int64(diskSize.Bytes())); err != nil {
		return "", "", "", errors.Wrapf(err, "failed to make backing file sparse with %d bytes", diskSize.Bytes())
	}
	fmt.Fprintf(handle, "Created sparse file of size %s from %q\n", diskSize.HumanReadable(), blockDeviceFile)
	if err := loopFile.Close(); err != nil {
		return "", "", "", errors.Wrap(err, "failed to close backing file")
	}

	// Create an ext4 file system in the device backing file.
	out, err := commandLogger(ctx, handle, "mkfs.ext4", blockDeviceFile)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "failed to create ext4 filesystem in backing file: %q", out)
	}

	fmt.Fprintf(handle, "Wrote ext4 filesystem to backing file %q\n", blockDeviceFile)

	// Create a loop device pointing to our block device.
	out, err = commandLogger(ctx, handle, "losetup", "--find", "--show", blockDeviceFile)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "failed to create loop device: %q", out)
	}
	blockDevice = strings.TrimSpace(out)
	defer func() {
		// If something further down in this function failed we detach the loop device
		// to not hoard them.
		if err != nil {
			err := detachLoopDevice(ctx, blockDevice, handle)
			if err != nil {
				fmt.Fprint(handle, "stderr: "+strings.ReplaceAll(strings.TrimSpace(err.Error()), "\n", "\nstderr: "))
			}
		}
	}()
	fmt.Fprintf(handle, "Created loop device at %q backed by %q\n", blockDevice, blockDeviceFile)

	// Mount the loop device at a temporary directory so we can write the workspace contents to it.
	tmpMountDir, err = mountLoopDevice(ctx, blockDevice, handle)
	if err != nil {
		// important to set at least blockDevice for the above defer
		return blockDeviceFile, "", blockDevice, err
	}
	fmt.Fprintf(handle, "Created temporary workspace mount location at %q\n", tmpMountDir)

	return blockDeviceFile, tmpMountDir, blockDevice, nil
}

// detachLoopDevice detaches a loop device by path (/dev/loopX).
func detachLoopDevice(ctx context.Context, blockDevice string, handle command.LogEntry) error {
	out, err := commandLogger(ctx, handle, "losetup", "--detach", blockDevice)
	if err != nil {
		return errors.Wrapf(err, "failed to detach loop device: %s", out)
	}
	return nil
}

// mountLoopDevice takes a path to a loop device (/dev/loopX) and mounts it at a
// random temporary mount point. The mount point is returned.
func mountLoopDevice(ctx context.Context, blockDevice string, handle command.LogEntry) (string, error) {
	tmpMountDir, err := MakeTempDirectory("workspace-mountpoints")
	if err != nil {
		return "", err
	}

	if out, err := commandLogger(ctx, handle, "mount", blockDevice, tmpMountDir); err != nil {
		_ = os.RemoveAll(tmpMountDir)
		return "", errors.Wrapf(err, "failed to mount loop device %q to %q: %q", blockDevice, tmpMountDir, out)
	}

	return tmpMountDir, nil
}
