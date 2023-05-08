package actor

import (
	"context"
	"time"

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
// at a regular interval.
func (s Sources) Worker(rootInterval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"sources", "sources sync worker",
		rootInterval,
		&sourcesPeriodicHandler{sources: s})
}

type sourcesPeriodicHandler struct {
	sources Sources
}

var _ goroutine.Handler = &sourcesPeriodicHandler{}

func (s *sourcesPeriodicHandler) Handle(ctx context.Context) error {
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
