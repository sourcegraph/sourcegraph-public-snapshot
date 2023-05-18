package actor

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var tracer = otel.GetTracerProvider().Tracer("llm-proxy/internal/actor")

type ErrNotFromSource struct{}

func (ErrNotFromSource) Error() string { return "token not from source" }

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

type Sources []Source

func (s Sources) Get(ctx context.Context, token string) (_ *Actor, err error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "Sources.Get")
	defer func() {
		span.RecordError(err) // don't set status, not necessarily a hard failure
		span.End()
	}()

	for _, src := range s {
		actor, err := src.Get(ctx, token)
		// Only if the Source indicates it doesn't know about this token do
		// we continue to the next Source.
		if err != nil && errors.Is(err, ErrNotFromSource{}) {
			continue
		}

		// Otherwise we continue with the first result we get.
		span.SetAttributes(attribute.String("matched_source", src.Name()))
		return actor, errors.Wrap(err, src.Name())
	}

	return nil, errors.New("no source found for token")
}

// Worker is a goroutine.BackgroundRoutine that runs any SourceSyncer implementations
// at a regular interval. It uses a redsync.Mutex to ensure only one worker is running
// at a time.
func (s Sources) Worker(logger log.Logger, rmux *redsync.Mutex, rootInterval time.Duration) goroutine.BackgroundRoutine {
	return &redisLockedBackgroundRoutine{
		logger: logger.Scoped("Sources.Worker", ""),
		rmux:   rmux,
		routine: goroutine.NewPeriodicGoroutine(
			context.Background(),
			"sources", "sources sync worker",
			rootInterval,
			&sourcesPeriodicHandler{
				logger:  logger.Scoped("sourcesPeriodicHandler", ""),
				rmux:    rmux,
				sources: s,
			}),
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
		_, err := s.rmux.Unlock()
		if err != nil {
			s.logger.Warn("Failed to unlock mutex after work completed", log.Error(err))
		}
	}
}

// sourcesPeriodicHandler is a handler for NewPeriodicGoroutine
type sourcesPeriodicHandler struct {
	logger  log.Logger
	rmux    *redsync.Mutex
	sources Sources
}

var _ goroutine.Handler = &sourcesPeriodicHandler{}

func (s *sourcesPeriodicHandler) Handle(ctx context.Context) (err error) {
	// If we are not holding a lock, try to acquire it.
	if expire := s.rmux.Until(); expire.IsZero() {
		// If another instance is working on background syncs, we don't want to
		// do anything. We should check every time still in case the current worker
		// goes offline, we want to be ready to pick up the work.
		if err := s.rmux.LockContext(ctx); errors.HasType(err, &redsync.ErrTaken{}) {
			s.logger.Debug("Not starting a new sync, another one is likely in progress")
			return nil // ignore lock contention errors
		} else if err != nil {
			return errors.Wrap(err, "acquire worker lock")
		}
	} else {
		s.logger.Debug("Extending lock duration")
		// Otherwise, extend our lock so that we can keep working.
		_, _ = s.rmux.ExtendContext(ctx)
	}

	p := pool.New().WithErrors().WithContext(ctx)
	for _, src := range s.sources {
		if src, ok := src.(SourceSyncer); ok {
			p.Go(func(ctx context.Context) error {
				var span trace.Span
				ctx, span = tracer.Start(ctx, "sourcesPeriodicHandler.Handle",
					trace.WithAttributes(attribute.String("source", src.Name())))
				defer func() {
					if err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, "sync failed")
					}
					span.End()
				}()

				logger := s.logger.
					With(log.String("syncer", fmt.Sprintf("%T", src))).
					WithTrace(log.TraceContext{
						TraceID: span.SpanContext().TraceID().String(),
						SpanID:  span.SpanContext().SpanID().String(),
					})

				start := time.Now()

				logger.Info("Starting a new sync")
				seen, err := src.Sync(ctx)
				if err != nil {
					logger.Error("Failed sync", log.Error(err))
					return errors.Wrapf(err, "failed to sync %s", src.Name())
				}
				logger.Info("Completed sync", log.Duration("sync_duration", time.Since(start)), log.Int("seen", seen))
				return nil
			})
		}
	}
	return p.Wait()
}
