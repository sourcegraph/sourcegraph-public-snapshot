package backend

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"

	"github.com/golang/groupcache/lru"
	"github.com/rogpeppe/rog-go/parallel"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-diff/diff"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/cache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

var Deltas sourcegraph.DeltasServer = &deltas{
	cache:          newDeltasCache(1e4), // ~1KB per gob encoded delta
	listFilesCache: newDeltasListFilesCache(1e4, 10*1024),
}

type deltas struct {
	// mockDiffFunc, if set, is called by (deltas).diff instead of the
	// main method body. It allows mocking (deltas).diff in tests.
	mockDiffFunc func(context.Context, sourcegraph.DeltaSpec) ([]*diff.FileDiff, *sourcegraph.Delta, error)

	// cache caches get delta requests, it does not cache results from
	// requests that return a non-nil error.
	cache *deltasCache

	// listFilesCache caches requests to list delta files, it does not cache
	// results from requests that return a non-nil error or diffs larger
	// than a certain size.
	listFilesCache *deltasListFileCache
}

var _ sourcegraph.DeltasServer = (*deltas)(nil)

func (s *deltas) Get(ctx context.Context, ds *sourcegraph.DeltaSpec) (*sourcegraph.Delta, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Deltas.Get", ds.Base.Repo); err != nil {
		return nil, err
	}
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "Deltas.Get", ds.Head.Repo); err != nil {
		return nil, err
	}

	if s.cache != nil {
		hit, ok := s.cache.Get(ds)
		if ok {
			return hit, nil
		}
	}

	d, err := s.fillDelta(ctx, &sourcegraph.Delta{Base: ds.Base, Head: ds.Head})
	if err != nil {
		return d, err
	}

	if s.cache != nil {
		s.cache.Add(ds, d)
	}
	return d, nil
}

func (s *deltas) fillDelta(ctx context.Context, d *sourcegraph.Delta) (*sourcegraph.Delta, error) {
	if d.Base.Repo != d.Head.Repo {
		return d, errors.New("base and head repo must be identical")
	}

	getCommit := func(repoRevSpec *sourcegraph.RepoRevSpec, commit **vcs.Commit) error {
		var err error
		*commit, err = svc.Repos(ctx).GetCommit(ctx, repoRevSpec)
		if err != nil {
			return err
		}
		repoRevSpec.CommitID = string((*commit).ID)
		return nil
	}

	par := parallel.NewRun(2)
	if d.BaseCommit == nil {
		par.Do(func() error { return getCommit(&d.Base, &d.BaseCommit) })
	}
	if d.HeadCommit == nil {
		par.Do(func() error { return getCommit(&d.Head, &d.HeadCommit) })
	}
	if err := par.Wait(); err != nil {
		return d, err
	}

	// Try to compute merge-base.
	vcsBaseRepo, err := store.RepoVCSFromContext(ctx).Open(ctx, d.Base.Repo)
	if err != nil {
		return d, err
	}

	id, err := vcsBaseRepo.MergeBase(vcs.CommitID(d.BaseCommit.ID), vcs.CommitID(d.HeadCommit.ID))
	if err != nil {
		return d, err
	}

	if d.BaseCommit.ID != id {
		// There is most likely a merge conflict here, so we update the
		// delta to contain the actual merge base used in this diff A...B
		d.Base.CommitID = string(id)
		d.BaseCommit = nil
		d, err = s.fillDelta(ctx, d)
		if err != nil {
			return d, err
		}
	}
	return d, nil
}

type deltasCache struct {
	cache.Cache
}

func newDeltasCache(maxEntries int) *deltasCache {
	return &deltasCache{cache.Sync(lru.New(maxEntries))}
}

func deltasCacheKey(spec *sourcegraph.DeltaSpec) string {
	return spec.Base.CommitID + ".." + spec.Head.CommitID
}

func (c *deltasCache) Add(spec *sourcegraph.DeltaSpec, delta *sourcegraph.Delta) {
	if spec.Base.CommitID == "" || spec.Head.CommitID == "" {
		return
	}
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(delta); err != nil {
		log.Println("error while encoding delta:", err.Error())
		return
	}
	c.Cache.Add(deltasCacheKey(spec), buf.Bytes())
}

func (c *deltasCache) Get(spec *sourcegraph.DeltaSpec) (*sourcegraph.Delta, bool) {
	if spec.Base.CommitID == "" || spec.Head.CommitID == "" {
		return nil, false
	}
	obj, ok := c.Cache.Get(deltasCacheKey(spec))
	if !ok {
		return nil, false
	}
	deltaBytes, isBytes := obj.([]byte)
	if !isBytes {
		return nil, false
	}
	var copy *sourcegraph.Delta
	dec := gob.NewDecoder(bytes.NewReader(deltaBytes))
	if err := dec.Decode(&copy); err != nil {
		log.Println("error while decoding delta:", err.Error())
		return nil, false
	}
	return copy, true
}
