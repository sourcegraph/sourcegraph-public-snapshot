package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
)

var m sync.Mutex
var durations = map[string][]float64{}

// queryGraphQL performs a GraphQL request and stores its latency not the global durations
// map. If the verbose flag is set, a line with the request's latency is printed.
func queryGraphQL(_ context.Context, queryName, query string, variables map[string]any, target any) error {
	requestStart := time.Now()

	if err := internal.GraphQLClient().GraphQL(internal.SourcegraphAccessToken, query, variables, target); err != nil {
		return err
	}

	duration := time.Since(requestStart)

	m.Lock()
	durations[queryName] = append(durations[queryName], float64(duration)/float64(time.Millisecond))
	m.Unlock()

	if verbose {
		fmt.Printf("[%5s] %s Completed %s request in %s\n", internal.TimeSince(start), internal.EmojiSuccess, queryName, duration)
	}

	return nil
}

// formatPercentiles returns a string slice describing latency histograms for each query.
func formatPercentiles() []string {
	names := queryNames()
	lines := make([]string, 0, len(names))
	sort.Strings(names)

	for _, queryName := range names {
		numRequests, percentileValues := percentiles(queryName, 0.50, 0.95, 0.99)

		lines = append(
			lines,
			fmt.Sprintf("queryName=%s\trequests=%d\tp50=%s\tp95=%s\tp99=%s",
				queryName,
				numRequests,
				percentileValues[0.50],
				percentileValues[0.95],
				percentileValues[0.99],
			))
	}

	return lines
}

// queryNames returns the keys of the duration map.
func queryNames() (names []string) {
	m.Lock()
	defer m.Unlock()

	names = make([]string, 0, len(durations))
	for queryName := range durations {
		names = append(names, queryName)
	}

	return names
}

// percentiles returns the number of samples and the ps[i]th percentile durations of the given query type.
func percentiles(queryName string, ps ...float64) (int, map[float64]time.Duration) {
	m.Lock()
	defer m.Unlock()

	queryDurations := durations[queryName]
	sort.Float64s(queryDurations)

	percentiles := make(map[float64]time.Duration, len(ps))
	for _, p := range ps {
		index := int(float64(len(queryDurations)) * p)
		percentiles[p] = time.Duration(queryDurations[index]) * time.Millisecond
	}

	return len(queryDurations), percentiles
}
