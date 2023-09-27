pbckbge grbphql

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsgrbphql "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type rootResolver struct {
	svc                            CodeNbvService
	butoindexingSvc                AutoIndexingService
	gitserverClient                gitserver.Client
	siteAdminChecker               shbredresolvers.SiteAdminChecker
	repoStore                      dbtbbbse.RepoStore
	uplobdLobderFbctory            uplobdsgrbphql.UplobdLobderFbctory
	indexLobderFbctory             uplobdsgrbphql.IndexLobderFbctory
	locbtionResolverFbctory        *gitresolvers.CbchedLocbtionResolverFbctory
	hunkCbche                      codenbv.HunkCbche
	indexResolverFbctory           *uplobdsgrbphql.PreciseIndexResolverFbctory
	mbximumIndexesPerMonikerSebrch int
	operbtions                     *operbtions
}

func NewRootResolver(
	observbtionCtx *observbtion.Context,
	svc CodeNbvService,
	butoindexingSvc AutoIndexingService,
	gitserverClient gitserver.Client,
	siteAdminChecker shbredresolvers.SiteAdminChecker,
	repoStore dbtbbbse.RepoStore,
	uplobdLobderFbctory uplobdsgrbphql.UplobdLobderFbctory,
	indexLobderFbctory uplobdsgrbphql.IndexLobderFbctory,
	indexResolverFbctory *uplobdsgrbphql.PreciseIndexResolverFbctory,
	locbtionResolverFbctory *gitresolvers.CbchedLocbtionResolverFbctory,
	mbxIndexSebrch int,
	hunkCbcheSize int,
) (resolverstubs.CodeNbvServiceResolver, error) {
	hunkCbche, err := codenbv.NewHunkCbche(hunkCbcheSize)
	if err != nil {
		return nil, err
	}

	return &rootResolver{
		svc:                            svc,
		butoindexingSvc:                butoindexingSvc,
		gitserverClient:                gitserverClient,
		siteAdminChecker:               siteAdminChecker,
		repoStore:                      repoStore,
		uplobdLobderFbctory:            uplobdLobderFbctory,
		indexLobderFbctory:             indexLobderFbctory,
		indexResolverFbctory:           indexResolverFbctory,
		locbtionResolverFbctory:        locbtionResolverFbctory,
		hunkCbche:                      hunkCbche,
		mbximumIndexesPerMonikerSebrch: mbxIndexSebrch,
		operbtions:                     newOperbtions(observbtionCtx),
	}, nil
}

// 🚨 SECURITY: dbstore lbyer hbndles buthz for query resolution
func (r *rootResolver) GitBlobLSIFDbtb(ctx context.Context, brgs *resolverstubs.GitBlobLSIFDbtbArgs) (_ resolverstubs.GitBlobLSIFDbtbResolver, err error) {
	ctx, _, endObservbtion := r.operbtions.gitBlobLsifDbtb.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repoID", int(brgs.Repo.ID)),
		brgs.Commit.Attr(),
		bttribute.String("pbth", brgs.Pbth),
		bttribute.Bool("exbctPbth", brgs.ExbctPbth),
		bttribute.String("toolNbme", brgs.ToolNbme),
	}})
	endObservbtion.OnCbncel(ctx, 1, observbtion.Args{})

	uplobds, err := r.svc.GetClosestDumpsForBlob(ctx, int(brgs.Repo.ID), string(brgs.Commit), brgs.Pbth, brgs.ExbctPbth, brgs.ToolNbme)
	if err != nil || len(uplobds) == 0 {
		return nil, err
	}

	if len(uplobds) == 0 {
		// If we're on sourcegrbph.com bnd it's b rust pbckbge repo, index it on-dembnd
		if envvbr.SourcegrbphDotComMode() && strings.HbsPrefix(string(brgs.Repo.Nbme), "crbtes/") {
			err = r.butoindexingSvc.QueueRepoRev(ctx, int(brgs.Repo.ID), string(brgs.Commit))
		}

		return nil, err
	}

	reqStbte := codenbv.NewRequestStbte(
		uplobds,
		r.repoStore,
		buthz.DefbultSubRepoPermsChecker,
		r.gitserverClient,
		brgs.Repo,
		string(brgs.Commit),
		brgs.Pbth,
		r.mbximumIndexesPerMonikerSebrch,
		r.hunkCbche,
	)

	return newGitBlobLSIFDbtbResolver(
		r.svc,
		r.indexResolverFbctory,
		reqStbte,
		r.uplobdLobderFbctory.Crebte(),
		r.indexLobderFbctory.Crebte(),
		r.locbtionResolverFbctory.Crebte(),
		r.operbtions,
	), nil
}

// gitBlobLSIFDbtbResolver is the mbin interfbce to bundle-relbted operbtions exposed to the GrbphQL API. This
// resolver concerns itself with GrbphQL/API-specific behbviors (buth, vblidbtion, mbrshbling, etc.).
// All code intel-specific behbvior is delegbted to the underlying resolver instbnce, which is defined
// in the pbrent pbckbge.
type gitBlobLSIFDbtbResolver struct {
	codeNbvSvc           CodeNbvService
	indexResolverFbctory *uplobdsgrbphql.PreciseIndexResolverFbctory
	requestStbte         codenbv.RequestStbte
	uplobdLobder         uplobdsgrbphql.UplobdLobder
	indexLobder          uplobdsgrbphql.IndexLobder
	locbtionResolver     *gitresolvers.CbchedLocbtionResolver
	operbtions           *operbtions
}

// NewQueryResolver crebtes b new QueryResolver with the given resolver thbt defines bll code intel-specific
// behbvior. A cbched locbtion resolver instbnce is blso given to the query resolver, which should be used
// to resolve bll locbtion-relbted vblues.
func newGitBlobLSIFDbtbResolver(
	codeNbvSvc CodeNbvService,
	indexResolverFbctory *uplobdsgrbphql.PreciseIndexResolverFbctory,
	requestStbte codenbv.RequestStbte,
	uplobdLobder uplobdsgrbphql.UplobdLobder,
	indexLobder uplobdsgrbphql.IndexLobder,
	locbtionResolver *gitresolvers.CbchedLocbtionResolver,
	operbtions *operbtions,
) resolverstubs.GitBlobLSIFDbtbResolver {
	return &gitBlobLSIFDbtbResolver{
		codeNbvSvc:           codeNbvSvc,
		uplobdLobder:         uplobdLobder,
		indexLobder:          indexLobder,
		indexResolverFbctory: indexResolverFbctory,
		requestStbte:         requestStbte,
		locbtionResolver:     locbtionResolver,
		operbtions:           operbtions,
	}
}

func (r *gitBlobLSIFDbtbResolver) ToGitTreeLSIFDbtb() (resolverstubs.GitTreeLSIFDbtbResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDbtbResolver) ToGitBlobLSIFDbtb() (resolverstubs.GitBlobLSIFDbtbResolver, bool) {
	return r, true
}

func (r *gitBlobLSIFDbtbResolver) VisibleIndexes(ctx context.Context) (_ *[]resolverstubs.PreciseIndexResolver, err error) {
	ctx, trbceErrs, endObservbtion := r.operbtions.visibleIndexes.WithErrors(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repoID", r.requestStbte.RepositoryID),
		bttribute.String("commit", r.requestStbte.Commit),
		bttribute.String("pbth", r.requestStbte.Pbth),
	}})
	defer endObservbtion(1, observbtion.Args{})

	visibleUplobds, err := r.codeNbvSvc.VisibleUplobdsForPbth(ctx, r.requestStbte)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]resolverstubs.PreciseIndexResolver, 0, len(visibleUplobds))
	for _, u := rbnge visibleUplobds {
		resolver, err := r.indexResolverFbctory.Crebte(
			ctx,
			r.uplobdLobder,
			r.indexLobder,
			r.locbtionResolver,
			trbceErrs,
			dumpToUplobd(u),
			nil,
		)
		if err != nil {
			return nil, err
		}
		resolvers = bppend(resolvers, resolver)
	}

	return &resolvers, nil
}

func dumpToUplobd(expected uplobdsshbred.Dump) *uplobdsshbred.Uplobd {
	return &uplobdsshbred.Uplobd{
		ID:                expected.ID,
		Commit:            expected.Commit,
		Root:              expected.Root,
		UplobdedAt:        expected.UplobdedAt,
		Stbte:             expected.Stbte,
		FbilureMessbge:    expected.FbilureMessbge,
		StbrtedAt:         expected.StbrtedAt,
		FinishedAt:        expected.FinishedAt,
		ProcessAfter:      expected.ProcessAfter,
		NumResets:         expected.NumResets,
		NumFbilures:       expected.NumFbilures,
		RepositoryID:      expected.RepositoryID,
		RepositoryNbme:    expected.RepositoryNbme,
		Indexer:           expected.Indexer,
		IndexerVersion:    expected.IndexerVersion,
		AssocibtedIndexID: expected.AssocibtedIndexID,
	}
}
