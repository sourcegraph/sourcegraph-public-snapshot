package ranking

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGetRepoRank(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	gitserverClient := NewMockGitserverClient()
	svc := newService(mockStore, nil, gitserverClient, &observation.TestContext)

	mockStore.GetStarRankFunc.SetDefaultReturn(0.6, nil)

	rank, err := svc.GetRepoRank(ctx, api.RepoName("foo"))
	if err != nil {
		t.Fatalf("unexpected error getting repo rank: %s", err)
	}

	if expected := float64(0); rank[0] != expected {
		t.Errorf("unexpected rank[0]. want=%.2f have=%.2f", expected, rank[0])
	}
	if expected := 0.4; rank[1] != expected {
		t.Errorf("unexpected rank[1]. want=%.2f have=%.2f", expected, rank[1])
	}
}

func TestGetDocumentRanks(t *testing.T) {
	ctx := context.Background()
	mockStore := NewMockStore()
	gitserverClient := NewMockGitserverClient()
	svc := newService(mockStore, nil, gitserverClient, &observation.TestContext)

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
		"code/a.go":           {0, 0, 0, 0.9, float64(0) / float64(13)},
		"code/b.go":           {0, 0, 0, 0.9, float64(1) / float64(13)},
		"code/c.go":           {0, 0, 0, 0.9, float64(2) / float64(13)},
		"code/d.go":           {0, 0, 0, 0.9, float64(3) / float64(13)},
		"main.go":             {0, 0, 0, 0.875, float64(4) / float64(13)},
		"node_modules/bar.js": {0, 1, 0, 0.95, float64(5) / float64(13)},
		"node_modules/baz.js": {0, 1, 0, 0.95, float64(6) / float64(13)},
		"node_modules/foo.js": {0, 1, 0, 0.95, float64(7) / float64(13)},
		"rendered/web/min.js": {1, 0, 0, 0.95, float64(8) / float64(13)},
		"test/a.go":           {0, 0, 1, 0.9, float64(9) / float64(13)},
		"test/b.go":           {0, 0, 1, 0.9, float64(10) / float64(13)},
		"test/c.go":           {0, 0, 1, 0.9, float64(11) / float64(13)},
		"test/d.go":           {0, 0, 1, 0.9, float64(12) / float64(13)},
	}

	if diff := cmp.Diff(expected, ranks); diff != "" {
		t.Errorf("unexpected ranks (-want +got):\n%s", diff)
	}
}
