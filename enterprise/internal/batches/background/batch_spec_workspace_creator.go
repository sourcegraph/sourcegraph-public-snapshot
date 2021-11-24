package background

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/cache"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
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
	ws := make([]*btypes.BatchSpecWorkspace, 0, len(workspaces))
	// Collect all cache keys so we can look them up in a single query.
	cacheKeyWorkspaces := make(map[string]struct {
		dbWorkspace *btypes.BatchSpecWorkspace
		repo        batcheslib.Repository
	})

	// Build workspaces DB objects.
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

		r := batcheslib.Repository{
			ID:          string(graphqlbackend.MarshalRepositoryID(w.Repo.ID)),
			Name:        string(w.Repo.Name),
			BaseRef:     w.Branch,
			BaseRev:     string(w.Commit),
			FileMatches: w.FileMatches,
		}

		key := cache.KeyForWorkspace(
			&template.BatchChangeAttributes{
				Name:        spec.Spec.Name,
				Description: spec.Spec.Description,
			},
			r,
			w.Path,
			w.OnlyFetchWorkspace,
			w.Steps,
		)
		rawKey, err := key.Key()
		if err != nil {
			return err
		}
		cacheKeyWorkspaces[rawKey] = struct {
			dbWorkspace *btypes.BatchSpecWorkspace
			repo        batcheslib.Repository
		}{
			dbWorkspace: workspace,
			repo:        r,
		}
	}

	// Fetch all cache entries by their keys.
	cacheKeys := make([]string, 0, len(cacheKeyWorkspaces))
	for key := range cacheKeyWorkspaces {
		cacheKeys = append(cacheKeys, key)
	}
	entriesByCacheKey := make(map[string]*btypes.BatchSpecExecutionCacheEntry)
	changesetsByWorkspace := make(map[*btypes.BatchSpecWorkspace][]*btypes.ChangesetSpec)
	if len(cacheKeys) > 0 {
		entries, err := tx.ListBatchSpecExecutionCacheEntries(ctx, store.ListBatchSpecExecutionCacheEntriesOpts{
			Keys: cacheKeys,
		})
		if err != nil {
			return err
		}
		for _, entry := range entries {
			entriesByCacheKey[entry.Key] = entry
		}
	}

	// All changeset specs to be created.
	cs := make([]*btypes.ChangesetSpec, 0)
	// Collect all IDs of used cache entries to mark them as recently used later.
	usedCacheEntries := make([]int64, 0)

	// Check for an existing cache entry for each of the workspaces.
	for rawKey, workspace := range cacheKeyWorkspaces {
		entry, ok := entriesByCacheKey[rawKey]
		if !ok {
			continue
		}

		workspace.dbWorkspace.CachedResultFound = true

		// Build the changeset specs from the cache entry.
		changesetSpecs, err := service.DBChangesetSpecsFromCache(spec.ID, workspace.dbWorkspace.RepoID, spec.UserID, spec.Spec, workspace.repo, entry)
		if err != nil {
			return err
		}
		cs = append(cs, changesetSpecs...)
		changesetsByWorkspace[workspace.dbWorkspace] = changesetSpecs

		// And mark the cache entries as used.
		usedCacheEntries = append(usedCacheEntries, entry.ID)
	}

	// Mark all used cache entries as recently used for cache eviction purposes.
	if err := tx.MarkUsedBatchSpecExecutionCacheEntries(ctx, usedCacheEntries); err != nil {
		return err
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
	for _, c := range specs {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(graphql.ID(c.BaseRepository))
		if err != nil {
			return err
		}
		changesetSpec := &btypes.ChangesetSpec{
			UserID:      spec.UserID,
			RepoID:      repoID,
			Spec:        c,
			BatchSpecID: spec.ID,
		}
		cs = append(cs, changesetSpec)
	}

	if err = tx.CreateChangesetSpec(ctx, cs...); err != nil {
		return err
	}

	// Associate the changeset specs with the workspace.
	for workspace, changesetSpecs := range changesetsByWorkspace {
		for _, spec := range changesetSpecs {
			workspace.ChangesetSpecIDs = append(workspace.ChangesetSpecIDs, spec.ID)
		}
	}

	return tx.CreateBatchSpecWorkspace(ctx, ws...)
}
