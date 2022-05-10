package compression

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

func TestInsertCommits(t *testing.T) {
	insightsDB := dbtest.NewInsightsDB(t)
	commitStore := NewCommitStore(insightsDB)

	commit1 := makeCommit("abc123", time.Date(2021, time.April, 21, 1, 1, 0, 0, time.UTC))
	commit2 := makeCommit("abc234", time.Date(2021, time.April, 21, 1, 2, 0, 0, time.UTC))
	commit3 := makeCommit("bcd123", time.Date(2021, time.May, 21, 1, 1, 0, 0, time.UTC))
	commit4 := makeCommit("bcd234", time.Date(2021, time.May, 21, 1, 2, 0, 0, time.UTC))

	testCases := []struct {
		before       time.Time
		firstInsert  []*gitdomain.Commit
		secondInsert []*gitdomain.Commit
		after        time.Time
		want         autogold.Value
	}{
		{
			before:       time.Date(2021, time.April, 20, 1, 1, 0, 0, time.UTC),
			firstInsert:  []*gitdomain.Commit{commit1},
			secondInsert: []*gitdomain.Commit{commit1, commit2},
			after:        time.Date(2021, time.April, 22, 1, 1, 0, 0, time.UTC),
			want:         autogold.Want("asc", "[{1 abc234 2021-04-21 01:02:00 +0000 UTC} {1 abc123 2021-04-21 01:01:00 +0000 UTC}]"),
		},
		{
			before:       time.Date(2021, time.May, 20, 1, 1, 0, 0, time.UTC),
			firstInsert:  []*gitdomain.Commit{commit3},
			secondInsert: []*gitdomain.Commit{commit4, commit3},
			after:        time.Date(2021, time.May, 22, 1, 1, 0, 0, time.UTC),
			want:         autogold.Want("desc", "[{1 bcd234 2021-05-21 01:02:00 +0000 UTC} {1 bcd123 2021-05-21 01:01:00 +0000 UTC}]")},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			err := commitStore.InsertCommits(context.Background(), 1, tc.firstInsert, time.Now(), "test")
			if err != nil {
				t.Errorf("unexpected insert error: %v want %v", err, nil)
			}
			err = commitStore.InsertCommits(context.Background(), 1, tc.secondInsert, time.Now(), "test")
			if err != nil {
				t.Errorf("unexpected insert error: %v want %v", err, nil)
			}

			savedCommits, err := commitStore.Get(context.Background(), 1, tc.before, tc.after)
			if err != nil {
				t.Errorf("unexpected get error: %v want %v", err, nil)
			}

			tc.want.Equal(t, commitStampsToString(savedCommits))
		})
	}
}

func commitStampsToString(stamps []CommitStamp) string {
	return fmt.Sprintf("%v", stamps)
}

func makeCommit(id api.CommitID, commitTime time.Time) *gitdomain.Commit {
	return &gitdomain.Commit{
		ID: id,
		Author: gitdomain.Signature{
			Name:  "user",
			Email: "user@user.com",
			Date:  commitTime,
		},
		Committer: &gitdomain.Signature{
			Date:  commitTime,
			Name:  "user",
			Email: "user@user.com",
		},
		Message: "commit message",
	}
}
