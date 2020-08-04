package commitupdater

import (
	"context"
	"errors"
	"time"

	"github.com/efritz/glock"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commits"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
)

// Updater periodically re-calculates the commit and upload visibility graph for repositories
// that are marked as dirty by the worker process. This is done out-of-band from the rest of
// the upload processing as it is likely that we are processing multiple uploads concurrently
// for the same repository and should not repeat the work since the last calculation performed
// will always be the one we want.
type Updater struct {
	store    store.Store
	updater  commits.Updater
	options  UpdaterOptions
	clock    glock.Clock
	ctx      context.Context // root context passed to the updater
	cancel   func()          // cancels the root context
	finished chan struct{}   // signals that Start has finished
}

type UpdaterOptions struct {
	Interval time.Duration
}

func NewUpdater(store store.Store, updater commits.Updater, options UpdaterOptions) *Updater {
	return newUpdater(store, updater, options, glock.NewRealClock())
}

func newUpdater(store store.Store, updater commits.Updater, options UpdaterOptions, clock glock.Clock) *Updater {
	ctx, cancel := context.WithCancel(context.Background())

	return &Updater{
		store:    store,
		updater:  updater,
		options:  options,
		clock:    clock,
		ctx:      ctx,
		cancel:   cancel,
		finished: make(chan struct{}),
	}
}

// Start begins periodically checking for dirty repositories and invoking the underlying
// updater on each one.
func (u *Updater) Start() {
	defer close(u.finished)

loop:
	for {
		repositoryIDs, err := u.store.DirtyRepositories(u.ctx)
		if err != nil {
			log15.Error("Failed to retrieve dirty repositories", "err", err)
		}

		for repositoryID, dirtyFlag := range repositoryIDs {
			if err := u.updater.TryUpdate(context.Background(), repositoryID, dirtyFlag); err != nil {
				for ex := err; ex != nil; ex = errors.Unwrap(ex) {
					if err == u.ctx.Err() {
						break loop
					}
				}

				log15.Error("Failed to update repository commit graph", "err", err)
			}
		}

		select {
		case <-u.clock.After(u.options.Interval):
		case <-u.ctx.Done():
			return
		}
	}
}

// Stop will cause the update loop to exit after the current iteration.
func (u *Updater) Stop() {
	u.cancel()
	<-u.finished
}
