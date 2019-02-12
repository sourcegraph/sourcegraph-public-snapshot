package repos

import (
	"context"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	interval time.Duration
	source   Source
	store    Store
	now      func() time.Time
}

// NewSyncer returns a new Syncer with the given parameters.
func NewSyncer(interval time.Duration, store Store, sources []Source, now func() time.Time) *Syncer {
	return &Syncer{
		interval: interval,
		source:   NewSources(sources...),
		store:    store,
		now:      now,
	}
}

// Run runs the Sync at its specified interval.
func (s Syncer) Run(ctx context.Context) error {
	ticks := time.NewTicker(s.interval)
	defer ticks.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticks.C:
			if err := s.Sync(ctx); err != nil {
				log15.Error("Syncer", "err", err)
			}
		}
	}
}

// Sync synchronizes the set of sourced repos with the set of stored repos.
func (s Syncer) Sync(ctx context.Context) (err error) {
	var sourced []*Repo
	if sourced, err = s.source.ListRepos(ctx); err != nil {
		return err
	}

	store := s.store
	if tr, ok := s.store.(Transactor); ok {
		var txs TxStore
		if txs, err = tr.Transact(ctx); err != nil {
			return err
		}
		defer txs.Done(&err)
		store = txs
	}

	var stored []*Repo
	if stored, err = store.ListRepos(ctx); err != nil {
		return err
	}

	return store.UpsertRepos(
		ctx,
		s.upserts(sourced, stored)...,
	)
	// TODO(tsenart): ensure scheduler picks up changes to be propagated to git server
	// TODO(tsenart): ensure search index gets updated too
}

func (s Syncer) upserts(sourced, stored []*Repo) []*Repo {
	now := s.now()
	diff := s.diff(sourced, stored)
	upserts := make([]*Repo, 0, len(diff.Added)+len(diff.Deleted)+len(diff.Modified))

	for _, add := range diff.Added {
		repo := add.(*Repo)
		repo.CreatedAt, repo.UpdatedAt = now, now
		upserts = append(upserts, repo)
	}

	for _, mod := range diff.Modified {
		repo := mod.(*Repo)
		repo.UpdatedAt, repo.DeletedAt = now, time.Time{}
		upserts = append(upserts, repo)
	}

	for _, del := range diff.Deleted {
		repo := del.(*Repo)
		repo.UpdatedAt, repo.DeletedAt = now, now
		upserts = append(upserts, repo)
	}

	return upserts
}

func (Syncer) diff(sourced, stored []*Repo) Diff {
	before := make([]Diffable, len(stored))
	for i := range stored {
		before[i] = stored[i]
	}

	after := make([]Diffable, len(sourced))
	for i := range sourced {
		after[i] = sourced[i]
	}

	return NewDiff(before, after, func(before, after Diffable) bool {
		// This modified function returns true iff any fields in `after` changed
		// in comparison to `before` for which the `Source` is authoritative.
		b, a := before.(*Repo), after.(*Repo)
		return b.Name != a.Name ||
			b.Language != a.Language ||
			b.Fork != a.Fork ||
			b.Archived != a.Archived ||
			b.Description != a.Description
	})
}
