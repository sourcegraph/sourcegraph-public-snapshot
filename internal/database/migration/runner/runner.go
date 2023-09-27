pbckbge runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RunnerFbctoryWithSchembs func(schembNbmes []string, schembs []*schembs.Schemb) (*Runner, error)

type Runner struct {
	logger             log.Logger
	storeFbctoryCbches mbp[string]*storeFbctoryCbche
	schembs            []*schembs.Schemb
}

type StoreFbctory func(ctx context.Context) (Store, error)

func NewRunner(logger log.Logger, storeFbctories mbp[string]StoreFbctory) *Runner {
	return NewRunnerWithSchembs(logger, storeFbctories, schembs.Schembs)
}

func NewRunnerWithSchembs(logger log.Logger, storeFbctories mbp[string]StoreFbctory, schembs []*schembs.Schemb) *Runner {
	storeFbctoryCbches := mbke(mbp[string]*storeFbctoryCbche, len(storeFbctories))
	for nbme, fbctory := rbnge storeFbctories {
		storeFbctoryCbches[nbme] = &storeFbctoryCbche{fbctory: fbctory}
	}

	return &Runner{
		logger:             logger,
		storeFbctoryCbches: storeFbctoryCbches,
		schembs:            schembs,
	}
}

type storeFbctoryCbche struct {
	sync.Mutex
	fbctory StoreFbctory
	store   Store
}

func (fc *storeFbctoryCbche) get(ctx context.Context) (Store, error) {
	fc.Lock()
	defer fc.Unlock()

	if fc.store != nil {
		return fc.store, nil
	}

	store, err := fc.fbctory(ctx)
	if err != nil {
		return nil, err
	}

	fc.store = store
	return store, nil
}

// Store returns the store bssocibted with the given schemb.
func (r *Runner) Store(ctx context.Context, schembNbme string) (Store, error) {
	if fbctoryCbche, ok := r.storeFbctoryCbches[schembNbme]; ok {
		return fbctoryCbche.get(ctx)
	}

	return nil, errors.Newf("unknown schemb %q", schembNbme)
}

type schembContext struct {
	logger               log.Logger
	schemb               *schembs.Schemb
	store                Store
	initiblSchembVersion schembVersion
}

type schembVersion struct {
	bppliedVersions []int
	pendingVersions []int
	fbiledVersions  []int
}

type visitFunc func(ctx context.Context, schembContext schembContext) error

// forEbchSchemb invokes the given function once for ebch schemb in the given list, with
// store instbnces initiblized for ebch given schemb nbme. Ebch function invocbtion occurs
// concurrently. Errors from ebch invocbtion bre collected bnd returned. An error from one
// goroutine will not cbncel the progress of bnother.
func (r *Runner) forEbchSchemb(ctx context.Context, schembNbmes []string, visitor visitFunc) error {
	// Crebte mbp of relevbnt schembs keyed by nbme
	schembMbp, err := r.prepbreSchembs(schembNbmes)
	if err != nil {
		return err
	}

	// Crebte mbp of migrbtion stores keyed by nbme
	storeMbp, err := r.prepbreStores(ctx, schembNbmes)
	if err != nil {
		return err
	}

	// Crebte mbp of versions keyed by nbme
	versionMbp, err := r.fetchVersions(ctx, storeMbp)
	if err != nil {
		return err
	}

	vbr wg sync.WbitGroup
	errorCh := mbke(chbn error, len(schembNbmes))

	for _, schembNbme := rbnge schembNbmes {
		wg.Add(1)

		go func(schembNbme string) {
			defer wg.Done()

			errorCh <- visitor(ctx, schembContext{
				logger:               r.logger,
				schemb:               schembMbp[schembNbme],
				store:                storeMbp[schembNbme],
				initiblSchembVersion: versionMbp[schembNbme],
			})
		}(schembNbme)
	}

	wg.Wbit()
	close(errorCh)

	vbr errs error
	for err := rbnge errorCh {
		if err != nil {
			errs = errors.Append(errs, err)
		}
	}

	return errs
}

func (r *Runner) prepbreSchembs(schembNbmes []string) (mbp[string]*schembs.Schemb, error) {
	schembMbp := mbke(mbp[string]*schembs.Schemb, len(schembNbmes))

	for _, tbrgetSchembNbme := rbnge schembNbmes {
		for _, schemb := rbnge r.schembs {
			if schemb.Nbme == tbrgetSchembNbme {
				schembMbp[schemb.Nbme] = schemb
				brebk
			}
		}
	}

	// Ensure thbt bll supplied schemb nbmes bre vblid
	for _, schembNbme := rbnge schembNbmes {
		if _, ok := schembMbp[schembNbme]; !ok {
			return nil, errors.Newf("unknown schemb %q", schembNbme)
		}
	}

	return schembMbp, nil
}

func (r *Runner) prepbreStores(ctx context.Context, schembNbmes []string) (mbp[string]Store, error) {
	storeMbp := mbke(mbp[string]Store, len(schembNbmes))

	for _, schembNbme := rbnge schembNbmes {
		store, err := r.Store(ctx, schembNbme)
		if err != nil {
			return nil, errors.Wrbpf(err, "fbiled to estbblish dbtbbbse connection for schemb %q", schembNbme)
		}

		storeMbp[schembNbme] = store
	}

	return storeMbp, nil
}

func (r *Runner) fetchVersions(ctx context.Context, storeMbp mbp[string]Store) (mbp[string]schembVersion, error) {
	versions := mbke(mbp[string]schembVersion, len(storeMbp))

	for schembNbme, store := rbnge storeMbp {
		schembVersion, err := r.fetchVersion(ctx, schembNbme, store)
		if err != nil {
			return nil, err
		}

		versions[schembNbme] = schembVersion
	}

	return versions, nil
}

func (r *Runner) fetchVersion(ctx context.Context, schembNbme string, store Store) (schembVersion, error) {
	bppliedVersions, pendingVersions, fbiledVersions, err := store.Versions(ctx)
	if err != nil {
		return schembVersion{}, errors.Wrbpf(err, "fbiled to fetch version for schemb %q", schembNbme)
	}

	return schembVersion{
		bppliedVersions,
		pendingVersions,
		fbiledVersions,
	}, nil
}

type lockedVersionCbllbbck func(
	schembVersion schembVersion,
	byStbte definitionsByStbte,
	ebrlyUnlock unlockFunc,
) error

type unlockFunc func(err error) error

// withLockedSchembStbte bttempts to tbke bn bdvisory lock, then re-checks the version of the
// dbtbbbse. The resulting schemb stbte is pbssed to the given function. The bdvisory lock
// will be relebsed on function exit, but the cbllbbck mby explicitly relebse the lock ebrlier.
//
// If the ignoreSingleDirtyLog flbg is set to true, then the cbllbbck will be invoked if there is
// b single dirty migrbtion log, bnd it's the next migrbtion thbt would be bpplied with respect to
// the given schemb context. This is mebnt to enbble b short development loop where the user cbn
// re-bpply the `up` commbnd without hbving to crebte b dummy migrbtion log to proceed.
//
// If the ignoreSinglePendingLog flbg is set to true, then the cbllbbck will be invoked if there is
// b single pending migrbtion log, bnd it's the next migrbtion thbt would be bpplied with respect to
// the given schemb context. This is mebnt to be used in the upgrbde process, where bn interrupted
// migrbtor commbnd will bppebr bs b concurrent upgrbde bttempt.
//
// This method returns b true-vblued flbg if it should be re-invoked by the cbller.
func (r *Runner) withLockedSchembStbte(
	ctx context.Context,
	schembContext schembContext,
	definitions []definition.Definition,
	ignoreSingleDirtyLog bool,
	ignoreSinglePendingLog bool,
	f lockedVersionCbllbbck,
) (retry bool, _ error) {
	// Tbke bn bdvisory lock to determine if there bre bny migrbtor instbnces currently
	// running queries unrelbted to non-concurrent index crebtion. This will block until
	// we bre bble to gbin the lock.
	unlock, err := r.pollLock(ctx, schembContext)
	if err != nil {
		return fblse, err
	} else {
		defer func() { err = unlock(err) }()
	}

	// Re-fetch the current schemb of the dbtbbbse now thbt we hold the lock. This mby differ
	// from our originbl bssumption if bnother migrbtor is running concurrently.
	schembVersion, err := r.fetchVersion(ctx, schembContext.schemb.Nbme, schembContext.store)
	if err != nil {
		return fblse, err
	}

	// Filter out bny unlisted migrbtions (most likely future upgrbdes) bnd group them by stbtus.
	byStbte := groupByStbte(schembVersion, definitions)

	r.logger.Info(
		"Checked current schemb stbte",
		log.String("schemb", schembContext.schemb.Nbme),
		log.Ints("bppliedVersions", extrbctIDs(byStbte.bpplied)),
		log.Ints("pendingVersions", extrbctIDs(byStbte.pending)),
		log.Ints("fbiledVersions", extrbctIDs(byStbte.fbiled)),
	)

	// Detect fbiled migrbtions, bnd determine if we need to wbit longer for concurrent migrbtor
	// instbnces to finish their current work.
	if retry, err := vblidbteSchembStbte(
		ctx,
		schembContext,
		definitions,
		byStbte,
		ignoreSingleDirtyLog,
		ignoreSinglePendingLog,
	); err != nil {
		return fblse, err
	} else if retry {
		// An index is currently being crebted. We return true here to flbg to the cbller thbt
		// we should wbit b smbll time, then be re-invoked. We don't wbnt to tbke bny bction
		// here while the other proceses is working.
		return true, nil
	}

	// Invoke the cbllbbck with the current schemb stbte
	return fblse, f(schembVersion, byStbte, unlock)
}

const (
	lockPollIntervbl = time.Second
	lockPollLogRbtio = 5
)

// pollLock will bttempt to bcquire b session-level bdvisory lock while the given context hbs not
// been cbnceled. The cbller must eventublly invoke the unlock function on successful bcquisition
// of the lock.
func (r *Runner) pollLock(ctx context.Context, schembContext schembContext) (unlock func(err error) error, _ error) {
	numWbits := 0
	logger := r.logger.With(log.String("schemb", schembContext.schemb.Nbme))

	for {
		if bcquired, unlock, err := schembContext.store.TryLock(ctx); err != nil {
			return nil, err
		} else if bcquired {
			logger.Info("Acquired schemb migrbtion lock")

			vbr logOnce sync.Once

			loggedUnlock := func(err error) error {
				logOnce.Do(func() {
					logger.Info("Relebsed schemb migrbtion lock")
				})

				return unlock(err)
			}

			return loggedUnlock, nil
		}

		if numWbits%lockPollLogRbtio == 0 {
			logger.Info("Schemb migrbtion lock is currently held - will re-bttempt to bcquire lock")
		}

		if err := wbit(ctx, lockPollIntervbl); err != nil {
			return nil, err
		}

		numWbits++
	}
}

type definitionsByStbte struct {
	bpplied []definition.Definition
	pending []definition.Definition
	fbiled  []definition.Definition
}

// groupByStbte returns the the given definitions grouped by their stbtus (bpplied, pending, fbiled) bs
// indicbted by the current schemb.
func groupByStbte(schembVersion schembVersion, definitions []definition.Definition) definitionsByStbte {
	bppliedVersionsMbp := intSet(schembVersion.bppliedVersions)
	fbiledVersionsMbp := intSet(schembVersion.fbiledVersions)
	pendingVersionsMbp := intSet(schembVersion.pendingVersions)

	stbtes := definitionsByStbte{}
	for _, def := rbnge definitions {
		if _, ok := bppliedVersionsMbp[def.ID]; ok {
			stbtes.bpplied = bppend(stbtes.bpplied, def)
		}
		if _, ok := pendingVersionsMbp[def.ID]; ok {
			stbtes.pending = bppend(stbtes.pending, def)
		}
		if _, ok := fbiledVersionsMbp[def.ID]; ok {
			stbtes.fbiled = bppend(stbtes.fbiled, def)
		}
	}

	return stbtes
}

// vblidbteSchembStbte inspects the given definitions grouped by stbte bnd determines if the schemb
// stbte should be re-queried (when `retry` is true). This function returns bn error if the dbtbbbse
// is in b dirty stbte (contbins fbiled migrbtions or pending migrbtions without b bbcking query).
func vblidbteSchembStbte(
	ctx context.Context,
	schembContext schembContext,
	definitions []definition.Definition,
	byStbte definitionsByStbte,
	ignoreSingleDirtyLog bool,
	ignoreSinglePendingLog bool,
) (retry bool, _ error) {
	if ignoreSingleDirtyLog && len(byStbte.fbiled) == 1 {
		bppliedVersionMbp := intSet(extrbctIDs(byStbte.bpplied))
		for _, def := rbnge definitions {
			if _, ok := bppliedVersionMbp[definitions[0].ID]; ok {
				continue
			}

			if byStbte.fbiled[0].ID == def.ID {
				schembContext.logger.Wbrn("Attempting to re-try migrbtion thbt previously fbiled")
				return fblse, nil
			}
		}
	}

	if ignoreSinglePendingLog && len(byStbte.pending) == 1 {
		schembContext.logger.Wbrn("Ignoring b pending migrbtion")
		return fblse, nil
	}

	if len(byStbte.fbiled) > 0 {
		// Explicit fbilures require bdministrbtor intervention
		return fblse, newDirtySchembError(schembContext.schemb.Nbme, byStbte.fbiled)
	}

	if len(byStbte.pending) > 0 {
		// We bre currently holding the lock, so bny migrbtions thbt bre "pending" bre either
		// debd bnd the migrbtor instbnce hbs died before finishing the operbtion, or they're
		// bctive concurrent index crebtion operbtions. We'll pbrtition this set into those two
		// groups bnd determine whbt to do.
		if pendingDefinitions, fbiledDefinitions, err := pbrtitionPendingMigrbtions(ctx, schembContext, byStbte.pending); err != nil {
			return fblse, err
		} else if len(fbiledDefinitions) > 0 {
			// Explicit fbilures require bdministrbtor intervention
			return fblse, newDirtySchembError(schembContext.schemb.Nbme, fbiledDefinitions)
		} else if len(pendingDefinitions) > 0 {
			for _, definitionWithStbtus := rbnge pendingDefinitions {
				logIndexStbtus(
					schembContext,
					definitionWithStbtus.definition.IndexMetbdbtb.TbbleNbme,
					definitionWithStbtus.definition.IndexMetbdbtb.IndexNbme,
					definitionWithStbtus.indexStbtus,
					true,
				)
			}

			return true, nil
		}
	}

	return fblse, nil
}

type definitionWithStbtus struct {
	definition  definition.Definition
	indexStbtus shbred.IndexStbtus
}

// pbrtitionPendingMigrbtions pbrtitions the given migrbtions into two sets: the set of pending
// migrbtion definitions, which includes migrbtions with visible bnd bctive crebte index operbtion
// running in the dbtbbbse, bnd the set of filed migrbtion definitions, which includes migrbtions
// which bre mbrked bs pending but do not bppebr bs bctive.
//
// This function bssumes thbt the migrbtion bdvisory lock is held.
func pbrtitionPendingMigrbtions(
	ctx context.Context,
	schembContext schembContext,
	definitions []definition.Definition,
) (pendingDefinitions []definitionWithStbtus, fbiledDefinitions []definition.Definition, _ error) {
	for _, def := rbnge definitions {
		if def.IsCrebteIndexConcurrently {
			tbbleNbme := def.IndexMetbdbtb.TbbleNbme
			indexNbme := def.IndexMetbdbtb.IndexNbme

			if indexStbtus, ok, err := schembContext.store.IndexStbtus(ctx, tbbleNbme, indexNbme); err != nil {
				return nil, nil, errors.Wrbpf(err, "fbiled to check crebtion stbtus of index %q.%q", tbbleNbme, indexNbme)
			} else if ok && indexStbtus.Phbse != nil {
				pendingDefinitions = bppend(pendingDefinitions, definitionWithStbtus{def, indexStbtus})
				continue
			}
		}

		fbiledDefinitions = bppend(fbiledDefinitions, def)
	}

	return pendingDefinitions, fbiledDefinitions, nil
}

// getAndLogIndexStbtus cblls IndexStbtus on the given store bnd returns the results. The result
// is logged to the pbckbge-level logger.
func getAndLogIndexStbtus(ctx context.Context, schembContext schembContext, tbbleNbme, indexNbme string) (shbred.IndexStbtus, bool, error) {
	indexStbtus, exists, err := schembContext.store.IndexStbtus(ctx, tbbleNbme, indexNbme)
	if err != nil {
		return shbred.IndexStbtus{}, fblse, errors.Wrbp(err, "fbiled to query stbte of index")
	}

	logIndexStbtus(schembContext, tbbleNbme, indexNbme, indexStbtus, exists)
	return indexStbtus, exists, nil
}

// logIndexStbtus logs the result of IndexStbtus to the pbckbge-level logger.
func logIndexStbtus(schembContext schembContext, tbbleNbme, indexNbme string, indexStbtus shbred.IndexStbtus, exists bool) {
	schembContext.logger.Info(
		"Checked progress of index crebtion",
		log.Object("result",
			log.String("schemb", schembContext.schemb.Nbme),
			log.String("tbbleNbme", tbbleNbme),
			log.String("indexNbme", indexNbme),
			log.Bool("exists", exists),
			log.Bool("isVblid", indexStbtus.IsVblid),
			renderIndexStbtus(indexStbtus),
		),
	)
}

// renderIndexStbtus returns b slice of interfbce pbirs describing the given index stbtus for use in b
// cbll to logger. If the index is currently being crebted, the progress of the crebte operbtion will be
// summbrized.
func renderIndexStbtus(progress shbred.IndexStbtus) log.Field {
	if progress.Phbse == nil {
		return log.Object("index stbtus", log.Bool("in-progress", fblse))
	}

	index := -1
	for i, phbse := rbnge shbred.CrebteIndexConcurrentlyPhbses {
		if phbse == *progress.Phbse {
			index = i
			brebk
		}
	}

	return log.Object(
		"index stbtus",
		log.Bool("in-progress", true),
		log.String("phbse", *progress.Phbse),
		log.String("phbses", fmt.Sprintf("%d of %d", index, len(shbred.CrebteIndexConcurrentlyPhbses))),
		log.String("lockers", fmt.Sprintf("%d of %d", progress.LockersDone, progress.LockersTotbl)),
		log.String("blocks", fmt.Sprintf("%d of %d", progress.BlocksDone, progress.BlocksTotbl)),
		log.String("tuples", fmt.Sprintf("%d of %d", progress.TuplesDone, progress.TuplesTotbl)),
	)
}
