package actor

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var tracer = otel.GetTracerProvider().Tracer("cody-gateway/internal/actor")

// ErrNotFromSource indicates that a Source doesn't care about an incoming
// token - it is not a hard-error case, and instead is a sentinel signal to
// indicate that we should try another Source.
type ErrNotFromSource struct{ Reason string }

func (e ErrNotFromSource) Error() string {
	if e.Reason == "" {
		return "token not from source"
	}
	return fmt.Sprintf("token not from source: %s", e.Reason)
}

func IsErrNotFromSource(err error) bool { return errors.As(err, &ErrNotFromSource{}) }

// Source is the interface for actor sources.
type Source interface {
	Name() string
	// Get retrieves an actor by an implementation-specific token retrieved from
	// request header 'Authorization: Bearer ${token}'.
	Get(ctx context.Context, token string) (*Actor, error)
}

// ErrActorRecentlyUpdated can be used to indicate that an actor cannot be
// updated because it was already updated more recently than allowed by a
// Source implementation.
type ErrActorRecentlyUpdated struct {
	RetryAt time.Time
}

func (e ErrActorRecentlyUpdated) Error() string {
	return fmt.Sprintf("actor was recently updated - try again in %s",
		time.Until(e.RetryAt).Truncate(time.Second).String())
}

func IsErrActorRecentlyUpdated(err error) bool { return errors.As(err, &ErrActorRecentlyUpdated{}) }

type SourceUpdater interface {
	Source
	// Update updates the given actor's state, though the implementation may
	// decide not to do so every time.
	//
	// Error can be ErrActorRecentlyUpdated if the actor was updated too recently.
	Update(ctx context.Context, actor *Actor) error
}

type SourceSyncer interface {
	Source
	// Sync retrieves all known actors from this source and updates its cache.
	// All Sync implementations are called periodically - implementations can decide
	// to skip syncs if the frequency is too high.
	// Sync should return the number of synced items.
	Sync(ctx context.Context) (int, error)
}

type Sources struct{ sources []Source }

func NewSources(sources ...Source) *Sources {
	return &Sources{sources: sources}
}

// Add appends sources to the set.
func (s *Sources) Add(sources ...Source) { s.sources = append(s.sources, sources...) }

// Get attempts to retrieve an actor from any source that can provide it.
// It returns the first non-ErrNotFromSource error encountered.
func (s *Sources) Get(ctx context.Context, token string) (_ *Actor, err error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "Sources.Get")
	defer func() {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	for _, src := range s.sources {
		actor, err := src.Get(ctx, token)
		// Only if the Source indicates it doesn't know about this token do
		// we continue to the next Source.
		if err != nil && IsErrNotFromSource(err) {
			continue
		}

		// Otherwise we continue with the first result we get. We also return
		// the error here, anything that's not ErrNotFromSource is a hard error.
		span.SetAttributes(attribute.String("matched_source", src.Name()))
		span.SetAttributes(actor.TraceAttributes()...)
		return actor, errors.Wrap(err, src.Name())
	}

	if token != "" {
		return nil, ErrNotFromSource{Reason: "no source found for token"}
	}
	return nil, ErrNotFromSource{Reason: "no token provided"}
}

// SyncAll immediately runs a sync on all sources implementing SourceSyncer.
// If multiple implementations are present, they will be run concurrently.
// Errors are aggregated.
//
// By default, this is only used by (Sources).Worker(), which ensures only
// a primary worker instance is running these in the background.
func (s *Sources) SyncAll(ctx context.Context, logger log.Logger) error {
	p := pool.New().WithErrors().WithContext(ctx)
	for _, src := range s.sources {
		if src, ok := src.(SourceSyncer); ok {
			p.Go(func(ctx context.Context) (err error) {
				var span trace.Span
				ctx, span = tracer.Start(ctx, src.Name()+".Sync")
				defer func() {
					if err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, "sync failed")
					}
					span.End()
				}()

				syncLogger := sgtrace.Logger(ctx, logger).
					With(log.String("source", src.Name()))

				start := time.Now()

				syncLogger.Info("Starting a new sync")
				seen, err := src.Sync(ctx)
				if err != nil {
					return errors.Wrapf(err, "failed to sync %s", src.Name())
				}
				syncLogger.Info("Completed sync", log.Duration("sync_duration", time.Since(start)), log.Int("seen", seen))
				return nil
			})
		}
	}
	if err := p.Wait(); err != nil {
		return err
	}

	logger.Info("All sources synced")
	return nil
}

// Worker is a goroutine.BackgroundRoutine that runs any SourceSyncer implementations
// at a regular interval. It uses a redsync.Mutex to ensure only one worker is running
// at a time.
func (s *Sources) Worker(obCtx *observation.Context, rmux *redsync.Mutex, rootInterval time.Duration) goroutine.BackgroundRoutine {
	logger := obCtx.Logger.Scoped("sources.worker")

	return &redisLockedBackgroundRoutine{
		logger: logger.Scoped("redisLock"),
		rmux:   rmux,

		routine: goroutine.NewPeriodicGoroutine(
			context.Background(),
			&sourcesSyncHandler{
				logger:       logger.Scoped("handler"),
				rmux:         rmux,
				sources:      s,
				syncInterval: rootInterval,
			},
			goroutine.WithName("sourcesSync"),
			goroutine.WithDescription("periodic full sources sync worker"),
			goroutine.WithInterval(rootInterval),
			goroutine.WithOperation(
				obCtx.Operation(observation.Op{
					Name:        "sourcesSync",
					Description: "sync actor sources",
				})),
		),
	}
}

// redisLockedBackgroundRoutine attempts to acquire a redsync lock before starting,
// and releases it when stopped.
type redisLockedBackgroundRoutine struct {
	logger log.Logger

	rmux    *redsync.Mutex
	routine goroutine.BackgroundRoutine
}

func (s *redisLockedBackgroundRoutine) Start() {
	s.logger.Info("Starting background sync routine")

	// Best-effort attempt to acquire lock immediately.
	// We check if we have the lock first because in tests we may manually acquire
	// it first to keep tests stable.
	if expire := s.rmux.Until(); expire.IsZero() {
		if err := s.rmux.LockContext(context.Background()); err != nil {
			s.logger.Info("Attempted to claim worker lock, but failed", log.Error(err))
		} else {
			s.logger.Info("Claimed worker lock")
		}
	} else {
		s.logger.Info("Did not claim worker lock")
	}

	s.routine.Start()
}

func (s *redisLockedBackgroundRoutine) Stop() {
	start := time.Now()
	s.logger.Info("Stopping background sync routine")
	s.routine.Stop()

	// If we have the lock, release it and let somebody else work
	if expire := s.rmux.Until(); !expire.IsZero() && expire.After(time.Now()) {
		s.logger.Info("Releasing held lock",
			log.Time("heldLockExpiry", expire))

		releaseCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		state, err := s.rmux.UnlockContext(releaseCtx)
		if err != nil {
			s.logger.Error("Failed to unlock mutex on worker shutdown",
				log.Bool("lockState", state),
				log.Time("heldLockExpiry", expire),
				log.Error(err))
		} else {
			s.logger.Info("Lock released successfully",
				log.Bool("lockState", state))
		}
	}

	s.logger.Info("Background sync successfully stopped",
		log.Duration("elapsed", time.Since(start)))
}

// sourcesSyncHandler is a handler for NewPeriodicGoroutine
type sourcesSyncHandler struct {
	logger  log.Logger
	rmux    *redsync.Mutex
	sources *Sources

	syncInterval time.Duration
}

var _ goroutine.Handler = &sourcesSyncHandler{}

func (s *sourcesSyncHandler) Handle(ctx context.Context) (err error) {
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, s.syncInterval)
	defer cancel()

	handleLogger := sgtrace.Logger(ctx, s.logger).
		With(log.Duration("handle.timeout", s.syncInterval))

	var skippedReason string
	span := trace.SpanFromContext(ctx)
	defer func() {
		// Annotate span to indicate whether we're actually doing work today
		span.SetAttributes(
			attribute.Bool("skipped", skippedReason != ""),
			attribute.String("skipped.reason", skippedReason))
	}()

	lockExpire := s.rmux.Until()
	switch {
	// If we are not holding a lock, or the lock we held has expired, try to
	// acquire it
	case lockExpire.IsZero() || lockExpire.Before(time.Now()):
		// If another instance is working on background syncs, we don't want to
		// do anything. We should check every time still in case the current worker
		// goes offline, we want to be ready to pick up the work.
		if err := s.rmux.LockContext(ctx); errors.HasType(err, &redsync.ErrTaken{}) {
			skippedReason = fmt.Sprintf("did not acquire lock, another worker is likely active: %s", err.Error())
			handleLogger.Debug(skippedReason)
			return nil // ignore lock contention errors
		} else if err != nil {
			err = errors.Wrap(err, "failed to acquire unclaimed worker lock")
			skippedReason = err.Error()
			return err
		}
		// We've succesfully acquired the lock, continue!
		span.SetAttributes(attribute.Bool("lock.acquired", true))

	// If the lock has not yet expired
	case lockExpire.After(time.Now()):
		handleLogger.Debug("Extending held lock duration")
		// Otherwise, if the lock has not yet expired, extend our lock so that
		// we can keep working.
		if _, err = s.rmux.ExtendContext(ctx); err != nil {
			err = errors.Wrap(err, "failed to extend claimed worker lock")
			skippedReason = err.Error()

			// Best-effort attempt to release the lock so that we don't get
			// stuck. If we are here we already think we "hold" the lock, so
			// worth a shot just in case.
			_, _ = s.rmux.UnlockContext(ctx)

			return err
		}
		// We've succesfully extended the lock, continue!
		span.SetAttributes(attribute.Bool("lock.extended", true))
	}

	handleLogger.Info("Running sources sync")
	return s.sources.SyncAll(ctx, handleLogger)
}

type FakeSource struct {
	SourceName codygateway.ActorSource
}

func (m FakeSource) Name() string {
	return string(m.SourceName)
}

func (m FakeSource) Get(_ context.Context, _ string) (*Actor, error) {
	// TODO implement me
	panic("implement me")
}

var _ Source = FakeSource{}
