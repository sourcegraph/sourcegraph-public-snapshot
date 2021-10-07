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
func queryGraphQL(ctx context.Context, queryName, query string, variables map[string]interface{}, target interface{}) error {
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

// percentile returns the pth percentile duration of the given query type.
func percentile(queryName string, p float64) time.Duration {
	m.Lock()
	defer m.Unlock()

	queryDurations := durations[queryName]
	sort.Float64s(queryDurations)
	index := int(float64(len(queryDurations)) * p)
	return time.Duration(queryDurations[index]) * time.Millisecond
}
