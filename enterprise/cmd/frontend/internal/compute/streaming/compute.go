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

func toComputeResultStream(ctx context.Context, db database.DB, cmd compute.Command, matches []result.Match, f func(compute.Result)) error {
	for _, m := range matches {
		result, err := cmd.Run(ctx, db, m)
		if err != nil {
			return err
		}
		f(result)
	}
	return nil
}

func NewComputeStream(ctx context.Context, db database.DB, query string) (<-chan Event, func() (*search.Alert, error)) {
	computeQuery, err := compute.Parse(query)
	if err != nil {
		return nil, func() (*search.Alert, error) { return nil, err }
	}

	searchQuery, err := computeQuery.ToSearchQuery()
	if err != nil {
		return nil, func() (*search.Alert, error) { return nil, err }
	}

	eventsC := make(chan Event)
	errorC := make(chan error, 1)
	stream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		if len(event.Results) > 0 {
			callback := func(result compute.Result) {
				eventsC <- Event{Results: []compute.Result{result}}
			}
			err = toComputeResultStream(ctx, db, computeQuery.Command, event.Results, callback)
			errorC <- err
		}
	})

	settings, err := graphqlbackend.DecodedViewerFinalSettings(ctx, db)
	if err != nil {
		close(eventsC)
		close(errorC)
		return eventsC, func() (*search.Alert, error) { return nil, err }
	}

	patternType := "regexp"
	searchClient := client.NewSearchClient(db, search.Indexed(), search.SearcherURLs())
	inputs, err := searchClient.Plan(ctx, "", &patternType, searchQuery, search.Streaming, settings, envvar.SourcegraphDotComMode())
	if err != nil {
		close(eventsC)
		close(errorC)

		return eventsC, func() (*search.Alert, error) { return nil, err }
	}

	type finalResult struct {
		alert *search.Alert
		err   error
	}
	final := make(chan finalResult, 1)
	go func() {
		defer close(final)
		defer close(eventsC)
		defer close(errorC)

		alert, err := searchClient.Execute(ctx, stream, inputs)
		final <- finalResult{alert: alert, err: err}
	}()

	return eventsC, func() (*search.Alert, error) {
		computeErr := <-errorC
		if computeErr != nil {
			return nil, computeErr
		}
		f := <-final
		return f.alert, f.err
	}
}
