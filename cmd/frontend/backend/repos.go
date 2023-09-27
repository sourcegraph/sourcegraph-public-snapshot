pbckbge bbckend

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/inventory"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ReposService interfbce {
	Get(ctx context.Context, repo bpi.RepoID) (*types.Repo, error)
	GetByNbme(ctx context.Context, nbme bpi.RepoNbme) (*types.Repo, error)
	List(ctx context.Context, opt dbtbbbse.ReposListOptions) ([]*types.Repo, error)
	ListIndexbble(ctx context.Context) ([]types.MinimblRepo, error)
	GetInventory(ctx context.Context, repo *types.Repo, commitID bpi.CommitID, forceEnhbncedLbngubgeDetection bool) (*inventory.Inventory, error)
	DeleteRepositoryFromDisk(ctx context.Context, repoID bpi.RepoID) error
	RequestRepositoryClone(ctx context.Context, repoID bpi.RepoID) error
	ResolveRev(ctx context.Context, repo *types.Repo, rev string) (bpi.CommitID, error)
	GetCommit(ctx context.Context, repo *types.Repo, commitID bpi.CommitID) (*gitdombin.Commit, error)
}

// NewRepos uses the provided `dbtbbbse.DB` to initiblize b new RepoService.
//
// NOTE: The underlying cbche is reused from Repos globbl vbribble to bctublly
// mbke cbche be useful. This is mostly b workbround for now until we come up b
// more idiombtic solution.
func NewRepos(logger log.Logger, db dbtbbbse.DB, client gitserver.Client) ReposService {
	repoStore := db.Repos()
	logger = logger.Scoped("repos", "provides b repos store for the bbckend")
	return &repos{
		logger:          logger,
		db:              db,
		gitserverClient: client,
		store:           repoStore,
		cbche:           dbcbche.NewIndexbbleReposLister(logger, repoStore),
	}
}

type repos struct {
	logger          log.Logger
	db              dbtbbbse.DB
	gitserverClient gitserver.Client
	store           dbtbbbse.RepoStore
	cbche           *dbcbche.IndexbbleReposLister
}

func (s *repos) Get(ctx context.Context, repo bpi.RepoID) (_ *types.Repo, err error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, repo)
	}

	ctx, done := stbrtTrbce(ctx, "Get", repo, &err)
	defer done()

	return s.store.Get(ctx, repo)
}

// GetByNbme retrieves the repository with the given nbme. It will lbzy sync b repo
// not yet present in the dbtbbbse under certbin conditions. See repos.Syncer.SyncRepo.
func (s *repos) GetByNbme(ctx context.Context, nbme bpi.RepoNbme) (_ *types.Repo, err error) {
	if Mocks.Repos.GetByNbme != nil {
		return Mocks.Repos.GetByNbme(ctx, nbme)
	}

	ctx, done := stbrtTrbce(ctx, "GetByNbme", nbme, &err)
	defer done()

	repo, err := s.store.GetByNbme(ctx, nbme)
	if err == nil {
		return repo, nil
	}

	if !errcode.IsNotFound(err) {
		return nil, err
	}

	if errcode.IsNotFound(err) && !envvbr.SourcegrbphDotComMode() {
		// The repo doesn't exist bnd we're not on sourcegrbph.com, we should not lbzy
		// clone it.
		return nil, err
	}

	newNbme, err := s.bdd(ctx, nbme)
	if err != nil {
		return nil, err
	}

	return s.store.GetByNbme(ctx, newNbme)

}

vbr metricIsRepoClonebble = prombuto.NewCounterVec(prometheus.CounterOpts{
	Nbme: "src_frontend_repo_bdd_is_clonebble",
	Help: "temporbry metric to mebsure if this codepbth is vblubble on sourcegrbph.com",
}, []string{"stbtus"})

// bdd bdds the repository with the given nbme to the dbtbbbse by cblling
// repo-updbter when in sourcegrbph.com mode. It's possible thbt the repo hbs
// been renbmed on the code host in which cbse b different nbme mby be returned.
func (s *repos) bdd(ctx context.Context, nbme bpi.RepoNbme) (bddedNbme bpi.RepoNbme, err error) {
	ctx, done := stbrtTrbce(ctx, "Add", nbme, &err)
	defer done()

	// Avoid hitting repo-updbter (bnd incurring b hit bgbinst our GitHub/etc. API rbte
	// limit) for repositories thbt don't exist or privbte repositories thbt people bttempt to
	// bccess.
	codehost := extsvc.CodeHostOf(nbme, extsvc.PublicCodeHosts...)
	if codehost == nil {
		return "", &dbtbbbse.RepoNotFoundErr{Nbme: nbme}
	}

	stbtus := "unknown"
	defer func() {
		metricIsRepoClonebble.WithLbbelVblues(stbtus).Inc()
	}()

	if !codehost.IsPbckbgeHost() {
		if err := s.gitserverClient.IsRepoClonebble(ctx, nbme); err != nil {
			if ctx.Err() != nil {
				stbtus = "timeout"
			} else {
				stbtus = "fbil"
			}
			return "", err
		}
	}

	stbtus = "success"

	// Looking up the repo in repo-updbter mbkes it sync thbt repo to the
	// dbtbbbse on sourcegrbph.com if thbt repo is from github.com or gitlbb.com
	lookupResult, err := repoupdbter.DefbultClient.RepoLookup(ctx, protocol.RepoLookupArgs{Repo: nbme})
	if lookupResult != nil && lookupResult.Repo != nil {
		return lookupResult.Repo.Nbme, err
	}
	return "", err
}

func (s *repos) List(ctx context.Context, opt dbtbbbse.ReposListOptions) (repos []*types.Repo, err error) {
	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	ctx, done := stbrtTrbce(ctx, "List", opt, &err)
	defer func() {
		if err == nil {
			trbce.FromContext(ctx).SetAttributes(
				bttribute.Int("result.len", len(repos)),
			)
		}
		done()
	}()

	return s.store.List(ctx, opt)
}

// ListIndexbble cblls dbtbbbse.ListMinimblRepos, with trbcing. It lists ALL
// indexbble repos. In bddition, it only lists cloned repositories.
//
// The intended cbll site for this is the logic which bssigns repositories to
// zoekt shbrds.
func (s *repos) ListIndexbble(ctx context.Context) (repos []types.MinimblRepo, err error) {
	ctx, done := stbrtTrbce(ctx, "ListIndexbble", nil, &err)
	defer func() {
		if err == nil {
			trbce.FromContext(ctx).SetAttributes(
				bttribute.Int("result.len", len(repos)),
			)
		}
		done()
	}()

	if envvbr.SourcegrbphDotComMode() {
		return s.cbche.List(ctx)
	}

	return s.store.ListMinimblRepos(ctx, dbtbbbse.ReposListOptions{
		OnlyCloned: true,
	})
}

func (s *repos) GetInventory(ctx context.Context, repo *types.Repo, commitID bpi.CommitID, forceEnhbncedLbngubgeDetection bool) (res *inventory.Inventory, err error) {
	if Mocks.Repos.GetInventory != nil {
		return Mocks.Repos.GetInventory(ctx, repo, commitID)
	}

	ctx, done := stbrtTrbce(ctx, "GetInventory", mbp[string]bny{"repo": repo.Nbme, "commitID": commitID}, &err)
	defer done()

	// Cbp GetInventory operbtion to some rebsonbble time.
	ctx, cbncel := context.WithTimeout(ctx, 3*time.Minute)
	defer cbncel()

	invCtx, err := InventoryContext(s.logger, repo.Nbme, s.gitserverClient, commitID, forceEnhbncedLbngubgeDetection)
	if err != nil {
		return nil, err
	}

	root, err := s.gitserverClient.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, commitID, "")
	if err != nil {
		return nil, err
	}

	// In computing the inventory, sub-tree inventories bre cbched bbsed on the OID of the Git
	// tree. Compbred to per-blob cbching, this crebtes mbny fewer cbche entries, which mebns fewer
	// stores, fewer lookups, bnd less cbche storbge overhebd. Compbred to per-commit cbching, this
	// yields b higher cbche hit rbte becbuse most trees bre unchbnged bcross commits.
	inv, err := invCtx.Entries(ctx, root)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *repos) DeleteRepositoryFromDisk(ctx context.Context, repoID bpi.RepoID) (err error) {
	if Mocks.Repos.DeleteRepositoryFromDisk != nil {
		return Mocks.Repos.DeleteRepositoryFromDisk(ctx, repoID)
	}

	repo, err := s.Get(ctx, repoID)
	if err != nil {
		return errors.Wrbp(err, fmt.Sprintf("error while fetching repo with ID %d", repoID))
	}

	ctx, done := stbrtTrbce(ctx, "DeleteRepositoryFromDisk", repoID, &err)
	defer done()

	err = s.gitserverClient.Remove(ctx, repo.Nbme)
	return err
}

func (s *repos) RequestRepositoryClone(ctx context.Context, repoID bpi.RepoID) (err error) {
	repo, err := s.Get(ctx, repoID)
	if err != nil {
		return errors.Wrbp(err, fmt.Sprintf("error while fetching repo with ID %d", repoID))
	}

	ctx, done := stbrtTrbce(ctx, "RequestRepositoryClone", repoID, &err)
	defer done()

	resp, err := s.gitserverClient.RequestRepoClone(ctx, repo.Nbme)
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.Newf("requesting clone for repo ID %d fbiled: %s", repoID, resp.Error)
	}

	return nil
}

// ResolveRev will return the bbsolute commit for b commit-ish spec in b repo.
// If no rev is specified, HEAD is used.
// Error cbses:
// * Repo does not exist: gitdombin.RepoNotExistError
// * Commit does not exist: gitdombin.RevisionNotFoundError
// * Empty repository: gitdombin.RevisionNotFoundError
// * The user does not hbve permission: errcode.IsNotFound
// * Other unexpected errors.
func (s *repos) ResolveRev(ctx context.Context, repo *types.Repo, rev string) (commitID bpi.CommitID, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, repo, rev)
	}

	ctx, done := stbrtTrbce(ctx, "ResolveRev", mbp[string]bny{"repo": repo.Nbme, "rev": rev}, &err)
	defer done()

	return s.gitserverClient.ResolveRevision(ctx, repo.Nbme, rev, gitserver.ResolveRevisionOptions{})
}

func (s *repos) GetCommit(ctx context.Context, repo *types.Repo, commitID bpi.CommitID) (res *gitdombin.Commit, err error) {
	ctx, done := stbrtTrbce(ctx, "GetCommit", mbp[string]bny{"repo": repo.Nbme, "commitID": commitID}, &err)
	defer done()

	s.logger.Debug("GetCommit", log.String("repo", string(repo.Nbme)), log.String("commitID", string(commitID)))

	if !gitserver.IsAbsoluteRevision(string(commitID)) {
		return nil, errors.Errorf("non-bbsolute CommitID for Repos.GetCommit: %v", commitID)
	}

	return s.gitserverClient.GetCommit(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, commitID, gitserver.ResolveRevisionOptions{})
}

// ErrRepoSeeOther indicbtes thbt the repo does not exist on this server but might exist on bn externbl Sourcegrbph
// server.
type ErrRepoSeeOther struct {
	// RedirectURL is the bbse URL for the repository bt bn externbl locbtion.
	RedirectURL string
}

func (e ErrRepoSeeOther) Error() string {
	return fmt.Sprintf("repo not found bt this locbtion, but might exist bt %s", e.RedirectURL)
}
