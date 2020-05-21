package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/schollz/progressbar/v3"
)

func main() {
	token := flag.String("token", os.Getenv("GITHUB_TOKEN"), "(required) GitHub personal access token")
	progressFilepath := flag.String("progress", "feeder.db", "path to a sqlite DB recording the progress made in the feeder (created if it doesn't exist)")
	baseURL := flag.String("baseURL", "", "(required) base URL of GHE instance to feed")
	uploadURL := flag.String("uploadURL", "", "upload URL of GHE instance to feed")
	numWorkers := flag.Int("numWorkers", 20, "number of workers")
	numGHEConcurrency := flag.Int("numGHEConcurrency", 10, "number of simultaneous GHE requests in flight")
	scratchDir := flag.String("scratchDir", "", "scratch dir where to temporarily clone repositories")
	limitPump := flag.Int64("limit", math.MaxInt64, "limit processing to this many repos (for debugging)")

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help || len(*baseURL) == 0 || len(*token) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if len(*uploadURL) == 0 {
		*uploadURL = *baseURL
	}

	if len(*scratchDir) == 0 {
		d, err := ioutil.TempDir("", "ghe-feeder")
		if err != nil {
			log.Fatal(err)
		}
		*scratchDir = d
	}

	ctx := context.Background()
	gheClient, err := newGHEClient(ctx, *baseURL, *uploadURL, *token)
	if err != nil {
		log.Fatal(err)
	}

	fdr, err := newFeederDB(*progressFilepath)
	if err != nil {
		log.Fatal(err)
	}

	gheSemaphore := make(chan struct{}, *numGHEConcurrency)

	spn := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spn.Start()

	numLines, err := numLinesTotal()
	if err != nil {
		log.Fatal(err)
	}

	if numLines > *limitPump {
		numLines = *limitPump
	}

	spn.Stop()

	bar := progressbar.New64(numLines)

	work := make(chan string)

	prdc := &producer{
		remaining: *limitPump,
		pipe:      work,
		fdr:       fdr,
	}

	var wg sync.WaitGroup

	wg.Add(*numWorkers)

	// trap Ctrl+C and call cancel on the context
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()
	go func() {
		select {
		case <-c:
			cancel()
		case <-ctx.Done():
		}
	}()

	var wkrs []*worker

	for i := 0; i < *numWorkers; i++ {
		name := fmt.Sprintf("worker-%d", i)
		wkrScratchDir := filepath.Join(*scratchDir, name)
		err := os.MkdirAll(wkrScratchDir, 0777)
		if err != nil {
			log.Fatal(err)
		}
		wkr := &worker{
			name:       name,
			client:     gheClient,
			sem:        gheSemaphore,
			index:      i,
			scratchDir: wkrScratchDir,
			work:       work,
			wg:         &wg,
			bar:        bar,
			fdr:        fdr,
		}
		wkrs = append(wkrs, wkr)
		go wkr.run(ctx)
	}

	err = prdc.pump(ctx)
	if err != nil {
		log.Fatal(err)
	}
	close(work)
	wg.Wait()

	printStats(wkrs, prdc)
}

func printStats(wkrs []*worker, prdc *producer) {
	var numProcessed, numSucceeded, numFailed int64

	for _, wkr := range wkrs {
		numProcessed += wkr.numSucceeded + wkr.numFailed
		numFailed += wkr.numFailed
		numSucceeded += wkr.numSucceeded
	}

	fmt.Printf("\n\nDone: processed %d, succeeded: %d, failed: %d, skipped: %d\n",
		numProcessed, numSucceeded, numFailed, prdc.numSkipped)
}
