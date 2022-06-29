package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
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
	sourcegraphVersion string
	timeout            time.Duration
	file               string
	tempDir            string
	cacheDir           string
	repoDir            string
}

func newExecutorModeFlags(flagSet *flag.FlagSet) (f *executorModeFlags) {
	f = &executorModeFlags{}
	flagSet.StringVar(&f.sourcegraphVersion, "sourcegraphVersion", "", "Sourcegraph backend version.")
	flagSet.DurationVar(&f.timeout, "timeout", 60*time.Minute, "The maximum duration a single batch spec step can take.")
	flagSet.StringVar(&f.file, "f", "", "The workspace execution input file to read.")
	flagSet.StringVar(&f.tempDir, "tmp", "", "Directory for storing temporary data.")
	flagSet.StringVar(&f.cacheDir, "cache", "", "Directory to read cached results from.")
	flagSet.StringVar(&f.repoDir, "repo", "", "Path of the checked out repo on disk.")

	return f
}

func validateExecutorModeFlags(f *executorModeFlags) error {
	if f.sourcegraphVersion == "" {
		return errors.New("sourcegraphVersion parameter missing")
	}
	if f.file == "" {
		return errors.New("input file parameter missing")
	}
	if f.tempDir == "" {
		return errors.New("tempDir parameter missing")
	}
	if f.cacheDir == "" {
		return errors.New("cacheDir parameter missing")
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
	defer func() {
		if err != nil {
			ui.ExecutionError(err)
		}
	}()

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

	repoDir := flags.repoDir
	if !filepath.IsAbs(repoDir) {
		repoDir, err = filepath.Abs(repoDir)
		if err != nil {
			return errors.Wrap(err, "getting absolute path for repo dir")
		}
	}

	svc := service.New(&service.Opts{
		// When this workspace made it to here, it's already been validated.
		AllowUnsupported: true,
		// When this workspace made it to here, it's already been validated.
		AllowIgnored: true,
		// We don't want src to talk to the sg instance, if it would, this
		// is a regression. Therefor, we have this dead client that kills the
		// process when something should try to talk to src.
		Client: &deadClient{},
	})

	if err := svc.SetFeatureFlagsForRelease(flags.sourcegraphVersion); err != nil {
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
	input, err := loadWorkspaceExecutionInput(flags.file)
	if err != nil {
		return err
	}

	// Since we already know which workspace we want to execute the steps in,
	// we can convert it to a RepoWorkspace and build a task only for that one.
	tasks := svc.BuildTasks(ctx, &input.BatchChangeAttributes, []service.RepoWorkspace{convertWorkspace(input)})

	if len(tasks) != 1 {
		return errors.New("invalid input, didn't yield exactly one task")
	}

	task := tasks[0]

	if len(task.Steps) == 0 {
		return errors.New("invalid execution, no steps to process")
	}

	{
		ui.PreparingContainerImages()
		_, err = svc.EnsureDockerImages(
			ctx,
			task.Steps,
			execPullParallelism,
			ui.PreparingContainerImagesProgress,
		)
		if err != nil {
			return err
		}
		ui.PreparingContainerImagesSuccess()
	}

	coord := svc.NewCoordinator(
		repozip.NewNoopRegistry(),
		log.NewNoopManager(),
		executor.NewCoordinatorOpts{
			Creator:     workspace.NewExecutorWorkspaceCreator(tempDir, repoDir),
			Cache:       &executor.ServerSideCache{CacheDir: flags.cacheDir, Writer: ui},
			Parallelism: 1,
			// TODO: Should be slightly less than the executor timeout. Can we somehow read that?
			Timeout: flags.timeout,
			TempDir: tempDir,
			// Don't allow to read from env.
			GlobalEnv: []string{},
			// Temporarily prevent the ability to sending a batch spec with a mount for server-side processing.
			IsRemote: true,
		},
	)

	// `src batch exec` uses server-side caching for changeset specs, so we
	// only need to call `CheckStepResultsCache` to make sure that per-step cache entries
	// are loaded and set on the tasks.
	if err := coord.CheckStepResultsCache(
		ctx,
		tasks,
		// Don't expose the executor env, we don't allow env forwarding anyways.
		[]string{},
	); err != nil {
		return err
	}

	// These arguments are unused in the json logs implementation, but the interface
	// dictates them.
	taskExecUI := ui.ExecutingTasks(false, 1)
	err = coord.Execute(ctx, tasks, taskExecUI)
	if err != nil {
		taskExecUI.Failed(err)
		return err
	}

	taskExecUI.Success()
	return nil
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

type deadClient struct{}

var _ api.Client = &deadClient{}

const deadClientPanicMsg = "Dead client invoked. This indicates a bug in src-cli in server-side execution, please report this."

func (c *deadClient) NewQuery(query string) api.Request {
	panic(deadClientPanicMsg)
}
func (c *deadClient) NewRequest(query string, vars map[string]interface{}) api.Request {
	panic(deadClientPanicMsg)
}
func (c *deadClient) NewGzippedRequest(query string, vars map[string]interface{}) api.Request {
	panic(deadClientPanicMsg)
}
func (c *deadClient) NewGzippedQuery(query string) api.Request {
	panic(deadClientPanicMsg)
}
func (c *deadClient) NewHTTPRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	panic(deadClientPanicMsg)
}
func (c *deadClient) Do(req *http.Request) (*http.Response, error) {
	panic(deadClientPanicMsg)
}
