package actor

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/sourcegraph/conc/pool"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
	Sync(ctx context.Context) error
}

type Sources []Source

func (s Sources) Get(ctx context.Context, token string) (*Actor, error) {
	for _, src := range s {
		actor, err := src.Get(ctx, token)
		// Only if the Source indicates it doesn't know about this token do
		// we continue to the next Source.
		if err != nil && errors.Is(err, ErrNotFromSource{}) {
			continue
		}

		// Otherwise we continue with the first result we get.
		return actor, errors.Wrap(err, src.Name())
	}
	return nil, errors.New("no source found for token")
}

// Worker is a goroutine.BackgroundRoutine that runs any SourceSyncer implementations
// at a regular interval. It uses a redsync.Mutex to ensure only one worker is running
// at a time.
func (s Sources) Worker(rmux *redsync.Mutex, rootInterval time.Duration) goroutine.BackgroundRoutine {
	return &redisLockedBackgroundRoutine{
		rmux: rmux,
		routine: goroutine.NewPeriodicGoroutine(
			context.Background(),
			"sources", "sources sync worker",
			rootInterval,
			&sourcesPeriodicHandler{
				rmux:    rmux,
				sources: s,
			}),
	}
}

// redisLockedBackgroundRoutine attempts to acquire a redsync lock before starting,
// and releases it when stopped.
type redisLockedBackgroundRoutine struct {
	rmux    *redsync.Mutex
	routine goroutine.BackgroundRoutine
}

func (s *redisLockedBackgroundRoutine) Start() {
	// Best-effort attempt to acquire lock immediately.
	// We check if we have the lock first because in tests we may manually acquire
	// it first to keep tests stable.
	if expire := s.rmux.Until(); expire.IsZero() {
		_ = s.rmux.LockContext(context.Background())
	}

	s.routine.Start()
}

func (s *redisLockedBackgroundRoutine) Stop() {
	s.routine.Stop()

	// If we have the lock, release it and let somebody else work
	if expire := s.rmux.Until(); !expire.IsZero() {
		s.rmux.Unlock()
	}
}

// sourcesPeriodicHandler is a handler for NewPeriodicGoroutine
type sourcesPeriodicHandler struct {
	rmux    *redsync.Mutex
	sources Sources
}

var _ goroutine.Handler = &sourcesPeriodicHandler{}

func (s *sourcesPeriodicHandler) Handle(ctx context.Context) error {
	// If we are not holding a lock, try to acquire it.
	if expire := s.rmux.Until(); expire.IsZero() {
		// If another instance is working on background syncs, we don't want to
		// do anything. We should check every time still in case the current worker
		// goes offline, we want to be ready to pick up the work.
		if err := s.rmux.LockContext(ctx); errors.Is(err, redsync.ErrFailed) {
			return nil // ignore lock contention errors
		} else if err != nil {
			return errors.Wrap(err, "acquire worker lock")
		}
	} else {
		// Otherwise, extend our lock so that we can keep working.
		_, _ = s.rmux.ExtendContext(ctx)
	}

	p := pool.New().WithErrors().WithContext(ctx)
	for _, src := range s.sources {
		if src, ok := src.(SourceSyncer); ok {
			p.Go(func(ctx context.Context) error {
				if err := src.Sync(ctx); err != nil {
					return errors.Wrap(err, src.Name())
				}
				return nil
			})
		}
	}
	return p.Wait()
}
