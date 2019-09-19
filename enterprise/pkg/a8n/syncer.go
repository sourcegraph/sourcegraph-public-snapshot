package a8n

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"gopkg.in/inconshreveable/log15.v2"
)

// A ChangesetSyncer periodically sync the metadata of the changesets
// saved in the database
type ChangesetSyncer struct {
	Store       *Store
	ReposStore  repos.Store
	HTTPFactory *httpcli.Factory
}

// Run runs the Sync at the specified interval.
func (s *ChangesetSyncer) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		seconds := 30 + rand.Intn(90)
		time.Sleep(time.Duration(seconds) * time.Second)

		s.syncAllChangesets(ctx)
	}

	return ctx.Err()
}

func (s *ChangesetSyncer) syncAllChangesets(ctx context.Context) {
	cs, err := s.listAllChangesets(ctx)
	if err != nil {
		log15.Error("ChangesetSyncer.listAllChangesets", "error", err)
		return
	}

	if err := s.Sync(ctx, cs...); err != nil {
		log15.Error("ChangesetSyncer", "error", err)
	}
}

// Sync refreshes the metadata of the specified changesets and updates them
// in the database
func (s *ChangesetSyncer) Sync(ctx context.Context, cs ...*a8n.Changeset) (err error) {
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

	for _, b := range batches {
		if err = b.LoadChangesets(ctx, b.Changesets...); err != nil {
			return err
		}
	}

	if err = s.Store.UpdateChangesets(ctx, cs...); err != nil {
		return err
	}

	return nil
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
