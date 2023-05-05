package actor

import (
	"context"
	"time"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"

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
func (s Sources) Worker(logger log.Logger, rootInterval time.Duration) goroutine.BackgroundRoutine {
	return &sourcesWorker{
		logger:  logger.Scoped("sources", "sources worker"),
		sources: s,
		ticker:  time.NewTicker(rootInterval),
		done:    make(chan struct{}),
	}
}

type sourcesWorker struct {
	logger  log.Logger
	sources Sources
	ticker  *time.Ticker
	done    chan struct{}
}

var _ goroutine.BackgroundRoutine = &sourcesWorker{}

func (s *sourcesWorker) Start() {
	for {
		select {
		case <-s.ticker.C:
			s.sync()
		case <-s.done:
			return
		}
	}
}

func (s *sourcesWorker) sync() {
	g := pool.New().WithErrors()
	for _, src := range s.sources {
		if src, ok := src.(SourceSyncer); ok {
			g.Go(func() error {
				if err := src.Sync(context.Background()); err != nil {
					return errors.Wrap(err, src.Name())
				}
				return nil
			})
		}
	}
	if err := g.Wait(); err != nil {
		s.logger.Error("some sources failed to sync", log.Error(err))
	}
}

func (s *sourcesWorker) Stop() {
	s.ticker.Stop()
	close(s.done)
}
