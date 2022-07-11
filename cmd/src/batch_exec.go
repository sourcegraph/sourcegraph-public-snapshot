package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/batches/docker"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/log"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
	"github.com/sourcegraph/src-cli/internal/batches/service"
	"github.com/sourcegraph/src-cli/internal/batches/ui"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
	"github.com/sourcegraph/src-cli/internal/cmderrors"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

const (
	execPullParallelism = 4
)

type executorModeFlags struct {
	timeout time.Duration
	file    string
	tempDir string
	repoDir string
}

func newExecutorModeFlags(flagSet *flag.FlagSet) (f *executorModeFlags) {
	f = &executorModeFlags{}
	flagSet.DurationVar(&f.timeout, "timeout", 60*time.Minute, "The maximum duration a single batch spec step can take.")
	flagSet.StringVar(&f.file, "f", "", "The workspace execution input file to read.")
	flagSet.StringVar(&f.tempDir, "tmp", "", "Directory for storing temporary data.")
	flagSet.StringVar(&f.repoDir, "repo", "", "Path of the checked out repo on disk.")

	return f
}

func validateExecutorModeFlags(f *executorModeFlags) error {
	if f.file == "" {
		return errors.New("input file parameter missing")
	}
	if f.tempDir == "" {
		return errors.New("tempDir parameter missing")
	}
	if f.repoDir == "" {
		return errors.New("repoDir parameter missing")
	}

	return nil
}

func init() {
	usage := `
INTERNAL USE ONLY: 'src batch exec' executes the given raw batch spec in the given workspaces.

The input file contains a JSON dump of the WorkspacesExecutionInput struct in
github.com/sourcegraph/sourcegraph/lib/batches.

Usage:

    src batch exec -f FILE -repo DIR [command options]

Examples:

    $ src batch exec -f batch-spec-with-workspaces.json

`

	flagSet := flag.NewFlagSet("exec", flag.ExitOnError)
	flags := newExecutorModeFlags(flagSet)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if len(flagSet.Args()) != 0 {
			return cmderrors.Usage("additional arguments not allowed")
		}

		if err := validateExecutorModeFlags(flags); err != nil {
			return cmderrors.ExitCode(1, err)
		}

		ctx, cancel := contextCancelOnInterrupt(context.Background())
		defer cancel()

		err := executeBatchSpecInWorkspaces(ctx, flags)
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

func executeBatchSpecInWorkspaces(ctx context.Context, flags *executorModeFlags) (err error) {
	ui := &ui.JSONLines{}

	// Ensure the temp dir exists.
	tempDir := flags.tempDir
	if !filepath.IsAbs(tempDir) {
		tempDir, err = filepath.Abs(tempDir)
		if err != nil {
			return errors.Wrap(err, "getting absolute path for temp dir")
		}

		if err := os.Mkdir(tempDir, os.ModePerm); err != nil {
			return errors.Wrap(err, "creating temp directory")
		}
	}

	// Grab the absolute path to the repo contents.
	repoDir := flags.repoDir
	if !filepath.IsAbs(repoDir) {
		repoDir, err = filepath.Abs(repoDir)
		if err != nil {
			return errors.Wrap(err, "getting absolute path for repo dir")
		}
	}

	// Test if git is available.
	if err := checkExecutable("git", "version"); err != nil {
		return err
	}
	// Test if docker is available.
	if err := checkExecutable("docker", "version"); err != nil {
		return err
	}

	// Read the input file that contains the raw spec and the workspaces in
	// which to execute it.
	input, err := loadWorkspaceExecutionInput(flags.file)
	if err != nil {
		return err
	}
	task := convertWorkspace(input)

	if len(task.Steps) == 0 {
		return errors.New("invalid execution, no steps to process")
	}

	imageCache := docker.NewImageCache()

	ui.PreparingContainerImages()
	_, err = service.New(&service.Opts{}).EnsureDockerImages(
		ctx,
		imageCache,
		task.Steps,
		execPullParallelism,
		ui.PreparingContainerImagesProgress,
	)
	if err != nil {
		return err
	}
	ui.PreparingContainerImagesSuccess()

	// Empty for now until we support secrets or env var settings in SSBC.
	globalEnv := []string{}
	isRemote := true

	// Set up the execution UI.
	taskExecUI := ui.ExecutingTasks(false, 1)
	taskExecUI.Start([]*executor.Task{task})
	taskExecUI.TaskStarted(task)

	opts := &executor.RunStepsOpts{
		Logger:      &log.NoopTaskLogger{},
		WC:          workspace.NewExecutorWorkspaceCreator(tempDir, repoDir),
		EnsureImage: imageCache.Ensure,
		Task:        task,
		// TODO: Should be slightly less than the executor timeout. Can we somehow read that?
		Timeout:   flags.timeout,
		TempDir:   tempDir,
		GlobalEnv: globalEnv,
		// Temporarily prevent the ability to sending a batch spec with a mount for server-side processing.
		IsRemote:    isRemote,
		RepoArchive: &repozip.NoopArchive{},
		UI:          taskExecUI.StepsExecutionUI(task),
	}
	results, err := executor.RunSteps(ctx, opts)

	// Write all step cache results for all results.
	for _, stepRes := range results {
		cacheKey := task.CacheKey(globalEnv, isRemote, stepRes.StepIndex)
		k, err := cacheKey.Key()
		if err != nil {
			return errors.Wrap(err, "calculating step cache key")
		}
		ui.WriteAfterStepResult(k, stepRes)
	}

	taskExecUI.TaskFinished(task, err)

	return err
}

func loadWorkspaceExecutionInput(file string) (input batcheslib.WorkspacesExecutionInput, err error) {
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

// convertWorkspace takes the WorkspacesExecutionInput and restructures it into
// an executor.Task.
func convertWorkspace(w batcheslib.WorkspacesExecutionInput) *executor.Task {
	fileMatches := make(map[string]bool)
	for _, path := range w.SearchResultPaths {
		fileMatches[path] = true
	}

	task := &executor.Task{
		Repository: &graphql.Repository{
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
		Path:                  w.Path,
		Steps:                 w.Steps,
		OnlyFetchWorkspace:    w.OnlyFetchWorkspace,
		BatchChangeAttributes: &w.BatchChangeAttributes,
		CachedStepResultFound: w.CachedStepResultFound,
		CachedStepResult:      w.CachedStepResult,
	}

	return task
}
