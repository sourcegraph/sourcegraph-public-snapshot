package streaming

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func toComputeResultStream(ctx context.Context, cmd compute.Command, matches []result.Match, f func(compute.Result)) error {
	for _, m := range matches {
		result, err := cmd.Run(ctx, m)
		if err != nil {
			return err
		}
		f(result)
	}
	return nil
}

func NewComputeStream(ctx context.Context, db database.DB, query string) (<-chan Event, func() error) {
	computeQuery, err := compute.Parse(query)
	if err != nil {
		return nil, func() error { return err }
	}

	searchQuery, err := computeQuery.ToSearchQuery()
	if err != nil {
		return nil, func() error { return err }
	}

	eventsC := make(chan Event)
	stream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		if len(event.Results) > 0 {
			callback := func(result compute.Result) {
				eventsC <- Event{Results: []compute.Result{result}}
			}
			_ = toComputeResultStream(ctx, computeQuery.Command, event.Results, callback)
			// TODO(rvantonder): compute err is currently ignored. Process it and send alerts/errors as needed.
		}
	})

	settings, err := graphqlbackend.DecodedViewerFinalSettings(ctx, db)
	if err != nil {
		close(eventsC)
		return eventsC, func() error { return err }
	}

	patternType := "regexp"
	searchClient := client.NewSearchClient(db, search.Indexed(), search.SearcherURLs())
	inputs, err := searchClient.Plan(ctx, "", &patternType, searchQuery, search.Streaming, settings, envvar.SourcegraphDotComMode())
	if err != nil {
		close(eventsC)
		return eventsC, func() error { return err }
	}

	type finalResult struct {
		err error
	}
	final := make(chan finalResult, 1)
	go func() {
		defer close(final)
		defer close(eventsC)

		_, err := searchClient.Execute(ctx, stream, inputs)
		final <- finalResult{err: err}
	}()

	return eventsC, func() error {
		f := <-final
		return f.err
	}
}
