package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/lib/errors"

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

	// Since we already know which workspace we want to execute the steps in,
	// we can convert it to a RepoWorkspace and build a task only for that one.
	repoWorkspace := convertWorkspace(input)

	var workspaceCreator workspace.Creator

	if len(input.Steps) > 0 {
		ui.PreparingContainerImages()
		images, err := svc.EnsureDockerImages(
			ctx, input.Steps, opts.flags.parallelism,
			ui.PreparingContainerImagesProgress,
		)
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
		Cache:         &executor.ServerSideCache{Writer: ui},
		SkipErrors:    opts.flags.skipErrors,
		CleanArchives: opts.flags.cleanArchives,
		Parallelism:   opts.flags.parallelism,
		Timeout:       opts.flags.timeout,
		KeepLogs:      opts.flags.keepLogs,
		TempDir:       opts.flags.tempDir,
		// Temporarily prevent the ability to sending a batch spec with a mount for server-side processing.
		AllowPathMounts: false,
	})

	// `src batch exec` uses server-side caching for changeset specs, so we
	// only need to call `CheckStepResultsCache` to make sure that per-step cache entries
	// are loaded and set on the tasks.
	tasks := svc.BuildTasks(ctx, &input.BatchChangeAttributes, []service.RepoWorkspace{repoWorkspace})
	if err := coord.CheckStepResultsCache(ctx, tasks); err != nil {
		return err
	}

	taskExecUI := ui.ExecutingTasks(*verbose, opts.flags.parallelism)
	err = coord.Execute(ctx, tasks, taskExecUI)
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

	return nil
}

func loadWorkspaceExecutionInput(file string) (batcheslib.WorkspacesExecutionInput, error) {
	var input batcheslib.WorkspacesExecutionInput

	f, err := batchOpenFileFlag(file)
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

func convertWorkspace(w batcheslib.WorkspacesExecutionInput) service.RepoWorkspace {
	fileMatches := make(map[string]bool)
	for _, path := range w.SearchResultPaths {
		fileMatches[path] = true
	}
	return service.RepoWorkspace{
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
	}
}
