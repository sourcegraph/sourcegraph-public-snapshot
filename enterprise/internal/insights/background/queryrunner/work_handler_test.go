package queryrunner

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query"
)

func TestGenerateComputeRecordings(t *testing.T) {
	ctx := context.Background()

	t.Run("compute job with no dependencies", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := Job{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
			ID:              1,
			State:           "queued",
		}

		mocked := mockComputeSearch([]computeSearch{
			{
				repoName: "github.com/sourcegraph/sourcegraph",
				repoId:   11,
				values: []computeValue{
					{
						value:   "1.15",
						count:   3,
						path:    "package1/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
					{
						value:   "1.14",
						count:   1,
						path:    "package1/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
				},
			},
		})

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			computeSearch:   mocked,
		}

		recordings, err := handler.generateComputeRecordings(ctx, &job)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("compute job with no dependencies", []string{
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.14 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.15 3.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute job with no dependencies multirepo", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := Job{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
			ID:              1,
			State:           "queued",
		}

		mocked := mockComputeSearch([]computeSearch{
			{
				repoName: "github.com/sourcegraph/sourcegraph",
				repoId:   11,
				values: []computeValue{
					{
						value:   "1.18",
						count:   8,
						path:    "package1/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
					{
						value:   "1.11",
						count:   2,
						path:    "package2/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
				},
			},
			{
				repoName: "github.com/sourcegraph/handbook",
				repoId:   5,
				values: []computeValue{
					{
						value:   "1.20",
						count:   1,
						path:    "package3/go.mod",
						revhash: "asdfsdfer32r234234",
					},
					{
						value:   "1.18",
						count:   2,
						path:    "package4/go.mod",
						revhash: "asdfsdfer32r234234",
					},
				},
			},
		})

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			computeSearch:   mocked,
		}

		recordings, err := handler.generateComputeRecordings(ctx, &job)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("compute job with no dependencies multirepo", []string{
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC 1.18 2.000000",
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC 1.20 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.11 2.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.18 8.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute job with dependencies", func(t *testing.T) {
		date := time.Date(2021, 8, 1, 0, 0, 0, 0, time.UTC)
		job := Job{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: []time.Time{date.AddDate(0, 1, 0), date.AddDate(0, 2, 0)},
			ID:              1,
			State:           "queued",
		}

		mocked := mockComputeSearch([]computeSearch{
			{
				repoName: "github.com/sourcegraph/sourcegraph",
				repoId:   11,
				values: []computeValue{
					{
						value:   "1.1",
						count:   1,
						path:    "package1/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
					{
						value:   "1.22",
						count:   2,
						path:    "package1/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
				},
			},
			{
				repoName: "github.com/sourcegraph/sourcegraph",
				repoId:   11,
				values: []computeValue{
					{
						value:   "1.33",
						count:   3,
						path:    "package3/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
					{
						value:   "1.22",
						count:   4,
						path:    "package3/go.mod",
						revhash: "asdfsadf1234qwrar234",
					},
				},
			},
		})

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			computeSearch:   mocked,
		}

		recordings, err := handler.generateComputeRecordings(ctx, &job)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("compute job with dependencies", []string{
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC 1.1 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC 1.22 6.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC 1.33 3.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC 1.1 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC 1.22 6.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC 1.33 3.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC 1.1 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC 1.22 6.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC 1.33 3.000000",
		}).Equal(t, stringified)
	})
}

// stringify will turn the results of the recording worker into a slice of strings to easily compare golden test files against using autogold
func stringify(recordings []store.RecordSeriesPointArgs) []string {
	stringified := make([]string, 0, len(recordings))
	for _, recording := range recordings {
		// reponame repoId time captured value count
		capture := ""
		if recording.Point.Capture != nil {
			capture = *recording.Point.Capture
		}
		repoName := ""
		if recording.RepoName != nil {
			repoName = *recording.RepoName
		}
		repoId := api.RepoID(0)
		if recording.RepoID != nil {
			repoId = *recording.RepoID
		}
		stringified = append(stringified, fmt.Sprintf("%s %d %s %s %f", repoName, repoId, recording.Point.Time, capture, recording.Point.Value))
	}
	// sort for test determinism
	sort.Strings(stringified)
	return stringified
}

type computeValue struct {
	value   string
	count   int
	path    string
	revhash string
}

type computeSearch struct {
	repoName string
	repoId   int
	values   []computeValue
}

func mockComputeSearch(results []computeSearch) func(context.Context, string) ([]query.ComputeResult, error) {
	var mock []query.ComputeResult
	for _, result := range results {
		for _, value := range result.values {
			for i := 0; i < value.count; i++ {
				mock = append(mock, query.ComputeMatchContext{
					Commit: value.revhash,
					Repository: struct {
						Name string
						Id   string
					}{
						Name: result.repoName,
						Id:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("repo:%d", result.repoId))),
					},
					Path: value.path,
					Matches: []query.ComputeMatch{{
						Value: "",
						Environment: []query.ComputeEnvironmentEntry{{
							Variable: "1",
							Value:    value.value,
						}},
					}},
				})
			}
		}
	}
	return func(ctx context.Context, s string) ([]query.ComputeResult, error) {
		return mock, nil
	}
}
