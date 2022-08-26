package resolvers

import (
	"context"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var errDisabledSourcegraphDotCom = errors.New("not enabled on sourcegraph.com")

type Resolver struct {
	db                edb.EnterpriseDB
	repoupdaterClient interface {
		SchedulePermsSync(ctx context.Context, args protocol.PermsSyncRequest) error
	}
}

// checkLicense returns a user-facing error if the provided feature is not purchased
// with the current license or any error occurred while validating the licence.
func (r *Resolver) checkLicense(feature licensing.Feature) error {
	err := licensing.Check(feature)
	if err != nil {
		if licensing.IsFeatureNotActivated(err) {
			return err
		}

		log15.Error("authz.Resolver.checkLicense", "err", err)
		return errors.New("Unable to check license feature, please refer to logs for actual error message.")
	}
	return nil
}

func NewResolver(db database.DB, clock func() time.Time) graphqlbackend.AuthzResolver {
	return &Resolver{
		db:                edb.NewEnterpriseDB(db),
		repoupdaterClient: repoupdater.DefaultClient,
	}
}

func (r *Resolver) SetRepositoryPermissionsForUsers(ctx context.Context, args *graphqlbackend.RepoPermsArgs) (*graphqlbackend.EmptyResponse, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errDisabledSourcegraphDotCom
	}

	if err := r.checkLicense(licensing.FeatureExplicitPermissionsAPI); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins can mutate repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}
	// Make sure the repo ID is valid.
	if _, err = r.db.Repos().Get(ctx, repoID); err != nil {
		return nil, err
	}

	bindIDs := make([]string, 0, len(args.UserPermissions))
	for _, up := range args.UserPermissions {
		bindIDs = append(bindIDs, up.BindID)
	}

	mapping, err := r.db.Perms().MapUsers(ctx, bindIDs, globals.PermissionsUserMapping())
	if err != nil {
		return nil, err
	}

	pendingBindIDs := make([]string, 0, len(bindIDs))
	for _, bindID := range bindIDs {
		if _, ok := mapping[bindID]; !ok {
			pendingBindIDs = append(pendingBindIDs, bindID)
		}
	}

	userIDs := make(map[int32]struct{}, len(mapping))
	for _, id := range mapping {
		userIDs[id] = struct{}{}
	}

	p := &authz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    authz.Read, // Note: We currently only support read for repository permissions.
		UserIDs: userIDs,
	}

	txs, err := r.db.Perms().Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start transaction")
	}
	defer func() { err = txs.Done(err) }()

	accounts := &extsvc.Accounts{
		ServiceType: authz.SourcegraphServiceType,
		ServiceID:   authz.SourcegraphServiceID,
		AccountIDs:  pendingBindIDs,
	}

	if err = txs.SetRepoPermissions(ctx, p); err != nil {
		return nil, errors.Wrap(err, "set repository permissions")
	} else if err = txs.SetRepoPendingPermissions(ctx, accounts, p); err != nil {
		return nil, errors.Wrap(err, "set repository pending permissions")
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) SetRepositoryPermissionsUnrestricted(ctx context.Context, args *graphqlbackend.RepoUnrestrictedArgs) (*graphqlbackend.EmptyResponse, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errDisabledSourcegraphDotCom
	}

	if err := r.checkLicense(licensing.FeatureExplicitPermissionsAPI); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only site admins can mutate repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	ids := make([]int32, 0, len(args.Repositories))
	for _, id := range args.Repositories {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(id)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshalling id")
		}
		ids = append(ids, int32(repoID))
	}

	if err := r.db.Perms().SetRepoPermissionsUnrestricted(ctx, ids, args.Unrestricted); err != nil {
		return nil, errors.Wrap(err, "setting unrestricted field")
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleRepositoryPermissionsSync(ctx context.Context, args *graphqlbackend.RepositoryIDArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := r.checkLicense(licensing.FeatureACLs); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	err = r.repoupdaterClient.SchedulePermsSync(ctx, protocol.PermsSyncRequest{
		RepoIDs: []api.RepoID{repoID},
	})
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleUserPermissionsSync(ctx context.Context, args *graphqlbackend.UserPermissionsSyncArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := r.checkLicense(licensing.FeatureACLs); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := graphqlbackend.UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	req := protocol.PermsSyncRequest{
		UserIDs: []int32{userID},
	}
	if args.Options != nil && args.Options.InvalidateCaches != nil && *args.Options.InvalidateCaches {
		req.Options.InvalidateCaches = true
	}

	if err := r.repoupdaterClient.SchedulePermsSync(ctx, req); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) SetSubRepositoryPermissionsForUsers(ctx context.Context, args *graphqlbackend.SubRepoPermsArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := r.checkLicense(licensing.FeatureExplicitPermissionsAPI); err != nil {
		return nil, err
	}
	if envvar.SourcegraphDotComMode() {
		return nil, errDisabledSourcegraphDotCom
	}

	// 🚨 SECURITY: Only site admins can mutate repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	db, err := r.db.Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start transaction")
	}
	defer func() { err = db.Done(err) }()

	// Make sure the repo ID is valid.
	if _, err = db.Repos().Get(ctx, repoID); err != nil {
		return nil, err
	}

	bindIDs := make([]string, 0, len(args.UserPermissions))
	for _, up := range args.UserPermissions {
		bindIDs = append(bindIDs, up.BindID)
	}

	mapping, err := r.db.Perms().MapUsers(ctx, bindIDs, globals.PermissionsUserMapping())
	if err != nil {
		return nil, err
	}

	for _, perm := range args.UserPermissions {
		userID, ok := mapping[perm.BindID]
		if !ok {
			return nil, errors.Errorf("user %q not found", perm.BindID)
		}

		if err := db.SubRepoPerms().Upsert(ctx, userID, repoID, authz.SubRepoPermissions{
			PathIncludes: perm.PathIncludes,
			PathExcludes: perm.PathExcludes,
		}); err != nil {
			return nil, errors.Wrap(err, "upserting sub-repo permissions")
		}
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) SetRepositoryPermissionsForBitbucketProject(
	ctx context.Context, args *graphqlbackend.RepoPermsBitbucketProjectArgs,
) (*graphqlbackend.EmptyResponse, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errDisabledSourcegraphDotCom
	}

	if err := r.checkLicense(licensing.FeatureExplicitPermissionsAPI); err != nil {
		return nil, err
	}

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	externalServiceID, err := graphqlbackend.UnmarshalExternalServiceID(args.CodeHost)
	if err != nil {
		return nil, err
	}

	unrestricted := false
	if args.Unrestricted != nil {
		unrestricted = *args.Unrestricted
	}

	// get the external service and check if it is Bitbucket Server
	svc, err := r.db.ExternalServices().GetByID(ctx, externalServiceID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get external service %d", externalServiceID)
	}

	if svc.Kind != extsvc.KindBitbucketServer {
		return nil, errors.Newf("expected Bitbucket Server external service, got: %s", svc.Kind)
	}

	jobID, err := r.db.BitbucketProjectPermissions().Enqueue(ctx, args.ProjectKey, externalServiceID, args.UserPermissions, unrestricted)
	if err != nil {
		return nil, err
	}

	log15.Debug("SetRepositoryPermissionsForBitbucketProject: job enqueued", "jobID", jobID)

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) AuthorizedUserRepositories(ctx context.Context, args *graphqlbackend.AuthorizedRepoArgs) (graphqlbackend.RepositoryConnectionResolver, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errDisabledSourcegraphDotCom
	}

	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var (
		err    error
		bindID string
		user   *types.User
	)
	if args.Email != nil {
		bindID = *args.Email
		// 🚨 SECURITY: It is critical to ensure the email is verified.
		user, err = r.db.Users().GetByVerifiedEmail(ctx, *args.Email)
	} else if args.Username != nil {
		bindID = *args.Username
		user, err = r.db.Users().GetByUsername(ctx, *args.Username)
	} else {
		return nil, errors.New("neither email nor username is given to identify a user")
	}
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	var ids []int32
	if user != nil {
		p := &authz.UserPermissions{
			UserID: user.ID,
			Perm:   authz.Read, // Note: We currently only support read for repository permissions.
			Type:   authz.PermRepos,
		}
		err = r.db.Perms().LoadUserPermissions(ctx, p)
		ids = p.GenerateSortedIDsSlice()
	} else {
		p := &authz.UserPendingPermissions{
			ServiceType: authz.SourcegraphServiceType,
			ServiceID:   authz.SourcegraphServiceID,
			BindID:      bindID,
			Perm:        authz.Read, // Note: We currently only support read for repository permissions.
			Type:        authz.PermRepos,
		}
		err = r.db.Perms().LoadUserPendingPermissions(ctx, p)
		ids = p.GenerateSortedIDsSlice()
	}
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}
	// If no row is found, we return an empty list to the consumer.
	if err == authz.ErrPermsNotFound {
		ids = []int32{}
	}

	return &repositoryConnectionResolver{
		db:    r.db,
		ids:   ids,
		first: args.First,
		after: args.After,
	}, nil
}

func (r *Resolver) UsersWithPendingPermissions(ctx context.Context) ([]string, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return r.db.Perms().ListPendingUsers(ctx, authz.SourcegraphServiceType, authz.SourcegraphServiceID)
}

func (r *Resolver) AuthorizedUsers(ctx context.Context, args *graphqlbackend.RepoAuthorizedUserArgs) (graphqlbackend.UserConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.RepositoryID)
	if err != nil {
		return nil, err
	}
	// Make sure the repo ID is valid.
	if _, err = r.db.Repos().Get(ctx, repoID); err != nil {
		return nil, err
	}

	p := &authz.RepoPermissions{
		RepoID: int32(repoID),
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
	}
	err = r.db.Perms().LoadRepoPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}
	// If no row is found, we return an empty list to the consumer.
	if err == authz.ErrPermsNotFound {
		p.UserIDs = map[int32]struct{}{}
	}

	return &userConnectionResolver{
		db:    r.db,
		ids:   p.GenerateSortedIDsSlice(),
		first: args.First,
		after: args.After,
	}, nil
}

var jobStatuses = map[string]bool{
	"queued":     true,
	"processing": true,
	"completed":  true,
	"canceled":   true,
	"errored":    true,
	"failed":     true,
}

func (r *Resolver) BitbucketProjectPermissionJobs(ctx context.Context, args *graphqlbackend.BitbucketProjectPermissionJobsArgs) (graphqlbackend.BitbucketProjectsPermissionJobsResolver, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errDisabledSourcegraphDotCom
	}
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	loweredAndTrimmedStatus := strings.ToLower(strings.TrimSpace(getOrDefault(args.Status)))
	if loweredAndTrimmedStatus != "" && !jobStatuses[loweredAndTrimmedStatus] {
		return nil, errors.New("Please provide one of the following job statuses: queued, processing, completed, canceled, errored, failed")
	}
	args.Status = &loweredAndTrimmedStatus

	jobs, err := r.db.BitbucketProjectPermissions().ListJobs(ctx, convertJobsArgsToOpts(args))
	if err != nil {
		return nil, errors.Wrap(err, "getting a list of Bitbucket Projects permission sync jobs")
	}
	return NewBitbucketProjectsPermissionJobsResolver(jobs), nil
}

func convertJobsArgsToOpts(args *graphqlbackend.BitbucketProjectPermissionJobsArgs) database.ListJobsOptions {
	if args == nil {
		return database.ListJobsOptions{}
	}

	return database.ListJobsOptions{
		ProjectKeys: getOrDefault(args.ProjectKeys),
		State:       getOrDefault(args.Status),
		Count:       getOrDefault(args.Count),
	}
}

// getOrDefault accepts a pointer of a type T and returns dereferenced value if the pointer
// is not nil, or zero-value for the given type otherwise
func getOrDefault[T any](ptr *T) T {
	var result T
	if ptr == nil {
		return result
	} else {
		return *ptr
	}
}

type permissionsInfoResolver struct {
	perms        authz.Perms
	syncedAt     time.Time
	updatedAt    time.Time
	unrestricted bool
}

func (r *permissionsInfoResolver) Permissions() []string {
	return strings.Split(strings.ToUpper(r.perms.String()), ",")
}

func (r *permissionsInfoResolver) SyncedAt() *graphqlbackend.DateTime {
	if r.syncedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.syncedAt}
}

func (r *permissionsInfoResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.updatedAt}
}

func (r *permissionsInfoResolver) Unrestricted() bool {
	return r.unrestricted
}

func (r *Resolver) RepositoryPermissionsInfo(ctx context.Context, id graphql.ID) (graphqlbackend.PermissionsInfoResolver, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(id)
	if err != nil {
		return nil, err
	}
	// Make sure the repo ID is valid and not soft-deleted.
	if _, err = r.db.Repos().Get(ctx, repoID); err != nil {
		return nil, err
	}

	p := &authz.RepoPermissions{
		RepoID: int32(repoID),
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
	}
	err = r.db.Perms().LoadRepoPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}

	if err == authz.ErrPermsNotFound {
		return nil, nil // It is acceptable to have no permissions information, i.e. nullable.
	}

	return &permissionsInfoResolver{
		perms:        p.Perm,
		syncedAt:     p.SyncedAt,
		updatedAt:    p.UpdatedAt,
		unrestricted: p.Unrestricted,
	}, nil
}

func (r *Resolver) UserPermissionsInfo(ctx context.Context, id graphql.ID) (graphqlbackend.PermissionsInfoResolver, error) {
	// 🚨 SECURITY: Only site admins can query user permissions.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := graphqlbackend.UnmarshalUserID(id)
	if err != nil {
		return nil, err
	}
	// Make sure the user ID is valid and not soft-deleted.
	if _, err = r.db.Users().GetByID(ctx, userID); err != nil {
		return nil, err
	}

	p := &authz.UserPermissions{
		UserID: userID,
		Perm:   authz.Read, // Note: We currently only support read for repository permissions.
		Type:   authz.PermRepos,
	}
	err = r.db.Perms().LoadUserPermissions(ctx, p)
	if err != nil && err != authz.ErrPermsNotFound {
		return nil, err
	}

	if err == authz.ErrPermsNotFound {
		return nil, nil // It is acceptable to have no permissions information, i.e. nullable.
	}

	return &permissionsInfoResolver{
		perms:     p.Perm,
		syncedAt:  p.SyncedAt,
		updatedAt: p.UpdatedAt,
	}, nil
}
