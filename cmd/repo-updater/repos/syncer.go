package repos

import (
	"context"
	"time"

	"github.com/k0kubun/pp"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A Syncer periodically synchronizes available repositories from all its given Sources
// with the stored Repositories in Sourcegraph.
type Syncer struct {
	interval time.Duration
	src      *Sourcerer
	store    Store
	kinds    []string
	synced   chan *SyncResult
	now      func() time.Time
}

// NewSyncer returns a new Syncer that periodically synchronizes stored repos with
// the repos yielded by the given kinds of sources (code hosts), retrieved by the
// given sourcerer from the frontend API. Each completed sync results in a diff that
// is sent to the given diffs channel.
func NewSyncer(
	interval time.Duration,
	store Store,
	src *Sourcerer,
	kinds []string,
	synced chan *SyncResult,
	now func() time.Time,
) *Syncer {
	return &Syncer{
		interval: interval,
		src:      src,
		store:    store,
		synced:   synced,
		now:      now,
	}
}

// Run runs the Sync at its specified interval.
func (s Syncer) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		if _, err := s.run(ctx); err != nil {
			log15.Error("Syncer", "err", err)
		}
		time.Sleep(s.interval)
	}

	return ctx.Err()
}

func (s Syncer) run(ctx context.Context) ([]*SyncResult, error) {
	sources, err := s.src.ListSources(ctx, s.kinds...)
	if err != nil {
		return nil, err
	}
	return s.SyncMany(ctx, sources...), nil
}

// SyncResult is returned by Sync to indicate the results of syncing a source.
type SyncResult struct {
	Source Source
	Diff   Diff
	Err    error
}

// SyncMany synchonizes the repos yielded by all the given Sources.
// It returns a SyncResults containing which repos were synced and which failed to.
func (s *Syncer) SyncMany(ctx context.Context, srcs ...Source) []*SyncResult {
	if len(srcs) == 0 {
		return nil
	}

	ch := make(chan *SyncResult, len(srcs))
	for _, src := range srcs {
		go func(src Source) {
			ch <- s.Sync(ctx, src)
		}(src)
	}

	results := make([]*SyncResult, 0, len(srcs))
	for i := 0; i < cap(ch); i++ {
		res := <-ch
		results = append(results, res)
	}

	return results
}

// Sync synchronizes the repositories of a single Source
func (s Syncer) Sync(ctx context.Context, src Source) *SyncResult {
	// TODO(tsenart): Ensure that transient failures do not remove
	// repositories. This means we need to use the store as a fallback Source
	// in the face of those kinds of errors, so that the diff results in Unmodified
	// entries. This logic can live here. We only need to make the returned error
	// more structured so we can identify which sources failed and for what reason.
	// See the SyncError type defined in other_external_services.go for inspiration.

	var (
		sourced Repos
		err     error
	)

	if sourced, err = src.ListRepos(ctx); err != nil {
		return &SyncResult{Source: src, Err: err}
	}

	store := s.store
	if tr, ok := s.store.(Transactor); ok {
		var txs TxStore
		if txs, err = tr.Transact(ctx); err != nil {
			return &SyncResult{Source: src, Err: err}
		}
		defer txs.Done(&err)
		store = txs
	}

	// TODO(tsenart): We want to list all stored repos of a given source / external service.
	// This requires us to changed what we store in the external_service_id column from a
	// URL to an ID of the corresponding row in the external_services table.
	// This will require a migration, so it doesn't currently work.
	var stored Repos
	if stored, err = store.ListRepos(ctx, src.URN()); err != nil {
		return &SyncResult{Source: src, Err: err}
	}

	diff := s.diff(sourced, stored)
	upserts := s.upserts(diff)

	if err = store.UpsertRepos(ctx, upserts...); err != nil {
		return &SyncResult{Source: src, Err: err}
	}

	res := &SyncResult{Source: src, Diff: diff}
	s.synced <- res

	return res
}

func (s Syncer) upserts(diff Diff) []*Repo {
	now := s.now()
	upserts := make([]*Repo, 0, len(diff.Added)+len(diff.Deleted)+len(diff.Modified))

	pp.Printf("diff: %s", diff)

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

	pp.Printf("\nupserts: %s", upserts)

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
