package client

import (
	"context"

	"github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/execute"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/schema"
)

//go:generate ../../../dev/mockgen.sh github.com/sourcegraph/sourcegraph/internal/search/client -i SearchClient -o mock_client.go
type SearchClient interface {
	Plan(
		ctx context.Context,
		version string,
		patternType *string,
		searchQuery string,
		protocol search.Protocol,
		settings *schema.Settings,
		sourcegraphDotComMode bool,
	) (*run.SearchInputs, error)

	Execute(
		ctx context.Context,
		stream streaming.Sender,
		inputs *run.SearchInputs,
	) (_ *search.Alert, err error)

	JobClients() job.RuntimeClients
}

func NewSearchClient(db database.DB, zoektStreamer zoekt.Streamer, searcherURLs *endpoint.Map) SearchClient {
	return &searchClient{
		db:           db,
		zoekt:        zoektStreamer,
		searcherURLs: searcherURLs,
	}
}

type searchClient struct {
	db           database.DB
	zoekt        zoekt.Streamer
	searcherURLs *endpoint.Map
}

func (s *searchClient) Plan(
	ctx context.Context,
	version string,
	patternType *string,
	searchQuery string,
	protocol search.Protocol,
	settings *schema.Settings,
	sourcegraphDotComMode bool,
) (*run.SearchInputs, error) {
	return run.NewSearchInputs(ctx, s.db, version, patternType, searchQuery, protocol, settings, sourcegraphDotComMode)
}

func (s *searchClient) Execute(
	ctx context.Context,
	stream streaming.Sender,
	inputs *run.SearchInputs,
) (*search.Alert, error) {
	return execute.Execute(ctx, stream, inputs, s.JobClients())
}

func (s *searchClient) JobClients() job.RuntimeClients {
	return job.RuntimeClients{
		DB:           s.db,
		Zoekt:        s.zoekt,
		SearcherURLs: s.searcherURLs,
		Gitserver:    gitserver.NewClient(s.db),
	}
}
