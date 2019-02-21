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
func (s Syncer) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		if _, err := s.Sync(ctx); err != nil {
			log15.Error("Syncer", "err", err)
		}
		time.Sleep(s.interval)
	}

	return ctx.Err()
}

// Sync synchronizes the repositories of a single Source
func (s Syncer) Sync(ctx context.Context) (_ Diff, err error) {
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

	diff := s.diff(sourced, stored)
	upserts := s.upserts(diff)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return Diff{}, err
	}

	if s.diffs != nil {
		s.diffs <- diff
	}

	return diff, nil
}

func (s Syncer) upserts(diff Diff) []*Repo {
	now := s.now()
	upserts := make([]*Repo, 0, len(diff.Added)+len(diff.Deleted)+len(diff.Modified))

	for _, repo := range diff.Deleted {
		repo.UpdatedAt, repo.DeletedAt = now, now
		repo.Sources = []string{}
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
	Added      []*Repo
	Deleted    []*Repo
	Modified   []*Repo
	Unmodified []*Repo
}

// Sort sorts all Diff elements by Repo.IDs.
func (d *Diff) Sort() {
	for _, ds := range [][]*Repo{
		d.Added,
		d.Deleted,
		d.Modified,
		d.Unmodified,
	} {
		sort.Slice(ds, func(i, j int) bool {
			l, r := ds[i].IDs(), ds[j].IDs()
			return l[0] < r[0]
		})
	}
}

func (Syncer) diff(sourced, stored []*Repo) (diff Diff) {
	byID := make(map[api.ExternalRepoSpec]*Repo, len(sourced))
	byName := make(map[string]*Repo, len(sourced))

	for _, r := range stored {
		byName[r.Name] = r
		if r.ExternalRepo != (api.ExternalRepoSpec{}) {
			byID[r.ExternalRepo] = r
		}
	}

	seen := make(map[string]*Repo, len(stored))
	for _, r := range sourced {
		if other := seen[r.Name]; other != nil {
			merge(other, r)
			continue
		}

		var old *Repo
		if r.ExternalRepo != (api.ExternalRepoSpec{}) {
			old = byID[r.ExternalRepo]
		}

		if old == nil {
			old = byName[r.Name]
		}

		if old == nil {
			seen[r.Name], diff.Added = r, append(diff.Added, r)
		} else if merge(old, r) {
			seen[r.Name], diff.Modified = old, append(diff.Modified, old)
		} else {
			seen[r.Name], diff.Unmodified = old, append(diff.Unmodified, old)
		}

	}

	for _, r := range stored {
		if seen[r.Name] == nil {
			diff.Deleted = append(diff.Deleted, r)
		}
	}

	return diff
}

// merge merges the newer Repo "n" into the older one "o", returning true if anything was modified.
func merge(o, n *Repo) (modified bool) {
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

	if o.Enabled != n.Enabled {
		o.Enabled, modified = n.Enabled, true
	}

	if o.Archived != n.Archived {
		o.Archived, modified = n.Archived, true
	}

	if o.Fork != n.Fork {
		o.Fork, modified = n.Fork, true
	}

	if !equal(o.Sources, n.Sources) {
		o.Sources, modified = dedup(n.Sources...), true
	}

	if !reflect.DeepEqual(o.Metadata, n.Metadata) {
		o.Metadata, modified = n.Metadata, true
	}

	return modified
}

func (s Syncer) sourced(ctx context.Context) ([]*Repo, error) {
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

func dedup(ss ...string) []string {
	uniq := make([]string, 0, len(ss))
	set := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		if _, ok := set[s]; !ok {
			set[s] = struct{}{}
			uniq = append(uniq, s)
		}
	}
	return uniq
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func set(s map[string]*Repo, d *Repo) {
	for _, id := range d.IDs() {
		s[id] = d
	}
}

func del(set map[string]*Repo, ids []string) {
	for _, id := range ids {
		delete(set, id)
	}
}

func elems(set map[string]*Repo, ids []string) []*Repo {
	ds := make([]*Repo, 0, len(ids))
	for _, id := range ids {
		if d, ok := set[id]; ok {
			ds = append(ds, d)
		}
	}
	return ds
}
