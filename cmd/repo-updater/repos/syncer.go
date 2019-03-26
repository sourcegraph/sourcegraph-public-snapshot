package repos

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/k0kubun/pp"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	store   Store
	sourcer Sourcer
	diffs   chan Diff
	now     func() time.Time
}

// NewSyncer returns a new Syncer that syncs stored repos with
// the repos yielded by the configured sources, retrieved by the given sourcer.
// Each completed sync results in a diff that is sent to the given diffs channel.
func NewSyncer(
	store Store,
	sourcer Sourcer,
	diffs chan Diff,
	now func() time.Time,
) *Syncer {
	return &Syncer{
		store:   store,
		sourcer: sourcer,
		diffs:   diffs,
		now:     now,
	}
}

// Run runs the Sync at the specified interval for the given external service kinds.
func (s *Syncer) Run(ctx context.Context, interval time.Duration, kinds ...string) error {
	for ctx.Err() == nil {
		if _, err := s.Sync(ctx, kinds...); err != nil {
			log15.Error("Syncer", "error", err)
		}
		time.Sleep(interval)
	}

	return ctx.Err()
}

// Sync synchronizes the repositories of the given external service kinds.
func (s *Syncer) Sync(ctx context.Context, kinds ...string) (_ Diff, err error) {
	var sourced Repos
	if sourced, err = s.sourced(ctx, kinds...); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.sourced")
	}

	store := s.store
	if tr, ok := s.store.(Transactor); ok {
		var txs TxStore
		if txs, err = tr.Transact(ctx); err != nil {
			return Diff{}, errors.Wrap(err, "syncer.sync.transact")
		}
		defer txs.Done(&err)
		store = txs
	}

	var stored Repos
	if stored, err = store.ListRepos(ctx, kinds...); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.store.list-repos")
	}

	diff := NewDiff(sourced, stored)
	upserts := s.upserts(diff)

	if len(diff.Added) > 0 {
		log15.Debug("syncer.sync", "diff.added", pp.Sprint(diff.Added))
	}

	if len(diff.Modified) > 0 {
		log15.Debug("syncer.sync", "diff.modified", pp.Sprint(diff.Modified))
	}

	if len(diff.Deleted) > 0 {
		log15.Debug("syncer.sync", "diff.deleted", pp.Sprint(diff.Deleted))
	}

	if len(diff.Unmodified) > 0 {
		log15.Debug("syncer.sync", "diff.unmodified", pp.Sprint(diff.Unmodified))
	}

	log15.Debug("syncer.sync", "upserts", pp.Sprint(upserts))

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.store.upsert-repos")
	}

	if s.diffs != nil {
		s.diffs <- diff
	}

	return diff, nil
}

func (s *Syncer) upserts(diff Diff) []*Repo {
	now := s.now()
	upserts := make([]*Repo, 0, len(diff.Added)+len(diff.Deleted)+len(diff.Modified))

	for _, repo := range diff.Deleted {
		repo.UpdatedAt, repo.DeletedAt = now, now
		repo.Sources = map[string]*SourceInfo{}
		repo.Enabled = false // Set for backwards compatibility
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Modified {
		repo.UpdatedAt, repo.DeletedAt = now, time.Time{}
		repo.Enabled = true // Set for backwards compatibility
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Added {
		repo.CreatedAt, repo.UpdatedAt, repo.DeletedAt = now, now, time.Time{}
		repo.Enabled = true // Set for backwards compatibility
		upserts = append(upserts, repo)
	}

	return upserts
}

// A Diff of two sets of Diffables.
type Diff struct {
	Added      Repos
	Deleted    Repos
	Modified   Repos
	Unmodified Repos
}

// Sort sorts all Diff elements by Repo.IDs.
func (d *Diff) Sort() {
	for _, ds := range []Repos{
		d.Added,
		d.Deleted,
		d.Modified,
		d.Unmodified,
	} {
		sort.Sort(ds)
	}
}

// NewDiff returns a diff from the given sourced and stored repos.
func NewDiff(sourced, stored []*Repo) (diff Diff) {
	byID := make(map[api.ExternalRepoSpec]*Repo, len(sourced))
	byName := make(map[string]*Repo, len(sourced))

	for _, r := range sourced {
		if r.ExternalRepo == (api.ExternalRepoSpec{}) {
			panic(fmt.Errorf("%s has no external repo spec", r.Name))
		} else if old := byID[r.ExternalRepo]; old != nil {
			merge(old, r)
		} else {
			byID[r.ExternalRepo], byName[r.Name] = r, r
		}
	}

	seenID := make(map[api.ExternalRepoSpec]bool, len(stored))
	seenName := make(map[string]bool, len(stored))

	for _, old := range stored {
		src := byID[old.ExternalRepo]
		if src == nil {
			src = byName[old.Name]
		}

		if src == nil {
			if !old.IsDeleted() {
				diff.Deleted = append(diff.Deleted, old)
			} else {
				diff.Unmodified = append(diff.Unmodified, old)
			}
		} else if !old.IsDeleted() {
			if old.Update(src) {
				diff.Modified = append(diff.Modified, old)
			} else {
				diff.Unmodified = append(diff.Unmodified, old)
			}
		} else {
			old.Update(src)
			diff.Added = append(diff.Added, old)
		}

		seenID[old.ExternalRepo] = true
		seenName[old.Name] = true
	}

	for _, r := range byID {
		if !seenID[r.ExternalRepo] && !seenName[r.Name] {
			diff.Added = append(diff.Added, r)
		}
	}

	return diff
}

func merge(o, n *Repo) {
	for id, src := range o.Sources {
		n.Sources[id] = src
	}
	o.Update(n)
}

func (s *Syncer) sourced(ctx context.Context, kinds ...string) ([]*Repo, error) {
	svcs, err := s.store.ListExternalServices(ctx, kinds...)
	if err != nil {
		return nil, err
	}

	srcs, err := s.sourcer(svcs...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return srcs.ListRepos(ctx)
}
