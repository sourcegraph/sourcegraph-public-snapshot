package compression

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

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
		got := commitFilter.FilterFrames(ctx, []types.Frame{}, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test one frame", func(t *testing.T) {
		input := []types.Frame{{
			maxHistorical, maxHistorical.Add(time.Second * 500), "abcdef",
		}}
		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test unable to fetch metadata", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []types.Frame{{
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
		input := []types.Frame{{
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
		input := []types.Frame{{
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

	t.Run("test multiple frames ensure previous frame is used for compression", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore

		// This test is a scenario from a bug discovered on a real insight. In this scenario there are 2 commits, each
		// in the middle of a frame. The goal of this test is to ensure that we use the correct frames to determine
		// if any changes have been made to query for the 'from' time point.
		input := []types.Frame{
			{
				toTime("2021-01-01"), toTime("2021-02-01"), "jan",
			},
			{
				toTime("2021-02-01"), toTime("2021-03-01"), "feb",
			},
			{
				toTime("2021-03-01"), toTime("2021-04-01"), "march",
			},
			{
				toTime("2021-04-01"), toTime("2021-05-01"), "april",
			},
			{
				toTime("2021-05-01"), toTime("2021-06-01"), "may",
			},
			{
				toTime("2021-06-01"), toTime("2021-07-01"), "june",
			},
			{
				toTime("2021-07-01"), toTime("2021-08-01"), "july",
			},
			{
				toTime("2021-08-01"), toTime("2021-08-15"), "aug",
			},
		}
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:        1,
			Enabled:       true,
			LastIndexedAt: toTime("2021-09-01"),
		}, nil)

		commitStore.GetFunc.PushReturn([]CommitStamp{
			{
				RepoID:      2,
				Commit:      "stamp1",
				CommittedAt: toTime("2021-01-16"),
			},
			{
				RepoID:      2,
				Commit:      "donotuse",
				CommittedAt: toTime("2021-01-15"),
			},
		}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{
			{
				RepoID:      2,
				Commit:      "stamp2",
				CommittedAt: toTime("2021-06-26"),
			},
		}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)

		got := commitFilter.FilterFrames(ctx, input, 2)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("test three frames middle has no commits but index is behind", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []types.Frame{{
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
