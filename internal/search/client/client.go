package client

import (
	"context"

	"github.com/google/zoekt"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

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

func NewSearchClient(logger log.Logger, db database.DB, zoektStreamer zoekt.Streamer, searcherURLs *endpoint.Map) SearchClient {
	return &searchClient{
		logger:       logger,
		db:           db,
		zoekt:        zoektStreamer,
		searcherURLs: searcherURLs,
	}
}

type searchClient struct {
	logger       log.Logger
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
) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "Execute", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	planJob, err := jobutil.NewPlanJob(inputs, inputs.Plan)
	if err != nil {
		return nil, err
	}

	return planJob.Run(ctx, s.JobClients(), stream)
}

func (s *searchClient) JobClients() job.RuntimeClients {
	return job.RuntimeClients{
		Logger:       s.logger,
		DB:           s.db,
		Zoekt:        s.zoekt,
		SearcherURLs: s.searcherURLs,
		Gitserver:    gitserver.NewClient(s.db),
	}
}
