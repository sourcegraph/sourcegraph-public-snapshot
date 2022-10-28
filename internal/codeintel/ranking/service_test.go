package ranking

import (
	"context"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetRepoRank(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	gitserverClient := NewMockGitserverClient()
	svc := newService(mockStore, nil, gitserverClient, nil, siteConfigQuerier{}, nil, &observation.TestContext)

	mockStore.GetStarRankFunc.SetDefaultReturn(0.6, nil)

	rank, err := svc.GetRepoRank(ctx, api.RepoName("foo"))
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
	gitserverClient := NewMockGitserverClient()
	mockConfigQuerier := NewMockSiteConfigQuerier()
	svc := newService(mockStore, nil, gitserverClient, nil, mockConfigQuerier, nil, &observation.TestContext)

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

	rank, err := svc.GetRepoRank(ctx, api.RepoName("github.com/foo/bar"))
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

func TestGetDocumentRanks(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	gitserverClient := NewMockGitserverClient()
	svc := newService(mockStore, nil, gitserverClient, nil, siteConfigQuerier{}, nil, &observation.TestContext)

	gitserverClient.ListFilesForRepoFunc.SetDefaultReturn([]string{
		"main.go",
		"code/c.go",
		"code/b.go",
		"code/a.go",
		"code/d.go",
		"test/c.go",           // test
		"test/b.go",           // test
		"test/a.go",           // test
		"test/d.go",           // test
		"rendered/web/min.js", // generated
		"node_modules/foo.js", // vendor
		"node_modules/bar.js", // vendor
		"node_modules/baz.js", // vendor
	}, nil)

	ranks, err := svc.GetDocumentRanks(ctx, api.RepoName("foo"))
	if err != nil {
		t.Fatalf("unexpected error getting repo rank: %s", err)
	}

	expected := map[string][]float64{
		"code/a.go":           {1, 1, 1, 0.100, 1 - (0.00 / 13.0)},
		"code/b.go":           {1, 1, 1, 0.100, 1 - (1.00 / 13.0)},
		"code/c.go":           {1, 1, 1, 0.100, 1 - (2.00 / 13.0)},
		"code/d.go":           {1, 1, 1, 0.100, 1 - (3.00 / 13.0)},
		"main.go":             {1, 1, 1, 0.125, 1 - (4.00 / 13.0)},
		"node_modules/bar.js": {1, 0, 1, 0.050, 1 - (5.00 / 13.0)},
		"node_modules/baz.js": {1, 0, 1, 0.050, 1 - (6.00 / 13.0)},
		"node_modules/foo.js": {1, 0, 1, 0.050, 1 - (7.00 / 13.0)},
		"rendered/web/min.js": {0, 1, 1, 0.050, 1 - (8.00 / 13.0)},
		"test/a.go":           {1, 1, 0, 0.100, 1 - (9.00 / 13.0)},
		"test/b.go":           {1, 1, 0, 0.100, 1 - (10.0 / 13.0)},
		"test/c.go":           {1, 1, 0, 0.100, 1 - (11.0 / 13.0)},
		"test/d.go":           {1, 1, 0, 0.100, 1 - (12.0 / 13.0)},
	}

	opt := cmp.Comparer(cmpFloat)

	if diff := cmp.Diff(expected, ranks, opt); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}
}

const epsilon = 0.00000001

func cmpFloat(x, y float64) bool {
	return math.Abs(x-y) < epsilon
}
