package main

import (
	"sort"
	"strconv"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
)

func generateHistory(builds []buildkite.Build, windowStart time.Time, opts CheckOptions) (totals map[string]int, flakes map[string]int, incidents map[string]int) {
	// day:count
	totals = make(map[string]int)
	for _, b := range builds {
		totals[buildDate(b.CreatedAt.Time)] += 1
	}
	// day:count
	flakes = make(map[string]int)
	// day:minutes
	incidents = make(map[string]int)

	// Scan over all builds
	scanBuilds := builds
	lastPassedBuild := windowStart
	for len(scanBuilds) > 0 {
		var firstFailedBuildIndex int
		for i, b := range scanBuilds {
			if isBuildFailed(b, opts.BuildTimeout) {
				firstFailedBuildIndex = i
				break
			} else if isBuildPassed(b) {
				lastPassedBuild = b.CreatedAt.Time
			}
		}
		scanBuilds = scanBuilds[max(firstFailedBuildIndex-1, 0):]

		failed, exceeded, scanned := checkConsecutiveFailures(
			scanBuilds, opts.FailuresThreshold, opts.BuildTimeout, true)
		if exceeded {
			// Time from last passed build to oldest build in series
			firstFailed := failed[len(failed)-1]
			redTime := lastPassedBuild.Sub(firstFailed.BuildCreated)
			incidents[buildDate(firstFailed.BuildCreated)] += int(redTime.Minutes())
		} else {
			for _, f := range failed {
				// Raw count of failed builds on date
				flakes[buildDate(f.BuildCreated)] += 1
			}
		}

		if len(scanBuilds) > scanned {
			// Set most recent passed build in last batch
			for _, b := range scanBuilds[:scanned+1] {
				if isBuildPassed(b) {
					lastPassedBuild = b.CreatedAt.Time
				}
			}
			// Scan next batch
			scanBuilds = scanBuilds[scanned+1:]
		} else {
			scanBuilds = []buildkite.Build{}
		}
	}

	return
}

func buildDate(created time.Time) string {
	return created.Format("2006/01/02")
}

func mapToRecords(m map[string]int) (records [][]string) {
	for k, v := range m {
		records = append(records, []string{k, strconv.Itoa(v)})
	}
	// Sort by date ascending
	sort.Slice(records, func(i, j int) bool {
		iDate, _ := time.Parse("2006/01/02", records[i][0])
		jDate, _ := time.Parse("2006/01/02", records[j][0])
		return iDate.Before(jDate)
	})
	// TODO Fill in the gaps maybe?
	// prev := records[0]
	// for _, r := range records {
	// 	rDate, _ := time.Parse("2006/01/02", r[0])
	// 	prevDate, _ := time.Parse("2006/01/02", prev[0])
	// 	if rDate.Sub(prevDate) > 24*time.Hour {
	// 		records = append(a[:index+1], a[index:]...)
	// 		a[index] = value
	// 	}
	// 	prev = r
	// }
	return
}
