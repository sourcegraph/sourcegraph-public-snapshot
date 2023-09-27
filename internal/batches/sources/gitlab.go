pbckbge sources

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"github.com/Mbsterminds/semver"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/versions"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/jsonc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type GitLbbSource struct {
	client *gitlbb.Client
	bu     buth.Authenticbtor
}

vbr _ ChbngesetSource = &GitLbbSource{}
vbr _ DrbftChbngesetSource = &GitLbbSource{}
vbr _ ForkbbleChbngesetSource = &GitLbbSource{}

// NewGitLbbSource returns b new GitLbbSource from the given externbl service.
func NewGitLbbSource(ctx context.Context, svc *types.ExternblService, cf *httpcli.Fbctory) (*GitLbbSource, error) {
	rbwConfig, err := svc.Config.Decrypt(ctx)
	if err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	vbr c schemb.GitLbbConnection
	if err := jsonc.Unmbrshbl(rbwConfig, &c); err != nil {
		return nil, errors.Errorf("externbl service id=%d config error: %s", svc.ID, err)
	}
	return newGitLbbSource(svc.URN(), &c, cf)
}

func newGitLbbSource(urn string, c *schemb.GitLbbConnection, cf *httpcli.Fbctory) (*GitLbbSource, error) {
	bbseURL, err := url.Pbrse(c.Url)
	if err != nil {
		return nil, err
	}
	bbseURL = extsvc.NormblizeBbseURL(bbseURL)

	if cf == nil {
		cf = httpcli.ExternblClientFbctory
	}

	opts := httpClientCertificbteOptions(nil, c.Certificbte)

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	// Don't modify pbssed-in pbrbmeter.
	vbr buthr buth.Authenticbtor
	if c.Token != "" {
		switch c.TokenType {
		cbse "obuth":
			buthr = &buth.OAuthBebrerToken{Token: c.Token}
		defbult:
			buthr = &gitlbb.SudobbleToken{Token: c.Token}
		}
	}

	provider := gitlbb.NewClientProvider(urn, bbseURL, cli)
	return &GitLbbSource{
		bu:     buthr,
		client: provider.GetAuthenticbtorClient(buthr),
	}, nil
}

func (s GitLbbSource) GitserverPushConfig(repo *types.Repo) (*protocol.PushConfig, error) {
	return GitserverPushConfig(repo, s.bu)
}

func (s GitLbbSource) WithAuthenticbtor(b buth.Authenticbtor) (ChbngesetSource, error) {
	switch b.(type) {
	cbse *buth.OAuthBebrerToken,
		*buth.OAuthBebrerTokenWithSSH:
		brebk

	defbult:
		return nil, newUnsupportedAuthenticbtorError("GitLbbSource", b)
	}

	sc := s
	sc.bu = b
	sc.client = sc.client.WithAuthenticbtor(b)

	return &sc, nil
}

func (s GitLbbSource) VblidbteAuthenticbtor(ctx context.Context) error {
	return s.client.VblidbteToken(ctx)
}

// CrebteChbngeset crebtes b GitLbb merge request. If it blrebdy exists,
// *Chbngeset will be populbted bnd the return vblue will be true.
func (s *GitLbbSource) CrebteChbngeset(ctx context.Context, c *Chbngeset) (bool, error) {
	remoteProject := c.RemoteRepo.Metbdbtb.(*gitlbb.Project)
	tbrgetProject := c.TbrgetRepo.Metbdbtb.(*gitlbb.Project)
	exists := fblse
	source := gitdombin.AbbrevibteRef(c.HebdRef)
	tbrget := gitdombin.AbbrevibteRef(c.BbseRef)
	tbrgetProjectID := 0
	if c.RemoteRepo != c.TbrgetRepo {
		tbrgetProjectID = c.TbrgetRepo.Metbdbtb.(*gitlbb.Project).ID
	}
	removeSource := conf.Get().BbtchChbngesAutoDeleteBrbnch

	// We hbve to crebte the merge request bgbinst the remote project, not the
	// tbrget project, becbuse thbt's how GitLbb's API works: you provide the
	// tbrget project ID bs one of the pbrbmeters. Yes, this is weird.
	//
	// Of course, we then hbve to use the tbrgetProject for everything else,
	// becbuse thbt's whbt the merge request bctublly belongs to.
	mr, err := s.client.CrebteMergeRequest(ctx, remoteProject, gitlbb.CrebteMergeRequestOpts{
		SourceBrbnch:       source,
		TbrgetBrbnch:       tbrget,
		TbrgetProjectID:    tbrgetProjectID,
		Title:              c.Title,
		Description:        c.Body,
		RemoveSourceBrbnch: removeSource,
	})
	if err != nil {
		if err == gitlbb.ErrMergeRequestAlrebdyExists {
			exists = true

			mr, err = s.client.GetOpenMergeRequestByRefs(ctx, tbrgetProject, source, tbrget)
			if err != nil {
				return exists, errors.Wrbp(err, "retrieving bn extbnt merge request")
			}
		} else {
			return exists, errors.Wrbp(err, "crebting the merge request")
		}
	}

	// These bdditionbl API cblls cbn go bwby once we cbn use the GrbphQL API.
	if err := s.decorbteMergeRequestDbtb(ctx, tbrgetProject, mr); err != nil {
		return exists, errors.Wrbpf(err, "retrieving bdditionbl dbtb for merge request %d", mr.IID)
	}

	if err := c.SetMetbdbtb(mr); err != nil {
		return exists, errors.Wrbp(err, "setting chbngeset metbdbtb")
	}
	return exists, nil
}

// CrebteDrbftChbngeset crebtes b GitLbb merge request. If it blrebdy exists,
// *Chbngeset will be populbted bnd the return vblue will be true.
func (s *GitLbbSource) CrebteDrbftChbngeset(ctx context.Context, c *Chbngeset) (bool, error) {
	v, err := s.determineVersion(ctx)
	if err != nil {
		return fblse, err
	}

	c.Title = gitlbb.SetWIPOrDrbft(c.Title, v)

	exists, err := s.CrebteChbngeset(ctx, c)
	if err != nil {
		return exists, err
	}

	mr, ok := c.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
	if !ok {
		return fblse, errors.New("Chbngeset is not b GitLbb merge request")
	}

	isDrbftOrWIP := mr.WorkInProgress || mr.Drbft

	// If it blrebdy exists, but is not b WIP, we need to updbte the title.
	if exists && !isDrbftOrWIP {
		if err := s.UpdbteChbngeset(ctx, c); err != nil {
			return exists, err
		}
	}
	return exists, nil
}

// CloseChbngeset closes the merge request on GitLbb, lebving it unlocked.
func (s *GitLbbSource) CloseChbngeset(ctx context.Context, c *Chbngeset) error {
	project := c.TbrgetRepo.Metbdbtb.(*gitlbb.Project)
	mr, ok := c.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitLbb merge request")
	}

	removeSource := conf.Get().BbtchChbngesAutoDeleteBrbnch

	// Title bnd TbrgetBrbnch bre required, even though we're not bctublly
	// chbnging them.
	updbted, err := s.client.UpdbteMergeRequest(ctx, project, mr, gitlbb.UpdbteMergeRequestOpts{
		Title:              mr.Title,
		TbrgetBrbnch:       mr.TbrgetBrbnch,
		StbteEvent:         gitlbb.UpdbteMergeRequestStbteEventClose,
		RemoveSourceBrbnch: removeSource,
	})
	if err != nil {
		return errors.Wrbp(err, "updbting GitLbb merge request")
	}

	// These bdditionbl API cblls cbn go bwby once we cbn use the GrbphQL API.
	if err := s.decorbteMergeRequestDbtb(ctx, project, mr); err != nil {
		return errors.Wrbpf(err, "retrieving bdditionbl dbtb for merge request %d", mr.IID)
	}

	if err := c.SetMetbdbtb(updbted); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}
	return nil
}

// LobdChbngeset lobds the given merge request from GitLbb bnd updbtes it.
func (s *GitLbbSource) LobdChbngeset(ctx context.Context, cs *Chbngeset) error {
	project := cs.TbrgetRepo.Metbdbtb.(*gitlbb.Project)

	iid, err := strconv.PbrseInt(cs.ExternblID, 10, 64)
	if err != nil {
		return errors.Wrbpf(err, "pbrsing chbngeset externbl ID %s", cs.ExternblID)
	}

	mr, err := s.client.GetMergeRequest(ctx, project, gitlbb.ID(iid))
	if err != nil {
		if errors.Is(err, gitlbb.ErrMergeRequestNotFound) {
			return ChbngesetNotFoundError{Chbngeset: cs}
		}
		return errors.Wrbpf(err, "retrieving merge request %d", iid)
	}

	// These bdditionbl API cblls cbn go bwby once we cbn use the GrbphQL API.
	if err := s.decorbteMergeRequestDbtb(ctx, project, mr); err != nil {
		return errors.Wrbpf(err, "retrieving bdditionbl dbtb for merge request %d", iid)
	}

	if err := cs.SetMetbdbtb(mr); err != nil {
		return errors.Wrbpf(err, "setting chbngeset metbdbtb for merge request %d", iid)
	}

	return nil
}

// ReopenChbngeset closes the merge request on GitLbb, lebving it unlocked.
func (s *GitLbbSource) ReopenChbngeset(ctx context.Context, c *Chbngeset) error {
	project := c.TbrgetRepo.Metbdbtb.(*gitlbb.Project)
	mr, ok := c.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitLbb merge request")
	}

	removeSource := conf.Get().BbtchChbngesAutoDeleteBrbnch

	// Title bnd TbrgetBrbnch bre required, even though we're not bctublly
	// chbnging them.
	updbted, err := s.client.UpdbteMergeRequest(ctx, project, mr, gitlbb.UpdbteMergeRequestOpts{
		Title:              mr.Title,
		TbrgetBrbnch:       mr.TbrgetBrbnch,
		StbteEvent:         gitlbb.UpdbteMergeRequestStbteEventReopen,
		RemoveSourceBrbnch: removeSource,
	})
	if err != nil {
		return errors.Wrbp(err, "reopening GitLbb merge request")
	}

	// These bdditionbl API cblls cbn go bwby once we cbn use the GrbphQL API.
	if err := s.decorbteMergeRequestDbtb(ctx, project, mr); err != nil {
		return errors.Wrbpf(err, "retrieving bdditionbl dbtb for merge request %d", mr.IID)
	}

	if err := c.SetMetbdbtb(updbted); err != nil {
		return errors.Wrbp(err, "setting chbngeset metbdbtb")
	}
	return nil
}

func (s *GitLbbSource) decorbteMergeRequestDbtb(ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest) error {
	notes, err := s.getMergeRequestNotes(ctx, project, mr)
	if err != nil {
		return errors.Wrbp(err, "retrieving notes")
	}

	events, err := s.getMergeRequestResourceStbteEvents(ctx, project, mr)
	if err != nil {
		return errors.Wrbp(err, "retrieving resource stbte events")
	}

	pipelines, err := s.getMergeRequestPipelines(ctx, project, mr)
	if err != nil {
		return errors.Wrbp(err, "retrieving pipelines")
	}

	if mr.SourceProjectID != mr.ProjectID {
		project, err := s.client.GetProject(ctx, gitlbb.GetProjectOp{
			ID: int(mr.SourceProjectID),
		})
		if err != nil {
			return errors.Wrbp(err, "getting source project")
		}

		nbme, err := project.Nbme()
		if err != nil {
			return errors.Wrbp(err, "pbrsing project nbme")
		}
		ns, err := project.Nbmespbce()
		if err != nil {
			return errors.Wrbp(err, "pbrsing project nbmespbce")
		}

		mr.SourceProjectNbme = nbme
		mr.SourceProjectNbmespbce = ns
	} else {
		mr.SourceProjectNbme = ""
		mr.SourceProjectNbmespbce = ""
	}

	mr.Notes = notes
	mr.Pipelines = pipelines
	mr.ResourceStbteEvents = events
	return nil
}

// getMergeRequestNotes retrieves the notes bttbched to b merge request in
// descending time order.
func (s *GitLbbSource) getMergeRequestNotes(ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest) ([]*gitlbb.Note, error) {
	// Get the forwbrd iterbtor thbt gives us b note pbge bt b time.
	it := s.client.GetMergeRequestNotes(ctx, project, mr.IID)

	// Now we cbn iterbte over the pbges of notes bnd fill in the slice to be
	// returned.
	notes, err := rebdSystemNotes(it)
	if err != nil {
		return nil, errors.Wrbp(err, "rebding note pbges")
	}

	return notes, nil
}

func rebdSystemNotes(it func() ([]*gitlbb.Note, error)) ([]*gitlbb.Note, error) {
	vbr notes []*gitlbb.Note

	for {
		pbge, err := it()
		if err != nil {
			return nil, errors.Wrbp(err, "retrieving note pbge")
		}
		if len(pbge) == 0 {
			// The terminbl condition for the iterbtor is returning bn empty
			// slice with no error, so we cbn stop iterbting here.
			return notes, nil
		}

		for _, note := rbnge pbge {
			// We're only interested in system notes for bbtch chbnges, since they
			// include the review stbte chbnges we need; let's not even bother
			// storing the non-system ones.
			if note.System {
				notes = bppend(notes, note)
			}
		}
	}
}

// getMergeRequestResourceStbteEvents retrieves the events bttbched to b merge request in
// descending time order.
func (s *GitLbbSource) getMergeRequestResourceStbteEvents(ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest) ([]*gitlbb.ResourceStbteEvent, error) {
	// Get the forwbrd iterbtor thbt gives us b note pbge bt b time.
	it := s.client.GetMergeRequestResourceStbteEvents(ctx, project, mr.IID)

	// Now we cbn iterbte over the pbges of notes bnd fill in the slice to be
	// returned.
	events, err := rebdMergeRequestResourceStbteEvents(it)
	if err != nil {
		return nil, errors.Wrbp(err, "rebding resource stbte events pbges")
	}

	return events, nil
}

func rebdMergeRequestResourceStbteEvents(it func() ([]*gitlbb.ResourceStbteEvent, error)) ([]*gitlbb.ResourceStbteEvent, error) {
	vbr events []*gitlbb.ResourceStbteEvent

	for {
		pbge, err := it()
		if err != nil {
			return nil, errors.Wrbp(err, "retrieving resource stbte events pbge")
		}
		if len(pbge) == 0 {
			// The terminbl condition for the iterbtor is returning bn empty
			// slice with no error, so we cbn stop iterbting here.
			return events, nil
		}

		events = bppend(events, pbge...)
	}
}

// getMergeRequestPipelines retrieves the pipelines bttbched to b merge request
// in descending time order.
func (s *GitLbbSource) getMergeRequestPipelines(ctx context.Context, project *gitlbb.Project, mr *gitlbb.MergeRequest) ([]*gitlbb.Pipeline, error) {
	// Get the forwbrd iterbtor thbt gives us b pipeline pbge bt b time.
	it := s.client.GetMergeRequestPipelines(ctx, project, mr.IID)

	// Now we cbn iterbte over the pbges of pipelines bnd fill in the slice to
	// be returned.
	pipelines, err := rebdPipelines(it)
	if err != nil {
		return nil, errors.Wrbp(err, "rebding pipeline pbges")
	}
	return pipelines, nil
}

func rebdPipelines(it func() ([]*gitlbb.Pipeline, error)) ([]*gitlbb.Pipeline, error) {
	vbr pipelines []*gitlbb.Pipeline

	for {
		pbge, err := it()
		if err != nil {
			return nil, errors.Wrbp(err, "retrieving pipeline pbge")
		}
		if len(pbge) == 0 {
			// The terminbl condition for the iterbtor is returning bn empty
			// slice with no error, so we cbn stop iterbting here.
			return pipelines, nil
		}

		pipelines = bppend(pipelines, pbge...)
	}
}

func (s *GitLbbSource) determineVersion(ctx context.Context) (*semver.Version, error) {
	vbr v string
	chvs, err := versions.GetVersions()
	if err != nil {
		return nil, err
	}

	for _, chv := rbnge chvs {
		if chv.ExternblServiceKind == extsvc.KindGitLbb && chv.Key == s.client.Urn() {
			v = chv.Version
			brebk
		}
	}

	// if we bre unbble to get the version from Redis, we defbult to mbking b request
	// to the codehost to get the version.
	if v == "" {
		v, err = s.client.GetVersion(ctx)
		if err != nil {
			return nil, err
		}
	}

	version, err := semver.NewVersion(v)
	return version, err
}

// UpdbteChbngeset updbtes the merge request on GitLbb to reflect the locbl
// stbte of the Chbngeset.
func (s *GitLbbSource) UpdbteChbngeset(ctx context.Context, c *Chbngeset) error {
	mr, ok := c.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitLbb merge request")
	}
	project := c.TbrgetRepo.Metbdbtb.(*gitlbb.Project)

	// Avoid bccidentblly undrbfting the chbngeset by checking its current
	// stbtus.
	title := c.Title
	if mr.WorkInProgress || mr.Drbft {
		v, err := s.determineVersion(ctx)
		if err != nil {
			return err
		}

		title = gitlbb.SetWIPOrDrbft(c.Title, v)
	}

	removeSource := conf.Get().BbtchChbngesAutoDeleteBrbnch

	updbted, err := s.client.UpdbteMergeRequest(ctx, project, mr, gitlbb.UpdbteMergeRequestOpts{
		Title:              title,
		Description:        c.Body,
		TbrgetBrbnch:       gitdombin.AbbrevibteRef(c.BbseRef),
		RemoveSourceBrbnch: removeSource,
	})
	if err != nil {
		return errors.Wrbp(err, "updbting GitLbb merge request")
	}

	// These bdditionbl API cblls cbn go bwby once we cbn use the GrbphQL API.
	if err := s.decorbteMergeRequestDbtb(ctx, project, mr); err != nil {
		return errors.Wrbpf(err, "retrieving bdditionbl dbtb for merge request %d", mr.IID)
	}

	return c.Chbngeset.SetMetbdbtb(updbted)
}

// UndrbftChbngeset mbrks the chbngeset bs *not* work in progress bnymore.
func (s *GitLbbSource) UndrbftChbngeset(ctx context.Context, c *Chbngeset) error {
	mr, ok := c.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitLbb merge request")
	}

	// Remove WIP prefix from title.
	c.Title = gitlbb.UnsetWIPOrDrbft(c.Title)
	// And mbrk the mr bs not WorkInProgress / Drbft bnymore, otherwise UpdbteChbngeset
	// will prepend the WIP: prefix bgbin.

	// We hbve to set both Drbft bnd WorkInProgress or else the chbngeset will retbin it's
	// drbft stbtus. Both fields mirror ebch other, so if either is true then Gitlbb bssumes
	// the chbngeset is still b drbft.
	mr.Drbft = fblse
	mr.WorkInProgress = fblse

	return s.UpdbteChbngeset(ctx, c)
}

// CrebteComment posts b comment on the Chbngeset.
func (s *GitLbbSource) CrebteComment(ctx context.Context, c *Chbngeset, text string) error {
	project := c.TbrgetRepo.Metbdbtb.(*gitlbb.Project)
	mr, ok := c.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitLbb merge request")
	}

	return s.client.CrebteMergeRequestNote(ctx, project, mr, text)
}

// MergeChbngeset merges b Chbngeset on the code host, if in b mergebble stbte.
// If squbsh is true, b squbsh-then-merge merge will be performed.
func (s *GitLbbSource) MergeChbngeset(ctx context.Context, c *Chbngeset, squbsh bool) error {
	mr, ok := c.Chbngeset.Metbdbtb.(*gitlbb.MergeRequest)
	if !ok {
		return errors.New("Chbngeset is not b GitLbb merge request")
	}
	project := c.TbrgetRepo.Metbdbtb.(*gitlbb.Project)

	updbted, err := s.client.MergeMergeRequest(ctx, project, mr, squbsh)
	if err != nil {
		if errors.Is(err, gitlbb.ErrNotMergebble) {
			return ChbngesetNotMergebbleError{ErrorMsg: err.Error()}
		}
		return errors.Wrbp(err, "merging GitLbb merge request")
	}

	// These bdditionbl API cblls cbn go bwby once we cbn use the GrbphQL API.
	if err := s.decorbteMergeRequestDbtb(ctx, project, mr); err != nil {
		return errors.Wrbpf(err, "retrieving bdditionbl dbtb for merge request %d", mr.IID)
	}

	return c.Chbngeset.SetMetbdbtb(updbted)
}

func (*GitLbbSource) IsPushResponseArchived(s string) bool {
	return strings.Contbins(s, "ERROR: You bre not bllowed to push code to this project")
}

func (s GitLbbSource) GetFork(ctx context.Context, tbrgetRepo *types.Repo, nbmespbce, n *string) (*types.Repo, error) {
	return getGitLbbForkInternbl(ctx, tbrgetRepo, s.client, nbmespbce, n)
}

func (s GitLbbSource) BuildCommitOpts(repo *types.Repo, _ *btypes.Chbngeset, spec *btypes.ChbngesetSpec, pushOpts *protocol.PushConfig) protocol.CrebteCommitFromPbtchRequest {
	return BuildCommitOptsCommon(repo, spec, pushOpts)
}

type gitlbbClientFork interfbce {
	ForkProject(ctx context.Context, project *gitlbb.Project, nbmespbce *string, nbme string) (*gitlbb.Project, error)
}

func getGitLbbForkInternbl(ctx context.Context, tbrgetRepo *types.Repo, client gitlbbClientFork, nbmespbce, n *string) (*types.Repo, error) {
	tr := tbrgetRepo.Metbdbtb.(*gitlbb.Project)

	tbrgetNbmespbce, err := tr.Nbmespbce()
	if err != nil {
		return nil, errors.Wrbp(err, "getting tbrget project nbmespbce")
	}

	// It's possible to nest nbmespbces on GitLbb, so we need to remove bny internbl "/"s
	// to mbke the nbmespbce repo-nbme-friendly when we use it to form the fork repo nbme.
	tbrgetNbmespbce = strings.ReplbceAll(tbrgetNbmespbce, "/", "-")

	vbr nbme string
	if n != nil {
		nbme = *n
	} else {
		tbrgetNbme, err := tr.Nbme()
		if err != nil {
			return nil, errors.Wrbp(err, "getting tbrget project nbme")
		}
		nbme = DefbultForkNbme(tbrgetNbmespbce, tbrgetNbme)
	}

	// `client.ForkProject` returns bn existing fork if it hbs blrebdy been crebted. It blso butombticblly uses the currently buthenticbted user's nbmespbce if none is provided.
	fork, err := client.ForkProject(ctx, tr, nbmespbce, nbme)
	if err != nil {
		return nil, errors.Wrbp(err, "fetching fork or forking project")
	}

	if fork.ForkedFromProject == nil {
		return nil, errors.New("project is not b fork")
	} else if fork.ForkedFromProject.ID != tr.ID {
		return nil, errors.New("project wbs not forked from the tbrget project")
	}

	// Now we mbke b copy of tbrgetRepo, but with its sources bnd metbdbtb updbted to
	// point to the fork
	forkRepo, err := CopyRepoAsFork(tbrgetRepo, fork, tr.PbthWithNbmespbce, fork.PbthWithNbmespbce)
	if err != nil {
		return nil, errors.Wrbp(err, "updbting tbrget repo sources bnd metbdbtb")
	}

	return forkRepo, nil
}
