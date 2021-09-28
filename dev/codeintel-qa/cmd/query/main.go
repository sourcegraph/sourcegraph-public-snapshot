package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

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

func buildQueries() <-chan queryFunc {
	fns := make(chan queryFunc)

	go func() {
		defer close(fns)

		for _, testCase := range testCases {
			// Definition returns defintion
			fns <- makeFunction(queryDefinitions, testCase.Definition, []Location{testCase.Definition})

			// References return definition
			for _, reference := range testCase.References {
				fns <- makeFunction(queryDefinitions, reference, []Location{testCase.Definition})
			}

			// Definition returns references
			fns <- makeFunction(queryReferences, testCase.Definition, testCase.References)

			// References return references
			if queryReferencesOfReferences {
				for _, reference := range testCase.References {
					fns <- makeFunction(queryReferences, reference, testCase.References)
				}
			}
		}
	}()

	return fns
}

// TODO - rename
func makeFunction(
	// TODO
	f func(ctx context.Context, location Location) ([]Location, error),
	source Location,
	expectedLocations []Location,
) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		locations, err := f(ctx, source)
		if err != nil {
			return err
		}

		if checkQueryResult {
			sortLocations(locations)

			if diff := cmp.Diff(expectedLocations, locations); diff != "" {
				return errors.Errorf("unexpected locations (-want +got):\n%s", diff)
			}
		}

		return nil
	}
}
