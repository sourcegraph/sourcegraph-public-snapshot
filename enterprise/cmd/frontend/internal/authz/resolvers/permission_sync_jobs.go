package resolvers

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/executor"
)

var _ graphqlbackend.PermissionSyncJobsResolver = &permissionSyncJobsResolver{}

type permissionSyncJobsResolver struct {
	db   database.DB
	once sync.Once
	jobs []*database.PermissionSyncJob
	err  error
}

func NewPermissionSyncJobsResolver(db database.DB) graphqlbackend.PermissionSyncJobsResolver {
	return &permissionSyncJobsResolver{db: db}
}

func (r *permissionSyncJobsResolver) PermissionSyncJobs(_ context.Context, args graphqlbackend.ListPermissionSyncJobsArgs) (*graphqlutil.ConnectionResolver[graphqlbackend.PermissionSyncJobResolver], error) {
	store := &permissionSyncJobConnectionStore{
		db:   r.db,
		args: args,
	}
	return graphqlutil.NewConnectionResolver[graphqlbackend.PermissionSyncJobResolver](store, &args.ConnectionResolverArgs, nil)
}

type permissionSyncJobConnectionStore struct {
	db   database.DB
	args graphqlbackend.ListPermissionSyncJobsArgs
}

func (s *permissionSyncJobConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	count, err := s.db.PermissionSyncJobs().Count(ctx)
	if err != nil {
		return nil, err
	}
	total := int32(count)
	return &total, nil
}

func (s *permissionSyncJobConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]graphqlbackend.PermissionSyncJobResolver, error) {
	jobs, err := s.db.PermissionSyncJobs().List(ctx, database.ListPermissionSyncJobOpts{PaginationArgs: args})
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.PermissionSyncJobResolver, 0, len(jobs))
	for _, job := range jobs {
		resolvers = append(resolvers, &permissionSyncJobResolver{
			db:  s.db,
			job: job,
		})
	}
	return resolvers, nil
}

func (s *permissionSyncJobConnectionStore) MarshalCursor(node graphqlbackend.PermissionSyncJobResolver, _ database.OrderBy) (*string, error) {
	id, err := unmarshalPermissionSyncJobID(node.ID())
	if err != nil {
		return nil, err
	}
	cursor := strconv.Itoa(id)
	return &cursor, nil
}

func (s *permissionSyncJobConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) (*string, error) {
	return &cursor, nil
}

type permissionSyncJobResolver struct {
	db  database.DB
	job *database.PermissionSyncJob
}

func (p *permissionSyncJobResolver) ID() graphql.ID {
	return marshalPermissionSyncJobID(p.job.ID)
}

func (p *permissionSyncJobResolver) State() string {
	return p.job.State
}

func (p *permissionSyncJobResolver) FailureMessage() *string {
	return p.job.FailureMessage
}

func (p *permissionSyncJobResolver) Reason() database.PermissionSyncJobReason {
	return p.job.Reason
}

func (p *permissionSyncJobResolver) CancellationReason() string {
	return p.job.CancellationReason
}

func (p *permissionSyncJobResolver) TriggeredByUserID() int32 {
	return p.job.TriggeredByUserID
}

func (p *permissionSyncJobResolver) QueuedAt() time.Time {
	return p.job.QueuedAt
}

func (p *permissionSyncJobResolver) StartedAt() time.Time {
	return p.job.StartedAt
}

func (p *permissionSyncJobResolver) FinishedAt() time.Time {
	return p.job.FinishedAt
}

func (p *permissionSyncJobResolver) ProcessAfter() time.Time {
	return p.job.ProcessAfter
}

func (p *permissionSyncJobResolver) NumResets() int {
	return p.job.NumResets
}

func (p *permissionSyncJobResolver) NumFailures() int {
	return p.job.NumFailures
}

func (p *permissionSyncJobResolver) LastHeartbeatAt() time.Time {
	return p.job.LastHeartbeatAt
}

func (p *permissionSyncJobResolver) ExecutionLogs() []executor.ExecutionLogEntry {
	return p.job.ExecutionLogs
}

func (p *permissionSyncJobResolver) WorkerHostname() string {
	return p.job.WorkerHostname
}

func (p *permissionSyncJobResolver) Cancel() bool {
	return p.job.Cancel
}

func (p *permissionSyncJobResolver) RepositoryID() graphql.ID {
	return graphqlbackend.MarshalRepositoryID(api.RepoID(p.job.RepositoryID))
}

func (p *permissionSyncJobResolver) UserID() graphql.ID {
	return graphqlbackend.MarshalUserID(int32(p.job.UserID))
}

func (p *permissionSyncJobResolver) Priority() database.PermissionSyncJobPriority {
	return p.job.Priority
}

func (p *permissionSyncJobResolver) NoPerms() bool {
	return p.job.NoPerms
}

func (p *permissionSyncJobResolver) InvalidateCaches() bool {
	return p.job.InvalidateCaches
}

func (p *permissionSyncJobResolver) PermissionsAdded() int {
	return p.job.PermissionsAdded
}

func (p *permissionSyncJobResolver) PermissionsRemoved() int {
	return p.job.PermissionsRemoved
}

func (p *permissionSyncJobResolver) PermissionsFound() int {
	return p.job.PermissionsFound
}

func marshalPermissionSyncJobID(id int) graphql.ID {
	return relay.MarshalID("PermissionSyncJob", id)
}

func unmarshalPermissionSyncJobID(id graphql.ID) (jobID int, err error) {
	err = relay.UnmarshalSpec(id, &jobID)
	return
}
