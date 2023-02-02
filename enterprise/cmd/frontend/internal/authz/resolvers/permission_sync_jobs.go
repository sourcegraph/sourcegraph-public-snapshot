package resolvers

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var _ graphqlbackend.PermissionSyncJobsResolver = &permissionSyncJobsResolver{}

type permissionSyncJobsResolver struct {
	db database.DB
}

func NewPermissionSyncJobsResolver(db database.DB) graphqlbackend.PermissionSyncJobsResolver {
	return &permissionSyncJobsResolver{db: db}
}

func (r *permissionSyncJobsResolver) PermissionSyncJobs(ctx context.Context, args *graphqlbackend.ListPermissionSyncJobsArgs) (graphqlbackend.PermissionSyncJobConnectionResolver, error) {
	return &permissionSyncJobConnectionResolver{db: r.db}, nil
}

type permissionSyncJobConnectionResolver struct {
	db   database.DB
	once sync.Once
	jobs []*database.PermissionSyncJob
	err  error
}

func (p *permissionSyncJobConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.PermissionSyncJobResolver, error) {
	jobs, err := p.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.PermissionSyncJobResolver, 0, len(jobs))
	for _, job := range jobs {
		resolvers = append(resolvers, &permissionSyncJobResolver{
			db:  p.db,
			job: job,
		})
	}
	return resolvers, nil
}

func (p *permissionSyncJobConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	if p.err != nil {
		return 0, p.err
	}
	return int32(len(p.jobs)), nil
}

func (p *permissionSyncJobConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return nil, nil
}

func (p *permissionSyncJobConnectionResolver) compute(ctx context.Context) ([]*database.PermissionSyncJob, error) {
	p.once.Do(func() {
		p.jobs, p.err = p.db.PermissionSyncJobs().List(ctx, database.ListPermissionSyncJobOpts{})
	})
	return p.jobs, p.err
}

type permissionSyncJobResolver struct {
	db  database.DB
	job *database.PermissionSyncJob
}

func (p *permissionSyncJobResolver) ID() graphql.ID {
	return marshalPermissionSyncJobID(p.job.ID)
}

func marshalPermissionSyncJobID(id int) graphql.ID {
	return relay.MarshalID("PermissionSyncJob", id)
}
