package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
)

var (
	numConcurrentRequests       int
	checkQueryResult            bool
	queryReferencesOfReferences bool
	verbose                     bool

	start = time.Now()
)

func init() {
	flag.IntVar(&numConcurrentRequests, "num-concurrent-requests", 5, "The maximum number of concurrent requests")
	flag.BoolVar(&checkQueryResult, "check-query-result", true, "Whether to confirm query results are correct")
	flag.BoolVar(&queryReferencesOfReferences, "query-references-of-references", false, "Whether to perform reference operations on test case references")
	flag.BoolVar(&verbose, "verbose", false, "Print every request")
}

func main() {
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	if err := mainErr(context.Background()); err != nil {
		fmt.Printf("%s error: %s\n", internal.EmojiFailure, err.Error())
		os.Exit(1)
	}
}

type queryFunc func(ctx context.Context) error

func mainErr(ctx context.Context) error {
	if err := internal.InitializeGraphQLClient(); err != nil {
		return err
	}

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

				atomic.AddUint64(&numRequestsFinished, 1)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

loop:
	for {
		select {
		case err, ok := <-errCh:
			if ok {
				return err
			}

			break loop

		case <-time.After(time.Second):
			if verbose {
				continue
			}

			val := atomic.LoadUint64(&numRequestsFinished)
			fmt.Printf("[%5s] %s %d queries completed\n\t%s\n", internal.TimeSince(start), internal.EmojiSuccess, val, strings.Join(formatPercentiles(), "\n\t"))
		}
	}

	fmt.Printf("[%5s] %s All %d queries completed\n", internal.TimeSince(start), internal.EmojiSuccess, numRequestsFinished)
	return nil
}
