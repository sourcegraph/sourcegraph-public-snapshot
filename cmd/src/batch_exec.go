package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/batches/ui"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
	"github.com/sourcegraph/src-cli/internal/cmderrors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func init() {
	usage := `
INTERNAL USE ONLY: 'src batch exec' executes the given raw batch spec in the given workspaces.

The input file contains a JSON dump of the WorkspacesExecutionInput struct in
github.com/sourcegraph/sourcegraph/lib/batches.

Usage:

    src batch exec -f FILE [command options]

Examples:

    $ src batch exec -f batch-spec-with-workspaces.json

`

	flagSet := flag.NewFlagSet("exec", flag.ExitOnError)
	flags := newBatchExecuteFlags(flagSet, true, batchDefaultCacheDir(), batchDefaultTempDirPrefix())

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return cmderrors.Usage("additional arguments not allowed")
		}

		ctx, cancel := contextCancelOnInterrupt(context.Background())
		defer cancel()

		err := executeBatchSpecInWorkspaces(ctx, &ui.JSONLines{}, executeBatchSpecOpts{
			flags:  flags,
			client: cfg.apiClient(flags.api, flagSet.Output()),
		})
		if err != nil {
			return cmderrors.ExitCode(1, nil)
		}

		return nil
	}

	batchCommands = append(batchCommands, &command{
		flagSet: flagSet,
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src batch %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

func executeBatchSpecInWorkspaces(ctx context.Context, ui *ui.JSONLines, opts executeBatchSpecOpts) (err error) {
	defer func() {
		if err != nil {
			ui.ExecutionError(err)
		}
	}()

	svc := service.New(&service.Opts{
		AllowUnsupported: opts.flags.allowUnsupported,
		AllowIgnored:     opts.flags.allowIgnored,
		AllowFiles:       false,
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

	// Read the input file that contains the raw spec and the workspaces in
	// which to execute it.
	input, err := loadWorkspaceExecutionInput(opts.flags.file)
	if err != nil {
		return err
	}

	// Since we already know which workspaces we want to execute the steps in,
	// we can convert them to RepoWorkspaces and build tasks only for those.
	repoWorkspaces := convertWorkspaces(input.Workspaces)

	// Parse the raw batch spec contained in the input
	ui.ParsingBatchSpec()
	batchSpec, err := svc.ParseBatchSpec([]byte(input.RawSpec))
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

	// EXECUTION OF TASKS
	coord := svc.NewCoordinator(executor.NewCoordinatorOpts{
		Creator:       workspaceCreator,
		CacheDir:      opts.flags.cacheDir,
		Cache:         &executor.JSONLinesCache{Writer: ui},
		ClearCache:    opts.flags.clearCache,
		SkipErrors:    opts.flags.skipErrors,
		CleanArchives: opts.flags.cleanArchives,
		Parallelism:   opts.flags.parallelism,
		Timeout:       opts.flags.timeout,
		KeepLogs:      opts.flags.keepLogs,
		TempDir:       opts.flags.tempDir,
		// Do not import changesets in `src batch exec`
		ImportChangesets: false,
	})

	ui.CheckingCache()
	tasks := svc.BuildTasks(ctx, batchSpec, repoWorkspaces)
	uncachedTasks, cachedSpecs, err := coord.CheckCache(ctx, tasks)
	if err != nil {
		return err
	}
	ui.CheckingCacheSuccess(len(cachedSpecs), len(uncachedTasks))

	taskExecUI := ui.ExecutingTasks(*verbose, opts.flags.parallelism)
	freshSpecs, _, err := coord.Execute(ctx, uncachedTasks, batchSpec, taskExecUI)
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

	specs := append(cachedSpecs, freshSpecs...)

	ids := make([]graphql.ChangesetSpecID, len(specs))

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

	return nil
}

func loadWorkspaceExecutionInput(file string) (batcheslib.WorkspacesExecutionInput, error) {
	var input batcheslib.WorkspacesExecutionInput

	f, err := batchOpenFileFlag(&file)
	if err != nil {
		return input, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return input, errors.Wrap(err, "reading workspace execution input file")
	}

	if err := json.Unmarshal(data, &input); err != nil {
		return input, errors.Wrap(err, "unmarshaling workspace execution input file")
	}

	return input, nil
}

func convertWorkspaces(ws []*batcheslib.Workspace) []service.RepoWorkspace {
	workspaces := make([]service.RepoWorkspace, 0, len(ws))

	for _, w := range ws {
		fileMatches := make(map[string]bool)
		for _, path := range w.SearchResultPaths {
			fileMatches[path] = true
		}

		workspaces = append(workspaces, service.RepoWorkspace{
			Repo: &graphql.Repository{
				ID:   w.Repository.ID,
				Name: w.Repository.Name,
				Branch: graphql.Branch{
					Name: w.Branch.Name,
					Target: graphql.Target{
						OID: w.Branch.Target.OID,
					},
				},
				Commit:      graphql.Target{OID: w.Branch.Target.OID},
				FileMatches: fileMatches,
			},
			Path:               w.Path,
			Steps:              w.Steps,
			OnlyFetchWorkspace: w.OnlyFetchWorkspace,
		})
	}

	return workspaces
}
