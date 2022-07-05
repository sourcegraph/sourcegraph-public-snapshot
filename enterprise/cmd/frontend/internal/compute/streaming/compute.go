package streaming

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func toComputeResult(ctx context.Context, db database.DB, cmd compute.Command, match result.Match) (out []compute.Result, _ error) {
	if v, ok := match.(*result.CommitMatch); ok && v.DiffPreview != nil {
		for _, diffMatch := range v.CommitToDiffMatches() {
			result, err := cmd.Run(ctx, db, diffMatch)
			if err != nil {
				return nil, err
			}
			out = append(out, result)
		}
	} else {
		result, err := cmd.Run(ctx, db, match)
		if err != nil {
			return nil, err
		}
		out = append(out, result)
	}
	return out, nil
}

func NewComputeStream(ctx context.Context, logger log.Logger, db database.DB, query string) (<-chan Event, func() (*search.Alert, error)) {
	computeQuery, err := compute.Parse(query)
	if err != nil {
		return nil, func() (*search.Alert, error) { return nil, err }
	}

	searchQuery, err := computeQuery.ToSearchQuery()
	if err != nil {
		return nil, func() (*search.Alert, error) { return nil, err }
	}

	eventsC := make(chan Event, 8)
	errorC := make(chan error, 1)
	g := group.NewWithStreaming[Event]().WithErrors().WithMaxConcurrency(8)
	cb := func(ev Event, err error) {
		if err != nil {
			select {
			case errorC <- err:
			default:
			}
		} else {
			eventsC <- ev
		}
	}
	stream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		if !event.Stats.Zero() {
			g.Go(func() (Event, error) {
				return Event{nil, event.Stats}, nil
			}, cb)
		}
		for _, match := range event.Results {
			match := match
			g.Go(func() (Event, error) {
				results, err := toComputeResult(ctx, db, computeQuery.Command, match)
				return Event{results, streaming.Stats{}}, err
			}, cb)
		}
	})

	settings, err := graphqlbackend.DecodedViewerFinalSettings(ctx, db)
	if err != nil {
		close(eventsC)
		close(errorC)
		return eventsC, func() (*search.Alert, error) { return nil, err }
	}

	patternType := "regexp"
	searchClient := client.NewSearchClient(logger, db, search.Indexed(), search.SearcherURLs())
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
		defer g.Wait()

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
