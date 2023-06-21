package streaming

import (
	"context"
	"sync/atomic"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type SearchClient interface {
	Search(ctx context.Context, query string, patternType *string, sender streaming.Sender) (*search.Alert, error)
}

func NewInsightsSearchClient(db database.DB, enterpriseJobs jobutil.EnterpriseJobs) SearchClient {
	logger := log.Scoped("insightsSearchClient", "")
	return &insightsSearchClient{
		db:           db,
		searchClient: client.New(logger, db, enterpriseJobs),
	}
}

type insightsSearchClient struct {
	db           database.DB
	searchClient client.SearchClient
}

func (r *insightsSearchClient) Search(ctx context.Context, query string, patternType *string, sender streaming.Sender) (*search.Alert, error) {
	inputs, err := r.searchClient.Plan(
		ctx,
		"",
		patternType,
		query,
		search.Precise,
		search.Streaming,
		"insights",
	)
	if err != nil {
		return nil, err
	}

	// Note: it may better to return the client.ExecutionResult, but for now it isn't as clear how to record a nice UserResultCount. Instead we just capture this ourselves.
	var resultCount atomic.Int64
	countSender := streaming.StreamFunc(func(event streaming.SearchEvent) {
		resultCount.Add(int64(len(event.Results)))
		sender.Send(event)
	})

	done := r.searchClient.Execute(ctx, countSender, inputs)
	return done(client.TelemetryArgs{
		UserResultSize: int(resultCount.Load()),
	})
}
