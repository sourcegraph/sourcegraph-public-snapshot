package compression

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

		oldest := toTime("2019-01-01") // sufficiently old to be before all of the inputs (non-relevant)
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:          1,
			Enabled:         true,
			LastIndexedAt:   toTime("2021-01-01"),
			OldestIndexedAt: &oldest,
		}, nil)

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

		oldest := toTime("2019-01-01") // sufficiently old to be before all of the inputs (non-relevant)
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:          1,
			Enabled:         true,
			LastIndexedAt:   toTime("2021-01-01"),
			OldestIndexedAt: &oldest,
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
		oldest := toTime("2019-01-01") // sufficiently old to be before all of the inputs (non-relevant)
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:          1,
			Enabled:         true,
			LastIndexedAt:   toTime("2021-09-01"),
			OldestIndexedAt: &oldest,
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

		oldest := toTime("2019-01-01") // sufficiently old to be before all of the inputs (non-relevant)
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:          1,
			Enabled:         true,
			LastIndexedAt:   toTime("2020-06-02"),
			OldestIndexedAt: &oldest,
		}, nil)

		commitStore.GetFunc.PushReturn([]CommitStamp{}, nil)
		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Equal(t, got, autogold.ExportedOnly())
	})

	t.Run("metadata indicates the index is empty (no commits are indexed)", func(t *testing.T) {
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
			RepoId:          1,
			Enabled:         true,
			LastIndexedAt:   toTime("2020-09-02"),
			OldestIndexedAt: nil, // this means no commits are in the index!
		}, nil)

		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Want("metadata indicates the index is empty (no commits are indexed)", "[{ 2020-05-01 00:00:00 +0000 UTC []},{ 2020-06-01 00:00:00 +0000 UTC []},{ 2020-07-01 00:00:00 +0000 UTC []}]").Equal(t, planToString(got), autogold.ExportedOnly())
	})

	t.Run("not enough history is indexed (oldest is after all commits)", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		input := []types.Frame{{
			toTime("2020-05-01"), toTime("2020-06-01"), "abcdef",
		}, {
			toTime("2020-06-01"), toTime("2020-07-01"), "fedcba",
		}, {
			toTime("2020-07-01"), toTime("2020-08-01"), "111222333",
		}}

		oldest := toTime("2020-07-02")
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:          1,
			Enabled:         true,
			LastIndexedAt:   toTime("2020-09-02"),
			OldestIndexedAt: &oldest,
		}, nil)

		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Want("not enough history is indexed (oldest is after all commits)", "[{ 2020-05-01 00:00:00 +0000 UTC []},{ 2020-06-01 00:00:00 +0000 UTC []},{ 2020-07-01 00:00:00 +0000 UTC []}]").Equal(t, planToString(got), autogold.ExportedOnly())
	})

	t.Run("not enough history is indexed (oldest is in the commit range)", func(t *testing.T) {
		commitStore := NewMockCommitStore()
		commitFilter.store = commitStore
		// in this test we should be able to compress the last commit. The first will always add (it's the first),
		// the second will fail to compress because the oldest commit falls inside the frame, the last commit
		// should compress into the second recording.
		input := []types.Frame{{
			toTime("2020-05-01"), toTime("2020-06-01"), "abcdef",
		}, {
			toTime("2020-06-01"), toTime("2020-07-01"), "fedcba",
		}, {
			toTime("2020-07-01"), toTime("2020-08-01"), "111222333",
		}}

		oldest := toTime("2020-06-02")
		commitStore.GetMetadataFunc.PushReturn(CommitIndexMetadata{
			RepoId:          1,
			Enabled:         true,
			LastIndexedAt:   toTime("2020-09-02"),
			OldestIndexedAt: &oldest,
		}, nil)

		got := commitFilter.FilterFrames(ctx, input, 1)
		autogold.Want("not enough history is indexed (oldest is in the commit range)", "[{ 2020-05-01 00:00:00 +0000 UTC []},{ 2020-06-01 00:00:00 +0000 UTC [2020-07-01 00:00:00 +0000 UTC]}]").Equal(t, planToString(got), autogold.ExportedOnly())
	})
}

func planToString(plan BackfillPlan) string {
	return fmt.Sprintf("%v", plan)
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

func Test_GitserverFilter(t *testing.T) {

	tests := []struct {
		want              autogold.Value
		fakeCommitFetcher fakeCommitFetcher
		times             []time.Time
	}{
		{
			want:              autogold.Want("no compression all times have a distinct commit", `{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":null},{"Revision":"2","RecordingTime":"2021-02-01T00:00:00Z","SharedRecordings":null},{"Revision":"3","RecordingTime":"2021-03-01T00:00:00Z","SharedRecordings":null},{"Revision":"4","RecordingTime":"2021-04-01T00:00:00Z","SharedRecordings":null}],"RecordCount":4}`),
			fakeCommitFetcher: buildFakeFetcher("1", "2", "3", "4"),
		},
		{
			want:              autogold.Want("compress inner values with 2 executions", `{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":["2021-02-01T00:00:00Z","2021-03-01T00:00:00Z"]},{"Revision":"2","RecordingTime":"2021-04-01T00:00:00Z","SharedRecordings":null}],"RecordCount":2}`),
			fakeCommitFetcher: buildFakeFetcher("1", "1", "1", "2"),
		},
		{
			want:              autogold.Want("all values compressed", `{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":["2021-02-01T00:00:00Z","2021-03-01T00:00:00Z","2021-04-01T00:00:00Z"]}],"RecordCount":1}`),
			fakeCommitFetcher: buildFakeFetcher("1", "1", "1", "1"),
		},
		{
			want:              autogold.Want("no compression with one error", `{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":null},{"Revision":"2","RecordingTime":"2021-02-01T00:00:00Z","SharedRecordings":null},{"Revision":"","RecordingTime":"2021-03-01T00:00:00Z","SharedRecordings":null},{"Revision":"4","RecordingTime":"2021-04-01T00:00:00Z","SharedRecordings":null}],"RecordCount":4}`),
			fakeCommitFetcher: buildFakeFetcher("1", "2", errors.New("asdf"), "4"),
		},
		{
			want:              autogold.Want("no commits return for any points", `{"Executions":[{"Revision":"","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":null},{"Revision":"","RecordingTime":"2021-02-01T00:00:00Z","SharedRecordings":null}],"RecordCount":2}`),
			fakeCommitFetcher: buildFakeFetcher(),
			times: []time.Time{time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	for _, test := range tests {
		t.Run(test.want.Name(), func(t *testing.T) {
			filter := gitserverFilter{commitFetcher: test.fakeCommitFetcher}
			if test.times == nil {
				test.times = test.fakeCommitFetcher.toTimes()
			}
			got := filter.GitserverFilter(context.Background(), test.times, "myrepo")
			jsonify, err := json.Marshal(got)
			if err != nil {
				t.Error(err)
			}
			test.want.Equal(t, string(jsonify))
		})
	}
}

// buildFakeFetcher returns a fake commit fetcher where each element in the input slice maps to a distinct timestamp in the provided order. Input
// can be either string (representing a hash) or an error
func buildFakeFetcher(input ...any) fakeCommitFetcher {
	current := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	fetcher := fakeCommitFetcher{
		hashes: make(map[time.Time]string),
		errors: make(map[time.Time]error),
	}
	for _, val := range input {
		switch v := val.(type) {
		case error:
			fetcher.errors[current] = v
		case string:
			fetcher.hashes[current] = v
		}
		current = current.AddDate(0, 1, 0)
	}
	return fetcher
}

type fakeCommitFetcher struct {
	hashes map[time.Time]string
	errors map[time.Time]error
}

func (f fakeCommitFetcher) toTimes() (times []time.Time) {
	for t := range f.hashes {
		times = append(times, t)
	}
	for t := range f.errors {
		times = append(times, t)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i].Before(times[j])
	})
	return times
}

func (f fakeCommitFetcher) RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time) ([]*gitdomain.Commit, error) {
	got, ok := f.hashes[target]
	if !ok {
		return nil, f.errors[target]
	}
	return []*gitdomain.Commit{{ID: api.CommitID(got), Committer: &gitdomain.Signature{Date: target.Add(time.Hour * -1)}}}, nil
}
