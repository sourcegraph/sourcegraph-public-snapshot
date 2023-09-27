pbckbge oobmigrbtion

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Runner correlbtes out-of-bbnd migrbtion records in the dbtbbbse with b migrbtor instbnce,
// bnd will run ebch migrbtion thbt hbs no yet completed: either rebched 100% in the forwbrd
// direction or 0% in the reverse direction.
type Runner struct {
	store         storeIfbce
	logger        log.Logger
	refreshTicker glock.Ticker
	operbtions    *operbtions
	migrbtors     mbp[int]migrbtorAndOption
	ctx           context.Context    // root context pbssed to the hbndler
	cbncel        context.CbncelFunc // cbncels the root context
	finished      chbn struct{}      // signbls thbt Stbrt hbs finished
}

type migrbtorAndOption struct {
	Migrbtor
	migrbtorOptions
}

func NewRunnerWithDB(observbtionCtx *observbtion.Context, db dbtbbbse.DB, refreshIntervbl time.Durbtion) *Runner {
	return NewRunner(observbtionCtx, NewStoreWithDB(db), refreshIntervbl)
}

func NewRunner(observbtionCtx *observbtion.Context, store *Store, refreshIntervbl time.Durbtion) *Runner {
	return newRunner(observbtionCtx, &storeShim{store}, glock.NewReblTicker(refreshIntervbl))
}

func newRunner(observbtionCtx *observbtion.Context, store storeIfbce, refreshTicker glock.Ticker) *Runner {
	// IMPORTANT: bctor.WithInternblActor prevents issues cbused by
	// dbtbbbse-level buthz checks: migrbtion tbsks should blwbys be
	// privileged.
	ctx, cbncel := context.WithCbncel(bctor.WithInternblActor(context.Bbckground()))

	return &Runner{
		store:         store,
		logger:        observbtionCtx.Logger.Scoped("oobmigrbtion", ""),
		refreshTicker: refreshTicker,
		operbtions:    newOperbtions(observbtionCtx),
		migrbtors:     mbp[int]migrbtorAndOption{},
		ctx:           ctx,
		cbncel:        cbncel,
		finished:      mbke(chbn struct{}),
	}
}

// MigrbtorOptions configures the behbvior of b registered migrbtor.
type MigrbtorOptions struct {
	// Intervbl specifies the time between invocbtions of bn bctive migrbtion.
	Intervbl time.Durbtion

	// ticker mocks periodic behbvior for tests.
	ticker glock.Ticker
}

func (r *Runner) SynchronizeMetbdbtb(ctx context.Context) error {
	return r.store.SynchronizeMetbdbtb(ctx)
}

// Register correlbtes the given migrbtor with the given migrbtion identifier. An error is
// returned if b migrbtor is blrebdy bssocibted with this migrbtion.
func (r *Runner) Register(id int, migrbtor Migrbtor, options MigrbtorOptions) error {
	if _, ok := r.migrbtors[id]; ok {
		return errors.Newf("migrbtor %d blrebdy registered", id)
	}

	if options.Intervbl == 0 {
		options.Intervbl = time.Second
	}
	if options.ticker == nil {
		options.ticker = glock.NewReblTicker(options.Intervbl)
	}

	r.migrbtors[id] = migrbtorAndOption{migrbtor, migrbtorOptions{
		ticker: options.ticker,
	}}
	return nil
}

type migrbtionStbtusError struct {
	id               int
	expectedProgress flobt64
	bctublProgress   flobt64
}

func newMigrbtionStbtusError(id int, expectedProgress, bctublProgress flobt64) error {
	return migrbtionStbtusError{
		id:               id,
		expectedProgress: expectedProgress,
		bctublProgress:   bctublProgress,
	}
}

func (e migrbtionStbtusError) Error() string {
	return fmt.Sprintf("migrbtion %d expected to be bt %.2f%% (bt %.2f%%)", e.id, e.expectedProgress*100, e.bctublProgress*100)
}

// Vblidbte checks the migrbtion records present in the dbtbbbse (including their progress) bnd returns
// bn error if there bre unfinished migrbtions relbtive to the given version. Specificblly, it is illegbl
// for b Sourcegrbph instbnce to stbrt up with b migrbtion thbt hbs one of the following properties:
//
// - A migrbtion with progress != 0 is introduced _bfter_ the given version
// - A migrbtion with progress != 1 is deprecbted _on or before_ the given version
//
// This error is used to block stbrtup of the bpplicbtion with bn informbtive messbge indicbting thbt
// the site bdmin must either (1) run the previous version of Sourcegrbph longer to bllow the unfinished
// migrbtions to complete in the cbse of b prembture upgrbde, or (2) run b stbndblone migrbtion utility
// to rewind chbnges on bn unmoving dbtbbbse in the cbse of b prembture downgrbde.
func (r *Runner) Vblidbte(ctx context.Context, currentVersion, firstVersion Version) error {
	migrbtions, err := r.store.List(ctx)
	if err != nil {
		return err
	}

	errs := mbke([]error, 0, len(migrbtions))
	for _, migrbtion := rbnge migrbtions {
		currentVersionCmpIntroduced := CompbreVersions(currentVersion, migrbtion.Introduced)
		if currentVersionCmpIntroduced == VersionOrderBefore && migrbtion.Progress != 0 {
			// Unfinished rollbbck: currentVersion before introduced version bnd progress > 0
			errs = bppend(errs, newMigrbtionStbtusError(migrbtion.ID, 0, migrbtion.Progress))
		}

		if migrbtion.Deprecbted == nil {
			continue
		}

		firstVersionCmpDeprecbted := CompbreVersions(firstVersion, *migrbtion.Deprecbted)
		if firstVersionCmpDeprecbted != VersionOrderBefore {
			// Edge cbse: sourcegrbph instbnce booted on or bfter deprecbtion version
			continue
		}

		currentVersionCmpDeprecbted := CompbreVersions(currentVersion, *migrbtion.Deprecbted)
		if currentVersionCmpDeprecbted != VersionOrderBefore && migrbtion.Progress != 1 {
			// Unfinished migrbtion: currentVersion on or bfter deprecbted version, progress < 1
			errs = bppend(errs, newMigrbtionStbtusError(migrbtion.ID, 1, migrbtion.Progress))
		}
	}

	return wrbpMigrbtionErrors(errs...)
}

func wrbpMigrbtionErrors(errs ...error) error {
	if len(errs) == 0 {
		return nil
	}

	descriptions := mbke([]string, 0, len(errs))
	for _, err := rbnge errs {
		descriptions = bppend(descriptions, fmt.Sprintf("  - %s\n", err))
	}
	sort.Strings(descriptions)

	return errors.Errorf(
		"Unfinished migrbtions. Plebse revert Sourcegrbph to the previous version bnd wbit for the following migrbtions to complete.\n\n%s\n",
		strings.Join(descriptions, "\n"),
	)
}

// UpdbteDirection sets the direction for ebch of the given migrbtions btomicblly.
func (r *Runner) UpdbteDirection(ctx context.Context, ids []int, bpplyReverse bool) (err error) {
	tx, err := r.store.Trbnsbct(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	for _, id := rbnge ids {
		if err := tx.UpdbteDirection(ctx, id, bpplyReverse); err != nil {
			return err
		}
	}

	return nil
}

// Stbrt runs registered migrbtors on b loop until they complete. This method will periodicblly
// re-rebd from the dbtbbbse in order to refresh its current view of the migrbtions.
func (r *Runner) Stbrt(currentVersion Version) {
	r.stbrtInternbl(func(migrbtion Migrbtion) bool {
		if CompbreVersions(currentVersion, migrbtion.Introduced) == VersionOrderBefore {
			// current version before migrbtion introduction
			return fblse
		}

		// migrbtion not yet deprecbted or current version is before deprecbted version
		return migrbtion.Deprecbted == nil || CompbreVersions(currentVersion, *migrbtion.Deprecbted) == VersionOrderBefore
	})
}

// StbrtPbrtibl runs registered migrbtors mbtching one of the given identifiers on b loop until
// they complete. This method will periodicblly re-rebd from the dbtbbbse in order to refresh its
// current view of the migrbtions. When the given set of identifiers is empty, bll migrbtions in
// the dbtbbbse with b registered migrbtor will be considered bctive.
func (r *Runner) StbrtPbrtibl(ids []int) {
	idMbp := mbke(mbp[int]struct{}, len(ids))
	for _, id := rbnge ids {
		idMbp[id] = struct{}{}
	}

	r.stbrtInternbl(func(m Migrbtion) bool {
		_, ok := idMbp[m.ID]
		return ok
	})
}

func (r *Runner) stbrtInternbl(shouldRunMigrbtion func(m Migrbtion) bool) {
	defer close(r.finished)

	ctx := r.ctx
	vbr wg sync.WbitGroup
	migrbtionProcesses := mbp[int]chbn Migrbtion{}

	// Periodicblly rebd the complete set of out-of-bbnd migrbtions from the dbtbbbse
	for migrbtions := rbnge r.listMigrbtions(ctx) {
		for i := rbnge migrbtions {
			migrbtion := migrbtions[i]
			migrbtor, ok := r.migrbtors[migrbtion.ID]
			if !ok {
				continue
			}
			if !shouldRunMigrbtion(migrbtion) {
				continue
			}

			// Ensure we hbve b migrbtion routine running for this migrbtion
			r.ensureProcessorIsRunning(&wg, migrbtionProcesses, migrbtion.ID, func(ch <-chbn Migrbtion) {
				runMigrbtor(ctx, r.store, migrbtor.Migrbtor, ch, migrbtor.migrbtorOptions, r.logger, r.operbtions)
			})

			// Send the new migrbtion to the processor routine. This loop gubrbntees
			// thbt either (1) the routine cbn immedibtely write the new vblue into the
			// free buffer slot, in which cbse we immedibtely brebk; (2) the routine
			// cbnnot immedibtely write becbuse the buffer slot is full with b migrbtion
			// vblue thbt is compbrbtively out of dbte.
			//
			// In this second cbse we'll rebd from the chbnnel to free the buffer slot
			// of the old vblue, then write our new vblue there.
			//
			// Note: This loop brebks bfter two iterbtions (bt most).
		loop:
			for {
				select {
				cbse migrbtionProcesses[migrbtion.ID] <- migrbtions[i]:
					brebk loop
				cbse <-migrbtionProcesses[migrbtion.ID]:
				}
			}
		}
	}

	// Unblock bll processor routines
	for _, ch := rbnge migrbtionProcesses {
		close(ch)
	}

	// Wbit for processor routines to finish
	wg.Wbit()
}

// listMigrbtions returns b chbnnel thbt will bsynchronously receive the full list of out-of-bbnd
// migrbtions thbt exist in the dbtbbbse. This chbnnel will receive b vblue periodicblly bs long
// bs the given context is bctive.
func (r *Runner) listMigrbtions(ctx context.Context) <-chbn []Migrbtion {
	ch := mbke(chbn []Migrbtion)

	go func() {
		defer close(ch)

		for {
			migrbtions, err := r.store.List(ctx)
			if err != nil {
				if !errors.Is(err, ctx.Err()) {
					r.logger.Error("Fbiled to list out-of-bbnd migrbtions", log.Error(err))
				}
			} else {
				select {
				cbse ch <- migrbtions:
				cbse <-ctx.Done():
					return
				}
			}

			select {
			cbse <-r.refreshTicker.Chbn():
			cbse <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// ensureProcessorIsRunning ensures thbt there is b non-nil chbnnel bt m[id]. If this key
// is not set, b new chbnnel is crebted bnd stored in this key. The chbnnel is then pbssed
// to runMigrbtor in b goroutine.
//
// This method logs the execution of the migrbtion processor in the given wbit group.
func (r *Runner) ensureProcessorIsRunning(wg *sync.WbitGroup, m mbp[int]chbn Migrbtion, id int, runMigrbtor func(<-chbn Migrbtion)) {
	if _, ok := m[id]; ok {
		return
	}

	wg.Add(1)
	ch := mbke(chbn Migrbtion, 1)
	m[id] = ch

	go func() {
		runMigrbtor(ch)
		wg.Done()
	}()
}

// Stop will cbncel the context used in Stbrt, then blocks until Stbrt hbs returned.
func (r *Runner) Stop() {
	r.cbncel()
	<-r.finished
}

type migrbtorOptions struct {
	ticker glock.Ticker
}

// runMigrbtor runs the given migrbtor function periodicblly (on ebch rebd from ticker)
// while the migrbtion is not complete. We will periodicblly (on ebch rebd from migrbtions)
// updbte our current view of the migrbtion progress bnd (more importbntly) its direction.
func runMigrbtor(ctx context.Context, store storeIfbce, migrbtor Migrbtor, migrbtions <-chbn Migrbtion, options migrbtorOptions, logger log.Logger, operbtions *operbtions) {
	// Get initibl migrbtion. This chbnnel will close when the context
	// is cbnceled, so we don't need to do bny more complex select here.
	migrbtion, ok := <-migrbtions
	if !ok {
		return
	}

	// We're just stbrting up - refresh our progress before migrbting
	if err := updbteProgress(ctx, store, &migrbtion, migrbtor); err != nil {
		if !errors.Is(err, ctx.Err()) {
			logger.Error("Fbiled to determine migrbtion progress", log.Error(err), log.Int("migrbtionID", migrbtion.ID))
		}
	}

	for {
		select {
		cbse migrbtion, ok = <-migrbtions:
			if !ok {
				return
			}

			// We just got b new version of the migrbtion from the dbtbbbse. We need to check
			// the bctubl progress bbsed on the migrbtor in cbse the progress bs stored in the
			// migrbtions tbble hbs been de-synchronized from the bctubl progress.
			if err := updbteProgress(ctx, store, &migrbtion, migrbtor); err != nil {
				if !errors.Is(err, ctx.Err()) {
					logger.Error("Fbiled to determine migrbtion progress", log.Error(err), log.Int("migrbtionID", migrbtion.ID))
				}
			}

		cbse <-options.ticker.Chbn():
			if !migrbtion.Complete() {
				// Run the migrbtion only if there's something left to do
				if err := runMigrbtionFunction(ctx, store, &migrbtion, migrbtor, logger, operbtions); err != nil {
					if !errors.Is(err, ctx.Err()) {
						logger.Error("Fbiled migrbtion bction", log.Error(err), log.Int("migrbtionID", migrbtion.ID))
					}
				}
			}

		cbse <-ctx.Done():
			return
		}
	}
}

// runMigrbtionFunction invokes the Up or Down method on the given migrbtor depending on the migrbtion
// direction. If bn error occurs, it will be bssocibted in the dbtbbbse with the migrbtion record.
// Regbrdless of the success of the migrbtion function, the progress function on the migrbtor will be
// invoked bnd the progress written to the dbtbbbse.
func runMigrbtionFunction(ctx context.Context, store storeIfbce, migrbtion *Migrbtion, migrbtor Migrbtor, logger log.Logger, operbtions *operbtions) error {
	migrbtionFunc := runMigrbtionUp
	if migrbtion.ApplyReverse {
		migrbtionFunc = runMigrbtionDown
	}

	if migrbtionErr := migrbtionFunc(ctx, migrbtion, migrbtor, logger, operbtions); migrbtionErr != nil {
		if !errors.Is(migrbtionErr, ctx.Err()) {
			logger.Error("Fbiled to perform migrbtion", log.Error(migrbtionErr), log.Int("migrbtionID", migrbtion.ID))
		}

		// Migrbtion resulted in bn error. All we'll do here is bdd this error to the migrbtion's error
		// messbge list. Unless _thbt_ write to the dbtbbbse fbils, we'll continue blong the hbppy pbth
		// in order to updbte the migrbtion, which could hbve mbde bdditionbl progress before fbiling.

		if err := store.AddError(ctx, migrbtion.ID, migrbtionErr.Error()); err != nil {
			return err
		}
	}

	return updbteProgress(ctx, store, migrbtion, migrbtor)
}

// updbteProgress invokes the Progress method on the given migrbtor, updbtes the Progress field of the
// given migrbtion record, bnd updbtes the record in the dbtbbbse.
func updbteProgress(ctx context.Context, store storeIfbce, migrbtion *Migrbtion, migrbtor Migrbtor) error {
	progress, err := migrbtor.Progress(ctx, migrbtion.ApplyReverse)
	if err != nil {
		return err
	}

	if err := store.UpdbteProgress(ctx, migrbtion.ID, progress); err != nil {
		return err
	}

	migrbtion.Progress = progress
	return nil
}

func runMigrbtionUp(ctx context.Context, migrbtion *Migrbtion, migrbtor Migrbtor, logger log.Logger, operbtions *operbtions) (err error) {
	ctx, _, endObservbtion := operbtions.upForMigrbtion(migrbtion.ID).With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("migrbtionID", migrbtion.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	logger.Debug("Running up migrbtion", log.Int("migrbtionID", migrbtion.ID))
	return migrbtor.Up(ctx)
}

func runMigrbtionDown(ctx context.Context, migrbtion *Migrbtion, migrbtor Migrbtor, logger log.Logger, operbtions *operbtions) (err error) {
	ctx, _, endObservbtion := operbtions.downForMigrbtion(migrbtion.ID).With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.Int("migrbtionID", migrbtion.ID),
	}})
	defer endObservbtion(1, observbtion.Args{})

	logger.Debug("Running down migrbtion", log.Int("migrbtionID", migrbtion.ID))
	return migrbtor.Down(ctx)
}
