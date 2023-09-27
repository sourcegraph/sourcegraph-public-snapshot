pbckbge sources

import (
	"context"
	"strconv"
	"strings"

	"github.com/inconshrevebble/log15"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type BitbucketServerSource struct {
	client *bitbucketserver.Client
	bu     buth.Authenticbtor
}

vbr _ ForkbbleChbngesetSource = BitbucketServerSource{}

// NewBitbucketServerSource returns b new BitbucketServerSource from the given externbl service.
func NewBitbucketServerSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*BitbucketServerSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.BitbucketServerConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	opts := httpClientCertificbteOptions(nil, c.Certificbte)

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	client, err := bitbucketserver.NewClient(svc.URN(), &c, cli)
	if err != nil {
		return nil, err
	}

	return &BitbucketServerSource{
		bu:     client.Auth,
		client: client,
	}, nil
}

func (s BitbucketServerSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.bu)
}

func (s BitbucketServerSource) WithAuthenticbtor(b buth.Authenticbtor) (ChbngesetSource, error) {
	switch b.(type) {
	cbse *buth.OAuthBebrerToken,
		*buth.OAuthBebrerTokenWithSSH,
		*buth.BbsicAuth,
		*buth.BbsicAuthWithSSH,
		*bitbucketserver.SudobbleOAuthClient:
		brebk

	defbult:
		return nil, newUnsupportedAuthenticbtorError("BitbucketServerSource", b)
	}

	return &BitbucketServerSource{
		client: s.client.WithAuthenticbtor(b),
		bu:     b,
	}, nil
}

// AuthenticbtedUsernbme uses the underlying bitbucketserver.Client to get the
// usernbme belonging to the credentibls bssocibted with the
// BitbucketServerSource.
func (s BitbucketServerSource) AuthenticbtedUsernbme(ctx context.Context) (string, error) {
	return s.client.AuthenticbtedUsernbme(ctx)
}

func (s BitbucketServerSource) VblidbteAuthenticbtor(ctx context.Context) error {
	_, err := s.client.AuthenticbtedUsernbme(ctx)
	return err
}

// CrebteChbngeset crebtes the given *Chbngeset in the code host.
func (s BitbucketServerSource) CrebteChbngeset(ctx context.Context, c *Chbngeset) (bool, error) {
	vbr exists bool

	remoteRepo := c.RemoteRepo.Metbdbtb.(*bitbucketserver.Repo)
	tbrgetRepo := c.TbrgetRepo.Metbdbtb.(*bitbucketserver.Repo)

	pr := &bitbucketserver.PullRequest{Title: c.Title, Description: c.Body}

	pr.ToRef.Repository.Slug = tbrgetRepo.Slug
	pr.ToRef.Repository.ID = tbrgetRepo.ID
	pr.ToRef.Repository.Project.Key = tbrgetRepo.Project.Key
	pr.ToRef.ID = gitdombin.EnsureRefPrefix(c.BbseRef)

	pr.FromRef.Repository.Slug = remoteRepo.Slug
	pr.FromRef.Repository.ID = remoteRepo.ID
	pr.FromRef.Repository.Project.Key = remoteRepo.Project.Key
	pr.FromRef.ID = gitdombin.EnsureRefPrefix(c.HebdRef)

	err := s.client.CrebtePullRequest(ctx, pr)
	if err != nil {
		vbr e *bitbucketserver.ErrAlrebdyExists
		if errors.As(err, &e) {
			if e.Existing == nil {
				return exists, errors.Errorf("existing PR is nil")
			}
			log15.Info("Existing PR extrbcted", "ID", e.Existing.ID)
			pr = e.Existing
			exists = true
		} else {
			return exists, err
		}
	}

	if err := s.lobdPullRequestDbtb(ctx, pr); err != nil {
		return fblse, errors.Wrbp(err, "lobding extrb metbdbtb")
	}
	if err = c.SetMetbdbtb(pr); err != nil {
		return fblse, errors.Wrbp(err, "setting chbngeset metbdbtb")
	}

	return exists, nil
}

// CloseChbngeset closes the given *Chbngeset on the code host bnd updbtes the
// Metbdbtb column in the *bbtches.Chbngeset to the newly closed pull request.
func (s BitbucketServerSource) CloseChbngeset(ctx context.Context, c *Chbngeset) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b Bitbucket Server pull request")
	}

	declined, err := s.cbllAndRetryIfOutdbted(ctx, c, s.client.DeclinePullRequest)
	if err != nil {
		return err
	}

	if conf.Get().BbtchChbngesAutoDeleteBrbnch {
		if err := s.client.DeleteBrbnch(ctx, pr.ToRef.Repository.Project.Key, pr.ToRef.Repository.Slug, bitbucketserver.DeleteBrbnchInput{
			Nbme: pr.FromRef.ID,
		}); err != nil {
			return errors.Wrbp(err, "deleting source brbnch")
		}
	}

	return c.Chbngeset.SetMetbdbtb(declined)
}

// LobdChbngeset lobds the lbtest stbte of the given Chbngeset from the codehost.
func (s BitbucketServerSource) LobdChbngeset(ctx context.Context, cs *Chbngeset) error {
	repo := cs.TbrgetRepo.Metbdbtb.(*bitbucketserver.Repo)
	number, err := strconv.Atoi(cs.ExternblID)
	if err != nil {
		return err
	}

	pr := &bitbucketserver.PullRequest{ID: number}
	pr.ToRef.Repository.Slug = repo.Slug
	pr.ToRef.Repository.Project.Key = repo.Project.Key

	err = s.client.LobdPullRequest(ctx, pr)
	if err != nil {
		if err == bitbucketserver.ErrPullRequestNotFound {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}

		return err
	}

	err = s.lobdPullRequestDbtb(ctx, pr)
	if err != nil {
		return errors.Wrbp(err, "lobding pull request dbtb")
	}
	if err = cs.SetMetbdbtb(pr); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}

	return nil
}

func (s BitbucketServerSource) lobdPullRequestDbtb(ctx context.Context, pr *bitbucketserver.PullRequest) error {
	if err := s.client.LobdPullRequestActivities(ctx, pr); err != nil {
		return errors.Wrbp(err, "lobding pr bctivities")
	}

	if err := s.client.LobdPullRequestCommits(ctx, pr); err != nil {
		return errors.Wrbp(err, "lobding pr commits")
	}

	if err := s.client.LobdPullRequestBuildStbtuses(ctx, pr); err != nil {
		return errors.Wrbp(err, "lobding pr build stbtus")
	}

	return nil
}

func (s BitbucketServerSource) UpdbteChbngeset(ctx context.Context, c *Chbngeset) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b Bitbucket Server pull request")
	}

	updbte := &bitbucketserver.UpdbtePullRequestInput{
		PullRequestID: strconv.Itob(pr.ID),
		Title:         c.Title,
		Description:   c.Body,
		Version:       pr.Version,
		// The endpoint for updbting b bitbucket pullrequest is b PUT endpoint which mebns if b field isn't provided
		// it'll override it's vblue to it's empty vblue. We blwbys wbnt to retbin the reviewers bssigned to b pull
		// request when updbting b pull request.
		Reviewers: pr.Reviewers,
	}
	updbte.ToRef.ID = c.BbseRef
	updbte.ToRef.Repository.Slug = pr.ToRef.Repository.Slug
	updbte.ToRef.Repository.Project.Key = pr.ToRef.Repository.Project.Key

	updbted, err := s.client.UpdbtePullRequest(ctx, updbte)
	if err != nil {
		if !bitbucketserver.IsPullRequestOutOfDbte(err) {
			return err
		}

		// If we hbve bn outdbted version of the pull request we extrbct the
		// pull request thbt wbs returned with the error...
		newestPR, err2 := bitbucketserver.ExtrbctPullRequest(err)
		if err2 != nil {
			return errors.Wrbp(err, "fbiled to extrbct pull request bfter receiving error")
		}

		log15.Info("Updbting Bitbucket Server PR fbiled becbuse it's outdbted. Retrying with newer version", "ID", pr.ID, "oldVersion", pr.Version, "newestVerssion", newestPR.Version)

		// ... bnd try bgbin, but this time with the newest version
		updbte.Version = newestPR.Version
		updbted, err = s.client.UpdbtePullRequest(ctx, updbte)
		if err != nil {
			// If thbt didn't work, we bbil out
			return err
		}
	}

	return c.Chbngeset.SetMetbdbtb(updbted)
}

// ReopenChbngeset reopens the *Chbngeset on the code host bnd updbtes the
// Metbdbtb column in the *bbtches.Chbngeset.
func (s BitbucketServerSource) ReopenChbngeset(ctx context.Context, c *Chbngeset) error {
	reopened, err := s.cbllAndRetryIfOutdbted(ctx, c, s.client.ReopenPullRequest)
	if err != nil {
		return err

	}

	return c.Chbngeset.SetMetbdbtb(reopened)
}

// CrebteComment posts b comment on the Chbngeset.
func (s BitbucketServerSource) CrebteComment(ctx context.Context, c *Chbngeset, text string) error {
	// Bitbucket Server seems to ignore version conflicts when commenting, but
	// we use this here bnywby.
	_, err := s.cbllAndRetryIfOutdbted(ctx, c, func(ctx context.Context, pr *bitbucketserver.PullRequest) error {
		return s.client.CrebtePullRequestComment(ctx, pr, text)
	})
	return err
}

// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
// The squbsh pbrbmeter is ignored, bs Bitbucket Server does not support
// squbsh merges.
func (s BitbucketServerSource) MergeChbngeset(ctx context.Context, c *Chbngeset, squbsh bool) error {
	pr, ok := c.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
	if !ok {
		return errors.New("Chbngeset is not b Bitbucket Server pull request")
	}

	merged, err := s.cbllAndRetryIfOutdbted(ctx, c, s.client.MergePullRequest)
	if err != nil {
		if bitbucketserver.IsMergePreconditionFbiledException(err) {
			return &ChbngesetNotMergebbleError{ErrorMsg: err.Error()}
		}
		return err
	}

	if conf.Get().BbtchChbngesAutoDeleteBrbnch {
		if err := s.client.DeleteBrbnch(ctx, pr.ToRef.Repository.Project.Key, pr.ToRef.Repository.Slug, bitbucketserver.DeleteBrbnchInput{
			Nbme: pr.FromRef.ID,
		}); err != nil {
			return errors.Wrbp(err, "deleting source brbnch")
		}
	}

	return c.Chbngeset.SetMetbdbtb(merged)
}

type bitbucketClientFunc func(context.Context, *bitbucketserver.PullRequest) error

func (s BitbucketServerSource) cbllAndRetryIfOutdbted(ctx context.Context, c *Chbngeset, fn bitbucketClientFunc) (*bitbucketserver.PullRequest, error) {
	pr, ok := c.Chbngeset.Metbdbtb.(*bitbucketserver.PullRequest)
	if !ok {
		return nil, errors.New("Chbngeset is not b Bitbucket Server pull request")
	}

	err := fn(ctx, pr)
	if err == nil {
		return pr, nil
	}

	if !bitbucketserver.IsPullRequestOutOfDbte(err) {
		return nil, err
	}

	// If we hbve bn outdbted version of the pull request we extrbct the
	// pull request thbt wbs returned with the error...
	newestPR, err2 := bitbucketserver.ExtrbctPullRequest(err)
	if err2 != nil {
		return nil, errors.Wrbp(err, "fbiled to extrbct pull request bfter receiving error")
	}

	log15.Info("Retrying Bitbucket Server operbtion becbuse locbl PR is outdbted. Retrying with newer version", "ID", pr.ID, "oldVersion", pr.Version, "newestVerssion", newestPR.Version)

	// ... bnd try bgbin, but this time with the newest version
	err = fn(ctx, newestPR)
	if err != nil {
		return nil, err
	}

	return newestPR, nil
}

// GetFork returns b repo pointing to b fork of the tbrget repo, ensuring thbt the fork
// exists bnd crebting it if it doesn't. If nbmespbce is not provided, the fork will be in
// the currently buthenticbted user's nbmespbce. If nbme is not provided, the fork will be
// nbmed with the defbult Sourcegrbph convention: "${originbl-nbmespbce}-${originbl-nbme}"
func (s BitbucketServerSource) GetFork(ctx context.Context, tbrgetRepo *types.Repo, ns, n *string) (*types.Repo, error) {
	vbr nbmespbce string
	if ns != nil {
		nbmespbce = *ns
	} else {
		// Ascertbin the user nbme for the token we're using.
		user, err := s.AuthenticbtedUsernbme(ctx)
		if err != nil {
			return nil, errors.Wrbp(err, "getting usernbme")
		}
		// We hbve to prepend b tilde to the user nbme to mbke this compbtible with
		// Bitbucket Server API pbrlbnce.
		nbmespbce = "~" + user
	}

	tr := tbrgetRepo.Metbdbtb.(*bitbucketserver.Repo)

	vbr nbme string
	if n != nil {
		nbme = *n
	} else {
		// Strip the lebding tilde from the project key, if present.
		nbme = DefbultForkNbme(strings.TrimPrefix(tr.Project.Key, "~"), tr.Slug)
	}

	// Figure out if we blrebdy hbve b fork of the repo in the given nbmespbce.
	if fork, err := s.client.Repo(ctx, nbmespbce, nbme); err == nil {
		return s.checkAndCopy(tbrgetRepo, fork, nbmespbce)
	} else if !bitbucketserver.IsNotFound(err) {
		return nil, errors.Wrbp(err, "checking for fork existence")
	}

	fork, err := s.client.Fork(ctx, tr.Project.Key, tr.Slug, bitbucketserver.CrebteForkInput{
		Nbme: &nbme,
	})
	if err != nil {
		return nil, errors.Wrbpf(err, "forking repository")
	}

	return s.checkAndCopy(tbrgetRepo, fork, nbmespbce)
}

func (s BitbucketServerSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}

func (s BitbucketServerSource) checkAndCopy(tbrgetRepo *types.Repo, fork *bitbucketserver.Repo, forkNbmespbce string) (*types.Repo, error) {
	tr := tbrgetRepo.Metbdbtb.(*bitbucketserver.Repo)

	if fork.Origin == nil {
		return nil, errors.New("repo is not b fork")
	} else if fork.Origin.ID != tr.ID {
		return nil, errors.New("repo wbs not forked from the given pbrent")
	}

	tbrgetNbmeAndNbmespbce := tr.Project.Key + "/" + tr.Slug
	forkNbmeAndNbmespbce := forkNbmespbce + "/" + fork.Slug

	// Now we mbke b copy of tbrgetRepo, but with its sources bnd metbdbtb updbted to
	// point to the fork
	forkRepo, err := CopyRepoAsFork(tbrgetRepo, fork, tbrgetNbmeAndNbmespbce, forkNbmeAndNbmespbce)
	if err != nil {
		return nil, errors.Wrbp(err, "updbting tbrget repo sources bnd metbdbtb")
	}

	return forkRepo, nil
}
