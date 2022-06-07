package workers

import (
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/log"

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
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// batchSpecWorkspaceCreator takes in BatchSpecs, resolves them into
// RepoWorkspaces and then persists those as pending BatchSpecWorkspaces.
type batchSpecWorkspaceCreator struct {
	store *store.Store
}

// HandlerFunc returns a workeruitl.HandlerFunc that can be passed to a
// workerutil.Worker to process queued changesets.
func (e *batchSpecWorkspaceCreator) HandlerFunc() workerutil.HandlerFunc {
	return func(ctx context.Context, logger log.Logger, record workerutil.Record) (err error) {
		job := record.(*btypes.BatchSpecResolutionJob)

		tx, err := e.store.Transact(ctx)
		if err != nil {
			return err
		}
		defer func() { err = tx.Done(err) }()

		return e.process(ctx, tx, service.NewWorkspaceResolver, job)
	}
}

type workspaceCacheKey struct {
	dbWorkspace   *btypes.BatchSpecWorkspace
	repo          batcheslib.Repository
	stepCacheKeys map[int]string
	skippedSteps  map[int32]struct{}
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
	cacheKeyWorkspaces := make(map[string]workspaceCacheKey)

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

			Unsupported: w.Unsupported,
			Ignored:     w.Ignored,
		}

		ws = append(ws, workspace)

		if spec.NoCache {
			continue
		}
		if !spec.AllowIgnored && w.Ignored {
			continue
		}
		if !spec.AllowUnsupported && w.Unsupported {
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
			spec.Spec.Steps,
		)

		rawKey, err := key.Key()
		if err != nil {
			return err
		}

		skippedSteps, err := batcheslib.SkippedStepsForRepo(spec.Spec, string(w.Repo.Name), w.FileMatches)
		if err != nil {
			return err
		}

		stepCacheKeys := make(map[int]string, len(spec.Spec.Steps))
		// Generate cache keys for all the step results as well.
		for i := 0; i < len(spec.Spec.Steps); i++ {
			if _, ok := skippedSteps[int32(i)]; ok {
				continue
			}
			key := cache.StepsCacheKey{ExecutionKey: &key, StepIndex: i}
			rawStepKey, err := key.Key()
			if err != nil {
				return nil
			}
			stepCacheKeys[i] = rawStepKey
		}

		cacheKeyWorkspaces[rawKey] = workspaceCacheKey{
			dbWorkspace:   workspace,
			repo:          r,
			stepCacheKeys: stepCacheKeys,
			skippedSteps:  skippedSteps,
		}
	}

	// Fetch all cache entries by their keys.
	cacheKeys := make([]string, 0, len(cacheKeyWorkspaces))
	stepCacheKeys := make([]string, 0, len(cacheKeyWorkspaces))
	for key, w := range cacheKeyWorkspaces {
		cacheKeys = append(cacheKeys, key)

		for _, key := range w.stepCacheKeys {
			stepCacheKeys = append(stepCacheKeys, key)
		}
	}
	entriesByCacheKey := make(map[string]*btypes.BatchSpecExecutionCacheEntry)
	if len(cacheKeys) > 0 {
		entries, err := tx.ListBatchSpecExecutionCacheEntries(ctx, store.ListBatchSpecExecutionCacheEntriesOpts{
			UserID: spec.UserID,
			Keys:   cacheKeys,
		})
		if err != nil {
			return err
		}
		for _, entry := range entries {
			entriesByCacheKey[entry.Key] = entry
		}
	}
	stepEntriesByCacheKey := make(map[string]*btypes.BatchSpecExecutionCacheEntry)
	if len(stepCacheKeys) > 0 {
		entries, err := tx.ListBatchSpecExecutionCacheEntries(ctx, store.ListBatchSpecExecutionCacheEntriesOpts{
			UserID: spec.UserID,
			Keys:   stepCacheKeys,
		})
		if err != nil {
			return err
		}
		for _, entry := range entries {
			stepEntriesByCacheKey[entry.Key] = entry
		}
	}

	// All changeset specs to be created.
	cs := make([]*btypes.ChangesetSpec, 0)
	// Collect all IDs of used cache entries to mark them as recently used later.
	usedCacheEntries := make([]int64, 0)
	changesetsByWorkspace := make(map[*btypes.BatchSpecWorkspace][]*btypes.ChangesetSpec)

	// Check for an existing cache entry for each of the workspaces.
	for rawKey, workspace := range cacheKeyWorkspaces {
		for idx, key := range workspace.stepCacheKeys {
			if c, ok := stepEntriesByCacheKey[key]; ok {
				var res execution.AfterStepResult
				if err := json.Unmarshal([]byte(c.Value), &res); err != nil {
					return err
				}
				workspace.dbWorkspace.SetStepCacheResult(idx+1, btypes.StepCacheResult{Key: key, Value: &res})

				// Mark the cache entry as used.
				usedCacheEntries = append(usedCacheEntries, c.ID)
			} else {
				// Only add cache entries up until we don't have the cache entry
				// for the previous step anymore.
				break
			}
		}

		entry, ok := entriesByCacheKey[rawKey]
		if !ok {
			// If no cache entry is found, maybe we have a step cache entry for the last step instead.
			// Since those can be converted without execution, we can recreate that entry here.
			// TODO: This technically means that we don't need execution entries at all, we could always just
			// run the code below to create a cache result from the execution cache entry.
			// This would reduce data stored in the DB. We can then also stop logging it in src-cli which also
			// makes the logs smaller (often by close to 50%, when there is just 1 step!).

			// Find the latest step that is not statically skipped.
			latestStepIdx := -1
			for i := len(spec.Spec.Steps) - 1; i >= 0; i-- {
				// Keep skipping steps until the first one is hit that we do want to run.
				if _, ok := workspace.skippedSteps[int32(i)]; ok {
					continue
				}
				// TODO: Is this required?
				i := i
				latestStepIdx = i
				break
			}
			if latestStepIdx != -1 {
				res, found := workspace.dbWorkspace.StepCacheResult(latestStepIdx + 1)
				if !found {
					// TODO: this is an error! It should have been set right above in l.222.
					continue
				}
				var execResult execution.Result
				// Set the Outputs to the cached outputs
				execResult.Outputs = res.Value.Outputs

				changes, err := git.ChangesInDiff([]byte(res.Value.Diff))
				if err != nil {
					return errors.Wrap(err, "parsing cached step diff")
				}

				execResult.Diff = res.Value.Diff
				execResult.ChangedFiles = &changes
				// TODO: This is not in src-cli, is it missing?
				execResult.Path = workspace.dbWorkspace.Path

				entry, err = btypes.NewCacheEntryFromResult(rawKey, execResult)
				if err != nil {
					return errors.Wrap(err, "NewCacheEntryFromResult")
				}
				entry.UserID = spec.UserID

				if err := tx.CreateBatchSpecExecutionCacheEntry(ctx, entry); err != nil {
					return errors.Wrap(err, "storing cache entry")
				}
			} else {
				continue
			}
		}

		workspace.dbWorkspace.CachedResultFound = true

		// Mark the cache entries as used.
		usedCacheEntries = append(usedCacheEntries, entry.ID)

		// Build the changeset specs from the cache entry.
		var executionResult execution.Result
		if err := json.Unmarshal([]byte(entry.Value), &executionResult); err != nil {
			return err
		}

		rawSpecs, err := cache.ChangesetSpecsFromCache(spec.Spec, workspace.repo, executionResult)
		if err != nil {
			return err
		}

		var specs []*btypes.ChangesetSpec
		for _, s := range rawSpecs {
			changesetSpec, err := btypes.NewChangesetSpecFromSpec(s)
			if err != nil {
				return err
			}
			changesetSpec.BatchSpecID = spec.ID
			changesetSpec.RepoID = workspace.dbWorkspace.RepoID
			changesetSpec.UserID = spec.UserID

			specs = append(specs, changesetSpec)
		}

		cs = append(cs, specs...)
		changesetsByWorkspace[workspace.dbWorkspace] = specs
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

	// Associate the changeset specs with the workspace now that they have IDs.
	for workspace, changesetSpecs := range changesetsByWorkspace {
		for _, spec := range changesetSpecs {
			workspace.ChangesetSpecIDs = append(workspace.ChangesetSpecIDs, spec.ID)
		}
	}

	return tx.CreateBatchSpecWorkspace(ctx, ws...)
}
