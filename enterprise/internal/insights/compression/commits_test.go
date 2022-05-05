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

	makeCommit := func(id api.CommitID, commitTime time.Time) gitdomain.Commit {
		return gitdomain.Commit{
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

	t.Run("inserting the same commit twice is not an error asc commit order", func(t *testing.T) {
		time1 := time.Date(2021, time.April, 21, 1, 1, 0, 0, time.UTC)

		now := time.Now().UTC()
		commit1 := makeCommit("abc123", time1)
		commits := []*gitdomain.Commit{&commit1}

		err := commitStore.InsertCommits(context.Background(), 1, commits, now, "test")
		if err != nil {
			t.Errorf("unexpected insert error: %v want %v", err, nil)
		}

		time2 := time.Date(2021, time.April, 21, 1, 2, 0, 0, time.UTC)
		commit2 := makeCommit("abc456", time2)
		commits = append(commits, &commit2)
		err = commitStore.InsertCommits(context.Background(), 1, commits, now, "test")
		if err != nil {
			t.Errorf("unexpected insert error: %v want %v", err, nil)
		}

		afterLastItem := time2.Add(time.Second)
		savedCommits, err := commitStore.Get(context.Background(), 1, time1, afterLastItem)
		if err != nil {
			t.Errorf("unexpected get error: %v want %v", err, nil)
		}

		want := autogold.Want("asc insert", "[{1 abc456 2021-04-21 01:02:00 +0000 UTC} {1 abc123 2021-04-21 01:01:00 +0000 UTC}]")

		want.Equal(t, commitStampsToString(savedCommits), autogold.ExportedOnly())
	})

	t.Run("inserting the same commit twice is not an error desc commit order", func(t *testing.T) {
		time1 := time.Date(2021, time.May, 21, 1, 1, 0, 0, time.UTC)

		now := time.Now().UTC()
		commit1 := makeCommit("abc123", time1)
		commits := []*gitdomain.Commit{&commit1}

		err := commitStore.InsertCommits(context.Background(), 1, commits, now, "test")
		if err != nil {
			t.Errorf("unexpected insert error: %v want %v", err, nil)
		}

		time2 := time.Date(2021, time.May, 21, 1, 2, 0, 0, time.UTC)
		commit2 := makeCommit("abc456", time2)
		reverseCommits := []*gitdomain.Commit{&commit2, &commit1}
		err = commitStore.InsertCommits(context.Background(), 1, reverseCommits, now, "test")
		if err != nil {
			t.Errorf("unexpected insert error: %v want %v", err, nil)
		}

		afterLastItem := time2.Add(time.Second)
		savedCommits, err := commitStore.Get(context.Background(), 1, time1, afterLastItem)
		if err != nil {
			t.Errorf("unexpected get error: %v want %v", err, nil)
		}

		want := autogold.Want("desc insert", "[{1 abc456 2021-05-21 01:02:00 +0000 UTC} {1 abc123 2021-05-21 01:01:00 +0000 UTC}]")

		want.Equal(t, commitStampsToString(savedCommits))
	})

}

func commitStampsToString(stamps []CommitStamp) string {
	return fmt.Sprintf("%v", stamps)
}
