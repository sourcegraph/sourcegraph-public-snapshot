package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewRepositoryIndexer(
	observationCtx *observation.Context,
	store store.Store,
	gitserverClient GitserverClient,
	symbolsClient SymbolsClient,
	interval time.Duration,
) goroutine.BackgroundRoutine {
	indexer := &indexer{
		store:           store,
		gitserverClient: gitserverClient,
		symbolsClient:   symbolsClient,
	}

	return goroutine.NewPeriodicGoroutine(
		context.Background(),
		"pagerank.repo-indexer", "builds page-rank indexes for repositories",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return indexer.indexRepositories(ctx)
		}),
	)
}

// We currently disable this until we find a better way to calculate a graph locally
// and its page rank
var rankingEnabled = false

type indexer struct {
	store           store.Store
	gitserverClient GitserverClient
	symbolsClient   SymbolsClient
}

func (i *indexer) indexRepositories(ctx context.Context) (err error) {
	if !rankingEnabled {
		return nil
	}

	repos, err := i.store.GetRepos(ctx)
	if err != nil {
		return err
	}

	for _, repoName := range repos {
		if err := i.indexRepository(ctx, repoName); err != nil {
			return err
		}
	}

	return nil
}

const fileReferenceGraphPrecision = 0.5

func (i *indexer) indexRepository(ctx context.Context, repoName api.RepoName) (err error) {
	graph, err := i.buildFileReferenceGraph(ctx, repoName)
	if err != nil {
		return err
	}

	ranks, err := pageRankFromStreamingGraph(ctx, graph)
	if err != nil {
		return err
	}

	return i.store.SetDocumentRanks(ctx, repoName, fileReferenceGraphPrecision, ranks)
}
