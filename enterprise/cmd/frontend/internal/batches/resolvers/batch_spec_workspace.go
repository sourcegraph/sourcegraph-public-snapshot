package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	gql "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const batchSpecWorkspaceIDKind = "BatchSpecWorkspace"

func marshalBatchSpecWorkspaceID(id int64) graphql.ID {
	return relay.MarshalID(batchSpecWorkspaceIDKind, id)
}

func unmarshalBatchSpecWorkspaceID(id graphql.ID) (batchSpecWorkspaceID int64, err error) {
	err = relay.UnmarshalSpec(id, &batchSpecWorkspaceID)
	return
}

func newBatchSpecWorkspaceResolver(ctx context.Context, store *store.Store, workspace *btypes.BatchSpecWorkspace, execution *btypes.BatchSpecWorkspaceExecutionJob, batchSpec *batcheslib.BatchSpec) (graphqlbackend.BatchSpecWorkspaceResolver, error) {
	repo, err := store.Repos().Get(ctx, workspace.RepoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	return newBatchSpecWorkspaceResolverWithRepo(store, workspace, execution, batchSpec, repo), nil
}

func newBatchSpecWorkspaceResolverWithRepo(store *store.Store, workspace *btypes.BatchSpecWorkspace, execution *btypes.BatchSpecWorkspaceExecutionJob, batchSpec *batcheslib.BatchSpec, repo *types.Repo) graphqlbackend.BatchSpecWorkspaceResolver {
	return &batchSpecWorkspaceResolver{
		store:        store,
		workspace:    workspace,
		execution:    execution,
		batchSpec:    batchSpec,
		repo:         repo,
		repoResolver: graphqlbackend.NewRepositoryResolver(store.DatabaseDB(), repo),
	}
}

type batchSpecWorkspaceResolver struct {
	store     *store.Store
	workspace *btypes.BatchSpecWorkspace
	execution *btypes.BatchSpecWorkspaceExecutionJob
	batchSpec *batcheslib.BatchSpec

	repo         *types.Repo
	repoResolver *graphqlbackend.RepositoryResolver

	changesetSpecs     []*btypes.ChangesetSpec
	changesetSpecsOnce sync.Once
	changesetSpecsErr  error
}

var _ graphqlbackend.BatchSpecWorkspaceResolver = &batchSpecWorkspaceResolver{}

func (r *batchSpecWorkspaceResolver) ToHiddenBatchSpecWorkspace() (graphqlbackend.HiddenBatchSpecWorkspaceResolver, bool) {
	if r.repo != nil {
		return nil, false
	}

	return r, true
}

func (r *batchSpecWorkspaceResolver) ToVisibleBatchSpecWorkspace() (graphqlbackend.VisibleBatchSpecWorkspaceResolver, bool) {
	if r.repo == nil {
		return nil, false
	}

	return r, true
}

func (r *batchSpecWorkspaceResolver) ID() graphql.ID {
	return marshalBatchSpecWorkspaceID(r.workspace.ID)
}

func (r *batchSpecWorkspaceResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	if _, ok := r.ToHiddenBatchSpecWorkspace(); ok {
		return nil, nil
	}

	return r.repoResolver, nil
}

func (r *batchSpecWorkspaceResolver) Branch(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	if _, ok := r.ToHiddenBatchSpecWorkspace(); ok {
		return nil, nil
	}

	return graphqlbackend.NewGitRefResolver(r.repoResolver, r.workspace.Branch, graphqlbackend.GitObjectID(r.workspace.Commit)), nil
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

func (r *batchSpecWorkspaceResolver) computeStepResolvers() ([]graphqlbackend.BatchSpecWorkspaceStepResolver, error) {
	if _, ok := r.ToHiddenBatchSpecWorkspace(); ok {
		return nil, nil
	}

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

	skippedSteps, err := batcheslib.SkippedStepsForRepo(r.batchSpec, r.repoResolver.Name(), r.workspace.FileMatches)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.BatchSpecWorkspaceStepResolver, 0, len(r.batchSpec.Steps))
	for idx, step := range r.batchSpec.Steps {
		si, ok := stepInfo[idx+1]
		if !ok {
			// Step hasn't run yet.
			si = &btypes.StepInfo{}
			// ..but also will never run.
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

		// If we have marked the step as to-be-skipped, we have to translate
		// that here into the workspace step info.
		if _, ok := skippedSteps[int32(idx)]; ok {
			si.Skipped = true
		}

		resolver := &batchSpecWorkspaceStepResolver{
			index:    idx,
			step:     step,
			stepInfo: si,
			store:    r.store,
			repo:     r.repoResolver,
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

func (r *batchSpecWorkspaceResolver) Steps() ([]graphqlbackend.BatchSpecWorkspaceStepResolver, error) {
	return r.computeStepResolvers()
}

func (r *batchSpecWorkspaceResolver) Step(args graphqlbackend.BatchSpecWorkspaceStepArgs) (graphqlbackend.BatchSpecWorkspaceStepResolver, error) {
	// Check if step exists.
	if int(args.Index) > len(r.batchSpec.Steps) {
		return nil, nil
	}

	resolvers, err := r.computeStepResolvers()
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

func (r *batchSpecWorkspaceResolver) ChangesetSpecs(ctx context.Context) (*[]graphqlbackend.VisibleChangesetSpecResolver, error) {
	// If this is a hidden resolver, we don't return changeset specs, since we only return visible changeset spec resolvers here.
	if _, ok := r.ToHiddenBatchSpecWorkspace(); ok {
		return nil, nil
	}

	// If the workspace has been skipped and no cached result was found, there are definitely no changeset specs.
	if r.workspace.Skipped && !r.CachedResultFound() {
		return nil, nil
	}

	specs, err := r.computeChangesetSpecs(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.VisibleChangesetSpecResolver, 0, len(specs))
	for _, spec := range specs {
		resolvers = append(resolvers, NewChangesetSpecResolverWithRepo(r.store, r.repo, spec))
	}

	return &resolvers, nil
}

func (r *batchSpecWorkspaceResolver) computeChangesetSpecs(ctx context.Context) ([]*btypes.ChangesetSpec, error) {
	r.changesetSpecsOnce.Do(func() {
		if len(r.workspace.ChangesetSpecIDs) == 0 {
			r.changesetSpecs = []*btypes.ChangesetSpec{}
			return
		}

		specs, _, err := r.store.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{IDs: r.workspace.ChangesetSpecIDs})
		if err != nil {
			r.changesetSpecsErr = err
			return
		}

		repoIDs := specs.RepoIDs()
		if len(repoIDs) > 1 {
			r.changesetSpecsErr = errors.New("changeset specs associated with workspace they don't belong to")
			return
		}
		if len(repoIDs) == 1 && repoIDs[0] != r.workspace.RepoID {
			r.changesetSpecsErr = errors.New("changeset specs associated with workspace they don't belong to")
			return
		}

		r.changesetSpecs = specs
	})

	return r.changesetSpecs, r.changesetSpecsErr
}

func (r *batchSpecWorkspaceResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	specs, err := r.computeChangesetSpecs(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.VisibleChangesetSpecResolver, 0, len(specs))
	for _, spec := range specs {
		resolvers = append(resolvers, NewChangesetSpecResolverWithRepo(r.store, r.repo, spec))
	}

	if len(resolvers) == 0 {
		return nil, nil
	}
	var totalDiff graphqlbackend.DiffStat
	for _, r := range resolvers {
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

func (r *batchSpecWorkspaceResolver) Executor(ctx context.Context) (*gql.ExecutorResolver, error) {
	if r.execution == nil {
		return nil, nil
	}

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.store.DatabaseDB()); err != nil {
		if err != backend.ErrMustBeSiteAdmin {
			return nil, err
		}
		return nil, nil
	}

	executor, err := gql.New(r.store.DatabaseDB()).ExecutorByHostname(ctx, r.execution.WorkerHostname)
	if err != nil {
		return nil, err
	}

	return executor, nil
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
