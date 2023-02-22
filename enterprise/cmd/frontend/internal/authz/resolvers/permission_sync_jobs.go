package resolvers

import (
	"context"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func NewPermissionSyncJobsResolver(db database.DB, args graphqlbackend.ListPermissionSyncJobsArgs) (*graphqlutil.ConnectionResolver[graphqlbackend.PermissionSyncJobResolver], error) {
	store := &permissionSyncJobConnectionStore{
		db:   db,
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
		syncSubject, err := s.resolveSubject(ctx, job)
		if err != nil {
			return nil, err
		}
		resolvers = append(resolvers, &permissionSyncJobResolver{
			db:          s.db,
			job:         job,
			syncSubject: syncSubject,
		})
	}
	return resolvers, nil
}

func (s *permissionSyncJobConnectionStore) resolveSubject(ctx context.Context, job *database.PermissionSyncJob) (graphqlbackend.PermissionSyncJobSubject, error) {
	var repoResolver *graphqlbackend.RepositoryResolver
	var userResolver *graphqlbackend.UserResolver

	if job.UserID > 0 {
		user, err := s.db.Users().GetByID(ctx, int32(job.UserID))
		if err != nil {
			return nil, err
		}
		userResolver = graphqlbackend.NewUserResolver(s.db, user)
	} else {
		repo, err := s.db.Repos().Get(ctx, api.RepoID(job.RepositoryID))
		if err != nil {
			return nil, err
		}
		repoResolver = graphqlbackend.NewRepositoryResolver(s.db, gitserver.NewClient(), repo)
	}

	return &subject{
		repo: repoResolver,
		user: userResolver,
	}, nil
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
	db          database.DB
	job         *database.PermissionSyncJob
	syncSubject graphqlbackend.PermissionSyncJobSubject
}

func (p *permissionSyncJobResolver) ID() graphql.ID {
	return marshalPermissionSyncJobID(p.job.ID)
}

func (p *permissionSyncJobResolver) State() string {
	return p.job.State.ToGraphQL()
}

func (p *permissionSyncJobResolver) FailureMessage() *string {
	return p.job.FailureMessage
}

func (p *permissionSyncJobResolver) Reason() graphqlbackend.PermissionSyncJobReasonResolver {
	reason := p.job.Reason
	return permissionSyncJobReasonResolver{group: reason.ResolveGroup(), message: string(reason)}
}

func (p *permissionSyncJobResolver) CancellationReason() *string {
	return p.job.CancellationReason
}

func (p *permissionSyncJobResolver) TriggeredByUser(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	userID := p.job.TriggeredByUserID
	if userID == 0 {
		return nil, nil
	}
	return graphqlbackend.UserByIDInt32(ctx, p.db, userID)
}

func (p *permissionSyncJobResolver) QueuedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: p.job.QueuedAt}
}

func (p *permissionSyncJobResolver) StartedAt() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.StartedAt)
}

func (p *permissionSyncJobResolver) FinishedAt() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.FinishedAt)
}

func (p *permissionSyncJobResolver) ProcessAfter() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.ProcessAfter)
}

func (p *permissionSyncJobResolver) RanForMs() *int32 {
	var ranFor int32
	if !p.job.FinishedAt.IsZero() {
		// Job runtime in ms shouldn't take more than a 32-bit int value.
		ranFor = int32(p.job.FinishedAt.Sub(p.job.StartedAt).Milliseconds())
	}
	return &ranFor
}

func (p *permissionSyncJobResolver) NumResets() *int32 {
	return intToInt32Ptr(p.job.NumResets)
}

func (p *permissionSyncJobResolver) NumFailures() *int32 {
	return intToInt32Ptr(p.job.NumFailures)
}

func (p *permissionSyncJobResolver) LastHeartbeatAt() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.LastHeartbeatAt)
}

func (p *permissionSyncJobResolver) WorkerHostname() string {
	return p.job.WorkerHostname
}

func (p *permissionSyncJobResolver) Cancel() bool {
	return p.job.Cancel
}

func (p *permissionSyncJobResolver) Subject() graphqlbackend.PermissionSyncJobSubject {
	return p.syncSubject
}

func (p *permissionSyncJobResolver) Priority() string {
	return p.job.Priority.ToString()
}

func (p *permissionSyncJobResolver) NoPerms() bool {
	return p.job.NoPerms
}

func (p *permissionSyncJobResolver) InvalidateCaches() bool {
	return p.job.InvalidateCaches
}

func (p *permissionSyncJobResolver) PermissionsAdded() int32 {
	return int32(p.job.PermissionsAdded)
}

func (p *permissionSyncJobResolver) PermissionsRemoved() int32 {
	return int32(p.job.PermissionsRemoved)
}

func (p *permissionSyncJobResolver) PermissionsFound() int32 {
	return int32(p.job.PermissionsFound)
}

func (p *permissionSyncJobResolver) CodeHostStates() []graphqlbackend.CodeHostStateResolver {
	resolvers := make([]graphqlbackend.CodeHostStateResolver, 0, len(p.job.CodeHostStates))
	for _, state := range p.job.CodeHostStates {
		resolvers = append(resolvers, codeHostStateResolver{state: state})
	}
	return resolvers
}

type codeHostStateResolver struct {
	state database.PermissionSyncCodeHostState
}

func (c codeHostStateResolver) ProviderID() string {
	return c.state.ProviderID
}

func (c codeHostStateResolver) ProviderType() string {
	return c.state.ProviderType
}

func (c codeHostStateResolver) Status() string {
	return c.state.Status
}

func (c codeHostStateResolver) Message() string {
	return c.state.Message
}

type permissionSyncJobReasonResolver struct {
	group   database.PermissionSyncJobReasonGroup
	message string
}

func (p permissionSyncJobReasonResolver) Group() string {
	return string(p.group)
}
func (p permissionSyncJobReasonResolver) Message() string {
	return p.message
}

type subject struct {
	repo *graphqlbackend.RepositoryResolver
	user *graphqlbackend.UserResolver
}

func (s subject) ToRepository() (*graphqlbackend.RepositoryResolver, bool) {
	return s.repo, s.repo != nil
}

func (s subject) ToUser() (*graphqlbackend.UserResolver, bool) {
	return s.user, s.user != nil
}

func marshalPermissionSyncJobID(id int) graphql.ID {
	return relay.MarshalID("PermissionSyncJob", id)
}

func unmarshalPermissionSyncJobID(id graphql.ID) (jobID int, err error) {
	err = relay.UnmarshalSpec(id, &jobID)
	return
}

func intToInt32Ptr(value int) *int32 {
	int32Value := int32(value)
	return &int32Value
}
