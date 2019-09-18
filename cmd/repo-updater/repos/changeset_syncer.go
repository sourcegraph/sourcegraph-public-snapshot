package repos

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A ChangesetSyncer periodically sync the metadata of the changesets
// saved in the database
type ChangesetSyncer struct {
	A8NStore    *a8n.Store
	ReposStore  Store
	HTTPFactory *httpcli.Factory
}

// Sync refreshes the metadata of the specified changesets and updates them
// in the database
func (s *ChangesetSyncer) Sync(ctx context.Context, cs ...*a8n.Changeset) (err error) {
	if len(cs) == 0 {
		return nil
	}

	var repoIDs []uint32
	repoSet := map[uint32]*Repo{}

	for _, c := range cs {
		id := uint32(c.RepoID)
		if _, ok := repoSet[id]; !ok {
			repoSet[id] = nil
			repoIDs = append(repoIDs, id)
		}
	}

	rs, err := s.ReposStore.ListRepos(ctx, StoreListReposArgs{IDs: repoIDs})
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

	es, err := s.ReposStore.ListExternalServices(ctx, StoreListExternalServicesArgs{RepoIDs: repoIDs})
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
		ChangesetSource
		Changesets []*Changeset
	}

	batches := make(map[int64]*batch, len(es))
	for _, e := range es {
		src, err := NewSource(e, s.HTTPFactory)
		if err != nil {
			return err
		}

		css, ok := src.(ChangesetSource)
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
		b.Changesets = append(b.Changesets, &Changeset{
			Changeset: c,
			Repo:      repoSet[repoID],
		})
	}

	for _, b := range batches {
		if err = b.LoadChangesets(ctx, b.Changesets...); err != nil {
			return err
		}
	}

	if err = s.A8NStore.UpdateChangesets(ctx, cs...); err != nil {
		return err
	}

	return nil
}

func (s *ChangesetSyncer) listAllChangesets(ctx context.Context) (all []*a8n.Changeset, err error) {
	for cursor := int64(-1); cursor != 0; {
		opts := a8n.ListChangesetsOpts{Cursor: cursor, Limit: 1000}
		cs, next, err := s.A8NStore.ListChangesets(ctx, opts)
		if err != nil {
			return nil, err
		}
		all, cursor = append(all, cs...), next
	}

	return all, err
}
