package authzchecked

import (
	"time"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/store"
)

// BuildLogs wraps base's methods with authorization checks.
func BuildLogs(base store.BuildLogs) store.BuildLogs { return &buildLogs{base} }

// buildLogs adds authorization checks to an underlying BuildLogs.
type buildLogs struct {
	noauthz store.BuildLogs
}

var _ store.BuildLogs = (*buildLogs)(nil)

func (s *buildLogs) Get(ctx context.Context, build sourcegraph.BuildSpec, tag, minID string, minTime, maxTime time.Time) (*sourcegraph.LogEntries, error) {
	if err := checkBuild(ctx, build, auth.Read); err != nil {
		return nil, err
	}
	return s.noauthz.Get(ctx, build, tag, minID, minTime, maxTime)
}
