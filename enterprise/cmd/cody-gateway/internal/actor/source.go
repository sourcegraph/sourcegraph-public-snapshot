package actor

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/log"

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

type SourceUpdater interface {
	Source
	// Update updates the given actor's state, though the implementation may
	// decide not to do so every time.
	Update(ctx context.Context, actor *Actor)
}

type SourceSyncer interface {
	Source
	// Sync retrieves all known actors from this source and updates its cache.
	// All Sync implementations are called periodically - implementations can decide
	// to skip syncs if the frequency is too high.
	// Sync should return the number of synced items.
	Sync(ctx context.Context) (int, error)
}

type Sources struct {
	sources []Source

	syncRunning atomic.Bool
}

func NewSources(sources ...Source) *Sources { return &Sources{sources: sources} }

func (s *Sources) Get(ctx context.Context, token string) (_ *Actor, err error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "Sources.Get")
	defer func() {
		span.RecordError(err) // don't set status, not necessarily a hard failure
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
		return actor, errors.Wrap(err, src.Name())
	}

	return nil, errors.New("no source found for token")
}

// SyncAll immediately runs a sync on all sources implementing SourceSyncer.
// If multiple implementations are present, they will be run concurrently.
// Errors are aggregated.
//
// By default, this is only used by (Sources).Worker(), which ensures only
// a primary worker instance is running these in the background.
func (s *Sources) SyncAll(ctx context.Context, logger log.Logger) error {
	if !s.syncRunning.CompareAndSwap(s.syncRunning.Load(), true) {
		return errors.New("sources.SyncAll already running")
	}

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
	logger := obCtx.Logger.Scoped("sources.worker", "sources background routie")

	return &redisLockedBackgroundRoutine{
		logger: logger.Scoped("redisLock", "distributed lock layer for sources sync"),
		rmux:   rmux,

		routine: goroutine.NewPeriodicGoroutine(
			context.Background(),
			&sourcesSyncHandler{
				logger:  logger.Scoped("handler", "handler for actor sources sync"),
				rmux:    rmux,
				sources: s,
			},
			goroutine.WithName("periodic.sourcesSync"),
			goroutine.WithDescription("periodic sources sync worker"),
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
		_ = s.rmux.LockContext(context.Background())
	}

	s.routine.Start()
}

func (s *redisLockedBackgroundRoutine) Stop() {
	s.logger.Info("Stopping background sync routine")
	s.routine.Stop()

	// If we have the lock, release it and let somebody else work
	if expire := s.rmux.Until(); !expire.IsZero() {
		s.logger.Info("Releasing held lock")
		_, err := s.rmux.Unlock()
		if err != nil {
			s.logger.Warn("Failed to unlock mutex after work completed", log.Error(err))
		}
	}
}

// sourcesSyncHandler is a handler for NewPeriodicGoroutine
type sourcesSyncHandler struct {
	logger  log.Logger
	rmux    *redsync.Mutex
	sources *Sources
}

var _ goroutine.Handler = &sourcesSyncHandler{}

func (s *sourcesSyncHandler) Handle(ctx context.Context) (err error) {
	handleLogger := sgtrace.Logger(ctx, s.logger)

	// If we are not holding a lock, try to acquire it.
	if expire := s.rmux.Until(); expire.IsZero() {
		// If another instance is working on background syncs, we don't want to
		// do anything. We should check every time still in case the current worker
		// goes offline, we want to be ready to pick up the work.
		if err := s.rmux.LockContext(ctx); errors.HasType(err, &redsync.ErrTaken{}) {
			msg := fmt.Sprintf("did not acquire lock, another worker is likely active: %s", err.Error())
			handleLogger.Debug(msg)
			trace.SpanFromContext(ctx).SetAttributes(
				attribute.Bool("skipped", true),
				attribute.String("reason", msg))

			return nil // ignore lock contention errors
		} else if err != nil {
			return errors.Wrap(err, "acquire worker lock")
		}
	} else {
		handleLogger.Debug("Extending lock duration")
		// Otherwise, extend our lock so that we can keep working.
		_, _ = s.rmux.ExtendContext(ctx)
	}

	// Annotate span to indicate we're actually doing work today!
	trace.SpanFromContext(ctx).SetAttributes(attribute.Bool("skipped", false))

	return s.sources.SyncAll(ctx, handleLogger)
}
