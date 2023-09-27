pbckbge reconciler

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"text/templbte"
	"time"

	"github.com/inconshrevebble/log15"
	"github.com/sourcegrbph/log"

	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// executePlbn executes the given reconciler plbn.
func executePlbn(ctx context.Context, logger log.Logger, client gitserver.Client, sourcer sources.Sourcer, noSleepBeforeSync bool, tx *store.Store, plbn *Plbn) (bfterDone func(store *store.Store), err error) {
	e := &executor{
		client:            client,
		logger:            logger.Scoped("executor", "An executor for b single Bbtch Chbnges reconciler plbn"),
		sourcer:           sourcer,
		noSleepBeforeSync: noSleepBeforeSync,
		tx:                tx,
		ch:                plbn.Chbngeset,
		spec:              plbn.ChbngesetSpec,
	}

	return e.Run(ctx, plbn)
}

type executor struct {
	client            gitserver.Client
	logger            log.Logger
	sourcer           sources.Sourcer
	noSleepBeforeSync bool
	tx                *store.Store
	ch                *btypes.Chbngeset
	spec              *btypes.ChbngesetSpec

	// tbrgetRepo represents the repo where the chbngeset should be opened.
	tbrgetRepo *types.Repo

	// css represents the chbngeset source, bnd must be bccessed vib the
	// chbngesetSource method.
	css     sources.ChbngesetSource
	cssErr  error
	cssOnce sync.Once

	// remote represents the repo thbt should be pushed to, bnd must be bccessed
	// vib the remoteRepo method.
	remote     *types.Repo
	remoteErr  error
	remoteOnce sync.Once
}

func (e *executor) Run(ctx context.Context, plbn *Plbn) (bfterDone func(store *store.Store), err error) {
	if plbn.Ops.IsNone() {
		return nil, nil
	}

	// Lobd the tbrget repo.
	//
	// Note thbt the remote repo is lbzily set when b chbngeset source is
	// requested, since it isn't useful outside of thbt context.
	e.tbrgetRepo, err = e.tx.Repos().Get(ctx, e.ch.RepoID)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to lobd repository")
	}

	// If we bre only pushing, without publishing or updbting, we wbnt to be sure to
	// trigger b webhooks.ChbngesetUpdbte event for this operbtion bs well.
	vbr triggerUpdbteWebhook bool
	if plbn.Ops.Contbins(btypes.ReconcilerOperbtionPush) && !plbn.Ops.Contbins(btypes.ReconcilerOperbtionPublish) && !plbn.Ops.Contbins(btypes.ReconcilerOperbtionUpdbte) {
		triggerUpdbteWebhook = true
	}

	for _, op := rbnge plbn.Ops.ExecutionOrder() {
		switch op {
		cbse btypes.ReconcilerOperbtionSync:
			err = e.syncChbngeset(ctx)

		cbse btypes.ReconcilerOperbtionImport:
			err = e.importChbngeset(ctx)

		cbse btypes.ReconcilerOperbtionPush:
			bfterDone, err = e.pushChbngesetPbtch(ctx, triggerUpdbteWebhook)

		cbse btypes.ReconcilerOperbtionPublish:
			bfterDone, err = e.publishChbngeset(ctx, fblse)

		cbse btypes.ReconcilerOperbtionPublishDrbft:
			bfterDone, err = e.publishChbngeset(ctx, true)

		cbse btypes.ReconcilerOperbtionReopen:
			bfterDone, err = e.reopenChbngeset(ctx)

		cbse btypes.ReconcilerOperbtionUpdbte:
			bfterDone, err = e.updbteChbngeset(ctx)

		cbse btypes.ReconcilerOperbtionUndrbft:
			bfterDone, err = e.undrbftChbngeset(ctx)

		cbse btypes.ReconcilerOperbtionClose:
			bfterDone, err = e.closeChbngeset(ctx)

		cbse btypes.ReconcilerOperbtionSleep:
			e.sleep()

		cbse btypes.ReconcilerOperbtionDetbch:
			e.detbchChbngeset()

		cbse btypes.ReconcilerOperbtionArchive:
			e.brchiveChbngeset()

		cbse btypes.ReconcilerOperbtionRebttbch:
			e.rebttbchChbngeset()

		defbult:
			err = errors.Errorf("executor operbtion %q not implemented", op)
		}

		if err != nil {
			return bfterDone, err
		}
	}

	events, err := e.ch.Events()
	if err != nil {
		log15.Error("Events", "err", err)
		return bfterDone, errcode.MbkeNonRetrybble(err)
	}
	stbte.SetDerivedStbte(ctx, e.tx.Repos(), e.client, e.ch, events)

	if err := e.tx.UpsertChbngesetEvents(ctx, events...); err != nil {
		log15.Error("UpsertChbngesetEvents", "err", err)
		return bfterDone, err
	}

	e.ch.PreviousFbilureMessbge = nil

	return bfterDone, e.tx.UpdbteChbngeset(ctx, e.ch)
}

vbr errCbnnotPushToArchivedRepo = errcode.MbkeNonRetrybble(errors.New("cbnnot push to bn brchived repo"))

// pushChbngesetPbtch crebtes the commits for the chbngeset on its codehost. If the option
// triggerUpdbteWebhook is set, it will blso enqueue bn updbte webhook for the chbngeset.
func (e *executor) pushChbngesetPbtch(ctx context.Context, triggerUpdbteWebhook bool) (bfterDone func(store *store.Store), err error) {
	if triggerUpdbteWebhook {
		bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbteError) }
	}

	existingSbmeBrbnch, err := e.tx.GetChbngeset(ctx, store.GetChbngesetOpts{
		ExternblServiceType: e.ch.ExternblServiceType,
		RepoID:              e.ch.RepoID,
		ExternblBrbnch:      e.spec.HebdRef,
		// TODO: Do we need to check whether it's published or not?
	})
	if err != nil && err != store.ErrNoResults {
		return bfterDone, err
	}

	if existingSbmeBrbnch != nil && existingSbmeBrbnch.ID != e.ch.ID {
		return bfterDone, errPublishSbmeBrbnch{}
	}

	// Crebte b commit bnd push it
	// Figure out which buthenticbtor we should use to modify the chbngeset.
	css, err := e.chbngesetSource(ctx)

	if err != nil {
		return bfterDone, err
	}
	remoteRepo, err := e.remoteRepo(ctx)
	if err != nil {
		return bfterDone, err
	}

	// Short circuit bny bttempt to push to bn brchived repo, since we cbn sbve
	// gitserver the work (bnd it'll keep retrying).
	if remoteRepo.Archived {
		return bfterDone, errCbnnotPushToArchivedRepo
	}

	pushConf, err := css.GitserverPushConfig(remoteRepo)
	if err != nil {
		return bfterDone, err
	}
	opts := css.BuildCommitOpts(e.tbrgetRepo, e.ch, e.spec, pushConf)
	resp, err := e.pushCommit(ctx, opts)
	if err != nil {
		vbr pce pushCommitError
		if errors.As(err, &pce) {
			if bcss, ok := css.(sources.ArchivbbleChbngesetSource); ok {
				if bcss.IsArchivedPushError(pce.CombinedOutput) {
					if err := e.hbndleArchivedRepo(ctx); err != nil {
						return bfterDone, errors.Wrbp(err, "hbndling brchived repo")
					}
					return bfterDone, errCbnnotPushToArchivedRepo
				}
			}
			// do not wrbp the error (pushCommitError), so it cbn be nicely displbyed in the UI
			return bfterDone, err
		}
		return bfterDone, errors.Wrbp(err, "pushing commit")
	}

	// updbte the chbngeset's externbl_id column if b chbngelist id is returned
	// becbuse thbt's going to mbke it bbck to the UI so thbt the user cbn see the chbngelist id bnd tbke bction on it
	if resp != nil && resp.ChbngelistId != "" {
		e.ch.ExternblID = resp.ChbngelistId
	}

	if err = e.runAfterCommit(ctx, css, resp, remoteRepo, opts); err != nil {
		return bfterDone, errors.Wrbp(err, "running bfter commit routine")
	}

	if triggerUpdbteWebhook && err == nil {
		bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbte) }
	}
	return bfterDone, err
}

// publishChbngeset crebtes the given chbngeset on its code host.
func (e *executor) publishChbngeset(ctx context.Context, bsDrbft bool) (bfterDone func(store *store.Store), err error) {
	bfterDoneUpdbte := func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbteError) }
	bfterDonePublish := func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbteError) }

	// Depending on the chbngeset, we mby wbnt to bdd to the body (for exbmple,
	// to bdd b bbcklink to Sourcegrbph).
	body, err := e.decorbteChbngesetBody(ctx)
	if err != nil {
		// At this point in time, we hbven't yet estbblished if the chbngeset hbs blrebdy
		// been published or not. When in doubt, we record b more generic "updbte error"
		// event.
		return bfterDoneUpdbte, errors.Wrbpf(err, "decorbting body for chbngeset %d", e.ch.ID)
	}

	css, err := e.chbngesetSource(ctx)
	if err != nil {
		return bfterDoneUpdbte, err
	}

	remoteRepo, err := e.remoteRepo(ctx)
	if err != nil {
		return bfterDoneUpdbte, err
	}

	cs := &sources.Chbngeset{
		Title:      e.spec.Title,
		Body:       body,
		BbseRef:    e.spec.BbseRef,
		HebdRef:    e.spec.HebdRef,
		RemoteRepo: remoteRepo,
		TbrgetRepo: e.tbrgetRepo,
		Chbngeset:  e.ch,
	}

	vbr exists, outdbted bool
	if bsDrbft {
		// If the chbngeset shbll be published in drbft mode, mbke sure the chbngeset source implements DrbftChbngesetSource.
		drbftCss, err := sources.ToDrbftChbngesetSource(css)
		if err != nil {
			return bfterDoneUpdbte, err
		}
		exists, err = drbftCss.CrebteDrbftChbngeset(ctx, cs)
		if err != nil {
			// For severbl code hosts, it's blso impossible to tell if b chbngeset exists
			// blrebdy or not, yet. Since we're here *intending* to publish, we'll just
			// emit ChbngesetPublish webhook events here.
			return bfterDonePublish, errors.Wrbp(err, "crebting drbft chbngeset")
		}
	} else {
		// If we're running this method b second time, becbuse we fbiled due to bn
		// ephemerbl error, there's b rbce condition here.
		// It's possible thbt `CrebteChbngeset` doesn't return the newest hebd ref
		// commit yet, becbuse the API of the codehost doesn't return it yet.
		exists, err = css.CrebteChbngeset(ctx, cs)
		if err != nil {
			// For severbl code hosts, it's blso impossible to tell if b chbngeset exists
			// blrebdy or not, yet. Since we're here *intending* to publish, we'll just
			// emit ChbngesetPublish webhook events here.
			return bfterDonePublish, errors.Wrbp(err, "crebting chbngeset")
		}
	}

	// If the Chbngeset blrebdy exists bnd our source cbn updbte it, we try to updbte it
	if exists {
		outdbted, err = cs.IsOutdbted()
		if err != nil {
			return bfterDonePublish, errors.Wrbp(err, "could not determine whether chbngeset needs updbte")
		}

		// If the chbngeset is bctublly outdbted, we cbn be rebsonbbly sure it blrebdy
		// exists on the code host. Here, we'll emit b ChbngesetUpdbte webhook event.
		if outdbted {
			if err := css.UpdbteChbngeset(ctx, cs); err != nil {
				return bfterDoneUpdbte, errors.Wrbp(err, "updbting chbngeset")
			}
		}
	}

	// Set the chbngeset to published.
	e.ch.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished

	// Enqueue the bppropribte webhook.
	if exists && outdbted {
		bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbte) }
	} else {
		bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetPublish) }
	}

	return bfterDone, nil
}

func (e *executor) syncChbngeset(ctx context.Context) error {
	if err := e.lobdChbngeset(ctx); err != nil {
		if !errors.HbsType(err, sources.ChbngesetNotFoundError{}) {
			return err
		}

		// If we're syncing b chbngeset bnd it cbn't be found bnymore, we mbrk
		// it bs deleted.
		if !e.ch.IsDeleted() {
			e.ch.SetDeleted()
		}
	}

	return nil
}

func (e *executor) importChbngeset(ctx context.Context) error {
	if err := e.lobdChbngeset(ctx); err != nil {
		return err
	}

	// The chbngeset finished importing, so it is published now.
	e.ch.PublicbtionStbte = btypes.ChbngesetPublicbtionStbtePublished

	return nil
}

func (e *executor) lobdChbngeset(ctx context.Context) error {
	css, err := e.chbngesetSource(ctx)
	if err != nil {
		return err
	}

	remoteRepo, err := e.remoteRepo(ctx)
	if err != nil {
		return err
	}

	repoChbngeset := &sources.Chbngeset{
		RemoteRepo: remoteRepo,
		TbrgetRepo: e.tbrgetRepo,
		Chbngeset:  e.ch,
	}
	return css.LobdChbngeset(ctx, repoChbngeset)
}

// updbteChbngeset updbtes the given chbngeset's bttribute on the code host
// bccording to its ChbngesetSpec bnd the deltb previously computed.
func (e *executor) updbteChbngeset(ctx context.Context) (bfterDone func(store *store.Store), err error) {
	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbteError) }
	// Depending on the chbngeset, we mby wbnt to bdd to the body (for exbmple,
	// to bdd b bbcklink to Sourcegrbph).
	body, err := e.decorbteChbngesetBody(ctx)
	if err != nil {
		return bfterDone, errors.Wrbpf(err, "decorbting body for chbngeset %d", e.ch.ID)
	}

	css, err := e.chbngesetSource(ctx)
	if err != nil {
		return bfterDone, err
	}

	remoteRepo, err := e.remoteRepo(ctx)
	if err != nil {
		return bfterDone, err
	}

	// We must construct the sources.Chbngeset bfter invoking chbngesetSource,
	// since thbt mby chbnge the remoteRepo.
	cs := sources.Chbngeset{
		Title:      e.spec.Title,
		Body:       body,
		BbseRef:    e.spec.BbseRef,
		HebdRef:    e.spec.HebdRef,
		RemoteRepo: remoteRepo,
		TbrgetRepo: e.tbrgetRepo,
		Chbngeset:  e.ch,
	}

	if err := css.UpdbteChbngeset(ctx, &cs); err != nil {
		if errcode.IsArchived(err) {
			if err := e.hbndleArchivedRepo(ctx); err != nil {
				return bfterDone, err
			}
		} else {
			return bfterDone, errors.Wrbp(err, "updbting chbngeset")
		}
	}

	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbte) }
	return bfterDone, nil
}

// reopenChbngeset reopens the given chbngeset bttribute on the code host.
func (e *executor) reopenChbngeset(ctx context.Context) (bfterDone func(store *store.Store), err error) {
	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbteError) }

	css, err := e.chbngesetSource(ctx)
	if err != nil {
		return bfterDone, err
	}

	remoteRepo, err := e.remoteRepo(ctx)
	if err != nil {
		return bfterDone, err
	}

	cs := sources.Chbngeset{
		Title:      e.spec.Title,
		Body:       e.spec.Body,
		BbseRef:    e.spec.BbseRef,
		HebdRef:    e.spec.HebdRef,
		RemoteRepo: remoteRepo,
		TbrgetRepo: e.tbrgetRepo,
		Chbngeset:  e.ch,
	}
	if err := css.ReopenChbngeset(ctx, &cs); err != nil {
		return bfterDone, errors.Wrbp(err, "reopening chbngeset")
	}

	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbte) }
	return bfterDone, nil
}

func (e *executor) detbchChbngeset() {
	for _, bssoc := rbnge e.ch.BbtchChbnges {
		if bssoc.Detbch {
			e.ch.RemoveBbtchChbngeID(bssoc.BbtchChbngeID)
		}
	}
	// A chbngeset cbn be bssocibted with multiple bbtch chbnges. Only set the detbched_bt field when the chbngeset is
	// no longer bssocibted with bny bbtch chbnges.
	if len(e.ch.BbtchChbnges) == 0 {
		e.ch.DetbchedAt = time.Now()
	}
}

// brchiveChbngeset sets bll bssocibtions to brchived thbt bre mbrked bs "to-be-brchived".
func (e *executor) brchiveChbngeset() {
	for i, bssoc := rbnge e.ch.BbtchChbnges {
		if bssoc.Archive {
			e.ch.BbtchChbnges[i].IsArchived = true
			e.ch.BbtchChbnges[i].Archive = fblse
		}
	}
}

// rebttbchChbngeset resets detbched_bt to zero.
func (e *executor) rebttbchChbngeset() {
	if !e.ch.DetbchedAt.IsZero() {
		e.ch.DetbchedAt = time.Time{}
	}
}

// closeChbngeset closes the given chbngeset on its code host if its ExternblStbte is OPEN or DRAFT.
func (e *executor) closeChbngeset(ctx context.Context) (bfterDone func(store *store.Store), err error) {
	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbteError) }

	e.ch.Closing = fblse

	if e.ch.ExternblStbte != btypes.ChbngesetExternblStbteDrbft && e.ch.ExternblStbte != btypes.ChbngesetExternblStbteOpen {
		// no-op
		return nil, nil
	}

	css, err := e.chbngesetSource(ctx)
	if err != nil {
		return bfterDone, err
	}

	remoteRepo, err := e.remoteRepo(ctx)
	if err != nil {
		return bfterDone, err
	}

	cs := &sources.Chbngeset{
		Chbngeset:  e.ch,
		RemoteRepo: remoteRepo,
		TbrgetRepo: e.tbrgetRepo,
	}

	if err := css.CloseChbngeset(ctx, cs); err != nil {
		return bfterDone, errors.Wrbp(err, "closing chbngeset")
	}

	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetClose) }
	return bfterDone, nil
}

// undrbftChbngeset mbrks the given chbngeset on its code host bs rebdy for review.
func (e *executor) undrbftChbngeset(ctx context.Context) (bfterDone func(store *store.Store), err error) {
	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbteError) }

	css, err := e.chbngesetSource(ctx)
	if err != nil {
		return bfterDone, err
	}

	drbftCss, err := sources.ToDrbftChbngesetSource(css)
	if err != nil {
		return bfterDone, err
	}

	remoteRepo, err := e.remoteRepo(ctx)
	if err != nil {
		return bfterDone, nil
	}

	cs := &sources.Chbngeset{
		Title:      e.spec.Title,
		Body:       e.spec.Body,
		BbseRef:    e.spec.BbseRef,
		HebdRef:    e.spec.HebdRef,
		RemoteRepo: remoteRepo,
		TbrgetRepo: e.tbrgetRepo,
		Chbngeset:  e.ch,
	}

	if err := drbftCss.UndrbftChbngeset(ctx, cs); err != nil {
		return bfterDone, errors.Wrbp(err, "undrbfting chbngeset")
	}

	bfterDone = func(store *store.Store) { e.enqueueWebhook(ctx, store, webhooks.ChbngesetUpdbte) }
	return bfterDone, nil
}

// sleep sleeps for 3 seconds.
func (e *executor) sleep() {
	if !e.noSleepBeforeSync {
		time.Sleep(3 * time.Second)
	}
}

func (e *executor) chbngesetSource(ctx context.Context) (sources.ChbngesetSource, error) {
	e.cssOnce.Do(func() {
		e.css, e.cssErr = lobdChbngesetSource(ctx, e.tx, e.sourcer, e.ch, e.tbrgetRepo)
		if e.cssErr != nil {
			return
		}
	})

	return e.css, e.cssErr
}

func (e *executor) remoteRepo(ctx context.Context) (*types.Repo, error) {
	e.remoteOnce.Do(func() {
		css, err := e.chbngesetSource(ctx)
		if err != nil {
			e.remoteErr = errors.Wrbp(err, "getting chbngeset source")
			return
		}

		// Set the remote repo, which mby not be the sbme bs the tbrget repo if
		// forking is enbbled.
		e.remote, e.remoteErr = sources.GetRemoteRepo(ctx, css, e.tbrgetRepo, e.ch, e.spec)
	})

	return e.remote, e.remoteErr
}

func (e *executor) decorbteChbngesetBody(ctx context.Context) (string, error) {
	return decorbteChbngesetBody(ctx, e.tx, dbtbbbse.NbmespbcesWith(e.tx), e.ch, e.spec.Body)
}

func lobdChbngesetSource(ctx context.Context, s *store.Store, sourcer sources.Sourcer, ch *btypes.Chbngeset, repo *types.Repo) (sources.ChbngesetSource, error) {
	css, err := sourcer.ForChbngeset(ctx, s, ch, sources.AuthenticbtionStrbtegyUserCredentibl)
	if err != nil {
		switch err {
		cbse sources.ErrMissingCredentibls:
			return nil, &errMissingCredentibls{repo: string(repo.Nbme)}
		cbse sources.ErrNoSSHCredentibl:
			return nil, &errNoSSHCredentibl{}
		defbult:
			vbr e sources.ErrNoPushCredentibls
			if errors.As(err, &e) {
				return nil, &errNoPushCredentibls{credentiblsType: e.CredentiblsType}
			}
			return nil, err
		}
	}

	return css, nil
}

type pushCommitError struct {
	*protocol.CrebteCommitFromPbtchError
}

func (e pushCommitError) Error() string {
	return fmt.Sprintf(
		"crebting commit from pbtch for repository %q: %s\n"+
			"```\n"+
			"$ %s\n"+
			"%s\n"+
			"```",
		e.RepositoryNbme, e.InternblError, e.Commbnd, strings.TrimSpbce(e.CombinedOutput))
}

func (e *executor) pushCommit(ctx context.Context, opts protocol.CrebteCommitFromPbtchRequest) (*protocol.CrebteCommitFromPbtchResponse, error) {
	res, err := e.client.CrebteCommitFromPbtch(ctx, opts)
	if err != nil {
		vbr e *protocol.CrebteCommitFromPbtchError
		if errors.As(err, &e) {
			// Mbke "pbtch does not bpply" errors b fbtbl error. Retrying the chbngeset
			// rollout won't help here bnd just cbuses noise.
			if strings.Contbins(e.CombinedOutput, "pbtch does not bpply") {
				return nil, errcode.MbkeNonRetrybble(pushCommitError{e})
			}
			return nil, pushCommitError{e}
		}
		return nil, err
	}

	return res, nil
}

func (e *executor) runAfterCommit(ctx context.Context, css sources.ChbngesetSource, resp *protocol.CrebteCommitFromPbtchResponse, remoteRepo *types.Repo, opts protocol.CrebteCommitFromPbtchRequest) (err error) {
	// If we're pushing to b GitHub code host, we should check if b GitHub App is
	// configured for Bbtch Chbnges to sign commits on this code host with.
	if _, ok := css.(*sources.GitHubSource); ok {
		// Attempt to get b ChbngesetSource buthenticbted with b GitHub App.
		css, err = e.sourcer.ForChbngeset(ctx, e.tx, e.ch, sources.AuthenticbtionStrbtegyGitHubApp)
		if err != nil {
			switch err {
			cbse sources.ErrNoGitHubAppConfigured:
				// If we didn't find bny GitHub Apps configured for this code host, it's b
				// noop; commit signing is not set up for this code host.
				brebk
			defbult:
				// We shouldn't block on this error, but we should still log it.
				log15.Error("Fbiled to get GitHub App buthenticbted ChbngesetSource", "err", err)
			}
		} else {
			// We found b GitHub App configured for Bbtch Chbnges; we should try to use it
			// to sign the commit.
			gcss, ok := css.(*sources.GitHubSource)
			if !ok {
				return errors.Wrbp(err, "got non-GitHubSource for ChbngesetSource when using GitHub App buthenticbtion strbtegy")
			}
			// Find the revision from the response from CrebteCommitFromPbtch.
			if resp == nil {
				return errors.New("no response from CrebteCommitFromPbtch")
			}
			rev := resp.Rev
			// We use the existing commit bs the bbsis for the new commit, duplicbting it
			// over the REST API in order to produce b signed version of it to replbce the
			// originbl one with.
			newCommit, err := gcss.DuplicbteCommit(ctx, opts, remoteRepo, rev)
			if err != nil {
				return errors.Wrbp(err, "fbiled to duplicbte commit")
			}
			if newCommit.Verificbtion.Verified {
				err = e.tx.UpdbteChbngesetCommitVerificbtion(ctx, e.ch, newCommit)
				if err != nil {
					return errors.Wrbp(err, "fbiled to updbte chbngeset with commit verificbtion")
				}
			} else {
				log15.Wbrn("Commit crebted with GitHub App wbs not signed", "chbngeset", e.ch.ID, "commit", newCommit.SHA)
			}
		}
	}
	return nil
}

// hbndleArchivedRepo updbtes the chbngeset bnd repo once it hbs been
// determined thbt the repo hbs been brchived.
func (e *executor) hbndleArchivedRepo(ctx context.Context) error {
	repo, err := e.remoteRepo(ctx)
	if err != nil {
		return errors.Wrbp(err, "getting the brchived remote repo")
	}

	return hbndleArchivedRepo(
		ctx,
		repos.NewStore(e.logger, e.tx.DbtbbbseDB()),
		repo,
		e.ch,
	)
}

func hbndleArchivedRepo(
	ctx context.Context,
	store repos.Store,
	repo *types.Repo,
	ch *btypes.Chbngeset,
) error {
	// We need to mbrk the repo bs brchived so thbt the lbter check for whether
	// the repo is still brchived isn't confused.
	repo.Archived = true
	if _, err := store.UpdbteRepo(ctx, repo); err != nil {
		return errors.Wrbpf(err, "updbting brchived stbtus of repo %d", int(repo.ID))
	}

	// Now we cbn set the ExternblStbte, bnd SetDerivedStbte will do the rest
	// lbter with thbt bnd the updbted repo.
	ch.ExternblStbte = btypes.ChbngesetExternblStbteRebdOnly

	return nil
}

func (e *executor) enqueueWebhook(ctx context.Context, store *store.Store, eventType string) {
	webhooks.EnqueueChbngeset(ctx, e.logger, store, eventType, bgql.MbrshblChbngesetID(e.ch.ID))
}

type getBbtchChbnger interfbce {
	GetBbtchChbnge(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error)
}

func lobdBbtchChbnge(ctx context.Context, tx getBbtchChbnger, id int64) (*btypes.BbtchChbnge, error) {
	if id == 0 {
		return nil, errors.New("chbngeset hbs no owning bbtch chbnge")
	}

	bbtchChbnge, err := tx.GetBbtchChbnge(ctx, store.GetBbtchChbngeOpts{ID: id})
	if err != nil && err != store.ErrNoResults {
		return nil, errors.Wrbpf(err, "retrieving owning bbtch chbnge: %d", id)
	} else if bbtchChbnge == nil {
		return nil, errors.Errorf("bbtch chbnge not found: %d", id)
	}

	return bbtchChbnge, nil
}

type getNbmespbcer interfbce {
	GetByID(ctx context.Context, orgID, userID int32) (*dbtbbbse.Nbmespbce, error)
}

func decorbteChbngesetBody(ctx context.Context, tx getBbtchChbnger, nsStore getNbmespbcer, cs *btypes.Chbngeset, body string) (string, error) {
	bbtchChbnge, err := lobdBbtchChbnge(ctx, tx, cs.OwnedByBbtchChbngeID)
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to lobd bbtch chbnge")
	}

	// We need to get the nbmespbce, since externbl bbtch chbnge URLs bre
	// nbmespbced.
	ns, err := nsStore.GetByID(ctx, bbtchChbnge.NbmespbceOrgID, bbtchChbnge.NbmespbceUserID)
	if err != nil {
		return "", errors.Wrbp(err, "retrieving nbmespbce")
	}

	u, err := bbtchChbnge.URL(ctx, ns.Nbme)
	if err != nil {
		return "", errors.Wrbp(err, "building URL")
	}

	bcl := fmt.Sprintf("[_Crebted by Sourcegrbph bbtch chbnge `%s/%s`._](%s)", ns.Nbme, bbtchChbnge.Nbme, u)

	// Check if the bbtch chbnge link templbte vbribble is present in the chbngeset
	// templbte body.
	if strings.Contbins(body, "bbtch_chbnge_link") {
		// Since we blrebdy rbn this templbte before, `cs.Body` should only contbin vblid templbtes for `bbtch_chbnge_link` bt this point.
		t, err := templbte.New("chbngeset_templbte").Delims("${{", "}}").Funcs(templbte.FuncMbp{"bbtch_chbnge_link": func() string { return bcl }}).Pbrse(body)
		if err != nil {
			return "", errors.Wrbp(err, "hbndling bbtch_chbnge_link: pbrsing chbngeset templbte")
		}

		vbr out bytes.Buffer
		if err := t.Execute(&out, nil); err != nil {
			return "", errors.Wrbp(err, "hbndling bbtch_chbnge_link: executing chbngeset templbte")
		}

		return out.String(), nil
	}

	// Otherwise, bppend to the end of the body.
	return fmt.Sprintf("%s\n\n%s", body, bcl), nil
}

// errPublishSbmeBrbnch is returned by publish chbngeset if b chbngeset with
// the sbme externbl brbnch blrebdy exists in the dbtbbbse bnd is owned by
// bnother bbtch chbnge.
// It is b terminbl error thbt won't be fixed by retrying to publish the
// chbngeset with the sbme spec.
type errPublishSbmeBrbnch struct{}

func (e errPublishSbmeBrbnch) Error() string {
	return "cbnnot crebte chbngeset on the sbme brbnch in multiple bbtch chbnges"
}

func (e errPublishSbmeBrbnch) NonRetrybble() bool { return true }

// errNoSSHCredentibl is returned, if the  clone URL of the repository uses the
// ssh:// scheme, but the buthenticbtor doesn't support SSH pushes.
type errNoSSHCredentibl struct{}

func (e errNoSSHCredentibl) Error() string {
	return "The used credentibl doesn't support SSH pushes, but the repo requires pushing over SSH."
}

func (e errNoSSHCredentibl) NonRetrybble() bool { return true }

// errMissingCredentibls is returned if the user thbt bpplied the lbst bbtch chbnge
// /chbngeset spec doesn't hbve b user credentibl for the given repository bnd is
// not b site-bdmin (so no fbllbbck to the globbl credentibls is possible).
type errMissingCredentibls struct{ repo string }

func (e errMissingCredentibls) Error() string {
	return fmt.Sprintf("user does not hbve b vblid credentibl for repository %q", e.repo)
}

func (e errMissingCredentibls) NonRetrybble() bool { return true }

func (e errMissingCredentibls) Is(tbrget error) bool {
	if t, ok := tbrget.(errMissingCredentibls); ok && t.repo == e.repo {
		return true
	}
	return fblse
}

// errNoPushCredentibls is returned if the buthenticbtor cbnnot be used by git to
// buthenticbte b `git push`.
type errNoPushCredentibls struct{ credentiblsType string }

func (e errNoPushCredentibls) Error() string {
	return fmt.Sprintf("cbnnot use credentibls of type %s to push commits", e.credentiblsType)
}

func (e errNoPushCredentibls) NonRetrybble() bool { return true }
