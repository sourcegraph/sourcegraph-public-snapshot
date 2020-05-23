package main

import (
	"bufio"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/schollz/progressbar/v3"
)

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

type producer struct {
	remaining  int64
	pipe       chan<- string
	fdr        *feederDB
	numSkipped int64
	logger     log15.Logger
	bar        *progressbar.ProgressBar
}

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
		line := strings.TrimSpace(scanner.Text())
		if isCSV {
			line = extractOwnerRepoFromCSVLine(line)
		} else {
			line = strings.Trim(line, "\"")
		}
		if len(line) == 0 {
			continue
		}
		skip, err := prdc.fdr.declareRepo(line)
		if err != nil {
			return err
		}
		if skip {
			prdc.numSkipped++
			_ = prdc.bar.Add(1)
			reposSkippedCounter.Inc()
			reposProcessedCounter.With(prometheus.Labels{"worker": "skipped"}).Inc()
			reposSucceededCounter.Inc()
			remainingWorkGauge.Add(-1.0)
			prdc.logger.Debug("skipping repo", "owner/repo", line)
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

func (prdc *producer) pump(ctx context.Context) error {
	for _, root := range flag.Args() {
		if ctx.Err() != nil || prdc.remaining <= 0 {
			return nil
		}

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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

func numLinesInFile(path string) (int64, error) {
	numLines := int64(0)
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		numLines++
	}

	return numLines, scanner.Err()
}

func numLinesTotal() (int64, error) {
	numLines := int64(0)

	for _, root := range flag.Args() {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if strings.HasSuffix(path, ".csv") || strings.HasSuffix(path, ".txt") ||
				strings.HasSuffix(path, ".json") {
				nl, err := numLinesInFile(path)
				if err != nil {
					return err
				}
				numLines += nl
			}
			return nil
		})

		if err != nil {
			return 0, err
		}
	}

	return numLines, nil
}
