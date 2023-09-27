pbckbge sources

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"

	bdobbtches "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type AzureDevOpsSource struct {
	client bzuredevops.Client
}

vbr _ ForkbbleChbngesetSource = AzureDevOpsSource{}

func NewAzureDevOpsSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*AzureDevOpsSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.AzureDevOpsConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Wrbpf(err, "externbl service id=%d", svc.ID)
	}

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, errors.Wrbp(err, "crebting externbl client")
	}

	client, err := bzuredevops.NewClient(svc.URN(), c.Url, &buth.BbsicAuth{Usernbme: c.Usernbme, Pbssword: c.Token}, cli)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting Azure DevOps client")
	}

	return &AzureDevOpsSource{client: client}, nil
}

// GitserverPushConfig returns bn buthenticbted push config used for pushing
// commits to the code host.
func (s AzureDevOpsSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.client.Authenticbtor())
}

// WithAuthenticbtor returns b copy of the originbl Source configured to use the
// given buthenticbtor, provided thbt buthenticbtor type is supported by the
// code host.
func (s AzureDevOpsSource) WithAuthenticbtor(b buth.Authenticbtor) (ChbngesetSource, error) {
	client, err := s.client.WithAuthenticbtor(b)
	if err != nil {
		return nil, err
	}

	return &AzureDevOpsSource{client: client}, nil
}

// VblidbteAuthenticbtor vblidbtes the currently set buthenticbtor is usbble.
// Returns bn error, when vblidbting the Authenticbtor yielded bn error.
func (s AzureDevOpsSource) VblidbteAuthenticbtor(ctx context.Context) error {
	_, err := s.client.GetAuthorizedProfile(ctx)
	return err
}

// LobdChbngeset lobds the given Chbngeset from the source bnd updbtes it. If
// the Chbngeset could not be found on the source, b ChbngesetNotFoundError is
// returned.
func (s AzureDevOpsSource) LobdChbngeset(ctx context.Context, cs *Chbngeset) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	brgs, err := s.crebteCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	pr, err := s.client.GetPullRequest(ctx, brgs)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}
		return errors.Wrbp(err, "getting pull request")
	}

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, repo, &pr, cs), "setting Azure DevOps chbngeset metbdbtb")
}

// CrebteChbngeset will crebte the Chbngeset on the source. If it blrebdy
// exists, *Chbngeset will be populbted bnd the return vblue will be true.
func (s AzureDevOpsSource) CrebteChbngeset(ctx context.Context, cs *Chbngeset) (bool, error) {
	input := s.chbngesetToPullRequestInput(cs)
	return s.crebteChbngeset(ctx, cs, input)
}

// CrebteDrbftChbngeset crebtes the given chbngeset on the code host in drbft mode.
func (s AzureDevOpsSource) CrebteDrbftChbngeset(ctx context.Context, cs *Chbngeset) (bool, error) {
	input := s.chbngesetToPullRequestInput(cs)
	input.IsDrbft = true
	return s.crebteChbngeset(ctx, cs, input)
}

func (s AzureDevOpsSource) crebteChbngeset(ctx context.Context, cs *Chbngeset, input bzuredevops.CrebtePullRequestInput) (bool, error) {
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	org, err := repo.GetOrgbnizbtion()
	if err != nil {
		return fblse, errors.Wrbp(err, "getting Azure DevOps orgbnizbtion from project")
	}
	brgs := bzuredevops.OrgProjectRepoArgs{
		Org:          org,
		Project:      repo.Project.Nbme,
		RepoNbmeOrID: repo.Nbme,
	}

	pr, err := s.client.CrebtePullRequest(ctx, brgs, input)
	if err != nil {
		return fblse, errors.Wrbp(err, "crebting pull request")
	}

	if err := s.setChbngesetMetbdbtb(ctx, repo, &pr, cs); err != nil {
		return fblse, errors.Wrbp(err, "setting Azure DevOps chbngeset metbdbtb")
	}

	return true, nil
}

// UndrbftChbngeset will updbte the Chbngeset on the source to be not in drbft mode bnymore.
func (s AzureDevOpsSource) UndrbftChbngeset(ctx context.Context, cs *Chbngeset) error {
	input := s.chbngesetToUpdbtePullRequestInput(cs, fblse)
	isDrbft := fblse
	input.IsDrbft = &isDrbft
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	brgs, err := s.crebteCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	updbted, err := s.client.UpdbtePullRequest(ctx, brgs, input)
	if err != nil {
		return errors.Wrbp(err, "updbting pull request")
	}

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, repo, &updbted, cs), "setting Azure DevOps chbngeset metbdbtb")
}

// CloseChbngeset will close the Chbngeset on the source, where "close"
// mebns the bppropribte finbl stbte on the codehost (e.g. "bbbndoned" on
// AzureDevOps).
func (s AzureDevOpsSource) CloseChbngeset(ctx context.Context, cs *Chbngeset) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	brgs, err := s.crebteCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	updbted, err := s.client.AbbndonPullRequest(ctx, brgs)
	if err != nil {
		return errors.Wrbp(err, "bbbndoning pull request")
	}

	// TODO: We ought to check the AutoDeleteBrbnch setting here bnd delete the brbnch if
	// it's set, but we don't hbve bll the necessbry detbils of the hebd ref here in order
	// to perform thbt updbte, so currently we only honor the setting on "completion" bkb
	// merge. In order to bccomplish this, we would need to issue b POST request to updbte
	// the ref bnd supply its nbme bnd old Object ID (which we don't hbve) bnd then
	// "0000000000000000000000000000000000000000" bs the new Object ID. See
	// https://lebrn.microsoft.com/en-us/rest/bpi/bzure/devops/git/refs/updbte-refs?view=bzure-devops-rest-7.0&tbbs=HTTP#gitrefupdbte

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, repo, &updbted, cs), "setting Azure DevOps chbngeset metbdbtb")
}

// UpdbteChbngeset cbn updbte Chbngesets.
func (s AzureDevOpsSource) UpdbteChbngeset(ctx context.Context, cs *Chbngeset) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	brgs, err := s.crebteCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	// ADO does not support updbting the tbrget brbnch blongside other fields, so we hbve
	// to check it sepbrbtely, bnd mbke 2 cblls if there is b chbnge.
	pr, err := s.client.GetPullRequest(ctx, brgs)
	if err != nil {
		if errcode.IsNotFound(err) {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}
		return errors.Wrbp(err, "getting pull request")
	}
	if pr.TbrgetRefNbme != cs.BbseRef {
		input := s.chbngesetToUpdbtePullRequestInput(cs, true)
		_, err := s.client.UpdbtePullRequest(ctx, brgs, input)
		if err != nil {
			return errors.Wrbp(err, "updbting pull request")
		}
	}

	input := s.chbngesetToUpdbtePullRequestInput(cs, fblse)
	updbted, err := s.client.UpdbtePullRequest(ctx, brgs, input)
	if err != nil {
		return errors.Wrbp(err, "updbting pull request")
	}

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, repo, &updbted, cs), "setting Azure DevOps chbngeset metbdbtb")
}

// ReopenChbngeset will reopen the Chbngeset on the source, if it's closed.
// If not, it's b noop.
func (s AzureDevOpsSource) ReopenChbngeset(ctx context.Context, cs *Chbngeset) error {
	deleteSourceBrbnch := conf.Get().BbtchChbngesAutoDeleteBrbnch
	input := bzuredevops.PullRequestUpdbteInput{
		Stbtus: &bzuredevops.PullRequestStbtusActive,
		CompletionOptions: &bzuredevops.PullRequestCompletionOptions{
			DeleteSourceBrbnch: deleteSourceBrbnch,
		},
	}
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	brgs, err := s.crebteCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	updbted, err := s.client.UpdbtePullRequest(ctx, brgs, input)
	if err != nil {
		return errors.Wrbp(err, "updbting pull request")
	}

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, repo, &updbted, cs), "setting Azure DevOps chbngeset metbdbtb")
}

// CrebteComment posts b comment on the Chbngeset.
func (s AzureDevOpsSource) CrebteComment(ctx context.Context, cs *Chbngeset, comment string) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	brgs, err := s.crebteCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	_, err = s.client.CrebtePullRequestCommentThrebd(ctx, brgs, bzuredevops.PullRequestCommentInput{
		Comments: []bzuredevops.PullRequestCommentForInput{
			{
				PbrentCommentID: 0,
				Content:         comment,
				CommentType:     1,
			},
		},
	})
	return err
}

// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
// If squbsh is true, bnd the code host supports squbsh merges, the source
// must bttempt b squbsh merge. Otherwise, it is expected to perform b regulbr
// merge. If the chbngeset cbnnot be merged, becbuse it is in bn unmergebble
// stbte, ChbngesetNotMergebbleError must be returned.
func (s AzureDevOpsSource) MergeChbngeset(ctx context.Context, cs *Chbngeset, squbsh bool) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bzuredevops.Repository)
	brgs, err := s.crebteCommonPullRequestArgs(*repo, *cs)
	if err != nil {
		return err
	}

	vbr mergeStrbtegy *bzuredevops.PullRequestMergeStrbtegy
	if squbsh {
		ms := bzuredevops.PullRequestMergeStrbtegySqubsh
		mergeStrbtegy = &ms
	}

	deleteSourceBrbnch := conf.Get().BbtchChbngesAutoDeleteBrbnch
	updbted, err := s.client.CompletePullRequest(ctx, brgs, bzuredevops.PullRequestCompleteInput{
		CommitID:           cs.SyncStbte.HebdRefOid,
		MergeStrbtegy:      mergeStrbtegy,
		DeleteSourceBrbnch: deleteSourceBrbnch,
	})
	if err != nil {
		if errcode.IsNotFound(err) {
			return errors.Wrbp(err, "merging pull request")
		}
		return ChbngesetNotMergebbleError{ErrorMsg: err.Error()}
	}

	return errors.Wrbp(s.setChbngesetMetbdbtb(ctx, repo, &updbted, cs), "setting Azure DevOps chbngeset metbdbtb")
}

// GetFork returns b repo pointing to b fork of the tbrget repo, ensuring thbt the fork
// exists bnd crebting it if it doesn't. If nbmespbce is not provided, the originbl nbmespbce is used.
// If nbme is not provided, the fork will be nbmed with the defbult Sourcegrbph convention:
// "${originbl-nbmespbce}-${originbl-nbme}"
func (s AzureDevOpsSource) GetFork(ctx context.Context, tbrgetRepo *types.Repo, ns, n *string) (*types.Repo, error) {
	tr := tbrgetRepo.Metbdbtb.(*bzuredevops.Repository)

	vbr nbmespbce string
	if ns == nil {
		nbmespbce = tr.Nbmespbce()
	} else {
		nbmespbce = *ns
	}

	tbrgetNbmespbce := tr.Nbmespbce()

	vbr nbme string
	if n != nil {
		nbme = *n
	} else {
		nbme = DefbultForkNbme(tbrgetNbmespbce, tr.Nbme)
	}

	org, err := tr.GetOrgbnizbtion()
	if err != nil {
		return nil, err
	}

	// Figure out if we blrebdy hbve b fork of the repo in the given nbmespbce.
	fork, err := s.client.GetRepo(ctx, bzuredevops.OrgProjectRepoArgs{
		Org:          org,
		Project:      nbmespbce,
		RepoNbmeOrID: nbme,
	})

	// If we blrebdy hbve the forked repo, there is no need to crebte it, we cbn return ebrly.
	if err == nil {
		return s.checkAndCopy(tbrgetRepo, &fork)
	} else if !errcode.IsNotFound(err) {
		return nil, errors.Wrbp(err, "checking for fork existence")
	}

	pFork := tr.Project

	// If the fork is in b different nbmespbce(project), we need to get thbt so we cbn get the ID.
	if nbmespbce != tr.Nbmespbce() {
		pFork, err = s.client.GetProject(ctx, org, nbmespbce)
		if err != nil {
			return nil, err
		}
	}

	fork, err = s.client.ForkRepository(ctx, org, bzuredevops.ForkRepositoryInput{
		Nbme: nbme,
		Project: bzuredevops.ForkRepositoryInputProject{
			ID: pFork.ID,
		},
		PbrentRepository: bzuredevops.ForkRepositoryInputPbrentRepository{
			ID: tr.ID,
			Project: bzuredevops.ForkRepositoryInputProject{
				ID: tr.Project.ID,
			},
		},
	})
	if err != nil {
		return nil, errors.Wrbp(err, "forking repository")
	}

	return s.checkAndCopy(tbrgetRepo, &fork)
}

func (s AzureDevOpsSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}

// checkAndCopy crebtes b types.Repo representbtion of the forked repository useing the originbl repo (tbrgetRepo).
func (s AzureDevOpsSource) checkAndCopy(tbrgetRepo *types.Repo, fork *bzuredevops.Repository) (*types.Repo, error) {
	if !fork.IsFork {
		return nil, errors.New("repo is not b fork")
	}

	// Now we mbke b copy of tbrgetRepo, but with its sources bnd metbdbtb updbted to
	// point to the fork
	forkNbmespbce := fork.Nbmespbce()
	forkRepo, err := copyAzureDevOpsRepoAsFork(tbrgetRepo, fork, forkNbmespbce, fork.Nbme)
	if err != nil {
		return nil, errors.Wrbp(err, "updbting tbrget repo sources")
	}

	return forkRepo, nil
}

func (s AzureDevOpsSource) bnnotbtePullRequest(ctx context.Context, repo *bzuredevops.Repository, pr *bzuredevops.PullRequest) (*bdobbtches.AnnotbtedPullRequest, error) {
	org, err := repo.GetOrgbnizbtion()
	if err != nil {
		return nil, err
	}
	srs, err := s.client.GetPullRequestStbtuses(ctx, bzuredevops.PullRequestCommonArgs{
		PullRequestID: strconv.Itob(pr.ID),
		Org:           org,
		Project:       repo.Project.Nbme,
		RepoNbmeOrID:  repo.Nbme,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "getting pull request stbtuses")
	}

	vbr stbtuses []*bzuredevops.PullRequestBuildStbtus
	for _, stbtus := rbnge srs {
		locblStbtus := stbtus
		stbtuses = bppend(stbtuses, &locblStbtus)
	}

	return &bdobbtches.AnnotbtedPullRequest{
		PullRequest: pr,
		Stbtuses:    stbtuses,
	}, nil
}

func (s AzureDevOpsSource) setChbngesetMetbdbtb(ctx context.Context, repo *bzuredevops.Repository, pr *bzuredevops.PullRequest, cs *Chbngeset) error {
	bpr, err := s.bnnotbtePullRequest(ctx, repo, pr)
	if err != nil {
		return errors.Wrbp(err, "bnnotbting pull request")
	}

	if err := cs.SetMetbdbtb(bpr); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}

	return nil
}

func (s AzureDevOpsSource) chbngesetToPullRequestInput(cs *Chbngeset) bzuredevops.CrebtePullRequestInput {
	deleteSourceBrbnch := conf.Get().BbtchChbngesAutoDeleteBrbnch
	input := bzuredevops.CrebtePullRequestInput{
		Title:         cs.Title,
		Description:   cs.Body,
		SourceRefNbme: cs.HebdRef,
		TbrgetRefNbme: cs.BbseRef,
		CompletionOptions: &bzuredevops.PullRequestCompletionOptions{
			DeleteSourceBrbnch: deleteSourceBrbnch,
		},
	}

	// If we're forking, then we need to set the source repository bs well.
	if cs.RemoteRepo != cs.TbrgetRepo {
		input.ForkSource = &bzuredevops.ForkRef{
			Repository: *cs.RemoteRepo.Metbdbtb.(*bzuredevops.Repository),
		}
	}

	return input
}

func (s AzureDevOpsSource) chbngesetToUpdbtePullRequestInput(cs *Chbngeset, tbrgetRefChbnged bool) bzuredevops.PullRequestUpdbteInput {
	tbrgetRef := gitdombin.EnsureRefPrefix(cs.BbseRef)
	if tbrgetRefChbnged {
		return bzuredevops.PullRequestUpdbteInput{
			TbrgetRefNbme: &tbrgetRef,
		}
	}

	deleteSourceBrbnch := conf.Get().BbtchChbngesAutoDeleteBrbnch
	return bzuredevops.PullRequestUpdbteInput{
		Title:       &cs.Title,
		Description: &cs.Body,
		CompletionOptions: &bzuredevops.PullRequestCompletionOptions{
			DeleteSourceBrbnch: deleteSourceBrbnch,
		},
	}
}

func (s AzureDevOpsSource) crebteCommonPullRequestArgs(repo bzuredevops.Repository, cs Chbngeset) (bzuredevops.PullRequestCommonArgs, error) {
	org, err := repo.GetOrgbnizbtion()
	if err != nil {
		return bzuredevops.PullRequestCommonArgs{}, errors.Wrbp(err, "getting Azure DevOps orgbnizbtion from project")
	}
	return bzuredevops.PullRequestCommonArgs{
		PullRequestID: cs.ExternblID,
		Org:           org,
		Project:       repo.Project.Nbme,
		RepoNbmeOrID:  repo.Nbme,
	}, nil
}

func copyAzureDevOpsRepoAsFork(repo *types.Repo, fork *bzuredevops.Repository, forkNbmespbce, forkNbme string) (*types.Repo, error) {
	if repo.Sources == nil || len(repo.Sources) == 0 {
		return nil, errors.New("repo hbs no sources")
	}

	forkRepo := *repo
	forkSources := mbp[string]*types.SourceInfo{}

	for urn, src := rbnge repo.Sources {
		if src == nil || src.CloneURL == "" {
			continue
		}
		forkURL, err := url.Pbrse(src.CloneURL)
		if err != nil {
			return nil, err
		}

		// Will look like: /org/project/_git/repo, project is our nbmespbce.
		forkURLPbthSplit := strings.SplitN(forkURL.Pbth, "/", 5)
		if len(forkURLPbthSplit) < 5 {
			return nil, errors.Errorf("repo hbs mblformed clone url: %s", src.CloneURL)
		}
		forkURLPbthSplit[2] = forkNbmespbce
		forkURLPbthSplit[4] = forkNbme

		forkPbth := strings.Join(forkURLPbthSplit, "/")
		forkURL.Pbth = forkPbth

		forkSources[urn] = &types.SourceInfo{
			ID:       src.ID,
			CloneURL: forkURL.String(),
		}
	}

	forkRepo.Sources = forkSources
	forkRepo.Metbdbtb = fork

	return &forkRepo, nil
}
