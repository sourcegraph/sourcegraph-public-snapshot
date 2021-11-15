package background

import (
	"context"
	"encoding/json"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

// batchSpecWorkspaceCreator takes in BatchSpecs, resolves them into
// RepoWorkspaces and then persists those as pending BatchSpecWorkspaces.
type batchSpecWorkspaceCreator struct {
	store *store.Store
}

// HandlerFunc returns a workeruitl.HandlerFunc that can be passed to a
// workerutil.Worker to process queued changesets.
func (e *batchSpecWorkspaceCreator) HandlerFunc() workerutil.HandlerFunc {
	return func(ctx context.Context, record workerutil.Record) (err error) {
		job := record.(*btypes.BatchSpecResolutionJob)

		tx, err := e.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		return e.process(ctx, tx, service.NewWorkspaceResolver, job)
	}
}

func (r *batchSpecWorkspaceCreator) process(
	ctx context.Context,
	tx *store.Store,
	newResolver service.WorkspaceResolverBuilder,
	job *btypes.BatchSpecResolutionJob,
) error {
	spec, err := tx.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: job.BatchSpecID})
	if err != nil {
		return err
	}

	evaluatableSpec, err := batcheslib.ParseBatchSpec([]byte(spec.RawSpec), batcheslib.ParseBatchSpecOptions{
		AllowTransformChanges: true,
		AllowConditionalExec:  true,
		// We don't allow forwarding of environment variables in server-side
		// batch changes, since we'd then leak the executor/Firecracker
		// internal environment.
		AllowArrayEnvironments: false,
	})
	if err != nil {
		return err
	}

	resolver := newResolver(tx)
	userCtx := actor.WithActor(ctx, actor.FromUser(spec.UserID))
	workspaces, err := resolver.ResolveWorkspacesForBatchSpec(userCtx, evaluatableSpec)
	if err != nil {
		return err
	}

	log15.Info("resolved workspaces for batch spec", "job", job.ID, "spec", spec.ID, "workspaces", len(workspaces))

	// Build DB workspaces and check for cache entries.
	var ws []*btypes.BatchSpecWorkspace
	for _, w := range workspaces {
		workspace := &btypes.BatchSpecWorkspace{
			BatchSpecID:      spec.ID,
			ChangesetSpecIDs: []int64{},

			RepoID:             w.Repo.ID,
			Branch:             w.Branch,
			Commit:             string(w.Commit),
			Path:               w.Path,
			FileMatches:        w.FileMatches,
			OnlyFetchWorkspace: w.OnlyFetchWorkspace,
			Steps:              w.Steps,

			Unsupported: w.Unsupported,
			Ignored:     w.Ignored,
		}

		ws = append(ws, workspace)

		if spec.NoCache {
			continue
		}

		rawKey, err := cacheKeyForWorkspace(spec, w)
		if err != nil {
			return err
		}

		entry, err := tx.GetBatchSpecExecutionCacheEntry(ctx, store.GetBatchSpecExecutionCacheEntryOpts{
			Key: rawKey,
		})
		if err != nil && err != store.ErrNoResults {
			return err
		}
		if err == store.ErrNoResults {
			continue
		}

		workspace.CachedResultFound = true

		changesetSpecs, err := changesetSpecsFromCache(spec, w, entry)
		if err != nil {
			return err
		}
		for _, spec := range changesetSpecs {
			if err := tx.CreateChangesetSpec(ctx, spec); err != nil {
				return err
			}
			workspace.ChangesetSpecIDs = append(workspace.ChangesetSpecIDs, spec.ID)
		}

		if err := tx.MarkUsedBatchSpecExecutionCacheEntry(ctx, entry.ID); err != nil {
			return err
		}
	}

	// If there are "importChangesets" statements in the spec we evaluate
	// them now and create ChangesetSpecs for them.
	reposStore := tx.Repos()
	specs, err := batcheslib.BuildImportChangesetSpecs(ctx, evaluatableSpec.ImportChangesets, func(ctx context.Context, repoNames []string) (map[string]string, error) {
		if len(repoNames) == 0 {
			return map[string]string{}, nil
		}

		// 🚨 SECURITY: We use database.Repos.List to get the ID and also to check
		// whether the user has access to the repository or not.
		repos, err := reposStore.List(ctx, database.ReposListOptions{Names: repoNames})
		if err != nil {
			return nil, err
		}

		repoNameIDs := make(map[string]string, len(repos))
		for _, r := range repos {
			repoNameIDs[string(r.Name)] = string(graphqlbackend.MarshalRepositoryID(r.ID))
		}
		return repoNameIDs, nil
	})
	if err != nil {
		return err
	}
	for _, cs := range specs {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(cs.BaseRepository))
		if err != nil {
			return err
		}
		changesetSpec := &btypes.ChangesetSpec{
			UserID:      spec.UserID,
			RepoID:      repoID,
			Spec:        cs,
			BatchSpecID: spec.ID,
		}

		if err = tx.CreateChangesetSpec(ctx, changesetSpec); err != nil {
			return err
		}
	}

	return tx.CreateBatchSpecWorkspace(ctx, ws...)
}

func cacheKeyForWorkspace(spec *btypes.BatchSpec, w *service.RepoWorkspace) (string, error) {
	fileMatches := w.FileMatches
	sort.Strings(fileMatches)

	executionKey := cache.ExecutionKey{
		Repository: batcheslib.Repository{
			ID:          string(graphqlbackend.MarshalRepositoryID(w.Repo.ID)),
			Name:        string(w.Repo.Name),
			BaseRef:     git.EnsureRefPrefix(w.Branch),
			BaseRev:     string(w.Commit),
			FileMatches: fileMatches,
		},
		Path:               w.Path,
		OnlyFetchWorkspace: w.OnlyFetchWorkspace,
		Steps:              w.Steps,
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        spec.Spec.Name,
			Description: spec.Spec.Description,
		},
	}
	return executionKey.Key()
}

func changesetSpecsFromCache(spec *btypes.BatchSpec, w *service.RepoWorkspace, entry *btypes.BatchSpecExecutionCacheEntry) ([]*btypes.ChangesetSpec, error) {
	var executionResult execution.Result
	if err := json.Unmarshal([]byte(entry.Value), &executionResult); err != nil {
		return nil, err
	}

	repoID := string(graphqlbackend.MarshalRepositoryID(w.Repo.ID))
	input := &batcheslib.ChangesetSpecInput{
		BaseRepositoryID: repoID,
		HeadRepositoryID: repoID,
		Repository: batcheslib.ChangesetSpecRepository{
			Name:        string(w.Repo.Name),
			FileMatches: w.FileMatches,
			BaseRef:     git.EnsureRefPrefix(w.Branch),
			BaseRev:     string(w.Commit),
		},
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        spec.Spec.Name,
			Description: spec.Spec.Description,
		},
		Template:         spec.Spec.ChangesetTemplate,
		TransformChanges: spec.Spec.TransformChanges,
		Result:           executionResult,
	}

	rawSpecs, err := batcheslib.BuildChangesetSpecs(input, batcheslib.ChangesetSpecFeatureFlags{
		IncludeAutoAuthorDetails: true,
		AllowOptionalPublished:   true,
	})
	if err != nil {
		return nil, err
	}

	var specs []*btypes.ChangesetSpec
	for _, s := range rawSpecs {
		changesetSpec, err := btypes.NewChangesetSpecFromSpec(s)
		if err != nil {
			return nil, err
		}
		changesetSpec.BatchSpecID = spec.ID
		changesetSpec.RepoID = w.Repo.ID
		changesetSpec.UserID = spec.UserID

		specs = append(specs, changesetSpec)

	}
	return specs, nil
}
