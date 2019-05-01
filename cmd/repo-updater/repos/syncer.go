package repos

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
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
func (s *Syncer) Sync(ctx context.Context, kinds ...string) (diff Diff, err error) {
	ctx, save := s.observe(ctx, "Syncer.Sync", strings.Join(kinds, " "))
	defer save(&diff, &err)

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
	args := StoreListReposArgs{Kinds: kinds}
	if stored, err = store.ListRepos(ctx, args); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.store.list-repos")
	}

	diff = NewDiff(sourced, stored)
	upserts := s.upserts(diff)

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
		repo.Enabled = true
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Modified {
		repo.UpdatedAt, repo.DeletedAt = now, time.Time{}
		repo.Enabled = true
		upserts = append(upserts, repo)
	}

	for _, repo := range diff.Added {
		repo.CreatedAt, repo.UpdatedAt, repo.DeletedAt = now, now, time.Time{}
		repo.Enabled = true
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

// Repos returns all repos in the Diff.
func (d Diff) Repos() Repos {
	all := make(Repos, 0, len(d.Added)+
		len(d.Deleted)+
		len(d.Modified)+
		len(d.Unmodified))

	for _, rs := range []Repos{
		d.Added,
		d.Deleted,
		d.Modified,
		d.Unmodified,
	} {
		all = append(all, rs...)
	}

	return all
}

// NewDiff returns a diff from the given sourced and stored repos.
func NewDiff(sourced, stored []*Repo) (diff Diff) {
	byID := make(map[api.ExternalRepoSpec]*Repo, len(sourced))
	for _, r := range sourced {
		if r.ExternalRepo == (api.ExternalRepoSpec{}) {
			panic(fmt.Errorf("%s has no external repo spec", r.Name))
		} else if old := byID[r.ExternalRepo]; old != nil {
			merge(old, r)
		} else {
			byID[r.ExternalRepo] = r
		}
	}

	byName := make(map[string]*Repo, len(byID))
	for _, r := range byID {
		byName[r.Name] = r
	}

	seenID := make(map[api.ExternalRepoSpec]bool, len(stored))
	seenName := make(map[string]bool, len(stored))

	// We are unsure if customer repositories can have ExternalRepo unset. We
	// know it can be unset for Sourcegraph.com. As such, we want to fallback
	// to associating stored repositories by name with the sourced
	// repositories.
	//
	// We do not want a stored repository without an externalrepo to be set
	sort.Stable(byExternalRepoSpecSet(stored))

	for _, old := range stored {
		src := byID[old.ExternalRepo]
		if src == nil && old.ExternalRepo == (api.ExternalRepoSpec{}) && !seenName[old.Name] {
			src = byName[old.Name]
		}

		if src == nil {
			diff.Deleted = append(diff.Deleted, old)
		} else if old.Update(src) {
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
	o.Update(n)
}

func (s *Syncer) sourced(ctx context.Context, kinds ...string) ([]*Repo, error) {
	svcs, err := s.store.ListExternalServices(ctx, StoreListExternalServicesArgs{
		Kinds: kinds,
	})

	if err != nil {
		return nil, err
	}

	srcs, err := s.sourcer(svcs...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, sourceTimeout)
	defer cancel()

	return srcs.ListRepos(ctx)
}

func (s *Syncer) observe(ctx context.Context, family, title string) (context.Context, func(*Diff, *error)) {
	began := s.now()
	tr, ctx := trace.New(ctx, family, title)

	return ctx, func(d *Diff, err *error) {
		now := s.now()
		took := s.now().Sub(began).Seconds()

		fields := make([]otlog.Field, 0, 7)
		for state, repos := range map[string]Repos{
			"added":      d.Added,
			"modified":   d.Modified,
			"deleted":    d.Deleted,
			"unmodified": d.Unmodified,
		} {
			fields = append(fields, otlog.Int(state+".count", len(repos)))
			if state != "unmodified" {
				fields = append(fields,
					otlog.Object(state+".repos", repos.Names()))
			}
			syncedTotal.WithLabelValues(state).Add(float64(len(repos)))
		}

		tr.LogFields(fields...)

		lastSync.WithLabelValues().Set(float64(now.Unix()))

		success := err == nil || *err == nil
		syncDuration.WithLabelValues(strconv.FormatBool(success)).Observe(took)

		if !success {
			tr.SetError(*err)
			syncErrors.WithLabelValues().Add(1)
		}

		tr.Finish()
	}
}

type byExternalRepoSpecSet []*Repo

func (rs byExternalRepoSpecSet) Len() int      { return len(rs) }
func (rs byExternalRepoSpecSet) Swap(i, j int) { rs[i], rs[j] = rs[j], rs[i] }
func (rs byExternalRepoSpecSet) Less(i, j int) bool {
	iSet := rs[i].ExternalRepo != (api.ExternalRepoSpec{})
	jSet := rs[j].ExternalRepo != (api.ExternalRepoSpec{})
	if iSet == jSet {
		return false
	}
	return iSet
}
