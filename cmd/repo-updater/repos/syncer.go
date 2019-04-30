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
	duplicates := map[string][]string{}
	ctx, save := s.observe(ctx, "Syncer.Sync", strings.Join(kinds, " "))
	defer save(&duplicates, &diff, &err)

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
	args := StoreListReposArgs{Kinds: kinds, Deleted: true}
	if stored, err = store.ListRepos(ctx, args); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.store.list-repos")
	}

	seenStored := make(map[api.ExternalRepoSpec]bool, len(stored))
	for _, repo := range stored {
		if seenStored[repo.ExternalRepo] {
			log15.Error("syncer.sync.stored.duplicate: detected duplicate stored repo (bug)", "repo", repo.Name)
			duplicates["stored"] = append(duplicates["stored"], repo.Name)
		}
		seenStored[repo.ExternalRepo] = true
	}

	seenSourced := make(map[api.ExternalRepoSpec]bool, len(sourced))
	for _, repo := range sourced {
		if seenSourced[repo.ExternalRepo] {
			log15.Error("syncer.sync.sourced.duplicate: detected duplicate sourced repo (bug)", "repo", repo.Name)
			duplicates["sourced"] = append(duplicates["sourced"], repo.Name)
		}
		seenSourced[repo.ExternalRepo] = true
	}

	diff = NewDiff(sourced, stored)
	upserts := s.upserts(diff, &duplicates)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return Diff{}, errors.Wrap(err, "syncer.sync.store.upsert-repos")
	}

	if s.diffs != nil {
		s.diffs <- diff
	}

	return diff, nil
}

func (s *Syncer) upserts(diff Diff, duplicates *map[string][]string) (result []*Repo) {
	now := s.now()
	upserts := make([]*Repo, 0, len(diff.Added)+len(diff.Deleted)+len(diff.Modified))

	seen := make(map[api.ExternalRepoSpec]bool, len(diff.Deleted))
	for _, repo := range diff.Deleted {
		// TODO: This is dumb mitigation for https://github.com/sourcegraph/sourcegraph/issues/3680
		// just so that one bad apple doesn't ruin the bunch (i.e., even with
		// this bug other repositories are still updating as expected).
		// Once that issue is resolved, we should remove this.
		if seen[repo.ExternalRepo] {
			(*duplicates)["upserts_deleted"] = append((*duplicates)["upserts_deleted"], repo.Name)
			log15.Error("syncer.sync.upserts.deleted.duplicate: ignoring duplicate Deleted repo (bug)", "repo", repo.Name)
			continue
		}
		seen[repo.ExternalRepo] = true

		repo.UpdatedAt, repo.DeletedAt = now, now
		repo.Sources = map[string]*SourceInfo{}
		repo.Enabled = true
		upserts = append(upserts, repo)
	}

	seen = make(map[api.ExternalRepoSpec]bool, len(diff.Modified))
	for _, repo := range diff.Modified {
		// TODO: This is dumb mitigation for https://github.com/sourcegraph/sourcegraph/issues/3680
		// just so that one bad apple doesn't ruin the bunch (i.e., even with
		// this bug other repositories are still updating as expected).
		// Once that issue is resolved, we should remove this.
		if seen[repo.ExternalRepo] {
			(*duplicates)["upserts_modified"] = append((*duplicates)["upserts_modified"], repo.Name)
			log15.Error("syncer.sync.upserts.modified.duplicate: ignoring duplicate Modified repo (bug)", "repo", repo.Name)
			continue
		}
		seen[repo.ExternalRepo] = true

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

func (s *Syncer) observe(ctx context.Context, family, title string) (context.Context, func(*map[string][]string, *Diff, *error)) {
	began := s.now()
	tr, ctx := trace.New(ctx, family, title)

	return ctx, func(duplicates *map[string][]string, d *Diff, err *error) {
		now := s.now()
		took := s.now().Sub(began).Seconds()

		fields := make([]otlog.Field, 0, 7)

		// Note: using otlog.String here because I (@slimsag) believe Jaeger
		// does not accept non-basic data types like otlog.Object with lists of
		// strings? The "added.repos" etc fields below don't show up in Jaeger.
		// Either that or those fields are too large for Jaeger's 65536 UDP
		// packet limit (but I think that would drop the entire span so that
		// seems less likely to me).
		for category, duplicates := range *duplicates {
			fields = append(fields, otlog.String(category, fmt.Sprintf("%q", duplicates)))
		}

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
