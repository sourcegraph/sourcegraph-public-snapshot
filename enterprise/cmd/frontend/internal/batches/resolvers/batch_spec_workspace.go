package resolvers

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
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

	repoOnce sync.Once
	repo     *graphqlbackend.RepositoryResolver
	repoErr  error
}

var _ graphqlbackend.BatchSpecWorkspaceResolver = &batchSpecWorkspaceResolver{}

func (r *batchSpecWorkspaceResolver) ID() graphql.ID {
	return marshalBatchSpecWorkspaceID(r.workspace.ID)
}

func (r *batchSpecWorkspaceResolver) computeRepo(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	r.repoOnce.Do(func() {
		var repo *types.Repo
		repo, r.repoErr = r.store.Repos().Get(ctx, r.workspace.RepoID)
		r.repo = graphqlbackend.NewRepositoryResolver(r.store.DB(), repo)
	})
	return r.repo, r.repoErr
}

func (r *batchSpecWorkspaceResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.computeRepo(ctx)
}

func (r *batchSpecWorkspaceResolver) Branch(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	repo, err := r.computeRepo(ctx)
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

func (r *batchSpecWorkspaceResolver) Steps(ctx context.Context) ([]graphqlbackend.BatchSpecWorkspaceStepResolver, error) {
	logLines := []batcheslib.LogEvent{}
	if r.execution != nil {
		entry, ok := findExecutionLogEntry(r.execution, "step.src.0")
		if ok {
			for _, line := range strings.Split(entry.Out, "\n") {
				if !strings.HasPrefix(line, "stdout: ") {
					continue
				}
				line = line[len("stdout: "):]
				var parsed batcheslib.LogEvent
				err := json.Unmarshal([]byte(line), &parsed)
				if err != nil {
					return nil, err
				}
				logLines = append(logLines, parsed)
			}
		}
	}

	repo, err := r.computeRepo(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.BatchSpecWorkspaceStepResolver, 0, len(r.workspace.Steps))
	for idx, step := range r.workspace.Steps {
		resolvers = append(resolvers, &batchSpecWorkspaceStepResolver{index: idx, step: step, logLines: logLines, store: r.store, repo: repo, baseRev: r.workspace.Commit})
	}

	return resolvers, nil
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
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecWorkspaceResolver) CachedResultFound() bool {
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecWorkspaceResolver) Stages() graphqlbackend.BatchSpecWorkspaceStagesResolver {
	if r.execution == nil {
		return nil
	}
	return &batchSpecWorkspaceStagesResolver{store: r.store, execution: r.execution}
}

func (r *batchSpecWorkspaceResolver) StartedAt() *graphqlbackend.DateTime {
	if r.execution == nil {
		return nil
	}
	if r.execution.StartedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.execution.StartedAt}
}

func (r *batchSpecWorkspaceResolver) FinishedAt() *graphqlbackend.DateTime {
	if r.execution == nil {
		return nil
	}
	if r.execution.FinishedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.execution.FinishedAt}
}

func (r *batchSpecWorkspaceResolver) FailureMessage() *string {
	if r.execution == nil {
		return nil
	}
	return r.execution.FailureMessage
}

func (r *batchSpecWorkspaceResolver) State() string {
	if r.execution == nil {
		return "QUEUED"
	}
	return r.execution.State.ToGraphQL()
}

func (r *batchSpecWorkspaceResolver) ChangesetSpecs(ctx context.Context) (*[]graphqlbackend.ChangesetSpecResolver, error) {
	if len(r.workspace.ChangesetSpecIDs) == 0 {
		none := []graphqlbackend.ChangesetSpecResolver{}
		return &none, nil
	}
	specs, _, err := r.store.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{IDs: r.workspace.ChangesetSpecIDs})
	if err != nil {
		return nil, err
	}
	repos, err := r.store.Repos().GetReposSetByIDs(ctx, specs.RepoIDs()...)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.ChangesetSpecResolver, 0, len(specs))
	for _, spec := range specs {
		resolvers = append(resolvers, NewChangesetSpecResolverWithRepo(r.store, repos[spec.RepoID], spec))
	}
	return &resolvers, nil
}

func (r *batchSpecWorkspaceResolver) PlaceInQueue() *int32 {
	// TODO(ssbc): not implemented
	return nil
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
		return graphqlbackend.NewExecutionLogEntryResolver(r.store.DB(), entry)
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
		r := graphqlbackend.NewExecutionLogEntryResolver(r.store.DB(), entry)
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
