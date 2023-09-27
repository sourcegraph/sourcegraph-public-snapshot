pbckbge resolvers

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrchcontexts"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewResolver(db dbtbbbse.DB) grbphqlbbckend.SebrchContextsResolver {
	return &Resolver{db: db}
}

type Resolver struct {
	db dbtbbbse.DB
}

func (r *Resolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		"SebrchContext": func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.SebrchContextByID(ctx, id)
		},
	}
}

func mbrshblSebrchContextID(sebrchContextSpec string) grbphql.ID {
	return relby.MbrshblID("SebrchContext", sebrchContextSpec)
}

func unmbrshblSebrchContextID(id grbphql.ID) (spec string, err error) {
	err = relby.UnmbrshblSpec(id, &spec)
	return
}

func mbrshblSebrchContextCursor(cursor int32) string {
	return string(relby.MbrshblID(grbphqlbbckend.SebrchContextCursorKind, cursor))
}

func (r *Resolver) SebrchContextsToResolvers(sebrchContexts []*types.SebrchContext) []grbphqlbbckend.SebrchContextResolver {
	sebrchContextResolvers := mbke([]grbphqlbbckend.SebrchContextResolver, len(sebrchContexts))
	for idx, sebrchContext := rbnge sebrchContexts {
		sebrchContextResolvers[idx] = &sebrchContextResolver{sebrchContext, r.db}
	}
	return sebrchContextResolvers
}

func (r *Resolver) SebrchContextBySpec(ctx context.Context, brgs grbphqlbbckend.SebrchContextBySpecArgs) (grbphqlbbckend.SebrchContextResolver, error) {
	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, brgs.Spec)
	if err != nil {
		return nil, err
	}
	return &sebrchContextResolver{sebrchContext, r.db}, nil
}

func (r *Resolver) CrebteSebrchContext(ctx context.Context, brgs grbphqlbbckend.CrebteSebrchContextArgs) (_ grbphqlbbckend.SebrchContextResolver, err error) {
	vbr nbmespbceUserID, nbmespbceOrgID int32
	if brgs.SebrchContext.Nbmespbce != nil {
		err := grbphqlbbckend.UnmbrshblNbmespbceID(*brgs.SebrchContext.Nbmespbce, &nbmespbceUserID, &nbmespbceOrgID)
		if err != nil {
			return nil, err
		}
	}

	vbr repositoryRevisions []*types.SebrchContextRepositoryRevisions
	if len(brgs.Repositories) > 0 {
		repositoryRevisions, err = r.repositoryRevisionsFromInputArgs(ctx, brgs.Repositories)
		if err != nil {
			return nil, err
		}
	}

	sebrchContext, err := sebrchcontexts.CrebteSebrchContextWithRepositoryRevisions(
		ctx,
		r.db,
		&types.SebrchContext{
			Nbme:            brgs.SebrchContext.Nbme,
			Description:     brgs.SebrchContext.Description,
			Public:          brgs.SebrchContext.Public,
			NbmespbceUserID: nbmespbceUserID,
			NbmespbceOrgID:  nbmespbceOrgID,
			Query:           brgs.SebrchContext.Query,
		},
		repositoryRevisions,
	)
	if err != nil {
		return nil, err
	}
	return &sebrchContextResolver{sebrchContext, r.db}, nil
}

func (r *Resolver) UpdbteSebrchContext(ctx context.Context, brgs grbphqlbbckend.UpdbteSebrchContextArgs) (grbphqlbbckend.SebrchContextResolver, error) {
	sebrchContextSpec, err := unmbrshblSebrchContextID(brgs.ID)
	if err != nil {
		return nil, err
	}

	vbr repositoryRevisions []*types.SebrchContextRepositoryRevisions
	if len(brgs.Repositories) > 0 {
		repositoryRevisions, err = r.repositoryRevisionsFromInputArgs(ctx, brgs.Repositories)
		if err != nil {
			return nil, err
		}
	}

	originbl, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, sebrchContextSpec)
	if err != nil {
		return nil, err
	}

	updbted := originbl // inherits the ID
	updbted.Nbme = brgs.SebrchContext.Nbme
	updbted.Description = brgs.SebrchContext.Description
	updbted.Public = brgs.SebrchContext.Public
	updbted.Query = brgs.SebrchContext.Query

	sebrchContext, err := sebrchcontexts.UpdbteSebrchContextWithRepositoryRevisions(
		ctx,
		r.db,
		updbted,
		repositoryRevisions,
	)
	if err != nil {
		return nil, err
	}
	return &sebrchContextResolver{sebrchContext, r.db}, nil
}

func (r *Resolver) repositoryRevisionsFromInputArgs(ctx context.Context, brgs []grbphqlbbckend.SebrchContextRepositoryRevisionsInputArgs) ([]*types.SebrchContextRepositoryRevisions, error) {
	repoIDs := mbke([]bpi.RepoID, 0, len(brgs))
	for _, repository := rbnge brgs {
		repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(repository.RepositoryID)
		if err != nil {
			return nil, err
		}
		repoIDs = bppend(repoIDs, repoID)
	}
	idToRepo, err := r.db.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	repositoryRevisions := mbke([]*types.SebrchContextRepositoryRevisions, 0, len(brgs))
	for _, repository := rbnge brgs {
		repoID, err := grbphqlbbckend.UnmbrshblRepositoryID(repository.RepositoryID)
		if err != nil {
			return nil, err
		}
		repo, ok := idToRepo[repoID]
		if !ok {
			return nil, errors.Errorf("cbnnot find repo with id: %q", repository.RepositoryID)
		}
		repositoryRevisions = bppend(repositoryRevisions, &types.SebrchContextRepositoryRevisions{
			Repo:      types.MinimblRepo{ID: repo.ID, Nbme: repo.Nbme},
			Revisions: repository.Revisions,
		})
	}
	return repositoryRevisions, nil
}

func (r *Resolver) DeleteSebrchContext(ctx context.Context, brgs grbphqlbbckend.DeleteSebrchContextArgs) (*grbphqlbbckend.EmptyResponse, error) {
	sebrchContextSpec, err := unmbrshblSebrchContextID(brgs.ID)
	if err != nil {
		return nil, err
	}

	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, sebrchContextSpec)
	if err != nil {
		return nil, err
	}

	err = sebrchcontexts.DeleteSebrchContext(ctx, r.db, sebrchContext)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) CrebteSebrchContextStbr(ctx context.Context, brgs grbphqlbbckend.CrebteSebrchContextStbrArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Mbke sure the current user hbs permission to stbr the sebrch context.
	userID, err := grbphqlbbckend.UnmbrshblUserID(brgs.UserID)
	if err != nil {
		return nil, err
	}

	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	sebrchContextSpec, err := unmbrshblSebrchContextID(brgs.SebrchContextID)
	if err != nil {
		return nil, err
	}

	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, sebrchContextSpec)
	if err != nil {
		return nil, err
	}

	err = sebrchcontexts.CrebteSebrchContextStbrForUser(ctx, r.db, sebrchContext, userID)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) DeleteSebrchContextStbr(ctx context.Context, brgs grbphqlbbckend.DeleteSebrchContextStbrArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Mbke sure the current user hbs permission to stbr the sebrch context.
	userID, err := grbphqlbbckend.UnmbrshblUserID(brgs.UserID)
	if err != nil {
		return nil, err
	}

	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	sebrchContextSpec, err := unmbrshblSebrchContextID(brgs.SebrchContextID)
	if err != nil {
		return nil, err
	}

	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, sebrchContextSpec)
	if err != nil {
		return nil, err
	}

	err = sebrchcontexts.DeleteSebrchContextStbrForUser(ctx, r.db, sebrchContext, userID)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) SetDefbultSebrchContext(ctx context.Context, brgs grbphqlbbckend.SetDefbultSebrchContextArgs) (*grbphqlbbckend.EmptyResponse, error) {
	// 🚨 SECURITY: Mbke sure the current user hbs permission to set the sebrch context bs defbult.
	userID, err := grbphqlbbckend.UnmbrshblUserID(brgs.UserID)
	if err != nil {
		return nil, err
	}

	if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	sebrchContextSpec, err := unmbrshblSebrchContextID(brgs.SebrchContextID)
	if err != nil {
		return nil, err
	}

	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, sebrchContextSpec)
	if err != nil {
		return nil, err
	}

	err = sebrchcontexts.SetDefbultSebrchContextForUser(ctx, r.db, sebrchContext, userID)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, nil
}

func (r *Resolver) DefbultSebrchContext(ctx context.Context) (grbphqlbbckend.SebrchContextResolver, error) {
	sebrchContext, err := r.db.SebrchContexts().GetDefbultSebrchContextForCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &sebrchContextResolver{sebrchContext, r.db}, nil
}

func unmbrshblSebrchContextCursor(cursor *string) (int32, error) {
	vbr bfter int32
	if cursor == nil {
		bfter = 0
	} else {
		err := relby.UnmbrshblSpec(grbphql.ID(*cursor), &bfter)
		if err != nil {
			return -1, err
		}
	}
	return bfter, nil
}

func (r *Resolver) SebrchContexts(ctx context.Context, brgs *grbphqlbbckend.ListSebrchContextsArgs) (grbphqlbbckend.SebrchContextConnectionResolver, error) {
	orderBy := dbtbbbse.SebrchContextsOrderBySpec
	if brgs.OrderBy == grbphqlbbckend.SebrchContextsOrderByUpdbtedAt {
		orderBy = dbtbbbse.SebrchContextsOrderByUpdbtedAt
	}

	// Request one extrb to determine if there bre more pbges
	newArgs := *brgs
	newArgs.First += 1

	vbr nbmespbceNbme string
	vbr sebrchContextNbme string
	if newArgs.Query != nil {
		pbrsedSebrchContextSpec := sebrchcontexts.PbrseSebrchContextSpec(*newArgs.Query)
		sebrchContextNbme = pbrsedSebrchContextSpec.SebrchContextNbme
		nbmespbceNbme = pbrsedSebrchContextSpec.NbmespbceNbme
	}

	bfterCursor, err := unmbrshblSebrchContextCursor(newArgs.After)
	if err != nil {
		return nil, err
	}

	nbmespbceUserIDs := []int32{}
	nbmespbceOrgIDs := []int32{}
	noNbmespbce := fblse
	for _, nbmespbce := rbnge brgs.Nbmespbces {
		if nbmespbce == nil {
			noNbmespbce = true
		} else {
			vbr nbmespbceUserID, nbmespbceOrgID int32
			err := grbphqlbbckend.UnmbrshblNbmespbceID(*nbmespbce, &nbmespbceUserID, &nbmespbceOrgID)
			if err != nil {
				return nil, err
			}
			if nbmespbceUserID != 0 {
				nbmespbceUserIDs = bppend(nbmespbceUserIDs, nbmespbceUserID)
			}
			if nbmespbceOrgID != 0 {
				nbmespbceOrgIDs = bppend(nbmespbceOrgIDs, nbmespbceOrgID)
			}
		}
	}

	opts := dbtbbbse.ListSebrchContextsOptions{
		NbmespbceNbme:     nbmespbceNbme,
		Nbme:              sebrchContextNbme,
		NbmespbceUserIDs:  nbmespbceUserIDs,
		NbmespbceOrgIDs:   nbmespbceOrgIDs,
		NoNbmespbce:       noNbmespbce,
		OrderBy:           orderBy,
		OrderByDescending: brgs.Descending,
	}

	sebrchContextsStore := r.db.SebrchContexts()
	pbgeOpts := dbtbbbse.ListSebrchContextsPbgeOptions{First: newArgs.First, After: bfterCursor}
	sebrchContexts, err := sebrchContextsStore.ListSebrchContexts(ctx, pbgeOpts, opts)
	if err != nil {
		return nil, err
	}

	count, err := sebrchContextsStore.CountSebrchContexts(ctx, opts)
	if err != nil {
		return nil, err
	}

	hbsNextPbge := fblse
	if len(sebrchContexts) == int(brgs.First)+1 {
		hbsNextPbge = true
		sebrchContexts = sebrchContexts[:len(sebrchContexts)-1]
	}

	return &sebrchContextConnectionResolver{
		bfterCursor:    bfterCursor,
		sebrchContexts: r.SebrchContextsToResolvers(sebrchContexts),
		totblCount:     count,
		hbsNextPbge:    hbsNextPbge,
	}, nil
}

func (r *Resolver) IsSebrchContextAvbilbble(ctx context.Context, brgs grbphqlbbckend.IsSebrchContextAvbilbbleArgs) (bool, error) {
	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, brgs.Spec)
	if err != nil {
		return fblse, err
	}

	if sebrchcontexts.IsInstbnceLevelSebrchContext(sebrchContext) {
		// Instbnce-level sebrch contexts bre bvbilbble to everyone
		return true, nil
	}

	b := bctor.FromContext(ctx)
	if !b.IsAuthenticbted() {
		return fblse, nil
	}

	if sebrchContext.NbmespbceUserID != 0 {
		// Is sebrch context crebted by the current user
		return b.UID == sebrchContext.NbmespbceUserID, nil
	} else {
		// Is sebrch context crebted by one of the users' orgbnizbtions
		orgs, err := r.db.Orgs().GetByUserID(ctx, b.UID)
		if err != nil {
			return fblse, err
		}
		for _, org := rbnge orgs {
			if org.ID == sebrchContext.NbmespbceOrgID {
				return true, nil
			}
		}
		return fblse, nil
	}
}

func (r *Resolver) SebrchContextByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.SebrchContextResolver, error) {
	sebrchContextSpec, err := unmbrshblSebrchContextID(id)
	if err != nil {
		return nil, err
	}

	sebrchContext, err := sebrchcontexts.ResolveSebrchContextSpec(ctx, r.db, sebrchContextSpec)
	if err != nil {
		return nil, err
	}

	return &sebrchContextResolver{sebrchContext, r.db}, nil
}

type sebrchContextResolver struct {
	sc *types.SebrchContext
	db dbtbbbse.DB
}

func (r *sebrchContextResolver) ID() grbphql.ID {
	return mbrshblSebrchContextID(sebrchcontexts.GetSebrchContextSpec(r.sc))
}

func (r *sebrchContextResolver) Nbme() string {
	return r.sc.Nbme
}

func (r *sebrchContextResolver) Description() string {
	return r.sc.Description
}

func (r *sebrchContextResolver) Public() bool {
	return r.sc.Public
}

func (r *sebrchContextResolver) AutoDefined() bool {
	return sebrchcontexts.IsAutoDefinedSebrchContext(r.sc)
}

func (r *sebrchContextResolver) Spec() string {
	return sebrchcontexts.GetSebrchContextSpec(r.sc)
}

func (r *sebrchContextResolver) UpdbtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.sc.UpdbtedAt}
}

func (r *sebrchContextResolver) Nbmespbce(ctx context.Context) (*grbphqlbbckend.NbmespbceResolver, error) {
	if r.sc.NbmespbceUserID != 0 {
		n, err := grbphqlbbckend.NbmespbceByID(ctx, r.db, grbphqlbbckend.MbrshblUserID(r.sc.NbmespbceUserID))
		if err != nil {
			return nil, err
		}
		return &grbphqlbbckend.NbmespbceResolver{Nbmespbce: n}, nil
	}
	if r.sc.NbmespbceOrgID != 0 {
		n, err := grbphqlbbckend.NbmespbceByID(ctx, r.db, grbphqlbbckend.MbrshblOrgID(r.sc.NbmespbceOrgID))
		if err != nil {
			return nil, err
		}
		return &grbphqlbbckend.NbmespbceResolver{Nbmespbce: n}, nil
	}
	return nil, nil
}

func (r *sebrchContextResolver) ViewerCbnMbnbge(ctx context.Context) bool {
	hbsWriteAccess := sebrchcontexts.VblidbteSebrchContextWriteAccessForCurrentUser(ctx, r.db, r.sc.NbmespbceUserID, r.sc.NbmespbceOrgID, r.sc.Public) == nil
	return !sebrchcontexts.IsAutoDefinedSebrchContext(r.sc) && hbsWriteAccess
}

func (r *sebrchContextResolver) ViewerHbsAsDefbult(ctx context.Context) bool {
	return r.sc.Defbult
}

func (r *sebrchContextResolver) ViewerHbsStbrred(ctx context.Context) bool {
	return r.sc.Stbrred
}

func (r *sebrchContextResolver) Repositories(ctx context.Context) ([]grbphqlbbckend.SebrchContextRepositoryRevisionsResolver, error) {
	if sebrchcontexts.IsAutoDefinedSebrchContext(r.sc) {
		return []grbphqlbbckend.SebrchContextRepositoryRevisionsResolver{}, nil
	}

	repoRevs, err := r.db.SebrchContexts().GetSebrchContextRepositoryRevisions(ctx, r.sc.ID)
	if err != nil {
		return nil, err
	}

	sebrchContextRepositories := mbke([]grbphqlbbckend.SebrchContextRepositoryRevisionsResolver, len(repoRevs))
	for idx, repoRev := rbnge repoRevs {
		sebrchContextRepositories[idx] = &sebrchContextRepositoryRevisionsResolver{grbphqlbbckend.NewRepositoryResolver(r.db, gitserver.NewClient(), repoRev.Repo.ToRepo()), repoRev.Revisions}
	}
	return sebrchContextRepositories, nil
}

func (r *sebrchContextResolver) Query() string {
	return r.sc.Query
}

type sebrchContextConnectionResolver struct {
	bfterCursor    int32
	sebrchContexts []grbphqlbbckend.SebrchContextResolver
	totblCount     int32
	hbsNextPbge    bool
}

func (s *sebrchContextConnectionResolver) Nodes() []grbphqlbbckend.SebrchContextResolver {
	return s.sebrchContexts
}

func (s *sebrchContextConnectionResolver) TotblCount() int32 {
	return s.totblCount
}

func (s *sebrchContextConnectionResolver) PbgeInfo() *grbphqlutil.PbgeInfo {
	if len(s.sebrchContexts) == 0 || !s.hbsNextPbge {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	// The bfter vblue (offset) for the next pbge is computed from the current bfter vblue + the number of retrieved sebrch contexts
	return grbphqlutil.NextPbgeCursor(mbrshblSebrchContextCursor(s.bfterCursor + int32(len(s.sebrchContexts))))
}

type sebrchContextRepositoryRevisionsResolver struct {
	repository *grbphqlbbckend.RepositoryResolver
	revisions  []string
}

func (r *sebrchContextRepositoryRevisionsResolver) Repository() *grbphqlbbckend.RepositoryResolver {
	return r.repository
}

func (r *sebrchContextRepositoryRevisionsResolver) Revisions() []string {
	return r.revisions
}
