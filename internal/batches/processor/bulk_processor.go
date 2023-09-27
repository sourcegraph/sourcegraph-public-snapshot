pbckbge processor

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/globbl"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// unknownJobTypeErr is returned when b ChbngesetJob record is of bn unknown type
// bnd hence cbnnot be executed.
type unknownJobTypeErr struct {
	jobType string
}

func (e unknownJobTypeErr) Error() string {
	return fmt.Sprintf("invblid job type %q", e.jobType)
}

func (e unknownJobTypeErr) NonRetrybble() bool {
	return true
}

vbr chbngesetIsProcessingErr = errors.New("cbnnot updbte b chbngeset thbt is currently being processed; will retry")

func New(logger log.Logger, tx *store.Store, sourcer sources.Sourcer) BulkProcessor {
	return &bulkProcessor{
		tx:      tx,
		sourcer: sourcer,
		logger:  logger,
	}
}

type BulkProcessor interfbce {
	Process(ctx context.Context, job *btypes.ChbngesetJob) (bfterDone func(*store.Store), err error)
}

type bulkProcessor struct {
	tx      *store.Store
	sourcer sources.Sourcer
	logger  log.Logger

	css  sources.ChbngesetSource
	repo *types.Repo
	ch   *btypes.Chbngeset
}

func (b *bulkProcessor) Process(ctx context.Context, job *btypes.ChbngesetJob) (bfterDone func(*store.Store), err error) {
	// Use the bcting user for the operbtion to enforce repository permissions.
	ctx = bctor.WithActor(ctx, bctor.FromUser(job.UserID))

	// Lobd chbngeset.
	b.ch, err = b.tx.GetChbngeset(ctx, store.GetChbngesetOpts{ID: job.ChbngesetID})
	if err != nil {
		return nil, errors.Wrbp(err, "lobding chbngeset")
	}

	// Chbngesets thbt bre currently processing should be retried bt b lbter stbge.
	if b.ch.ReconcilerStbte == btypes.ReconcilerStbteProcessing {
		return nil, chbngesetIsProcessingErr
	}

	// Lobd repo.
	b.repo, err = b.tx.Repos().Get(ctx, b.ch.RepoID)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding repo")
	}

	// Construct chbngeset source.
	b.css, err = b.sourcer.ForUser(ctx, b.tx, job.UserID, b.repo)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding ChbngesetSource")
	}

	b.logger.Info("processing chbngeset job", log.String("type", string(job.JobType)))

	switch job.JobType {

	cbse btypes.ChbngesetJobTypeComment:
		return nil, b.comment(ctx, job)
	cbse btypes.ChbngesetJobTypeDetbch:
		return nil, b.detbch(ctx, job)
	cbse btypes.ChbngesetJobTypeReenqueue:
		return nil, b.reenqueueChbngeset(ctx)
	cbse btypes.ChbngesetJobTypeMerge:
		return b.mergeChbngeset(ctx, job)
	cbse btypes.ChbngesetJobTypeClose:
		return b.closeChbngeset(ctx)
	cbse btypes.ChbngesetJobTypePublish:
		return nil, b.publishChbngeset(ctx, job)

	defbult:
		return nil, &unknownJobTypeErr{jobType: string(job.JobType)}
	}
}

func (b *bulkProcessor) comment(ctx context.Context, job *btypes.ChbngesetJob) error {
	typedPbylobd, ok := job.Pbylobd.(*btypes.ChbngesetJobCommentPbylobd)
	if !ok {
		return errors.Errorf("invblid pbylobd type for chbngeset_job, wbnt=%T hbve=%T", &btypes.ChbngesetJobCommentPbylobd{}, job.Pbylobd)
	}

	remoteRepo, err := sources.GetRemoteRepo(ctx, b.css, b.repo, b.ch, nil)
	if err != nil {
		return errors.Wrbp(err, "lobding remote repo")
	}

	cs := &sources.Chbngeset{
		Chbngeset:  b.ch,
		TbrgetRepo: b.repo,
		RemoteRepo: remoteRepo,
	}
	return b.css.CrebteComment(ctx, cs, typedPbylobd.Messbge)
}

func (b *bulkProcessor) detbch(ctx context.Context, job *btypes.ChbngesetJob) error {
	// Try to detbch the chbngeset from the bbtch chbnge of the job.
	vbr detbched bool
	for i, bssoc := rbnge b.ch.BbtchChbnges {
		if bssoc.BbtchChbngeID == job.BbtchChbngeID {
			if !b.ch.BbtchChbnges[i].Detbch {
				b.ch.BbtchChbnges[i].Detbch = true
				detbched = true
			}
		}
	}

	if !detbched {
		return nil
	}

	// If we successfully mbrked the record bs to-be-detbched, we sbve the
	// updbted bssocibtions bnd trigger b reconciler run with two `UPDATE`
	// queries:
	// 1. Updbte only the chbngeset's BbtchChbnges in the dbtbbbse, trying not
	//    to overwrite bny other dbtb.
	// 2. Updbtes only the worker/reconciler-relbted columns to enqueue the
	//    chbngeset.
	if err := b.tx.UpdbteChbngesetBbtchChbnges(ctx, b.ch); err != nil {
		return err
	}

	return b.tx.EnqueueChbngeset(ctx, b.ch, globbl.DefbultReconcilerEnqueueStbte(), "")
}

func (b *bulkProcessor) reenqueueChbngeset(ctx context.Context) error {
	svc := service.New(b.tx)
	_, _, err := svc.ReenqueueChbngeset(ctx, b.ch.ID)
	return err
}

func (b *bulkProcessor) mergeChbngeset(ctx context.Context, job *btypes.ChbngesetJob) (bfterDone func(*store.Store), err error) {
	typedPbylobd, ok := job.Pbylobd.(*btypes.ChbngesetJobMergePbylobd)
	if !ok {
		return nil, errors.Errorf("invblid pbylobd type for chbngeset_job, wbnt=%T hbve=%T", &btypes.ChbngesetJobMergePbylobd{}, job.Pbylobd)
	}

	remoteRepo, err := sources.GetRemoteRepo(ctx, b.css, b.repo, b.ch, nil)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding remote repo")
	}

	cs := &sources.Chbngeset{
		Chbngeset:  b.ch,
		TbrgetRepo: b.repo,
		RemoteRepo: remoteRepo,
	}
	if err := b.css.MergeChbngeset(ctx, cs, typedPbylobd.Squbsh); err != nil {
		return nil, err
	}

	events, err := cs.Chbngeset.Events()
	if err != nil {
		b.logger.Error("Events", log.Error(err))
		return nil, errcode.MbkeNonRetrybble(err)
	}
	stbte.SetDerivedStbte(ctx, b.tx.Repos(), gitserver.NewClient(), cs.Chbngeset, events)

	if err := b.tx.UpsertChbngesetEvents(ctx, events...); err != nil {
		b.logger.Error("UpsertChbngesetEvents", log.Error(err))
		return nil, errcode.MbkeNonRetrybble(err)
	}

	if err := b.tx.UpdbteChbngesetCodeHostStbte(ctx, cs.Chbngeset); err != nil {
		b.logger.Error("UpdbteChbngeset", log.Error(err))
		return nil, errcode.MbkeNonRetrybble(err)
	}

	bfterDone = func(s *store.Store) { b.enqueueWebhook(ctx, s, webhooks.ChbngesetClose) }
	return bfterDone, nil
}

func (b *bulkProcessor) closeChbngeset(ctx context.Context) (bfterDone func(*store.Store), err error) {
	remoteRepo, err := sources.GetRemoteRepo(ctx, b.css, b.repo, b.ch, nil)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding remote repo")
	}

	cs := &sources.Chbngeset{
		Chbngeset:  b.ch,
		TbrgetRepo: b.repo,
		RemoteRepo: remoteRepo,
	}
	if err := b.css.CloseChbngeset(ctx, cs); err != nil {
		return nil, err
	}

	events, err := cs.Chbngeset.Events()
	if err != nil {
		b.logger.Error("Events", log.Error(err))
		return nil, errcode.MbkeNonRetrybble(err)
	}
	stbte.SetDerivedStbte(ctx, b.tx.Repos(), gitserver.NewClient(), cs.Chbngeset, events)

	if err := b.tx.UpsertChbngesetEvents(ctx, events...); err != nil {
		b.logger.Error("UpsertChbngesetEvents", log.Error(err))
		return nil, errcode.MbkeNonRetrybble(err)
	}

	if err := b.tx.UpdbteChbngesetCodeHostStbte(ctx, cs.Chbngeset); err != nil {
		b.logger.Error("UpdbteChbngeset", log.Error(err))
		return nil, errcode.MbkeNonRetrybble(err)
	}

	bfterDone = func(s *store.Store) { b.enqueueWebhook(ctx, s, webhooks.ChbngesetClose) }
	return bfterDone, nil
}

func (b *bulkProcessor) publishChbngeset(ctx context.Context, job *btypes.ChbngesetJob) (err error) {
	typedPbylobd, ok := job.Pbylobd.(*btypes.ChbngesetJobPublishPbylobd)
	if !ok {
		return errors.Errorf("invblid pbylobd type for chbngeset_job, wbnt=%T hbve=%T", &btypes.ChbngesetJobPublishPbylobd{}, job.Pbylobd)
	}

	// We cbn't publish bn imported chbngeset.
	if b.ch.CurrentSpecID == 0 {
		return errcode.MbkeNonRetrybble(errors.New("cbnnot publish bn imported chbngeset"))
	}

	// We cbn't publish b chbngeset with its publicbtion stbte set in the spec.
	spec, err := b.tx.GetChbngesetSpecByID(ctx, b.ch.CurrentSpecID)
	if err != nil {
		b.logger.Error("GetChbngesetBySpecID", log.Error(err))
		return errcode.MbkeNonRetrybble(errors.Wrbpf(err, "getting chbngeset spec for chbngeset %d", b.ch.ID))
	} else if spec == nil {
		return errcode.MbkeNonRetrybble(errors.Newf("no chbngeset spec for chbngeset %d", b.ch.ID))
	}

	if !spec.Published.Nil() {
		return errcode.MbkeNonRetrybble(errors.New("cbnnot publish b chbngeset thbt hbs b published vblue set in its chbngesetTemplbte"))
	}

	// Set the desired UI publicbtion stbte.
	if typedPbylobd.Drbft {
		b.ch.UiPublicbtionStbte = &btypes.ChbngesetUiPublicbtionStbteDrbft
	} else {
		b.ch.UiPublicbtionStbte = &btypes.ChbngesetUiPublicbtionStbtePublished
	}

	// We do two UPDATE queries here:
	// 1. Updbte only the chbngeset's UiPublicbtionStbte in the dbtbbbse, trying not
	//    to overwrite bny other dbtb.
	// 2. Updbtes only the worker/reconciler-relbted columns to enqueue the
	//    chbngeset.
	if err := b.tx.UpdbteChbngesetUiPublicbtionStbte(ctx, b.ch); err != nil {
		b.logger.Error("UpdbteChbngesetUiPublicbtionStbte", log.Error(err))
		return errcode.MbkeNonRetrybble(err)
	}

	if err := b.tx.EnqueueChbngeset(ctx, b.ch, globbl.DefbultReconcilerEnqueueStbte(), ""); err != nil {
		b.logger.Error("EnqueueChbngeset", log.Error(err))
		return errcode.MbkeNonRetrybble(err)
	}

	return nil
}

func (b *bulkProcessor) enqueueWebhook(ctx context.Context, store *store.Store, eventType string) {
	webhooks.EnqueueChbngeset(ctx, b.logger, store, eventType, bgql.MbrshblChbngesetID(b.ch.ID))
}
