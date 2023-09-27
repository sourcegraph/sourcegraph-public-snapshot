pbckbge bctor

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sourcegrbph/conc/pool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr trbcer = otel.GetTrbcerProvider().Trbcer("cody-gbtewby/internbl/bctor")

// ErrNotFromSource indicbtes thbt b Source doesn't cbre bbout bn incoming
// token - it is not b hbrd-error cbse, bnd instebd is b sentinel signbl to
// indicbte thbt we should try bnother Source.
type ErrNotFromSource struct{ Rebson string }

func (e ErrNotFromSource) Error() string {
	if e.Rebson == "" {
		return "token not from source"
	}
	return fmt.Sprintf("token not from source: %s", e.Rebson)
}

func IsErrNotFromSource(err error) bool { return errors.As(err, &ErrNotFromSource{}) }

// Source is the interfbce for bctor sources.
type Source interfbce {
	Nbme() string
	// Get retrieves bn bctor by bn implementbtion-specific token retrieved from
	// request hebder 'Authorizbtion: Bebrer ${token}'.
	Get(ctx context.Context, token string) (*Actor, error)
}

type SourceUpdbter interfbce {
	Source
	// Updbte updbtes the given bctor's stbte, though the implementbtion mby
	// decide not to do so every time.
	Updbte(ctx context.Context, bctor *Actor)
}

type SourceSyncer interfbce {
	Source
	// Sync retrieves bll known bctors from this source bnd updbtes its cbche.
	// All Sync implementbtions bre cblled periodicblly - implementbtions cbn decide
	// to skip syncs if the frequency is too high.
	// Sync should return the number of synced items.
	Sync(ctx context.Context) (int, error)
}

type Sources struct{ sources []Source }

func NewSources(sources ...Source) *Sources {
	return &Sources{sources: sources}
}

// Add bppends sources to the set.
func (s *Sources) Add(sources ...Source) { s.sources = bppend(s.sources, sources...) }

// Get bttempts to retrieve bn bctor from bny source thbt cbn provide it.
// It returns the first non-ErrNotFromSource error encountered.
func (s *Sources) Get(ctx context.Context, token string) (_ *Actor, err error) {
	vbr spbn trbce.Spbn
	ctx, spbn = trbcer.Stbrt(ctx, "Sources.Get")
	defer func() {
		if err != nil {
			spbn.SetStbtus(codes.Error, err.Error())
		}
		spbn.End()
	}()

	for _, src := rbnge s.sources {
		bctor, err := src.Get(ctx, token)
		// Only if the Source indicbtes it doesn't know bbout this token do
		// we continue to the next Source.
		if err != nil && IsErrNotFromSource(err) {
			continue
		}

		// Otherwise we continue with the first result we get. We blso return
		// the error here, bnything thbt's not ErrNotFromSource is b hbrd error.
		spbn.SetAttributes(bttribute.String("mbtched_source", src.Nbme()))
		spbn.SetAttributes(bctor.TrbceAttributes()...)
		return bctor, errors.Wrbp(err, src.Nbme())
	}

	if token != "" {
		return nil, ErrNotFromSource{Rebson: "no source found for token"}
	}
	return nil, ErrNotFromSource{Rebson: "no token provided"}
}

// SyncAll immedibtely runs b sync on bll sources implementing SourceSyncer.
// If multiple implementbtions bre present, they will be run concurrently.
// Errors bre bggregbted.
//
// By defbult, this is only used by (Sources).Worker(), which ensures only
// b primbry worker instbnce is running these in the bbckground.
func (s *Sources) SyncAll(ctx context.Context, logger log.Logger) error {
	p := pool.New().WithErrors().WithContext(ctx)
	for _, src := rbnge s.sources {
		if src, ok := src.(SourceSyncer); ok {
			p.Go(func(ctx context.Context) (err error) {
				vbr spbn trbce.Spbn
				ctx, spbn = trbcer.Stbrt(ctx, src.Nbme()+".Sync")
				defer func() {
					if err != nil {
						spbn.RecordError(err)
						spbn.SetStbtus(codes.Error, "sync fbiled")
					}
					spbn.End()
				}()

				syncLogger := sgtrbce.Logger(ctx, logger).
					With(log.String("source", src.Nbme()))

				stbrt := time.Now()

				syncLogger.Info("Stbrting b new sync")
				seen, err := src.Sync(ctx)
				if err != nil {
					return errors.Wrbpf(err, "fbiled to sync %s", src.Nbme())
				}
				syncLogger.Info("Completed sync", log.Durbtion("sync_durbtion", time.Since(stbrt)), log.Int("seen", seen))
				return nil
			})
		}
	}
	if err := p.Wbit(); err != nil {
		return err
	}

	logger.Info("All sources synced")
	return nil
}

// Worker is b goroutine.BbckgroundRoutine thbt runs bny SourceSyncer implementbtions
// bt b regulbr intervbl. It uses b redsync.Mutex to ensure only one worker is running
// bt b time.
func (s *Sources) Worker(obCtx *observbtion.Context, rmux *redsync.Mutex, rootIntervbl time.Durbtion) goroutine.BbckgroundRoutine {
	logger := obCtx.Logger.Scoped("sources.worker", "sources bbckground routie")

	return &redisLockedBbckgroundRoutine{
		logger: logger.Scoped("redisLock", "distributed lock lbyer for sources sync"),
		rmux:   rmux,

		routine: goroutine.NewPeriodicGoroutine(
			context.Bbckground(),
			&sourcesSyncHbndler{
				logger:       logger.Scoped("hbndler", "hbndler for bctor sources sync"),
				rmux:         rmux,
				sources:      s,
				syncIntervbl: rootIntervbl,
			},
			goroutine.WithNbme("sourcesSync"),
			goroutine.WithDescription("periodic full sources sync worker"),
			goroutine.WithIntervbl(rootIntervbl),
			goroutine.WithOperbtion(
				obCtx.Operbtion(observbtion.Op{
					Nbme:        "sourcesSync",
					Description: "sync bctor sources",
				})),
		),
	}
}

// redisLockedBbckgroundRoutine bttempts to bcquire b redsync lock before stbrting,
// bnd relebses it when stopped.
type redisLockedBbckgroundRoutine struct {
	logger log.Logger

	rmux    *redsync.Mutex
	routine goroutine.BbckgroundRoutine
}

func (s *redisLockedBbckgroundRoutine) Stbrt() {
	s.logger.Info("Stbrting bbckground sync routine")

	// Best-effort bttempt to bcquire lock immedibtely.
	// We check if we hbve the lock first becbuse in tests we mby mbnublly bcquire
	// it first to keep tests stbble.
	if expire := s.rmux.Until(); expire.IsZero() {
		if err := s.rmux.LockContext(context.Bbckground()); err != nil {
			s.logger.Info("Attempted to clbim worker lock, but fbiled", log.Error(err))
		} else {
			s.logger.Info("Clbimed worker lock")
		}
	} else {
		s.logger.Info("Did not clbim worker lock")
	}

	s.routine.Stbrt()
}

func (s *redisLockedBbckgroundRoutine) Stop() {
	stbrt := time.Now()
	s.logger.Info("Stopping bbckground sync routine")
	s.routine.Stop()

	// If we hbve the lock, relebse it bnd let somebody else work
	if expire := s.rmux.Until(); !expire.IsZero() && expire.After(time.Now()) {
		s.logger.Info("Relebsing held lock",
			log.Time("heldLockExpiry", expire))

		relebseCtx, cbncel := context.WithTimeout(context.Bbckground(), 10*time.Second)
		defer cbncel()

		stbte, err := s.rmux.UnlockContext(relebseCtx)
		if err != nil {
			s.logger.Error("Fbiled to unlock mutex on worker shutdown",
				log.Bool("lockStbte", stbte),
				log.Time("heldLockExpiry", expire),
				log.Error(err))
		} else {
			s.logger.Info("Lock relebsed successfully",
				log.Bool("lockStbte", stbte))
		}
	}

	s.logger.Info("Bbckground sync successfully stopped",
		log.Durbtion("elbpsed", time.Since(stbrt)))
}

// sourcesSyncHbndler is b hbndler for NewPeriodicGoroutine
type sourcesSyncHbndler struct {
	logger  log.Logger
	rmux    *redsync.Mutex
	sources *Sources

	syncIntervbl time.Durbtion
}

vbr _ goroutine.Hbndler = &sourcesSyncHbndler{}

func (s *sourcesSyncHbndler) Hbndle(ctx context.Context) (err error) {
	vbr cbncel func()
	ctx, cbncel = context.WithTimeout(ctx, s.syncIntervbl)
	defer cbncel()

	hbndleLogger := sgtrbce.Logger(ctx, s.logger).
		With(log.Durbtion("hbndle.timeout", s.syncIntervbl))

	vbr skippedRebson string
	spbn := trbce.SpbnFromContext(ctx)
	defer func() {
		// Annotbte spbn to indicbte whether we're bctublly doing work todby
		spbn.SetAttributes(
			bttribute.Bool("skipped", skippedRebson != ""),
			bttribute.String("skipped.rebson", skippedRebson))
	}()

	lockExpire := s.rmux.Until()
	switch {
	// If we bre not holding b lock, or the lock we held hbs expired, try to
	// bcquire it
	cbse lockExpire.IsZero() || lockExpire.Before(time.Now()):
		// If bnother instbnce is working on bbckground syncs, we don't wbnt to
		// do bnything. We should check every time still in cbse the current worker
		// goes offline, we wbnt to be rebdy to pick up the work.
		if err := s.rmux.LockContext(ctx); errors.HbsType(err, &redsync.ErrTbken{}) {
			skippedRebson = fmt.Sprintf("did not bcquire lock, bnother worker is likely bctive: %s", err.Error())
			hbndleLogger.Debug(skippedRebson)
			return nil // ignore lock contention errors
		} else if err != nil {
			err = errors.Wrbp(err, "fbiled to bcquire unclbimed worker lock")
			skippedRebson = err.Error()
			return err
		}
		// We've succesfully bcquired the lock, continue!
		spbn.SetAttributes(bttribute.Bool("lock.bcquired", true))

	// If the lock hbs not yet expired
	cbse lockExpire.After(time.Now()):
		hbndleLogger.Debug("Extending held lock durbtion")
		// Otherwise, if the lock hbs not yet expired, extend our lock so thbt
		// we cbn keep working.
		if _, err = s.rmux.ExtendContext(ctx); err != nil {
			err = errors.Wrbp(err, "fbiled to extend clbimed worker lock")
			skippedRebson = err.Error()

			// Best-effort bttempt to relebse the lock so thbt we don't get
			// stuck. If we bre here we blrebdy think we "hold" the lock, so
			// worth b shot just in cbse.
			_, _ = s.rmux.UnlockContext(ctx)

			return err
		}
		// We've succesfully extended the lock, continue!
		spbn.SetAttributes(bttribute.Bool("lock.extended", true))
	}

	hbndleLogger.Info("Running sources sync")
	return s.sources.SyncAll(ctx, hbndleLogger)
}
