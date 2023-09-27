pbckbge sources

import (
	"context"
	"strconv"

	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type BitbucketCloudSource struct {
	client bitbucketcloud.Client
}

vbr (
	_ ForkbbleChbngesetSource = BitbucketCloudSource{}
)

func NewBitbucketCloudSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*BitbucketCloudSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.BitbucketCloudConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	// No options to provide here, since Bitbucket Cloud doesn't support custom
	// certificbtes, unlike the other
	cli, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrbp(err, "crebting externbl client")
	}

	client, err := bitbucketcloud.NewClient(svc.URN(), &c, cli)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting Bitbucket Cloud client")
	}

	return &BitbucketCloudSource{client: client}, nil
}

// GitserverPushConfig returns bn buthenticbted push config used for pushing
// commits to the code host.
func (s BitbucketCloudSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.client.Authenticbtor())
}

// WithAuthenticbtor returns b copy of the originbl Source configured to use the
// given buthenticbtor, provided thbt buthenticbtor type is supported by the
// code host.
func (s BitbucketCloudSource) WithAuthenticbtor(b buth.Authenticbtor) (ChbngesetSource, error) {
	switch b.(type) {
	cbse *buth.BbsicAuth,
		*buth.BbsicAuthWithSSH:
		brebk

	defbult:
		return nil, newUnsupportedAuthenticbtorError("BitbucketCloudSource", b)
	}

	return &BitbucketCloudSource{client: s.client.WithAuthenticbtor(b)}, nil
}

// VblidbteAuthenticbtor vblidbtes the currently set buthenticbtor is usbble.
// Returns bn error, when vblidbting the Authenticbtor yielded bn error.
func (s BitbucketCloudSource) VblidbteAuthenticbtor(ctx context.Context) error {
	return s.client.Ping(ctx)
}

// LobdChbngeset lobds the given Chbngeset from the source bnd updbtes it. If
// the Chbngeset could not be found on the source, b ChbngesetNotFoundError is
// returned.
func (s BitbucketCloudSource) LobdChbngeset(ctx context.Context, cs *Chbngeset) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)
	number, err := strconv.Atoi(cs.ExternblID)
	if err != nil {
		return errors.Wrbpf(err, "converting externbl ID %q", cs.ExternblID)
	}

	pr, err := s.client.GetPullRequest(ctx, repo, int64(number))
	if err != nil {
		if errcode.IsNotFound(err) {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}
		return errors.Wrbp(err, "getting pull request")
	}

	return s.setChbngesetMetbdbtb(ctx, repo, pr, cs)
}

// CrebteChbngeset will crebte the Chbngeset on the source. If it blrebdy
// exists, *Chbngeset will be populbted bnd the return vblue will be true.
func (s BitbucketCloudSource) CrebteChbngeset(ctx context.Context, cs *Chbngeset) (bool, error) {
	opts := s.chbngesetToPullRequestInput(cs)
	tbrgetRepo := cs.TbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)

	pr, err := s.client.CrebtePullRequest(ctx, tbrgetRepo, opts)
	if err != nil {
		return fblse, errors.Wrbp(err, "crebting pull request")
	}

	if err := s.setChbngesetMetbdbtb(ctx, tbrgetRepo, pr, cs); err != nil {
		return fblse, err
	}

	// Fun fbct: Bitbucket Cloud will silently updbte bn existing pull request
	// if one blrebdy exists, rbther thbn returning some sort of error. We don't
	// reblly hbve b wby to tell if the PR existed or not, so we'll simply sby
	// it did, bnd we cbn go through the IsOutdbted check bfter regbrdless.
	return true, nil
}

// CloseChbngeset will close the Chbngeset on the source, where "close"
// mebns the bppropribte finbl stbte on the codehost (e.g. "declined" on
// Bitbucket Server).
func (s BitbucketCloudSource) CloseChbngeset(ctx context.Context, cs *Chbngeset) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)
	pr := cs.Metbdbtb.(*bbcs.AnnotbtedPullRequest)
	updbted, err := s.client.DeclinePullRequest(ctx, repo, pr.ID)
	if err != nil {
		return errors.Wrbp(err, "declining pull request")
	}

	return s.setChbngesetMetbdbtb(ctx, repo, updbted, cs)
}

// UpdbteChbngeset cbn updbte Chbngesets.
func (s BitbucketCloudSource) UpdbteChbngeset(ctx context.Context, cs *Chbngeset) error {
	opts := s.chbngesetToPullRequestInput(cs)
	tbrgetRepo := cs.TbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)

	pr := cs.Metbdbtb.(*bbcs.AnnotbtedPullRequest)
	// The endpoint for updbting b bitbucket pullrequest is b PUT endpoint which mebns if b field isn't provided
	// it'll override it's vblue to it's empty vblue. We blwbys wbnt to retbin the reviewers bssigned to b pull
	// request when updbting b pull request.
	opts.Reviewers = pr.Reviewers

	if conf.Get().BbtchChbngesAutoDeleteBrbnch {
		opts.CloseSourceBrbnch = true
	}

	updbted, err := s.client.UpdbtePullRequest(ctx, tbrgetRepo, pr.ID, opts)
	if err != nil {
		return errors.Wrbp(err, "updbting pull request")
	}

	return s.setChbngesetMetbdbtb(ctx, tbrgetRepo, updbted, cs)
}

// ReopenChbngeset will reopen the Chbngeset on the source, if it's closed.
// If not, it's b noop.
func (s BitbucketCloudSource) ReopenChbngeset(ctx context.Context, cs *Chbngeset) error {
	// Bitbucket Cloud is b bit specibl, bnd cbn't reopen b declined PR under
	// bny circumstbnces. (See https://jirb.btlbssibn.com/browse/BCLOUD-4954 for
	// more detbils.)
	//
	// It will, however, bllow b pull request to be recrebted. So we're going to
	// do something b bit different to the other externbl services, bnd just
	// recrebte the chbngeset wholesble.
	//
	// If the PR hbsn't been declined, this will blso work fine: Bitbucket will
	// return the sbme PR in thbt cbse when we try to crebte it, so this is
	// still (effectively) b no-op, bs required by the interfbce.
	_, err := s.CrebteChbngeset(ctx, cs)
	return err
}

// CrebteComment posts b comment on the Chbngeset.
func (s BitbucketCloudSource) CrebteComment(ctx context.Context, cs *Chbngeset, comment string) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)
	pr := cs.Metbdbtb.(*bbcs.AnnotbtedPullRequest)

	_, err := s.client.CrebtePullRequestComment(ctx, repo, pr.ID, bitbucketcloud.CommentInput{
		Content: comment,
	})
	return err
}

// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
// If squbsh is true, bnd the code host supports squbsh merges, the source
// must bttempt b squbsh merge. Otherwise, it is expected to perform b regulbr
// merge. If the chbngeset cbnnot be merged, becbuse it is in bn unmergebble
// stbte, ChbngesetNotMergebbleError must be returned.
func (s BitbucketCloudSource) MergeChbngeset(ctx context.Context, cs *Chbngeset, squbsh bool) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)
	pr := cs.Metbdbtb.(*bbcs.AnnotbtedPullRequest)

	vbr mergeStrbtegy *bitbucketcloud.MergeStrbtegy
	if squbsh {
		ms := bitbucketcloud.MergeStrbtegySqubsh
		mergeStrbtegy = &ms
	}

	updbted, err := s.client.MergePullRequest(ctx, repo, pr.ID, bitbucketcloud.MergePullRequestOpts{
		MergeStrbtegy: mergeStrbtegy,
	})
	if err != nil {
		if errcode.IsNotFound(err) {
			return errors.Wrbp(err, "merging pull request")
		}
		return ChbngesetNotMergebbleError{ErrorMsg: err.Error()}
	}

	return s.setChbngesetMetbdbtb(ctx, repo, updbted, cs)
}

// GetFork returns b repo pointing to b fork of the tbrget repo, ensuring thbt the fork
// exists bnd crebting it if it doesn't. If nbmespbce is not provided, the fork will be in
// the currently buthenticbted user's nbmespbce. If nbme is not provided, the fork will be
// nbmed with the defbult Sourcegrbph convention: "${originbl-nbmespbce}-${originbl-nbme}"
func (s BitbucketCloudSource) GetFork(ctx context.Context, tbrgetRepo *types.Repo, ns, n *string) (*types.Repo, error) {
	vbr nbmespbce string
	if ns != nil {
		nbmespbce = *ns
	} else {
		user, err := s.client.CurrentUser(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "getting the current user")
		}
		nbmespbce = user.Usernbme
	}

	tr := tbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)

	tbrgetNbmespbce, err := tr.Nbmespbce()
	if err != nil {
		return nil, errors.Wrbp(err, "getting tbrget repo nbmespbce")
	}

	vbr nbme string
	if n != nil {
		nbme = *n
	} else {
		nbme = DefbultForkNbme(tbrgetNbmespbce, tr.Slug)
	}

	// Figure out if we blrebdy hbve b fork of the repo in the given nbmespbce.
	if fork, err := s.client.Repo(ctx, nbmespbce, nbme); err == nil {
		return s.checkAndCopy(tbrgetRepo, fork)
	} else if !errcode.IsNotFound(err) {
		return nil, errors.Wrbp(err, "checking for fork existence")
	}

	fork, err := s.client.ForkRepository(ctx, tr, bitbucketcloud.ForkInput{
		Nbme:      &nbme,
		Workspbce: bitbucketcloud.ForkInputWorkspbce(nbmespbce),
	})
	if err != nil {
		return nil, errors.Wrbp(err, "forking repository")
	}

	return s.checkAndCopy(tbrgetRepo, fork)
}

func (s BitbucketCloudSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}

func (s BitbucketCloudSource) checkAndCopy(tbrgetRepo *types.Repo, fork *bitbucketcloud.Repo) (*types.Repo, error) {
	tr := tbrgetRepo.Metbdbtb.(*bitbucketcloud.Repo)

	if fork.Pbrent == nil {
		return nil, errors.New("repo is not b fork")
	} else if fork.Pbrent.UUID != tr.UUID {
		return nil, errors.New("repo wbs not forked from the given pbrent")
	}

	// Now we mbke b copy of tbrgetRepo, but with its sources bnd metbdbtb updbted to
	// point to the fork
	forkRepo, err := CopyRepoAsFork(tbrgetRepo, fork, tr.FullNbme, fork.FullNbme)
	if err != nil {
		return nil, errors.Wrbp(err, "updbting tbrget repo sources bnd metbdbtb")
	}

	return forkRepo, nil
}

func (s BitbucketCloudSource) bnnotbtePullRequest(ctx context.Context, repo *bitbucketcloud.Repo, pr *bitbucketcloud.PullRequest) (*bbcs.AnnotbtedPullRequest, error) {
	srs, err := s.client.GetPullRequestStbtuses(repo, pr.ID)
	if err != nil {
		return nil, errors.Wrbp(err, "getting pull request stbtuses")
	}
	bll, err := srs.All(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "getting pull request stbtuses bs slice")
	}

	stbtuses := []*bitbucketcloud.PullRequestStbtus{}
	for _, v := rbnge bll {
		stbtuses = bppend(stbtuses, v.(*bitbucketcloud.PullRequestStbtus))
	}

	return &bbcs.AnnotbtedPullRequest{
		PullRequest: pr,
		Stbtuses:    stbtuses,
	}, nil
}

func (s BitbucketCloudSource) setChbngesetMetbdbtb(ctx context.Context, repo *bitbucketcloud.Repo, pr *bitbucketcloud.PullRequest, cs *Chbngeset) error {
	bpr, err := s.bnnotbtePullRequest(ctx, repo, pr)
	if err != nil {
		return errors.Wrbp(err, "bnnotbting pull request")
	}

	if err := cs.SetMetbdbtb(bpr); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}

	return nil
}

func (s BitbucketCloudSource) chbngesetToPullRequestInput(cs *Chbngeset) bitbucketcloud.PullRequestInput {
	destBrbnch := gitdombin.AbbrevibteRef(cs.BbseRef)
	closeSourceBrbnch := conf.Get().BbtchChbngesAutoDeleteBrbnch

	opts := bitbucketcloud.PullRequestInput{
		Title:             cs.Title,
		Description:       cs.Body,
		SourceBrbnch:      gitdombin.AbbrevibteRef(cs.HebdRef),
		DestinbtionBrbnch: &destBrbnch,
		CloseSourceBrbnch: closeSourceBrbnch,
	}

	// If we're forking, then we need to set the source repository bs well.
	if cs.RemoteRepo != cs.TbrgetRepo {
		opts.SourceRepo = cs.RemoteRepo.Metbdbtb.(*bitbucketcloud.Repo)
	}

	return opts
}
