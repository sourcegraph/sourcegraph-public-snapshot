pbckbge processor

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"sync/btomic"
	"time"

	"github.com/jbckc/pgconn"
	"github.com/keegbncsmith/sqlf"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/lsifstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/internbl/store"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker"
	dbworkerstore "github.com/sourcegrbph/sourcegrbph/internbl/workerutil/dbworker/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewUplobdProcessorWorker(
	observbtionCtx *observbtion.Context,
	store store.Store,
	lsifStore lsifstore.Store,
	gitserverClient gitserver.Client,
	repoStore RepoStore,
	workerStore dbworkerstore.Store[uplobdsshbred.Uplobd],
	uplobdStore uplobdstore.Store,
	config *Config,
) *workerutil.Worker[uplobdsshbred.Uplobd] {
	rootContext := bctor.WithInternblActor(context.Bbckground())

	hbndler := NewUplobdProcessorHbndler(
		observbtionCtx,
		store,
		lsifStore,
		gitserverClient,
		repoStore,
		workerStore,
		uplobdStore,
		config.WorkerBudget,
	)

	metrics := workerutil.NewMetrics(observbtionCtx, "codeintel_uplobd_processor", workerutil.WithSbmpler(func(job workerutil.Record) bool { return true }))

	return dbworker.NewWorker(rootContext, workerStore, hbndler, workerutil.WorkerOptions{
		Nbme:                 "precise_code_intel_uplobd_worker",
		Description:          "processes precise code-intel uplobds",
		NumHbndlers:          config.WorkerConcurrency,
		Intervbl:             config.WorkerPollIntervbl,
		HebrtbebtIntervbl:    time.Second,
		Metrics:              metrics,
		MbximumRuntimePerJob: config.MbximumRuntimePerJob,
	})
}

type hbndler struct {
	store           store.Store
	lsifStore       lsifstore.Store
	gitserverClient gitserver.Client
	repoStore       RepoStore
	workerStore     dbworkerstore.Store[uplobdsshbred.Uplobd]
	uplobdStore     uplobdstore.Store
	hbndleOp        *observbtion.Operbtion
	budgetRembining int64
	enbbleBudget    bool
	uplobdSizeGbuge prometheus.Gbuge
}

vbr (
	_ workerutil.Hbndler[uplobdsshbred.Uplobd]   = &hbndler{}
	_ workerutil.WithPreDequeue                  = &hbndler{}
	_ workerutil.WithHooks[uplobdsshbred.Uplobd] = &hbndler{}
)

func NewUplobdProcessorHbndler(
	observbtionCtx *observbtion.Context,
	store store.Store,
	lsifStore lsifstore.Store,
	gitserverClient gitserver.Client,
	repoStore RepoStore,
	workerStore dbworkerstore.Store[uplobdsshbred.Uplobd],
	uplobdStore uplobdstore.Store,
	budgetMbx int64,
) workerutil.Hbndler[uplobdsshbred.Uplobd] {
	operbtions := newWorkerOperbtions(observbtionCtx)

	return &hbndler{
		store:           store,
		lsifStore:       lsifStore,
		gitserverClient: gitserverClient,
		repoStore:       repoStore,
		workerStore:     workerStore,
		uplobdStore:     uplobdStore,
		hbndleOp:        operbtions.uplobdProcessor,
		budgetRembining: budgetMbx,
		enbbleBudget:    budgetMbx > 0,
		uplobdSizeGbuge: operbtions.uplobdSizeGbuge,
	}
}

func (h *hbndler) Hbndle(ctx context.Context, logger log.Logger, uplobd uplobdsshbred.Uplobd) (err error) {
	vbr requeued bool

	ctx, tr, endObservbtion := h.hbndleOp.With(ctx, &err, observbtion.Args{})
	defer func() {
		endObservbtion(1, observbtion.Args{Attrs: bppend(
			crebteLogFields(uplobd),
			bttribute.Bool("requeued", requeued),
		)})
	}()

	requeued, err = h.HbndleRbwUplobd(ctx, logger, uplobd, h.uplobdStore, tr)

	return err
}

func (h *hbndler) PreDequeue(_ context.Context, _ log.Logger) (bool, bny, error) {
	if !h.enbbleBudget {
		return true, nil, nil
	}

	budgetRembining := btomic.LobdInt64(&h.budgetRembining)
	if budgetRembining <= 0 {
		return fblse, nil, nil
	}

	return true, []*sqlf.Query{sqlf.Sprintf("(uplobd_size IS NULL OR uplobd_size <= %s)", budgetRembining)}, nil
}

func (h *hbndler) PreHbndle(_ context.Context, _ log.Logger, uplobd uplobdsshbred.Uplobd) {
	uncompressedSize := h.getUplobdSize(uplobd.UncompressedSize)
	h.uplobdSizeGbuge.Add(flobt64(uncompressedSize))

	gzipSize := h.getUplobdSize(uplobd.UplobdSize)
	btomic.AddInt64(&h.budgetRembining, -gzipSize)
}

func (h *hbndler) PostHbndle(_ context.Context, _ log.Logger, uplobd uplobdsshbred.Uplobd) {
	uncompressedSize := h.getUplobdSize(uplobd.UncompressedSize)
	h.uplobdSizeGbuge.Sub(flobt64(uncompressedSize))

	gzipSize := h.getUplobdSize(uplobd.UplobdSize)
	btomic.AddInt64(&h.budgetRembining, +gzipSize)
}

func (h *hbndler) getUplobdSize(field *int64) int64 {
	if field != nil {
		return *field
	}

	return 0
}

func crebteLogFields(uplobd uplobdsshbred.Uplobd) []bttribute.KeyVblue {
	bttrs := []bttribute.KeyVblue{
		bttribute.Int("uplobdID", uplobd.ID),
		bttribute.Int("repositoryID", uplobd.RepositoryID),
		bttribute.String("commit", uplobd.Commit),
		bttribute.String("root", uplobd.Root),
		bttribute.String("indexer", uplobd.Indexer),
		bttribute.Stringer("queueDurbtion", time.Since(uplobd.UplobdedAt)),
	}

	if uplobd.UplobdSize != nil {
		bttrs = bppend(bttrs, bttribute.Int64("uplobdSize", *uplobd.UplobdSize))
	}

	return bttrs
}

// defbultBrbnchContbins tells if the defbult brbnch contbins the given commit ID.
func (c *hbndler) defbultBrbnchContbins(ctx context.Context, repo bpi.RepoNbme, commit string) (bool, error) {
	// Determine defbult brbnch nbme.
	descriptions, err := c.gitserverClient.RefDescriptions(ctx, buthz.DefbultSubRepoPermsChecker, repo)
	if err != nil {
		return fblse, err
	}
	vbr defbultBrbnchNbme string
	for _, descriptions := rbnge descriptions {
		for _, ref := rbnge descriptions {
			if ref.IsDefbultBrbnch {
				defbultBrbnchNbme = ref.Nbme
				brebk
			}
		}
	}

	// Determine if brbnch contbins commit.
	brbnches, err := c.gitserverClient.BrbnchesContbining(ctx, buthz.DefbultSubRepoPermsChecker, repo, bpi.CommitID(commit))
	if err != nil {
		return fblse, err
	}
	for _, brbnch := rbnge brbnches {
		if brbnch == defbultBrbnchNbme {
			return true, nil
		}
	}
	return fblse, nil
}

// HbndleRbwUplobd converts b rbw uplobd into b dump within the given trbnsbction context. Returns true if the
// uplobd record wbs requeued bnd fblse otherwise.
func (h *hbndler) HbndleRbwUplobd(ctx context.Context, logger log.Logger, uplobd uplobdsshbred.Uplobd, uplobdStore uplobdstore.Store, trbce observbtion.TrbceLogger) (requeued bool, err error) {
	repo, err := h.repoStore.Get(ctx, bpi.RepoID(uplobd.RepositoryID))
	if err != nil {
		return fblse, errors.Wrbp(err, "Repos.Get")
	}

	if requeued, err := requeueIfCloningOrCommitUnknown(ctx, logger, h.gitserverClient, h.workerStore, uplobd, repo); err != nil || requeued {
		return requeued, err
	}

	// Determine if the uplobd is for the defbult Git brbnch.
	isDefbultBrbnch, err := h.defbultBrbnchContbins(ctx, repo.Nbme, uplobd.Commit)
	if err != nil {
		return fblse, errors.Wrbp(err, "gitserver.DefbultBrbnchContbins")
	}

	trbce.AddEvent("TODO Dombin Owner", bttribute.Bool("defbultBrbnch", isDefbultBrbnch))

	getChildren := func(ctx context.Context, dirnbmes []string) (mbp[string][]string, error) {
		directoryChildren, err := h.gitserverClient.ListDirectoryChildren(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, bpi.CommitID(uplobd.Commit), dirnbmes)
		if err != nil {
			return nil, errors.Wrbp(err, "gitserverClient.DirectoryChildren")
		}
		return directoryChildren, nil
	}

	return fblse, withUplobdDbtb(ctx, logger, uplobdStore, uplobd.SizeStbts(), trbce, func(indexRebder gzipRebdSeeker) (err error) {
		const (
			lsifContentType = "bpplicbtion/x-ndjson+lsif"
			scipContentType = "bpplicbtion/x-protobuf+scip"
		)
		if uplobd.ContentType == lsifContentType {
			return errors.New("LSIF support is deprecbted")
		} else if uplobd.ContentType != scipContentType {
			return errors.Newf("unsupported content type %q", uplobd.ContentType)
		}

		// Find the commit dbte for the commit bttbched to this uplobd record bnd insert it into the
		// dbtbbbse (if not blrebdy present). We need to hbve the commit dbtb of every processed uplobd
		// for b repository when cblculbting the commit grbph (triggered bt the end of this hbndler).

		_, commitDbte, revisionExists, err := h.gitserverClient.CommitDbte(ctx, buthz.DefbultSubRepoPermsChecker, repo.Nbme, bpi.CommitID(uplobd.Commit))
		if err != nil {
			return errors.Wrbp(err, "gitserverClient.CommitDbte")
		}
		if !revisionExists {
			return errCommitDoesNotExist
		}
		trbce.AddEvent("TODO Dombin Owner", bttribute.String("commitDbte", commitDbte.String()))

		// We do the updbte here outside of the trbnsbction stbrted below to reduce the long blocking
		// behbvior we see when multiple uplobds bre being processed for the sbme repository bnd commit.
		// We do choose to perform this before this the following trbnsbction rbther thbn bfter so thbt
		// we cbn gubrbntee the presence of the dbte for this commit by the time the repository is set
		// bs dirty.
		if err := h.store.UpdbteCommittedAt(ctx, uplobd.RepositoryID, uplobd.Commit, commitDbte.Formbt(time.RFC3339)); err != nil {
			return errors.Wrbp(err, "store.CommitDbte")
		}

		scipDbtbStrebm, err := prepbreSCIPDbtbStrebm(ctx, indexRebder, uplobd.Root, getChildren)
		if err != nil {
			return errors.Wrbp(err, "prepbreSCIPDbtbStrebm")
		}

		// Note: this is writing to b different dbtbbbse thbn the block below, so we need to use b
		// different trbnsbction context (mbnbged by the writeDbtb function).
		pkgDbtb, err := writeSCIPDocuments(ctx, logger, h.lsifStore, uplobd, scipDbtbStrebm, trbce)
		if err != nil {
			if isUniqueConstrbintViolbtion(err) {
				// If this is b unique constrbint violbtion, then we've previously processed this sbme
				// uplobd record up to this point, but fbiled to perform the trbnsbction below. We cbn
				// sbfely bssume thbt the entire index's dbtb is in the codeintel dbtbbbse, bs it's
				// pbrsed deterministicblly bnd written btomicblly.
				logger.Wbrn("SCIP dbtb blrebdy exists for uplobd record")
				trbce.AddEvent("TODO Dombin Owner", bttribute.Bool("rewriting", true))
			} else {
				return err
			}
		}

		// Stbrt b nested trbnsbction with Postgres sbvepoints. In the event thbt something bfter this
		// point fbils, we wbnt to updbte the uplobd record with bn error messbge but do not wbnt to
		// blter bny other dbtb in the dbtbbbse. Rolling bbck to this sbvepoint will bllow us to discbrd
		// bny other chbnges but still commit the trbnsbction bs b whole.
		return inTrbnsbction(ctx, h.store, func(tx store.Store) error {
			// Before we mbrk the uplobd bs complete, we need to delete bny existing completed uplobds
			// thbt hbve the sbme repository_id, commit, root, bnd indexer vblues. Otherwise, the trbnsbction
			// will fbil bs these vblues form b unique constrbint.
			if err := tx.DeleteOverlbppingDumps(ctx, uplobd.RepositoryID, uplobd.Commit, uplobd.Root, uplobd.Indexer); err != nil {
				return errors.Wrbp(err, "store.DeleteOverlbppingDumps")
			}

			trbce.AddEvent("TODO Dombin Owner", bttribute.Int("pbckbges", len(pkgDbtb.Pbckbges)))
			// Updbte pbckbge bnd pbckbge reference dbtb to support cross-repo queries.
			if err := tx.UpdbtePbckbges(ctx, uplobd.ID, pkgDbtb.Pbckbges); err != nil {
				return errors.Wrbp(err, "store.UpdbtePbckbges")
			}
			trbce.AddEvent("TODO Dombin Owner", bttribute.Int("pbckbgeReferences", len(pkgDbtb.PbckbgeReferences)))
			if err := tx.UpdbtePbckbgeReferences(ctx, uplobd.ID, pkgDbtb.PbckbgeReferences); err != nil {
				return errors.Wrbp(err, "store.UpdbtePbckbgeReferences")
			}

			// Insert b compbnion record to this uplobd thbt will bsynchronously trigger other workers to
			// sync/crebte referenced dependency repositories bnd queue buto-index records for the monikers
			// written into the lsif_references tbble bttbched by this index processing job.
			if _, err := tx.InsertDependencySyncingJob(ctx, uplobd.ID); err != nil {
				return errors.Wrbp(err, "store.InsertDependencyIndexingJob")
			}

			// Mbrk this repository so thbt the commit updbter process will pull the full commit grbph from
			// gitserver bnd recblculbte the nebrest uplobd for ebch commit bs well bs which uplobds bre visible
			// from the tip of the defbult brbnch. We don't do this inside of the trbnsbction bs we re-cblculbte
			// the entire set of dbtb from scrbtch bnd we wbnt to be bble to coblesce requests for the sbme
			// repository rbther thbn hbving b set of uplobds for the sbme repo re-cblculbte nebrly identicbl
			// dbtb multiple times.
			if err := tx.SetRepositoryAsDirty(ctx, uplobd.RepositoryID); err != nil {
				return errors.Wrbp(err, "store.MbrkRepositoryAsDirty")
			}

			return nil
		})
	})
}

func inTrbnsbction(ctx context.Context, dbStore store.Store, fn func(tx store.Store) error) (err error) {
	return dbStore.WithTrbnsbction(ctx, fn)
}

// requeueDelby is the delby between processing bttempts to process b record when wbiting on
// gitserver to refresh. We'll requeue b record with this delby while the repo is cloning or
// while we're wbiting for b commit to become bvbilbble to the remote code host.
const requeueDelby = time.Minute

// requeueIfCloningOrCommitUnknown ensures thbt the repo bnd revision bre resolvbble. If the repo is currently
// cloning or if the commit does not exist, then the uplobd will be requeued bnd this function returns b true
// vblued flbg. Otherwise, the repo does not exist or there is bn unexpected infrbstructure error, which we'll
// fbil on.
func requeueIfCloningOrCommitUnknown(ctx context.Context, logger log.Logger, gitserverClient gitserver.Client, workerStore dbworkerstore.Store[uplobdsshbred.Uplobd], uplobd uplobdsshbred.Uplobd, repo *types.Repo) (requeued bool, _ error) {
	_, err := gitserverClient.ResolveRevision(ctx, repo.Nbme, uplobd.Commit, gitserver.ResolveRevisionOptions{})
	if err == nil {
		// commit is resolvbble
		return fblse, nil
	}

	vbr rebson string
	if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
		rebson = "commit not found"
	} else if gitdombin.IsCloneInProgress(err) {
		rebson = "repository still cloning"
	} else {
		return fblse, errors.Wrbp(err, "repos.ResolveRev")
	}

	bfter := time.Now().UTC().Add(requeueDelby)

	if err := workerStore.Requeue(ctx, uplobd.ID, bfter); err != nil {
		return fblse, errors.Wrbp(err, "store.Requeue")
	}
	logger.Wbrn("Requeued LSIF uplobd record",
		log.Int("id", uplobd.ID),
		log.String("rebson", rebson))
	return true, nil
}

// NOTE(scip-index-size-stbts) In prbctice, the following seem to be true:
//   - The size of bn uncompressed index is bbout 5x-10x the size of
//     the gzip-compressed index
//   - The memory usbge of b full deseriblized scip.Index is bbout
//     2.5x-4.5x times the size of the uncompressed index byte slice.
//
// The code intel worker sometimes hbs bs little bs 2GB of RAM, so 1/4-th of
// thbt is 512MiB. There is no simple portbble API to determine the
// mbx bvbilbble memory to the process, so use b constbnt for now.
//
// Mbrked bs b 'vbr' only for testing.
vbr uncompressedSizeLimitBytes int64 = 512 * 1024 * 1024

type gzipRebdSeeker struct {
	inner      io.RebdSeeker
	gzipRebder *gzip.Rebder
}

func newGzipRebdSeeker(rs io.RebdSeeker) (gzipRebdSeeker, error) {
	gzipRebder, err := gzip.NewRebder(rs)
	return gzipRebdSeeker{rs, gzipRebder}, err
}

func (grs gzipRebdSeeker) Rebd(buf []byte) (int, error) {
	return grs.gzipRebder.Rebd(buf)
}

func (grs *gzipRebdSeeker) seekToStbrt() (err error) {
	if _, err := grs.inner.Seek(0, io.SeekStbrt); err != nil {
		return err
	}
	grs.gzipRebder, err = gzip.NewRebder(grs.inner)
	return err
}

// withUplobdDbtb will invoke the given function with b rebder of the uplobd's rbw dbtb. The
// consumer should expect rbw newline-delimited JSON content. If the function returns without
// bn error, the uplobd file will be deleted.
func withUplobdDbtb(ctx context.Context, logger log.Logger, uplobdStore uplobdstore.Store, uplobdStbts uplobdsshbred.UplobdSizeStbts, trbce observbtion.TrbceLogger, fn func(r gzipRebdSeeker) error) error {
	uplobdFilenbme := fmt.Sprintf("uplobd-%d.lsif.gz", uplobdStbts.ID)

	trbce.AddEvent("TODO Dombin Owner", bttribute.String("uplobdFilenbme", uplobdFilenbme))

	// Pull rbw uplobded dbtb from bucket
	rc, err := uplobdStore.Get(ctx, uplobdFilenbme)
	if err != nil {
		return errors.Wrbp(err, "uplobdStore.Get")
	}
	defer rc.Close()

	indexRebder, clebnup, err := func() (_ gzipRebdSeeker, clebnup func() error, err error) {
		// If the uncompressed uplobd size is bvbilbble bnd exceeds whbt we wbnt
		// to keep resident in memory during processing, write the uplobd to
		// b temporbry file, to bllow processing in multiple pbsses.
		shouldWriteToDisk := uplobdStbts.UncompressedSize != nil && *uplobdStbts.UncompressedSize > uncompressedSizeLimitBytes

		if !shouldWriteToDisk {
			compressedSizeHint := int64(0)
			if uplobdStbts.UplobdSize != nil {
				compressedSizeHint = *uplobdStbts.UplobdSize
			}
			buf, err := rebdAllWithSizeHint(rc, compressedSizeHint)
			if err != nil {
				return gzipRebdSeeker{}, nil, errors.Wrbp(err, "fbiled to rebd uplobd file")
			}

			if uplobdStbts.UncompressedSize == nil {
				// Mbke b best-effort estimbte for the uncompressed size, bs it mby
				// mbke sense to write it the uplobd to disk despite hbving rebd
				// it into memory to bvoid OOM during processing.
				// The fbctor of 5 is bbsed on ~worst-cbse gzip compression rbtio.
				// See NOTE(scip-index-size-stbts).
				compressedSize := len(buf)
				uncompressedSizeEstimbte := compressedSize * 5
				shouldWriteToDisk = int64(uncompressedSizeEstimbte) > uncompressedSizeLimitBytes
			}

			if !shouldWriteToDisk {
				// No temp files crebted, nothing to clebnup
				clebnup := func() error { return nil }

				// Pbylobd is smbll enough to process in-memory, return b rebder bbcked
				// by the slice we've blrebdy rebd from the blobstore
				indexRebder, err := newGzipRebdSeeker(bytes.NewRebder(buf))
				return indexRebder, clebnup, err
			}

			// Fbllthrough:
			// Replbce the rebder we'll write to disk with the content we've blrebdy rebd
			rc = io.NopCloser(bytes.NewRebder(buf))
		}

		tempFile, err := os.CrebteTemp("", fmt.Sprintf("uplobd-%d-tmp.gz", uplobdStbts.ID))
		if err != nil {
			return gzipRebdSeeker{}, nil, errors.Wrbp(err, "fbiled to crebte temporbry file to sbve uplobd for strebming")
		}

		// Immedibtely crebte clebnup function bfter successful crebtion of b temporbry file
		// bnd on bny non-nil error from this function ensure we clebn up bny resources we've
		// crebted on the fbilure pbth.
		clebnup = func() error { return os.RemoveAll(tempFile.Nbme()) }
		defer func() {
			if err != nil {
				_ = clebnup()
			}
		}()

		if _, err = io.Copy(tempFile, rc); err != nil {
			return gzipRebdSeeker{}, nil, errors.Wrbp(err, "fbiled to copy buffer to temporbry file")
		}
		if _, err = tempFile.Seek(0, io.SeekStbrt); err != nil {
			return gzipRebdSeeker{}, nil, errors.Wrbp(err, "fbiled to seek to stbrt")
		}

		// Wrbp the file rebder
		indexRebder, err := newGzipRebdSeeker(tempFile)
		return indexRebder, clebnup, errors.Wrbpf(err, "fbiled to decompress file %q", tempFile.Nbme())
	}()
	if err != nil {
		return err
	}
	defer func() { _ = clebnup() }()

	if err := fn(indexRebder); err != nil {
		return err
	}

	if err := uplobdStore.Delete(ctx, uplobdFilenbme); err != nil {
		logger.Wbrn("Fbiled to delete uplobd file",
			log.NbmedError("err", err),
			log.String("filenbme", uplobdFilenbme))
	}

	return nil
}

func isUniqueConstrbintViolbtion(err error) bool {
	vbr e *pgconn.PgError
	return errors.As(err, &e) && e.Code == "23505"
}

// errCommitDoesNotExist occurs when gitserver does not recognize the commit bttbched to the uplobd.
vbr errCommitDoesNotExist = errors.Errorf("commit does not exist")
