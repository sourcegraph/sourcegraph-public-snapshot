package streaming

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type SearchClient interface {
	Search(ctx context.Context, query string, patternType *string, sender streaming.Sender) (*search.Alert, error)
}

func NewInsightsSearchClient(db database.DB) SearchClient {
	logger := log.Scoped("insightsSearchClient")
	return &insightsSearchClient{
		db:           db,
		searchClient: client.New(logger, db, gitserver.NewClient("insights.search")),
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
		pointers.Ptr(int32(0)),
	)
	if err != nil {
		return nil, err
	}
	return r.searchClient.Execute(ctx, sender, inputs)
}
