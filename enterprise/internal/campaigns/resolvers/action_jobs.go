package resolvers

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

const actionJobIDKind = "ActionJob"

func marshalActionJobID(id int64) graphql.ID {
	return relay.MarshalID(actionJobIDKind, id)
}

func unmarshalActionJobID(id graphql.ID) (actionJobID int64, err error) {
	err = relay.UnmarshalSpec(id, &actionJobID)
	return
}

type actionJobResolver struct {
	store *ee.Store
	job   campaigns.ActionJob
	// todo: pass in from parent if known to avoid duplicate sql query
	actionExecution *campaigns.ActionExecution

	repoOnce sync.Once
	repo     *graphqlbackend.RepositoryResolver
	repoErr  error
}

func (r *Resolver) ActionJobByID(ctx context.Context, id graphql.ID) (graphqlbackend.ActionJobResolver, error) {
	// todo: permissions

	dbId, err := unmarshalActionJobID(id)
	if err != nil {
		return nil, err
	}

	actionJob, err := r.store.ActionJobByID(ctx, ee.ActionJobByIDOpts{
		ID: dbId,
	})
	if err != nil {
		return nil, err
	}
	if actionJob.ID == 0 {
		return nil, nil
	}
	return &actionJobResolver{store: r.store, job: *actionJob}, nil
}

func (r *actionJobResolver) ID() graphql.ID {
	return marshalActionJobID(r.job.ID)
}

func (r *actionJobResolver) Definition(ctx context.Context) (graphqlbackend.ActionDefinitionResolver, error) {
	if r.actionExecution != nil {
		return &actionDefinitionResolver{steps: r.actionExecution.Steps, envStr: *r.actionExecution.EnvStr}, nil
	}
	actionExecution, err := r.store.ActionExecutionByID(ctx, ee.ActionExecutionByIDOpts{ID: r.job.ExecutionID})
	if err != nil {
		return nil, err
	}
	return &actionDefinitionResolver{steps: actionExecution.Steps, envStr: *actionExecution.EnvStr}, nil
}

func (r *actionJobResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.computeRepo(ctx)
}

func (r *actionJobResolver) BaseRevision() string {
	return r.job.BaseRevision
}

func (r *actionJobResolver) State() campaigns.ActionJobState {
	return r.job.State
}

func (r *actionJobResolver) Agent() graphqlbackend.AgentResolver {
	if r.job.State == campaigns.ActionJobStatePending {
		return nil
	}
	return &agentResolver{}
}

func (r *actionJobResolver) BaseRepository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.Repository(ctx)
}
func (r *actionJobResolver) Diff() graphqlbackend.ActionJobResolver {
	if r.job.Patch != nil {
		return r
	}
	return nil
}
func (r *actionJobResolver) FileDiffs(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.PreviewFileDiffConnection, error) {
	repo, err := r.computeRepo(ctx)
	if err != nil {
		return nil, err
	}
	commit, err := repo.Commit(ctx, &graphqlbackend.RepositoryCommitArgs{Rev: r.BaseRevision()})
	return &previewFileDiffConnectionResolver{
		diff:   r.job.Patch,
		commit: commit,
		first:  args.First,
	}, nil
}

func (r *actionJobResolver) ExecutionStart() *graphqlbackend.DateTime {
	if r.job.ExecutionStartAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.job.ExecutionStartAt}
}

func (r *actionJobResolver) ExecutionEnd() *graphqlbackend.DateTime {
	if r.job.ExecutionEndAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.job.ExecutionEndAt}
}

func (r *actionJobResolver) Log() *string {
	return r.job.Log
}

func (r *actionJobResolver) computeRepo(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	r.repoOnce.Do(func() {
		r.repo, r.repoErr = graphqlbackend.RepositoryByIDInt32(ctx, api.RepoID(r.job.RepoID))
	})
	return r.repo, r.repoErr
}
