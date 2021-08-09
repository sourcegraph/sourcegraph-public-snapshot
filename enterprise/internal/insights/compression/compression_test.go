package compression

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/insights/priority"

	"github.com/hexops/autogold"

	"github.com/cockroachdb/errors"
)

func TestFilterFrames(t *testing.T) {

	ctx := context.Background()

	maxHistorical := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	commitFilter := CommitFilter{
		maxHistorical: maxHistorical,
	}

	t.Run("test empty frames", func(t *testing.T) {
		got := commitFilter.FilterFrames(ctx, []Frame{}, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test one frame", func(t *testing.T) {
		input := []Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}}
		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test unable to fetch metadata", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}, {
			maxHistorical.Add(time.Second * 500), maxHistorical.Add(time.Second * 1000), "fedcba",
		}}
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{}, errors.New("really bad error"))

		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test no commits two frames", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}, {
			maxHistorical.Add(time.Second * 500), maxHistorical.Add(time.Second * 1000), "fedcba",
		}}

		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test three frames middle has no commits", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []Frame{{
			toTime("2020-05-01"), toTime("2020-06-01"), "abcdef",
		}, {
			toTime("2020-06-01"), toTime("2020-07-01"), "fedcba",
		}, {
			toTime("2020-07-01"), toTime("2020-08-01"), "111222333",
		}}

		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:        1,
			Enabled:       true,
			LastIndexedAt: toTime("2021-01-01"),
		}, nil)

		// The middle commit will actually be the first one to call Get
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{
			{
				RepoID:      2,
				Commit:      "21342134",
				CommittedAt: toTime("2020-07-02"),
			},
		}, nil)

		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test three frames middle has no commits but index is behind", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []Frame{{
			toTime("2020-05-01"), toTime("2020-06-01"), "abcdef",
		}, {
			toTime("2020-06-01"), toTime("2020-07-01"), "fedcba",
		}, {
			toTime("2020-07-01"), toTime("2020-08-01"), "111222333",
		}}

		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:        1,
			Enabled:       true,
			LastIndexedAt: toTime("2020-06-02"),
		}, nil)

		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}

func toTime(date string) time.Time {
	result, _ := time.Parse("2006-01-02", date)
	return result
}

func TestQueryExecution_ToRecording(t *testing.T) {
	bTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("test to recording with dependents", func(t *testing.T) {
		var exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"
		exec.SharedRecordings = append(exec.SharedRecordings, bTime.Add(time.Hour*24))

		got := exec.ToRecording("series1", "repoName1", 1, 5.0)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test to recording without dependents", func(t *testing.T) {
		var exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"

		got := exec.ToRecording("series1", "repoName1", 1, 5.0)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}

func TestQueryExecution_ToQueueJob(t *testing.T) {
	bTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("test to job with dependents", func(t *testing.T) {
		var exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"
		exec.SharedRecordings = append(exec.SharedRecordings, bTime.Add(time.Hour*24))

		got := exec.ToQueueJob("series1", "sourcegraphquery1", priority.Cost(500), priority.Low)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
	t.Run("test to job without dependents", func(t *testing.T) {
		var exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"

		got := exec.ToQueueJob("series1", "sourcegraphquery1", priority.Cost(500), priority.Low)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})
}
