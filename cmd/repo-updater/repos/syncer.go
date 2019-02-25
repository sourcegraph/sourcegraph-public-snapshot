package repos

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	interval time.Duration
	store    Store
	sourcer  Sourcer
	diffs    chan Diff
	now      func() time.Time
}

// NewSyncer returns a new Syncer that periodically synchronizes stored repos with
// the repos yielded by the configured sources, retrieved by the given sourcer.
// Each completed sync results in a diff that is sent to the given diffs channel.
func NewSyncer(
	interval time.Duration,
	store Store,
	sourcer Sourcer,
	diffs chan Diff,
	now func() time.Time,
) *Syncer {
	return &Syncer{
		interval: interval,
		store:    store,
		sourcer:  sourcer,
		diffs:    diffs,
		now:      now,
	}
}

// Run runs the Sync at its specified interval.
func (s *Syncer) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		if _, err := s.Sync(ctx); err != nil {
			log15.Error("Syncer", "err", err)
		}
		time.Sleep(s.interval)
	}

	return ctx.Err()
}

// Sync synchronizes the repositories of a single Source
func (s *Syncer) Sync(ctx context.Context) (_ Diff, err error) {
	// TODO(tsenart): Ensure that transient failures do not remove
	// repositories. This means we need to use the store as a fallback Source
	// in the face of those kinds of errors, so that the diff results in Unmodified
	// entries. This logic can live here. We only need to make the returned error
	// more structured so we can identify which sources failed and for what reason.
	// See the SyncError type defined in other_external_services.go for inspiration.

	var sourced Repos
	if sourced, err = s.sourced(ctx); err != nil {
		return Diff{}, err
	}

	store := s.store
	if tr, ok := s.store.(Transactor); ok {
		var txs TxStore
		if txs, err = tr.Transact(ctx); err != nil {
			return Diff{}, err
		}
		defer txs.Done(&err)
		store = txs
	}

	var stored Repos
	if stored, err = store.ListRepos(ctx, sourced.Names()...); err != nil {
		return Diff{}, err
	}

	diff := NewDiff(sourced, stored)
	upserts := s.upserts(diff)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return Diff{}, err
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
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Modified {
		repo.UpdatedAt, repo.DeletedAt = now, time.Time{}
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Added {
		repo.CreatedAt, repo.DeletedAt = now, time.Time{}
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
			diff.Deleted = append(diff.Deleted, old)
		} else if upsert(old, src) {
			diff.Modified = append(diff.Modified, old)
		} else {
			diff.Unmodified = append(diff.Unmodified, old)
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
	upsert(o, n)
}

func upsert(o, n *Repo) (modified bool) {
	if !o.ExternalRepo.Equal(&n.ExternalRepo) && o.Name != n.Name {
		panic(fmt.Errorf("merge called with distinct repos: older: %+v, newer: %+v", o, n))
	}

	if o.Name != n.Name {
		o.Name, modified = n.Name, true
	}

	if o.Description != n.Description {
		o.Description, modified = n.Description, true
	}

	if o.Language != n.Language {
		o.Language, modified = n.Language, true
	}

	if !o.ExternalRepo.Equal(&n.ExternalRepo) {
		o.ExternalRepo, modified = n.ExternalRepo, true
	}

	if o.Archived != n.Archived {
		o.Archived, modified = n.Archived, true
	}

	if o.Fork != n.Fork {
		o.Fork, modified = n.Fork, true
	}

	if !reflect.DeepEqual(o.Sources, n.Sources) {
		o.Sources, modified = n.Sources, true
	}

	if !reflect.DeepEqual(o.Metadata, n.Metadata) {
		o.Metadata, modified = n.Metadata, true
	}

	return modified
}

func (s *Syncer) sourced(ctx context.Context) ([]*Repo, error) {
	sources, err := s.sourcer.ListSources(ctx)
	if err != nil {
		return nil, err
	}

	type result struct {
		src   Source
		repos []*Repo
		err   error
	}

	ch := make(chan result, len(sources))
	for _, src := range sources {
		go func(src Source) {
			if repos, err := src.ListRepos(ctx); err != nil {
				ch <- result{src: src, err: err}
			} else {
				ch <- result{src: src, repos: repos}
			}
		}(src)
	}

	var repos []*Repo
	var errs *multierror.Error

	for i := 0; i < cap(ch); i++ {
		if r := <-ch; r.err != nil {
			errs = multierror.Append(errs, r.err)
		} else {
			repos = append(repos, r.repos...)
		}
	}

	return repos, errs.ErrorOrNil()
}
