package pipeline

import (
	"context"
	golog "log"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	itypes "github.com/sourcegraph/sourcegraph/internal/types"
)

type BackfillRequest struct {
	Series *types.InsightSeries
	Repo   *itypes.MinimalRepo
}

type Backfiller interface {
	Run(ctx context.Context, request BackfillRequest) error
}

type SearchJobGeneratorOutput struct {
	*BackfillRequest
	Job *queryrunner.Job
}

type SearchResultOutput struct {
	*BackfillRequest
	err    error
	result searchResult
}

type SearchJobGenerator func(ctx context.Context, req BackfillRequest) <-chan SearchJobGeneratorOutput
type SearchRunner func(ctx context.Context, input <-chan SearchJobGeneratorOutput) <-chan SearchResultOutput
type ResultsPersister func(ctx context.Context, input <-chan SearchResultOutput) error

func NewBackfiller(jobGenerator SearchJobGenerator, searchRunner SearchRunner, resultsPersister ResultsPersister) Backfiller {
	return &backfiller{
		searchJobGenerator: jobGenerator,
		searchRunner:       searchRunner,
		persister:          resultsPersister,
		logger:             log.Scoped("insights_backfill_pipeline", ""),
	}

}

type backfiller struct {
	//dependencies
	searchJobGenerator SearchJobGenerator
	searchRunner       SearchRunner
	persister          ResultsPersister
	logger             log.Logger
}

func (b *backfiller) Run(ctx context.Context, req BackfillRequest) error {

	jobsChan := b.searchJobGenerator(ctx, req)
	searchResultsChan := b.searchRunner(ctx, jobsChan)
	return b.persister(ctx, searchResultsChan)

}

// Implimentation of steps for Backfill process

type searchResult struct {
	count       int
	capture     string
	repo        *itypes.MinimalRepo
	pointInTime time.Time
}

func makeRunSearchFunc(logger log.Logger, searchClient streaming.SearchClient) func(context.Context, <-chan SearchJobGeneratorOutput) <-chan SearchResultOutput {
	return func(ctx context.Context, in <-chan SearchJobGeneratorOutput) <-chan SearchResultOutput {

		out := make(chan SearchResultOutput)
		go func(ctx context.Context, outputChannel chan SearchResultOutput) {
			defer func() {
				if err := recover(); err != nil {
					stack := debug.Stack()
					golog.Printf("goroutine panic: %v\n%s", err, stack)
				}
				close(out)
			}()
			for r := range in {

				// run search
				// some made up values
				time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
				logger.Debug("running the search job")
				outputChannel <- SearchResultOutput{
					BackfillRequest: r.BackfillRequest,
					result:          searchResult{count: 10, capture: "", repo: r.BackfillRequest.Repo, pointInTime: *r.Job.RecordTime},
					err:             nil}
			}
		}(ctx, out)
		return out
	}
}
