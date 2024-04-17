package main

import (
	"context"
	"flag"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/batches/ui"
	"github.com/sourcegraph/src-cli/internal/batches/watchdog"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

// We check for docker responsiveness every minute
const dockerWatchDuration = 1 * time.Minute

// batchExecutionFlags are common to batch changes that are executed both
// locally and remotely.
type batchExecutionFlags struct {
	allowUnsupported bool
	allowIgnored     bool
	api              *api.Flags
	clearCache       bool
	namespace        string
}

func newBatchExecutionFlags(flagSet *flag.FlagSet) *batchExecutionFlags {
	bef := &batchExecutionFlags{
		api: api.NewFlags(flagSet),
	}

	flagSet.BoolVar(
		&bef.allowUnsupported, "allow-unsupported", false,
		"Allow unsupported code hosts.",
	)
	flagSet.BoolVar(
		&bef.clearCache, "clear-cache", false,
		"If true, clears the execution cache and executes all steps anew.",
	)
	flagSet.BoolVar(
		&bef.allowIgnored, "force-override-ignore", false,
		"Do not ignore repositories that have a .batchignore file.",
	)
	flagSet.StringVar(
		&bef.namespace, "namespace", "",
		"The user or organization namespace to place the batch change within. Default is the currently authenticated user.",
	)
	flagSet.StringVar(&bef.namespace, "n", "", "Alias for -namespace.")

	return bef
}

// batchExecuteFlags are used when executing batch changes locally.
type batchExecuteFlags struct {
	*batchExecutionFlags

	apply         bool
	cacheDir      string
	tempDir       string
	file          string
	keepLogs      bool
	parallelism   int
	timeout       time.Duration
	workspace     string
	cleanArchives bool
	skipErrors    bool
	runAsRoot     bool

	// EXPERIMENTAL
	textOnly bool
}

func newBatchExecuteFlags(flagSet *flag.FlagSet, cacheDir, tempDir string) *batchExecuteFlags {
	caf := &batchExecuteFlags{
		batchExecutionFlags: newBatchExecutionFlags(flagSet),
	}

	flagSet.BoolVar(
		&caf.textOnly, "text-only", false,
		"INTERNAL USE ONLY. EXPERIMENTAL. Switches off the TUI to only print JSON lines.",
	)

	flagSet.BoolVar(
		&caf.apply, "apply", false,
		"Ignored.",
	)

	flagSet.BoolVar(
		&caf.keepLogs, "keep-logs", false,
		"Retain logs after executing steps.",
	)

	flagSet.StringVar(
		&caf.cacheDir, "cache", cacheDir,
		"Directory for caching results and repository archives.",
	)

	flagSet.StringVar(
		&caf.tempDir, "tmp", tempDir,
		"Directory for storing temporary data, such as log files. Default is /tmp. Can also be set with environment variable SRC_BATCH_TMP_DIR; if both are set, this flag will be used and not the environment variable.",
	)

	flagSet.StringVar(
		&caf.file, "f", "",
		"The batch spec file to read, or - to read from standard input.",
	)

	flagSet.IntVar(
		&caf.parallelism, "j", 0,
		"The maximum number of parallel jobs. Default (or 0) is the number of CPU cores available to Docker.",
	)

	flagSet.DurationVar(
		&caf.timeout, "timeout", 60*time.Minute,
		"The maximum duration a single batch spec step can take.",
	)

	flagSet.BoolVar(
		&caf.cleanArchives, "clean-archives", true,
		"If true, deletes downloaded repository archives after executing batch spec steps. Note that only the archives related to the actual repositories matched by the batch spec will be cleaned up, and clean up will not occur if src exits unexpectedly.",
	)

	flagSet.BoolVar(
		&caf.skipErrors, "skip-errors", false,
		"If true, errors encountered while executing steps in a repository won't stop the execution of the batch spec but only cause that repository to be skipped.",
	)

	flagSet.StringVar(
		&caf.workspace, "workspace", "auto",
		`Workspace mode to use ("auto", "bind", or "volume")`,
	)

	flagSet.BoolVar(verbose, "v", false, "print verbose output")

	flagSet.BoolVar(
		&caf.runAsRoot, "run-as-root", false,
		"If true, forces all step containers to run as root.",
	)

	return caf
}

var errAdditionalArguments = cmderrors.Usage("additional arguments not allowed")

func getBatchSpecFile(flagSet *flag.FlagSet, fileFlag *string) (string, error) {
	if fileFlag == nil || *fileFlag != "" {
		if flagSet.NArg() != 0 {
			return "", errAdditionalArguments
		}
		if fileFlag == nil {
			return "", nil
		}
		return *fileFlag, nil
	} else if flagSet.NArg() > 1 {
		return "", errAdditionalArguments
	}
	return flagSet.Arg(0), nil
}

func batchDefaultCacheDir() string {
	uc, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	dir := path.Join(uc, "sourcegraph", "batch")

	return dir
}

// batchDefaultTempDirPrefix returns the prefix to be passed to ioutil.TempFile.
// If the environment variable SRC_BATCH_TMP_DIR is set, that is used as the prefix.
// Otherwise we use "/tmp".
func batchDefaultTempDirPrefix() string {
	if p := os.Getenv("SRC_BATCH_TMP_DIR"); p != "" {
		return p
	}

	// On macOS, we use an explicit prefix for our temp directories, because
	// otherwise Go would use $TMPDIR, which is set to `/var/folders` per
	// default on macOS. But Docker for Mac doesn't have `/var/folders` in its
	// default set of shared folders, but it does have `/tmp` in there.
	if runtime.GOOS == "darwin" {
		return "/tmp"
	}

	return os.TempDir()
}

func batchOpenFileFlag(flag string) (*os.File, error) {
	if flag == "" || flag == "-" {
		if flag != "-" {
			// If the flag wasn't set, we want to check stdin. If it's not a TTY,
			// then we'll assume that we always want to read from it. If it is a TTY,
			// then we'll briefly pause to see if data is getting piped in, otherwise
			// we'll error out, because it's likely that the user forgot the `-f` on
			// the command line.
			fd := os.Stdin.Fd()
			if isatty.IsTerminal(fd) {
				has, err := ui.HasInput(os.Stdin.Fd(), 250*time.Millisecond)
				if err != nil {
					return nil, errors.Wrap(err, "checking for input on stdin")
				} else if !has {
					return nil, errors.New("-f specified, but no input was detected on stdin; did you forget to pipe a batch spec in?")
				}
			}
		}
		// https://github.com/golang/go/issues/24842
		if err := syscall.SetNonblock(0, true); err != nil {
			panic(err)
		}
		stdin := os.NewFile(0, "stdin")

		return stdin, nil
	}

	file, err := os.Open(flag)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file %s", flag)
	}
	return file, nil
}

type executeBatchSpecOpts struct {
	flags *batchExecuteFlags

	applyBatchSpec bool
	file           string

	client api.Client
}

func createDockerWatchdog(ctx context.Context, execUI ui.ExecUI) *watchdog.WatchDog {
	return watchdog.New(dockerWatchDuration, func() {
		_, err := docker.NCPU(ctx)
		if err != nil {
			execUI.DockerWatchDogWarning(errors.Wrap(err, "docker watchdog"))
		}
	})
}

// executeBatchSpec performs all the steps required to upload the batch spec to
// Sourcegraph, including execution as needed and applying the resulting batch
// spec if specified.
func executeBatchSpec(ctx context.Context, opts executeBatchSpecOpts) (err error) {
	var execUI ui.ExecUI
	if opts.flags.textOnly {
		execUI = &ui.JSONLines{}
	} else {
		out := output.NewOutput(os.Stderr, output.OutputOpts{Verbose: *verbose})
		execUI = &ui.TUI{Out: out}
	}

	w := createDockerWatchdog(ctx, execUI)
	go w.Start()

	defer func() {
		w.Stop()
		if err != nil {
			execUI.ExecutionError(err)
		}
	}()

	svc := service.New(&service.Opts{
		Client: opts.client,
	})

	lr, ffs, err := svc.DetermineLicenseAndFeatureFlags(ctx)
	if err != nil {
		return err
	}

	// Once we know about feature flags, reconfigure the UI if needed.
	if opts.flags.textOnly && ffs.BinaryDiffs {
		execUI = &ui.JSONLines{BinaryDiffs: true}
	}

	imageCache := docker.NewImageCache()

	if err := validateSourcegraphVersionConstraint(ctx, ffs); err != nil {
		return err
	}

	if err := checkExecutable("git", "version"); err != nil {
		return err
	}

	// In the past, we relied on `getBatchParallelism` to ascertain if docker is running,
	// however, we don't always check for the number of CPUs (especially when the -j parallelis)
	// flag is passed. This is a more explicit check to confirm docker is working.
	if err := docker.CheckVersion(ctx); err != nil {
		return err
	}

	parallelism, err := getBatchParallelism(ctx, opts.flags.parallelism)
	if err != nil {
		return err
	}

	// On Linux only, we also need to figure out if we need to override the
	// temporary directory — Docker Desktop restricts file mounts to /home only
	// by default.
	//
	// We could interrogate ~/.docker/desktop/settings.json for extra bonus
	// points here, but that feels like overkill. Basically, if it's
	// desktop-linux, we'll just assume the user has the default /home mount
	// available and go from there.
	if runtime.GOOS == "linux" && opts.flags.tempDir == batchDefaultTempDirPrefix() {
		context, err := docker.CurrentContext(ctx)
		if err != nil {
			return err
		}

		if context == "desktop-linux" {
			opts.flags.tempDir = path.Join(path.Dir(opts.flags.cacheDir), "batch-tmp")

			// Ensure the directory exists and is writable.
			os.MkdirAll(opts.flags.tempDir, os.ModePerm)
		}
	}

	// Parse flags and build up our service and executor options.
	execUI.ParsingBatchSpec()
	batchSpec, batchSpecDir, rawSpec, err := parseBatchSpec(ctx, opts.file, svc)
	if err != nil {
		var multiErr errors.MultiError
		if errors.As(err, &multiErr) {
			execUI.ParsingBatchSpecFailure(multiErr)
			return cmderrors.ExitCode(2, nil)
		} else {
			// This shouldn't happen; let's just punt and let the normal
			// rendering occur.
			return err
		}
	}
	execUI.ParsingBatchSpecSuccess()

	execUI.ResolvingNamespace()
	namespace, err := svc.ResolveNamespace(ctx, opts.flags.namespace)
	if err != nil {
		return err
	}
	execUI.ResolvingNamespaceSuccess(namespace.ID)

	var workspaceCreator workspace.Creator

	if len(batchSpec.Steps) > 0 {
		execUI.PreparingContainerImages()
		images, err := svc.EnsureDockerImages(
			ctx,
			imageCache,
			batchSpec.Steps,
			parallelism,
			execUI.PreparingContainerImagesProgress,
		)
		if err != nil {
			return err
		}
		execUI.PreparingContainerImagesSuccess()

		execUI.DeterminingWorkspaceCreatorType()
		var typ workspace.CreatorType
		workspaceCreator, typ = workspace.NewCreator(ctx, opts.flags.workspace, opts.flags.cacheDir, opts.flags.tempDir, images)
		if typ == workspace.CreatorTypeVolume {
			// This creator type requires an additional image, so let's ensure it exists.
			_, err = imageCache.Ensure(ctx, workspace.DockerVolumeWorkspaceImage)
			if err != nil {
				return err
			}
		}
		execUI.DeterminingWorkspaceCreatorTypeSuccess(typ)
	}

	execUI.DeterminingWorkspaces()
	workspaces, repos, err := svc.ResolveWorkspacesForBatchSpec(ctx, batchSpec, opts.flags.allowUnsupported, opts.flags.allowIgnored)
	if err != nil {
		if repoSet, ok := err.(batches.UnsupportedRepoSet); ok {
			execUI.DeterminingWorkspacesSuccess(len(workspaces), len(repos), repoSet, nil)
		} else if repoSet, ok := err.(batches.IgnoredRepoSet); ok {
			execUI.DeterminingWorkspacesSuccess(len(workspaces), len(repos), nil, repoSet)
		} else {
			return errors.Wrap(err, "resolving repositories")
		}
	} else {
		execUI.DeterminingWorkspacesSuccess(len(workspaces), len(repos), nil, nil)
	}

	archiveRegistry := repozip.NewArchiveRegistry(opts.client, opts.flags.cacheDir, opts.flags.cleanArchives)
	logManager := log.NewDiskManager(opts.flags.tempDir, opts.flags.keepLogs)
	coord := executor.NewCoordinator(
		executor.NewCoordinatorOpts{
			ExecOpts: executor.NewExecutorOpts{
				Logger:              logManager,
				RepoArchiveRegistry: archiveRegistry,
				Creator:             workspaceCreator,
				EnsureImage:         imageCache.Ensure,
				Parallelism:         parallelism,
				WorkingDirectory:    batchSpecDir,
				Timeout:             opts.flags.timeout,
				TempDir:             opts.flags.tempDir,
				GlobalEnv:           os.Environ(),
				ForceRoot:           opts.flags.runAsRoot,
				BinaryDiffs:         ffs.BinaryDiffs,
			},
			Logger:      logManager,
			Cache:       executor.NewDiskCache(opts.flags.cacheDir),
			BinaryDiffs: ffs.BinaryDiffs,
			GlobalEnv:   os.Environ(),
		},
	)

	execUI.CheckingCache()
	tasks := svc.BuildTasks(
		&template.BatchChangeAttributes{
			Name:        batchSpec.Name,
			Description: batchSpec.Description,
		},
		batchSpec.Steps,
		workspaces,
	)
	var (
		specs         []*batcheslib.ChangesetSpec
		uncachedTasks []*executor.Task
	)
	if opts.flags.clearCache {
		if err := coord.ClearCache(ctx, tasks); err != nil {
			return err
		}
		uncachedTasks = tasks
	} else {
		// Check the cache for completely cached executions.
		uncachedTasks, specs, err = coord.CheckCache(ctx, batchSpec, tasks)
		if err != nil {
			return err
		}
	}
	execUI.CheckingCacheSuccess(len(specs), len(uncachedTasks))

	taskExecUI := execUI.ExecutingTasks(*verbose, parallelism)
	freshSpecs, logFiles, execErr := coord.ExecuteAndBuildSpecs(ctx, batchSpec, uncachedTasks, taskExecUI)
	// Add external changeset specs.
	importedSpecs, importErr := svc.CreateImportChangesetSpecs(ctx, batchSpec)
	if execErr != nil {
		err = errors.Append(err, execErr)
	}
	if importErr != nil {
		err = errors.Append(err, importErr)
	}
	if err != nil && !opts.flags.skipErrors {
		return err
	}
	if err == nil || opts.flags.skipErrors {
		if err == nil {
			taskExecUI.Success()
		} else {
			execUI.ExecutingTasksSkippingErrors(err)
		}
	} else {
		if err != nil {
			taskExecUI.Failed(err)
			return err
		}
	}

	if len(logFiles) > 0 && opts.flags.keepLogs {
		execUI.LogFilesKept(logFiles)
	}

	specs = append(specs, freshSpecs...)
	specs = append(specs, importedSpecs...)

	err = svc.ValidateChangesetSpecs(repos, specs)
	if err != nil {
		return err
	}

	ids := make([]graphql.ChangesetSpecID, len(specs))

	if len(specs) > 0 {
		execUI.UploadingChangesetSpecs(len(specs))

		for i, spec := range specs {
			id, err := svc.CreateChangesetSpec(ctx, spec)
			if err != nil {
				return err
			}
			ids[i] = id
			execUI.UploadingChangesetSpecsProgress(i+1, len(specs))
		}

		execUI.UploadingChangesetSpecsSuccess(ids)
	} else if len(repos) == 0 {
		execUI.NoChangesetSpecs()
	}

	execUI.CreatingBatchSpec()
	id, url, err := svc.CreateBatchSpec(ctx, namespace.ID, rawSpec, ids)
	if err != nil {
		return execUI.CreatingBatchSpecError(lr.MaxUnlicensedChangesets, err)
	}
	previewURL := cfg.Endpoint + url
	execUI.CreatingBatchSpecSuccess(previewURL)

	hasWorkspaceFiles := false
	for _, step := range batchSpec.Steps {
		if len(step.Mount) > 0 {
			hasWorkspaceFiles = true
			break
		}
	}
	if hasWorkspaceFiles {
		execUI.UploadingWorkspaceFiles()
		if err := svc.UploadBatchSpecWorkspaceFiles(ctx, batchSpecDir, string(id), batchSpec.Steps); err != nil {
			// Since failing to upload workspace files should not stop processing, just warn
			execUI.UploadingWorkspaceFilesWarning(errors.Wrap(err, "uploading workspace files"))
		} else {
			execUI.UploadingWorkspaceFilesSuccess()
		}
	}

	if !opts.applyBatchSpec {
		execUI.PreviewBatchSpec(previewURL)
		return
	}

	execUI.ApplyingBatchSpec()
	batch, err := svc.ApplyBatchChange(ctx, id)
	if err != nil {
		return err
	}
	execUI.ApplyingBatchSpecSuccess(cfg.Endpoint + batch.URL)

	return nil
}

func setReadDeadlineOnCancel(ctx context.Context, f *os.File) {
	go func() {
		// When user cancels, we set the read deadline to now() so the runtime
		// cancels the read and we don't block.
		<-ctx.Done()
		f.SetReadDeadline(time.Now())
	}()
}

// parseBatchSpec parses and validates the given batch spec. If the spec has
// validation errors, they are returned.
func parseBatchSpec(ctx context.Context, file string, svc *service.Service) (*batcheslib.BatchSpec, string, string, error) {
	f, err := batchOpenFileFlag(file)
	if err != nil {
		return nil, "", "", err
	}
	defer f.Close()

	// Create new ctx so we ensure that the goroutine in
	// setReadDeadlineOnCancel exits at end of function.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	setReadDeadlineOnCancel(ctx, f)

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, "", "", errors.Wrap(err, "reading batch spec")
	}

	dir, err := getBatchSpecDirectory(file)
	if err != nil {
		return nil, "", "", errors.Wrap(err, "batch spec path")
	}

	spec, err := svc.ParseBatchSpec(dir, data)
	return spec, dir, string(data), err
}

func getBatchSpecDirectory(file string) (string, error) {
	var workingDirectory string
	var err error
	// if the batch spec is being provided via standard input, set the working directory to the current directory
	if file == "" || file == "-" {
		workingDirectory, err = os.Getwd()
		if err != nil {
			return "", errors.Wrap(err, "batch spec path")
		}
	} else {
		p, err := filepath.Abs(file)
		if err != nil {
			return "", errors.Wrap(err, "batch spec path")
		}
		workingDirectory = filepath.Dir(p)
	}
	return workingDirectory, nil
}

func checkExecutable(cmd string, args ...string) error {
	if err := exec.Command(cmd, args...).Run(); err != nil {
		return errors.Newf(
			"failed to execute \"%s %s\":\n\t%s\n\n'src batch' requires %q to be available.",
			cmd,
			strings.Join(args, " "),
			err,
			cmd,
		)
	}
	return nil
}

func contextCancelOnInterrupt(parent context.Context) (context.Context, func()) {
	ctx, ctxCancel := context.WithCancel(parent)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case <-c:
			ctxCancel()
		case <-ctx.Done():
		}
	}()

	return ctx, func() {
		signal.Stop(c)
		ctxCancel()
	}
}

func getBatchParallelism(ctx context.Context, flag int) (int, error) {
	if flag > 0 {
		return flag, nil
	}

	return docker.NCPU(ctx)
}

func validateSourcegraphVersionConstraint(ctx context.Context, ffs *batches.FeatureFlags) error {
	if ffs.Sourcegraph40 {
		return nil
	}
	return errors.Newf("\n\n * Warning:\n This version of src-cli requires Sourcegraph version 4.0 or newer. If you're not on Sourcegraph 4.0 or newer, please use the 3.x release of src-cli that corresponds to your Sourcegraph version.\n\n")
}
