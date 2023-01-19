package resolvers

import (
	"context"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/authz/syncjobs"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var errDisabledSourcegraphDotCom = errors.New("not enabled on sourcegraph.com")

type Resolver struct {
	logger          log.Logger
	db              edb.EnterpriseDB
	syncJobsRecords interface {
		Get(timestamp time.Time) (*syncjobs.Status, error)
		GetAll(ctx context.Context, first int) ([]syncjobs.Status, error)
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

		r.logger.Error("Unable to check license for feature", log.Error(err))
		return errors.New("Unable to check license feature, please refer to logs for actual error message.")
	}
	return nil
}

func NewResolver(observationCtx *observation.Context, db database.DB, clock func() time.Time) graphqlbackend.AuthzResolver {
	return &Resolver{
		logger:          observationCtx.Logger.Scoped("authz.Resolver", ""),
		db:              edb.NewEnterpriseDB(db),
		syncJobsRecords: syncjobs.NewRecordsReader(),
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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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

	// 🚨 SECURITY: Only site admins can trigger repository permissions syncs.
	user, err := auth.CheckCurrentUserIsSiteAdminAndReturn(ctx, r.db)
	if err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	userID := int32(0)
	// user is nil in case of internal actor
	if user != nil {
		userID = user.ID
	}
	req := protocol.PermsSyncRequest{RepoIDs: []api.RepoID{repoID}, Reason: permssync.ReasonManualRepoSync, TriggeredByUserID: userID}
	permssync.SchedulePermsSync(ctx, r.logger, r.db, req)

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleUserPermissionsSync(ctx context.Context, args *graphqlbackend.UserPermissionsSyncArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := r.checkLicense(licensing.FeatureACLs); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site admins can trigger user permissions syncs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := graphqlbackend.UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	req := protocol.PermsSyncRequest{UserIDs: []int32{userID}, Reason: permssync.ReasonManualUserSync, TriggeredByUserID: userID}
	if args.Options != nil && args.Options.InvalidateCaches != nil && *args.Options.InvalidateCaches {
		req.Options.InvalidateCaches = true
	}

	permssync.SchedulePermsSync(ctx, r.logger, r.db, req)

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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	ossDB, err := r.db.Transact(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "start transaction")
	}
	defer func() { err = ossDB.Done(err) }()
	db := edb.NewEnterpriseDB(ossDB)

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
		if (perm.PathIncludes == nil || perm.PathExcludes == nil) && perm.Paths == nil {
			return nil, errors.New("either both pathIncludes and pathExcludes needs to be set, or paths needs to be set")
		}
	}

	for _, perm := range args.UserPermissions {
		userID, ok := mapping[perm.BindID]
		if !ok {
			return nil, errors.Errorf("user %q not found", perm.BindID)
		}

		var paths []string
		if perm.Paths == nil {
			paths = make([]string, 0, len(*perm.PathIncludes)+len(*perm.PathExcludes))
			for _, include := range *perm.PathIncludes {
				if !strings.HasPrefix(include, "/") { // ensure leading slash
					include = "/" + include
				}
				paths = append(paths, include)
			}
			for _, exclude := range *perm.PathExcludes {
				if !strings.HasPrefix(exclude, "/") { // ensure leading slash
					exclude = "/" + exclude
				}
				paths = append(paths, "-"+exclude) // excludes start with a minus (-)
			}
		} else {
			paths = make([]string, 0, len(*perm.Paths))
			for _, path := range *perm.Paths {
				if strings.HasPrefix(path, "-") {
					if !strings.HasPrefix(path, "-/") {
						path = "-/" + strings.TrimPrefix(path, "-")
					}
				} else {
					if !strings.HasPrefix(path, "/") {
						path = "/" + path
					}
				}
				paths = append(paths, path)
			}
		}

		if err := db.SubRepoPerms().Upsert(ctx, userID, repoID, authz.SubRepoPermissions{
			Paths: paths,
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

	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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

	r.logger.Debug("Bitbucket project permissions job enqueued", log.Int("jobID", jobID))

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) AuthorizedUserRepositories(ctx context.Context, args *graphqlbackend.AuthorizedRepoArgs) (graphqlbackend.RepositoryConnectionResolver, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errDisabledSourcegraphDotCom
	}

	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return r.db.Perms().ListPendingUsers(ctx, authz.SourcegraphServiceType, authz.SourcegraphServiceID)
}

func (r *Resolver) AuthorizedUsers(ctx context.Context, args *graphqlbackend.RepoAuthorizedUserArgs) (graphqlbackend.UserConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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

func (r *Resolver) AuthzProviderTypes(ctx context.Context) ([]string, error) {
	// 🚨 SECURITY: Only site admins can query for authz providers.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	_, providers := authz.GetProviders()
	providerTypes := make([]string, 0, len(providers))
	for _, p := range providers {
		providerTypes = append(providerTypes, p.ServiceType())
	}
	return providerTypes, nil
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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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

func (r *permissionsInfoResolver) SyncedAt() *gqlutil.DateTime {
	if r.syncedAt.IsZero() {
		return nil
	}
	return &gqlutil.DateTime{Time: r.syncedAt}
}

func (r *permissionsInfoResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.updatedAt}
}

func (r *permissionsInfoResolver) Unrestricted() bool {
	return r.unrestricted
}

func (r *Resolver) RepositoryPermissionsInfo(ctx context.Context, id graphql.ID) (graphqlbackend.PermissionsInfoResolver, error) {
	// 🚨 SECURITY: Only site admins can query repository permissions.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
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

func (r *Resolver) PermissionsSyncJobs(ctx context.Context, args *graphqlbackend.PermissionsSyncJobsArgs) (graphqlbackend.PermissionsSyncJobsConnection, error) {
	// 🚨 SECURITY: Only site admins can query sync jobs records.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	if args.First == 0 {
		return nil, errors.Newf("expected non-zero 'first', got %d", args.First)
	}

	records, err := r.syncJobsRecords.GetAll(ctx, int(args.First))
	if err != nil {
		return nil, err
	}

	jobs := &permissionsSyncJobsConnection{
		jobs: make([]graphqlbackend.PermissionsSyncJobResolver, 0, len(records)),
	}
	for _, j := range records {
		// If status is not provided, add all - otherwise, check if the job's status
		// matches the argument status.
		if args.Status == nil {
			jobs.jobs = append(jobs.jobs, permissionsSyncJobResolver{j})
		} else if j.Status == *args.Status {
			jobs.jobs = append(jobs.jobs, permissionsSyncJobResolver{j})
		}
	}

	return jobs, nil
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		permissionsSyncJobKind: getPermissionsSyncJobByIDFunc(r),
	}
}
