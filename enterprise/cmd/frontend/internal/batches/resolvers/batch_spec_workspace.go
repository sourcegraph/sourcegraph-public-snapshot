package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

const batchSpecWorkspaceIDKind = "BatchSpecWorkspace"

func marshalBatchSpecWorkspaceID(id int64) graphql.ID {
	return relay.MarshalID(batchSpecWorkspaceIDKind, id)
}

func unmarshalBatchSpecWorkspaceID(id graphql.ID) (batchSpecWorkspaceID int64, err error) {
	err = relay.UnmarshalSpec(id, &batchSpecWorkspaceID)
	return
}

type batchSpecWorkspaceResolver struct {
	store     *store.Store
	workspace *btypes.BatchSpecWorkspace
	execution *btypes.BatchSpecWorkspaceExecutionJob

	preloadedRepo *types.Repo

	repoResolverOnce sync.Once
	repoResolver     *graphqlbackend.RepositoryResolver
	repoResolverErr  error
}

var _ graphqlbackend.BatchSpecWorkspaceResolver = &batchSpecWorkspaceResolver{}

func (r *batchSpecWorkspaceResolver) ID() graphql.ID {
	return marshalBatchSpecWorkspaceID(r.workspace.ID)
}

func (r *batchSpecWorkspaceResolver) computeRepoResolver(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	r.repoResolverOnce.Do(func() {
		var repo *types.Repo
		if r.preloadedRepo != nil {
			repo = r.preloadedRepo
		} else {
			repo, r.repoResolverErr = r.store.Repos().Get(ctx, r.workspace.RepoID)
			if r.repoResolverErr != nil {
				return
			}
		}

		r.repoResolver = graphqlbackend.NewRepositoryResolver(r.store.DatabaseDB(), repo)
	})
	return r.repoResolver, r.repoResolverErr
}

func (r *batchSpecWorkspaceResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.computeRepoResolver(ctx)
}

func (r *batchSpecWorkspaceResolver) Branch(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	repo, err := r.computeRepoResolver(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewGitRefResolver(repo, r.workspace.Branch, graphqlbackend.GitObjectID(r.workspace.Commit)), nil
}

func (r *batchSpecWorkspaceResolver) Path() string {
	return r.workspace.Path
}

func (r *batchSpecWorkspaceResolver) OnlyFetchWorkspace() bool {
	return r.workspace.OnlyFetchWorkspace
}

func (r *batchSpecWorkspaceResolver) SearchResultPaths() []string {
	return r.workspace.FileMatches
}

func (r *batchSpecWorkspaceResolver) computeStepResolvers(ctx context.Context) ([]graphqlbackend.BatchSpecWorkspaceStepResolver, error) {
	var stepInfo = make(map[int]*btypes.StepInfo)
	var entryExitCode *int
	if r.execution != nil {
		entry, ok := findExecutionLogEntry(r.execution, "step.src.0")
		if ok {
			logLines := btypes.ParseJSONLogsFromOutput(entry.Out)
			stepInfo = btypes.ParseLogLines(entry, logLines)
			entryExitCode = entry.ExitCode
		}
	}

	repoResolver, err := r.computeRepoResolver(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.BatchSpecWorkspaceStepResolver, 0, len(r.workspace.Steps))
	for idx, step := range r.workspace.Steps {
		si, ok := stepInfo[idx+1]
		if !ok {
			// Step hasn't run yet.
			si = &btypes.StepInfo{}
			// But also will never run
			if entryExitCode != nil {
				si.Skipped = true
			}
		}

		// Mark all steps as skipped when a cached result was found.
		if r.CachedResultFound() {
			si.Skipped = true
		}

		// Mark all steps as skipped when a workspace is skipped.
		if r.workspace.Skipped {
			si.Skipped = true
		}

		resolver := &batchSpecWorkspaceStepResolver{
			index:    idx,
			step:     step,
			stepInfo: si,
			store:    r.store,
			repo:     repoResolver,
			baseRev:  r.workspace.Commit,
		}

		// See if we have a cache result for this step.
		if cachedResult, ok := r.workspace.StepCacheResult(idx + 1); ok {
			resolver.cachedResult = cachedResult.Value
		}

		resolvers = append(resolvers, resolver)
	}

	return resolvers, nil
}

func (r *batchSpecWorkspaceResolver) Steps(ctx context.Context) ([]graphqlbackend.BatchSpecWorkspaceStepResolver, error) {
	return r.computeStepResolvers(ctx)
}

func (r *batchSpecWorkspaceResolver) Step(ctx context.Context, args graphqlbackend.BatchSpecWorkspaceStepArgs) (graphqlbackend.BatchSpecWorkspaceStepResolver, error) {
	// Check if step exists.
	if int(args.Index) > len(r.workspace.Steps) {
		return nil, nil
	}

	resolvers, err := r.computeStepResolvers(ctx)
	if err != nil {
		return nil, err
	}
	return resolvers[args.Index-1], nil
}

func (r *batchSpecWorkspaceResolver) BatchSpec(ctx context.Context) (graphqlbackend.BatchSpecResolver, error) {
	if r.workspace.BatchSpecID == 0 {
		return nil, nil
	}
	batchSpec, err := r.store.GetBatchSpec(ctx, store.GetBatchSpecOpts{ID: r.workspace.BatchSpecID})
	if err != nil {
		return nil, err
	}
	return &batchSpecResolver{store: r.store, batchSpec: batchSpec}, nil
}

func (r *batchSpecWorkspaceResolver) Ignored() bool {
	return r.workspace.Ignored
}

func (r *batchSpecWorkspaceResolver) Unsupported() bool {
	return r.workspace.Unsupported
}

func (r *batchSpecWorkspaceResolver) CachedResultFound() bool {
	return r.workspace.CachedResultFound
}

func (r *batchSpecWorkspaceResolver) Stages() graphqlbackend.BatchSpecWorkspaceStagesResolver {
	if r.execution == nil {
		return nil
	}
	return &batchSpecWorkspaceStagesResolver{store: r.store, execution: r.execution}
}

func (r *batchSpecWorkspaceResolver) StartedAt() *graphqlbackend.DateTime {
	if r.workspace.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	if r.execution.StartedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.execution.StartedAt}
}

func (r *batchSpecWorkspaceResolver) QueuedAt() *graphqlbackend.DateTime {
	if r.workspace.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	if r.execution.CreatedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.execution.CreatedAt}
}

func (r *batchSpecWorkspaceResolver) FinishedAt() *graphqlbackend.DateTime {
	if r.workspace.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	if r.execution.FinishedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.execution.FinishedAt}
}

func (r *batchSpecWorkspaceResolver) FailureMessage() *string {
	if r.workspace.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	return r.execution.FailureMessage
}

func (r *batchSpecWorkspaceResolver) State() string {
	if r.CachedResultFound() {
		return "COMPLETED"
	}
	if r.workspace.Skipped {
		return "SKIPPED"
	}
	if r.execution == nil {
		return "PENDING"
	}
	return r.execution.State.ToGraphQL()
}

func (r *batchSpecWorkspaceResolver) ChangesetSpecs(ctx context.Context) (*[]graphqlbackend.ChangesetSpecResolver, error) {
	if r.workspace.Skipped && !r.CachedResultFound() {
		return nil, nil
	}

	if len(r.workspace.ChangesetSpecIDs) == 0 {
		none := []graphqlbackend.ChangesetSpecResolver{}
		return &none, nil
	}
	specs, _, err := r.store.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{IDs: r.workspace.ChangesetSpecIDs})
	if err != nil {
		return nil, err
	}
	var repos map[api.RepoID]*types.Repo
	repoIDs := specs.RepoIDs()
	if len(repoIDs) > 0 {
		repos, err = r.store.Repos().GetReposSetByIDs(ctx, specs.RepoIDs()...)
		if err != nil {
			return nil, err
		}
	}
	resolvers := make([]graphqlbackend.ChangesetSpecResolver, 0, len(specs))
	for _, spec := range specs {
		resolvers = append(resolvers, NewChangesetSpecResolverWithRepo(r.store, repos[spec.RepoID], spec))
	}
	return &resolvers, nil
}

func (r *batchSpecWorkspaceResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	// TODO: Cache this computation.
	resolvers, err := r.ChangesetSpecs(ctx)
	if err != nil {
		return nil, err
	}
	if resolvers == nil || len(*resolvers) == 0 {
		return nil, nil
	}
	var totalDiff graphqlbackend.DiffStat
	for _, r := range *resolvers {
		// If changeset is not visible to user, skip it.
		v, ok := r.ToVisibleChangesetSpec()
		if !ok {
			continue
		}
		desc, err := v.Description(ctx)
		if err != nil {
			return nil, err
		}
		// We only need to count "branch" changeset specs.
		d, ok := desc.ToGitBranchChangesetDescription()
		if !ok {
			continue
		}
		if diff := d.DiffStat(); diff != nil {
			totalDiff.AddDiffStat(diff)
		}
	}
	return &totalDiff, nil
}

func (r *batchSpecWorkspaceResolver) PlaceInQueue() *int32 {
	if r.execution == nil {
		return nil
	}
	if r.execution.State != btypes.BatchSpecWorkspaceExecutionJobStateQueued {
		return nil
	}

	i32 := int32(r.execution.PlaceInQueue)
	return &i32
}

func (r *batchSpecWorkspaceResolver) Executor(ctx context.Context) (*graphqlbackend.ExecutorResolver, error) {
	if r.execution == nil {
		return nil, nil
	}

	return graphqlbackend.ExecutorByHostname(ctx, r.store.DatabaseDB(), r.execution.WorkerHostname)
}

type batchSpecWorkspaceStagesResolver struct {
	store     *store.Store
	execution *btypes.BatchSpecWorkspaceExecutionJob
}

var _ graphqlbackend.BatchSpecWorkspaceStagesResolver = &batchSpecWorkspaceStagesResolver{}

func (r *batchSpecWorkspaceStagesResolver) Setup() []graphqlbackend.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix("setup.")
}

func (r *batchSpecWorkspaceStagesResolver) SrcExec() graphqlbackend.ExecutionLogEntryResolver {
	if entry, ok := findExecutionLogEntry(r.execution, "step.src.0"); ok {
		return graphqlbackend.NewExecutionLogEntryResolver(r.store.DatabaseDB(), entry)
	}

	return nil
}

func (r *batchSpecWorkspaceStagesResolver) Teardown() []graphqlbackend.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix("teardown.")
}

func (r *batchSpecWorkspaceStagesResolver) executionLogEntryResolversWithPrefix(prefix string) []graphqlbackend.ExecutionLogEntryResolver {
	var resolvers []graphqlbackend.ExecutionLogEntryResolver
	for _, entry := range r.execution.ExecutionLogs {
		if !strings.HasPrefix(entry.Key, prefix) {
			continue
		}
		r := graphqlbackend.NewExecutionLogEntryResolver(r.store.DatabaseDB(), entry)
		resolvers = append(resolvers, r)
	}

	return resolvers
}

func findExecutionLogEntry(execution *btypes.BatchSpecWorkspaceExecutionJob, key string) (workerutil.ExecutionLogEntry, bool) {
	for _, entry := range execution.ExecutionLogs {
		if entry.Key == key {
			return entry, true
		}
	}

	return workerutil.ExecutionLogEntry{}, false
}
