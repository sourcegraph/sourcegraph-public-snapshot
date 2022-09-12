package worker

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/c2h5oh/datasize"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/command"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const SchemeExecutorToken = "token-executor"

// These env vars should be set for git commands. We want to make sure it never hangs on interactive input.
var gitStdEnv = []string{"GIT_TERMINAL_PROMPT=0"}

type Workspace interface {
	// Path represents the block device path when firecracker is enabled and the
	// directory when firecracker is disabled where the workspace is configured.
	Path() string
	// ScriptFilenames holds the ordered set of script filenames to be invoked.
	ScriptFilenames() []string
	Remove(ctx context.Context)
}

type firecrackerWorkspace struct {
	path                  string
	scriptFilenames       []string
	keepWorkspaces        bool
	loopPath, blockDevice string
	commandLogger         command.Logger
}

func (w firecrackerWorkspace) Path() string {
	return w.path
}

func (w firecrackerWorkspace) ScriptFilenames() []string {
	return w.scriptFilenames
}

func (w firecrackerWorkspace) Remove(ctx context.Context) {
	handle := w.commandLogger.Log("teardown.fs", nil)
	defer handle.Close()

	if w.keepWorkspaces {
		fmt.Fprintf(handle, "Preserving workspace files (block device: %s, loop file: %s) as per config", w.blockDevice, w.loopPath)
		handle.Finalize(0)
		return
	}

	fmt.Fprintf(handle, "Removing loop device %s\n", w.loopPath)

	cmd := exec.CommandContext(ctx, "losetup", "--detach", w.blockDevice)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(handle, "Command 'losetup --detach' failed: %q", out)
	}

	if err := os.Remove(w.loopPath); err != nil {
		fmt.Fprintf(handle, "Error removing loop device: %v", err)
	}

	// We always finish this with exit code 0 even if it errored, because workspace
	// cleanup doesn't fail the execution job. We can deal with it separately.
	handle.Finalize(0)
}

type dockerWorkspace struct {
	path            string
	scriptFilenames []string
	keepWorkspaces  bool
	workspaceDir    string
	commandLogger   command.Logger
}

func (w dockerWorkspace) Path() string {
	return w.path
}

func (w dockerWorkspace) ScriptFilenames() []string {
	return w.scriptFilenames
}

func (w dockerWorkspace) Remove(ctx context.Context) {
	handle := w.commandLogger.Log("teardown.fs", nil)
	defer handle.Close()

	if w.keepWorkspaces {
		fmt.Fprintf(handle, "Preserving workspace (%s) as per config", w.workspaceDir)
	}

	fmt.Fprintf(handle, "Removing %s\n", w.workspaceDir)
	if rmErr := os.RemoveAll(w.workspaceDir); rmErr != nil {
		fmt.Fprintf(handle, "Operation failed: %s\n", rmErr.Error())
	}

	// We always finish this with exit code 0 even if it errored, because workspace
	// cleanup doesn't fail the execution job. We can deal with it separately.
	handle.Finalize(0)
}

// prepareWorkspace creates and returns a temporary directory in which acts the workspace
// while processing a single job. It is up to the caller to ensure that this directory is
// removed after the job has finished processing. If a repository name is supplied, then
// that repository will be cloned (through the frontend API) into the workspace.
func (h *handler) prepareWorkspace(
	ctx context.Context,
	commandRunner command.Runner,
	job executor.Job,
	commandLogger command.Logger,
) (workspace Workspace, err error) {
	if h.options.FirecrackerOptions.Enabled {
		loopFileName, tmpMountDir, blockDevice, err := setupLoopDevice(
			ctx,
			job.ID,
			h.options.ResourceOptions.DiskSpace,
			h.options.KeepWorkspaces,
			commandLogger,
		)
		if err != nil {
			return nil, err
		}
		defer func() {
			if !h.options.KeepWorkspaces {
				syscall.Unmount(tmpMountDir, 0)
				os.RemoveAll(tmpMountDir)
			}
		}()

		scriptPaths, err := h.prepareCloneAndScripts(ctx, job, tmpMountDir, commandRunner, commandLogger)
		if err != nil {
			return nil, err
		}

		return &firecrackerWorkspace{
			path:            blockDevice,
			scriptFilenames: scriptPaths,
			keepWorkspaces:  h.options.KeepWorkspaces,
			loopPath:        loopFileName,
			blockDevice:     blockDevice,
			commandLogger:   commandLogger,
		}, err
	}

	workspaceDir, err := makeTempDirectory("workspace-" + strconv.Itoa(job.ID))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(workspaceDir)
		}
	}()

	scriptPaths, err := h.prepareCloneAndScripts(ctx, job, workspaceDir, commandRunner, commandLogger)
	if err != nil {
		return nil, err
	}

	return &dockerWorkspace{
		path:            workspaceDir,
		scriptFilenames: scriptPaths,
		keepWorkspaces:  h.options.KeepWorkspaces,
		workspaceDir:    workspaceDir,
		commandLogger:   commandLogger,
	}, nil
}

func (h *handler) prepareCloneAndScripts(
	ctx context.Context,
	job executor.Job,
	workspaceDir string,
	commandRunner command.Runner,
	commandLogger command.Logger,
) ([]string, error) {
	if job.RepositoryName != "" {
		repoPath := workspaceDir
		if job.RepositoryDirectory != "" {
			repoPath = filepath.Join(workspaceDir, job.RepositoryDirectory)

			if !strings.HasPrefix(repoPath, workspaceDir) {
				return nil, errors.Newf("invalid repo path %q not a subdirectory of %q", repoPath, workspaceDir)
			}

			if err := os.MkdirAll(repoPath, os.ModePerm); err != nil {
				return nil, errors.Wrap(err, "creating repo directory")
			}
		}

		cloneURL, err := makeRelativeURL(
			h.options.ClientOptions.EndpointOptions.URL,
			h.options.GitServicePath,
			job.RepositoryName,
		)
		if err != nil {
			return nil, err
		}

		authorizationOption := fmt.Sprintf(
			"http.extraHeader=Authorization: %s %s",
			SchemeExecutorToken,
			h.options.ClientOptions.EndpointOptions.Token,
		)

		fetchCommand := []string{
			"git",
			"-C", repoPath,
			"-c", "protocol.version=2",
			"-c", authorizationOption,
			"-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal",
			"fetch",
			"--progress",
			"--no-recurse-submodules",
			"origin",
			job.Commit,
		}

		appendFetchArg := func(arg string) {
			l := len(fetchCommand)
			insertPos := l - 2
			fetchCommand = append(fetchCommand[:insertPos+1], fetchCommand[insertPos:]...)
			fetchCommand[insertPos] = arg
		}

		if job.FetchTags {
			appendFetchArg("--tags")
		}

		if job.ShallowClone {
			if !job.FetchTags {
				appendFetchArg("--no-tags")
			}
			appendFetchArg("--depth=1")
		}

		// For a sparse checkout, we want to add a blob filter so we only fetch the minimum set of files initially.
		if len(job.SparseCheckout) > 0 {
			appendFetchArg("--filter=blob:none")
		}

		gitCommands := []command.CommandSpec{
			{Key: "setup.git.init", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "init"}, Operation: h.operations.SetupGitInit},
			{Key: "setup.git.add-remote", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "remote", "add", "origin", cloneURL.String()}, Operation: h.operations.SetupAddRemote},
			// Disable gc, this can improve performance and should never run for executor clones.
			{Key: "setup.git.disable-gc", Env: gitStdEnv, Command: []string{"git", "-C", repoPath, "config", "--local", "gc.auto", "0"}, Operation: h.operations.SetupGitDisableGC},
			{Key: "setup.git.fetch", Env: gitStdEnv, Command: fetchCommand, Operation: h.operations.SetupGitFetch},
		}

		if len(job.SparseCheckout) > 0 {
			gitCommands = append(gitCommands, command.CommandSpec{
				Key:       "setup.git.sparse-checkout-config",
				Env:       gitStdEnv,
				Command:   []string{"git", "-C", repoPath, "config", "--local", "core.sparseCheckout", "1"},
				Operation: h.operations.SetupGitSparseCheckoutConfig,
			})
			gitCommands = append(gitCommands, command.CommandSpec{
				Key:       "setup.git.sparse-checkout-set",
				Env:       gitStdEnv,
				Command:   append([]string{"git", "-C", repoPath, "sparse-checkout", "set", "--no-cone", "--"}, job.SparseCheckout...),
				Operation: h.operations.SetupGitSparseCheckoutSet,
			})
		}

		checkoutCommand := []string{
			"git",
			"-C", repoPath,
			"checkout",
			"--progress",
			"--force",
			job.Commit,
		}

		// Sparse checkouts need to fetch additional blobs, so we need to add
		// auth config here.
		if len(job.SparseCheckout) > 0 {
			checkoutCommand = []string{
				"git",
				"-C", repoPath,
				"-c", "protocol.version=2", "-c", authorizationOption, "-c", "http.extraHeader=X-Sourcegraph-Actor-UID: internal",
				"checkout",
				"--progress",
				"--force",
				job.Commit,
			}
		}

		gitCommands = append(gitCommands, command.CommandSpec{
			Key:       "setup.git.checkout",
			Env:       gitStdEnv,
			Command:   checkoutCommand,
			Operation: h.operations.SetupGitCheckout,
		})

		// This is for LSIF, it relies on the origin being set to the upstream repo
		// for indexing.
		gitCommands = append(gitCommands, command.CommandSpec{
			Key: "setup.git.set-remote",
			Env: gitStdEnv,
			Command: []string{
				"git",
				"-C", repoPath,
				"remote",
				"set-url",
				"origin",
				job.RepositoryName,
			},
			Operation: h.operations.SetupGitSetRemoteUrl,
		})

		for _, spec := range gitCommands {
			if err := commandRunner.Run(ctx, spec); err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("failed %s", spec.Key))
			}
		}
	}

	// Create the scripts path.
	if err := os.MkdirAll(filepath.Join(workspaceDir, command.ScriptsPath), os.ModePerm); err != nil {
		return nil, errors.Wrap(err, "creating script path")
	}

	// Construct a map from filenames to file content that should be accessible to jobs
	// within the workspace. This consists of files supplied within the job record itself,
	// as well as file-version of each script step.
	workspaceFileContentsByPath := map[string][]byte{}

	for relativePath, content := range job.VirtualMachineFiles {
		path, err := filepath.Abs(filepath.Join(workspaceDir, relativePath))
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(path, workspaceDir) {
			return nil, errors.Errorf("refusing to write outside of working directory")
		}

		workspaceFileContentsByPath[path] = []byte(content)
	}

	scriptNames := make([]string, 0, len(job.DockerSteps))
	for i, dockerStep := range job.DockerSteps {
		scriptName := scriptNameFromJobStep(job, i)
		scriptNames = append(scriptNames, scriptName)

		path := filepath.Join(workspaceDir, command.ScriptsPath, scriptName)
		workspaceFileContentsByPath[path] = buildScript(dockerStep)
	}

	if err := writeFiles(workspaceFileContentsByPath, commandLogger); err != nil {
		return nil, errors.Wrap(err, "failed to write virtual machine files")
	}

	return scriptNames, nil
}

func makeRelativeURL(base string, path ...string) (*url.URL, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, err
	}

	urlx, err := baseURL.ResolveReference(&url.URL{Path: filepath.Join(path...)}), nil
	if err != nil {
		return nil, err
	}

	urlx.User = url.User("executor")
	return urlx, nil
}

// makeTempFile defaults to makeTemporaryFile and can be replaced for testing
// with determinstic workspace/scripts directories.
var makeTempFile = makeTemporaryFile

func makeTemporaryFile(prefix string) (*os.File, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return nil, err
		}
		return os.CreateTemp(tempdir, prefix+"-*")
	}

	return os.CreateTemp("", prefix+"-*")
}

// makeTempDirectory defaults to makeTemporaryDirectory and can be replaced for testing
// with determinstic workspace/scripts directories.
var makeTempDirectory = makeTemporaryDirectory

func makeTemporaryDirectory(prefix string) (string, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return "", err
		}
		return os.MkdirTemp(tempdir, prefix+"-*")
	}

	return os.MkdirTemp("", prefix+"-*")
}

var scriptPreamble = `
set -x
`

func buildScript(dockerStep executor.DockerStep) []byte {
	return []byte(strings.Join(append([]string{scriptPreamble, ""}, dockerStep.Commands...), "\n") + "\n")
}

func scriptNameFromJobStep(job executor.Job, i int) string {
	return fmt.Sprintf("%d.%d_%s@%s.sh", job.ID, i, strings.ReplaceAll(job.RepositoryName, "/", "_"), job.Commit)
}

// writeFiles writes to the filesystem the content in the given map.
func writeFiles(workspaceFileContentsByPath map[string][]byte, logger command.Logger) (err error) {
	// Bail out early if nothing to do, we don't need to spawn an empty log group.
	if len(workspaceFileContentsByPath) == 0 {
		return nil
	}

	handle := logger.Log("setup.fs.extras", nil)
	defer func() {
		if err == nil {
			handle.Finalize(0)
		} else {
			handle.Finalize(1)
		}

		handle.Close()
	}()

	for path, content := range workspaceFileContentsByPath {
		// Ensure the path exists.
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(path, content, os.ModePerm); err != nil {
			return err
		}

		fmt.Fprintf(handle, "Wrote %s\n", path)
	}

	return nil
}

func setupLoopDevice(
	ctx context.Context,
	jobID int,
	diskSpace string,
	keepWorkspaces bool,
	commandLogger command.Logger,
) (loopFileName, tmpMountDir, blockDevice string, err error) {
	handle := commandLogger.Log("setup.fs.workspace", nil)
	defer func() {
		if err != nil {
			handle.Finalize(1)
		} else {
			handle.Finalize(0)
		}
		handle.Close()
	}()

	tempFile, err := makeTempFile("workspace-loop-" + strconv.Itoa(jobID))
	if err != nil {
		return "", "", "", err
	}
	loopFileName = tempFile.Name()

	fmt.Fprintf(handle, "Created backing workspace file at %q\n", loopFileName)

	defer func() {
		if !keepWorkspaces {
			os.Remove(tempFile.Name())
		}
	}()

	diskSize, err := datasize.ParseString(diskSpace)
	if err != nil {
		return "", "", "", errors.Wrapf(err, "invalid disk size provided: %q", diskSpace)
	}

	if err := tempFile.Truncate(int64(diskSize.Bytes())); err != nil {
		return "", "", "", errors.Wrapf(err, "failed to make backing file sparse with %d bytes", diskSize.Bytes())
	}

	fmt.Fprintf(handle, "Created sparse file of size %s from %q\n", diskSize.HumanReadable(), loopFileName)

	tempFile.Close()

	cmd := exec.CommandContext(ctx, "mkfs.ext4", loopFileName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", "", errors.Newf("failed to create ext4 filesystem in backing file: %q", out)
	}

	mkfsOutput := "stderr: " + strings.ReplaceAll(strings.TrimSpace(string(out)), "\n", "\nstderr: ")
	fmt.Fprintf(handle, "Wrote ext4 filesystem to backing file %q:\n%s\n", loopFileName, mkfsOutput)

	tmpMountDir, err = makeTempDirectory("workspace-mount-" + strconv.Itoa(jobID))
	if err != nil {
		return "", "", "", err
	}

	fmt.Fprintf(handle, "Created temporary workspace mount location at %q\n", tmpMountDir)

	cmd = exec.CommandContext(ctx, "losetup", "--find", "--show", loopFileName)
	out, err = cmd.CombinedOutput()
	if err != nil {
		return "", "", "", errors.Newf("failed to create loop device: %q", out)
	}

	blockDevice = strings.TrimSpace(string(out))

	fmt.Fprintf(handle, "Created loop device at %q backed by %q\n", blockDevice, loopFileName)

	mountCmd := exec.CommandContext(ctx, "mount", blockDevice, tmpMountDir)
	if out, err := mountCmd.CombinedOutput(); err != nil {
		if !keepWorkspaces {
			os.RemoveAll(tmpMountDir)
		}
		return "", "", "", errors.Newf("failed to mount loop device %q to %q: %q", loopFileName, tmpMountDir, out)
	}

	return loopFileName, tmpMountDir, blockDevice, nil
}
