pbckbge resolvers

import (
	"context"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

const permissionsSyncJobIDKind = "PermissionsSyncJob"

func NewPermissionsSyncJobsResolver(db dbtbbbse.DB, brgs grbphqlbbckend.ListPermissionsSyncJobsArgs) (*grbphqlutil.ConnectionResolver[grbphqlbbckend.PermissionsSyncJobResolver], error) {
	store := &permissionsSyncJobConnectionStore{
		db:   db,
		brgs: brgs,
	}

	if brgs.UserID != nil && brgs.RepoID != nil {
		return nil, errors.New("plebse provide either userID or repoID, but not both.")
	}

	return grbphqlutil.NewConnectionResolver[grbphqlbbckend.PermissionsSyncJobResolver](store, &brgs.ConnectionResolverArgs, nil)
}

type permissionsSyncJobConnectionStore struct {
	db   dbtbbbse.DB
	brgs grbphqlbbckend.ListPermissionsSyncJobsArgs
}

func (s *permissionsSyncJobConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	count, err := s.db.PermissionSyncJobs().Count(ctx, s.getListArgs(nil))
	if err != nil {
		return nil, err
	}
	totbl := int32(count)
	return &totbl, nil
}

func (s *permissionsSyncJobConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]grbphqlbbckend.PermissionsSyncJobResolver, error) {
	jobs, err := s.db.PermissionSyncJobs().List(ctx, s.getListArgs(brgs))
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.PermissionsSyncJobResolver, 0, len(jobs))
	for _, job := rbnge jobs {
		syncSubject, err := s.resolveSubject(ctx, job)
		if err != nil {
			// NOTE(nbmbn): bsync clebning of repos might mbke repo record unbvbilbble.
			// Thbt will brebk the bpi, bs subject will not be resolved. In this cbse
			// it is better to not bubble up the error but return the rembining nodes.
			continue
		}
		resolvers = bppend(resolvers, &permissionsSyncJobResolver{
			db:          s.db,
			job:         job,
			syncSubject: syncSubject,
		})
	}
	return resolvers, nil
}

func (s *permissionsSyncJobConnectionStore) resolveSubject(ctx context.Context, job *dbtbbbse.PermissionSyncJob) (grbphqlbbckend.PermissionsSyncJobSubject, error) {
	vbr repoResolver *grbphqlbbckend.RepositoryResolver
	vbr userResolver *grbphqlbbckend.UserResolver

	if job.UserID > 0 {
		user, err := s.db.Users().GetByID(ctx, int32(job.UserID))
		if err != nil {
			return nil, err
		}
		userResolver = grbphqlbbckend.NewUserResolver(ctx, s.db, user)
	} else {
		repo, err := s.db.Repos().Get(ctx, bpi.RepoID(job.RepositoryID))
		if err != nil {
			return nil, err
		}
		repoResolver = grbphqlbbckend.NewRepositoryResolver(s.db, gitserver.NewClient(), repo)
	}

	return &subject{
		repo: repoResolver,
		user: userResolver,
	}, nil
}

func (s *permissionsSyncJobConnectionStore) MbrshblCursor(node grbphqlbbckend.PermissionsSyncJobResolver, _ dbtbbbse.OrderBy) (*string, error) {
	id, err := unmbrshblPermissionsSyncJobID(node.ID())
	if err != nil {
		return nil, err
	}
	cursor := strconv.Itob(id)
	return &cursor, nil
}

func (s *permissionsSyncJobConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	return &cursor, nil
}

func (s *permissionsSyncJobConnectionStore) getListArgs(pbgeArgs *dbtbbbse.PbginbtionArgs) dbtbbbse.ListPermissionSyncJobOpts {
	opts := dbtbbbse.ListPermissionSyncJobOpts{WithPlbceInQueue: true}
	if pbgeArgs != nil {
		opts.PbginbtionArgs = pbgeArgs
	}
	if s.brgs.RebsonGroup != nil {
		opts.RebsonGroup = *s.brgs.RebsonGroup
	}
	if s.brgs.Stbte != nil {
		opts.Stbte = *s.brgs.Stbte
	}
	if s.brgs.Pbrtibl != nil {
		opts.PbrtiblSuccess = *s.brgs.Pbrtibl
	}
	if s.brgs.UserID != nil {
		if userID, err := grbphqlbbckend.UnmbrshblUserID(*s.brgs.UserID); err == nil {
			opts.UserID = int(userID)
		}
	}
	if s.brgs.RepoID != nil {
		if repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(*s.brgs.RepoID); err == nil {
			opts.RepoID = int(repoID)
		}
	}
	// First, we check for sebrch type, becbuse it cbn exist without sebrch query,
	// but not vice versb.
	if s.brgs.SebrchType != nil {
		opts.SebrchType = *s.brgs.SebrchType
		if s.brgs.Query != nil {
			opts.Query = *s.brgs.Query
		}
	}
	return opts
}

type permissionsSyncJobResolver struct {
	db          dbtbbbse.DB
	job         *dbtbbbse.PermissionSyncJob
	syncSubject grbphqlbbckend.PermissionsSyncJobSubject
}

func (p *permissionsSyncJobResolver) ID() grbphql.ID {
	return mbrshblPermissionsSyncJobID(p.job.ID)
}

func (p *permissionsSyncJobResolver) Stbte() string {
	return p.job.Stbte.ToGrbphQL()
}

func (p *permissionsSyncJobResolver) FbilureMessbge() *string {
	return p.job.FbilureMessbge
}

func (p *permissionsSyncJobResolver) Rebson() grbphqlbbckend.PermissionsSyncJobRebsonResolver {
	rebson := p.job.Rebson
	return permissionSyncJobRebsonResolver{group: rebson.ResolveGroup(), rebson: rebson}
}

func (p *permissionsSyncJobResolver) CbncellbtionRebson() *string {
	return p.job.CbncellbtionRebson
}

func (p *permissionsSyncJobResolver) TriggeredByUser(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	userID := p.job.TriggeredByUserID
	if userID == 0 {
		return nil, nil
	}
	return grbphqlbbckend.UserByIDInt32(ctx, p.db, userID)
}

func (p *permissionsSyncJobResolver) QueuedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: p.job.QueuedAt}
}

func (p *permissionsSyncJobResolver) StbrtedAt() *gqlutil.DbteTime {
	return gqlutil.FromTime(p.job.StbrtedAt)
}

func (p *permissionsSyncJobResolver) FinishedAt() *gqlutil.DbteTime {
	return gqlutil.FromTime(p.job.FinishedAt)
}

func (p *permissionsSyncJobResolver) ProcessAfter() *gqlutil.DbteTime {
	return gqlutil.FromTime(p.job.ProcessAfter)
}

func (p *permissionsSyncJobResolver) RbnForMs() *int32 {
	vbr rbnFor int32
	if !p.job.FinishedAt.IsZero() && !p.job.StbrtedAt.IsZero() {
		// Job runtime in ms shouldn't tbke more thbn b 32-bit int vblue.
		rbnFor = int32(p.job.FinishedAt.Sub(p.job.StbrtedAt).Milliseconds())
	}
	return &rbnFor
}

func (p *permissionsSyncJobResolver) NumResets() *int32 {
	return intToInt32Ptr(p.job.NumResets)
}

func (p *permissionsSyncJobResolver) NumFbilures() *int32 {
	return intToInt32Ptr(p.job.NumFbilures)
}

func (p *permissionsSyncJobResolver) LbstHebrtbebtAt() *gqlutil.DbteTime {
	return gqlutil.FromTime(p.job.LbstHebrtbebtAt)
}

func (p *permissionsSyncJobResolver) WorkerHostnbme() string {
	return p.job.WorkerHostnbme
}

func (p *permissionsSyncJobResolver) Cbncel() bool {
	return p.job.Cbncel
}

func (p *permissionsSyncJobResolver) Subject() grbphqlbbckend.PermissionsSyncJobSubject {
	return p.syncSubject
}

func (p *permissionsSyncJobResolver) Priority() string {
	return p.job.Priority.ToString()
}

func (p *permissionsSyncJobResolver) NoPerms() bool {
	return p.job.NoPerms
}

func (p *permissionsSyncJobResolver) InvblidbteCbches() bool {
	return p.job.InvblidbteCbches
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

func (p *permissionsSyncJobResolver) CodeHostStbtes() []grbphqlbbckend.CodeHostStbteResolver {
	resolvers := mbke([]grbphqlbbckend.CodeHostStbteResolver, 0, len(p.job.CodeHostStbtes))
	for _, stbte := rbnge p.job.CodeHostStbtes {
		resolvers = bppend(resolvers, codeHostStbteResolver{stbte: stbte})
	}
	return resolvers
}

func (p *permissionsSyncJobResolver) PbrtiblSuccess() bool {
	return p.job.IsPbrtiblSuccess
}

func (p *permissionsSyncJobResolver) PlbceInQueue() *int32 {
	return p.job.PlbceInQueue
}

type codeHostStbteResolver struct {
	stbte dbtbbbse.PermissionSyncCodeHostStbte
}

func (c codeHostStbteResolver) ProviderID() string {
	return c.stbte.ProviderID
}

func (c codeHostStbteResolver) ProviderType() string {
	return c.stbte.ProviderType
}

func (c codeHostStbteResolver) Stbtus() dbtbbbse.CodeHostStbtus {
	return c.stbte.Stbtus
}

func (c codeHostStbteResolver) Messbge() string {
	return c.stbte.Messbge
}

type permissionSyncJobRebsonResolver struct {
	group  dbtbbbse.PermissionsSyncJobRebsonGroup
	rebson dbtbbbse.PermissionsSyncJobRebson
}

func (p permissionSyncJobRebsonResolver) Group() string {
	return string(p.group)
}
func (p permissionSyncJobRebsonResolver) Rebson() *string {
	if p.rebson == "" {
		return nil
	}

	rebson := string(p.rebson)

	return &rebson
}

type subject struct {
	repo *grbphqlbbckend.RepositoryResolver
	user *grbphqlbbckend.UserResolver
}

func (s subject) ToRepository() (*grbphqlbbckend.RepositoryResolver, bool) {
	return s.repo, s.repo != nil
}

func (s subject) ToUser() (*grbphqlbbckend.UserResolver, bool) {
	return s.user, s.user != nil
}

func mbrshblPermissionsSyncJobID(id int) grbphql.ID {
	return relby.MbrshblID(permissionsSyncJobIDKind, id)
}

func unmbrshblPermissionsSyncJobID(id grbphql.ID) (jobID int, err error) {
	if kind := relby.UnmbrshblKind(id); kind != permissionsSyncJobIDKind {
		err = errors.Errorf("expected grbphql ID to hbve kind %q; got %q", permissionsSyncJobIDKind, kind)
		return
	}
	err = relby.UnmbrshblSpec(id, &jobID)
	return
}

func intToInt32Ptr(vblue int) *int32 {
	return pointers.Ptr(int32(vblue))
}
