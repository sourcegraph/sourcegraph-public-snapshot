package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/batches/ui"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

type batchExecuteFlags struct {
	allowUnsupported bool
	allowIgnored     bool
	api              *api.Flags
	apply            bool
	cacheDir         string
	tempDir          string
	clearCache       bool
	file             string
	keepLogs         bool
	namespace        string
	parallelism      int
	timeout          time.Duration
	workspace        string
	cleanArchives    bool
	skipErrors       bool

	// EXPERIMENTAL
	textOnly bool
}

func newBatchExecuteFlags(flagSet *flag.FlagSet, workspaceExecution bool, cacheDir, tempDir string) *batchExecuteFlags {
	caf := &batchExecuteFlags{
		api: api.NewFlags(flagSet),
	}

	if !workspaceExecution {
		flagSet.BoolVar(
			&caf.textOnly, "text-only", false,
			"INTERNAL USE ONLY. EXPERIMENTAL. Switches off the TUI to only print JSON lines.",
		)
		flagSet.BoolVar(
			&caf.allowUnsupported, "allow-unsupported", false,
			"Allow unsupported code hosts.",
		)
		flagSet.BoolVar(
			&caf.allowIgnored, "force-override-ignore", false,
			"Do not ignore repositories that have a .batchignore file.",
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
			&caf.namespace, "namespace", "",
			"The user or organization namespace to place the batch change within. Default is the currently authenticated user.",
		)
		flagSet.StringVar(&caf.namespace, "n", "", "Alias for -namespace.")
	}

	flagSet.StringVar(
		&caf.cacheDir, "cache", cacheDir,
		"Directory for caching results and repository archives.",
	)
	flagSet.BoolVar(
		&caf.clearCache, "clear-cache", false,
		"If true, clears the execution cache and executes all steps anew.",
	)
	flagSet.StringVar(
		&caf.tempDir, "tmp", tempDir,
		"Directory for storing temporary data, such as log files. Default is /tmp. Can also be set with environment variable SRC_BATCH_TMP_DIR; if both are set, this flag will be used and not the environment variable.",
	)

	flagSet.StringVar(
		&caf.file, "f", "",
		"The batch spec file to read.",
	)

	flagSet.IntVar(
		&caf.parallelism, "j", runtime.GOMAXPROCS(0),
		"The maximum number of parallel jobs. Default is GOMAXPROCS.",
	)
	flagSet.DurationVar(
		&caf.timeout, "timeout", 60*time.Minute,
		"The maximum duration a single batch spec step can take.",
	)
	flagSet.BoolVar(
		&caf.cleanArchives, "clean-archives", true,
		"If true, deletes downloaded repository archives after executing batch spec steps.",
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

	return caf
}

func batchDefaultCacheDir() string {
	uc, err := os.UserCacheDir()
	if err != nil {
		return ""
	}

	// Check if there's an old campaigns cache directory but not a new batch
	// directory: if so, we should rename the old directory and carry on.
	//
	// TODO(campaigns-deprecation): we can remove this migration shim after June
	// 2021.
	old := path.Join(uc, "sourcegraph", "campaigns")
	dir := path.Join(uc, "sourcegraph", "batch")
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if _, err := os.Stat(old); os.IsExist(err) {
			// We'll just try to do this without checking for an error: if it
			// fails, we'll carry on and let the normal cache directory handling
			// logic take care of it.
			os.Rename(old, dir)
		}
	}

	return dir
}

// batchDefaultTempDirPrefix returns the prefix to be passed to ioutil.TempFile.
// If one of the environment variables SRC_BATCH_TMP_DIR or
// SRC_CAMPAIGNS_TMP_DIR is set, that is used as the prefix. Otherwise we use
// "/tmp".
func batchDefaultTempDirPrefix() string {
	// TODO(campaigns-deprecation): we can remove this migration shim in
	// Sourcegraph 4.0.
	for _, env := range []string{"SRC_BATCH_TMP_DIR", "SRC_CAMPAIGNS_TMP_DIR"} {
		if p := os.Getenv(env); p != "" {
			return p
		}
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

func batchOpenFileFlag(flag *string) (io.ReadCloser, error) {
	if flag == nil || *flag == "" || *flag == "-" {
		return os.Stdin, nil
	}

	file, err := os.Open(*flag)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file %q", *flag)
	}
	return file, nil
}

type executeBatchSpecOpts struct {
	flags *batchExecuteFlags

	applyBatchSpec bool

	client api.Client
}

// executeBatchSpec performs all the steps required to upload the batch spec to
// Sourcegraph, including execution as needed and applying the resulting batch
// spec if specified.
func executeBatchSpec(ctx context.Context, ui ui.ExecUI, opts executeBatchSpecOpts) (err error) {
	defer func() {
		if err != nil {
			ui.ExecutionError(err)
		}
	}()

	svc := service.New(&service.Opts{
		AllowUnsupported: opts.flags.allowUnsupported,
		AllowIgnored:     opts.flags.allowIgnored,
		AllowFiles:       true,
		Client:           opts.client,
	})

	if err := svc.DetermineFeatureFlags(ctx); err != nil {
		return err
	}

	if err := checkExecutable("git", "version"); err != nil {
		return err
	}

	if err := checkExecutable("docker", "version"); err != nil {
		return err
	}

	// Parse flags and build up our service and executor options.
	ui.ParsingBatchSpec()
	batchSpec, rawSpec, err := parseBatchSpec(&opts.flags.file, svc)
	if err != nil {
		var multiErr *multierror.Error
		if errors.As(err, &multiErr) {
			ui.ParsingBatchSpecFailure(multiErr)
			return cmderrors.ExitCode(2, nil)
		} else {
			// This shouldn't happen; let's just punt and let the normal
			// rendering occur.
			return err
		}
	}
	ui.ParsingBatchSpecSuccess()

	ui.ResolvingNamespace()
	namespace, err := svc.ResolveNamespace(ctx, opts.flags.namespace)
	if err != nil {
		return err
	}
	ui.ResolvingNamespaceSuccess(namespace)

	var workspaceCreator workspace.Creator

	if svc.HasDockerImages(batchSpec) {
		ui.PreparingContainerImages()
		images, err := svc.EnsureDockerImages(ctx, batchSpec, ui.PreparingContainerImagesProgress)
		if err != nil {
			return err
		}
		ui.PreparingContainerImagesSuccess()

		ui.DeterminingWorkspaceCreatorType()
		workspaceCreator = workspace.NewCreator(ctx, opts.flags.workspace, opts.flags.cacheDir, opts.flags.tempDir, images)
		if workspaceCreator.Type() == workspace.CreatorTypeVolume {
			_, err = svc.EnsureImage(ctx, workspace.DockerVolumeWorkspaceImage)
			if err != nil {
				return err
			}
		}
		ui.DeterminingWorkspaceCreatorTypeSuccess(workspaceCreator.Type())
	}

	ui.ResolvingRepositories()
	repos, err := svc.ResolveRepositories(ctx, batchSpec)
	if err != nil {
		if repoSet, ok := err.(batches.UnsupportedRepoSet); ok {
			ui.ResolvingRepositoriesDone(repos, repoSet, nil)
		} else if repoSet, ok := err.(batches.IgnoredRepoSet); ok {
			ui.ResolvingRepositoriesDone(repos, nil, repoSet)
		} else {
			return errors.Wrap(err, "resolving repositories")
		}
	} else {
		ui.ResolvingRepositoriesDone(repos, nil, nil)
	}

	ui.DeterminingWorkspaces()
	workspaces, err := svc.DetermineWorkspaces(ctx, repos, batchSpec)
	if err != nil {
		return err
	}
	ui.DeterminingWorkspacesSuccess(len(workspaces))

	// EXECUTION OF TASKS
	coord := svc.NewCoordinator(executor.NewCoordinatorOpts{
		Creator:          workspaceCreator,
		CacheDir:         opts.flags.cacheDir,
		Cache:            executor.NewDiskCache(opts.flags.cacheDir),
		ClearCache:       opts.flags.clearCache,
		SkipErrors:       opts.flags.skipErrors,
		CleanArchives:    opts.flags.cleanArchives,
		Parallelism:      opts.flags.parallelism,
		Timeout:          opts.flags.timeout,
		KeepLogs:         opts.flags.keepLogs,
		TempDir:          opts.flags.tempDir,
		ImportChangesets: true,
	})

	ui.CheckingCache()
	tasks := svc.BuildTasks(ctx, batchSpec, workspaces)
	uncachedTasks, cachedSpecs, err := coord.CheckCache(ctx, tasks)
	if err != nil {
		return err
	}
	ui.CheckingCacheSuccess(len(cachedSpecs), len(uncachedTasks))

	taskExecUI := ui.ExecutingTasks(*verbose, opts.flags.parallelism)
	freshSpecs, logFiles, err := coord.Execute(ctx, uncachedTasks, batchSpec, taskExecUI)
	if err != nil && !opts.flags.skipErrors {
		return err
	}
	if err == nil || opts.flags.skipErrors {
		if err == nil {
			taskExecUI.Success()
		} else {
			ui.ExecutingTasksSkippingErrors(err)
		}
	} else {
		if err != nil {
			taskExecUI.Failed(err)
			return err
		}
	}

	if len(logFiles) > 0 && opts.flags.keepLogs {
		ui.LogFilesKept(logFiles)
	}

	specs := append(cachedSpecs, freshSpecs...)

	err = svc.ValidateChangesetSpecs(repos, specs)
	if err != nil {
		return err
	}

	ids := make([]graphql.ChangesetSpecID, len(specs))

	if len(specs) > 0 {
		ui.UploadingChangesetSpecs(len(specs))

		for i, spec := range specs {
			id, err := svc.CreateChangesetSpec(ctx, spec)
			if err != nil {
				return err
			}
			ids[i] = id
			ui.UploadingChangesetSpecsProgress(i+1, len(specs))
		}

		ui.UploadingChangesetSpecsSuccess(ids)
	} else if len(repos) == 0 {
		ui.NoChangesetSpecs()
	}

	ui.CreatingBatchSpec()
	id, url, err := svc.CreateBatchSpec(ctx, namespace, rawSpec, ids)
	if err != nil {
		return ui.CreatingBatchSpecError(err)
	}
	previewURL := cfg.Endpoint + url
	ui.CreatingBatchSpecSuccess(previewURL)

	if !opts.applyBatchSpec {
		ui.PreviewBatchSpec(previewURL)
		return
	}

	ui.ApplyingBatchSpec()
	batch, err := svc.ApplyBatchChange(ctx, id)
	if err != nil {
		return err
	}
	ui.ApplyingBatchSpecSuccess(cfg.Endpoint + batch.URL)

	return nil
}

// parseBatchSpec parses and validates the given batch spec. If the spec has
// validation errors, they are returned.
func parseBatchSpec(file *string, svc *service.Service) (*batcheslib.BatchSpec, string, error) {
	f, err := batchOpenFileFlag(file)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, "", errors.Wrap(err, "reading batch spec")
	}

	spec, err := svc.ParseBatchSpec(data)
	return spec, string(data), err
}

func checkExecutable(cmd string, args ...string) error {
	if err := exec.Command(cmd, args...).Run(); err != nil {
		return fmt.Errorf(
			"failed to execute \"%s %s\":\n\t%s\n\n'src batch' require %q to be available.",
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
