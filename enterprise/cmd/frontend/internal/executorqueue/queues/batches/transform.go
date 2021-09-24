package batches

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

const (
	accessTokenNote  = "batch-spec-execution"
	accessTokenScope = "user:all"
)

func createAccessToken(ctx context.Context, db dbutil.DB, userID int32) (string, error) {
	_, token, err := database.AccessTokens(db).Create(ctx, userID, []string{accessTokenScope}, accessTokenNote, userID)
	if err != nil {
		return "", err
	}
	return token, err
}

func makeURL(base, username, password string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	u.User = url.UserPassword(username, password)
	return u.String(), nil
}

type batchesStore interface {
	GetBatchSpecWorkspace(context.Context, store.GetBatchSpecWorkspaceOpts) (*btypes.BatchSpecWorkspace, error)
	GetBatchSpec(context.Context, store.GetBatchSpecOpts) (*btypes.BatchSpec, error)

	DB() dbutil.DB
}

// transformBatchSpecWorkspaceExecutionJobRecord transforms a *btypes.BatchSpecWorkspaceExecutionJob into an apiclient.Job.
func transformBatchSpecWorkspaceExecutionJobRecord(ctx context.Context, s batchesStore, job *btypes.BatchSpecWorkspaceExecutionJob, config *Config) (apiclient.Job, error) {
	// MAYBE: We could create a view in which batch_spec and repo are joined
	// against the batch_spec_workspace_job so we don't have to load them
	// separately.
	workspace, err := s.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: job.BatchSpecWorkspaceID})
	if err != nil {
		return apiclient.Job{}, errors.Wrapf(err, "fetching workspace %d", job.BatchSpecWorkspaceID)
	}

	batchSpec, err := s.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: workspace.BatchSpecID})
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching batch spec")
	}

	// 🚨 SECURITY: Set the actor on the context so we check for permissions
	// when loading the repository.
	ctx = actor.WithActor(ctx, actor.FromUser(batchSpec.UserID))

	repo, err := database.Repos(s.DB()).Get(ctx, workspace.RepoID)
	if err != nil {
		return apiclient.Job{}, errors.Wrap(err, "fetching repo")
	}

	executionInput := batcheslib.WorkspacesExecutionInput{
		RawSpec: batchSpec.RawSpec,
		Workspaces: []*batcheslib.Workspace{
			{
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
				Steps:              workspace.Steps,
				SearchResultPaths:  workspace.FileMatches,
			},
		},
	}

	// TODO: createAccessToken is a bit of technical debt until we figure out a
	// better solution. The problem is that src-cli needs to make requests to
	// the Sourcegraph instance *on behalf of the user*.
	//
	// Ideally we'd have something like one-time tokens that
	// * we could hand to src-cli
	// * are not visible to the user in the Sourcegraph web UI
	// * valid only for the duration of the batch spec execution
	// * and cleaned up after batch spec is executed
	//
	// Until then we create a fresh access token every time.
	//
	// GetOrCreate doesn't work because once an access token has been created
	// in the database Sourcegraph can't access the plain-text token anymore.
	// Only a hash for verification is kept in the database.
	token, err := createAccessToken(ctx, s.DB(), batchSpec.UserID)
	if err != nil {
		return apiclient.Job{}, err
	}

	frontendURL := conf.Get().ExternalURL

	srcEndpoint, err := makeURL(frontendURL, config.Shared.FrontendUsername, config.Shared.FrontendPassword)
	if err != nil {
		return apiclient.Job{}, err
	}

	redactedSrcEndpoint, err := makeURL(frontendURL, "USERNAME_REMOVED", "PASSWORD_REMOVED")
	if err != nil {
		return apiclient.Job{}, err
	}

	cliEnv := []string{
		fmt.Sprintf("SRC_ENDPOINT=%s", srcEndpoint),
		fmt.Sprintf("SRC_ACCESS_TOKEN=%s", token),
	}

	marshaledInput, err := json.Marshal(executionInput)
	if err != nil {
		return apiclient.Job{}, err
	}

	return apiclient.Job{
		ID:                  int(job.ID),
		VirtualMachineFiles: map[string]string{"input.json": string(marshaledInput)},
		CliSteps: []apiclient.CliStep{
			{
				Commands: []string{
					"batch",
					"exec",
					"-f", "input.json",
					"-skip-errors",
				},
				Dir: ".",
				Env: cliEnv,
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
			config.Shared.FrontendUsername: "USERNAME_REMOVED",
			config.Shared.FrontendPassword: "PASSWORD_REMOVED",

			// 🚨 SECURITY: Redact the access token used for src-cli to talk to
			// Sourcegraph instance.
			token: "SRC_ACCESS_TOKEN_REMOVED",
		},
	}, nil
}
