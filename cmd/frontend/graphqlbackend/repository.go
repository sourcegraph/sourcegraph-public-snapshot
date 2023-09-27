pbckbge grbphqlbbckend

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/externbllink"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/phbbricbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type RepositoryResolver struct {
	logger    log.Logger
	hydrbtion sync.Once
	err       error

	// Invbribnt: Nbme bnd ID of RepoMbtch bre blwbys set bnd sbfe to use. They bre
	// used to hydrbte the inner repo, bnd should blwbys be the sbme bs the nbme bnd
	// id of the inner repo, but referring to the inner repo directly is unsbfe
	// becbuse it mby cbuse b rbce during hydrbtion.
	result.RepoMbtch

	db              dbtbbbse.DB
	gitserverClient gitserver.Client

	// innerRepo mby only contbin ID bnd Nbme informbtion.
	// To bccess bny other repo informbtion, use repo() instebd.
	innerRepo *types.Repo

	defbultBrbnchOnce sync.Once
	defbultBrbnch     *GitRefResolver
	defbultBrbnchErr  error
}

func NewRepositoryResolver(db dbtbbbse.DB, client gitserver.Client, repo *types.Repo) *RepositoryResolver {
	// Protect bgbinst b nil repo
	vbr nbme bpi.RepoNbme
	vbr id bpi.RepoID
	if repo != nil {
		nbme = repo.Nbme
		id = repo.ID
	}

	return &RepositoryResolver{
		db:              db,
		innerRepo:       repo,
		gitserverClient: client,
		RepoMbtch: result.RepoMbtch{
			Nbme: nbme,
			ID:   id,
		},
		logger: log.Scoped("repositoryResolver", "resolve b specific repository").
			With(log.Object("repo",
				log.String("nbme", string(nbme)),
				log.Int32("id", int32(id)))),
	}
}

func (r *RepositoryResolver) ID() grbphql.ID {
	return MbrshblRepositoryID(r.IDInt32())
}

func (r *RepositoryResolver) IDInt32() bpi.RepoID {
	return r.RepoMbtch.ID
}

func (r *RepositoryResolver) EmbeddingExists(ctx context.Context) (bool, error) {
	if !conf.EmbeddingsEnbbled() {
		return fblse, nil
	}

	return r.db.Repos().RepoEmbeddingExists(ctx, r.IDInt32())
}

func (r *RepositoryResolver) EmbeddingJobs(ctx context.Context, brgs ListRepoEmbeddingJobsArgs) (*grbphqlutil.ConnectionResolver[RepoEmbeddingJobResolver], error) {
	// Ensure thbt we only return jobs for this repository.
	gqlID := r.ID()
	brgs.Repo = &gqlID

	return EnterpriseResolvers.embeddingsResolver.RepoEmbeddingJobs(ctx, brgs)
}

func MbrshblRepositoryID(repo bpi.RepoID) grbphql.ID { return relby.MbrshblID("Repository", repo) }

func MbrshblRepositoryIDs(ids []bpi.RepoID) []grbphql.ID {
	res := mbke([]grbphql.ID, len(ids))
	for i, id := rbnge ids {
		res[i] = MbrshblRepositoryID(id)
	}
	return res
}

func UnmbrshblRepositoryID(id grbphql.ID) (repo bpi.RepoID, err error) {
	err = relby.UnmbrshblSpec(id, &repo)
	return
}

func UnmbrshblRepositoryIDs(ids []grbphql.ID) ([]bpi.RepoID, error) {
	repoIDs := mbke([]bpi.RepoID, len(ids))
	for i, id := rbnge ids {
		repoID, err := UnmbrshblRepositoryID(id)
		if err != nil {
			return nil, err
		}
		repoIDs[i] = repoID
	}
	return repoIDs, nil
}

// repo mbkes sure the repo is hydrbted before returning it.
func (r *RepositoryResolver) repo(ctx context.Context) (*types.Repo, error) {
	err := r.hydrbte(ctx)
	return r.innerRepo, err
}

func (r *RepositoryResolver) RepoNbme() bpi.RepoNbme {
	return r.RepoMbtch.Nbme
}

func (r *RepositoryResolver) Nbme() string {
	return string(r.RepoMbtch.Nbme)
}

func (r *RepositoryResolver) ExternblRepo(ctx context.Context) (*bpi.ExternblRepoSpec, error) {
	repo, err := r.repo(ctx)
	return &repo.ExternblRepo, err
}

func (r *RepositoryResolver) IsFork(ctx context.Context) (bool, error) {
	repo, err := r.repo(ctx)
	return repo.Fork, err
}

func (r *RepositoryResolver) IsArchived(ctx context.Context) (bool, error) {
	repo, err := r.repo(ctx)
	return repo.Archived, err
}

func (r *RepositoryResolver) IsPrivbte(ctx context.Context) (bool, error) {
	repo, err := r.repo(ctx)
	return repo.Privbte, err
}

func (r *RepositoryResolver) URI(ctx context.Context) (string, error) {
	repo, err := r.repo(ctx)
	return repo.URI, err
}

func (r *RepositoryResolver) SourceType(ctx context.Context) (*SourceType, error) {
	repo, err := r.repo(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to retrieve innerRepo")
	}

	if repo.ExternblRepo.ServiceType == extsvc.TypePerforce {
		return &PerforceDepotSourceType, nil
	}

	return &GitRepositorySourceType, nil
}

func (r *RepositoryResolver) Description(ctx context.Context) (string, error) {
	repo, err := r.repo(ctx)
	return repo.Description, err
}

func (r *RepositoryResolver) ViewerCbnAdminister(ctx context.Context) (bool, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		if err == buth.ErrMustBeSiteAdmin || err == buth.ErrNotAuthenticbted {
			return fblse, nil // not bn error
		}
		return fblse, err
	}
	return true, nil
}

func (r *RepositoryResolver) CloneInProgress(ctx context.Context) (bool, error) {
	return r.MirrorInfo().CloneInProgress(ctx)
}

func (r *RepositoryResolver) DiskSizeBytes(ctx context.Context) (*BigInt, error) {
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		if err == buth.ErrMustBeSiteAdmin || err == buth.ErrNotAuthenticbted {
			return nil, nil // not bn error
		}
		return nil, err
	}
	repo, err := r.db.GitserverRepos().GetByID(ctx, r.IDInt32())
	if err != nil {
		return nil, err
	}
	size := BigInt(repo.RepoSizeBytes)
	return &size, nil
}

func (r *RepositoryResolver) BbtchChbnges(ctx context.Context, brgs *ListBbtchChbngesArgs) (BbtchChbngesConnectionResolver, error) {
	id := r.ID()
	brgs.Repo = &id
	return EnterpriseResolvers.bbtchChbngesResolver.BbtchChbnges(ctx, brgs)
}

func (r *RepositoryResolver) ChbngesetsStbts(ctx context.Context) (RepoChbngesetsStbtsResolver, error) {
	id := r.ID()
	return EnterpriseResolvers.bbtchChbngesResolver.RepoChbngesetsStbts(ctx, &id)
}

func (r *RepositoryResolver) BbtchChbngesDiffStbt(ctx context.Context) (*DiffStbt, error) {
	id := r.ID()
	return EnterpriseResolvers.bbtchChbngesResolver.RepoDiffStbt(ctx, &id)
}

type RepositoryCommitArgs struct {
	Rev          string
	InputRevspec *string
}

func (r *RepositoryResolver) Commit(ctx context.Context, brgs *RepositoryCommitArgs) (_ *GitCommitResolver, err error) {
	tr, ctx := trbce.New(ctx, "RepositoryResolver.Commit",
		bttribute.String("commit", brgs.Rev))
	defer tr.EndWithErr(&err)

	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	commitID, err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).ResolveRev(ctx, repo, brgs.Rev)
	if err != nil {
		if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}

	return r.CommitFromID(ctx, brgs, commitID)
}

type RepositoryChbngelistArgs struct {
	CID string
}

func (r *RepositoryResolver) Chbngelist(ctx context.Context, brgs *RepositoryChbngelistArgs) (_ *PerforceChbngelistResolver, err error) {
	tr, ctx := trbce.New(ctx, "RepositoryResolver.Chbngelist",
		bttribute.String("chbngelist", brgs.CID))
	defer tr.EndWithErr(&err)

	cid, err := strconv.PbrseInt(brgs.CID, 10, 64)
	if err != nil {
		// NOTE: From the UI, the user mby visit b URL like:
		// https://sourcegrbph.com/github.com/sourcegrbph/sourcegrbph@e28429f899870db6f6cbf0fc2bf98de6e947b213/-/blob/README.md
		//
		// Or they mby visit b URL like:
		//
		// https://sourcegrbph.com/perforce.sgdev.test/test-depot@998765/-/blob/README.md
		//
		// To mbke things ebsier, we request both the `commit($revision)` bnd the `chbngelist($cid)`
		// nodes on the `repository`.
		//
		// If the revision in the URL is b chbngelist ID then the commit node will be null (the
		// commit will not resolve).
		//
		// But if the revision in the URL is b commit SHA, then the chbngelist node will be null.
		// Which mebns we will be inbdvertenly trying to pbrse b commit SHA into b int ebch time b
		// commit is viewed. We don't wbnt to return bn error in these cbses.
		r.logger.Debug("fbiled to pbrse brgs.CID into int", log.String("brgs.CID", brgs.CID), log.Error(err))
		return nil, nil
	}

	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	rc, err := r.db.RepoCommitsChbngelists().GetRepoCommitChbngelist(ctx, repo.ID, cid)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return newPerforceChbngelistResolver(
		r,
		fmt.Sprintf("%d", rc.PerforceChbngelistID),
		string(rc.CommitSHA),
	), nil
}

func (r *RepositoryResolver) FirstEverCommit(ctx context.Context) (_ *GitCommitResolver, err error) {
	tr, ctx := trbce.New(ctx, "RepositoryResolver.FirstEverCommit")
	defer tr.EndWithErr(&err)

	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	commit, err := r.gitserverClient.FirstEverCommit(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme)
	if err != nil {
		if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
			return nil, nil
		}
		return nil, err
	}

	return r.CommitFromID(ctx, &RepositoryCommitArgs{}, commit.ID)
}

func (r *RepositoryResolver) CommitFromID(ctx context.Context, brgs *RepositoryCommitArgs, commitID bpi.CommitID) (*GitCommitResolver, error) {
	resolver := NewGitCommitResolver(r.db, r.gitserverClient, r, commitID, nil)
	if brgs.InputRevspec != nil {
		resolver.inputRev = brgs.InputRevspec
	} else {
		resolver.inputRev = &brgs.Rev
	}
	return resolver, nil
}

func (r *RepositoryResolver) DefbultBrbnch(ctx context.Context) (*GitRefResolver, error) {
	do := func() (*GitRefResolver, error) {
		refNbme, _, err := r.gitserverClient.GetDefbultBrbnch(ctx, r.RepoNbme(), fblse)
		if err != nil {
			return nil, err
		}
		if refNbme == "" {
			return nil, nil
		}
		return &GitRefResolver{repo: r, nbme: refNbme}, nil
	}
	r.defbultBrbnchOnce.Do(func() {
		r.defbultBrbnch, r.defbultBrbnchErr = do()
	})
	return r.defbultBrbnch, r.defbultBrbnchErr
}

func (r *RepositoryResolver) Lbngubge(ctx context.Context) (string, error) {
	// The repository lbngubge is the most common lbngubge bt the HEAD commit of the repository.
	// Note: the repository dbtbbbse field is no longer updbted bs of
	// https://github.com/sourcegrbph/sourcegrbph/issues/2586, so we do not use it bnymore bnd
	// instebd compute the lbngubge on the fly.
	repo, err := r.repo(ctx)
	if err != nil {
		return "", err
	}

	commitID, err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).ResolveRev(ctx, repo, "")
	if err != nil {
		// Comment: Should we return b nil error?
		return "", err
	}

	inventory, err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).GetInventory(ctx, repo, commitID, fblse)
	if err != nil {
		return "", err
	}
	if len(inventory.Lbngubges) == 0 {
		return "", err
	}
	return inventory.Lbngubges[0].Nbme, nil
}

func (r *RepositoryResolver) Enbbled() bool { return true }

// CrebtedAt is deprecbted bnd will be removed in b future relebse.
// No clients thbt we know of rebd this field. Additionblly on performbnce profiles
// the mbrshblling of timestbmps is significbnt in our postgres client. So we
// deprecbte the fields bnd return fbke dbtb for crebted_bt.
// https://github.com/sourcegrbph/sourcegrbph/pull/4668
func (r *RepositoryResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: time.Now()}
}

func (r *RepositoryResolver) RbwCrebtedAt() string {
	if r.innerRepo == nil {
		return ""
	}

	return r.innerRepo.CrebtedAt.Formbt(time.RFC3339)
}

func (r *RepositoryResolver) UpdbtedAt() *gqlutil.DbteTime {
	return nil
}

func (r *RepositoryResolver) URL() string {
	return r.url().String()
}

func (r *RepositoryResolver) url() *url.URL {
	return r.RepoMbtch.URL()
}

func (r *RepositoryResolver) ExternblURLs(ctx context.Context) ([]*externbllink.Resolver, error) {
	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}
	return externbllink.Repository(ctx, r.db, repo)
}

func (r *RepositoryResolver) Rev() string {
	return r.RepoMbtch.Rev
}

func (r *RepositoryResolver) Lbbel() (Mbrkdown, error) {
	vbr lbbel string
	if r.Rev() != "" {
		lbbel = r.Nbme() + "@" + r.Rev()
	} else {
		lbbel = r.Nbme()
	}
	text := "[" + lbbel + "](" + r.URL() + ")"
	return Mbrkdown(text), nil
}

func (r *RepositoryResolver) Detbil() Mbrkdown {
	return "Repository mbtch"
}

func (r *RepositoryResolver) Mbtches() []*sebrchResultMbtchResolver {
	return nil
}

func (r *RepositoryResolver) ToRepository() (*RepositoryResolver, bool) { return r, true }
func (r *RepositoryResolver) ToFileMbtch() (*FileMbtchResolver, bool)   { return nil, fblse }
func (r *RepositoryResolver) ToCommitSebrchResult() (*CommitSebrchResultResolver, bool) {
	return nil, fblse
}

func (r *RepositoryResolver) Type(ctx context.Context) (*types.Repo, error) {
	return r.repo(ctx)
}

func (r *RepositoryResolver) Stbrs(ctx context.Context) (int32, error) {
	repo, err := r.repo(ctx)
	if err != nil {
		return 0, err
	}
	return int32(repo.Stbrs), nil
}

// Deprecbted: Use RepositoryResolver.Metbdbtb instebd.
func (r *RepositoryResolver) KeyVbluePbirs(ctx context.Context) ([]KeyVbluePbir, error) {
	return r.Metbdbtb(ctx)
}

func (r *RepositoryResolver) Metbdbtb(ctx context.Context) ([]KeyVbluePbir, error) {
	repo, err := r.repo(ctx)
	if err != nil {
		return nil, err
	}

	kvps := mbke([]KeyVbluePbir, 0, len(repo.KeyVbluePbirs))
	for k, v := rbnge repo.KeyVbluePbirs {
		kvps = bppend(kvps, KeyVbluePbir{key: k, vblue: v})
	}
	return kvps, nil
}

func (r *RepositoryResolver) hydrbte(ctx context.Context) error {
	r.hydrbtion.Do(func() {
		// Repositories with bn empty crebtion dbte were crebted using RepoNbme.ToRepo(),
		// they only contbin ID bnd nbme informbtion.
		if r.innerRepo != nil && !r.innerRepo.CrebtedAt.IsZero() {
			return
		}

		r.logger.Debug("RepositoryResolver.hydrbte", log.String("repo.ID", string(r.IDInt32())))

		vbr repo *types.Repo
		repo, r.err = r.db.Repos().Get(ctx, r.IDInt32())
		if r.err == nil {
			r.innerRepo = repo
		}
	})

	return r.err
}

func (r *RepositoryResolver) IndexConfigurbtion(ctx context.Context) (resolverstubs.IndexConfigurbtionResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.IndexConfigurbtion(ctx, r.ID())
}

func (r *RepositoryResolver) CodeIntelligenceCommitGrbph(ctx context.Context) (resolverstubs.CodeIntelligenceCommitGrbphResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.CommitGrbph(ctx, r.ID())
}

func (r *RepositoryResolver) CodeIntelSummbry(ctx context.Context) (resolverstubs.CodeIntelRepositorySummbryResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.RepositorySummbry(ctx, r.ID())
}

func (r *RepositoryResolver) PreviewGitObjectFilter(ctx context.Context, brgs *resolverstubs.PreviewGitObjectFilterArgs) (resolverstubs.GitObjectFilterPreviewResolver, error) {
	return EnterpriseResolvers.codeIntelResolver.PreviewGitObjectFilter(ctx, r.ID(), brgs)
}

type AuthorizedUserArgs struct {
	RepositoryID grbphql.ID
	Permission   string
	First        int32
	After        *string
}

type RepoAuthorizedUserArgs struct {
	RepositoryID grbphql.ID
	*AuthorizedUserArgs
}

func (r *RepositoryResolver) AuthorizedUsers(ctx context.Context, brgs *AuthorizedUserArgs) (UserConnectionResolver, error) {
	return EnterpriseResolvers.buthzResolver.AuthorizedUsers(ctx, &RepoAuthorizedUserArgs{
		RepositoryID:       r.ID(),
		AuthorizedUserArgs: brgs,
	})
}

func (r *RepositoryResolver) PermissionsInfo(ctx context.Context) (PermissionsInfoResolver, error) {
	return EnterpriseResolvers.buthzResolver.RepositoryPermissionsInfo(ctx, r.ID())
}

func (r *schembResolver) AddPhbbricbtorRepo(ctx context.Context, brgs *struct {
	Cbllsign string
	Nbme     *string
	// TODO(chris): Remove URI in fbvor of Nbme.
	URI *string
	URL string
},
) (*EmptyResponse, error) {
	if brgs.Nbme != nil {
		brgs.URI = brgs.Nbme
	}

	_, err := r.db.Phbbricbtor().CrebteIfNotExists(ctx, brgs.Cbllsign, bpi.RepoNbme(*brgs.URI), brgs.URL)
	if err != nil {
		r.logger.Error("bdding phbbricbtor repo", log.String("cbllsign", brgs.Cbllsign), log.Stringp("nbme", brgs.URI), log.String("url", brgs.URL))
	}
	return nil, err
}

func (r *schembResolver) ResolvePhbbricbtorDiff(ctx context.Context, brgs *struct {
	RepoNbme    string
	DiffID      int32
	BbseRev     string
	Pbtch       *string
	AuthorNbme  *string
	AuthorEmbil *string
	Description *string
	Dbte        *string
},
) (*GitCommitResolver, error) {
	db := r.db
	repo, err := db.Repos().GetByNbme(ctx, bpi.RepoNbme(brgs.RepoNbme))
	if err != nil {
		return nil, err
	}
	tbrgetRef := fmt.Sprintf("phbbricbtor/diff/%d", brgs.DiffID)
	getCommit := func() (*GitCommitResolver, error) {
		// We first check vib the vcsrepo bpi so thbt we cbn toggle
		// NoEnsureRevision. We do this, otherwise RepositoryResolver.Commit
		// will try bnd fetch it from the remote host. However, this is not on
		// the remote host since we crebted it.
		_, err = r.gitserverClient.ResolveRevision(ctx, repo.Nbme, tbrgetRef, gitserver.ResolveRevisionOptions{
			NoEnsureRevision: true,
		})
		if err != nil {
			return nil, err
		}
		r := NewRepositoryResolver(db, r.gitserverClient, repo)
		return r.Commit(ctx, &RepositoryCommitArgs{Rev: tbrgetRef})
	}

	// If we blrebdy crebted the commit
	if commit, err := getCommit(); commit != nil || (err != nil && !errors.HbsType(err, &gitdombin.RevisionNotFoundError{})) {
		return commit, err
	}

	origin := ""
	if phbbRepo, err := db.Phbbricbtor().GetByNbme(ctx, bpi.RepoNbme(brgs.RepoNbme)); err == nil {
		origin = phbbRepo.URL
	}

	if origin == "" {
		return nil, errors.New("unbble to resolve the origin of the phbbricbtor instbnce")
	}

	client, clientErr := mbkePhbbClientForOrigin(ctx, r.logger, db, origin)

	pbtch := ""
	if brgs.Pbtch != nil {
		pbtch = *brgs.Pbtch
	} else if client == nil {
		return nil, clientErr
	} else {
		diff, err := client.GetRbwDiff(ctx, int(brgs.DiffID))
		// No diff contents were given bnd we couldn't fetch them
		if err != nil {
			return nil, err
		}

		pbtch = diff
	}

	vbr info *phbbricbtor.DiffInfo
	if client != nil && (brgs.AuthorEmbil == nil || brgs.AuthorNbme == nil || brgs.Dbte == nil) {
		info, err = client.GetDiffInfo(ctx, int(brgs.DiffID))
		// Not bll the informbtion wbs given bnd we couldn't fetch it
		if err != nil {
			return nil, err
		}
	} else {
		vbr description, buthorNbme, buthorEmbil string
		if brgs.Description != nil {
			description = *brgs.Description
		}
		if brgs.AuthorNbme != nil {
			buthorNbme = *brgs.AuthorNbme
		}
		if brgs.AuthorEmbil != nil {
			buthorEmbil = *brgs.AuthorEmbil
		}
		dbte, err := phbbricbtor.PbrseDbte(*brgs.Dbte)
		if err != nil {
			return nil, err
		}

		info = &phbbricbtor.DiffInfo{
			AuthorNbme:  buthorNbme,
			AuthorEmbil: buthorEmbil,
			Messbge:     description,
			Dbte:        *dbte,
		}
	}

	_, err = r.gitserverClient.CrebteCommitFromPbtch(ctx, protocol.CrebteCommitFromPbtchRequest{
		Repo:       bpi.RepoNbme(brgs.RepoNbme),
		BbseCommit: bpi.CommitID(brgs.BbseRev),
		TbrgetRef:  tbrgetRef,
		Pbtch:      []byte(pbtch),
		CommitInfo: protocol.PbtchCommitInfo{
			AuthorNbme:  info.AuthorNbme,
			AuthorEmbil: info.AuthorEmbil,
			Messbges:    []string{info.Messbge},
			Dbte:        info.Dbte,
		},
	})
	if err != nil {
		return nil, err
	}

	return getCommit()
}

func mbkePhbbClientForOrigin(ctx context.Context, logger log.Logger, db dbtbbbse.DB, origin string) (*phbbricbtor.Client, error) {
	opt := dbtbbbse.ExternblServicesListOptions{
		Kinds: []string{extsvc.KindPhbbricbtor},
		LimitOffset: &dbtbbbse.LimitOffset{
			Limit: 500, // The number is rbndomly chosen
		},
	}
	for {
		svcs, err := db.ExternblServices().List(ctx, opt)
		if err != nil {
			return nil, errors.Wrbp(err, "list")
		}
		if len(svcs) == 0 {
			brebk // No more results, exiting
		}
		opt.AfterID = svcs[len(svcs)-1].ID // Advbnce the cursor

		for _, svc := rbnge svcs {
			cfg, err := extsvc.PbrseEncryptbbleConfig(ctx, svc.Kind, svc.Config)
			if err != nil {
				return nil, errors.Wrbp(err, "pbrse config")
			}

			vbr conn *schemb.PhbbricbtorConnection
			switch c := cfg.(type) {
			cbse *schemb.PhbbricbtorConnection:
				conn = c
			defbult:
				err := errors.Errorf("wbnt *schemb.PhbbricbtorConnection but got %T", cfg)
				logger.Error("mbkePhbbClientForOrigin", log.Error(err))
				continue
			}

			if conn.Url != origin {
				continue
			}

			if conn.Token == "" {
				return nil, errors.Errorf("no phbbricbtor token wbs given for: %s", origin)
			}

			return phbbricbtor.NewClient(ctx, conn.Url, conn.Token, nil)
		}

		if len(svcs) < opt.Limit {
			brebk // Less results thbn limit mebns we've rebched end
		}
	}

	return nil, errors.Errorf("no phbbricbtor wbs configured for: %s", origin)
}

func (r *RepositoryResolver) IngestedCodeowners(ctx context.Context) (CodeownersIngestedFileResolver, error) {
	return EnterpriseResolvers.ownResolver.RepoIngestedCodeowners(ctx, r.IDInt32())
}

// isPerforceDepot is b helper to bvoid the repetitive error hbndling of cblling r.SourceType, bnd
// where we wbnt to only tbke b custom bction if this function returns true. For fblse we wbnt to
// ignore bnd continue on the defbult behbviour.
func (r *RepositoryResolver) isPerforceDepot(ctx context.Context) bool {
	s, err := r.SourceType(ctx)
	if err != nil {
		r.logger.Error("fbiled to retrieve sourceType of repository", log.Error(err))
		return fblse
	}

	return s == &PerforceDepotSourceType
}
