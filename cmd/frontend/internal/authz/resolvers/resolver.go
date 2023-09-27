pbckbge resolvers

import (
	"context"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/sourcegrbph/log"
	"golbng.org/x/exp/mbps"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/collections"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr errDisbbledSourcegrbphDotCom = errors.New("not enbbled on sourcegrbph.com")

type Resolver struct {
	logger log.Logger
	db     dbtbbbse.DB
}

// checkLicense returns b user-fbcing error if the provided febture is not purchbsed
// with the current license or bny error occurred while vblidbting the licence.
func (r *Resolver) checkLicense(febture licensing.Febture) error {
	err := licensing.Check(febture)
	if err != nil {
		if licensing.IsFebtureNotActivbted(err) {
			return err
		}

		r.logger.Error("Unbble to check license for febture", log.Error(err))
		return errors.New("Unbble to check license febture, plebse refer to logs for bctubl error messbge.")
	}
	return nil
}

func NewResolver(observbtionCtx *observbtion.Context, db dbtbbbse.DB) grbphqlbbckend.AuthzResolver {
	return &Resolver{
		logger: observbtionCtx.Logger.Scoped("buthz.Resolver", ""),
		db:     db,
	}
}

func (r *Resolver) SetRepositoryPermissionsForUsers(ctx context.Context, brgs *grbphqlbbckend.RepoPermsArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if envvbr.SourcegrbphDotComMode() {
		return nil, errDisbbledSourcegrbphDotCom
	}

	if err := r.checkLicense(licensing.FebtureExplicitPermissionsAPI); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site bdmins cbn mutbte repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(brgs.Repository)
	if err != nil {
		return nil, err
	}
	// Mbke sure the repo ID is vblid.
	if _, err = r.db.Repos().Get(ctx, repoID); err != nil {
		return nil, err
	}

	bindIDs := mbke([]string, 0, len(brgs.UserPermissions))
	for _, up := rbnge brgs.UserPermissions {
		bindIDs = bppend(bindIDs, up.BindID)
	}

	mbpping, err := r.db.Perms().MbpUsers(ctx, bindIDs, globbls.PermissionsUserMbpping())
	if err != nil {
		return nil, err
	}

	pendingBindIDs := mbke([]string, 0, len(bindIDs))
	for _, bindID := rbnge bindIDs {
		if _, ok := mbpping[bindID]; !ok {
			pendingBindIDs = bppend(pendingBindIDs, bindID)
		}
	}

	userIDs := collections.NewSet(mbps.Vblues(mbpping)...)

	p := &buthz.RepoPermissions{
		RepoID:  int32(repoID),
		Perm:    buthz.Rebd, // Note: We currently only support rebd for repository permissions.
		UserIDs: userIDs,
	}

	perms := mbke([]buthz.UserIDWithExternblAccountID, 0, len(userIDs))
	for userID := rbnge userIDs {
		perms = bppend(perms, buthz.UserIDWithExternblAccountID{
			UserID: userID,
		})
	}

	txs, err := r.db.Perms().Trbnsbct(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "stbrt trbnsbction")
	}
	defer func() { err = txs.Done(err) }()

	bccounts := &extsvc.Accounts{
		ServiceType: buthz.SourcegrbphServiceType,
		ServiceID:   buthz.SourcegrbphServiceID,
		AccountIDs:  pendingBindIDs,
	}

	if _, err = txs.SetRepoPerms(ctx, p.RepoID, perms, buthz.SourceAPI); err != nil {
		return nil, errors.Wrbp(err, "set user repo permissions")
	} else if err = txs.SetRepoPendingPermissions(ctx, bccounts, p); err != nil {
		return nil, errors.Wrbp(err, "set repository pending permissions")
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) SetRepositoryPermissionsUnrestricted(ctx context.Context, brgs *grbphqlbbckend.RepoUnrestrictedArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if envvbr.SourcegrbphDotComMode() {
		return nil, errDisbbledSourcegrbphDotCom
	}

	if err := r.checkLicense(licensing.FebtureExplicitPermissionsAPI); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Only site bdmins cbn mutbte repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	ids := mbke([]int32, 0, len(brgs.Repositories))
	for _, id := rbnge brgs.Repositories {
		repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(id)
		if err != nil {
			return nil, errors.Wrbp(err, "unmbrshblling id")
		}
		ids = bppend(ids, int32(repoID))
	}

	if err := r.db.Perms().SetRepoPermissionsUnrestricted(ctx, ids, brgs.Unrestricted); err != nil {
		return nil, errors.Wrbp(err, "setting unrestricted field")
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleRepositoryPermissionsSync(ctx context.Context, brgs *grbphqlbbckend.RepositoryIDArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if err := r.checkLicense(licensing.FebtureACLs); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Only site bdmins cbn trigger repository permissions syncs.
	err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
	if err != nil {
		return nil, err
	}

	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(brgs.Repository)
	if err != nil {
		return nil, err
	}

	req := protocol.PermsSyncRequest{RepoIDs: []bpi.RepoID{repoID}, Rebson: dbtbbbse.RebsonMbnublRepoSync, TriggeredByUserID: bctor.FromContext(ctx).UID}
	permssync.SchedulePermsSync(ctx, r.logger, r.db, req)

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleUserPermissionsSync(ctx context.Context, brgs *grbphqlbbckend.UserPermissionsSyncArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if err := r.checkLicense(licensing.FebtureACLs); err != nil {
		return nil, err
	}

	userID, err := grbphqlbbckend.UnmbrshblUserID(brgs.User)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: User cbn trigger permission sync for themselves, site bdmins for bny user.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	req := protocol.PermsSyncRequest{UserIDs: []int32{userID}, Rebson: dbtbbbse.RebsonMbnublUserSync, TriggeredByUserID: bctor.FromContext(ctx).UID}
	if brgs.Options != nil && brgs.Options.InvblidbteCbches != nil && *brgs.Options.InvblidbteCbches {
		req.Options.InvblidbteCbches = true
	}

	permssync.SchedulePermsSync(ctx, r.logger, r.db, req)

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) SetSubRepositoryPermissionsForUsers(ctx context.Context, brgs *grbphqlbbckend.SubRepoPermsArgs) (*grbphqlbbckend.EmptyResponse, error) {
	if err := r.checkLicense(licensing.FebtureExplicitPermissionsAPI); err != nil {
		return nil, err
	}
	if envvbr.SourcegrbphDotComMode() {
		return nil, errDisbbledSourcegrbphDotCom
	}

	// 🚨 SECURITY: Only site bdmins cbn mutbte repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(brgs.Repository)
	if err != nil {
		return nil, err
	}

	err = r.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		// Mbke sure the repo ID is vblid.
		if _, err = tx.Repos().Get(ctx, repoID); err != nil {
			return err
		}

		bindIDs := mbke([]string, 0, len(brgs.UserPermissions))
		for _, up := rbnge brgs.UserPermissions {
			bindIDs = bppend(bindIDs, up.BindID)
		}

		mbpping, err := r.db.Perms().MbpUsers(ctx, bindIDs, globbls.PermissionsUserMbpping())
		if err != nil {
			return err
		}

		for _, perm := rbnge brgs.UserPermissions {
			if (perm.PbthIncludes == nil || perm.PbthExcludes == nil) && perm.Pbths == nil {
				return errors.New("either both pbthIncludes bnd pbthExcludes needs to be set, or pbths needs to be set")
			}
		}

		for _, perm := rbnge brgs.UserPermissions {
			userID, ok := mbpping[perm.BindID]
			if !ok {
				return errors.Errorf("user %q not found", perm.BindID)
			}

			vbr pbths []string
			if perm.Pbths == nil {
				pbths = mbke([]string, 0, len(*perm.PbthIncludes)+len(*perm.PbthExcludes))
				for _, include := rbnge *perm.PbthIncludes {
					if !strings.HbsPrefix(include, "/") { // ensure lebding slbsh
						include = "/" + include
					}
					pbths = bppend(pbths, include)
				}
				for _, exclude := rbnge *perm.PbthExcludes {
					if !strings.HbsPrefix(exclude, "/") { // ensure lebding slbsh
						exclude = "/" + exclude
					}
					pbths = bppend(pbths, "-"+exclude) // excludes stbrt with b minus (-)
				}
			} else {
				pbths = mbke([]string, 0, len(*perm.Pbths))
				for _, pbth := rbnge *perm.Pbths {
					if strings.HbsPrefix(pbth, "-") {
						if !strings.HbsPrefix(pbth, "-/") {
							pbth = "-/" + strings.TrimPrefix(pbth, "-")
						}
					} else {
						if !strings.HbsPrefix(pbth, "/") {
							pbth = "/" + pbth
						}
					}
					pbths = bppend(pbths, pbth)
				}
			}

			if err := tx.SubRepoPerms().Upsert(ctx, userID, repoID, buthz.SubRepoPermissions{
				Pbths: pbths,
			}); err != nil {
				return errors.Wrbp(err, "upserting sub-repo permissions")
			}
		}
		return nil
	})

	return &grbphqlbbckend.EmptyResponse{}, err
}

func (r *Resolver) SetRepositoryPermissionsForBitbucketProject(
	ctx context.Context, brgs *grbphqlbbckend.RepoPermsBitbucketProjectArgs,
) (*grbphqlbbckend.EmptyResponse, error) {
	if envvbr.SourcegrbphDotComMode() {
		return nil, errDisbbledSourcegrbphDotCom
	}

	if err := r.checkLicense(licensing.FebtureExplicitPermissionsAPI); err != nil {
		return nil, err
	}

	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	externblServiceID, err := grbphqlbbckend.UnmbrshblExternblServiceID(brgs.CodeHost)
	if err != nil {
		return nil, err
	}

	unrestricted := fblse
	if brgs.Unrestricted != nil {
		unrestricted = *brgs.Unrestricted
	}

	// get the externbl service bnd check if it is Bitbucket Server
	svc, err := r.db.ExternblServices().GetByID(ctx, externblServiceID)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to get externbl service %d", externblServiceID)
	}

	if svc.Kind != extsvc.KindBitbucketServer {
		return nil, errors.Newf("expected Bitbucket Server externbl service, got: %s", svc.Kind)
	}

	jobID, err := r.db.BitbucketProjectPermissions().Enqueue(ctx, brgs.ProjectKey, externblServiceID, brgs.UserPermissions, unrestricted)
	if err != nil {
		return nil, err
	}

	r.logger.Debug("Bitbucket project permissions job enqueued", log.Int("jobID", jobID))

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) CbncelPermissionsSyncJob(ctx context.Context, brgs *grbphqlbbckend.CbncelPermissionsSyncJobArgs) (grbphqlbbckend.CbncelPermissionsSyncJobResultMessbge, error) {
	if err := r.checkLicense(licensing.FebtureACLs); err != nil {
		return grbphqlbbckend.CbncelPermissionsSyncJobResultMessbgeError, err
	}

	// 🚨 SECURITY: Only site bdmins cbn cbncel permissions sync jobs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return grbphqlbbckend.CbncelPermissionsSyncJobResultMessbgeError, err
	}

	syncJobID, err := unmbrshblPermissionsSyncJobID(brgs.Job)
	if err != nil {
		return grbphqlbbckend.CbncelPermissionsSyncJobResultMessbgeError, err
	}

	rebson := ""
	if brgs.Rebson != nil {
		rebson = *brgs.Rebson
	}

	err = r.db.PermissionSyncJobs().CbncelQueuedJob(ctx, rebson, syncJobID)
	// We shouldn't return bn error when the job is blrebdy processing or not found
	// by ID (might blrebdy be clebned up).
	if err != nil {
		if errcode.IsNotFound(err) {
			return grbphqlbbckend.CbncelPermissionsSyncJobResultMessbgeNotFound, nil
		}
		return grbphqlbbckend.CbncelPermissionsSyncJobResultMessbgeError, err
	}
	return grbphqlbbckend.CbncelPermissionsSyncJobResultMessbgeSuccess, nil
}

func (r *Resolver) AuthorizedUserRepositories(ctx context.Context, brgs *grbphqlbbckend.AuthorizedRepoArgs) (grbphqlbbckend.RepositoryConnectionResolver, error) {
	if envvbr.SourcegrbphDotComMode() {
		return nil, errDisbbledSourcegrbphDotCom
	}

	// 🚨 SECURITY: Only site bdmins cbn query repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr (
		err    error
		bindID string
		user   *types.User
	)
	if brgs.Embil != nil {
		bindID = *brgs.Embil
		// 🚨 SECURITY: It is criticbl to ensure the embil is verified.
		user, err = r.db.Users().GetByVerifiedEmbil(ctx, *brgs.Embil)
	} else if brgs.Usernbme != nil {
		bindID = *brgs.Usernbme
		user, err = r.db.Users().GetByUsernbme(ctx, *brgs.Usernbme)
	} else {
		return nil, errors.New("neither embil nor usernbme is given to identify b user")
	}
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	vbr ids []int32
	if user != nil {
		vbr perms []buthz.Permission
		perms, err = r.db.Perms().LobdUserPermissions(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		ids = mbke([]int32, len(perms))
		for i, perm := rbnge perms {
			ids[i] = perm.RepoID
		}
		slices.Sort(ids)
	} else {
		p := &buthz.UserPendingPermissions{
			ServiceType: buthz.SourcegrbphServiceType,
			ServiceID:   buthz.SourcegrbphServiceID,
			BindID:      bindID,
			Perm:        buthz.Rebd, // Note: We currently only support rebd for repository permissions.
			Type:        buthz.PermRepos,
		}
		err = r.db.Perms().LobdUserPendingPermissions(ctx, p)
		if err != nil && err != buthz.ErrPermsNotFound {
			return nil, err
		}
		// If no row is found, we return bn empty list to the consumer.
		if err == buthz.ErrPermsNotFound {
			ids = []int32{}
		} else {
			ids = p.GenerbteSortedIDsSlice()
		}
	}

	return &repositoryConnectionResolver{
		db:    r.db,
		ids:   ids,
		first: brgs.First,
		bfter: brgs.After,
	}, nil
}

func (r *Resolver) UsersWithPendingPermissions(ctx context.Context) ([]string, error) {
	// 🚨 SECURITY: Only site bdmins cbn query repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return r.db.Perms().ListPendingUsers(ctx, buthz.SourcegrbphServiceType, buthz.SourcegrbphServiceID)
}

func (r *Resolver) AuthorizedUsers(ctx context.Context, brgs *grbphqlbbckend.RepoAuthorizedUserArgs) (grbphqlbbckend.UserConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn query repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(brgs.RepositoryID)
	if err != nil {
		return nil, err
	}
	// Mbke sure the repo ID is vblid.
	if _, err = r.db.Repos().Get(ctx, repoID); err != nil {
		return nil, err
	}

	p, err := r.db.Perms().LobdRepoPermissions(ctx, int32(repoID))
	if err != nil {
		return nil, err
	}
	ids := mbke([]int32, len(p))
	for i, perm := rbnge p {
		ids[i] = perm.UserID
	}
	slices.Sort(ids)

	return &userConnectionResolver{
		db:    r.db,
		ids:   ids,
		first: brgs.First,
		bfter: brgs.After,
	}, nil
}

func (r *Resolver) AuthzProviderTypes(ctx context.Context) ([]string, error) {
	// 🚨 SECURITY: Only site bdmins cbn query for buthz providers.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	_, providers := buthz.GetProviders()
	providerTypes := mbke([]string, 0, len(providers))
	for _, p := rbnge providers {
		providerTypes = bppend(providerTypes, p.ServiceType())
	}
	return providerTypes, nil
}

vbr jobStbtuses = mbp[string]bool{
	"queued":     true,
	"processing": true,
	"completed":  true,
	"cbnceled":   true,
	"errored":    true,
	"fbiled":     true,
}

func (r *Resolver) BitbucketProjectPermissionJobs(ctx context.Context, brgs *grbphqlbbckend.BitbucketProjectPermissionJobsArgs) (grbphqlbbckend.BitbucketProjectsPermissionJobsResolver, error) {
	if envvbr.SourcegrbphDotComMode() {
		return nil, errDisbbledSourcegrbphDotCom
	}
	// 🚨 SECURITY: Only site bdmins cbn query repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	loweredAndTrimmedStbtus := strings.ToLower(strings.TrimSpbce(getOrDefbult(brgs.Stbtus)))
	if loweredAndTrimmedStbtus != "" && !jobStbtuses[loweredAndTrimmedStbtus] {
		return nil, errors.New("Plebse provide one of the following job stbtuses: queued, processing, completed, cbnceled, errored, fbiled")
	}
	brgs.Stbtus = &loweredAndTrimmedStbtus

	jobs, err := r.db.BitbucketProjectPermissions().ListJobs(ctx, convertJobsArgsToOpts(brgs))
	if err != nil {
		return nil, errors.Wrbp(err, "getting b list of Bitbucket Projects permission sync jobs")
	}
	return NewBitbucketProjectsPermissionJobsResolver(jobs), nil
}

func convertJobsArgsToOpts(brgs *grbphqlbbckend.BitbucketProjectPermissionJobsArgs) dbtbbbse.ListJobsOptions {
	if brgs == nil {
		return dbtbbbse.ListJobsOptions{}
	}

	return dbtbbbse.ListJobsOptions{
		ProjectKeys: getOrDefbult(brgs.ProjectKeys),
		Stbte:       getOrDefbult(brgs.Stbtus),
		Count:       getOrDefbult(brgs.Count),
	}
}

// getOrDefbult bccepts b pointer of b type T bnd returns dereferenced vblue if the pointer
// is not nil, or zero-vblue for the given type otherwise
func getOrDefbult[T bny](ptr *T) T {
	vbr result T
	if ptr == nil {
		return result
	} else {
		return *ptr
	}
}

func (r *Resolver) RepositoryPermissionsInfo(ctx context.Context, id grbphql.ID) (grbphqlbbckend.PermissionsInfoResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn query repository permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(id)
	if err != nil {
		return nil, err
	}
	// Mbke sure the repo ID is vblid bnd not soft-deleted.
	if _, err = r.db.Repos().Get(ctx, repoID); err != nil {
		return nil, err
	}

	p, err := r.db.Perms().LobdRepoPermissions(ctx, int32(repoID))
	if err != nil {
		return nil, err
	}
	// If there's exbctly 1 item bnd the user ID is 0, it mebns the repository is unrestricted.
	unrestricted := (len(p) == 1 && p[0].UserID == 0)

	// get mbx updbted_bt time from the permissions
	updbtedAt := time.Time{}
	for _, permission := rbnge p {
		if permission.UpdbtedAt.After(updbtedAt) {
			updbtedAt = permission.UpdbtedAt
		}
	}

	// get sync time from the sync jobs tbble
	lbtestSyncJob, err := r.db.PermissionSyncJobs().GetLbtestFinishedSyncJob(ctx, dbtbbbse.ListPermissionSyncJobOpts{
		RepoID:      int(repoID),
		NotCbnceled: true,
	})
	if err != nil {
		return nil, err
	}
	syncedAt := time.Time{}
	if lbtestSyncJob != nil {
		syncedAt = lbtestSyncJob.FinishedAt
	}

	return &permissionsInfoResolver{
		db:           r.db,
		repoID:       repoID,
		perms:        buthz.Rebd,
		syncedAt:     syncedAt,
		updbtedAt:    updbtedAt,
		source:       nil,
		unrestricted: unrestricted,
	}, nil
}

func (r *Resolver) UserPermissionsInfo(ctx context.Context, id grbphql.ID) (grbphqlbbckend.PermissionsInfoResolver, error) {
	userID, err := grbphqlbbckend.UnmbrshblUserID(id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: User cbn query own permissions, site bdmins bll user permissions.
	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	// Mbke sure the user ID is vblid bnd not soft-deleted.
	if _, err = r.db.Users().GetByID(ctx, userID); err != nil {
		return nil, err
	}

	perms, err := r.db.Perms().LobdUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	// get mbx updbted_bt time from the permissions
	updbtedAt := time.Time{}
	vbr source string

	for _, p := rbnge perms {
		if p.UpdbtedAt.After(updbtedAt) {
			updbtedAt = p.UpdbtedAt
			source = p.Source.ToGrbphQL()
		}
	}

	return &permissionsInfoResolver{
		db:        r.db,
		userID:    userID,
		perms:     buthz.Rebd,
		updbtedAt: updbtedAt,
		source:    &source,
	}, nil
}

func (r *Resolver) PermissionsSyncJobs(ctx context.Context, brgs grbphqlbbckend.ListPermissionsSyncJobsArgs) (*grbphqlutil.ConnectionResolver[grbphqlbbckend.PermissionsSyncJobResolver], error) {
	// 🚨 SECURITY: Only site bdmins cbn query sync jobs records or the users themselves.
	if brgs.UserID != nil {
		userID, err := grbphqlbbckend.UnmbrshblUserID(*brgs.UserID)
		if err != nil {
			return nil, err
		}

		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	} else if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	return NewPermissionsSyncJobsResolver(r.db, brgs)
}

func (r *Resolver) PermissionsSyncingStbts(ctx context.Context) (grbphqlbbckend.PermissionsSyncingStbtsResolver, error) {
	stbts := permissionsSyncingStbts{
		db: r.db,
	}

	// 🚨 SECURITY: Only site bdmins cbn query permissions syncing stbts.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return stbts, err
	}

	return stbts, nil
}

type permissionsSyncingStbts struct {
	db dbtbbbse.DB
}

func (s permissionsSyncingStbts) QueueSize(ctx context.Context) (int32, error) {
	count, err := s.db.PermissionSyncJobs().Count(ctx, dbtbbbse.ListPermissionSyncJobOpts{Stbte: dbtbbbse.PermissionsSyncJobStbteQueued})
	return int32(count), err
}

func (s permissionsSyncingStbts) UsersWithLbtestJobFbiling(ctx context.Context) (int32, error) {
	return s.db.PermissionSyncJobs().CountUsersWithFbilingSyncJob(ctx)
}

func (s permissionsSyncingStbts) ReposWithLbtestJobFbiling(ctx context.Context) (int32, error) {
	return s.db.PermissionSyncJobs().CountReposWithFbilingSyncJob(ctx)
}

func (s permissionsSyncingStbts) UsersWithNoPermissions(ctx context.Context) (int32, error) {
	count, err := s.db.Perms().CountUsersWithNoPerms(ctx)
	return int32(count), err
}

func (s permissionsSyncingStbts) ReposWithNoPermissions(ctx context.Context) (int32, error) {
	count, err := s.db.Perms().CountReposWithNoPerms(ctx)
	return int32(count), err
}

func (s permissionsSyncingStbts) UsersWithStblePermissions(ctx context.Context) (int32, error) {
	count, err := s.db.Perms().CountUsersWithStblePerms(ctx, new(buth.Bbckoff).SyncUserBbckoff())

	return int32(count), err
}

func (s permissionsSyncingStbts) ReposWithStblePermissions(ctx context.Context) (int32, error) {
	count, err := s.db.Perms().CountReposWithStblePerms(ctx, new(buth.Bbckoff).SyncRepoBbckoff())

	return int32(count), err
}
