package batches

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/version"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	srcInputPath = "input.json"
	srcCacheDir  = "cache"
	srcTempDir   = ".src-tmp"
	srcRepoDir   = "repository"
)

type BatchesStore interface {
	GetBatchSpecWorkspace(context.Context, store.GetBatchSpecWorkspaceOpts) (*btypes.BatchSpecWorkspace, error)
	GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error)

	DatabaseDB() database.DB
}

// transformRecord transforms a *btypes.BatchSpecWorkspaceExecutionJob into an apiclient.Job.
func transformRecord(ctx context.Context, logger log.Logger, s BatchesStore, job *btypes.BatchSpecWorkspaceExecutionJob, accessToken string) (apiclient.Job, error) {
	workspace, err := s.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: job.BatchSpecWorkspaceID})
	if err != nil {
		return apiclient.Job{}, errors.Wrapf(err, "fetching workspace %d", job.BatchSpecWorkspaceID)
	}

	batchSpec, err := s.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: workspace.BatchSpecID})
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching batch spec")
	}

	// This should never happen. To get some easier debugging when a user sees strange
	// behavior, we log some additional context.
	if job.UserID != batchSpec.UserID {
		logger.Error("bad DB state: batch spec workspace execution job did not have the same user ID as the associated batch spec")
	}

	// 🚨 SECURITY: Set the actor on the context so we check for permissions
	// when loading the repository.
	ctx = actor.WithActor(ctx, actor.FromUser(job.UserID))

	repo, err := s.DatabaseDB().Repos().Get(ctx, workspace.RepoID)
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching repo")
	}

	executionInput := batcheslib.WorkspacesExecutionInput{
		Repository: batcheslib.WorkspaceRepo{
			ID:   string(graphqlbackend.MarshalRepositoryID(repo.ID)),
			Name: string(repo.Name),
		},
		Branch: batcheslib.WorkspaceBranch{
			Name:   workspace.Branch,
			Target: batcheslib.Commit{OID: workspace.Commit},
		},
		Path:               workspace.Path,
		OnlyFetchWorkspace: workspace.OnlyFetchWorkspace,
		// TODO: We can further optimize here later and tell src-cli to
		// not run those steps so there is no discrepancy between the backend
		// and src-cli calculating the if conditions.
		Steps:             batchSpec.Spec.Steps,
		SearchResultPaths: workspace.FileMatches,
		BatchChangeAttributes: template.BatchChangeAttributes{
			Name:        batchSpec.Spec.Name,
			Description: batchSpec.Spec.Description,
		},
	}

	frontendURL := conf.ExecutorsFrontendURL()

	srcEndpoint, err := makeURL(frontendURL, accessToken)
	if err != nil {
		return apiclient.Job{}, err
	}

	redactedSrcEndpoint, err := makeURL(frontendURL, "PASSWORD_REMOVED")
	if err != nil {
		return apiclient.Job{}, err
	}

	marshaledInput, err := json.Marshal(executionInput)
	if err != nil {
		return apiclient.Job{}, err
	}

	files := map[string]string{srcInputPath: string(marshaledInput)}

	if !batchSpec.NoCache {
		// Find the cache entry for the _last_ step. src-cli only needs the most
		// recent cache entry to do its work.
		latestIndex := -1
		for idx := range workspace.StepCacheResults {
			if idx > latestIndex {
				latestIndex = idx
			}
		}
		if latestIndex != -1 {
			cacheEntry, _ := workspace.StepCacheResult(latestIndex)
			serializedCacheEntry, err := json.Marshal(cacheEntry.Value)
			if err != nil {
				return apiclient.Job{}, errors.Wrap(err, "serializing cache entry")
			}
			// Add file to BirtualMachineFiles.
			cacheFilePath := filepath.Join(srcCacheDir, fmt.Sprintf("%s.json", cacheEntry.Key))
			files[cacheFilePath] = string(serializedCacheEntry)
		}
	}

	sparseCheckout := []string{}
	if workspace.OnlyFetchWorkspace {
		sparseCheckout = []string{
			fmt.Sprintf("%s/*", workspace.Path),
		}
	}

	return apiclient.Job{
		ID:                  int(job.ID),
		VirtualMachineFiles: files,
		RepositoryName:      string(repo.Name),
		RepositoryDirectory: srcRepoDir,
		Commit:              workspace.Commit,
		// We only care about the current repos content, so a shallow clone is good enough.
		// Later we might allow to tweak more git parameters, like submodules and LFS.
		ShallowClone:   true,
		SparseCheckout: sparseCheckout,
		CliSteps: []apiclient.CliStep{
			{
				Commands: []string{
					"batch",
					"exec",
					"-f", srcInputPath,
					"-repo", srcRepoDir,
					// Tell src where to look for cache files.
					"-cache", srcCacheDir,
					// Tell src to store tmp files inside the workspace. Src currently
					// runs on the host and we don't want pollution outside of the workspace.
					"-tmp", srcTempDir,
					// Tell src which version of Sourcegraph it's talking to
					// to enable the right feature gates.
					"-sourcegraphVersion", version.Version(),
				},
				Dir: ".",
				Env: []string{
					// Make sure src never talks to any Sourcegraph instance.
					// TODO: This can go away once the execution mode is stripped down even
					// further. There should be no code path doing a request, but you never
					// know, software.
					fmt.Sprintf("SRC_ENDPOINT=%s", "http://127.0.0.1:10001"),
				},
			},
		},
		RedactedValues: map[string]string{
			// 🚨 SECURITY: Catch leak of upload endpoint. This is necessary in addition
			// to the below in case the username or password contains illegal URL characters,
			// which are then urlencoded and are not replaceable via byte comparison.
			srcEndpoint: redactedSrcEndpoint,

			// 🚨 SECURITY: Catch uses of fragments pulled from URL to construct another target
			// (in src-cli). We only pass the constructed URL to src-cli, which we trust not to
			// ship the values to a third party, but not to trust to ensure the values are absent
			// from the command's stdout or stderr streams.
			accessToken: "PASSWORD_REMOVED",
		},
	}, nil
}

func makeURL(base, password string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	u.User = url.UserPassword("sourcegraph", password)
	return u.String(), nil
}
