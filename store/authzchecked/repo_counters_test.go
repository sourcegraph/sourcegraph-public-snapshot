package authzchecked

import (
	"testing"
	"time"

	"src.sourcegraph.com/sourcegraph/store/mockstore"

	"golang.org/x/net/context"
)

func TestRepoCounters_RecordHit(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledRecordHit bool
	s := RepoCounters(&mockstore.RepoCounters{
		RecordHit_: func(ctx context.Context, repo string) error { calledRecordHit = true; return nil },
	})

	if err := s.RecordHit(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if !calledRecordHit {
		t.Error("!calledRecordHit")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}

func TestRepoCounters_CountHits(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledCountHits bool
	s := RepoCounters(&mockstore.RepoCounters{
		CountHits_: func(ctx context.Context, repo string, since time.Time) (int, error) {
			calledCountHits = true
			return 0, nil
		},
	})

	if _, err := s.CountHits(ctx, "", time.Time{}); err != nil {
		t.Fatal(err)
	}
	if !calledCountHits {
		t.Error("!calledCountHits")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}
