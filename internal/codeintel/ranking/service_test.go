package ranking

import (
	"context"
	"math"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetRepoRank(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	svc := newService(observation.TestContextTB(t), mockStore, nil, conf.DefaultClient())

	mockStore.GetStarRankFunc.SetDefaultReturn(0.6, nil)

	rank, err := svc.GetRepoRank(ctx, "foo")
	if err != nil {
		t.Fatalf("unexpected error getting repo rank: %s", err)
	}

	if expected := 0.0; !cmpFloat(rank[0], expected) {
		t.Errorf("unexpected rank[0]. want=%.5f have=%.5f", expected, rank[0])
	}
	if expected := 0.6; !cmpFloat(rank[1], expected) {
		t.Errorf("unexpected rank[1]. want=%.5f have=%.5f", expected, rank[1])
	}
}

func TestGetRepoRankWithUserBoostedScores(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	mockConfigQuerier := NewMockSiteConfigQuerier()
	svc := newService(observation.TestContextTB(t), mockStore, nil, mockConfigQuerier)

	mockStore.GetStarRankFunc.SetDefaultReturn(0.6, nil)
	mockConfigQuerier.SiteConfigFunc.SetDefaultReturn(schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			Ranking: &schema.Ranking{
				RepoScores: map[string]float64{
					"github.com/foo":     400, // matches
					"github.com/foo/baz": 600, // no match
					"github.com/bar":     200, // no match
				},
			},
		},
	})

	rank, err := svc.GetRepoRank(ctx, "github.com/foo/bar")
	if err != nil {
		t.Fatalf("unexpected error getting repo rank: %s", err)
	}

	if expected := 400.0 / 401.0; !cmpFloat(rank[0], expected) {
		t.Errorf("unexpected rank[0]. want=%.5f have=%.5f", expected, rank[0])
	}
	if expected := 0.6; !cmpFloat(rank[1], expected) {
		t.Errorf("unexpected rank[1]. want=%.5f have=%.5f", expected, rank[1])
	}
}

const epsilon = 0.00000001

func cmpFloat(x, y float64) bool {
	return math.Abs(x-y) < epsilon
}
