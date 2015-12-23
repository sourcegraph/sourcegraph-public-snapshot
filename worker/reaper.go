package worker

import (
	"math/rand"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/grpccache"
	"sourcegraph.com/sqs/pbtypes"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"golang.org/x/net/context"
)

// buildReaper periodically removes builds that have not sent a
// heartbeat within the allowable timeout. It should be run in a
// goroutine by itself.
func buildReaper(ctx context.Context) {
	// Add margin for clock skew.
	const heartbeatTimeout = serverHeartbeatInterval + 2*time.Minute

	listAllHeartbeatExpiredBuilds := func(ctx context.Context) ([]*sourcegraph.Build, error) {
		var expired []*sourcegraph.Build

		cl := sourcegraph.NewClientFromContext(ctx)
		for page := int32(1); ; page++ {
			builds, err := cl.Builds.List(grpccache.NoCache(ctx), &sourcegraph.BuildListOptions{
				Active:      true,
				ListOptions: sourcegraph.ListOptions{Page: page, PerPage: 100},
			})
			if err != nil {
				return nil, err
			}

			if len(builds.Builds) == 0 {
				break
			}

			for _, b := range builds.Builds {
				// A build that was just started might not yet have a
				// heartbeat, so accept the StartedAt time as the last
				// heartbeat-ish time.
				var last time.Time
				if b.HeartbeatAt != nil && b.HeartbeatAt.Time().After(last) {
					last = b.HeartbeatAt.Time()
				}
				if b.StartedAt != nil && b.StartedAt.Time().After(last) {
					last = b.StartedAt.Time()
				}
				if last.IsZero() {
					log15.Error("Active build has no StartedAt nor HeartbeatAt date", "build", b.Spec())
					continue
				}

				if time.Since(last) > heartbeatTimeout {
					expired = append(expired, b)
				}
			}
		}

		return expired, nil
	}

	// Random sleep to avoid thundering herd.
	time.Sleep(time.Second * time.Duration(3+rand.Intn(5)))

	for {
		cl := sourcegraph.NewClientFromContext(ctx)
		expiredBuilds, err := listAllHeartbeatExpiredBuilds(ctx)
		if err != nil {
			log15.Error("Error listing all heartbeat-expired builds", "err", err)
			continue
		}

		for _, b := range expiredBuilds {
			log15.Info("Marking heartbeat-expired build as failed", "build", b.Spec())
			now := pbtypes.NewTimestamp(time.Now())
			_, err := cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
				Build: b.Spec(),
				Info:  sourcegraph.BuildUpdate{Failure: true, Killed: true, EndedAt: &now},
			})
			if err != nil {
				log15.Error("Error updating heartbeat-expired build", "build", b.Spec(), "err", err)
			}
		}

		// Random sleep to avoid thundering herd.
		time.Sleep(time.Minute * time.Duration(3+rand.Intn(5)))
	}
}
