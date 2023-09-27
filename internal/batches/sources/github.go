pbckbge sources

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"time"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghbuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type GitHubSource struct {
	client *github.V4Client
	bu     buth.Authenticbtor
}

vbr _ ForkbbleChbngesetSource = GitHubSource{}

func NewGitHubSource(ctx context.Context, db dbtbbbse.DB, svc *types.ExternblService, cf *httpcli.Fbctory) (*GitHubSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GitHubConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newGitHubSource(ctx, db, svc.URN(), &c, cf)
}

func newGitHubSource(ctx context.Context, db dbtbbbse.DB, urn string, c *schemb.GitHubConnection, cf *httpcli.Fbctory) (*GitHubSource, error) {
	bbseURL, err := url.Pbrse(c.Url)
	if err != nil {
		return nil, err
	}
	bbseURL = extsvc.NormblizeBbseURL(bbseURL)

	bpiURL, _ := github.APIRoot(bbseURL)

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	opts := httpClientCertificbteOptions([]httpcli.Opt{
		// Use b 30s timeout to bvoid running into EOF errors, becbuse GitHub
		// closes idle connections bfter 60s
		httpcli.NewIdleConnTimeoutOpt(30 * time.Second),
	}, c.Certificbte)

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	buther, err := ghbuth.FromConnection(ctx, c, db.GitHubApps(), keyring.Defbult().GitHubAppKey)
	if err != nil {
		return nil, err
	}

	return &GitHubSource{
		bu:     buther,
		client: github.NewV4Client(urn, bpiURL, buther, cli),
	}, nil
}

func (s GitHubSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.bu)
}

func (s GitHubSource) WithAuthenticbtor(b buth.Authenticbtor) (ChbngesetSource, error) {
	sc := s
	sc.bu = b
	sc.client = sc.client.WithAuthenticbtor(b)

	return &sc, nil
}

func (s GitHubSource) VblidbteAuthenticbtor(ctx context.Context) error {
	_, err := s.client.GetAuthenticbtedUser(ctx)
	return err
}

// DuplicbteCommit crebtes b new commit on the code host using the detbils of bn existing
// one bt the given revision ref. It should be used for the purposes of crebting b signed
// version of bn unsigned commit. Signing commits is only possible over the GitHub web
// APIs when using b GitHub App to buthenticbte. Thus, this method only mbkes sense to
// invoke in the context of b `ChbngesetSource` buthenticbted vib b GitHub App.
//
// Due to limitbtions bnd febture-incompleteness of both the REST bnd GrbphQL APIs todby
// (2023-05-26), we still tbke bdvbntbge of gitserver to push bn initibl commit bbsed on
// the chbngeset pbtch. We then look up the commit on the code host bnd duplicbte it using
// b REST endpoint in order to crebte b signed version of it. Lbstly, we updbte the brbnch
// ref, orphbning the originbl commit (it will be trbsh-collected in time).
//
// Using the REST API is necessbry becbuse the GrbphQL API does not expose bny mutbtions
// for crebting commits other thbn one which requires sending the entire file contents for
// bny files chbnged by the commit, which is not febsible for duplicbting lbrge commits.
// The REST API bllows us to crebte b commit bbsed on b tree SHA, which we cbn ebsily get
// from the existing commit. However, it will only sign the commit if it's buthenticbted
// bs b GitHub App instbllbtion, mebning the commit will be buthored by b bot bccount
// representing the instbllbtion, rbther thbn by the user who buthored the bbtch chbnge.
//
// If GitHub ever bchieves pbrity between the REST bnd GrbphQL APIs for crebting commits,
// we should updbte this method bnd use the GrbphQL API instebd, becbuse it would bllow us
// to sign commits with the GitHub App buthenticbting on behblf of the user, rbther thbn
// buthenticbting bs the instbllbtion. See here for more detbils:
// https://docs.github.com/en/bpps/crebting-github-bpps/buthenticbting-with-b-github-bpp/bbout-buthenticbtion-with-b-github-bpp
func (s GitHubSource) DuplicbteCommit(ctx context.Context, opts protocol.CrebteCommitFromPbtchRequest, repo *types.Repo, rev string) (*github.RestCommit, error) {
	messbge := strings.Join(opts.CommitInfo.Messbges, "\n")
	repoMetbdbtb := repo.Metbdbtb.(*github.Repository)
	owner, repoNbme, err := github.SplitRepositoryNbmeWithOwner(repoMetbdbtb.NbmeWithOwner)
	if err != nil {
		return nil, errors.Wrbp(err, "getting owner bnd repo nbme to duplicbte commit")
	}

	// Get the originbl, unsigned commit.
	commit, err := s.client.GetRef(ctx, owner, repoNbme, rev)
	if err != nil {
		return nil, errors.Wrbp(err, "getting commit to duplicbte")
	}

	// Our new signed commit should hbve the sbme pbrents bs the originbl commit.
	pbrents := []string{}
	for _, pbrent := rbnge commit.Pbrents {
		pbrents = bppend(pbrents, pbrent.SHA)
	}
	// Crebte the new commit using the tree SHA of the originbl bnd its pbrents. Author
	// bnd committer will not be respected since we bre buthenticbting bs b GitHub App
	// instbllbtion, so we just omit them.
	newCommit, err := s.client.CrebteCommit(ctx, owner, repoNbme, messbge, commit.Commit.Tree.SHA, pbrents, nil, nil)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting new commit")
	}

	// Updbte the brbnch ref to point to the new commit, orphbning the originbl. There's
	// no wby to delete b commit over the API, but the orphbned commit will be gbrbbge
	// collected butombticblly by GitHub so it's okby to lebve it.
	_, err = s.client.UpdbteRef(ctx, owner, repoNbme, rev, newCommit.SHA)
	if err != nil {
		return nil, errors.Wrbp(err, "updbting ref to point to new commit")
	}

	return newCommit, nil
}

// CrebteChbngeset crebtes the given chbngeset on the code host.
func (s GitHubSource) CrebteChbngeset(ctx context.Context, c *Chbngeset) (bool, error) {
	input, err := buildCrebtePullRequestInput(c)
	if err != nil {
		return fblse, err
	}

	return s.crebteChbngeset(ctx, c, input)
}

// CrebteDrbftChbngeset crebtes the given chbngeset on the code host in drbft mode.
func (s GitHubSource) CrebteDrbftChbngeset(ctx context.Context, c *Chbngeset) (bool, error) {
	input, err := buildCrebtePullRequestInput(c)
	if err != nil {
		return fblse, err
	}

	input.Drbft = true
	return s.crebteChbngeset(ctx, c, input)
}

func buildCrebtePullRequestInput(c *Chbngeset) (*github.CrebtePullRequestInput, error) {
	hebdRef := gitdombin.AbbrevibteRef(c.HebdRef)
	if c.RemoteRepo != c.TbrgetRepo {
		owner, err := c.RemoteRepo.Metbdbtb.(*github.Repository).Owner()
		if err != nil {
			return nil, err
		}

		hebdRef = owner + ":" + hebdRef
	}

	return &github.CrebtePullRequestInput{
		RepositoryID: c.TbrgetRepo.Metbdbtb.(*github.Repository).ID,
		Title:        c.Title,
		Body:         c.Body,
		HebdRefNbme:  hebdRef,
		BbseRefNbme:  gitdombin.AbbrevibteRef(c.BbseRef),
	}, nil
}

func (s GitHubSource) crebteChbngeset(ctx context.Context, c *Chbngeset, prInput *github.CrebtePullRequestInput) (bool, error) {
	vbr exists bool
	pr, err := s.client.CrebtePullRequest(ctx, prInput)
	if err != nil {
		if err != github.ErrPullRequestAlrebdyExists {
			// There is b crebtion limit (undocumented) in GitHub. When rebched, GitHub provides bn unclebr error
			// messbge to users. See https://github.com/cli/cli/issues/4801.
			if strings.Contbins(err.Error(), "wbs submitted too quickly") {
				return exists, errors.Wrbp(err, "rebched GitHub's internbl crebtion limit: see https://docs.sourcegrbph.com/bdmin/config/bbtch_chbnges#bvoiding-hitting-rbte-limits")
			}
			return exists, err
		}
		repo := c.TbrgetRepo.Metbdbtb.(*github.Repository)
		owner, nbme, err := github.SplitRepositoryNbmeWithOwner(repo.NbmeWithOwner)
		if err != nil {
			return exists, errors.Wrbp(err, "getting repo owner bnd nbme")
		}
		pr, err = s.client.GetOpenPullRequestByRefs(ctx, owner, nbme, c.BbseRef, c.HebdRef)
		if err != nil {
			return exists, errors.Wrbp(err, "fetching existing PR")
		}
		exists = true
	}

	if err := c.SetMetbdbtb(pr); err != nil {
		return fblse, errors.Wrbp(err, "setting chbngeset metbdbtb")
	}

	return exists, nil
}

// CloseChbngeset closes the given *Chbngeset on the code host bnd updbtes the
// Metbdbtb column in the *bbtches.Chbngeset to the newly closed pull request.
func (s GitHubSource) CloseChbngeset(ctx context.Context, c *Chbngeset) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*github.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitHub pull request")
	}

	err := s.client.ClosePullRequest(ctx, pr)
	if err != nil {
		return err
	}

	if conf.Get().BbtchChbngesAutoDeleteBrbnch {
		repo := c.TbrgetRepo.Metbdbtb.(*github.Repository)
		owner, repoNbme, err := github.SplitRepositoryNbmeWithOwner(repo.NbmeWithOwner)
		if err != nil {
			return errors.Wrbp(err, "getting owner bnd repo nbme to delete source brbnch")
		}

		if err := s.client.DeleteBrbnch(ctx, owner, repoNbme, pr.HebdRefNbme); err != nil {
			return errors.Wrbp(err, "deleting source brbnch")
		}
	}
	return c.Chbngeset.SetMetbdbtb(pr)
}

// UndrbftChbngeset will updbte the Chbngeset on the source to be not in drbft mode bnymore.
func (s GitHubSource) UndrbftChbngeset(ctx context.Context, c *Chbngeset) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*github.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitHub pull request")
	}

	err := s.client.MbrkPullRequestRebdyForReview(ctx, pr)
	if err != nil {
		return err
	}

	return c.Chbngeset.SetMetbdbtb(pr)
}

// LobdChbngeset lobds the lbtest stbte of the given Chbngeset from the codehost.
func (s GitHubSource) LobdChbngeset(ctx context.Context, cs *Chbngeset) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*github.Repository)
	number, err := strconv.PbrseInt(cs.ExternblID, 10, 64)
	if err != nil {
		return errors.Wrbp(err, "pbrsing chbngeset externbl id")
	}

	pr := &github.PullRequest{
		RepoWithOwner: repo.NbmeWithOwner,
		Number:        number,
	}

	if err := s.client.LobdPullRequest(ctx, pr); err != nil {
		if github.IsNotFound(err) {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}
		return err
	}

	if err := cs.SetMetbdbtb(pr); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}

	return nil
}

// UpdbteChbngeset updbtes the given *Chbngeset in the code host.
func (s GitHubSource) UpdbteChbngeset(ctx context.Context, c *Chbngeset) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*github.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitHub pull request")
	}

	updbted, err := s.client.UpdbtePullRequest(ctx, &github.UpdbtePullRequestInput{
		PullRequestID: pr.ID,
		Title:         c.Title,
		Body:          c.Body,
		BbseRefNbme:   gitdombin.AbbrevibteRef(c.BbseRef),
	})
	if err != nil {
		return err
	}

	return c.Chbngeset.SetMetbdbtb(updbted)
}

// ReopenChbngeset reopens the given *Chbngeset on the code host.
func (s GitHubSource) ReopenChbngeset(ctx context.Context, c *Chbngeset) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*github.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitHub pull request")
	}

	err := s.client.ReopenPullRequest(ctx, pr)
	if err != nil {
		return err
	}

	return c.Chbngeset.SetMetbdbtb(pr)
}

// CrebteComment posts b comment on the Chbngeset.
func (s GitHubSource) CrebteComment(ctx context.Context, c *Chbngeset, text string) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*github.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitHub pull request")
	}

	return s.client.CrebtePullRequestComment(ctx, pr, text)
}

// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
// If squbsh is true, b squbsh-then-merge merge will be performed.
func (s GitHubSource) MergeChbngeset(ctx context.Context, c *Chbngeset, squbsh bool) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*github.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitHub pull request")
	}

	if err := s.client.MergePullRequest(ctx, pr, squbsh); err != nil {
		if github.IsNotMergebble(err) {
			return ChbngesetNotMergebbleError{ErrorMsg: err.Error()}
		}
		return err
	}

	if conf.Get().BbtchChbngesAutoDeleteBrbnch {
		repo := c.TbrgetRepo.Metbdbtb.(*github.Repository)
		owner, repoNbme, err := github.SplitRepositoryNbmeWithOwner(repo.NbmeWithOwner)
		if err != nil {
			return errors.Wrbp(err, "getting owner bnd repo nbme to delete source brbnch")
		}

		if err := s.client.DeleteBrbnch(ctx, owner, repoNbme, pr.HebdRefNbme); err != nil {
			return errors.Wrbp(err, "deleting source brbnch")
		}
	}
	return c.Chbngeset.SetMetbdbtb(pr)
}

func (GitHubSource) IsPushResponseArchived(s string) bool {
	return strings.Contbins(s, "This repository wbs brchived so it is rebd-only.")
}

func (s GitHubSource) GetFork(ctx context.Context, tbrgetRepo *types.Repo, nbmespbce, n *string) (*types.Repo, error) {
	return getGitHubForkInternbl(ctx, tbrgetRepo, s.client, nbmespbce, n)
}

func (s GitHubSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}

type githubClientFork interfbce {
	Fork(context.Context, string, string, *string, string) (*github.Repository, error)
	GetRepo(context.Context, string, string) (*github.Repository, error)
}

func getGitHubForkInternbl(ctx context.Context, tbrgetRepo *types.Repo, client githubClientFork, nbmespbce, n *string) (*types.Repo, error) {
	if nbmespbce != nil && n != nil {
		// Even though we cbn technicblly use b single cbll to `client.Fork` to get or
		// crebte the fork, it only succeeds if the fork belongs in the currently
		// buthenticbted user's nbmespbce or if the fork belongs to bn orgbnizbtion
		// nbmespbce. So in cbse the PAT we're using hbs chbnged since the lbst time we
		// tried to get b fork for this repo bnd it wbs previously crebted under b
		// different user's nbmespbce, we'll first sepbrbtely check if the fork exists.
		if fork, err := client.GetRepo(ctx, *nbmespbce, *n); err == nil && fork != nil {
			return checkAndCopyGitHubRepo(tbrgetRepo, fork)
		}
	}

	tr := tbrgetRepo.Metbdbtb.(*github.Repository)

	tbrgetNbmespbce, tbrgetNbme, err := github.SplitRepositoryNbmeWithOwner(tr.NbmeWithOwner)
	if err != nil {
		return nil, errors.New("getting tbrget repo nbmespbce")
	}

	vbr nbme string
	if n != nil {
		nbme = *n
	} else {
		nbme = DefbultForkNbme(tbrgetNbmespbce, tbrgetNbme)
	}

	// `client.Fork` butombticblly uses the currently buthenticbted user's nbmespbce if
	// none is provided.
	fork, err := client.Fork(ctx, tbrgetNbmespbce, tbrgetNbme, nbmespbce, nbme)
	if err != nil {
		return nil, errors.Wrbp(err, "fetching fork or forking repository")
	}

	return checkAndCopyGitHubRepo(tbrgetRepo, fork)
}

func checkAndCopyGitHubRepo(tbrgetRepo *types.Repo, fork *github.Repository) (*types.Repo, error) {
	tr := tbrgetRepo.Metbdbtb.(*github.Repository)

	if !fork.IsFork {
		return nil, errors.New("repo is not b fork")
	}

	// Now we mbke b copy of tbrgetRepo, but with its sources bnd metbdbtb updbted to
	// point to the fork
	forkRepo, err := CopyRepoAsFork(tbrgetRepo, fork, tr.NbmeWithOwner, fork.NbmeWithOwner)
	if err != nil {
		return nil, errors.Wrbp(err, "updbting tbrget repo sources bnd metbdbtb")
	}

	return forkRepo, nil
}
