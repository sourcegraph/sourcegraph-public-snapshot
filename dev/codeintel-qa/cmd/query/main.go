package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
)

var (
	numConcurrentRequests       int
	checkQueryResult            bool
	queryReferencesOfReferences bool
)

func init() {
	flag.IntVar(&numConcurrentRequests, "numConcurrentRequests", 5, "The maximum number of concurrent requests")
	flag.BoolVar(&checkQueryResult, "checkQueryResult", true, "Whether to confirm query results are correct")
	flag.BoolVar(&queryReferencesOfReferences, "queryReferencesOfReferences", false, "Whether to perform reference operations on test case references")
}

func main() {
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := mainErr(context.Background(), time.Now()); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

type queryFunc func(ctx context.Context) error

func mainErr(ctx context.Context, start time.Time) error {
	var wg sync.WaitGroup
	var numRequestsFinished uint64
	queries := buildQueries()
	errCh := make(chan error)

	for i := 0; i < numConcurrentRequests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for fn := range queries {
				if err := fn(ctx); err != nil {
					errCh <- err
				}

				if val := atomic.AddUint64(&numRequestsFinished, 1); val%100 == 0 {
					fmt.Printf("[%5s] %d queries completed\n", internal.TimeSince(start), val)
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return err
	}

	fmt.Printf("[%5s] All %d queries completed\n", internal.TimeSince(start), numRequestsFinished)
	return nil
}
