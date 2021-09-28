package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"
)

var (
	indexDir             string
	numConcurrentUploads int
)

func init() {
	// Default assumes running from the dev/codeintel-qa directory
	flag.StringVar(&indexDir, "indexDir", "./testdata/indexes", "The location of the testdata directory")
	flag.IntVar(&numConcurrentUploads, "numConcurrentUploads", 5, "The maximum number of concurrent uploads")
}

func main() {
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	limiter := newLimiter(numConcurrentUploads)
	defer limiter.close()

	if err := mainErr(context.Background(), limiter, time.Now()); err != nil {
		fmt.Printf("error: %s\n", err.Error())
		os.Exit(1)
	}
}

func mainErr(ctx context.Context, limiter *limiter, start time.Time) error {
	commitsByRepo, err := commitsByRepo()
	if err != nil {
		return err
	}

	repoNames := make([]string, 0, len(commitsByRepo))
	for name := range commitsByRepo {
		repoNames = append(repoNames, name)
	}
	sort.Strings(repoNames)

	uploads, err := uploadAll(ctx, commitsByRepo, limiter, start)
	if err != nil {
		return err
	}

	if err := monitor(ctx, repoNames, uploads, start); err != nil {
		return err
	}

	return nil
}
