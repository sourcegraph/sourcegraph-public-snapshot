package a8n

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"gopkg.in/inconshreveable/log15.v2"
)

// A ChangesetSyncer periodically sync the metadata of the changesets
// saved in the database
type ChangesetSyncer struct {
	Store       *Store
	ReposStore  repos.Store
	HTTPFactory *httpcli.Factory
}

// Sync refreshes the metadata of all changesets and updates them in the
// database
func (s *ChangesetSyncer) Sync(ctx context.Context) error {
	cs, err := s.listAllChangesets(ctx)
	if err != nil {
		log15.Error("ChangesetSyncer.listAllChangesets", "error", err)
		return err
	}

	if err := s.SyncChangesets(ctx, cs...); err != nil {
		log15.Error("ChangesetSyncer", "error", err)
		return err
	}
	return nil
}

// SyncChangesets refreshes the metadata of the given changesets and
// updates them in the database
func (s *ChangesetSyncer) SyncChangesets(ctx context.Context, cs ...*a8n.Changeset) (err error) {
	if len(cs) == 0 {
		return nil
	}

	var repoIDs []uint32
	repoSet := map[uint32]*repos.Repo{}

	for _, c := range cs {
		id := uint32(c.RepoID)
		if _, ok := repoSet[id]; !ok {
			repoSet[id] = nil
			repoIDs = append(repoIDs, id)
		}
	}

	rs, err := s.ReposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
	if err != nil {
		return err
	}

	for _, r := range rs {
		repoSet[r.ID] = r
	}

	for _, c := range cs {
		repo := repoSet[uint32(c.RepoID)]
		if repo == nil {
			log15.Warn("changeset not synced, repo not in database", "changeset_id", c.ID, "repo_id", c.RepoID)
		}
	}

	es, err := s.ReposStore.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{RepoIDs: repoIDs})
	if err != nil {
		return err
	}

	byRepo := make(map[uint32]int64, len(rs))
	for _, r := range rs {
		eids := r.ExternalServiceIDs()
		for _, id := range eids {
			if _, ok := byRepo[r.ID]; !ok {
				byRepo[r.ID] = id
				break
			}
		}
	}

	type batch struct {
		repos.ChangesetSource
		Changesets []*repos.Changeset
	}

	batches := make(map[int64]*batch, len(es))
	for _, e := range es {
		src, err := repos.NewSource(e, s.HTTPFactory)
		if err != nil {
			return err
		}

		css, ok := src.(repos.ChangesetSource)
		if !ok {
			return errors.Errorf("unsupported repo type %q", e.Kind)
		}

		batches[e.ID] = &batch{ChangesetSource: css}
	}

	for _, c := range cs {
		repoID := uint32(c.RepoID)
		b := batches[byRepo[repoID]]
		if b == nil {
			continue
		}
		b.Changesets = append(b.Changesets, &repos.Changeset{
			Changeset: c,
			Repo:      repoSet[repoID],
		})
	}

	var events []*a8n.ChangesetEvent
	for _, b := range batches {
		if err = b.LoadChangesets(ctx, b.Changesets...); err != nil {
			return err
		}

		for _, c := range b.Changesets {
			events = append(events, c.Events()...)
		}
	}

	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}

	defer tx.Done(&err)

	if err = tx.UpdateChangesets(ctx, cs...); err != nil {
		return err
	}

	return tx.UpsertChangesetEvents(ctx, events...)
}

func (s *ChangesetSyncer) listAllChangesets(ctx context.Context) (all []*a8n.Changeset, err error) {
	for cursor := int64(-1); cursor != 0; {
		opts := ListChangesetsOpts{Cursor: cursor, Limit: 1000}
		cs, next, err := s.Store.ListChangesets(ctx, opts)
		if err != nil {
			return nil, err
		}
		all, cursor = append(all, cs...), next
	}

	return all, err
}
