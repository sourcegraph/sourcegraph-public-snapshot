package main

import (
	"bufio"
	"context"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/schollz/progressbar/v3"
	"github.com/sourcegraph/log"
)

// extractOwnerRepoFromCSVLine extracts the owner and repo from a line that comes from a CSV file that a GHE instance
// created in a repo report (so it has a certain number of fields).
// for example: 2019-05-23 15:24:16 -0700,4,Organization,sourcegraph,9,tsenart-vegeta,public,1.64 MB,1683,0,false,false
// we're looking for field number 6 (tsenart-vegeta in the example) and split it into owner/repo by replacing the first
// '-' with a '/' (the owner and repo were merged when added, this is the owner on github.com, not in the GHE)
func extractOwnerRepoFromCSVLine(line string) string {
	if len(line) == 0 {
		return line
	}

	elems := strings.Split(line, ",")
	if len(elems) != 12 {
		return ""
	}

	var ownerRepo = elems[5]
	return strings.Replace(ownerRepo, "-", "/", 1)
}

// producer is pumping input line by line into the pipe channel for processing by the workers.
type producer struct {
	// how many lines are remaining to be processed
	remaining int64
	// where to send each ownerRepo. the workers expect 'owner/repo' strings
	pipe chan<- string
	// sqlite DB where each ownerRepo is declared (to keep progress and to implement resume functionality)
	fdr *feederDB
	// how many we have already processed
	numAlreadyDone int64
	// logger for the pump
	logger log.Logger
	// terminal UI progress bar
	bar *progressbar.ProgressBar
	// skips this many lines from the input before starting to feed into the pipe
	skipNumLines int64
}

// pumpFile reads the specified file line by line and feeds ownerRepo strings into the pipe
func (prdc *producer) pumpFile(ctx context.Context, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	isCSV := strings.HasSuffix(path, ".csv")

	scanner := bufio.NewScanner(file)
	lineNum := int64(0)
	for scanner.Scan() && prdc.remaining > 0 {
		if prdc.skipNumLines > 0 {
			prdc.skipNumLines--
			continue
		}
		line := strings.TrimSpace(scanner.Text())
		if isCSV {
			line = extractOwnerRepoFromCSVLine(line)
		} else {
			line = strings.Trim(line, "\"")
		}
		if len(line) == 0 {
			continue
		}
		alreadyDone, err := prdc.fdr.declareRepo(line)
		if err != nil {
			return err
		}
		if alreadyDone {
			prdc.numAlreadyDone++
			_ = prdc.bar.Add(1)
			reposAlreadyDoneCounter.Inc()
			reposProcessedCounter.With(prometheus.Labels{"worker": "skipped"}).Inc()
			reposSucceededCounter.Inc()
			remainingWorkGauge.Add(-1.0)
			prdc.remaining--
			prdc.logger.Debug("repo already done in previous run", log.String("owner/repo", line))
			continue
		}
		select {
		case prdc.pipe <- line:
			prdc.remaining--
		case <-ctx.Done():
			return scanner.Err()
		}
		lineNum++
	}

	return scanner.Err()
}

// pump finds all the input files specified as command line by recursively going through all specified directories
// and looking for '*.csv', '*.json' and '*.txt' files.
func (prdc *producer) pump(ctx context.Context) error {
	for _, root := range flag.Args() {
		if ctx.Err() != nil || prdc.remaining <= 0 {
			return nil
		}

		err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if prdc.remaining <= 0 {
				return nil
			}
			if strings.HasSuffix(path, ".csv") || strings.HasSuffix(path, ".txt") ||
				strings.HasSuffix(path, ".json") {
				err := prdc.pumpFile(ctx, path)
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}

// numLinesInFile counts how many lines are in the specified file (it starts counting only after skipNumLines have been
// skipped from the file). Returns counted lines, how many lines were skipped and any errors.
func numLinesInFile(path string, skipNumLines int64) (int64, int64, error) {
	var numLines, skippedLines int64

	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	counting := skipNumLines == 0
	for scanner.Scan() {
		if counting {
			numLines++
		} else {
			skippedLines++
		}
		if skippedLines == skipNumLines {
			counting = true
		}
	}

	return numLines, skippedLines, scanner.Err()
}

// numLinesTotal goes through all the inputs and counts how many lines are available for processing.
func numLinesTotal(skipNumLines int64) (int64, error) {
	var numLines int64
	skippedLines := skipNumLines

	for _, root := range flag.Args() {
		err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if strings.HasSuffix(path, ".csv") || strings.HasSuffix(path, ".txt") ||
				strings.HasSuffix(path, ".json") {
				nl, sl, err := numLinesInFile(path, skippedLines)
				if err != nil {
					return err
				}
				numLines += nl
				skippedLines -= sl
			}
			return nil
		})

		if err != nil {
			return 0, err
		}
	}

	return numLines, nil
}
