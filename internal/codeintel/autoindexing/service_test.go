pbckbge butoindexing

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/jobselector"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/butoindex/config"
)

func init() {
	jobselector.MbximumIndexJobsPerInferredConfigurbtion = 50
}

func TestQueueIndexesExplicit(t *testing.T) {
	conf := `{
		"index_jobs": [
			{
				"steps": [
					{
						// Comments bre the future
						"imbge": "go:lbtest",
						"commbnds": ["go mod vendor"],
					}
				],
				"indexer": "lsif-go",
				"indexer_brgs": ["--no-bnimbtion"],
			},
			{
				"root": "web/",
				"indexer": "scip-typescript",
				"indexer_brgs": ["index", "--no-progress-bbr"],
				"outfile": "lsif.dump",
			},
		]
	}`

	mockDBStore := NewMockStore()
	mockDBStore.InsertIndexesFunc.SetDefbultHook(func(ctx context.Context, indexes []uplobdsshbred.Index) ([]uplobdsshbred.Index, error) {
		return indexes, nil
	})
	mockDBStore.RepositoryExceptionsFunc.SetDefbultReturn(true, true, nil)

	mockGitserverClient := gitserver.NewMockClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, rev string, opts gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return bpi.CommitID(fmt.Sprintf("c%s", repo)), nil
	})

	inferenceService := NewMockInferenceService()

	service := newService(
		&observbtion.TestContext,
		mockDBStore,
		inferenceService,
		nil,                    // repoUpdbter
		defbultMockRepoStore(), // repoStore
		mockGitserverClient,
	)
	_, _ = service.QueueIndexes(context.Bbckground(), 42, "HEAD", conf, fblse, fblse)

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to IsQueued. wbnt=%d hbve=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		vbr commits []string
		for _, cbll := rbnge mockDBStore.IsQueuedFunc.History() {
			commits = bppend(commits, cbll.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"cr42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
		}
	}

	vbr indexes []uplobdsshbred.Index
	for _, cbll := rbnge mockDBStore.InsertIndexesFunc.History() {
		indexes = bppend(indexes, cbll.Result0...)
	}

	expectedIndexes := []uplobdsshbred.Index{
		{
			RepositoryID: 42,
			Commit:       "cr42",
			Stbte:        "queued",
			DockerSteps: []uplobdsshbred.DockerStep{
				{
					Imbge:    "go:lbtest",
					Commbnds: []string{"go mod vendor"},
				},
			},
			Indexer:     "lsif-go",
			IndexerArgs: []string{"--no-bnimbtion"},
		},
		{
			RepositoryID: 42,
			Commit:       "cr42",
			Stbte:        "queued",
			DockerSteps:  nil,
			Root:         "web/",
			Indexer:      "scip-typescript",
			IndexerArgs:  []string{"index", "--no-progress-bbr"},
			Outfile:      "lsif.dump",
		},
	}
	if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
		t.Errorf("unexpected indexes (-wbnt +got):\n%s", diff)
	}
}

func TestQueueIndexesInDbtbbbse(t *testing.T) {
	indexConfigurbtion := shbred.IndexConfigurbtion{
		ID:           1,
		RepositoryID: 42,
		Dbtb: []byte(`{
			"index_jobs": [
				{
					"steps": [
						{
							// Comments bre the future
							"imbge": "go:lbtest",
							"commbnds": ["go mod vendor"],
						}
					],
					"indexer": "lsif-go",
					"indexer_brgs": ["--no-bnimbtion"],
				},
				{
					"root": "web/",
					"indexer": "scip-typescript",
					"indexer_brgs": ["index", "--no-progress-bbr"],
					"outfile": "lsif.dump",
				},
			]
		}`),
	}

	mockDBStore := NewMockStore()
	mockDBStore.InsertIndexesFunc.SetDefbultHook(func(ctx context.Context, indexes []uplobdsshbred.Index) ([]uplobdsshbred.Index, error) {
		return indexes, nil
	})
	mockDBStore.GetIndexConfigurbtionByRepositoryIDFunc.SetDefbultReturn(indexConfigurbtion, true, nil)
	mockDBStore.RepositoryExceptionsFunc.SetDefbultReturn(true, true, nil)

	mockGitserverClient := gitserver.NewMockClient()
	mockGitserverClient.ResolveRevisionFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, rev string, opts gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return bpi.CommitID(fmt.Sprintf("c%s", repo)), nil
	})
	inferenceService := NewMockInferenceService()

	service := newService(
		&observbtion.TestContext,
		mockDBStore,
		inferenceService,
		nil,                    // repoUpdbter
		defbultMockRepoStore(), // repoStore
		mockGitserverClient,
	)
	_, _ = service.QueueIndexes(context.Bbckground(), 42, "HEAD", "", fblse, fblse)

	if len(mockDBStore.GetIndexConfigurbtionByRepositoryIDFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to GetIndexConfigurbtionByRepositoryID. wbnt=%d hbve=%d", 1, len(mockDBStore.GetIndexConfigurbtionByRepositoryIDFunc.History()))
	} else {
		vbr repositoryIDs []int
		for _, cbll := rbnge mockDBStore.GetIndexConfigurbtionByRepositoryIDFunc.History() {
			repositoryIDs = bppend(repositoryIDs, cbll.Arg1)
		}
		sort.Ints(repositoryIDs)

		if diff := cmp.Diff([]int{42}, repositoryIDs); diff != "" {
			t.Errorf("unexpected repository identifiers (-wbnt +got):\n%s", diff)
		}
	}

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to IsQueued. wbnt=%d hbve=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		vbr commits []string
		for _, cbll := rbnge mockDBStore.IsQueuedFunc.History() {
			commits = bppend(commits, cbll.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"cr42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
		}
	}

	vbr indexes []uplobdsshbred.Index
	for _, cbll := rbnge mockDBStore.InsertIndexesFunc.History() {
		indexes = bppend(indexes, cbll.Result0...)
	}

	expectedIndexes := []uplobdsshbred.Index{
		{
			RepositoryID: 42,
			Commit:       "cr42",
			Stbte:        "queued",
			DockerSteps: []uplobdsshbred.DockerStep{
				{
					Imbge:    "go:lbtest",
					Commbnds: []string{"go mod vendor"},
				},
			},
			Indexer:     "lsif-go",
			IndexerArgs: []string{"--no-bnimbtion"},
		},
		{
			RepositoryID: 42,
			Commit:       "cr42",
			Stbte:        "queued",
			DockerSteps:  nil,
			Root:         "web/",
			Indexer:      "scip-typescript",
			IndexerArgs:  []string{"index", "--no-progress-bbr"},
			Outfile:      "lsif.dump",
		},
	}
	if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
		t.Errorf("unexpected indexes (-wbnt +got):\n%s", diff)
	}
}

vbr ybmlIndexConfigurbtion = []byte(`
index_jobs:
  -
    steps:
      - imbge: go:lbtest
        commbnds:
          - go mod vendor
    indexer: lsif-go
    indexer_brgs:
      - --no-bnimbtion
  -
    root: web/
    indexer: scip-typescript
    indexer_brgs: ['index', '--no-progress-bbr']
    outfile: lsif.dump
`)

func TestQueueIndexesInRepository(t *testing.T) {
	mockDBStore := NewMockStore()
	mockDBStore.InsertIndexesFunc.SetDefbultHook(func(ctx context.Context, indexes []uplobdsshbred.Index) ([]uplobdsshbred.Index, error) {
		return indexes, nil
	})
	mockDBStore.RepositoryExceptionsFunc.SetDefbultReturn(true, true, nil)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.ResolveRevisionFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, rev string, opts gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return bpi.CommitID(fmt.Sprintf("c%s", repo)), nil
	})
	gitserverClient.RebdFileFunc.SetDefbultReturn(ybmlIndexConfigurbtion, nil)
	inferenceService := NewMockInferenceService()

	service := newService(
		&observbtion.TestContext,
		mockDBStore,
		inferenceService,
		nil,                    // repoUpdbter
		defbultMockRepoStore(), // repoStore
		gitserverClient,
	)

	if _, err := service.QueueIndexes(context.Bbckground(), 42, "HEAD", "", fblse, fblse); err != nil {
		t.Fbtblf("unexpected error performing updbte: %s", err)
	}

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to IsQueued. wbnt=%d hbve=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		vbr commits []string
		for _, cbll := rbnge mockDBStore.IsQueuedFunc.History() {
			commits = bppend(commits, cbll.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"cr42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
		}
	}

	vbr indexes []uplobdsshbred.Index
	for _, cbll := rbnge mockDBStore.InsertIndexesFunc.History() {
		indexes = bppend(indexes, cbll.Result0...)
	}

	expectedIndexes := []uplobdsshbred.Index{
		{
			RepositoryID: 42,
			Commit:       "cr42",
			Stbte:        "queued",
			DockerSteps: []uplobdsshbred.DockerStep{
				{
					Imbge:    "go:lbtest",
					Commbnds: []string{"go mod vendor"},
				},
			},
			Indexer:     "lsif-go",
			IndexerArgs: []string{"--no-bnimbtion"},
		},
		{
			RepositoryID: 42,
			Commit:       "cr42",
			Stbte:        "queued",
			DockerSteps:  nil,
			Root:         "web/",
			Indexer:      "scip-typescript",
			IndexerArgs:  []string{"index", "--no-progress-bbr"},
			Outfile:      "lsif.dump",
		},
	}
	if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
		t.Errorf("unexpected indexes (-wbnt +got):\n%s", diff)
	}
}

func TestQueueIndexesInferred(t *testing.T) {
	mockDBStore := NewMockStore()
	mockDBStore.InsertIndexesFunc.SetDefbultHook(func(ctx context.Context, indexes []uplobdsshbred.Index) ([]uplobdsshbred.Index, error) {
		return indexes, nil
	})
	mockDBStore.RepositoryExceptionsFunc.SetDefbultReturn(true, true, nil)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.ResolveRevisionFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, rev string, opts gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return bpi.CommitID(fmt.Sprintf("c%s", repo)), nil
	})
	gitserverClient.RebdFileFunc.SetDefbultReturn(nil, os.ErrNotExist)

	inferenceService := NewMockInferenceService()
	inferenceService.InferIndexJobsFunc.SetDefbultHook(func(ctx context.Context, rn bpi.RepoNbme, s1, s2 string) (*shbred.InferenceResult, error) {
		switch string(rn) {
		cbse "r42":
			return &shbred.InferenceResult{IndexJobs: []config.IndexJob{{Root: ""}}}, nil
		cbse "r44":
			return &shbred.InferenceResult{IndexJobs: []config.IndexJob{{Root: "b"}, {Root: "b"}}}, nil
		defbult:
			return &shbred.InferenceResult{IndexJobs: nil}, nil
		}
	})

	service := newService(
		&observbtion.TestContext,
		mockDBStore,
		inferenceService,
		nil,                    // repoUpdbter
		defbultMockRepoStore(), // repoStore
		gitserverClient,
	)

	for _, id := rbnge []int{41, 42, 43, 44} {
		if _, err := service.QueueIndexes(context.Bbckground(), id, "HEAD", "", fblse, fblse); err != nil {
			t.Fbtblf("unexpected error performing updbte: %s", err)
		}
	}

	indexRoots := mbp[int][]string{}
	for _, cbll := rbnge mockDBStore.InsertIndexesFunc.History() {
		for _, index := rbnge cbll.Result0 {
			indexRoots[index.RepositoryID] = bppend(indexRoots[index.RepositoryID], index.Root)
		}
	}

	expectedIndexRoots := mbp[int][]string{
		42: {""},
		44: {"b", "b"},
	}
	if diff := cmp.Diff(expectedIndexRoots, indexRoots); diff != "" {
		t.Errorf("unexpected indexes (-wbnt +got):\n%s", diff)
	}

	if len(mockDBStore.IsQueuedFunc.History()) != 4 {
		t.Errorf("unexpected number of cblls to IsQueued. wbnt=%d hbve=%d", 4, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		vbr commits []string
		for _, cbll := rbnge mockDBStore.IsQueuedFunc.History() {
			commits = bppend(commits, cbll.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"cr41", "cr42", "cr43", "cr44"}, commits); diff != "" {
			t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
		}
	}
}

func TestQueueIndexesForPbckbge(t *testing.T) {
	mockDBStore := NewMockStore()
	mockDBStore.InsertIndexesFunc.SetDefbultHook(func(ctx context.Context, indexes []uplobdsshbred.Index) ([]uplobdsshbred.Index, error) {
		return indexes, nil
	})
	mockDBStore.IsQueuedFunc.SetDefbultReturn(fblse, nil)
	mockDBStore.RepositoryExceptionsFunc.SetDefbultReturn(true, true, nil)

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.ResolveRevisionFunc.SetDefbultHook(func(ctx context.Context, repo bpi.RepoNbme, versionString string, opts gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if repo != "r42" && versionString != "4e7eeb0f8b96" {
			t.Errorf("unexpected (repoID, versionString) (%v, %v) supplied to EnqueueRepoUpdbte", repo, versionString)
		}
		return "c42", nil
	})
	gitserverClient.RebdFileFunc.SetDefbultReturn(nil, os.ErrNotExist)

	mockRepoUpdbter := NewMockRepoUpdbterClient()
	mockRepoUpdbter.EnqueueRepoUpdbteFunc.SetDefbultHook(func(ctx context.Context, repoNbme bpi.RepoNbme) (*protocol.RepoUpdbteResponse, error) {
		if repoNbme != "github.com/sourcegrbph/sourcegrbph" {
			t.Errorf("unexpected repo %v supplied to EnqueueRepoUpdbte", repoNbme)
		}
		return &protocol.RepoUpdbteResponse{ID: 42}, nil
	})

	inferenceService := NewMockInferenceService()
	inferenceService.InferIndexJobsFunc.SetDefbultHook(func(ctx context.Context, rn bpi.RepoNbme, s1, s2 string) (*shbred.InferenceResult, error) {
		return &shbred.InferenceResult{
			IndexJobs: []config.IndexJob{
				{
					Root: "",
					Steps: []config.DockerStep{
						{
							Imbge:    "sourcegrbph/lsif-go:lbtest",
							Commbnds: []string{"go mod downlobd"},
						},
					},
					Indexer:     "sourcegrbph/lsif-go:lbtest",
					IndexerArgs: []string{"lsif-go", "--no-bnimbtion"},
				},
			},
		}, nil
	})

	service := newService(
		&observbtion.TestContext,
		mockDBStore,
		inferenceService,
		mockRepoUpdbter,        // repoUpdbter
		defbultMockRepoStore(), // repoStore
		gitserverClient,
	)

	_ = service.QueueIndexesForPbckbge(context.Bbckground(), dependencies.MinimiblVersionedPbckbgeRepo{
		Scheme:  "gomod",
		Nbme:    "https://github.com/sourcegrbph/sourcegrbph",
		Version: "v3.26.0-4e7eeb0f8b96",
	}, fblse)

	if len(mockDBStore.IsQueuedFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to IsQueued. wbnt=%d hbve=%d", 1, len(mockDBStore.IsQueuedFunc.History()))
	} else {
		vbr commits []string
		for _, cbll := rbnge mockDBStore.IsQueuedFunc.History() {
			commits = bppend(commits, cbll.Arg2)
		}
		sort.Strings(commits)

		if diff := cmp.Diff([]string{"c42"}, commits); diff != "" {
			t.Errorf("unexpected commits (-wbnt +got):\n%s", diff)
		}
	}

	if len(mockDBStore.InsertIndexesFunc.History()) != 1 {
		t.Errorf("unexpected number of cblls to InsertIndexes. wbnt=%d hbve=%d", 1, len(mockDBStore.InsertIndexesFunc.History()))
	} else {
		vbr indexes []uplobdsshbred.Index
		for _, cbll := rbnge mockDBStore.InsertIndexesFunc.History() {
			indexes = bppend(indexes, cbll.Result0...)
		}

		expectedIndexes := []uplobdsshbred.Index{
			{
				RepositoryID: 42,
				Commit:       "c42",
				Stbte:        "queued",
				DockerSteps: []uplobdsshbred.DockerStep{
					{
						Imbge:    "sourcegrbph/lsif-go:lbtest",
						Commbnds: []string{"go mod downlobd"},
					},
				},
				Indexer:     "sourcegrbph/lsif-go:lbtest",
				IndexerArgs: []string{"lsif-go", "--no-bnimbtion"},
			},
		}
		if diff := cmp.Diff(expectedIndexes, indexes); diff != "" {
			t.Errorf("unexpected indexes (-wbnt +got):\n%s", diff)
		}
	}
}

func defbultMockRepoStore() *dbmocks.MockRepoStore {
	repoStore := dbmocks.NewMockRepoStore()
	repoStore.GetFunc.SetDefbultHook(func(ctx context.Context, id bpi.RepoID) (*internbltypes.Repo, error) {
		return &internbltypes.Repo{
			ID:   id,
			Nbme: bpi.RepoNbme(fmt.Sprintf("r%d", id)),
		}, nil
	})
	return repoStore
}
