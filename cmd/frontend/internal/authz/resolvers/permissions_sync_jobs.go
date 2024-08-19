package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const permissionsSyncJobIDKind = "PermissionsSyncJob"

func NewPermissionsSyncJobsResolver(logger log.Logger, db database.DB, args graphqlbackend.ListPermissionsSyncJobsArgs) (*gqlutil.ConnectionResolver[graphqlbackend.PermissionsSyncJobResolver], error) {
	store := &permissionsSyncJobConnectionStore{
		logger: logger.Scoped("permissions_sync_jobs_resolver"),
		db:     db,
		args:   args,
	}

	if args.UserID != nil && args.RepoID != nil {
		return nil, errors.New("please provide either userID or repoID, but not both.")
	}

	return gqlutil.NewConnectionResolver[graphqlbackend.PermissionsSyncJobResolver](
		store, &args.ConnectionResolverArgs, nil)
}

type permissionsSyncJobConnectionStore struct {
	db     database.DB
	args   graphqlbackend.ListPermissionsSyncJobsArgs
	logger log.Logger
}

func (s *permissionsSyncJobConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.PermissionSyncJobs().Count(ctx, s.getListArgs(nil))
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (s *permissionsSyncJobConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]graphqlbackend.PermissionsSyncJobResolver, error) {
	jobs, err := s.db.PermissionSyncJobs().List(ctx, s.getListArgs(args))
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.PermissionsSyncJobResolver, 0, len(jobs))
	var errs []error
	for _, job := range jobs {
		syncSubject, err := s.resolveSubject(ctx, job)
		if err != nil {
			if !errcode.IsNotFound(err) {
				// We don't surface errors for deleted repositories or users. It's frequently the case that asynchronous
				// cleanup jobs will delete repositories and users that are referenced by permissions sync jobs.
				errs = append(errs, errors.Wrapf(err, "resolving permissions sync job subject for job %d", job.ID))
			}

			continue
		}

		resolvers = append(resolvers, &permissionsSyncJobResolver{
			db:          s.db,
			job:         job,
			syncSubject: syncSubject,
		})
	}

	if len(errs) > 0 {
		s.logger.Warn("Failed to resolve permissions sync job subjects", log.Error(errors.Append(nil, errs...)))

		tr := trace.FromContext(ctx)
		for _, e := range errs {
			tr.RecordError(e)
		}
	}

	return resolvers, nil
}

func (s *permissionsSyncJobConnectionStore) resolveSubject(ctx context.Context, job *database.PermissionSyncJob) (graphqlbackend.PermissionsSyncJobSubject, error) {
	var repoResolver *graphqlbackend.RepositoryResolver
	var userResolver *graphqlbackend.UserResolver

	if job.UserID > 0 {
		user, err := s.db.Users().GetByID(ctx, int32(job.UserID))
		if err != nil {
			return nil, err
		}
		userResolver = graphqlbackend.NewUserResolver(ctx, s.db, user)
	} else {
		repo, err := s.db.Repos().Get(ctx, api.RepoID(job.RepositoryID))
		if err != nil {
			return nil, err
		}
		repoResolver = graphqlbackend.NewRepositoryResolver(s.db, gitserver.NewClient("graphql.authz.syncjobs"), repo)
	}

	return &subject{
		repo: repoResolver,
		user: userResolver,
	}, nil
}

func (s *permissionsSyncJobConnectionStore) MarshalCursor(node graphqlbackend.PermissionsSyncJobResolver, _ database.OrderBy) (*string, error) {
	cur := string(node.ID())
	return &cur, nil
}

func (s *permissionsSyncJobConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	id, err := unmarshalPermissionsSyncJobID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	return []any{id}, nil
}

func (s *permissionsSyncJobConnectionStore) getListArgs(pageArgs *database.PaginationArgs) database.ListPermissionSyncJobOpts {
	opts := database.ListPermissionSyncJobOpts{WithPlaceInQueue: true}
	if pageArgs != nil {
		opts.PaginationArgs = pageArgs
	}
	if s.args.ReasonGroup != nil {
		opts.ReasonGroup = *s.args.ReasonGroup
	}
	if s.args.State != nil {
		opts.State = *s.args.State
	}
	if s.args.Partial != nil {
		opts.PartialSuccess = *s.args.Partial
	}
	if s.args.UserID != nil {
		if userID, err := graphqlbackend.UnmarshalUserID(*s.args.UserID); err == nil {
			opts.UserID = int(userID)
		}
	}
	if s.args.RepoID != nil {
		if repoID, err := graphqlbackend.UnmarshalRepositoryID(*s.args.RepoID); err == nil {
			opts.RepoID = int(repoID)
		}
	}
	// First, we check for search type, because it can exist without search query,
	// but not vice versa.
	if s.args.SearchType != nil {
		opts.SearchType = *s.args.SearchType
		if s.args.Query != nil {
			opts.Query = *s.args.Query
		}
	}
	return opts
}

type permissionsSyncJobResolver struct {
	db          database.DB
	job         *database.PermissionSyncJob
	syncSubject graphqlbackend.PermissionsSyncJobSubject
}

func (p *permissionsSyncJobResolver) ID() graphql.ID {
	return marshalPermissionsSyncJobID(p.job.ID)
}

func (p *permissionsSyncJobResolver) State() string {
	return p.job.State.ToGraphQL()
}

func (p *permissionsSyncJobResolver) FailureMessage() *string {
	return p.job.FailureMessage
}

func (p *permissionsSyncJobResolver) Reason() graphqlbackend.PermissionsSyncJobReasonResolver {
	reason := p.job.Reason
	return permissionSyncJobReasonResolver{group: reason.ResolveGroup(), reason: reason}
}

func (p *permissionsSyncJobResolver) CancellationReason() *string {
	return p.job.CancellationReason
}

func (p *permissionsSyncJobResolver) TriggeredByUser(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	userID := p.job.TriggeredByUserID
	if userID == 0 {
		return nil, nil
	}
	return graphqlbackend.UserByIDInt32(ctx, p.db, userID)
}

func (p *permissionsSyncJobResolver) QueuedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: p.job.QueuedAt}
}

func (p *permissionsSyncJobResolver) StartedAt() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.StartedAt)
}

func (p *permissionsSyncJobResolver) FinishedAt() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.FinishedAt)
}

func (p *permissionsSyncJobResolver) ProcessAfter() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.ProcessAfter)
}

func (p *permissionsSyncJobResolver) RanForMs() *int32 {
	var ranFor int32
	if !p.job.FinishedAt.IsZero() && !p.job.StartedAt.IsZero() {
		// Job runtime in ms shouldn't take more than a 32-bit int value.
		ranFor = int32(p.job.FinishedAt.Sub(p.job.StartedAt).Milliseconds())
	}
	return &ranFor
}

func (p *permissionsSyncJobResolver) NumResets() *int32 {
	return intToInt32Ptr(p.job.NumResets)
}

func (p *permissionsSyncJobResolver) NumFailures() *int32 {
	return intToInt32Ptr(p.job.NumFailures)
}

func (p *permissionsSyncJobResolver) LastHeartbeatAt() *gqlutil.DateTime {
	return gqlutil.FromTime(p.job.LastHeartbeatAt)
}

func (p *permissionsSyncJobResolver) WorkerHostname() string {
	return p.job.WorkerHostname
}

func (p *permissionsSyncJobResolver) Cancel() bool {
	return p.job.Cancel
}

func (p *permissionsSyncJobResolver) Subject() graphqlbackend.PermissionsSyncJobSubject {
	return p.syncSubject
}

func (p *permissionsSyncJobResolver) Priority() string {
	return p.job.Priority.ToString()
}

func (p *permissionsSyncJobResolver) NoPerms() bool {
	return p.job.NoPerms
}

func (p *permissionsSyncJobResolver) InvalidateCaches() bool {
	return p.job.InvalidateCaches
}

func (p *permissionsSyncJobResolver) PermissionsAdded() int32 {
	return int32(p.job.PermissionsAdded)
}

func (p *permissionsSyncJobResolver) PermissionsRemoved() int32 {
	return int32(p.job.PermissionsRemoved)
}

func (p *permissionsSyncJobResolver) PermissionsFound() int32 {
	return int32(p.job.PermissionsFound)
}

func (p *permissionsSyncJobResolver) CodeHostStates() []graphqlbackend.CodeHostStateResolver {
	resolvers := make([]graphqlbackend.CodeHostStateResolver, 0, len(p.job.CodeHostStates))
	for _, state := range p.job.CodeHostStates {
		resolvers = append(resolvers, codeHostStateResolver{state: state})
	}
	return resolvers
}

func (p *permissionsSyncJobResolver) PartialSuccess() bool {
	return p.job.IsPartialSuccess
}

func (p *permissionsSyncJobResolver) PlaceInQueue() *int32 {
	return p.job.PlaceInQueue
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

func (c codeHostStateResolver) Status() database.CodeHostStatus {
	return c.state.Status
}

func (c codeHostStateResolver) Message() string {
	return c.state.Message
}

type permissionSyncJobReasonResolver struct {
	group  database.PermissionsSyncJobReasonGroup
	reason database.PermissionsSyncJobReason
}

func (p permissionSyncJobReasonResolver) Group() string {
	return string(p.group)
}
func (p permissionSyncJobReasonResolver) Reason() *string {
	if p.reason == "" {
		return nil
	}

	reason := string(p.reason)

	return &reason
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

func marshalPermissionsSyncJobID(id int) graphql.ID {
	return relay.MarshalID(permissionsSyncJobIDKind, id)
}

func unmarshalPermissionsSyncJobID(id graphql.ID) (jobID int, err error) {
	if kind := relay.UnmarshalKind(id); kind != permissionsSyncJobIDKind {
		err = errors.Errorf("expected graphql ID to have kind %q; got %q", permissionsSyncJobIDKind, kind)
		return
	}
	err = relay.UnmarshalSpec(id, &jobID)
	return
}

func intToInt32Ptr(value int) *int32 {
	return pointers.Ptr(int32(value))
}
