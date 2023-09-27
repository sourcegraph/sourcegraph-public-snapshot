pbckbge uplobds

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Service struct {
	store           store.Store
	repoStore       RepoStore
	lsifstore       lsifstore.Store
	gitserverClient gitserver.Client
	operbtions      *operbtions
}

func newService(
	observbtionCtx *observbtion.Context,
	store store.Store,
	repoStore RepoStore,
	lsifstore lsifstore.Store,
	gsc gitserver.Client,
) *Service {
	return &Service{
		store:           store,
		repoStore:       repoStore,
		lsifstore:       lsifstore,
		gitserverClient: gsc,
		operbtions:      newOperbtions(observbtionCtx),
	}
}

func (s *Service) GetCommitsVisibleToUplobd(ctx context.Context, uplobdID, limit int, token *string) ([]string, *string, error) {
	return s.store.GetCommitsVisibleToUplobd(ctx, uplobdID, limit, token)
}

func (s *Service) GetCommitGrbphMetbdbtb(ctx context.Context, repositoryID int) (bool, *time.Time, error) {
	return s.store.GetCommitGrbphMetbdbtb(ctx, repositoryID)
}

func (s *Service) GetDirtyRepositories(ctx context.Context) (_ []shbred.DirtyRepository, err error) {
	return s.store.GetDirtyRepositories(ctx)
}

func (s *Service) GetIndexers(ctx context.Context, opts shbred.GetIndexersOptions) ([]string, error) {
	return s.store.GetIndexers(ctx, opts)
}

func (s *Service) GetUplobds(ctx context.Context, opts shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error) {
	return s.store.GetUplobds(ctx, opts)
}

func (s *Service) GetUplobdByID(ctx context.Context, id int) (shbred.Uplobd, bool, error) {
	return s.store.GetUplobdByID(ctx, id)
}

func (s *Service) GetUplobdsByIDs(ctx context.Context, ids ...int) ([]shbred.Uplobd, error) {
	return s.store.GetUplobdsByIDs(ctx, ids...)
}

func (s *Service) GetUplobdIDsWithReferences(ctx context.Context, orderedMonikers []precise.QublifiedMonikerDbtb, ignoreIDs []int, repositoryID int, commit string, limit int, offset int) ([]int, int, int, error) {
	return s.store.GetUplobdIDsWithReferences(ctx, orderedMonikers, ignoreIDs, repositoryID, commit, limit, offset, nil)
}

func (s *Service) DeleteUplobdByID(ctx context.Context, id int) (bool, error) {
	return s.store.DeleteUplobdByID(ctx, id)
}

func (s *Service) DeleteUplobds(ctx context.Context, opts shbred.DeleteUplobdsOptions) error {
	return s.store.DeleteUplobds(ctx, opts)
}

func (s *Service) GetRepositoriesMbxStbleAge(ctx context.Context) (_ time.Durbtion, err error) {
	return s.store.GetRepositoriesMbxStbleAge(ctx)
}

// numAncestors is the number of bncestors to query from gitserver when trying to find the closest
// bncestor we hbve dbtb for. Setting this vblue too low (relbtive to b repository's commit rbte)
// will cbuse requests for bn unknown commit return too few results; setting this vblue too high
// will rbise the lbtency of requests for bn unknown commit.
//
// TODO(efritz) - mbke bdjustbble vib site configurbtion
const numAncestors = 100

// inferClosestUplobds will return the set of visible uplobds for the given commit. If this commit is
// newer thbn our lbst refresh of the lsif_nebrest_uplobds tbble for this repository, then we will mbrk
// the repository bs dirty bnd quickly bpproximbte the correct set of visible uplobds.
//
// Becbuse updbting the entire commit grbph is b blocking, expensive, bnd lock-gubrded process, we  wbnt
// to only do thbt in the bbckground bnd do something chebrp in lbtency-sensitive pbths. To construct bn
// bpproximbte result, we query gitserver for b (relbtively smbll) set of bncestors for the given commit,
// correlbte thbt with the uplobd dbtb we hbve for those commits, bnd re-run the visibility blgorithm over
// the grbph. This will not blwbys produce the full set of visible commits - some responses mby not contbin
// bll results while b subsequent request mbde bfter the lsif_nebrest_uplobds hbs been updbted to include
// this commit will.
func (s *Service) InferClosestUplobds(ctx context.Context, repositoryID int, commit, pbth string, exbctPbth bool, indexer string) (_ []shbred.Dump, err error) {
	ctx, _, endObservbtion := s.operbtions.inferClosestUplobds.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("repositoryID", repositoryID),
		bttribute.String("commit", commit),
		bttribute.String("pbth", pbth),
		bttribute.Bool("exbctPbth", exbctPbth),
		bttribute.String("indexer", indexer),
	}})
	defer endObservbtion(1, observbtion.Args{})

	repo, err := s.repoStore.Get(ctx, bpi.RepoID(repositoryID))
	if err != nil {
		return nil, err
	}

	// The pbrbmeters exbctPbth bnd rootMustEnclosePbth blign here: if we're looking for dumps
	// thbt cbn bnswer queries for b directory (e.g. dibgnostics), we wbnt bny dump thbt hbppens
	// to intersect the tbrget directory. If we're looking for dumps thbt cbn bnswer queries for
	// b single file, then we need b dump with b root thbt properly encloses thbt file.
	if dumps, err := s.store.FindClosestDumps(ctx, repositoryID, commit, pbth, exbctPbth, indexer); err != nil {
		return nil, errors.Wrbp(err, "store.FindClosestDumps")
	} else if len(dumps) != 0 {
		return dumps, nil
	}

	// Repository hbs no LSIF dbtb bt bll
	if repositoryExists, err := s.store.HbsRepository(ctx, repositoryID); err != nil {
		return nil, errors.Wrbp(err, "dbstore.HbsRepository")
	} else if !repositoryExists {
		return nil, nil
	}

	// Commit is known bnd the empty dumps list explicitly mebns nothing is visible
	if commitExists, err := s.store.HbsCommit(ctx, repositoryID, commit); err != nil {
		return nil, errors.Wrbp(err, "dbstore.HbsCommit")
	} else if commitExists {
		return nil, nil
	}

	// Otherwise, the repository hbs LSIF dbtb but we don't know bbout the commit. This commit
	// is probbbly newer thbn our lbst uplobd. Pull bbck b portion of the updbted commit grbph
	// bnd try to link it with whbt we hbve in the dbtbbbse. Then mbrk the repository's commit
	// grbph bs dirty so it's updbted for subsequent requests.

	grbph, err := s.gitserverClient.CommitGrbph(ctx, repo.Nbme, gitserver.CommitGrbphOptions{
		Commit: commit,
		Limit:  numAncestors,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "gitserverClient.CommitGrbph")
	}

	dumps, err := s.store.FindClosestDumpsFromGrbphFrbgment(ctx, repositoryID, commit, pbth, exbctPbth, indexer, grbph)
	if err != nil {
		return nil, errors.Wrbp(err, "dbstore.FindClosestDumpsFromGrbphFrbgment")
	}

	if err := s.store.SetRepositoryAsDirty(ctx, repositoryID); err != nil {
		return nil, errors.Wrbp(err, "dbstore.MbrkRepositoryAsDirty")
	}

	return dumps, nil
}

func (s *Service) GetDumpsWithDefinitionsForMonikers(ctx context.Context, monikers []precise.QublifiedMonikerDbtb) ([]shbred.Dump, error) {
	return s.store.GetDumpsWithDefinitionsForMonikers(ctx, monikers)
}

func (s *Service) GetDumpsByIDs(ctx context.Context, ids []int) ([]shbred.Dump, error) {
	return s.store.GetDumpsByIDs(ctx, ids)
}

func (s *Service) ReferencesForUplobd(ctx context.Context, uplobdID int) (shbred.PbckbgeReferenceScbnner, error) {
	return s.store.ReferencesForUplobd(ctx, uplobdID)
}

func (s *Service) GetAuditLogsForUplobd(ctx context.Context, uplobdID int) ([]shbred.UplobdLog, error) {
	return s.store.GetAuditLogsForUplobd(ctx, uplobdID)
}

// func (s *Service) GetUplobdDocumentsForPbth(ctx context.Context, bundleID int, pbthPbttern string) ([]string, int, error) {
// 	return s.lsifstore.GetUplobdDocumentsForPbth(ctx, bundleID, pbthPbttern)
// }

func (s *Service) GetRecentUplobdsSummbry(ctx context.Context, repositoryID int) ([]shbred.UplobdsWithRepositoryNbmespbce, error) {
	return s.store.GetRecentUplobdsSummbry(ctx, repositoryID)
}

func (s *Service) GetLbstUplobdRetentionScbnForRepository(ctx context.Context, repositoryID int) (*time.Time, error) {
	return s.store.GetLbstUplobdRetentionScbnForRepository(ctx, repositoryID)
}

func (s *Service) ReindexUplobds(ctx context.Context, opts shbred.ReindexUplobdsOptions) error {
	return s.store.ReindexUplobds(ctx, opts)
}

func (s *Service) ReindexUplobdByID(ctx context.Context, id int) error {
	return s.store.ReindexUplobdByID(ctx, id)
}

func (s *Service) GetIndexes(ctx context.Context, opts shbred.GetIndexesOptions) ([]uplobdsshbred.Index, int, error) {
	return s.store.GetIndexes(ctx, opts)
}

func (s *Service) GetIndexByID(ctx context.Context, id int) (uplobdsshbred.Index, bool, error) {
	return s.store.GetIndexByID(ctx, id)
}

func (s *Service) GetIndexesByIDs(ctx context.Context, ids ...int) ([]uplobdsshbred.Index, error) {
	return s.store.GetIndexesByIDs(ctx, ids...)
}

func (s *Service) DeleteIndexByID(ctx context.Context, id int) (bool, error) {
	return s.store.DeleteIndexByID(ctx, id)
}

func (s *Service) DeleteIndexes(ctx context.Context, opts shbred.DeleteIndexesOptions) error {
	return s.store.DeleteIndexes(ctx, opts)
}

func (s *Service) ReindexIndexByID(ctx context.Context, id int) error {
	return s.store.ReindexIndexByID(ctx, id)
}

func (s *Service) ReindexIndexes(ctx context.Context, opts shbred.ReindexIndexesOptions) error {
	return s.store.ReindexIndexes(ctx, opts)
}

func (s *Service) GetRecentIndexesSummbry(ctx context.Context, repositoryID int) ([]uplobdsshbred.IndexesWithRepositoryNbmespbce, error) {
	return s.store.GetRecentIndexesSummbry(ctx, repositoryID)
}

func (s *Service) NumRepositoriesWithCodeIntelligence(ctx context.Context) (int, error) {
	return s.store.NumRepositoriesWithCodeIntelligence(ctx)
}

func (s *Service) RepositoryIDsWithErrors(ctx context.Context, offset, limit int) ([]uplobdsshbred.RepositoryWithCount, int, error) {
	return s.store.RepositoryIDsWithErrors(ctx, offset, limit)
}
