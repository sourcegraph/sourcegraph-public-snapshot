package queryrunner

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/insights/query/streaming"
	"github.com/sourcegraph/sourcegraph/internal/insights/store"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	dbtypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestGenerateComputeRecordingsStream(t *testing.T) {
	t.Run("compute stream job with no dependencies", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Expect([]string{
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.14 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.15 3.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute stream job with sub-repo permissions", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIDFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			if id == 11 {
				return true, nil
			} else {
				return false, errors.New("Wrong repoID, try again")
			}
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
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
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIDFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			return false, errors.New("Oops")
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
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
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Expect([]string{
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC 1.18 2.000000",
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC 1.20 1.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.11 3.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC 1.18 1.000000",
		}).Equal(t, stringified)
	})

	t.Run("compute stream job with dependencies", func(t *testing.T) {
		date := time.Date(2021, 8, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: []time.Time{date.AddDate(0, 1, 0), date.AddDate(0, 2, 0)},
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

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Expect([]string{
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
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{}, errors.New("error")
		}

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on compute stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
	})

	t.Run("compute stream job returns retryable error event", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"error event"},
				},
			}, nil
		}

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on compute stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if strings.Contains(err.Error(), "terminal") {
			t.Errorf("Expected retryable error, got %v", err)
		}
	})

	t.Run("compute stream job returns terminal error event", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"not terminal", "invalid query"},
				},
			}, nil
		}

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on compute stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		var terminalError TerminalStreamingError
		if !errors.As(err, &terminalError) {
			t.Errorf("Expected terminal error, got %v", err)
		}
	})

	t.Run("compute stream job returns alert event", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.ComputeTabulationResult, error) {
			return &streaming.ComputeTabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Alerts: []string{"event"},
				},
			}, nil
		}

		recordings, err := generateComputeRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
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
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		// Bearing in mind search series points don't store any values apart from count as the
		// value is the query. This translates into an empty space.
		stringified := stringify(recordings)
		autogold.Expect([]string{
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC  5.000000",
		}).Equal(t, stringified)
	})

	t.Run("search stream job with sub-repo permissions", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIDFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			if id == 11 {
				return true, nil
			} else {
				return false, errors.New("Wrong repoID, try again")
			}
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
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
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		checker := authz.NewMockSubRepoPermissionChecker()
		checker.EnabledFunc.SetDefaultHook(func() bool {
			return true
		})
		checker.EnabledForRepoIDFunc.SetDefaultHook(func(ctx context.Context, id api.RepoID) (bool, error) {
			return false, errors.New("Oops")
		})

		// sub-repo permissions are enabled
		authz.DefaultSubRepoPermsChecker = checker

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
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
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
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

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Expect([]string{
			"github.com/sourcegraph/handbook 5 2021-12-01 00:00:00 +0000 UTC  20.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-12-01 00:00:00 +0000 UTC  5.000000",
		}).Equal(t, stringified)
	})

	t.Run("search stream job with dependencies", func(t *testing.T) {
		date := time.Date(2021, 8, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: []time.Time{date.AddDate(0, 1, 0), date.AddDate(0, 2, 0)},
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

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if err != nil {
			t.Error(err)
		}
		stringified := stringify(recordings)
		autogold.Expect([]string{
			"github.com/sourcegraph/sourcegraph 11 2021-08-01 00:00:00 +0000 UTC  5.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-09-01 00:00:00 +0000 UTC  5.000000",
			"github.com/sourcegraph/sourcegraph 11 2021-10-01 00:00:00 +0000 UTC  5.000000",
		}).Equal(t, stringified)
	})

	t.Run("search stream job returns errors", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{}, errors.New("error")
		}

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
	})

	t.Run("search stream job returns retryable error event", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"error event"},
				},
			}, nil
		}

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		if strings.Contains(err.Error(), "terminal") {
			t.Errorf("Expected retryable error, got %v", err)
		}
	})

	t.Run("search stream job returns retryable error event", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"retryable event", "invalid query"},
				},
			}, nil
		}

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
		if len(recordings) != 0 {
			t.Error("No records should be returned as we errored on stream")
		}
		if err == nil {
			t.Error("Expected error but received nil")
		}
		var terminalError TerminalStreamingError
		if !errors.As(err, &terminalError) {
			t.Errorf("Expected terminal error, got %v", err)
		}
	})

	t.Run("search stream job returns alert event", func(t *testing.T) {
		date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
		job := SearchJob{
			SeriesID:        "testseries1",
			SearchQuery:     "searchit",
			RecordTime:      &date,
			PersistMode:     "record",
			DependentFrames: nil,
		}

		mocked := func(context.Context, string) (*streaming.TabulationResult, error) {
			return &streaming.TabulationResult{
				StreamDecoderEvents: streaming.StreamDecoderEvents{
					Errors: []string{"alert"},
				},
			}, nil
		}

		recordings, err := generateSearchRecordingsStream(context.Background(), &job, date, mocked, logtest.Scoped(t))
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

func TestFilterRecordsingsByRepo(t *testing.T) {
	date := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)
	repo1 := &dbtypes.Repo{ID: 1, Name: "repo1"}
	repo2 := &dbtypes.Repo{ID: 2, Name: "repo2"}
	repo3 := &dbtypes.Repo{ID: 3, Name: "repo3"}
	repo4 := &dbtypes.Repo{ID: 4, Name: "repo4"}
	allRepos := []*dbtypes.Repo{repo1, repo2, repo3, repo4}
	oddRepos := []*dbtypes.Repo{repo1, repo3}

	r1p1 := store.RecordSeriesPointArgs{RepoID: &repo1.ID, RepoName: (*string)(&repo1.Name)}
	r1p2 := store.RecordSeriesPointArgs{RepoID: &repo1.ID, RepoName: (*string)(&repo1.Name)}
	r2p1 := store.RecordSeriesPointArgs{RepoID: &repo2.ID, RepoName: (*string)(&repo2.Name)}
	r2p2 := store.RecordSeriesPointArgs{RepoID: &repo2.ID, RepoName: (*string)(&repo2.Name)}
	r3p1 := store.RecordSeriesPointArgs{RepoID: &repo3.ID, RepoName: (*string)(&repo3.Name)}
	r3p2 := store.RecordSeriesPointArgs{RepoID: &repo3.ID, RepoName: (*string)(&repo3.Name)}
	r4p1 := store.RecordSeriesPointArgs{RepoID: &repo4.ID, RepoName: (*string)(&repo4.Name)}
	r4p2 := store.RecordSeriesPointArgs{RepoID: &repo4.ID, RepoName: (*string)(&repo4.Name)}
	nonRepoPoint := store.RecordSeriesPointArgs{Point: store.SeriesPoint{SeriesID: "testseries1", Value: 10}}

	recordings := []store.RecordSeriesPointArgs{r1p1, r1p2, r2p1, r2p2, r3p1, r3p2, r4p1, r4p2, nonRepoPoint}

	testJob := SearchJob{
		SeriesID:        "testseries1",
		SearchQuery:     "searchit",
		RecordTime:      &date,
		PersistMode:     "record",
		DependentFrames: nil,
	}
	testCases := []struct {
		name       string
		job        SearchJob
		series     types.InsightSeries
		repoList   []*dbtypes.Repo
		recordings []store.RecordSeriesPointArgs
		want       autogold.Value
	}{
		{
			name:       "AllReposEmptySlice",
			job:        testJob,
			series:     types.InsightSeries{Repositories: []string{}},
			repoList:   allRepos,
			recordings: recordings,
			want: autogold.Expect([]string{
				" 0 0001-01-01 00:00:00 +0000 UTC  10.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
		{
			name:       "AllReposNil",
			job:        testJob,
			series:     types.InsightSeries{Repositories: nil},
			repoList:   allRepos,
			recordings: recordings,
			want: autogold.Expect([]string{
				" 0 0001-01-01 00:00:00 +0000 UTC  10.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo4 4 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
		{
			name:       "OddRepos",
			job:        testJob,
			series:     types.InsightSeries{Repositories: []string{string(repo1.Name), string(repo3.Name)}},
			repoList:   oddRepos,
			recordings: recordings,
			want: autogold.Expect([]string{
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo1 1 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo3 3 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
		{
			name:       "Repo4NotFound",
			job:        testJob,
			series:     types.InsightSeries{Repositories: []string{string(repo2.Name), string(repo4.Name)}},
			repoList:   []*dbtypes.Repo{repo2},
			recordings: recordings,
			want: autogold.Expect([]string{
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
				"repo2 2 0001-01-01 00:00:00 +0000 UTC  0.000000",
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepoStore := dbmocks.NewMockRepoStore()
			mockRepoStore.ListFunc.SetDefaultReturn(tc.repoList, nil)

			got, _ := filterRecordingsBySeriesRepos(context.Background(), mockRepoStore, &tc.series, recordings)
			tc.want.Equal(t, stringify(got))
		})
	}
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
