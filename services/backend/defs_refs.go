package backend

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

var Defs = &defs{}

type defs struct{}

func (s *defs) TotalRefs(ctx context.Context, source string) (res int, err error) {
	ctx, done := trace(ctx, "Defs", "TotalRefs", source, &err)
	defer done()
	return localstore.GlobalRefs.TotalRefs(ctx, source)
}

func (s *defs) RefLocations(ctx context.Context, op sourcegraph.RefLocationsOptions) (res *sourcegraph.RefLocations, err error) {
	ctx, done := trace(ctx, "Defs", "RefLocations", op, &err)
	defer done()
	return localstore.GlobalRefs.RefLocations(ctx, op)
}

var indexDuration = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Name:      "index_duration_seconds",
	Help:      "Duration of time that indexing a repository takes.",
})

func init() {
	prometheus.MustRegister(indexDuration)
}

func (s *defs) RefreshIndex(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) (err error) {
	start := time.Now()
	if Mocks.Defs.RefreshIndex != nil {
		return Mocks.Defs.RefreshIndex(ctx, op)
	}

	ctx, done := trace(ctx, "Defs", "RefreshIndex", op, &err)
	defer done()

	// Refuse to index private repositories. For the time being, we do not. We
	// must decide on an approach, and there are serious implications to both
	// approaches.
	repo, err := Repos.Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
	if err != nil {
		return err
	}
	if repo.Private {
		log15.Warn("Refusing to index private repository", "repo", repo.URI)
		return nil
	}

	// Only set the index duration gauge if the repository is not private,
	// otherwise we would skew the index duration times by quite a large
	// margin.
	defer func() {
		indexDuration.Set(time.Since(start).Seconds())
	}()

	rev, err := Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: op.Repo})
	if err != nil {
		return err
	}

	// Refresh global references indexes.
	return localstore.GlobalRefs.RefreshIndex(ctx, repo.URI, rev.CommitID)
}

type MockDefs struct {
	RefreshIndex func(v0 context.Context, v1 *sourcegraph.DefsRefreshIndexOp) error
}

func (s *MockDefs) MockRefreshIndex(t *testing.T, wantOp *sourcegraph.DefsRefreshIndexOp) (called *bool) {
	called = new(bool)
	s.RefreshIndex = func(ctx context.Context, op *sourcegraph.DefsRefreshIndexOp) error {
		*called = true
		if !reflect.DeepEqual(op, wantOp) {
			t.Fatalf("unexpected DefsRefreshIndexOp, got %+v != %+v", op, wantOp)
		}
		return nil
	}
	return
}
