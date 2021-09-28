package resolvers

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

type batchSpecWorkspaceResolver struct {
	store *store.Store
	node  *service.RepoWorkspace
}

var _ graphqlbackend.BatchSpecWorkspaceResolver = &batchSpecWorkspaceResolver{}

func (r *batchSpecWorkspaceResolver) ID() graphql.ID {
	// TODO(ssbc): not implemented
	return graphql.ID("not implemented")
}
func (r *batchSpecWorkspaceResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.NewRepositoryResolver(r.store.DB(), r.node.Repo), nil
}

func (r *batchSpecWorkspaceResolver) Branch(ctx context.Context) (*graphqlbackend.GitRefResolver, error) {
	repo, _ := r.Repository(ctx)
	return graphqlbackend.NewGitRefResolver(repo, r.node.Branch, graphqlbackend.GitObjectID(r.node.Commit)), nil
}

func (r *batchSpecWorkspaceResolver) Path() string {
	return r.node.Path
}

func (r *batchSpecWorkspaceResolver) OnlyFetchWorkspace() bool {
	return r.node.OnlyFetchWorkspace
}

func (r *batchSpecWorkspaceResolver) SearchResultPaths() []string {
	return r.node.FileMatches
}

func (r *batchSpecWorkspaceResolver) Steps() []graphqlbackend.BatchSpecWorkspaceStepResolver {
	resolvers := make([]graphqlbackend.BatchSpecWorkspaceStepResolver, 0, len(r.node.Steps))
	for _, step := range r.node.Steps {
		resolvers = append(resolvers, &batchSpecWorkspaceStepResolver{step})
	}
	return resolvers
}

func (r *batchSpecWorkspaceResolver) BatchSpec(context.Context) (graphqlbackend.BatchSpecResolver, error) {
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented")
}

func (r *batchSpecWorkspaceResolver) Ignored() bool {
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecWorkspaceResolver) CachedResultFound() bool {
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecWorkspaceResolver) Stages() (graphqlbackend.BatchSpecWorkspaceStagesResolver, error) {
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *batchSpecWorkspaceResolver) StartedAt() *graphqlbackend.DateTime {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceResolver) FinishedAt() *graphqlbackend.DateTime {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceResolver) FailureMessage() *string {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceResolver) State() string {
	// TODO(ssbc): not implemented
	return "FAILED"
}

func (r *batchSpecWorkspaceResolver) ChangesetSpecs() *[]graphqlbackend.ChangesetSpecResolver {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceResolver) PlaceInQueue() *int32 {
	// TODO(ssbc): not implemented
	var p int32 = 9999
	return &p
}

// -----------------------------------------------

type batchSpecWorkspaceStepResolver struct {
	step batcheslib.Step
}

func (r *batchSpecWorkspaceStepResolver) Run() string {
	return r.step.Run
}

func (r *batchSpecWorkspaceStepResolver) Container() string {
	return r.step.Container
}

func (r *batchSpecWorkspaceStepResolver) CachedResultFound() bool {
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecWorkspaceStepResolver) Skipped() bool {
	// TODO(ssbc): not implemented
	return false
}

func (r *batchSpecWorkspaceStepResolver) OutputLines(ctx context.Context, args *graphqlbackend.BatchSpecWorkspaceStepOutputLinesArgs) (*[]string, error) {
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}

func (r *batchSpecWorkspaceStepResolver) StartedAt() *graphqlbackend.DateTime {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceStepResolver) FinishedAt() *graphqlbackend.DateTime {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceStepResolver) ExitCode() *int32 {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceStepResolver) Environment() []graphqlbackend.BatchSpecWorkspaceEnvironmentVariableResolver {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceStepResolver) OutputVariables() *[]graphqlbackend.BatchSpecWorkspaceOutputVariableResolver {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceStepResolver) DiffStat() *graphqlbackend.DiffStat {
	// TODO(ssbc): not implemented
	return nil
}

func (r *batchSpecWorkspaceStepResolver) Diff(ctx context.Context) (graphqlbackend.PreviewRepositoryComparisonResolver, error) {
	// TODO(ssbc): not implemented
	return nil, errors.New("not implemented yet")
}
