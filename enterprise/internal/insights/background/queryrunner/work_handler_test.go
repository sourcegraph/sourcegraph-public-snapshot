package queryrunner

import (
	"context"
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/lib/errors"

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

		recordings, err := handler.generateComputeRecordings(ctx, &job, date)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("compute job with no dependencies", []string{
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.14 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.15 3.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute job with sub-repo permissions", func(t *testing.T) {
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

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIdFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			if id == 11 {
				return true, nil
			} else {
				return false, errors.New("Wrong repoID, try again")
			}
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := handler.generateComputeRecordings(ctx, &job, date)
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned as given repo has sub-repo permissions")
		}

		// Resetting DefaultSubRepoPermsChecker, so it won't affect further tests
		authz.DefaultSubRepoPermsChecker = nil
	})

	t.Run("compute job with sub-repo permissions resulted in error", func(t *testing.T) {
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

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIdFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			return false, errors.New("Oops")
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := handler.generateComputeRecordings(ctx, &job, date)
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned as given repo has an error during sub-repo permissions check")
		}

		// Resetting DefaultSubRepoPermsChecker, so it won't affect further tests
		authz.DefaultSubRepoPermsChecker = nil
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

		recordings, err := handler.generateComputeRecordings(ctx, &job, date)
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

		recordings, err := handler.generateComputeRecordings(ctx, &job, date)
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

func TestGenerateComputeRecordingsStream(t *testing.T) {
	t.Run("compute stream job with no dependencies", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				RepoCounts: map[string]*streaming.ComputeMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						ValueCounts: map[string]int{
							"1.15": 3,
							"1.14": 1,
						},
					},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("compute stream job with no dependencies", []string{
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.14 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.15 3.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute stream job with sub-repo permissions", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				RepoCounts: map[string]*streaming.ComputeMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						ValueCounts: map[string]int{
							"1.15": 3,
							"1.14": 1,
						},
					},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIdFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			if id == 11 {
				return true, nil
			} else {
				return false, errors.New("Wrong repoID, try again")
			}
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned as given repo has sub-repo permissions")
		}

		// Resetting DefaultSubRepoPermsChecker, so it won't affect further tests
		authz.DefaultSubRepoPermsChecker = nil
	})

	t.Run("compute stream job with sub-repo permissions resulted in error", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				RepoCounts: map[string]*streaming.ComputeMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						ValueCounts: map[string]int{
							"1.15": 3,
							"1.14": 1,
						},
					},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIdFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			return false, errors.New("Oops")
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned as given repo has an error during sub-repo permissions check")
		}

		// Resetting DefaultSubRepoPermsChecker, so it won't affect further tests
		authz.DefaultSubRepoPermsChecker = nil
	})

	t.Run("compute stream job with no dependencies multirepo", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				RepoCounts: map[string]*streaming.ComputeMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						ValueCounts: map[string]int{
							"1.11": 3,
							"1.18": 1,
						},
					},
					"github.com/sourcegraph/handbook": {
						RepositoryID:   5,
						RepositoryName: "github.com/sourcegraph/handbook",
						ValueCounts: map[string]int{
							"1.18": 2,
							"1.20": 1,
						},
					},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("compute stream job with no dependencies multirepo", []string{
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC 1.18 2.000000",
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC 1.20 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.18 1.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute stream job with dependencies", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				RepoCounts: map[string]*streaming.ComputeMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						ValueCounts: map[string]int{
							"1.11": 3,
							"1.18": 1,
							"1.33": 6,
						},
					},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("compute stream job with dependencies", []string{
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC 1.18 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC 1.33 6.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC 1.18 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC 1.33 6.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC 1.18 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC 1.33 6.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute stream job returns errors", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{}, errors.New("error")
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on compute stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
	})

	t.Run("compute stream job returns error event", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"error event"},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on compute stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		var streamingErr StreamingError
		if !errors.As(err, &streamingErr) {
			t.Errorf("Expected StreamingError, got %v", err)
		}
	})

	t.Run("compute stream job returns alert event", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Alerts: []string{"event"},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore:     nil,
			insightsStore:       nil,
			metadadataStore:     nil,
			limiter:             nil,
			mu:                  sync.RWMutex{},
			seriesCache:         nil,
			computeSearchStream: mocked,
		}

		recordings, err := handler.generateComputeRecordingsStream(context.Background(), &job, date)
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on compute stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if !strings.Contains(err.Error(), "alert") {
			t.Errorf("Expected alerts to return, got %v", err)
		}
	})
}

func TestGenerateSearchRecordingsStream(t *testing.T) {
	t.Run("search stream job with no dependencies", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				RepoCounts: map[string]*streaming.SearchMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						MatchCount:     5,
					},
				},
				TotalCount: 5,
			}, nil
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		// We only use series data for inserting dirty queries in the non-stream path so we can
		// ignore the argument here.
		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if err != nil {
			t.Error(err)
		}
		// Bearing in mind search series points don't store any values apart from count as the
		// value is the query. This translates into an empty space.
		stringified := stringify(recordings)
		autogold.Want("search stream job with no dependencies", []string{
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC  5.000000",
		}).Equal(t, stringified)
	})

	t.Run("search stream job with sub-repo permissions", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				RepoCounts: map[string]*streaming.SearchMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						MatchCount:     5,
					},
				},
				TotalCount: 5,
			}, nil
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIdFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			if id == 11 {
				return true, nil
			} else {
				return false, errors.New("Wrong repoID, try again")
			}
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		// We only use series data for inserting dirty queries in the non-stream path so we can
		// ignore the argument here.
		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned as given repo has sub-repo permissions")
		}

		// Resetting DefaultSubRepoPermsChecker, so it won't affect further tests
		authz.DefaultSubRepoPermsChecker = nil
	})

	t.Run("search stream job with sub-repo permissions resulted in error", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				RepoCounts: map[string]*streaming.SearchMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						MatchCount:     5,
					},
				},
				TotalCount: 5,
			}, nil
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIdFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			return false, errors.New("Oops")
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		// We only use series data for inserting dirty queries in the non-stream path so we can
		// ignore the argument here.
		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if err != nil {
			t.Error(err)
		}
		if len(recordings) != 0 {
			t.Error("No records should be returned as given repo has an error during sub-repo permissions check")
		}

		// Resetting DefaultSubRepoPermsChecker, so it won't affect further tests
		authz.DefaultSubRepoPermsChecker = nil
	})

	t.Run("search stream job with no dependencies multirepo", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				RepoCounts: map[string]*streaming.SearchMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						MatchCount:     5,
					},
					"github.com/sourcegraph/handbook": {
						RepositoryID:   5,
						RepositoryName: "github.com/sourcegraph/handbook",
						MatchCount:     20,
					},
				},
				TotalCount: 25,
			}, nil
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		// We only use series data for inserting dirty queries in the non-stream path so we can
		// ignore the argument here.
		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("search stream job with no dependencies multirepo", []string{
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC  20.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC  5.000000",
		}).Equal(t, stringified)
	})

	t.Run("search stream job with dependencies", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				RepoCounts: map[string]*streaming.SearchMatch{
					"github.com/sourcegraph/sourcegraph": {
						RepositoryID:   11,
						RepositoryName: "github.com/sourcegraph/sourcegraph",
						MatchCount:     5,
					},
				},
				TotalCount: 5,
			}, nil
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		// We only use series data for inserting dirty queries in the non-stream path so we can
		// ignore the argument here.
		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Want("search stream job with dependencies", []string{
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC  5.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC  5.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC  5.000000",
		}).Equal(t, stringified)
	})

	t.Run("search stream job returns errors", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{}, errors.New("error")
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
	})

	t.Run("search stream job returns error event", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"error event"},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on stream")
		}
		var streamingErr StreamingError
		if !errors.As(err, &streamingErr) {
			t.Errorf("Expected StreamingError, got %v", err)
		}
	})

	t.Run("search stream job returns alert event", func(t *testing.T) {
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

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"alert"},
				},
			}, nil
		}

		handler := workHandler{
			baseWorkerStore: nil,
			insightsStore:   nil,
			metadadataStore: nil,
			limiter:         nil,
			mu:              sync.RWMutex{},
			seriesCache:     nil,
			searchStream:    mocked,
		}

		recordings, err := handler.generateSearchRecordingsStream(context.Background(), &job, nil, date)
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if !strings.Contains(err.Error(), "alert") {
			t.Errorf("Expected alerts to return, got %v", err)
		}
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

func TestGetSeries(t *testing.T) {
	insightsDB := dbtest.NewInsightsDB(t)
	now := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC).Truncate(time.Microsecond).Round(0)
	metadataStore := store.NewInsightStore(insightsDB)
	metadataStore.Now = func() time.Time {
		return now
	}
	ctx := context.Background()

	workHandler := workHandler{
		metadadataStore: metadataStore,
		mu:              sync.RWMutex{},
		seriesCache:     make(map[string]*types.InsightSeries),
	}

	t.Run("series definition does not exist", func(t *testing.T) {
		_, err := workHandler.getSeries(ctx, "seriesshouldnotexist")
		if err == nil {
			t.Fatal("expected error from getSeries")
		}
		autogold.Want("series definition does not exist", "workHandler.getSeries: insight definition not found for series_id: seriesshouldnotexist").Equal(t, err.Error())
	})

	t.Run("series definition does exist", func(t *testing.T) {
		series, err := metadataStore.CreateSeries(ctx, types.InsightSeries{
			SeriesID:                   "arealseries",
			Query:                      "query1",
			CreatedAt:                  now,
			OldestHistoricalAt:         now,
			LastRecordedAt:             now,
			NextRecordingAfter:         now,
			LastSnapshotAt:             now,
			NextSnapshotAfter:          now,
			BackfillQueuedAt:           now,
			Enabled:                    true,
			Repositories:               nil,
			SampleIntervalUnit:         string(types.Month),
			SampleIntervalValue:        1,
			GeneratedFromCaptureGroups: false,
			JustInTime:                 false,
			GenerationMethod:           types.Search,
		})
		if err != nil {
			t.Error(err)
		}
		got, err := workHandler.getSeries(ctx, series.SeriesID)
		if err != nil {
			t.Fatal("unexpected error from getseries")
		}
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

}
