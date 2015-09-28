package handlerutil

import (
	"reflect"
	"testing"
	"time"

	"strings"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
)

func TestGroupCommitsByDay(t *testing.T) {
	tm := func(value string) pbtypes.Timestamp {
		tm, err := time.Parse(time.RFC822Z, value)
		if err != nil {
			t.Fatal(err)
		}
		return pbtypes.NewTimestamp(tm.In(time.UTC))
	}

	tests := []struct {
		commits []*vcs.Commit
		want    map[string][]vcs.CommitID // day start instant time string -> commit IDs
	}{
		{
			commits: []*vcs.Commit{
				{ID: commitID("a"), Committer: &vcs.Signature{Date: tm("06 Jan 01 23:00 -0700")}}, // 7th (utc)
				{ID: commitID("b"), Committer: &vcs.Signature{Date: tm("06 Jan 01 12:00 -0700")}}, // 6th (utc)
				{ID: commitID("c"), Committer: &vcs.Signature{Date: tm("06 Jan 01 01:00 -0700")}}, // 6th (utc)
				{ID: "d", Committer: &vcs.Signature{Date: tm("05 Jan 01 19:00 -0700")}},           // 6th (utc)
				{ID: commitID("e"), Committer: &vcs.Signature{Date: tm("06 Jan 01 03:00 +0500")}}, // 5th (utc)
			},
			want: map[string][]vcs.CommitID{
				tm("07 Jan 01 00:00 -0000").Time().In(time.UTC).String(): []vcs.CommitID{commitID("a")},
				tm("06 Jan 01 00:00 -0000").Time().In(time.UTC).String(): []vcs.CommitID{commitID("b"), commitID("c"), "d"},
				tm("05 Jan 01 00:00 -0000").Time().In(time.UTC).String(): []vcs.CommitID{commitID("e")},
			},
		},
	}
	for _, test := range tests {
		days := GroupCommitsByDay(test.commits)
		for _, day := range days {
			wantCommitIDs := test.want[day.Start.String()]
			gotCommitIDs := extractCommitIDs(day.Commits)
			if !reflect.DeepEqual(gotCommitIDs, wantCommitIDs) {
				t.Errorf("day %s: got commit IDs %v, want %v", day.Start, gotCommitIDs, wantCommitIDs)
			}
		}
	}
}

func extractCommitIDs(commits []*vcs.Commit) []vcs.CommitID {
	ids := make([]vcs.CommitID, len(commits))
	for i, c := range commits {
		ids[i] = c.ID
	}
	return ids
}

func commitID(c string) vcs.CommitID { return vcs.CommitID(strings.Repeat(c, 40)) }
