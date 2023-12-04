package compression

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestQueryExecution_ToRecording(t *testing.T) {
	bTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("test to recording with dependents", func(t *testing.T) {
		var exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"
		exec.SharedRecordings = append(exec.SharedRecordings, bTime.Add(time.Hour*24))

		got := exec.ToRecording("series1", "repoName1", 1, 5.0)
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})

	t.Run("test to recording without dependents", func(t *testing.T) {
		var exec QueryExecution
		exec.RecordingTime = bTime
		exec.Revision = "asdf1234"

		got := exec.ToRecording("series1", "repoName1", 1, 5.0)
		autogold.ExpectFile(t, got, autogold.ExportedOnly())
	})
}

func Test_GitserverFilter(t *testing.T) {

	tests := []struct {
		name              string
		want              autogold.Value
		fakeCommitFetcher fakeCommitFetcher
		times             []time.Time
	}{
		{
			name:              "no compression all times have a distinct commit",
			want:              autogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":null},{"Revision":"2","RecordingTime":"2021-02-01T00:00:00Z","SharedRecordings":null},{"Revision":"3","RecordingTime":"2021-03-01T00:00:00Z","SharedRecordings":null},{"Revision":"4","RecordingTime":"2021-04-01T00:00:00Z","SharedRecordings":null}],"RecordCount":4}`),
			fakeCommitFetcher: buildFakeFetcher("1", "2", "3", "4"),
		},
		{
			name:              "compress inner values with 2 executions",
			want:              autogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":["2021-02-01T00:00:00Z","2021-03-01T00:00:00Z"]},{"Revision":"2","RecordingTime":"2021-04-01T00:00:00Z","SharedRecordings":null}],"RecordCount":2}`),
			fakeCommitFetcher: buildFakeFetcher("1", "1", "1", "2"),
		},
		{
			name:              "all values compressed",
			want:              autogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":["2021-02-01T00:00:00Z","2021-03-01T00:00:00Z","2021-04-01T00:00:00Z"]}],"RecordCount":1}`),
			fakeCommitFetcher: buildFakeFetcher("1", "1", "1", "1"),
		},
		{
			name:              "no compression with one error",
			want:              autogold.Expect(`{"Executions":[{"Revision":"1","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":null},{"Revision":"2","RecordingTime":"2021-02-01T00:00:00Z","SharedRecordings":null},{"Revision":"","RecordingTime":"2021-03-01T00:00:00Z","SharedRecordings":null},{"Revision":"4","RecordingTime":"2021-04-01T00:00:00Z","SharedRecordings":null}],"RecordCount":4}`),
			fakeCommitFetcher: buildFakeFetcher("1", "2", errors.New("asdf"), "4"),
		},
		{
			name:              "no commits return for any points",
			want:              autogold.Expect(`{"Executions":[{"Revision":"","RecordingTime":"2021-01-01T00:00:00Z","SharedRecordings":null},{"Revision":"","RecordingTime":"2021-02-01T00:00:00Z","SharedRecordings":null}],"RecordCount":2}`),
			fakeCommitFetcher: buildFakeFetcher(),
			times: []time.Time{time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := gitserverFilter{commitFetcher: test.fakeCommitFetcher, logger: logtest.Scoped(t)}
			if test.times == nil {
				test.times = test.fakeCommitFetcher.toTimes()
			}
			got := filter.Filter(context.Background(), test.times, "myrepo")
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

func (f fakeCommitFetcher) RecentCommits(ctx context.Context, repoName api.RepoName, target time.Time, revision string) ([]*gitdomain.Commit, error) {
	got, ok := f.hashes[target]
	if !ok {
		return nil, f.errors[target]
	}
	return []*gitdomain.Commit{{ID: api.CommitID(got), Committer: &gitdomain.Signature{Date: target.Add(time.Hour * -1)}}}, nil
}
