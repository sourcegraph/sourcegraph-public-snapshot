package authzchecked

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store/mockstore"
)

func TestBuildLogs_Get(t *testing.T) {
	ctx, rc := mockRepoCheckerContext()

	var calledGet bool
	s := BuildLogs(&mockstore.BuildLogs{
		Get_: func(ctx context.Context, build sourcegraph.BuildSpec, tag, minID string, minTime, maxTime time.Time) (*sourcegraph.LogEntries, error) {
			calledGet = true
			return nil, nil
		},
	})

	if _, err := s.Get(ctx, sourcegraph.BuildSpec{}, "", "", time.Time{}, time.Time{}); err != nil {
		t.Fatal(err)
	}
	if !calledGet {
		t.Error("!calledGet")
	}
	if !rc.calledCheckRepo {
		t.Error("!calledCheckRepo")
	}
}
